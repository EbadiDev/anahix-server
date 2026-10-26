package admin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/EbadiDev/anahix-server/internal/logic/provider/smsir"
	"github.com/EbadiDev/anahix-server/internal/model"
	"github.com/EbadiDev/anahix-server/internal/svc"
	"github.com/EbadiDev/anahix-server/internal/types"
	"github.com/EbadiDev/anahix-server/pkg/xerr"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CustomerAdminLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCustomerAdminLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CustomerAdminLogic {
	return &CustomerAdminLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// ListCustomers retrieves a paginated and filtered list of customers for the admin workbench
func (l *CustomerAdminLogic) ListCustomers(req *types.AdminListCustomersRequest) (*types.AdminListCustomersResponse, error) {
	page := req.Page
	if page <= 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 15
	}
	if pageSize > 100 {
		pageSize = 100
	}

	query := l.svcCtx.DB.WithContext(l.ctx).Model(&model.Customer{})

	// Search filter: phone, name, email, referral code
	if term := strings.TrimSpace(req.Search); term != "" {
		searchLike := "%" + term + "%"
		query = query.Where(
			"phone LIKE ? OR full_name LIKE ? OR first_name LIKE ? OR last_name LIKE ? OR email LIKE ? OR referral_code LIKE ?",
			searchLike, searchLike, searchLike, searchLike, searchLike, searchLike,
		)
	}

	// Filter by account status
	if req.Status == "active" {
		query = query.Where("is_active = ?", true)
	} else if req.Status == "blocked" {
		query = query.Where("is_active = ?", false)
	}

	// Filter by wallet balance
	if req.HasBalance != nil && *req.HasBalance {
		query = query.Where("balance_toman > 0")
	}

	// Filter by referral earnings
	if req.HasReferral != nil && *req.HasReferral {
		query = query.Where("referral_balance_toman > 0")
	}

	// Filter by completed profile
	if req.ProfileCompleted != nil && *req.ProfileCompleted {
		query = query.Where("full_name != '' AND full_name IS NOT NULL AND email != '' AND email IS NOT NULL")
	}

	// Filter by user tier
	if req.Tier != "" && req.Tier != "all" {
		query = query.Where("tier = ?", req.Tier)
	}

	// Total filtered count
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, xerr.NewErrCode(xerr.DatabaseError)
	}

	// Summary statistics across entire system
	var totalActive int64
	var totalBlocked int64
	var totalVIP int64
	var totalBalanceToman int64

	l.svcCtx.DB.WithContext(l.ctx).Model(&model.Customer{}).Where("is_active = ?", true).Count(&totalActive)
	l.svcCtx.DB.WithContext(l.ctx).Model(&model.Customer{}).Where("is_active = ?", false).Count(&totalBlocked)
	l.svcCtx.DB.WithContext(l.ctx).Model(&model.Customer{}).Where("tier = ?", "vip").Count(&totalVIP)
	
	type BalanceSum struct {
		Sum int64
	}
	var bSum BalanceSum
	l.svcCtx.DB.WithContext(l.ctx).Model(&model.Customer{}).Select("COALESCE(SUM(balance_toman), 0) as sum").Scan(&bSum)
	totalBalanceToman = bSum.Sum

	// Sorting
	orderClause := "created_at DESC"
	sortOrder := strings.ToUpper(strings.TrimSpace(req.SortOrder))
	if sortOrder != "ASC" && sortOrder != "DESC" {
		sortOrder = "DESC"
	}

	switch req.SortBy {
	case "last_login_at":
		orderClause = fmt.Sprintf("last_login_at %s NULLS LAST", sortOrder)
	case "balance":
		orderClause = fmt.Sprintf("balance_toman %s", sortOrder)
	case "referral_balance":
		orderClause = fmt.Sprintf("referral_balance_toman %s", sortOrder)
	case "profile_completed":
		orderClause = fmt.Sprintf("CASE WHEN (full_name != '' AND full_name IS NOT NULL AND email != '' AND email IS NOT NULL) THEN 1 ELSE 0 END %s, created_at DESC", sortOrder)
	case "created_at":
		orderClause = fmt.Sprintf("created_at %s", sortOrder)
	default:
		orderClause = fmt.Sprintf("created_at %s", sortOrder)
	}

	var customers []model.Customer
	offset := (page - 1) * pageSize
	if err := query.Order(orderClause).Limit(pageSize).Offset(offset).Find(&customers).Error; err != nil {
		return nil, xerr.NewErrCode(xerr.DatabaseError)
	}

	now := time.Now()
	items := make([]types.AdminCustomerListItem, 0, len(customers))

	for _, c := range customers {
		// Calculate store orders count and total spent
		var storeOrdersCount int64
		var storeTotalSpent int64
		l.svcCtx.DB.WithContext(l.ctx).Model(&model.Order{}).
			Where("(customer_id = ? OR customer_phone = ?) AND status NOT IN ('pending_payment', 'cancelled')", c.ID, c.Phone).
			Count(&storeOrdersCount)

		type SpentSum struct {
			Sum int64
		}
		var sSum SpentSum
		l.svcCtx.DB.WithContext(l.ctx).Model(&model.Order{}).
			Select("COALESCE(SUM(total_amount_toman), 0) as sum").
			Where("(customer_id = ? OR customer_phone = ?) AND status NOT IN ('pending_payment', 'cancelled')", c.ID, c.Phone).
			Scan(&sSum)
		storeTotalSpent = sSum.Sum

		// Calculate virtual number orders count and spent
		var vNumOrdersCount int64
		var vNumTotalSpent int64
		l.svcCtx.DB.WithContext(l.ctx).Model(&model.VirtualNumberOrder{}).
			Where("customer_phone = ? AND status NOT IN ('CANCELED', 'TIMEOUT')", c.Phone).
			Count(&vNumOrdersCount)

		var vSum SpentSum
		l.svcCtx.DB.WithContext(l.ctx).Model(&model.VirtualNumberOrder{}).
			Select("COALESCE(SUM(price_toman), 0) as sum").
			Where("customer_phone = ? AND status NOT IN ('CANCELED', 'TIMEOUT')", c.Phone).
			Scan(&vSum)
		vNumTotalSpent = vSum.Sum

		totalOrdersCount := storeOrdersCount + vNumOrdersCount
		totalSpentToman := storeTotalSpent + vNumTotalSpent

		isOnline := false
		if c.LastLoginAt != nil && now.Sub(*c.LastLoginAt) < 15*time.Minute {
			isOnline = true
		}

		tier := c.Tier
		if tier == "" {
			tier = "standard"
		}

		fullName := c.FullName
		if fullName == "" {
			fullName = strings.TrimSpace(c.FirstName + " " + c.LastName)
		}
		if fullName == "" {
			fullName = "کاربر بدون نام"
		}

		items = append(items, types.AdminCustomerListItem{
			ID:                   c.ID.String(),
			Phone:                c.Phone,
			AvatarURL:            c.AvatarURL,
			FirstName:            c.FirstName,
			LastName:             c.LastName,
			FullName:             fullName,
			Email:                c.Email,
			Tier:                 tier,
			BalanceToman:         c.BalanceToman,
			ReferralCode:         c.ReferralCode,
			ReferralBalanceToman: c.ReferralBalanceToman,
			TotalOrdersCount:     totalOrdersCount,
			StoreOrdersCount:     storeOrdersCount,
			VirtualOrdersCount:   vNumOrdersCount,
			TotalSpentToman:      totalSpentToman,
			IsActive:             c.IsActive,
			IsAdmin:              c.IsAdmin,
			BlockReason:          c.BlockReason,
			InternalNotes:        c.InternalNotes,
			LastLoginAt:          c.LastLoginAt,
			IsOnline:             isOnline,
			CreatedAt:            c.CreatedAt,
		})
	}

	return &types.AdminListCustomersResponse{
		Customers:         items,
		Total:             total,
		Page:              page,
		PageSize:          pageSize,
		TotalActive:       totalActive,
		TotalBlocked:      totalBlocked,
		TotalVIP:          totalVIP,
		TotalBalanceToman: totalBalanceToman,
	}, nil
}

