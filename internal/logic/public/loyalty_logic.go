package public

import (
	"context"
	"encoding/json"
	"math"
	"strings"
	"time"

	"github.com/EbadiDev/anahix-server/internal/model"
	"github.com/EbadiDev/anahix-server/internal/svc"
	"github.com/EbadiDev/anahix-server/internal/types"
	"github.com/EbadiDev/anahix-server/pkg/xerr"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LoyaltyPublicLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLoyaltyPublicLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoyaltyPublicLogic {
	return &LoyaltyPublicLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// GetCustomerLoyaltyProfile returns the customer's current VIP status, progression bar, and perks
func (l *LoyaltyPublicLogic) GetCustomerLoyaltyProfile(customerPhone string) (*types.CustomerLoyaltyProfileResponse, error) {
	var customer model.Customer
	if customerPhone != "" {
		_ = l.svcCtx.DB.WithContext(l.ctx).Where("phone = ?", customerPhone).First(&customer).Error
	}

	// 1. Fetch all active tiers ordered by min spend
	var tiers []model.LoyaltyTier
	if err := l.svcCtx.DB.WithContext(l.ctx).Where("is_active = ?", true).Order("order_index asc, min_spend_toman asc").Find(&tiers).Error; err != nil {
		return nil, xerr.NewErrCode(xerr.DatabaseError)
	}
	if len(tiers) == 0 {
		return nil, xerr.NewErrMsg("هیچ سطح وفاداری در سیستم تعریف نشده است")
	}

	// 2. Fetch loyalty settings
	systemStatus := "active"
	var enabledSetting model.SystemSetting
	if err := l.svcCtx.DB.WithContext(l.ctx).Where("category = ? AND key = ?", "loyalty", "loyalty_enabled").First(&enabledSetting).Error; err == nil {
		val := strings.ToLower(strings.TrimSpace(enabledSetting.Value))
		if val == "false" || val == "disabled" {
			systemStatus = "disabled"
		} else if val == "coming_soon" {
			systemStatus = "coming_soon"
		} else {
			systemStatus = "active"
		}
	}

	period := "monthly_30d"
	var periodSetting model.SystemSetting
	if err := l.svcCtx.DB.WithContext(l.ctx).Where("category = ? AND key = ?", "loyalty", "loyalty_calculation_period").First(&periodSetting).Error; err == nil {
		if periodSetting.Value != "" {
			period = periodSetting.Value
		}
	}

	// 3. Compute live spend for this customer
	var totalSpentToman int64 = 0
	var monthlySpentToman int64 = 0
	var totalSavedToman int64 = 0

	if customer.ID != uuid.Nil {
		now := time.Now()
		monthAgo := now.AddDate(0, 0, -30)

		type OrderStats struct {
			TotalPaid   int64
			MonthlyPaid int64
			TotalSaved  int64
		}
		var stat OrderStats
		l.svcCtx.DB.WithContext(l.ctx).Model(&model.Order{}).
			Select(`
				COALESCE(SUM(CASE WHEN status IN ('paid','completed','in_queue','processing','awaiting_customer_action') THEN total_amount_toman ELSE 0 END), 0) as total_paid,
				COALESCE(SUM(CASE WHEN status IN ('paid','completed','in_queue','processing','awaiting_customer_action') AND created_at >= ? THEN total_amount_toman ELSE 0 END), 0) as monthly_paid,
				COALESCE(SUM(CASE WHEN status IN ('paid','completed','in_queue','processing','awaiting_customer_action') THEN loyalty_discount_toman + discount_amount_toman ELSE 0 END), 0) as total_saved
			`, monthAgo).
			Where("customer_phone = ? OR customer_id = ?", customer.Phone, customer.ID).
			Scan(&stat)

		totalSpentToman = stat.TotalPaid
		monthlySpentToman = stat.MonthlyPaid
		totalSavedToman = stat.TotalSaved

		// Include virtual numbers
		var vnTotal int64 = 0
		var vnMonthly int64 = 0
		l.svcCtx.DB.WithContext(l.ctx).Model(&model.VirtualNumberOrder{}).
			Select(`
				COALESCE(SUM(CASE WHEN status IN ('RECEIVED','FINISHED') THEN price_toman ELSE 0 END), 0) as vn_total,
				COALESCE(SUM(CASE WHEN status IN ('RECEIVED','FINISHED') AND created_at >= ? THEN price_toman ELSE 0 END), 0) as vn_monthly
			`, monthAgo).
			Where("customer_phone = ?", customer.Phone).
			Row().Scan(&vnTotal, &vnMonthly)

		totalSpentToman += vnTotal
		monthlySpentToman += vnMonthly
	}

	activeSpent := totalSpentToman
	if period == "monthly_30d" {
		activeSpent = monthlySpentToman
	}

	// 4. Resolve current and next tier based on active spend
	// Find highest tier where MinSpendToman <= activeSpent
	var currentTier model.LoyaltyTier = tiers[0]
	var nextTier *model.LoyaltyTier = nil

	for i := len(tiers) - 1; i >= 0; i-- {
		if activeSpent >= tiers[i].MinSpendToman {
			currentTier = tiers[i]
			if i+1 < len(tiers) {
				nextTier = &tiers[i+1]
			}
			break
		}
	}

	// If customer explicitly had a manual VIP tier assigned, prioritize it if higher
	if customer.Tier != "" {
		for i, t := range tiers {
			if strings.EqualFold(t.Code, customer.Tier) {
				if t.MinSpendToman > currentTier.MinSpendToman {
					currentTier = t
					if i+1 < len(tiers) {
						nextTier = &tiers[i+1]
					} else {
						nextTier = nil
					}
				}
				break
			}
		}
	}

	// 5. Calculate progress towards next tier
	var progressPercent float64 = 100.0
	var requiredForNext int64 = 0

	if nextTier != nil {
		tierSpan := nextTier.MinSpendToman - currentTier.MinSpendToman
		spendInTier := activeSpent - currentTier.MinSpendToman
		if spendInTier < 0 {
			spendInTier = 0
		}
		if tierSpan > 0 {
			progressPercent = math.Round((float64(spendInTier)/float64(tierSpan))*1000) / 10
			if progressPercent > 100 {
				progressPercent = 100
			}
		}
		requiredForNext = nextTier.MinSpendToman - activeSpent
		if requiredForNext < 0 {
			requiredForNext = 0
		}
	}

	// Convert all tiers to DTOs
	var allTiersDTO []types.LoyaltyTierDTO
	for _, t := range tiers {
		var perks []string
		if t.PerksJSON != "" {
			_ = json.Unmarshal([]byte(t.PerksJSON), &perks)
		}
		allTiersDTO = append(allTiersDTO, types.LoyaltyTierDTO{
			ID:              t.ID.String(),
			Code:            t.Code,
			Name:            t.Name,
			NameEN:          t.NameEN,
			MinSpendToman:   t.MinSpendToman,
			DiscountPercent: t.DiscountPercent,
			QueuePriority:   t.QueuePriority,
			BadgeColor:      t.BadgeColor,
			BadgeGradient:   t.BadgeGradient,
			Icon:            t.Icon,
			Description:     t.Description,
			Perks:           perks,
			OrderIndex:      t.OrderIndex,
			IsActive:        t.IsActive,
		})
	}

	var currentPerks []string
	if currentTier.PerksJSON != "" {
		_ = json.Unmarshal([]byte(currentTier.PerksJSON), &currentPerks)
	}

	currentDTO := types.LoyaltyTierDTO{
		ID:              currentTier.ID.String(),
		Code:            currentTier.Code,
		Name:            currentTier.Name,
		NameEN:          currentTier.NameEN,
		MinSpendToman:   currentTier.MinSpendToman,
		DiscountPercent: currentTier.DiscountPercent,
		QueuePriority:   currentTier.QueuePriority,
		BadgeColor:      currentTier.BadgeColor,
		BadgeGradient:   currentTier.BadgeGradient,
		Icon:            currentTier.Icon,
		Description:     currentTier.Description,
		Perks:           currentPerks,
		OrderIndex:      currentTier.OrderIndex,
		IsActive:        currentTier.IsActive,
	}

	var nextDTO *types.LoyaltyTierDTO
	if nextTier != nil {
		var nextPerks []string
		if nextTier.PerksJSON != "" {
			_ = json.Unmarshal([]byte(nextTier.PerksJSON), &nextPerks)
		}
		nextDTO = &types.LoyaltyTierDTO{
			ID:              nextTier.ID.String(),
			Code:            nextTier.Code,
			Name:            nextTier.Name,
			NameEN:          nextTier.NameEN,
			MinSpendToman:   nextTier.MinSpendToman,
			DiscountPercent: nextTier.DiscountPercent,
			QueuePriority:   nextTier.QueuePriority,
			BadgeColor:      nextTier.BadgeColor,
			BadgeGradient:   nextTier.BadgeGradient,
			Icon:            nextTier.Icon,
			Description:     nextTier.Description,
			Perks:           nextPerks,
			OrderIndex:      nextTier.OrderIndex,
			IsActive:        nextTier.IsActive,
		}
	}

	return &types.CustomerLoyaltyProfileResponse{
		SystemStatus:      systemStatus,
		CurrentTier:       currentDTO,
		NextTier:          nextDTO,
		TotalSpentToman:   totalSpentToman,
		MonthlySpentToman: monthlySpentToman,
		Period:            period,
		ActiveSpentToman:  activeSpent,
		RequiredForNext:   requiredForNext,
		ProgressPercent:   progressPercent,
		AllTiers:          allTiersDTO,
		ActiveDiscountPct: currentTier.DiscountPercent,
		TotalSavedToman:   totalSavedToman,
	}, nil
}

// RecalculateCustomerTier updates a customer's cached spend and tier dynamically
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

	// Virtual numbers
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

	// Check calculation period setting
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

	// Fetch active tiers
	var tiers []model.LoyaltyTier
	db.WithContext(ctx).Where("is_active = ?", true).Order("min_spend_toman asc").Find(&tiers)
	if len(tiers) > 0 {
		bestTier := tiers[0].Code
		for _, t := range tiers {
			if activeSpent >= t.MinSpendToman {
				bestTier = t.Code
			}
		}

		if customer.Tier != bestTier {
			customer.Tier = bestTier
			customer.TierUpdatedAt = &now
		}
	}

	return db.WithContext(ctx).Save(&customer).Error
}
