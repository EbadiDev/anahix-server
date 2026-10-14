package admin

import (
	"context"
	"fmt"
	"math"
	"strconv"

	"github.com/EbadiDev/anahix-server/internal/model"
	"github.com/EbadiDev/anahix-server/internal/svc"
	"github.com/EbadiDev/anahix-server/internal/types"
	"github.com/EbadiDev/anahix-server/pkg/exchangerate"
	"github.com/EbadiDev/anahix-server/pkg/xerr"
	"github.com/google/uuid"
	"gorm.io/gorm/clause"
)

type DeliveryEngineLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeliveryEngineLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeliveryEngineLogic {
	return &DeliveryEngineLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeliveryEngineLogic) getInstantDeliveryConfig() (enabled bool, apiKey string, margin float64) {
	var settings []model.SystemSetting
	_ = l.svcCtx.DB.WithContext(l.ctx).Where("category IN (?, ?)", "instant_delivery", "insightx").Find(&settings).Error

	configMap := make(map[string]string)
	for _, s := range settings {
		configMap[s.Key] = s.Value
	}

	enabled = configMap["provider_enabled"] == "true" || configMap["insightx_enabled"] == "true"
	apiKey = configMap["provider_api_key"]
	if apiKey == "" {
		apiKey = configMap["insightx_api_key"]
	}

	if marginStr, ok := configMap["provider_margin_percent"]; ok {
		margin, _ = strconv.ParseFloat(marginStr, 64)
	} else if marginStr, ok := configMap["insightx_margin_percent"]; ok {
		margin, _ = strconv.ParseFloat(marginStr, 64)
	} else {
		margin = 25.0
	}
	return
}

func (l *DeliveryEngineLogic) GetInstantDeliverySettings() (*types.InstantDeliverySettingsResponse, error) {
	enabled, apiKey, margin := l.getInstantDeliveryConfig()

	hasKey := apiKey != ""
	maskedKey := ""
	if len(apiKey) > 8 {
		maskedKey = apiKey[:4] + "••••••••" + apiKey[len(apiKey)-4:]
	} else if hasKey {
		maskedKey = "••••••••"
	}

	// Resolve exchange rate using SwapWallet Dual-Priority Engine
	ratePriority := "manual"
	var vnSetting model.SystemSetting
	if err := l.svcCtx.DB.WithContext(l.ctx).Where("key IN (?, ?)", "vn_rate_priority", "fivesim_rate_priority").First(&vnSetting).Error; err == nil {
		ratePriority = vnSetting.Value
	}
	var usdRateSetting model.SystemSetting
	usdRate := int64(95000)
	if err := l.svcCtx.DB.WithContext(l.ctx).Where("key IN (?, ?)", "vn_usd_rate", "fivesim_usd_rate").First(&usdRateSetting).Error; err == nil {
		if val, err := strconv.ParseInt(usdRateSetting.Value, 10, 64); err == nil {
			usdRate = val
		}
	}
	effectiveRate, _, _ := exchangerate.GetDefaultClient().ResolveRate(ratePriority, usdRate)

	balanceUSDT := 0.0
	rateLimit := 60
	statusMsg := "کلید اتصال مخزن تحویل آنی هنوز تنظیم نشده است"

	if hasKey {
		balResp, err := l.svcCtx.FulfillmentProvider.GetBalance(apiKey)
		if err != nil {
			statusMsg = fmt.Sprintf("خطا در استعلام حساب مخزن تحویل آنی: %v", err)
		} else {
			balanceUSDT = balResp.BalanceUSDT
			rateLimit = balResp.RateLimitPerMin
			statusMsg = fmt.Sprintf("اتصال موفق (وضعیت: %s | سقف: %d درخواست/دقیقه)", balResp.Status, rateLimit)
		}
	}

	balanceToman := int64(balanceUSDT * float64(effectiveRate))

	return &types.InstantDeliverySettingsResponse{
		Enabled:          enabled,
		APIKeyMasked:     maskedKey,
		HasAPIKey:        hasKey,
		BalanceUSDT:      balanceUSDT,
		BalanceToman:     balanceToman,
		EffectiveUSDRate: effectiveRate,
		MarginPercent:    margin,
		RateLimit:        rateLimit,
		StatusMessage:    statusMsg,
	}, nil
}

func (l *DeliveryEngineLogic) GetInsightXSettings() (*types.InstantDeliverySettingsResponse, error) {
	return l.GetInstantDeliverySettings()
}