// GetCustomer360 compiles the full 360-degree customer profile across all systems
func (l *CustomerAdminLogic) GetCustomer360(customerIDStr string) (*types.AdminCustomerDetail360Response, error) {
	var customer model.Customer
	customerUUID, err := uuid.Parse(customerIDStr)
	if err == nil {
		if err := l.svcCtx.DB.WithContext(l.ctx).Where("id = ?", customerUUID).First(&customer).Error; err != nil {
			return nil, xerr.NewErrMsg("کاربر مورد نظر یافت نشد")
		}
	} else {
		// Fallback by phone
		if err := l.svcCtx.DB.WithContext(l.ctx).Where("phone = ?", customerIDStr).First(&customer).Error; err != nil {
			return nil, xerr.NewErrMsg("کاربر مورد نظر یافت نشد")
		}
	}

	// 1. Compile base customer info
	now := time.Now()
	isOnline := false
	if customer.LastLoginAt != nil && now.Sub(*customer.LastLoginAt) < 15*time.Minute {
		isOnline = true
	}
	tier := customer.Tier
	if tier == "" {
		tier = "standard"
	}
	fullName := customer.FullName
	if fullName == "" {
		fullName = strings.TrimSpace(customer.FirstName + " " + customer.LastName)
	}
	if fullName == "" {
		fullName = "کاربر بدون نام"
	}

	// 2. Fetch Wallet Transactions
	var transactions []model.Transaction
	_ = l.svcCtx.DB.WithContext(l.ctx).
		Where("customer_id = ? OR customer_phone = ?", customer.ID, customer.Phone).
		Order("created_at DESC").
		Limit(50).
		Find(&transactions).Error

	// 3. Fetch Store Orders
	var orders []model.Order
	_ = l.svcCtx.DB.WithContext(l.ctx).
		Preload("Items").
		Where("customer_id = ? OR customer_phone = ?", customer.ID, customer.Phone).
		Order("created_at DESC").
		Limit(50).
		Find(&orders).Error

	orderBriefs := make([]types.AdminCustomerOrderBrief, 0, len(orders))
	var totalSpentToman int64
	for _, o := range orders {
		prodTitle := ""
		delivered := ""
		if len(o.Items) > 0 {
			prodTitle = o.Items[0].ProductTitle
			delivered = o.Items[0].DeliveredPayload
		}
		if o.Status != model.OrderPendingPayment && o.Status != model.OrderCancelled {
			totalSpentToman += o.TotalAmountToman
		}
		orderBriefs = append(orderBriefs, types.AdminCustomerOrderBrief{
			ID:               o.ID.String(),
			OrderNumber:      o.OrderNumber,
			Status:           o.Status,
			DeliveryType:     o.DeliveryType,
			TotalAmountToman: o.TotalAmountToman,
			PaymentMethod:    o.PaymentMethod,
			ProductTitle:     prodTitle,
			DeliveredPayload: delivered,
			Items:            o.Items,
			CreatedAt:        o.CreatedAt,
			PaidAt:           o.PaidAt,
		})
	}

	// 4. Fetch Virtual Number Orders
	var vnums []model.VirtualNumberOrder
	_ = l.svcCtx.DB.WithContext(l.ctx).
		Where("customer_phone = ?", customer.Phone).
		Order("created_at DESC").
		Limit(50).
		Find(&vnums).Error

	for _, v := range vnums {
		if v.Status != model.VNumStatusCanceled && v.Status != model.VNumStatusTimeout {
			totalSpentToman += v.PriceToman
		}
	}

	// 5. Fetch Affiliate & Referral Network Stats
	var referrals []model.Referral
	_ = l.svcCtx.DB.WithContext(l.ctx).
		Where("referrer_customer_id = ?", customer.ID).
		Order("created_at DESC").
		Find(&referrals).Error

	referredUsers := make([]types.AdminReferredUserItem, 0, len(referrals))
	for _, r := range referrals {
		var refUser model.Customer
		if err := l.svcCtx.DB.WithContext(l.ctx).Where("id = ?", r.ReferredCustomerID).First(&refUser).Error; err == nil {
			rName := refUser.FullName
			if rName == "" {
				rName = strings.TrimSpace(refUser.FirstName + " " + refUser.LastName)
			}
			if rName == "" {
				rName = "کاربر زیرمجموعه"
			}
			referredUsers = append(referredUsers, types.AdminReferredUserItem{
				ID:        refUser.ID.String(),
				Phone:     refUser.Phone,
				FullName:  rName,
				CreatedAt: r.CreatedAt,
			})
		}
	}

	var commissions []model.ReferralCommission
	_ = l.svcCtx.DB.WithContext(l.ctx).
		Where("referrer_customer_id = ?", customer.ID).
		Order("created_at DESC").
		Find(&commissions).Error

	var totalComm int64
	var transferredComm int64
	for _, comm := range commissions {
		totalComm += comm.CommissionToman
		if comm.Status == model.CommissionTransferredToWallet {
			transferredComm += comm.CommissionToman
		}
	}
	availableComm := totalComm - transferredComm
	if availableComm < 0 {
		availableComm = 0
	}

	affiliateStats := types.AdminCustomerAffiliateStats{
		ReferralCode:               customer.ReferralCode,
		ReferredCount:              int64(len(referrals)),
		TotalCommissionToman:       totalComm,
		TransferredCommissionToman: transferredComm,
		AvailableCommissionToman:   availableComm,
		ReferredUsers:              referredUsers,
		Commissions:                commissions,
	}

	// 6. Fetch Admin Action / Audit Logs for this customer
	var auditLogs []model.AdminActionLog
	_ = l.svcCtx.DB.WithContext(l.ctx).
		Where("resource_id = ? OR details LIKE ? OR description LIKE ?", customer.ID.String(), "%"+customer.Phone+"%", "%"+customer.Phone+"%").
		Order("created_at DESC").
		Limit(50).
		Find(&auditLogs).Error

	customerItem := types.AdminCustomerListItem{
		ID:                   customer.ID.String(),
		Phone:                customer.Phone,
		AvatarURL:            customer.AvatarURL,
		FirstName:            customer.FirstName,
		LastName:             customer.LastName,
		FullName:             fullName,
		Email:                customer.Email,
		Tier:                 tier,
		BalanceToman:         customer.BalanceToman,
		ReferralCode:         customer.ReferralCode,
		ReferralBalanceToman: customer.ReferralBalanceToman,
		TotalOrdersCount:     int64(len(orders) + len(vnums)),
		StoreOrdersCount:     int64(len(orders)),
		VirtualOrdersCount:   int64(len(vnums)),
		TotalSpentToman:      totalSpentToman,
		IsActive:             customer.IsActive,
		IsAdmin:              customer.IsAdmin,
		BlockReason:          customer.BlockReason,
		InternalNotes:        customer.InternalNotes,
		LastLoginAt:          customer.LastLoginAt,
		IsOnline:             isOnline,
		CreatedAt:            customer.CreatedAt,
	}

	return &types.AdminCustomerDetail360Response{
		Customer:           customerItem,
		WalletTransactions: transactions,
		Orders:             orderBriefs,
		VirtualNumbers:     vnums,
		AffiliateStats:     affiliateStats,
		AuditLogs:          auditLogs,
	}, nil
}

