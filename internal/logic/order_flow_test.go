package logic_test

import (
	"context"
	"testing"

	"github.com/EbadiDev/anahix-server/internal/config"
	"github.com/EbadiDev/anahix-server/internal/logic/admin"
	"github.com/EbadiDev/anahix-server/internal/logic/public"
	"github.com/EbadiDev/anahix-server/internal/model"
	"github.com/EbadiDev/anahix-server/internal/svc"
	"github.com/EbadiDev/anahix-server/internal/types"
	"github.com/EbadiDev/anahix-server/pkg/orm"
	"github.com/EbadiDev/anahix-server/pkg/xerr"
	"github.com/google/uuid"
)

func setupTestContext(t *testing.T) *svc.ServiceContext {
	db, err := orm.InitTestDB()
	if err != nil {
		t.Fatalf("Failed to init test DB: %v", err)
	}

	cfg := &config.Config{
		Vault: config.VaultConfig{
			MasterKey: "test-vault-secret-key-32-bytes-long!",
		},
	}

	svcCtx, err := svc.NewServiceContext(cfg, db)
	if err != nil {
		t.Fatalf("Failed to init ServiceContext: %v", err)
	}

	return svcCtx
}

func TestInstantLinkOrderAndAtomicStockAllocation(t *testing.T) {
	svcCtx := setupTestContext(t)
	ctx := context.Background()

	// 1. Create Instant Delivery Product (Gemini 18m)
	prodID := uuid.New()
	geminiProduct := model.Product{
		ID:           prodID,
		Title:        "Gemini Advanced 18 Months",
		TitleFA:      "اکانت جمینای ۱۸ ماهه",
		Slug:         "gemini-18m-test",
		Category:     model.CategoryAILink,
		DeliveryType: model.DeliveryInstant,
		PriceToman:   1450000,
		IsActive:     true,
	}
	if err := svcCtx.DB.Create(&geminiProduct).Error; err != nil {
		t.Fatalf("Failed to create product: %v", err)
	}

	// 2. Bulk import 2 activation links via Admin InventoryLogic
	adminInvLogic := admin.NewInventoryLogic(ctx, svcCtx)
	importResp, err := adminInvLogic.BulkImport(&types.BulkImportInventoryRequest{
		ProductID: prodID.String(),
		ItemType:  "activation_url",
		Items: []string{
			"https://gemini.google.com/redeem?code=GEMINI_PROMO_LINK_001",
			"https://gemini.google.com/redeem?code=GEMINI_PROMO_LINK_002",
		},
		BatchTag: "BATCH_TEST_1",
	})
	if err != nil {
		t.Fatalf("Failed to bulk import items: %v", err)
	}
	if importResp.ImportedCount != 2 {
		t.Fatalf("Expected 2 items imported, got %d", importResp.ImportedCount)
	}

	// Verify links in DB are ENCRYPTED and not plain text
	var itemsInDB []model.InventoryItem
	svcCtx.DB.Where("product_id = ?", prodID).Find(&itemsInDB)
	if len(itemsInDB) != 2 {
		t.Fatalf("Expected 2 items in DB, got %d", len(itemsInDB))
	}
	for _, item := range itemsInDB {
		if item.PayloadEncrypted == "https://gemini.google.com/redeem?code=GEMINI_PROMO_LINK_001" {
			t.Fatal("SECURITY ERROR: Payload was stored in plain text!")
		}
	}

	// 3. Customer 1 places order -> Stock allocated
	publicOrderLogic := public.NewOrderLogic(ctx, svcCtx)
	order1, err := publicOrderLogic.CreateOrder(&types.CreateOrderRequest{
		ProductID:     prodID.String(),
		CustomerPhone: "09121111111",
		PaymentMethod: "zarinpal",
	})
	if err != nil {
		t.Fatalf("Customer 1 order failed: %v", err)
	}
	if order1.TotalAmountToman != 1450000 {
		t.Fatalf("Expected price 1450000, got %d", order1.TotalAmountToman)
	}

	// 4. Customer 2 places order -> Stock allocated
	order2, err := publicOrderLogic.CreateOrder(&types.CreateOrderRequest{
		ProductID:     prodID.String(),
		CustomerPhone: "09122222222",
		PaymentMethod: "zarinpal",
	})
	if err != nil {
		t.Fatalf("Customer 2 order failed: %v", err)
	}
	if order2.OrderID == "" {
		t.Fatal("Expected valid order ID for customer 2")
	}

	// 5. Customer 3 places order -> MUST FAIL with OutOfStock!
	_, err = publicOrderLogic.CreateOrder(&types.CreateOrderRequest{
		ProductID:     prodID.String(),
		CustomerPhone: "09123333333",
		PaymentMethod: "zarinpal",
	})
	if err == nil {
		t.Fatal("Expected OutOfStock error for Customer 3, but order succeeded!")
	}
	if xerrCode, ok := err.(*xerr.CodeError); ok {
		if xerrCode.GetErrCode() != xerr.OutOfStock {
			t.Fatalf("Expected OutOfStock code (%d), got %d", xerr.OutOfStock, xerrCode.GetErrCode())
		}
	} else {
		t.Fatalf("Expected *xerr.CodeError, got %T: %v", err, err)
	}

	// 6. Simulate payment for Order 1 and verify decrypted activation link is delivered
	svcCtx.DB.Model(&model.Order{}).
		Where("id = ?", order1.OrderID).
		Updates(map[string]interface{}{
			"status": model.OrderPaid,
		})

	detail1, err := publicOrderLogic.GetOrderDetail(order1.OrderID, order1.AccessToken)
	if err != nil {
		t.Fatalf("Failed to get order detail: %v", err)
	}
	if len(detail1.DeliveredItems) != 1 {
		t.Fatalf("Expected 1 delivered item, got %d", len(detail1.DeliveredItems))
	}
	deliveredLink := detail1.DeliveredItems[0].Payload
	if deliveredLink != "https://gemini.google.com/redeem?code=GEMINI_PROMO_LINK_001" &&
		deliveredLink != "https://gemini.google.com/redeem?code=GEMINI_PROMO_LINK_002" {
		t.Fatalf("Invalid decrypted link: %s", deliveredLink)
	}
}

