package admin

import (
	"context"
	"testing"
	"time"

	"github.com/EbadiDev/anahix-server/internal/config"
	"github.com/EbadiDev/anahix-server/internal/logic/provider/smsir"
	"github.com/EbadiDev/anahix-server/internal/model"
	"github.com/EbadiDev/anahix-server/internal/svc"
	"github.com/EbadiDev/anahix-server/internal/types"
	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func setupTestCustomerAdminDB(t *testing.T) (*svc.ServiceContext, model.Customer) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite memory db: %v", err)
	}

	err = db.AutoMigrate(
		&model.Customer{},
		&model.Order{},
		&model.OrderItem{},
		&model.VirtualNumberOrder{},
		&model.Transaction{},
		&model.Referral{},
		&model.ReferralCommission{},
		&model.SystemSetting{},
		&model.AdminActionLog{},
	)
	if err != nil {
		t.Fatalf("failed to auto migrate tables: %v", err)
	}

	cfg := &config.Config{
		JWT: config.JWTConfig{
			AccessSecret:      "test-secret-key-1234567890",
			AccessExpireHours: 720,
		},
	}

	svcCtx := &svc.ServiceContext{
		DB:          db,
		Config:      cfg,
		SMSIRClient: smsir.NewClient(),
	}

	// Seed customer
	customerID := uuid.New()
	now := time.Now()
	cust := model.Customer{
		ID:                   customerID,
		Phone:                "09121112233",
		FullName:             "علی علوی",
		AvatarURL:            "https://example.com/avatar.png",
		Email:                "ali@example.com",
		BalanceToman:         100000,
		ReferralCode:         "ANX-TEST",
		ReferralBalanceToman: 25000,
		Tier:                 "vip",
		IsActive:             true,
		LastLoginAt:          &now,
		CreatedAt:            now,
	}
	if err := db.Create(&cust).Error; err != nil {
		t.Fatalf("failed to seed customer: %v", err)
	}

	// Seed an order
	order := model.Order{
		ID:               uuid.New(),
		OrderNumber:      "ORD-2026-9901",
		CustomerID:       &cust.ID,
		CustomerPhone:    cust.Phone,
		Status:           model.OrderPaid,
		DeliveryType:     "activation_code",
		TotalAmountToman: 450000,
		PaymentMethod:    model.PaymentMethodBalance,
		CreatedAt:        now,
		Items: []model.OrderItem{
			{
				ID:               uuid.New(),
				ProductTitle:     "اشتراک ChatGPT Plus",
				ProductSlug:      "chatgpt-plus",
				DeliveryType:     "activation_code",
				UnitPriceToman:   450000,
				Quantity:         1,
				DeliveredPayload: "KEY-1234-ABCD",
			},
		},
	}
	db.Create(&order)

	// Seed a virtual number order
	vnum := model.VirtualNumberOrder{
		ID:            uuid.New(),
		OrderNo:       "VN-9901",
		Phone:         "+447891234567",
		Country:       "england",
		CountryNameFA: "انگلستان",
		Service:       "telegram",
		ServiceNameFA: "تلگرام",
		PriceToman:    65000,
		Status:        model.VNumStatusReceived,
		SMSCode:       "54321",
		CustomerPhone: cust.Phone,
		CreatedAt:     now,
	}
	db.Create(&vnum)

	return svcCtx, cust
}