// AdjustWallet performs an atomic manual credit or debit on a customer's wallet with mandatory reason and audit log
func (l *CustomerAdminLogic) AdjustWallet(customerIDStr string, req *types.AdminAdjustCustomerWalletRequest, adminUsername string) (*types.AdminCustomerListItem, error) {
	if req.AmountToman <= 0 {
		return nil, xerr.NewErrMsg("مبلغ تراکنش باید بزرگتر از صفر باشد")
	}
	req.Reason = strings.TrimSpace(req.Reason)
	if req.Reason == "" {
		return nil, xerr.NewErrMsg("درج علت تراکنش برای ثبت در حسابرسی و اسناد مالی الزامی است")
	}
	action := strings.ToLower(strings.TrimSpace(req.Action))
	if action != "credit" && action != "debit" {
		return nil, xerr.NewErrMsg("نوع عملیات کیف پول نامعتبر است (تنها credit یا debit مجاز است)")
	}

	customerUUID, err := uuid.Parse(customerIDStr)
	if err != nil {
		return nil, xerr.NewErrMsg("شناسه کاربر نامعتبر است")
	}

	var updatedCustomer model.Customer

	// Atomic database transaction with row-level locking
	err = l.svcCtx.DB.WithContext(l.ctx).Transaction(func(tx *gorm.DB) error {
		var customer model.Customer
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", customerUUID).First(&customer).Error; err != nil {
			return errors.New("کاربر مورد نظر یافت نشد")
		}

		prevBalance := customer.BalanceToman
		var newBalance int64

		if action == "credit" {
			newBalance = prevBalance + req.AmountToman
		} else {
			if prevBalance < req.AmountToman {
				return fmt.Errorf("موجودی فعلی کاربر (%d تومان) کمتر از مبلغ کسر درخواستی (%d تومان) است", prevBalance, req.AmountToman)
			}
			newBalance = prevBalance - req.AmountToman
		}

		customer.BalanceToman = newBalance
		if err := tx.Save(&customer).Error; err != nil {
			return err
		}

		// Record financial transaction ledger
		txType := model.TxTypeAdminAdjustment
		actionDescFA := "شارژ دستی"
		if action == "debit" {
			actionDescFA = "کسر دستی"
		}

		metadataMap := map[string]interface{}{
			"admin_username": adminUsername,
			"action":         action,
			"reason":         req.Reason,
			"prev_balance":   prevBalance,
			"new_balance":    newBalance,
			"amount_toman":   req.AmountToman,
		}
		metaBytes, _ := json.Marshal(metadataMap)

		ledgerTx := model.Transaction{
			ID:                uuid.New(),
			TransactionNumber: fmt.Sprintf("TX-ADJ-%d-%s", time.Now().UnixNano(), uuid.New().String()[:6]),
			CustomerID:        &customer.ID,
			CustomerPhone:     customer.Phone,
			CustomerEmail:     customer.Email,
			AmountToman:       req.AmountToman,
			OrderType:         txType,
			PaymentPlatform:   model.PlatformBalance,
			Status:            model.TxStatusSuccess,
			Description:       fmt.Sprintf("%s توسط مدیر (%s): %s", actionDescFA, adminUsername, req.Reason),
			Metadata:          string(metaBytes),
		}
		if err := tx.Create(&ledgerTx).Error; err != nil {
			return err
		}

		// Record immutable audit log
		auditLog := model.AdminActionLog{
			AdminUsername: adminUsername,
			Action:        "adjust_customer_wallet",
			Resource:      "customer/wallet",
			ResourceID:    customer.ID.String(),
			Description:   fmt.Sprintf("%s موجودی کاربر %s به مبلغ %d تومان توسط %s (علت: %s)", actionDescFA, customer.Phone, req.AmountToman, adminUsername, req.Reason),
			Details:       string(metaBytes),
		}
		if err := tx.Create(&auditLog).Error; err != nil {
			return err
		}

		updatedCustomer = customer
		return nil
	})

	if err != nil {
		return nil, xerr.NewErrMsg(err.Error())
	}

	fullName := updatedCustomer.FullName
	if fullName == "" {
		fullName = strings.TrimSpace(updatedCustomer.FirstName + " " + updatedCustomer.LastName)
	}

	return &types.AdminCustomerListItem{
		ID:                   updatedCustomer.ID.String(),
		Phone:                updatedCustomer.Phone,
		AvatarURL:            updatedCustomer.AvatarURL,
		FirstName:            updatedCustomer.FirstName,
		LastName:             updatedCustomer.LastName,
		FullName:             fullName,
		Email:                updatedCustomer.Email,
		Tier:                 updatedCustomer.Tier,
		BalanceToman:         updatedCustomer.BalanceToman,
		ReferralCode:         updatedCustomer.ReferralCode,
		ReferralBalanceToman: updatedCustomer.ReferralBalanceToman,
		IsActive:             updatedCustomer.IsActive,
		IsAdmin:              updatedCustomer.IsAdmin,
		BlockReason:          updatedCustomer.BlockReason,
		InternalNotes:        updatedCustomer.InternalNotes,
		LastLoginAt:          updatedCustomer.LastLoginAt,
		CreatedAt:            updatedCustomer.CreatedAt,
	}, nil
}

