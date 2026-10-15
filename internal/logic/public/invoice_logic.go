package public

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/EbadiDev/anahix-server/internal/model"
	"github.com/EbadiDev/anahix-server/internal/svc"
	"github.com/EbadiDev/anahix-server/internal/types"
	"github.com/EbadiDev/anahix-server/pkg/jalali"
	"github.com/EbadiDev/anahix-server/pkg/numtofa"
	"github.com/EbadiDev/anahix-server/pkg/xerr"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type InvoiceLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewInvoiceLogic(ctx context.Context, svcCtx *svc.ServiceContext) *InvoiceLogic {
	return &InvoiceLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// GetOrderInvoice builds an official, tax-compliant commercial invoice or proforma
func (l *InvoiceLogic) GetOrderInvoice(orderIDOrNo string, token string, currentCustomerID *uuid.UUID, isAdmin bool) (*types.OfficialInvoiceResponse, error) {
	var order model.Order
	query := l.svcCtx.DB.WithContext(l.ctx).Preload("Items")

	if orderUUID, err := uuid.Parse(orderIDOrNo); err == nil {
		query = query.Where("id = ?", orderUUID)
	} else {
		query = query.Where("order_number = ?", orderIDOrNo)
	}

	if err := query.First(&order).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, xerr.NewErrCode(xerr.OrderNotFound)
		}
		return nil, xerr.NewErrCode(xerr.DatabaseError)
	}

	// Permission & Verification:
	tokenMatches := token != "" && order.AccessToken == token
	isOwner := false
	if currentCustomerID != nil && order.CustomerID != nil && *order.CustomerID == *currentCustomerID {
		isOwner = true
	}

	// 1. Load Seller Billing Configuration from system table (category: "billing")
	seller := types.SellerBillingInfo{
		CompanyName:       "فروشگاه هوش مصنوعی آناهیکس",
		BrandFA:           "فروشگاه هوش مصنوعی آناهیکس",
		NationalID:        "",
		EconomicCode:      "",
		RegistrationNo:    "",
		PostalCode:        "",
		Province:          "",
		City:              "",
		Address:           "",
		Phone:             "021-88776655",
		Email:             "support@anahix.com",
		Website:           "https://anahix.com",
		TaxMemoryID:       "",
		StampSignatureURL: "",
	}

	invoiceNotes := "از خرید و اعتماد شما به آناهیکس صمیمانه متشکریم! پشتیبانی ۲۴ ساعته در کنار شماست."

	var sysSettings []model.SystemSetting
	if err := l.svcCtx.DB.WithContext(l.ctx).Where("category = ?", "billing").Find(&sysSettings).Error; err == nil {
		for _, s := range sysSettings {
			val := strings.TrimSpace(s.Value)
			switch s.Key {
			case "billing_seller_company_name":
				if val != "" && !strings.Contains(val, "نوآوران داده‌پرداز") {
					seller.CompanyName = val
				}
			case "billing_seller_brand_fa":
				if val != "" {
					seller.BrandFA = val
				}
			case "billing_seller_phone":
				if val != "" {
					seller.Phone = val
				}
			case "billing_seller_email":
				if val != "" {
					seller.Email = val
				}
			case "billing_seller_website":
				if val != "" {
					seller.Website = val
				}
			case "billing_invoice_notes":
				if val != "" && !strings.Contains(val, "ماده ۱۹") {
					invoiceNotes = val
				}
			}
		}
	}

	// 2. Prepare Buyer Information (Strict privacy: NEVER exposes Mode C passwords or account upgrade credentials)
	maskedPhone := order.CustomerPhone
	if !tokenMatches && !isAdmin && !isOwner && len(maskedPhone) >= 7 {
		maskedPhone = maskedPhone[:4] + "***" + maskedPhone[len(maskedPhone)-4:]
	}

	buyerFullName := "کاربر گرامی"
	buyerEmail := ""
	if order.CustomerID != nil {
		var cust model.Customer
		if err := l.svcCtx.DB.WithContext(l.ctx).Where("id = ?", *order.CustomerID).First(&cust).Error; err == nil {
			if cust.FullName != "" {
				buyerFullName = cust.FullName
			}
			if cust.Email != "" {
				buyerEmail = cust.Email
			}
		}
	}

	buyer := types.BuyerBillingInfo{
		IsCorporate:    false,
		CompanyName:    "",
		FullName:       buyerFullName,
		NationalID:     "",
		EconomicCode:   "",
		RegistrationNo: "",
		PostalCode:     "",
		Address:        "",
		Phone:          maskedPhone,
		Email:          buyerEmail,
	}

	// 3. Itemize calculation
	var invoiceItems []types.InvoiceItemDTO
	subtotalToman := order.SubtotalAmountToman
	totalDiscountToman := order.DiscountAmountToman
	if subtotalToman <= 0 && len(order.Items) > 0 {
		for _, it := range order.Items {
			subtotalToman += it.UnitPriceToman * int64(it.Quantity)
		}
	}

	for idx, item := range order.Items {
		qty := item.Quantity
		if qty <= 0 {
			qty = 1
		}
		itemSubtotal := item.UnitPriceToman * int64(qty)

		// Proportional discount allocation
		var itemDiscount int64
		if subtotalToman > 0 && totalDiscountToman > 0 {
			itemDiscount = (itemSubtotal * totalDiscountToman) / subtotalToman
		}
		if itemDiscount > itemSubtotal {
			itemDiscount = itemSubtotal
		}

		itemFinal := itemSubtotal - itemDiscount

		deliveryTypeLabel := "تحویل آنلاین و آنی"
		if item.DeliveryType == model.DeliveryManualTimed {
			deliveryTypeLabel = "فعال‌سازی روی اکانت شخصی"
		}

		invoiceItems = append(invoiceItems, types.InvoiceItemDTO{
			RowIndex:                idx + 1,
			ProductID:               item.ProductID.String(),
			ProductTitle:            item.ProductTitle,
			ProductSlug:             item.ProductSlug,
			DeliveryType:            deliveryTypeLabel,
			Quantity:                qty,
			UnitPriceToman:          item.UnitPriceToman,
			SubtotalToman:           itemSubtotal,
			DiscountToman:           itemDiscount,
			TotalAfterDiscountToman: itemFinal,
			TaxRatePercent:          0,
			TaxToman:                0,
			FinalAmountToman:        itemFinal,
		})
	}

	amountAfterDiscount := subtotalToman - totalDiscountToman
	if amountAfterDiscount < 0 {
		amountAfterDiscount = 0
	}
	finalTotalToman := amountAfterDiscount

	// 4. Payment status & invoice type
	isPaid := order.Status == model.OrderPaid || order.Status == model.OrderCompleted
	invoiceType := "proforma"
	invoiceTypeFA := "پیش‌فاکتور سفارش"
	paymentStatus := string(order.Status)
	paymentStatusFA := "در انتظار پرداخت"

	if isPaid {
		invoiceType = "official"
		invoiceTypeFA = "رسید و صورتحساب خرید"
		paymentStatusFA = "پرداخت موفق و تایید شده"
	} else if order.Status == model.OrderCancelled {
		paymentStatusFA = "لغو شده"
	} else if order.Status == model.OrderRefunded {
		paymentStatusFA = "استرداد وجه"
	}

	paymentMethodFA := "پرداخت الکترونیک شاپرک"
	switch order.PaymentMethod {
	case model.PaymentMethodBalance:
		paymentMethodFA = "کسر مستقیم از کیف پول آناهیکس"
	case model.PaymentMethodZarinpal:
		paymentMethodFA = "درگاه پرداخت امن شاپرک (زرین‌پال)"
	case model.PaymentMethodCardToCard:
		paymentMethodFA = "کارت به کارت شتابی با تایید رسید"
	case model.PaymentMethodCrypto:
		paymentMethodFA = "پرداخت بین‌المللی رمزارز (USDT)"
	}

	// Dates in Jalali
	issueDateJalali := jalali.FormatDateTimeJalali(order.CreatedAt)
	var paidAtJalali string
	if order.PaidAt != nil {
		paidAtJalali = jalali.FormatDateTimeJalali(*order.PaidAt)
	}

	// Formatted payment reference / Shaparak tracking code
	shaparakRef := order.PaymentRefID
	if shaparakRef == "" && isPaid {
		shaparakRef = order.PaymentAuthority
	}

	words := numtofa.AmountToPersianWords(finalTotalToman)

	return &types.OfficialInvoiceResponse{
		InvoiceNumber:            fmt.Sprintf("INV-%s", strings.TrimPrefix(order.OrderNumber, "ANX-")),
		OrderNumber:              order.OrderNumber,
		OrderID:                  order.ID.String(),
		InvoiceType:              invoiceType,
		InvoiceTypeFA:            invoiceTypeFA,
		IssueDate:                order.CreatedAt,
		IssueDateJalali:          issueDateJalali,
		PaidAt:                   order.PaidAt,
		PaidAtJalali:             paidAtJalali,
		PaymentMethod:            string(order.PaymentMethod),
		PaymentMethodFA:          paymentMethodFA,
		PaymentRefID:             shaparakRef,
		PaymentAuthority:         order.PaymentAuthority,
		PaymentStatus:            paymentStatus,
		PaymentStatusFA:          paymentStatusFA,
		Seller:                   seller,
		Buyer:                    buyer,
		Items:                    invoiceItems,
		SubtotalToman:            subtotalToman,
		TotalDiscountToman:       totalDiscountToman,
		AmountAfterDiscountToman: amountAfterDiscount,
		TaxRatePercent:           0,
		TotalTaxToman:            0,
		FinalTotalToman:          finalTotalToman,
		FinalTotalTomanWords:     words,
		FinalTotalRials:          finalTotalToman * 10,
		Notes:                    invoiceNotes,
	}, nil
}

