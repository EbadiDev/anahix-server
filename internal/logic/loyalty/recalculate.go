package loyalty

import (
	"context"
	"strings"
	"time"

	"github.com/EbadiDev/anahix-server/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RecalculateCustomerTier calculates the customer's cumulative or monthly spend and upgrades/downgrades tier
func RecalculateCustomerTier(ctx context.Context, db *gorm.DB, customerID uuid.UUID) error {
	var customer model.Customer
	if err := db.WithContext(ctx).Where("id = ?", customerID).First(&customer).Error; err != nil {
		return err
	}

	now := time.Now()
	monthAgo := now.AddDate(0, 0, -30)

	type SpendSummary struct {
		Total   int64
		Monthly int64
	}
	var sum SpendSummary
	db.WithContext(ctx).Model(&model.Order{}).
		Select(`
			COALESCE(SUM(CASE WHEN status IN ('paid','completed','in_queue','processing','awaiting_customer_action') THEN total_amount_toman ELSE 0 END), 0) as total,
			COALESCE(SUM(CASE WHEN status IN ('paid','completed','in_queue','processing','awaiting_customer_action') AND created_at >= ? THEN total_amount_toman ELSE 0 END), 0) as monthly
		`, monthAgo).
		Where("customer_phone = ? OR customer_id = ?", customer.Phone, customer.ID).
		Scan(&sum)

	// Virtual numbers spend
	var vnTotal int64 = 0
	var vnMonthly int64 = 0
	db.WithContext(ctx).Model(&model.VirtualNumberOrder{}).
		Select(`
			COALESCE(SUM(CASE WHEN status IN ('RECEIVED','FINISHED') THEN price_toman ELSE 0 END), 0) as vn_total,
			COALESCE(SUM(CASE WHEN status IN ('RECEIVED','FINISHED') AND created_at >= ? THEN price_toman ELSE 0 END), 0) as vn_monthly
		`, monthAgo).
		Where("customer_phone = ?", customer.Phone).
		Row().Scan(&vnTotal, &vnMonthly)

	customer.TotalSpentToman = sum.Total + vnTotal
	customer.MonthlySpentToman = sum.Monthly + vnMonthly

	// Check calculation period setting from database
	period := "monthly_30d"
	var s model.SystemSetting
	if err := db.WithContext(ctx).Where("category = ? AND key = ?", "loyalty", "loyalty_calculation_period").First(&s).Error; err == nil {
		if s.Value != "" {
			period = s.Value
		}
	}

	activeSpent := customer.TotalSpentToman
	if period == "monthly_30d" {
		activeSpent = customer.MonthlySpentToman
	}

	// Fetch active tiers from database
	var tiers []model.LoyaltyTier
	db.WithContext(ctx).Where("is_active = ?", true).Order("min_spend_toman asc").Find(&tiers)
	if len(tiers) > 0 {
		bestTier := tiers[0].Code
		for _, t := range tiers {
			if activeSpent >= t.MinSpendToman {
				bestTier = t.Code
			}
		}

		if !strings.EqualFold(customer.Tier, bestTier) {
			customer.Tier = bestTier
			customer.TierUpdatedAt = &now
		}
	}

	return db.WithContext(ctx).Save(&customer).Error
}

// RecalculateByPhone recalculates customer tier using customer phone
func RecalculateByPhone(ctx context.Context, db *gorm.DB, phone string) error {
	if phone == "" {
		return nil
	}
	var customer model.Customer
	if err := db.WithContext(ctx).Where("phone = ?", phone).First(&customer).Error; err != nil {
		return err
	}
	return RecalculateCustomerTier(ctx, db, customer.ID)
}