// UpdateCustomer updates customer access control, tier, notes, or profile fields
func (l *CustomerAdminLogic) UpdateCustomer(customerIDStr string, req *types.AdminUpdateCustomerStatusRequest, adminUsername string) (*types.AdminCustomerListItem, error) {
	customerUUID, err := uuid.Parse(customerIDStr)
	if err != nil {
		return nil, xerr.NewErrMsg("شناسه کاربر نامعتبر است")
	}

	var customer model.Customer
	if err := l.svcCtx.DB.WithContext(l.ctx).Where("id = ?", customerUUID).First(&customer).Error; err != nil {
		return nil, xerr.NewErrMsg("کاربر مورد نظر یافت نشد")
	}

	changes := make(map[string]interface{})

	if req.IsActive != nil {
		customer.IsActive = *req.IsActive
		changes["is_active"] = *req.IsActive
	}
	if req.BlockReason != nil {
		customer.BlockReason = strings.TrimSpace(*req.BlockReason)
		changes["block_reason"] = customer.BlockReason
	}
	if req.Tier != nil && *req.Tier != "" {
		tier := strings.ToLower(strings.TrimSpace(*req.Tier))
		if tier == "standard" || tier == "colleague" || tier == "vip" {
			customer.Tier = tier
			changes["tier"] = tier
		}
	}
	if req.InternalNotes != nil {
		customer.InternalNotes = *req.InternalNotes
		changes["internal_notes"] = customer.InternalNotes
	}
	if req.FullName != nil {
		customer.FullName = strings.TrimSpace(*req.FullName)
		changes["full_name"] = customer.FullName
	}
	if req.Email != nil {
		customer.Email = strings.TrimSpace(*req.Email)
		changes["email"] = customer.Email
	}
	if req.NewPassword != nil && strings.TrimSpace(*req.NewPassword) != "" {
		newPass := strings.TrimSpace(*req.NewPassword)
		if len(newPass) < 6 {
			return nil, xerr.NewErrMsg("رمز عبور جدید باید حداقل ۶ کاراکتر باشد")
		}
		if err := customer.SetPassword(newPass); err != nil {
			return nil, xerr.NewErrMsg("خطا در هش‌گذاری رمز عبور جدید")
		}
		changes["password_changed"] = true
	}

	if err := l.svcCtx.DB.WithContext(l.ctx).Save(&customer).Error; err != nil {
		return nil, xerr.NewErrCode(xerr.DatabaseError)
	}

	changesJSON, _ := json.Marshal(changes)
	auditLog := model.AdminActionLog{
		AdminUsername: adminUsername,
		Action:        "update_customer_profile",
		Resource:      "customer",
		ResourceID:    customer.ID.String(),
		Description:   fmt.Sprintf("به‌روزرسانی پرونده و دسترسی‌های کاربر %s توسط مدیر %s", customer.Phone, adminUsername),
		Details:       string(changesJSON),
	}
	_ = l.svcCtx.DB.WithContext(l.ctx).Create(&auditLog)

	fullName := customer.FullName
	if fullName == "" {
		fullName = strings.TrimSpace(customer.FirstName + " " + customer.LastName)
	}

	return &types.AdminCustomerListItem{
		ID:                   customer.ID.String(),
		Phone:                customer.Phone,
		AvatarURL:            customer.AvatarURL,
		FirstName:            customer.FirstName,
		LastName:             customer.LastName,
		FullName:             fullName,
		Email:                customer.Email,
		Tier:                 customer.Tier,
		BalanceToman:         customer.BalanceToman,
		ReferralCode:         customer.ReferralCode,
		ReferralBalanceToman: customer.ReferralBalanceToman,
		IsActive:             customer.IsActive,
		IsAdmin:              customer.IsAdmin,
		BlockReason:          customer.BlockReason,
		InternalNotes:        customer.InternalNotes,
		LastLoginAt:          customer.LastLoginAt,
		CreatedAt:            customer.CreatedAt,
	}, nil
}