func TestCustomerAdminLogic_ListAnd360(t *testing.T) {
	ctx := context.Background()
	svcCtx, cust := setupTestCustomerAdminDB(t)
	logic := NewCustomerAdminLogic(ctx, svcCtx)

	// 1. List Customers with search
	listResp, err := logic.ListCustomers(&types.AdminListCustomersRequest{
		Search:   "0912111",
		Status:   "active",
		Tier:     "vip",
		Page:     1,
		PageSize: 10,
	})
	if err != nil {
		t.Fatalf("ListCustomers failed: %v", err)
	}
	if listResp.Total != 1 {
		t.Fatalf("expected total 1, got %d", listResp.Total)
	}
	if listResp.Customers[0].Phone != cust.Phone {
		t.Fatalf("expected phone %s, got %s", cust.Phone, listResp.Customers[0].Phone)
	}
	if listResp.Customers[0].AvatarURL != "https://example.com/avatar.png" {
		t.Fatalf("expected avatar_url https://example.com/avatar.png, got %s", listResp.Customers[0].AvatarURL)
	}
	if listResp.Customers[0].Tier != "vip" {
		t.Fatalf("expected tier vip, got %s", listResp.Customers[0].Tier)
	}
	if !listResp.Customers[0].IsOnline {
		t.Fatalf("expected is_online to be true")
	}

	// 2. Get Customer 360 Detail
	detail360, err := logic.GetCustomer360(cust.ID.String())
	if err != nil {
		t.Fatalf("GetCustomer360 failed: %v", err)
	}
	if detail360.Customer.Phone != cust.Phone {
		t.Fatalf("expected phone %s, got %s", cust.Phone, detail360.Customer.Phone)
	}
	if detail360.Customer.AvatarURL != "https://example.com/avatar.png" {
		t.Fatalf("expected 360 avatar_url https://example.com/avatar.png, got %s", detail360.Customer.AvatarURL)
	}
	if len(detail360.Orders) != 1 {
		t.Fatalf("expected 1 order, got %d", len(detail360.Orders))
	}
	if len(detail360.VirtualNumbers) != 1 {
		t.Fatalf("expected 1 virtual number, got %d", len(detail360.VirtualNumbers))
	}
	if detail360.VirtualNumbers[0].SMSCode != "54321" {
		t.Fatalf("expected SMSCode 54321, got %s", detail360.VirtualNumbers[0].SMSCode)
	}
}

func TestCustomerAdminLogic_AdjustWallet(t *testing.T) {
	ctx := context.Background()
	svcCtx, cust := setupTestCustomerAdminDB(t)
	logic := NewCustomerAdminLogic(ctx, svcCtx)

	// 1. Credit wallet
	creditReq := &types.AdminAdjustCustomerWalletRequest{
		AmountToman: 50000,
		Action:      "credit",
		Reason:      "شارژ بابت واریز کارت به کارت فیش ۱۲۳",
	}
	updatedCust, err := logic.AdjustWallet(cust.ID.String(), creditReq, "admin_test")
	if err != nil {
		t.Fatalf("AdjustWallet credit failed: %v", err)
	}
	if updatedCust.BalanceToman != 150000 {
		t.Fatalf("expected balance 150000, got %d", updatedCust.BalanceToman)
	}

	// Verify transaction record
	var tx model.Transaction
	if err := svcCtx.DB.Where("customer_phone = ? AND order_type = ?", cust.Phone, model.TxTypeAdminAdjustment).First(&tx).Error; err != nil {
		t.Fatalf("failed to find adjustment transaction: %v", err)
	}
	if tx.AmountToman != 50000 {
		t.Fatalf("expected tx amount 50000, got %d", tx.AmountToman)
	}

	// 2. Debit wallet
	debitReq := &types.AdminAdjustCustomerWalletRequest{
		AmountToman: 30000,
		Action:      "debit",
		Reason:      "کسر بابت سفارش حضوری",
	}
	updatedCust2, err := logic.AdjustWallet(cust.ID.String(), debitReq, "admin_test")
	if err != nil {
		t.Fatalf("AdjustWallet debit failed: %v", err)
	}
	if updatedCust2.BalanceToman != 120000 {
		t.Fatalf("expected balance 120000, got %d", updatedCust2.BalanceToman)
	}

	// 3. Debit more than available (overdraft protection check)
	overdraftReq := &types.AdminAdjustCustomerWalletRequest{
		AmountToman: 999999,
		Action:      "debit",
		Reason:      "کسر بیش از موجودی",
	}
	_, err = logic.AdjustWallet(cust.ID.String(), overdraftReq, "admin_test")
	if err == nil {
		t.Fatalf("expected overdraft error, got nil")
	}
}

