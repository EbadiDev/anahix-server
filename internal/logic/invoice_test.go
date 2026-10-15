package logic

import (
	"context"
	"testing"
	"time"

	"github.com/EbadiDev/anahix-server/internal/config"
	"github.com/EbadiDev/anahix-server/internal/logic/admin"
	"github.com/EbadiDev/anahix-server/internal/logic/public"
	"github.com/EbadiDev/anahix-server/internal/model"
	"github.com/EbadiDev/anahix-server/internal/svc"
	"github.com/EbadiDev/anahix-server/internal/types"
	"github.com/EbadiDev/anahix-server/pkg/orm"
	"github.com/google/uuid"
)

func TestOfficialInvoiceAndBillingLogic(t *testing.T) {
	db, err := orm.InitTestDB()
	if err != nil {
		t.Fatalf("Failed to init test db: %v", err)
	}

	cfg := config.Config{
		JWT: config.JWTConfig{AccessSecret: "test-secret-key-1234567890"},
	}
	svcCtx := &svc.ServiceContext{
		DB:     db,
		Config: &cfg,
	}

	ctx := context.Background()

	// 1. Seed Billing Settings
	billingAdmin := admin.NewBillingSettingsLogic(ctx, svcCtx)
	companyName := "شرکت فناوری اطلاعات آناهیکس آزمایشی"
	vatRate := 10
	vatEnabled := true
	updateErr := billingAdmin.UpdateBillingSettings(&types.UpdateBillingSettingsRequest{
		SellerCompanyName: &companyName,
		VATPercent:        &vatRate,
		VATEnabled:        &vatEnabled,
	})
	if updateErr != nil {
		t.Fatalf("Failed to update billing settings: %v", updateErr)
	}

	loadedSettings, getErr := billingAdmin.GetBillingSettings()
	if getErr != nil {
		t.Fatalf("Failed to get billing settings: %v", getErr)
	}
	if loadedSettings.SellerCompanyName != companyName {
		t.Errorf("Expected seller company '%s', got '%s'", companyName, loadedSettings.SellerCompanyName)
	}
	if loadedSettings.VATPercent != 10 {
		t.Errorf("Expected VAT 10, got %d", loadedSettings.VATPercent)
	}

	// 2. Create sample product and order
	prodID := uuid.New()
	prod := model.Product{
		ID:           prodID,
		TitleFA:      "اشتراک پریمیوم چت‌جی‌پی‌تی ۴ پلاس",
		Slug:         "chatgpt-plus",
		Category:     "ai_chat",
		DeliveryType: model.DeliveryInstant,
		PriceToman:   1000000, // 1,000,000 Toman
		IsActive:     true,
	}
	db.Create(&prod)

	orderID := uuid.New()
	orderNo := "ANX-261015-TEST"
	orderToken := "secure-token-123"
	now := time.Now()

	order := model.Order{
		ID:                  orderID,
		OrderNumber:         orderNo,
		CustomerPhone:       "09120001122",
		CustomerEmail:       "client@test.ir",
		AccessToken:         orderToken,
		Status:              model.OrderPaid,
		DeliveryType:        model.DeliveryInstant,
		SubtotalAmountToman: 1000000,
		DiscountAmountToman: 100000, // 100,000 discount -> after discount = 900,000
		TotalAmountToman:    900000,
		PaymentMethod:       model.PaymentMethodZarinpal,
		PaymentRefID:        "SHP-987654321",
		PaidAt:              &now,
		IsCorporate:         false,
		Items: []model.OrderItem{
			{
				ID:             uuid.New(),
				OrderID:        orderID,
				ProductID:      prodID,
				ProductTitle:   prod.TitleFA,
				ProductSlug:    prod.Slug,
				DeliveryType:   prod.DeliveryType,
				UnitPriceToman: 1000000,
				Quantity:       1,
				SubtotalToman:  1000000,
			},
		},
	}
	db.Create(&order)

	// 3. Retrieve Official Invoice
	invoiceLogic := public.NewInvoiceLogic(ctx, svcCtx)
	inv, invErr := invoiceLogic.GetOrderInvoice(orderNo, orderToken, nil, false)
	if invErr != nil {
		t.Fatalf("Failed to retrieve invoice: %v", invErr)
	}

	if inv.InvoiceType != "official" {
		t.Errorf("Expected official invoice, got %s", inv.InvoiceType)
	}
	if inv.Seller.CompanyName != companyName {
		t.Errorf("Expected seller '%s', got '%s'", companyName, inv.Seller.CompanyName)
	}
	if inv.SubtotalToman != 1000000 {
		t.Errorf("Expected subtotal 1000000, got %d", inv.SubtotalToman)
	}
	if inv.TotalDiscountToman != 100000 {
		t.Errorf("Expected discount 100000, got %d", inv.TotalDiscountToman)
	}
	if inv.AmountAfterDiscountToman != 900000 {
		t.Errorf("Expected amount after discount 900000, got %d", inv.AmountAfterDiscountToman)
	}
	if inv.TotalTaxToman != 0 {
		t.Errorf("Expected VAT 0, got %d", inv.TotalTaxToman)
	}
	if inv.FinalTotalToman != 900000 {
		t.Errorf("Expected final total 900000, got %d", inv.FinalTotalToman)
	}
	if inv.FinalTotalTomanWords == "" {
		t.Errorf("Expected non-empty Persian amount in words")
	}

	// 4. Verify clean buyer and seller details (privacy-safe, no credentials)
	if inv.Buyer.FullName == "" {
		t.Errorf("Expected non-empty buyer full name")
	}
	if inv.Buyer.Phone != "09120001122" {
		t.Errorf("Expected buyer phone '09120001122', got '%s'", inv.Buyer.Phone)
	}
	if inv.Seller.BrandFA == "" {
		t.Errorf("Expected non-empty seller brand name")
	}
	if len(inv.Items) != 1 || inv.Items[0].ProductTitle != prod.TitleFA {
		t.Errorf("Expected item title '%s', got '%s'", prod.TitleFA, inv.Items[0].ProductTitle)
	}
}
