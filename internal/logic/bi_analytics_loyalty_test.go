package logic_test

import (
	"context"
	"testing"
	"time"

	"github.com/EbadiDev/anahix-server/internal/logic/admin"
	"github.com/EbadiDev/anahix-server/internal/logic/loyalty"
	"github.com/EbadiDev/anahix-server/internal/logic/public"
	"github.com/EbadiDev/anahix-server/internal/model"
	"github.com/EbadiDev/anahix-server/internal/types"
	"github.com/google/uuid"
)

func TestBIAnalyticsAndLoyaltySystem(t *testing.T) {
	svcCtx := setupTestContext(t)
	ctx := context.Background()

	// 1. Seed dynamic Loyalty Tiers into DB
	tierBronze := model.LoyaltyTier{
		ID:              uuid.New(),
		Code:            "bronze",
		Name:            "برنزی (پایه)",
		MinSpendToman:   0,
		DiscountPercent: 0.0,
		QueuePriority:   0,
		BadgeColor:      "#cd7f32",
		IsActive:        true,
	}
	tierSilver := model.LoyaltyTier{
		ID:              uuid.New(),
		Code:            "silver",
		Name:            "نقره‌ای (ویژه)",
		MinSpendToman:   5000000, // 5M
		DiscountPercent: 2.0,     // 2%
		QueuePriority:   10,
		BadgeColor:      "#94a3b8",
		IsActive:        true,
	}
	tierGold := model.LoyaltyTier{
		ID:              uuid.New(),
		Code:            "gold",
		Name:            "طلایی (ریسلرها)",
		MinSpendToman:   20000000, // 20M
		DiscountPercent: 5.0,      // 5%
		QueuePriority:   50,
		BadgeColor:      "#eab308",
		IsActive:        true,
	}
	svcCtx.DB.Create(&tierBronze)
	svcCtx.DB.Create(&tierSilver)
	svcCtx.DB.Create(&tierGold)

	// 2. Create products with explicit Cost of Goods Sold (BaseCostUSD)
	prod1ID := uuid.New()
	prod1 := model.Product{
		ID:           prod1ID,
		Title:        "ChatGPT Plus 1 Month",
		TitleFA:      "اکانت چت جی‌پی‌تی پلاس",
		Slug:         "chatgpt-plus-1m",
		Category:     model.CategoryPersonalUpgrade,
		DeliveryType: model.DeliveryManualTimed,
		PriceToman:   2000000, // 2M
		BaseCostUSD:  15.0,    // ~$15 cost
		IsActive:     true,
	}
	prod2ID := uuid.New()
	prod2 := model.Product{
		ID:           prod2ID,
		Title:        "Claude Pro 1 Month",
		TitleFA:      "اکانت کلود پرو",
		Slug:         "claude-pro-1m",
		Category:     model.CategoryPersonalUpgrade,
		DeliveryType: model.DeliveryManualTimed,
		PriceToman:   2200000,
		BaseCostUSD:  16.0,
		IsActive:     true,
	}
	svcCtx.DB.Create(&prod1)
	svcCtx.DB.Create(&prod2)

	// 3. Create VIP Customer (Already Silver with 2% discount)
	custPhone := "09121112233"
	cust := model.Customer{
		ID:           uuid.New(),
		Phone:        custPhone,
		FullName:     "کاربر نقره‌ای آزمایشی",
		Tier:         "silver",
		BalanceToman: 10000000, // 10M
		IsActive:     true,
	}
	svcCtx.DB.Create(&cust)

	// 4. Test Order Creation with Automatic Loyalty Discount
	orderLogic := public.NewOrderLogic(ctx, svcCtx)
	createReq := types.CreateOrderRequest{
		ProductID:     prod1ID.String(),
		CustomerPhone: custPhone,
		PaymentMethod: string(model.PaymentMethodBalance), // Paid instantly with wallet
	}

	orderRes, err := orderLogic.CreateOrder(&createReq)
	if err != nil {
		t.Fatalf("Failed to create order with loyalty discount: %v", err)
	}

	// Verify order discounts and priority in DB
	var createdOrder model.Order
	if err := svcCtx.DB.Preload("Items").Where("id = ?", orderRes.OrderID).First(&createdOrder).Error; err != nil {
		t.Fatalf("Order not found in DB: %v", err)
	}

	// 2M * 2% discount = 40,000 Toman discount
	expectedDiscount := int64(40000)
	if createdOrder.LoyaltyDiscountToman != expectedDiscount {
		t.Errorf("Expected loyalty discount %d, got %d", expectedDiscount, createdOrder.LoyaltyDiscountToman)
	}
	expectedTotal := int64(2000000) - expectedDiscount
	if createdOrder.TotalAmountToman != expectedTotal {
		t.Errorf("Expected total amount %d, got %d", expectedTotal, createdOrder.TotalAmountToman)
	}
	if createdOrder.PriorityScore != 10 {
		t.Errorf("Expected queue priority 10 for silver customer, got %d", createdOrder.PriorityScore)
	}
	if createdOrder.CustomerTier != "silver" {
		t.Errorf("Expected customer tier 'silver', got '%s'", createdOrder.CustomerTier)
	}

	// 5. Create a 5SIM Virtual Number order
	vnOrder := model.VirtualNumberOrder{
		ID:            uuid.New(),
		OrderNo:       "VN-TEST-1001",
		Phone:         "+447123456789",
		Country:       "england",
		CountryNameFA: "انگلستان",
		CountryFlag:   "🇬🇧",
		Service:       "telegram",
		ServiceNameFA: "تلگرام",
		BaseCost:      0.80, // $0.80
		PriceToman:    150000,
		Status:        model.VNumStatusReceived,
		CustomerPhone: custPhone,
		CreatedAt:     time.Now(),
	}
	svcCtx.DB.Create(&vnOrder)

	// 6. Test BI Analytics Overview Engine
	biLogic := admin.NewBIAnalyticsLogic(ctx, svcCtx)
	biOverview, err := biLogic.GetOverview(&types.AdminBIOverviewRequest{
		Range:   "30d",
		GroupBy: "daily",
	})
	if err != nil {
		t.Fatalf("Failed to get BI analytics overview: %v", err)
	}

	if biOverview == nil {
		t.Fatal("BI overview returned nil")
	}

	kpis := biOverview.KPIs
	if kpis.PaidOrders < 2 {
		t.Errorf("Expected at least 2 paid orders (1 product + 1 vn), got %d", kpis.PaidOrders)
	}
	if kpis.TotalRevenueToman < 2000000 {
		t.Errorf("Expected revenue > 2M, got %d", kpis.TotalRevenueToman)
	}
	if kpis.GrossProfitToman <= 0 {
		t.Errorf("Expected positive gross profit, got %d", kpis.GrossProfitToman)
	}
	if kpis.GrossMarginPercent <= 0 {
		t.Errorf("Expected positive gross margin percent, got %f", kpis.GrossMarginPercent)
	}
	if kpis.AverageOrderValueToman <= 0 {
		t.Errorf("Expected positive AOV, got %d", kpis.AverageOrderValueToman)
	}

	// Verify Top Countries in BI
	if len(biOverview.TopCountries) == 0 {
		t.Error("Expected top countries to contain England for 5SIM")
	} else if biOverview.TopCountries[0].Country != "england" {
		t.Errorf("Expected top country to be 'england', got '%s'", biOverview.TopCountries[0].Country)
	}

	// 7. Test Auto-Upgrade to Gold (Buy large amount)
	// Create another order of 25,000,000 to push lifetime spend above 20M gold threshold
	largeOrder := model.Order{
		ID:                  uuid.New(),
		OrderNumber:         "ANX-LARGE-TEST",
		CustomerPhone:       custPhone,
		CustomerID:          &cust.ID,
		Status:              model.OrderPaid,
		TotalAmountToman:    25000000,
		SubtotalAmountToman: 25000000,
		CreatedAt:           time.Now(),
	}
	svcCtx.DB.Create(&largeOrder)

	err = loyalty.RecalculateCustomerTier(ctx, svcCtx.DB, cust.ID)
	if err != nil {
		t.Fatalf("Failed to recalculate customer tier: %v", err)
	}

	var updatedCust model.Customer
	svcCtx.DB.Where("id = ?", cust.ID).First(&updatedCust)
	if updatedCust.Tier != "gold" {
		t.Errorf("Expected customer to be auto-upgraded to 'gold', got '%s'", updatedCust.Tier)
	}

	// 8. Test Queue Priority Sorting for Operator Desk
	orderDeskLogic := admin.NewOrderDeskLogic(ctx, svcCtx)
	queueOrders, err := orderDeskLogic.ListQueueOrders("", "")
	if err != nil {
		t.Fatalf("Failed to list queue orders: %v", err)
	}

	if len(queueOrders) > 0 {
		// First order must have priority_score >= 10
		if queueOrders[0].PriorityScore < 10 {
			t.Errorf("Expected VIP order to be prioritized at top of operator queue, got priority %d", queueOrders[0].PriorityScore)
		}
	}
}