func TestCustomerAdminLogic_UpdateCustomerAndReplaceKey(t *testing.T) {
	ctx := context.Background()
	svcCtx, cust := setupTestCustomerAdminDB(t)
	logic := NewCustomerAdminLogic(ctx, svcCtx)

	// 1. Update Customer status & notes
	isBlocked := false
	blockReason := "تلاش برای فعالیت مشکوک"
	tier := "colleague"
	notes := "مشتری قدیمی همکار"
	updateReq := &types.AdminUpdateCustomerStatusRequest{
		IsActive:      &isBlocked,
		BlockReason:   &blockReason,
		Tier:          &tier,
		InternalNotes: &notes,
	}

	updated, err := logic.UpdateCustomer(cust.ID.String(), updateReq, "admin_test")
	if err != nil {
		t.Fatalf("UpdateCustomer failed: %v", err)
	}
	if updated.IsActive != false {
		t.Fatalf("expected is_active to be false")
	}
	if updated.Tier != "colleague" {
		t.Fatalf("expected tier colleague, got %s", updated.Tier)
	}
	if updated.InternalNotes != notes {
		t.Fatalf("expected internal notes %s, got %s", notes, updated.InternalNotes)
	}

	// 2. Replace Order Key
	var item model.OrderItem
	if err := svcCtx.DB.First(&item).Error; err != nil {
		t.Fatalf("failed to find order item: %v", err)
	}

	repReq := &types.AdminReplaceCustomerKeyRequest{
		OrderItemID: item.ID.String(),
		NewPayload:  "NEW-KEY-REPLACED-9999",
		Reason:      "تمدید اشتراک",
	}
	err = logic.ReplaceOrderKey(cust.ID.String(), repReq, "admin_test")
	if err != nil {
		t.Fatalf("ReplaceOrderKey failed: %v", err)
	}

	var updatedItem model.OrderItem
	svcCtx.DB.Where("id = ?", item.ID).First(&updatedItem)
	if updatedItem.DeliveredPayload != "NEW-KEY-REPLACED-9999" {
		t.Fatalf("expected replaced key NEW-KEY-REPLACED-9999, got %s", updatedItem.DeliveredPayload)
	}

	// 3. Direct SMS send
	smsReq := &types.AdminSendCustomerSMSRequest{
		Message: "کلید جدید اکانت شما فعال گردید",
	}
	err = logic.SendDirectSMS(cust.ID.String(), smsReq, "admin_test")
	if err != nil {
		t.Fatalf("SendDirectSMS failed: %v", err)
	}
}

func TestOrderDeskLogic_ListAllAndRefund(t *testing.T) {
	ctx := context.Background()
	svcCtx, cust := setupTestCustomerAdminDB(t)
	orderLogic := NewOrderDeskLogic(ctx, svcCtx)

	// 1. List All Orders
	res, err := orderLogic.ListAllOrders("all", "all", "all", "", 1, 10)
	if err != nil {
		t.Fatalf("ListAllOrders failed: %v", err)
	}
	if res.Total != 1 {
		t.Fatalf("expected 1 order, got %d", res.Total)
	}

	// 2. Update Order Status to Refunded with RefundToWallet = true
	var order model.Order
	svcCtx.DB.First(&order)

	prevBalance := cust.BalanceToman // 100,000
	orderAmount := order.TotalAmountToman // 450,000

	updateReq := &types.AdminUpdateOrderStatusRequest{
		Status:         model.OrderRefunded,
		OperatorNotes:  "مرجوعی به دلیل عدم رضایت خریدار",
		RefundToWallet: true,
	}

	err = orderLogic.UpdateOrderStatus(order.ID.String(), updateReq, "admin_test")
	if err != nil {
		t.Fatalf("UpdateOrderStatus failed: %v", err)
	}

	// Verify customer balance increased by order amount
	var refreshedCust model.Customer
	svcCtx.DB.Where("id = ?", cust.ID).First(&refreshedCust)
	expectedBalance := prevBalance + orderAmount
	if refreshedCust.BalanceToman != expectedBalance {
		t.Fatalf("expected balance %d after refund, got %d", expectedBalance, refreshedCust.BalanceToman)
	}

	// Verify refund transaction
	var refundTx model.Transaction
	if err := svcCtx.DB.Where("order_id = ? AND order_type = ?", order.ID, model.TxTypeRefund).First(&refundTx).Error; err != nil {
		t.Fatalf("failed to find refund transaction: %v", err)
	}
	if refundTx.AmountToman != orderAmount {
		t.Fatalf("expected refund amount %d, got %d", orderAmount, refundTx.AmountToman)
	}
}