func (l *DeliveryEngineLogic) UpdateInstantDeliverySettings(req *types.UpdateInstantDeliverySettingsRequest) error {
	upsertSetting := func(key, value, desc, valType string) error {
		setting := model.SystemSetting{
			Category: "instant_delivery",
			Key:      key,
			Value:    value,
			Desc:     desc,
			Type:     valType,
		}
		return l.svcCtx.DB.WithContext(l.ctx).Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "key"}},
			DoUpdates: clause.AssignmentColumns([]string{"value", "updated_at"}),
		}).Create(&setting).Error
	}

	if req.Enabled != nil {
		enVal := "false"
		if *req.Enabled {
			enVal = "true"
		}
		_ = upsertSetting("provider_enabled", enVal, "فعال‌سازی موتور تحویل آنی آناهیکس", "bool")
	}

	if req.APIKey != "" {
		_ = upsertSetting("provider_api_key", req.APIKey, "کلید اتصال مخزن تحویل آنی آناهیکس", "string")
	}

	if req.MarginPercent != nil {
		marginStr := fmt.Sprintf("%.2f", *req.MarginPercent)
		_ = upsertSetting("provider_margin_percent", marginStr, "درصد مارجین سود لایسنس‌ها", "float")
	}

	return nil
}

func (l *DeliveryEngineLogic) UpdateInsightXSettings(req *types.UpdateInstantDeliverySettingsRequest) error {
	return l.UpdateInstantDeliverySettings(req)
}

func (l *DeliveryEngineLogic) ListProviderCatalog() ([]types.ProviderCatalogItemDTO, error) {
	_, apiKey, margin := l.getInstantDeliveryConfig()
	if apiKey == "" {
		return []types.ProviderCatalogItemDTO{}, nil
	}

	prods, err := l.svcCtx.FulfillmentProvider.GetProducts(apiKey, false)
	if err != nil {
		return []types.ProviderCatalogItemDTO{}, nil
	}

	// Resolve exchange rate
	var usdRateSetting model.SystemSetting
	usdRate := int64(95000)
	if err := l.svcCtx.DB.WithContext(l.ctx).Where("key IN (?, ?)", "vn_usd_rate", "fivesim_usd_rate").First(&usdRateSetting).Error; err == nil {
		if val, err := strconv.ParseInt(usdRateSetting.Value, 10, 64); err == nil {
			usdRate = val
		}
	}
	var vnSetting model.SystemSetting
	ratePriority := "manual"
	if err := l.svcCtx.DB.WithContext(l.ctx).Where("key IN (?, ?)", "vn_rate_priority", "fivesim_rate_priority").First(&vnSetting).Error; err == nil {
		ratePriority = vnSetting.Value
	}
	effectiveRate, _, _ := exchangerate.GetDefaultClient().ResolveRate(ratePriority, usdRate)

	// Fetch existing linked products
	var existingProds []model.Product
	_ = l.svcCtx.DB.WithContext(l.ctx).Where("external_product_id > 0").Find(&existingProds).Error
	linkedMap := make(map[int64][]string)
	for _, ep := range existingProds {
		linkedMap[ep.ExternalProductID] = append(linkedMap[ep.ExternalProductID], ep.TitleFA)
	}

	var dtoList []types.ProviderCatalogItemDTO
	for _, p := range prods {
		cost := p.PriceUSDT
		toman := int64(math.Ceil((cost*float64(effectiveRate)*(1.0+margin/100.0))/500.0) * 500.0)
		if toman < 5000 {
			toman = 5000
		}

		linkedTitles := linkedMap[p.ID]
		dtoList = append(dtoList, types.ProviderCatalogItemDTO{
			ID:             p.ID,
			Name:           p.Name,
			Emoji:          p.Emoji,
			Description:    p.DescriptionText,
			PriceUSDT:      p.PriceUSDT,
			PriceToman:     toman,
			Stock:          p.Stock,
			Available:      p.Available,
			Unlimited:      p.Unlimited,
			IsLinked:       len(linkedTitles) > 0,
			LinkedProducts: linkedTitles,
		})
	}

	return dtoList, nil
}

func (l *DeliveryEngineLogic) ListInsightXProducts() ([]types.ProviderCatalogItemDTO, error) {
	return l.ListProviderCatalog()
}