// SendDirectSMS sends a direct SMS notification to the customer via configured SMS.ir gateway
func (l *CustomerAdminLogic) SendDirectSMS(customerIDStr string, req *types.AdminSendCustomerSMSRequest, adminUsername string) error {
	msg := strings.TrimSpace(req.Message)
	if msg == "" {
		return xerr.NewErrMsg("متن پیامک ارسالی نمی‌تواند خالی باشد")
	}

	customerUUID, err := uuid.Parse(customerIDStr)
	if err != nil {
		return xerr.NewErrMsg("شناسه کاربر نامعتبر است")
	}

	var customer model.Customer
	if err := l.svcCtx.DB.WithContext(l.ctx).Where("id = ?", customerUUID).First(&customer).Error; err != nil {
		return xerr.NewErrMsg("کاربر مورد نظر یافت نشد")
	}

	phone := smsir.NormalizeMobile(customer.Phone)
	if phone == "" {
		return xerr.NewErrMsg("شماره موبایل کاربر نامعتبر است")
	}

	// Fetch SMS.ir configuration from system settings
	var settings []model.SystemSetting
	_ = l.svcCtx.DB.WithContext(l.ctx).Where("category = ?", "sms").Find(&settings).Error

	configMap := make(map[string]string)
	for _, s := range settings {
		configMap[s.Key] = s.Value
	}

	apiKey := strings.TrimSpace(configMap["smsir_api_key"])
	lineStr := strings.Trim(configMap["smsir_line_number"], "'\" \t\r\n")
	lineNum, _ := strconv.ParseInt(lineStr, 10, 64)

	// If template ID is specified or fast verify mode is requested
	if req.TemplateID != nil && *req.TemplateID > 0 {
		_, sendErr := l.svcCtx.SMSIRClient.SendVerify(l.ctx, apiKey, phone, *req.TemplateID, msg)
		if sendErr != nil {
			return fmt.Errorf("خطا در ارسال پیامک سریع: %w", sendErr)
		}
	} else {
		// Bulk standard text send
		_, sendErr := l.svcCtx.SMSIRClient.SendBulk(l.ctx, apiKey, lineNum, msg, []string{phone})
		if sendErr != nil {
			return fmt.Errorf("خطا در ارسال پیامک: %w", sendErr)
		}
	}

	// Log audit trail
	auditLog := model.AdminActionLog{
		AdminUsername: adminUsername,
		Action:        "send_customer_sms",
		Resource:      "customer/sms",
		ResourceID:    customer.ID.String(),
		Description:   fmt.Sprintf("ارسال پیامک به کاربر %s توسط مدیر %s: %s", customer.Phone, adminUsername, msg),
		Details:       fmt.Sprintf(`{"mobile":"%s","message":"%s"}`, phone, msg),
	}
	_ = l.svcCtx.DB.WithContext(l.ctx).Create(&auditLog)

	return nil
}

