package admin

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/EbadiDev/anahix-server/internal/model"
	"github.com/EbadiDev/anahix-server/internal/svc"
	"github.com/EbadiDev/anahix-server/internal/types"
	"github.com/EbadiDev/anahix-server/pkg/exchangerate"
	"gorm.io/gorm/clause"
)

type VirtualNumberAdminLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewVirtualNumberAdminLogic(ctx context.Context, svcCtx *svc.ServiceContext) *VirtualNumberAdminLogic {
	return &VirtualNumberAdminLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// FiveSimAdminLogic alias for backwards compatibility
type FiveSimAdminLogic = VirtualNumberAdminLogic

func NewFiveSimAdminLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FiveSimAdminLogic {
	return NewVirtualNumberAdminLogic(ctx, svcCtx)
}

func (l *VirtualNumberAdminLogic) GetSettings() (*types.VirtualNumberSettingsResponse, error) {
	var settings []model.SystemSetting
	err := l.svcCtx.DB.WithContext(l.ctx).
		Where("category IN (?, ?)", "virtual_numbers", "fivesim").
		Find(&settings).Error
	if err != nil {
		return nil, err
	}

	configMap := make(map[string]string)
	for _, s := range settings {
		configMap[s.Key] = s.Value
	}

	apiKey := configMap["vn_api_key"]
	if apiKey == "" {
		apiKey = configMap["fivesim_api_key"]
	}
	hasKey := strings.TrimSpace(apiKey) != ""
	maskedKey := ""
	if hasKey {
		if len(apiKey) > 8 {
			maskedKey = apiKey[:4] + "••••••••" + apiKey[len(apiKey)-4:]
		} else {
			maskedKey = "••••••••"
		}
	}

	enabled := configMap["vn_enabled"] == "true" || configMap["fivesim_enabled"] == "true"
	marginStr := configMap["vn_margin_percent"]
	if marginStr == "" {
		marginStr = configMap["fivesim_margin_percent"]
	}
	margin, _ := strconv.ParseFloat(marginStr, 64)
	if margin <= 0 {
		margin = 25
	}

	usdRateStr := configMap["vn_usd_rate"]
	if usdRateStr == "" {
		usdRateStr = configMap["fivesim_usd_rate"]
	}
	usdRate, _ := strconv.ParseInt(usdRateStr, 10, 64)
	if usdRate <= 0 {
		usdRate = 95000
	}

	rubRateStr := configMap["vn_rub_rate"]
	if rubRateStr == "" {
		rubRateStr = configMap["fivesim_rub_rate"]
	}
	rubRate, _ := strconv.ParseInt(rubRateStr, 10, 64)
	if rubRate <= 0 {
		rubRate = 1050
	}

	defaultOperator := configMap["vn_default_operator"]
	if defaultOperator == "" {
		defaultOperator = configMap["fivesim_default_operator"]
	}
	if defaultOperator == "" {
		defaultOperator = "any"
	}

	ratePriority := configMap["vn_rate_priority"]
	if ratePriority == "" {
		ratePriority = configMap["fivesim_rate_priority"]
	}
	if ratePriority == "" {
		ratePriority = "manual"
	}

	// Resolve exchange rate with SwapWallet converter
	rateClient := exchangerate.GetDefaultClient()
	effectiveRate, activeSource, liveUSDT := rateClient.ResolveRate(ratePriority, usdRate)

	// Probe SMS provider connection & profile
	prof, profErr := l.svcCtx.SMSProvider.GetProfile(apiKey)
	balance := 0.0
	rating := 100.0
	statusMsg := "سامانه شماره‌های مجازی آناهیکس آماده دریافت درخواست است (حالت دمو/شبیه‌ساز)"
	isMock := true

	if profErr != nil {
		statusMsg = fmt.Sprintf("خطا در ارتباط با زیرساخت شماره‌های مجازی: %v", profErr)
	} else if prof != nil {
		balance = prof.Balance
		rating = prof.Rating
		isMock = prof.IsMock
		if !prof.IsMock {
			statusMsg = fmt.Sprintf("اتصال موفق به حساب کاربری با امتیاز %.1f", prof.Rating)
		}
	}

	return &types.VirtualNumberSettingsResponse{
		Enabled:          enabled,
		APIKeyMasked:     maskedKey,
		HasAPIKey:        hasKey,
		MarginPercent:    margin,
		USDRate:          usdRate,
		RUBRate:          rubRate,
		DefaultOperator:  defaultOperator,
		RatePriority:     ratePriority,
		LiveUSDRate:      liveUSDT,
		EffectiveUSDRate: effectiveRate,
		ActiveRateSource: activeSource,
		Balance:          balance,
		Rating:           rating,
		StatusMessage:    statusMsg,
		IsMock:           isMock,
	}, nil
}

func (l *VirtualNumberAdminLogic) UpdateSettings(req *types.UpdateVirtualNumberSettingsRequest) (*types.VirtualNumberSettingsResponse, error) {
	upsertSetting := func(key, val, valType, desc string) error {
		item := model.SystemSetting{
			Category: "virtual_numbers",
			Key:      key,
			Value:    val,
			Type:     valType,
			Desc:     desc,
		}
		return l.svcCtx.DB.WithContext(l.ctx).Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "key"}},
			DoUpdates: clause.AssignmentColumns([]string{"value", "updated_at"}),
		}).Create(&item).Error
	}

	if req.Enabled != nil {
		enVal := "false"
		if *req.Enabled {
			enVal = "true"
		}
		if err := upsertSetting("vn_enabled", enVal, "bool", "فعال بودن فروش شماره مجازی"); err != nil {
			return nil, err
		}
	}

	if req.APIKey != "" {
		if err := upsertSetting("vn_api_key", strings.TrimSpace(req.APIKey), "string", "کلید فعال‌سازی زیرساخت شماره‌های مجازی آناهیکس"); err != nil {
			return nil, err
		}
	}

	if req.MarginPercent > 0 {
		if err := upsertSetting("vn_margin_percent", fmt.Sprintf("%.2f", req.MarginPercent), "number", "درصد مارجین سود"); err != nil {
			return nil, err
		}
	}

	if req.USDRate > 0 {
		if err := upsertSetting("vn_usd_rate", strconv.FormatInt(req.USDRate, 10), "number", "نرخ دلار به تومان"); err != nil {
			return nil, err
		}
	}

	if req.RUBRate > 0 {
		if err := upsertSetting("vn_rub_rate", strconv.FormatInt(req.RUBRate, 10), "number", "نرخ روبل به تومان"); err != nil {
			return nil, err
		}
	}

	if req.DefaultOperator != "" {
		if err := upsertSetting("vn_default_operator", req.DefaultOperator, "string", "اپراتور پیش‌فرض شماره مجازی"); err != nil {
			return nil, err
		}
	}

	if req.RatePriority != "" {
		if err := upsertSetting("vn_rate_priority", req.RatePriority, "string", "اولویت نرخ ارز"); err != nil {
			return nil, err
		}
	}

	return l.GetSettings()
}

func (l *VirtualNumberAdminLogic) ListVirtualOrders() ([]model.VirtualNumberOrder, error) {
	var orders []model.VirtualNumberOrder
	err := l.svcCtx.DB.WithContext(l.ctx).
		Order("created_at desc").
		Limit(100).
		Find(&orders).Error
	return orders, err
}