// UpdateOrderBuyerInvoice updates the official corporate or individual buyer information for an order
func (l *InvoiceLogic) UpdateOrderBuyerInvoice(orderIDOrNo string, token string, currentCustomerID *uuid.UUID, isAdmin bool, req *types.UpdateBuyerInvoiceRequest) error {
	var order model.Order
	query := l.svcCtx.DB.WithContext(l.ctx)

	if orderUUID, err := uuid.Parse(orderIDOrNo); err == nil {
		query = query.Where("id = ?", orderUUID)
	} else {
		query = query.Where("order_number = ?", orderIDOrNo)
	}

	if err := query.First(&order).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return xerr.NewErrCode(xerr.OrderNotFound)
		}
		return xerr.NewErrCode(xerr.DatabaseError)
	}

	// Authorization
	tokenMatches := token != "" && order.AccessToken == token
	isOwner := currentCustomerID != nil && order.CustomerID != nil && *order.CustomerID == *currentCustomerID

	if !isAdmin && !tokenMatches && !isOwner {
		return xerr.NewErrMsg("دسترسی غیرمجاز برای ویرایش فاکتور رسمی این سفارش")
	}

	updates := map[string]interface{}{
		"is_corporate":    req.IsCorporate,
		"company_name":    strings.TrimSpace(req.CompanyName),
		"national_id":     strings.TrimSpace(req.NationalID),
		"economic_code":   strings.TrimSpace(req.EconomicCode),
		"registration_no": strings.TrimSpace(req.RegistrationNo),
		"postal_code":     strings.TrimSpace(req.PostalCode),
		"buyer_address":   strings.TrimSpace(req.Address),
		"buyer_phone":     strings.TrimSpace(req.Phone),
		"updated_at":      time.Now(),
	}

	if err := l.svcCtx.DB.WithContext(l.ctx).Model(&order).Updates(updates).Error; err != nil {
		return xerr.NewErrCode(xerr.DatabaseError)
	}

	// Also if owner is logged in and is updating corporate info, optionally update Customer profile
	if isOwner && currentCustomerID != nil && req.IsCorporate {
		l.svcCtx.DB.WithContext(l.ctx).Model(&model.Customer{}).Where("id = ?", *currentCustomerID).Updates(map[string]interface{}{
			"is_corporate":    true,
			"company_name":    strings.TrimSpace(req.CompanyName),
			"national_id":     strings.TrimSpace(req.NationalID),
			"economic_code":   strings.TrimSpace(req.EconomicCode),
			"registration_no": strings.TrimSpace(req.RegistrationNo),
			"postal_code":     strings.TrimSpace(req.PostalCode),
			"company_address": strings.TrimSpace(req.Address),
			"company_phone":   strings.TrimSpace(req.Phone),
		})
	}

	return nil
}