// ReplaceOrderKey updates a delivered license key or account payload on a specific order item
func (l *CustomerAdminLogic) ReplaceOrderKey(customerIDStr string, req *types.AdminReplaceCustomerKeyRequest, adminUsername string) error {
	itemUUID, err := uuid.Parse(req.OrderItemID)
	if err != nil {
		return xerr.NewErrMsg("شناسه ردیف سفارش نامعتبر است")
	}

	newPayload := strings.TrimSpace(req.NewPayload)
	if newPayload == "" {
		return xerr.NewErrMsg("کلید یا لایسنس جدید نمی‌تواند خالی باشد")
	}

	var item model.OrderItem
	if err := l.svcCtx.DB.WithContext(l.ctx).Where("id = ?", itemUUID).First(&item).Error; err != nil {
		return xerr.NewErrMsg("ردیف سفارش مورد نظر یافت نشد")
	}

	oldPayload := item.DeliveredPayload
	item.DeliveredPayload = newPayload
	if err := l.svcCtx.DB.WithContext(l.ctx).Save(&item).Error; err != nil {
		return xerr.NewErrCode(xerr.DatabaseError)
	}

	// Record audit log
	detailsMap := map[string]interface{}{
		"order_item_id": item.ID.String(),
		"order_id":      item.OrderID.String(),
		"product_title": item.ProductTitle,
		"old_payload":   oldPayload,
		"new_payload":   newPayload,
		"reason":        req.Reason,
	}
	detailsJSON, _ := json.Marshal(detailsMap)

	auditLog := model.AdminActionLog{
		AdminUsername: adminUsername,
		Action:        "replace_order_key",
		Resource:      "order/item",
		ResourceID:    item.ID.String(),
		Description:   fmt.Sprintf("تعویض کلید تحویلی سفارش %s توسط مدیر %s (علت: %s)", item.ProductTitle, adminUsername, req.Reason),
		Details:       string(detailsJSON),
	}
	_ = l.svcCtx.DB.WithContext(l.ctx).Create(&auditLog)

	return nil
}