func TestModeCOnAccountUpgradeAndTwoFactorHandshake(t *testing.T) {
	svcCtx := setupTestContext(t)
	ctx := context.Background()

	// 1. Create Timed Personal Upgrade Product (ChatGPT Plus)
	prodID := uuid.New()
	chatgptProduct := model.Product{
		ID:             prodID,
		Title:          "ChatGPT Plus Personal Upgrade",
		TitleFA:        "فعال‌سازی چت‌جی‌پی‌تی پلاس روی اکانت شخصی",
		Slug:           "chatgpt-plus-personal-test",
		Category:       model.CategoryPersonalUpgrade,
		DeliveryType:   model.DeliveryManualTimed,
		SLAMinMinutes:  30,
		SLAMaxMinutes:  180,
		PriceToman:     1950000,
		IsActive:       true,
		OperatingHours: "09:00 - 23:00",
	}
	svcCtx.DB.Create(&chatgptProduct)

	// 2. Customer creates order with their private account credentials
	publicOrderLogic := public.NewOrderLogic(ctx, svcCtx)
	orderResp, err := publicOrderLogic.CreateOrder(&types.CreateOrderRequest{
		ProductID:               prodID.String(),
		CustomerPhone:           "09351234567",
		CustomerEmail:           "customer@gmail.com",
		PaymentMethod:           "zarinpal",
		CustomerAccountEmail:    "customer@gmail.com",
		CustomerAccountPassword: "SuperSecretPassword123!",
		CustomerNotes:           "لطفا پلن سالانه فعال نشود، ماهانه باشد",
	})
	if err != nil {
		t.Fatalf("Failed to create order: %v", err)
	}

	// 3. Operator views order in Admin Desk and inspects decrypted credentials
	adminDesk := admin.NewOrderDeskLogic(ctx, svcCtx)
	adminOrder, err := adminDesk.GetOrderDetail(orderResp.OrderID)
	if err != nil {
		t.Fatalf("Admin failed to view order: %v", err)
	}
	if adminOrder.CustomerCredentials == "" {
		t.Fatal("Admin expected decrypted customer credentials, but was empty")
	}

	// 4. Operator requests 2FA code from customer
	err = adminDesk.RequestTwoFactor(orderResp.OrderID, &types.RequestTwoFactorRequest{
		PromptMessage: "لطفاً کد ۶ رقمی ارسالی به ایمیل خود را وارد نمایید",
	})
	if err != nil {
		t.Fatalf("Failed to request 2FA: %v", err)
	}

	// Customer checks order tracking view -> sees TwoFactorRequested
	customerView, err := publicOrderLogic.GetOrderDetail(orderResp.OrderID, orderResp.AccessToken)
	if err != nil {
		t.Fatalf("Customer failed to view order: %v", err)
	}
	if customerView.TwoFactorStatus != string(model.TwoFactorRequested) {
		t.Fatalf("Expected 2FA status 'requested', got %s", customerView.TwoFactorStatus)
	}

	// 5. Customer submits 2FA OTP code
	err = publicOrderLogic.SubmitTwoFactor(orderResp.OrderID, orderResp.AccessToken, "654321")
	if err != nil {
		t.Fatalf("Customer failed to submit 2FA: %v", err)
	}

	// Admin checks order -> sees submitted OTP
	updatedAdminOrder, _ := adminDesk.GetOrderDetail(orderResp.OrderID)
	if updatedAdminOrder.Order.TwoFactorCode != "654321" {
		t.Fatalf("Expected 2FA code '654321', got '%s'", updatedAdminOrder.Order.TwoFactorCode)
	}

	// 6. Admin completes order -> Customer credentials MUST BE ZEROIZED for privacy
	err = adminDesk.CompleteOrder(orderResp.OrderID, &types.CompleteOrderRequest{
		OperatorNotes: "با مسترکارت ترکیه فعال شد",
	})
	if err != nil {
		t.Fatalf("Admin failed to complete order: %v", err)
	}

	// Verify credentials are now zeroized in DB
	var finalOrder model.Order
	svcCtx.DB.Where("id = ?", orderResp.OrderID).First(&finalOrder)
	if finalOrder.CustomerCredentialsEncrypted != "" {
		t.Fatal("SECURITY ERROR: Customer credentials were not purged after order completion!")
	}
	if finalOrder.Status != model.OrderCompleted {
		t.Fatalf("Expected OrderCompleted, got %s", finalOrder.Status)
	}
}
