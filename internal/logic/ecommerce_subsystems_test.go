package logic

import (
	"context"
	"testing"
	"time"

	"github.com/EbadiDev/anahix-server/internal/logic/admin"
	"github.com/EbadiDev/anahix-server/internal/logic/public"
	"github.com/EbadiDev/anahix-server/internal/model"
	"github.com/EbadiDev/anahix-server/internal/svc"
	"github.com/EbadiDev/anahix-server/internal/types"
	"github.com/EbadiDev/anahix-server/pkg/crypto"
	"github.com/EbadiDev/anahix-server/pkg/orm"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func setupTestSubsystems(t *testing.T) (*svc.ServiceContext, model.Customer, model.Product) {
	db, err := orm.InitTestDB()
	assert.NoError(t, err)

	vault, err := crypto.NewVault("test-master-key-32bytes-secret!")
	assert.NoError(t, err)

	svcCtx := &svc.ServiceContext{
		DB:    db,
		Vault: vault,
	}

	uid := uuid.New().String()[:6]
	cust := model.Customer{
		ID:                   uuid.New(),
		Phone:                "0912" + uid,
		FullName:             "تستر خریدار",
		BalanceToman:         5000000,
		ReferralCode:         "ANX-" + uid,
		ReferralBalanceToman: 200000,
		IsActive:             true,
	}
	assert.NoError(t, db.Create(&cust).Error)

	prod := model.Product{
		ID:               uuid.New(),
		Title:            "Gemini Pro Test " + uid,
		TitleFA:          "گوگل جمینای تست " + uid,
		Slug:             "gemini-pro-test-" + uid,
		Category:         model.CategoryAILink,
		DeliveryType:     model.DeliveryInstant,
		DeliveryStrategy: "local_inventory",
		PriceToman:       1000000,
		IsActive:         true,
	}
	assert.NoError(t, db.Create(&prod).Error)

	return svcCtx, cust, prod
}

func TestCouponValidationLogic(t *testing.T) {
	svcCtx, cust, _ := setupTestSubsystems(t)
	couponLogic := public.NewCouponLogic(context.Background(), svcCtx)

	// Create test coupon: 20% off, min order 500,000, max discount 300,000
	coupon := model.Coupon{
		ID:                  uuid.New(),
		Code:                "TEST20",
		DiscountType:        model.DiscountTypePercentage,
		DiscountValue:       20,
		MinOrderAmountToman: 500000,
		MaxDiscountToman:    300000,
		PerUserLimit:        1,
		IsActive:            true,
	}
	assert.NoError(t, svcCtx.DB.Create(&coupon).Error)

	// 1. Valid application on 1,000,000 Toman -> 200,000 Toman discount
	res, err := couponLogic.ValidateCoupon(&types.ValidateCouponRequest{
		Code:          "TEST20",
		SubtotalToman: 1000000,
		CustomerPhone: cust.Phone,
	})
	assert.NoError(t, err)
	assert.True(t, res.Valid)
	assert.Equal(t, int64(200000), res.DiscountToman)
	assert.Equal(t, int64(800000), res.FinalAmountToman)

	// 2. Minimum amount not met (400,000 < 500,000)
	resMin, err := couponLogic.ValidateCoupon(&types.ValidateCouponRequest{
		Code:          "TEST20",
		SubtotalToman: 400000,
		CustomerPhone: cust.Phone,
	})
	assert.NoError(t, err)
	assert.False(t, resMin.Valid)

	// 3. Invalid coupon code
	resInv, err := couponLogic.ValidateCoupon(&types.ValidateCouponRequest{
		Code:          "NONEXISTENT",
		SubtotalToman: 1000000,
	})
	assert.NoError(t, err)
	assert.False(t, resInv.Valid)
}

func TestReferralStatsAndCommissionTransfer(t *testing.T) {
	svcCtx, cust, _ := setupTestSubsystems(t)
	refLogic := public.NewReferralLogic(context.Background(), svcCtx)

	// 1. Fetch stats
	stats, err := refLogic.GetReferralStats(cust.Phone)
	assert.NoError(t, err)
	assert.Equal(t, cust.ReferralCode, stats.ReferralCode)
	assert.Equal(t, int64(200000), stats.WithdrawableCommissionToman)

	// 2. Transfer referral commission to main wallet
	initialBalance := cust.BalanceToman
	transferRes, err := refLogic.TransferToWallet(cust.Phone)
	assert.NoError(t, err)
	assert.Equal(t, int64(200000), transferRes.TransferredToman)
	assert.Equal(t, initialBalance+200000, transferRes.NewBalanceToman)

	// Check customer in DB
	var updatedCust model.Customer
	assert.NoError(t, svcCtx.DB.Where("id = ?", cust.ID).First(&updatedCust).Error)
	assert.Equal(t, int64(0), updatedCust.ReferralBalanceToman)
	assert.Equal(t, initialBalance+200000, updatedCust.BalanceToman)
}

func TestInventoryStatsAndKeyReplacement(t *testing.T) {
	svcCtx, _, prod := setupTestSubsystems(t)
	invLogic := admin.NewInventoryLogic(context.Background(), svcCtx)

	// 1. Bulk import keys
	importRes, err := invLogic.BulkImport(&types.BulkImportInventoryRequest{
		ProductID: prod.ID.String(),
		ItemType:  "license_key",
		Items:     []string{"KEY-1111-AAA", "KEY-2222-BBB", "KEY-3333-CCC"},
	})
	assert.NoError(t, err)
	assert.Equal(t, 3, importRes.ImportedCount)

	// 2. Stats
	stats, err := invLogic.GetStats()
	assert.NoError(t, err)
	assert.Equal(t, int64(3), stats.TotalCount)
	assert.Equal(t, int64(3), stats.AvailableCount)

	// 3. Simulate an order reserving KEY-1111-AAA
	var firstItem model.InventoryItem
	assert.NoError(t, svcCtx.DB.Where("product_id = ?", prod.ID).First(&firstItem).Error)
	firstItem.Status = model.InventorySold
	now := time.Now()
	firstItem.SoldAt = &now
	orderID := uuid.New()
	firstItem.ReservedOrderID = &orderID
	assert.NoError(t, svcCtx.DB.Save(&firstItem).Error)

	// Create order item
	orderItem := model.OrderItem{
		ID:               uuid.New(),
		OrderID:          orderID,
		ProductID:        prod.ID,
		InventoryItemID:  &firstItem.ID,
		ProductTitle:     prod.TitleFA,
		DeliveredPayload: "KEY-1111-AAA",
	}
	assert.NoError(t, svcCtx.DB.Create(&orderItem).Error)

	// 4. Operator key replacement
	repRes, err := invLogic.ReplaceKey(&types.ReplaceKeyRequest{
		OrderID:            orderID.String(),
		OldInventoryItemID: firstItem.ID.String(),
		Reason:             "کلید نامعتبر گزارش شد",
	})
	assert.NoError(t, err)
	assert.True(t, repRes.Success)
	assert.NotEmpty(t, repRes.NewInventoryItemID)
	assert.NotEqual(t, firstItem.ID.String(), repRes.NewInventoryItemID)

	// Check old item is VOID
	var oldCheck model.InventoryItem
	assert.NoError(t, svcCtx.DB.Where("id = ?", firstItem.ID).First(&oldCheck).Error)
	assert.Equal(t, model.InventoryVoid, oldCheck.Status)
}