func (l *DeliveryEngineLogic) ListProductGroups() ([]types.ProductGroupDTO, error) {
	var products []model.Product
	if err := l.svcCtx.DB.WithContext(l.ctx).Order("group_tag asc, created_at asc").Find(&products).Error; err != nil {
		return nil, err
	}

	// Group by group_tag
	groupMap := make(map[string]*types.ProductGroupDTO)
	var groupOrder []string

	for _, p := range products {
		tag := p.GroupTag
		if tag == "" {
			tag = "general"
		}

		grp, exists := groupMap[tag]
		if !exists {
			titleFA := p.GroupTitleFA
			if titleFA == "" {
				switch tag {
				case "gemini":
					titleFA = "گوگل جمینای (Google Gemini)"
				case "chatgpt":
					titleFA = "اوپن‌ای‌آی و چت‌جی‌پی‌تی (ChatGPT)"
				case "dev_tools":
					titleFA = "دستیارهای برنامه‌نویسی و ابزارهای توسعه"
				case "vpn":
					titleFA = "شبکه و ابزارهای حریم خصوصی"
				default:
					titleFA = "سرویس‌های دیجیتال عمومی"
				}
			}

			strat := p.DeliveryStrategy
			if strat == "" || strat == "auto" {
				strat = "manual_operator"
			}

			grp = &types.ProductGroupDTO{
				GroupTag:                tag,
				GroupTitleFA:            titleFA,
				DeliveryStrategy:        strat,
				LinkedProviderProductID: p.ExternalProductID,
				LinkedProviderName:      p.ExternalProductName,
				Products:                []types.ProductGroupItemDTO{},
			}
			groupMap[tag] = grp
			groupOrder = append(groupOrder, tag)
		}

		if grp.LinkedProviderProductID == 0 && p.ExternalProductID > 0 {
			grp.LinkedProviderProductID = p.ExternalProductID
			grp.LinkedProviderName = p.ExternalProductName
		}

		itemDTO := types.ProductGroupItemDTO{
			ID:                  p.ID.String(),
			Title:               p.Title,
			TitleFA:             p.TitleFA,
			Slug:                p.Slug,
			PriceToman:          p.PriceToman,
			DeliveryType:        string(p.DeliveryType),
			DeliveryStrategy:    p.DeliveryStrategy,
			ExternalProvider:    p.ExternalProvider,
			ExternalProductID:   p.ExternalProductID,
			ExternalProductName: p.ExternalProductName,
			IsActive:            p.IsActive,
		}
		grp.Products = append(grp.Products, itemDTO)
		grp.ProductsCount++
	}

	var result []types.ProductGroupDTO
	for _, tag := range groupOrder {
		result = append(result, *groupMap[tag])
	}
	return result, nil
}

func (l *DeliveryEngineLogic) UpdateGroupDeliveryStrategy(req *types.UpdateGroupStrategyRequest) error {
	if req.GroupTag == "" || req.DeliveryStrategy == "" {
		return xerr.NewErrCode(xerr.InvalidParams)
	}

	// Update all products in this group
	return l.svcCtx.DB.WithContext(l.ctx).Model(&model.Product{}).
		Where("group_tag = ?", req.GroupTag).
		Updates(map[string]interface{}{
			"delivery_strategy": req.DeliveryStrategy,
		}).Error
}

func (l *DeliveryEngineLogic) LinkProductToProvider(req *types.LinkProductProviderRequest) error {
	prodUUID, err := uuid.Parse(req.ProductID)
	if err != nil {
		return xerr.NewErrCode(xerr.InvalidParams)
	}

	var product model.Product
	if err := l.svcCtx.DB.WithContext(l.ctx).Where("id = ?", prodUUID).First(&product).Error; err != nil {
		return xerr.NewErrCode(xerr.ProductNotFound)
	}

	updates := map[string]interface{}{
		"external_provider":   "third_party_api",
		"external_product_id": req.ExternalProductID,
	}

	if req.ExternalProductID > 0 {
		_, apiKey, _ := l.getInstantDeliveryConfig()
		if apiKey != "" {
			if prods, err := l.svcCtx.FulfillmentProvider.GetProducts(apiKey, false); err == nil {
				for _, ip := range prods {
					if ip.ID == req.ExternalProductID {
						updates["external_product_name"] = ip.Name
						updates["base_cost_usd"] = ip.PriceUSDT
						break
					}
				}
			}
		}
	} else {
		updates["external_product_name"] = ""
	}

	if req.DeliveryStrategy != "" {
		updates["delivery_strategy"] = req.DeliveryStrategy
	}

	return l.svcCtx.DB.WithContext(l.ctx).Model(&product).Updates(updates).Error
}

func (l *DeliveryEngineLogic) LinkProductToInsightX(req *types.LinkProductProviderRequest) error {
	return l.LinkProductToProvider(req)
}
