package admin

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/EbadiDev/anahix-server/internal/model"
	"github.com/EbadiDev/anahix-server/internal/svc"
	"github.com/EbadiDev/anahix-server/internal/types"
	"github.com/EbadiDev/anahix-server/pkg/xerr"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LoyaltyAdminLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLoyaltyAdminLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoyaltyAdminLogic {
	return &LoyaltyAdminLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// ListTiers returns all configured loyalty tiers with live user counts
func (l *LoyaltyAdminLogic) ListTiers() ([]types.LoyaltyTierDTO, error) {
	var tiers []model.LoyaltyTier
	if err := l.svcCtx.DB.WithContext(l.ctx).Order("order_index asc, min_spend_toman asc").Find(&tiers).Error; err != nil {
		return nil, xerr.NewErrCode(xerr.DatabaseError)
	}

	// Count users per tier
	type tierCount struct {
		Tier  string
		Count int64
	}
	var counts []tierCount
	l.svcCtx.DB.WithContext(l.ctx).Model(&model.Customer{}).Select("tier, count(*) as count").Group("tier").Scan(&counts)
	countMap := make(map[string]int64)
	for _, c := range counts {
		countMap[strings.ToLower(c.Tier)] = c.Count
	}

	result := make([]types.LoyaltyTierDTO, 0, len(tiers))
	for _, t := range tiers {
		var perks []string
		if t.PerksJSON != "" {
			_ = json.Unmarshal([]byte(t.PerksJSON), &perks)
		}

		result = append(result, types.LoyaltyTierDTO{
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
			UserCount:       countMap[strings.ToLower(t.Code)],
			CreatedAt:       t.CreatedAt,
			UpdatedAt:       t.UpdatedAt,
		})
	}

	return result, nil
}

// CreateTier creates a new customizable VIP tier in database
func (l *LoyaltyAdminLogic) CreateTier(req *types.CreateOrUpdateLoyaltyTierRequest) (*types.LoyaltyTierDTO, error) {
	code := strings.ToLower(strings.TrimSpace(req.Code))
	if code == "" {
		return nil, xerr.NewErrMsg("کد یکتای سطح الزامی است")
	}

	// Check unique code
	var existing model.LoyaltyTier
	if err := l.svcCtx.DB.WithContext(l.ctx).Where("code = ?", code).First(&existing).Error; err == nil {
		return nil, xerr.NewErrMsg("سطحی با این کد انگلیسی قبلاً ثبت شده است")
	}

	perksJSON := "[]"
	if len(req.Perks) > 0 {
		if b, err := json.Marshal(req.Perks); err == nil {
			perksJSON = string(b)
		}
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	tier := model.LoyaltyTier{
		ID:              uuid.New(),
		Code:            code,
		Name:            req.Name,
		NameEN:          req.NameEN,
		MinSpendToman:   req.MinSpendToman,
		DiscountPercent: req.DiscountPercent,
		QueuePriority:   req.QueuePriority,
		BadgeColor:      req.BadgeColor,
		BadgeGradient:   req.BadgeGradient,
		Icon:            req.Icon,
		Description:     req.Description,
		PerksJSON:       perksJSON,
		OrderIndex:      req.OrderIndex,
		IsActive:        isActive,
	}

	if err := l.svcCtx.DB.WithContext(l.ctx).Create(&tier).Error; err != nil {
		return nil, xerr.NewErrCode(xerr.DatabaseError)
	}

	return &types.LoyaltyTierDTO{
		ID:              tier.ID.String(),
		Code:            tier.Code,
		Name:            tier.Name,
		NameEN:          tier.NameEN,
		MinSpendToman:   tier.MinSpendToman,
		DiscountPercent: tier.DiscountPercent,
		QueuePriority:   tier.QueuePriority,
		BadgeColor:      tier.BadgeColor,
		BadgeGradient:   tier.BadgeGradient,
		Icon:            tier.Icon,
		Description:     tier.Description,
		Perks:           req.Perks,
		OrderIndex:      tier.OrderIndex,
		IsActive:        tier.IsActive,
		CreatedAt:       tier.CreatedAt,
		UpdatedAt:       tier.UpdatedAt,
	}, nil
}

// UpdateTier modifies an existing loyalty tier
func (l *LoyaltyAdminLogic) UpdateTier(idStr string, req *types.CreateOrUpdateLoyaltyTierRequest) (*types.LoyaltyTierDTO, error) {
	id, err := uuid.Parse(idStr)
	if err != nil {
		return nil, xerr.NewErrMsg("شناسه سطح نامعتبر است")
	}

	var tier model.LoyaltyTier
	if err := l.svcCtx.DB.WithContext(l.ctx).Where("id = ?", id).First(&tier).Error; err != nil {
		return nil, xerr.NewErrMsg("سطح وفاداری مورد نظر یافت نشد")
	}

	// Prevent changing bronze code to preserve base fallback
	newCode := strings.ToLower(strings.TrimSpace(req.Code))
	if tier.Code == "bronze" && newCode != "bronze" && newCode != "" {
		return nil, xerr.NewErrMsg("کد سطح پایه (bronze) قابل تغییر نیست")
	}

	if newCode != "" && newCode != tier.Code {
		var conflict model.LoyaltyTier
		if err := l.svcCtx.DB.WithContext(l.ctx).Where("code = ? AND id != ?", newCode, id).First(&conflict).Error; err == nil {
			return nil, xerr.NewErrMsg("کد انگلیسی وارد شده تکراری است")
		}
		tier.Code = newCode
	}

	tier.Name = req.Name
	tier.NameEN = req.NameEN
	tier.MinSpendToman = req.MinSpendToman
	tier.DiscountPercent = req.DiscountPercent
	tier.QueuePriority = req.QueuePriority
	tier.BadgeColor = req.BadgeColor
	tier.BadgeGradient = req.BadgeGradient
	tier.Icon = req.Icon
	tier.Description = req.Description
	tier.OrderIndex = req.OrderIndex

	if req.IsActive != nil {
		tier.IsActive = *req.IsActive
	}

	if req.Perks != nil {
		if b, err := json.Marshal(req.Perks); err == nil {
			tier.PerksJSON = string(b)
		}
	}

	if err := l.svcCtx.DB.WithContext(l.ctx).Save(&tier).Error; err != nil {
		return nil, xerr.NewErrCode(xerr.DatabaseError)
	}

	var perks []string
	if tier.PerksJSON != "" {
		_ = json.Unmarshal([]byte(tier.PerksJSON), &perks)
	}

	return &types.LoyaltyTierDTO{
		ID:              tier.ID.String(),
		Code:            tier.Code,
		Name:            tier.Name,
		NameEN:          tier.NameEN,
		MinSpendToman:   tier.MinSpendToman,
		DiscountPercent: tier.DiscountPercent,
		QueuePriority:   tier.QueuePriority,
		BadgeColor:      tier.BadgeColor,
		BadgeGradient:   tier.BadgeGradient,
		Icon:            tier.Icon,
		Description:     tier.Description,
		Perks:           perks,
		OrderIndex:      tier.OrderIndex,
		IsActive:        tier.IsActive,
		CreatedAt:       tier.CreatedAt,
		UpdatedAt:       tier.UpdatedAt,
	}, nil
}

// DeleteTier deletes a tier if not bronze
func (l *LoyaltyAdminLogic) DeleteTier(idStr string) error {
	id, err := uuid.Parse(idStr)
	if err != nil {
		return xerr.NewErrMsg("شناسه سطح نامعتبر است")
	}

	var tier model.LoyaltyTier
	if err := l.svcCtx.DB.WithContext(l.ctx).Where("id = ?", id).First(&tier).Error; err != nil {
		return xerr.NewErrMsg("سطح وفاداری یافت نشد")
	}

	if tier.Code == "bronze" {
		return xerr.NewErrMsg("سطح پایه برنزی سیستم قابل حذف نمی‌باشد")
	}

	// Reassign customers belonging to this tier back to bronze
	_ = l.svcCtx.DB.WithContext(l.ctx).Model(&model.Customer{}).Where("tier = ?", tier.Code).Update("tier", "bronze")

	if err := l.svcCtx.DB.WithContext(l.ctx).Delete(&tier).Error; err != nil {
		return xerr.NewErrCode(xerr.DatabaseError)
	}

	return nil
}

// GetLoyaltySettings returns system level settings for loyalty club
func (l *LoyaltyAdminLogic) GetLoyaltySettings() (map[string]string, error) {
	var settings []model.SystemSetting
	l.svcCtx.DB.WithContext(l.ctx).Where("category = ?", "loyalty").Find(&settings)

	res := map[string]string{
		"loyalty_enabled":            "true",
		"loyalty_calculation_period": "monthly_30d",
		"loyalty_auto_upgrade":       "true",
	}

	for _, s := range settings {
		res[s.Key] = s.Value
	}

	return res, nil
}

// UpdateLoyaltySettings saves system level loyalty settings
func (l *LoyaltyAdminLogic) UpdateLoyaltySettings(settings map[string]string) error {
	for k, v := range settings {
		var s model.SystemSetting
		err := l.svcCtx.DB.WithContext(l.ctx).Where("category = ? AND key = ?", "loyalty", k).First(&s).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s = model.SystemSetting{
				Category: "loyalty",
				Key:      k,
				Value:    v,
				Type:     "string",
				Desc:     "تنظیمات باشگاه وفاداری و مشتریان VIP",
			}
			_ = l.svcCtx.DB.WithContext(l.ctx).Create(&s)
		} else if err == nil {
			s.Value = v
			_ = l.svcCtx.DB.WithContext(l.ctx).Save(&s)
		}
	}
	return nil
}
