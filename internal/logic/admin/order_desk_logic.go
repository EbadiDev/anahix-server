package admin

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/EbadiDev/anahix-server/internal/logic/loyalty"
	"github.com/EbadiDev/anahix-server/internal/logic/notification"
	"github.com/EbadiDev/anahix-server/internal/model"
	"github.com/EbadiDev/anahix-server/internal/svc"
	"github.com/EbadiDev/anahix-server/internal/types"
	"github.com/EbadiDev/anahix-server/pkg/xerr"
	"github.com/google/uuid"
)

type OrderDeskLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewOrderDeskLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OrderDeskLogic {
	return &OrderDeskLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

type AdminQueueOrderItem struct {
	ID                   uuid.UUID             `json:"id"`
	OrderID              string                `json:"order_id"`
	OrderNumber          string                `json:"order_number"`
	CustomerPhone        string                `json:"customer_phone"`
	CustomerEmail        string                `json:"customer_email"`
	Status               model.OrderStatus     `json:"status"`
	DeliveryType         model.DeliveryType    `json:"delivery_type"`
	TotalAmountToman     int64                 `json:"total_amount_toman"`
	LoyaltyDiscountToman int64                 `json:"loyalty_discount_toman"`
	CustomerTier         string                `json:"customer_tier"`
	PriorityScore        int                   `json:"priority_score"`
	PaymentMethod        model.PaymentMethod   `json:"payment_method"`
	ProductTitle         string                `json:"product_title"`
	CustomerCredentials  string                `json:"customer_credentials,omitempty"`
	AccountEmail         string                `json:"account_email,omitempty"`
	AccountPassword      string                `json:"account_password,omitempty"`
	TwoFactorPrompt      string                `json:"two_factor_prompt,omitempty"`
	TwoFactorCode        string                `json:"two_factor_code,omitempty"`
	TwoFactorStatus      model.TwoFactorStatus `json:"two_factor_status"`
	TwoFactorRequestedAt *time.Time            `json:"two_factor_requested_at,omitempty"`
	CreatedAt            time.Time             `json:"created_at"`
	UpdatedAt            time.Time             `json:"updated_at"`
}

type AdminOrderDetailResponse struct {
	Order               model.Order `json:"order"`
	CustomerCredentials string      `json:"customer_credentials,omitempty"`
}

func (l *OrderDeskLogic) ListQueueOrders(status string, deliveryType string) ([]AdminQueueOrderItem, error) {
	query := l.svcCtx.DB.WithContext(l.ctx).Preload("Items").Order("priority_score desc, created_at asc")

	if status != "" {
		query = query.Where("status = ?", status)
	} else {
		// Default to orders needing operator attention
		query = query.Where("status IN ?", []model.OrderStatus{
			model.OrderPaid,
			model.OrderInQueue,
			model.OrderProcessing,
			model.OrderAwaitingCustomerAction,
		})
	}

	if deliveryType != "" {
		query = query.Where("delivery_type = ?", deliveryType)
	}

	var orders []model.Order
	if err := query.Find(&orders).Error; err != nil {
		return nil, xerr.NewErrCode(xerr.DatabaseError)
	}

	resultList := make([]AdminQueueOrderItem, 0, len(orders))
	for _, o := range orders {
		item := AdminQueueOrderItem{
			ID:                   o.ID,
			OrderID:              o.ID.String(),
			OrderNumber:          o.OrderNumber,
			CustomerPhone:        o.CustomerPhone,
			CustomerEmail:        o.CustomerEmail,
			Status:               o.Status,
			DeliveryType:         o.DeliveryType,
			TotalAmountToman:     o.TotalAmountToman,
			LoyaltyDiscountToman: o.LoyaltyDiscountToman,
			CustomerTier:         o.CustomerTier,
			PriorityScore:        o.PriorityScore,
			PaymentMethod:        o.PaymentMethod,
			TwoFactorPrompt:      o.TwoFactorPrompt,
			TwoFactorCode:        o.TwoFactorCode,
			TwoFactorStatus:      o.TwoFactorStatus,
			TwoFactorRequestedAt: o.TwoFactorRequestedAt,
			CreatedAt:            o.CreatedAt,
			UpdatedAt:            o.UpdatedAt,
		}

		if len(o.Items) > 0 {
			item.ProductTitle = o.Items[0].ProductTitle
		}

		if o.CustomerCredentialsEncrypted != "" && l.svcCtx.Vault != nil {
			dec, decErr := l.svcCtx.Vault.Decrypt(o.CustomerCredentialsEncrypted)
			if decErr == nil && dec != "" {
				item.CustomerCredentials = dec
				// Parse "Email: ... | Pass: ... | Notes: ..."
				parts := strings.Split(dec, " | ")
				for _, part := range parts {
					if strings.HasPrefix(part, "Email: ") {
						item.AccountEmail = strings.TrimPrefix(part, "Email: ")
					} else if strings.HasPrefix(part, "Pass: ") {
						item.AccountPassword = strings.TrimPrefix(part, "Pass: ")
					}
				}
			}
		}

		resultList = append(resultList, item)
	}

	return resultList, nil
}

func (l *OrderDeskLogic) GetOrderDetail(orderIDStr string) (*AdminOrderDetailResponse, error) {
	var order model.Order
	orderUUID, err := uuid.Parse(orderIDStr)
	if err == nil {
		if err := l.svcCtx.DB.WithContext(l.ctx).Preload("Items").Where("id = ?", orderUUID).First(&order).Error; err != nil {
			return nil, xerr.NewErrCode(xerr.OrderNotFound)
		}
	} else {
		if err := l.svcCtx.DB.WithContext(l.ctx).Preload("Items").Where("order_number = ?", orderIDStr).First(&order).Error; err != nil {
			return nil, xerr.NewErrCode(xerr.OrderNotFound)
		}
	}

	plainCreds := ""
	if order.CustomerCredentialsEncrypted != "" && l.svcCtx.Vault != nil {
		dec, decErr := l.svcCtx.Vault.Decrypt(order.CustomerCredentialsEncrypted)
		if decErr == nil {
			plainCreds = dec
		}
	}

	return &AdminOrderDetailResponse{
		Order:               order,
		CustomerCredentials: plainCreds,
	}, nil
}

func (l *OrderDeskLogic) RequestTwoFactor(orderIDStr string, req *types.RequestTwoFactorRequest) error {
	var order model.Order
	orderUUID, err := uuid.Parse(orderIDStr)
	if err == nil {
		if err := l.svcCtx.DB.WithContext(l.ctx).Where("id = ?", orderUUID).First(&order).Error; err != nil {
			return xerr.NewErrCode(xerr.OrderNotFound)
		}
	} else {
		if err := l.svcCtx.DB.WithContext(l.ctx).Where("order_number = ?", orderIDStr).First(&order).Error; err != nil {
			return xerr.NewErrCode(xerr.OrderNotFound)
		}
	}

	now := time.Now()
	order.Status = model.OrderAwaitingCustomerAction
	order.TwoFactorStatus = model.TwoFactorRequested
	order.TwoFactorPrompt = req.PromptMessage
	order.TwoFactorRequestedAt = &now

	if err := l.svcCtx.DB.WithContext(l.ctx).Save(&order).Error; err != nil {
		return xerr.NewErrCode(xerr.DatabaseError)
	}

	// Dispatch Pattern SMS to customer alerting them of 2FA request (bypasses blacklist)
	if order.CustomerPhone != "" {
		notifier := notification.NewSMSPatternNotifier(l.svcCtx)
		notifier.SendTwoFactorRequestAsync(order.CustomerPhone, order.OrderNumber, req.PromptMessage)
	}

	return nil
}

func (l *OrderDeskLogic) CompleteOrder(orderIDStr string, req *types.CompleteOrderRequest) error {
	var order model.Order
	orderUUID, err := uuid.Parse(orderIDStr)
	if err == nil {
		if err := l.svcCtx.DB.WithContext(l.ctx).Where("id = ?", orderUUID).First(&order).Error; err != nil {
			return xerr.NewErrCode(xerr.OrderNotFound)
		}
	} else {
		if err := l.svcCtx.DB.WithContext(l.ctx).Where("order_number = ?", orderIDStr).First(&order).Error; err != nil {
			return xerr.NewErrCode(xerr.OrderNotFound)
		}
	}

	now := time.Now()
	order.Status = model.OrderCompleted
	order.CompletedAt = &now
	order.TwoFactorStatus = model.TwoFactorVerified
	if req.OperatorNotes != "" {
		order.OperatorNotes = req.OperatorNotes
	}

	// ZEROIZE / PURGE customer credentials for privacy guardrail
	order.CustomerCredentialsEncrypted = ""

	if err := l.svcCtx.DB.WithContext(l.ctx).Save(&order).Error; err != nil {
		return xerr.NewErrCode(xerr.DatabaseError)
	}

	// Dispatch Pattern SMS upon completion and license delivery
	if order.CustomerPhone != "" {
		// Recalculate customer tier and lifetime spend asynchronously
		go func(phone string) {
			_ = loyalty.RecalculateByPhone(context.Background(), l.svcCtx.DB, phone)
		}(order.CustomerPhone)

		var item model.OrderItem
		_ = l.svcCtx.DB.WithContext(l.ctx).Where("order_id = ?", order.ID).First(&item).Error
		payload := item.DeliveredPayload
		if payload == "" {
			payload = item.ProductTitle
		}
		notifier := notification.NewSMSPatternNotifier(l.svcCtx)
		notifier.SendOrderCompletedAsync(order.CustomerPhone, order.OrderNumber, payload)
	}

	return nil
}

type AdminOrderListResult struct {
	Orders   []AdminQueueOrderItem `json:"orders"`
	Total    int64                 `json:"total"`
	Page     int                   `json:"page"`
	PageSize int                   `json:"page_size"`
}

// ListAllOrders lists all orders with advanced search and status/payment filters
func (l *OrderDeskLogic) ListAllOrders(status string, deliveryType string, paymentMethod string, search string, page int, pageSize int) (*AdminOrderListResult, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	query := l.svcCtx.DB.WithContext(l.ctx).Model(&model.Order{}).Preload("Items")

	if status != "" && status != "all" {
		query = query.Where("status = ?", status)
	}

	if deliveryType != "" && deliveryType != "all" {
		query = query.Where("delivery_type = ?", deliveryType)
	}

	if paymentMethod != "" && paymentMethod != "all" {
		query = query.Where("payment_method = ?", paymentMethod)
	}

	if term := strings.TrimSpace(search); term != "" {
		sLike := "%" + term + "%"
		query = query.Where("order_number LIKE ? OR customer_phone LIKE ? OR customer_email LIKE ?", sLike, sLike, sLike)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, xerr.NewErrCode(xerr.DatabaseError)
	}

	var orders []model.Order
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Limit(pageSize).Offset(offset).Find(&orders).Error; err != nil {
		return nil, xerr.NewErrCode(xerr.DatabaseError)
	}

	resultList := make([]AdminQueueOrderItem, 0, len(orders))
	for _, o := range orders {
		item := AdminQueueOrderItem{
			ID:                   o.ID,
			OrderID:              o.ID.String(),
			OrderNumber:          o.OrderNumber,
			CustomerPhone:        o.CustomerPhone,
			CustomerEmail:        o.CustomerEmail,
			Status:               o.Status,
			DeliveryType:         o.DeliveryType,
			TotalAmountToman:     o.TotalAmountToman,
			LoyaltyDiscountToman: o.LoyaltyDiscountToman,
			CustomerTier:         o.CustomerTier,
			PriorityScore:        o.PriorityScore,
			PaymentMethod:        o.PaymentMethod,
			TwoFactorPrompt:      o.TwoFactorPrompt,
			TwoFactorCode:        o.TwoFactorCode,
			TwoFactorStatus:      o.TwoFactorStatus,
			TwoFactorRequestedAt: o.TwoFactorRequestedAt,
			CreatedAt:            o.CreatedAt,
			UpdatedAt:            o.UpdatedAt,
		}

		if len(o.Items) > 0 {
			item.ProductTitle = o.Items[0].ProductTitle
		}

		if o.CustomerCredentialsEncrypted != "" && l.svcCtx.Vault != nil {
			dec, decErr := l.svcCtx.Vault.Decrypt(o.CustomerCredentialsEncrypted)
			if decErr == nil && dec != "" {
				item.CustomerCredentials = dec
				parts := strings.Split(dec, " | ")
				for _, part := range parts {
					if strings.HasPrefix(part, "Email: ") {
						item.AccountEmail = strings.TrimPrefix(part, "Email: ")
					} else if strings.HasPrefix(part, "Pass: ") {
						item.AccountPassword = strings.TrimPrefix(part, "Pass: ")
					}
				}
			}
		}

		resultList = append(resultList, item)
	}

	return &AdminOrderListResult{
		Orders:   resultList,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// UpdateOrderStatus modifies order status and optionally triggers automated wallet refunds
func (l *OrderDeskLogic) UpdateOrderStatus(orderIDStr string, req *types.AdminUpdateOrderStatusRequest, adminUsername string) error {
	var order model.Order
	orderUUID, err := uuid.Parse(orderIDStr)
	if err == nil {
		if err := l.svcCtx.DB.WithContext(l.ctx).Where("id = ?", orderUUID).First(&order).Error; err != nil {
			return xerr.NewErrCode(xerr.OrderNotFound)
		}
	} else {
		if err := l.svcCtx.DB.WithContext(l.ctx).Where("order_number = ?", orderIDStr).First(&order).Error; err != nil {
			return xerr.NewErrCode(xerr.OrderNotFound)
		}
	}

	oldStatus := order.Status
	newStatus := req.Status
	now := time.Now()

	order.Status = newStatus
	if req.OperatorNotes != "" {
		order.OperatorNotes = req.OperatorNotes
	}

	if newStatus == model.OrderCompleted {
		order.CompletedAt = &now
		order.CustomerCredentialsEncrypted = "" // Zeroize
	} else if newStatus == model.OrderPaid && order.PaidAt == nil {
		order.PaidAt = &now
	}

	// Refund to wallet handling if requested
	if req.RefundToWallet && (newStatus == model.OrderCancelled || newStatus == model.OrderRefunded) {
		var customer model.Customer
		foundCust := false
		if order.CustomerID != nil {
			if err := l.svcCtx.DB.WithContext(l.ctx).Where("id = ?", *order.CustomerID).First(&customer).Error; err == nil {
				foundCust = true
			}
		}
		if !foundCust && order.CustomerPhone != "" {
			if err := l.svcCtx.DB.WithContext(l.ctx).Where("phone = ?", order.CustomerPhone).First(&customer).Error; err == nil {
				foundCust = true
			}
		}

		if foundCust {
			customer.BalanceToman += order.TotalAmountToman
			_ = l.svcCtx.DB.WithContext(l.ctx).Save(&customer)

			refundTx := model.Transaction{
				ID:                uuid.New(),
				TransactionNumber: fmt.Sprintf("TX-REF-%d-%s", time.Now().UnixNano(), uuid.New().String()[:6]),
				CustomerID:        &customer.ID,
				OrderID:           &order.ID,
				CustomerPhone:     customer.Phone,
				CustomerEmail:     customer.Email,
				AmountToman:       order.TotalAmountToman,
				OrderType:         model.TxTypeRefund,
				PaymentPlatform:   model.PlatformBalance,
				Status:            model.TxStatusSuccess,
				Description:       fmt.Sprintf("استرداد وجه سفارش %s به کیف پول توسط مدیر %s", order.OrderNumber, adminUsername),
			}
			_ = l.svcCtx.DB.WithContext(l.ctx).Create(&refundTx)
		}
	}

	if err := l.svcCtx.DB.WithContext(l.ctx).Save(&order).Error; err != nil {
		return xerr.NewErrCode(xerr.DatabaseError)
	}

	auditLog := model.AdminActionLog{
		AdminUsername: adminUsername,
		Action:        "update_order_status",
		Resource:      "order",
		ResourceID:    order.ID.String(),
		Description:   fmt.Sprintf("تغییر وضعیت سفارش %s از %s به %s توسط مدیر %s (استرداد به کیف: %v)", order.OrderNumber, oldStatus, newStatus, adminUsername, req.RefundToWallet),
	}
	_ = l.svcCtx.DB.WithContext(l.ctx).Create(&auditLog)

	return nil
}
