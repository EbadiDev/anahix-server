package admin

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/EbadiDev/anahix-server/internal/logic/provider/fulfillment"
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

func (l *DeliveryEngineLogic) ListBotProviderSettings() ([]types.BotProviderSettingDTO, error) {
	effectiveRate := l.getEffectiveRate()

	var settings []model.SystemSetting
	_ = l.svcCtx.DB.WithContext(l.ctx).Where("category = ?", "bot_providers").Find(&settings).Error

	settingMap := make(map[string]string)
	for _, s := range settings {
		settingMap[s.Key] = s.Value
	}

	adapters := l.svcCtx.FulfillmentManager.ListAdapters()
	result := make([]types.BotProviderSettingDTO, 0, len(adapters))

	for _, adapter := range adapters {
		key := adapter.GetProviderKey()
		nameFA := adapter.GetProviderNameFA()

		enabled := settingMap["provider_"+key+"_enabled"] == "true"
		apiKey := settingMap["provider_"+key+"_api_key"]
		baseURL := settingMap["provider_"+key+"_base_url"]
		if baseURL == "" {
			baseURL = adapter.GetDefaultBaseURL()
		}

		margin := 25.0
		if mStr, ok := settingMap["provider_"+key+"_margin"]; ok && mStr != "" {
			if mVal, err := strconv.ParseFloat(mStr, 64); err == nil && mVal >= 0 {
				margin = mVal
			}
		}

		// Fallback for InsightX legacy settings
		if key == fulfillment.ProviderInsightX && apiKey == "" {
			_, legacyKey, legacyMargin := l.getInstantDeliveryConfig()
			if legacyKey != "" {
				apiKey = legacyKey
				margin = legacyMargin
				enabled = true
			}
		}

		hasKey := apiKey != ""
		maskedKey := ""
		if len(apiKey) > 8 {
			maskedKey = apiKey[:4] + "••••••••" + apiKey[len(apiKey)-4:]
		} else if hasKey {
			maskedKey = "••••••••"
		}

		cfg := fulfillment.ProviderConfig{
			Key:           key,
			NameFA:        nameFA,
			Enabled:       enabled,
			APIKey:        apiKey,
			BaseURL:       baseURL,
			MarginPercent: margin,
		}

		balanceUSD := 0.0
		statusMsg := "کلید اتصال تنظیم نشده است"
		isConnected := false

		if hasKey {
			bal, err := adapter.GetBalance(http.DefaultClient, cfg)
			if err != nil {
				statusMsg = fmt.Sprintf("خطا: %v", err)
			} else {
				balanceUSD = bal.BalanceUSD
				statusMsg = "اتصال موفق و فعال"
				isConnected = true
			}
		}

		balanceToman := int64(balanceUSD * float64(effectiveRate))

		result = append(result, types.BotProviderSettingDTO{
			Key:           key,
			NameFA:        nameFA,
			Enabled:       enabled,
			HasAPIKey:     hasKey,
			APIKeyMasked:  maskedKey,
			BaseURL:       baseURL,
			MarginPercent: margin,
			BalanceUSD:    balanceUSD,
			BalanceToman:  balanceToman,
			StatusMessage: statusMsg,
			IsConnected:   isConnected,
		})
	}

	return result, nil
}

func (l *DeliveryEngineLogic) getEffectiveRate() int64 {
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
	rate, _, _ := exchangerate.GetDefaultClient().ResolveRate(ratePriority, usdRate)
	return rate
}

func (l *DeliveryEngineLogic) UpdateBotProviderSettings(req *types.UpdateBotProviderSettingsRequest) error {
	if req.Key == "" {
		return xerr.NewErrCode(xerr.InvalidParams)
	}

	upsertSetting := func(k, val string) error {
		var setting model.SystemSetting
		err := l.svcCtx.DB.WithContext(l.ctx).Where("`key` = ?", k).First(&setting).Error
		if err == nil {
			return l.svcCtx.DB.WithContext(l.ctx).Model(&setting).Updates(map[string]interface{}{
				"category":   "bot_providers",
				"value":      val,
				"updated_at": time.Now(),
			}).Error
		}
		newSetting := model.SystemSetting{
			Category:  "bot_providers",
			Key:       k,
			Value:     val,
			Type:      "string",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		return l.svcCtx.DB.WithContext(l.ctx).Create(&newSetting).Error
	}

	if req.Enabled != nil {
		enStr := "false"
		if *req.Enabled {
			enStr = "true"
		}
		if err := upsertSetting("provider_"+req.Key+"_enabled", enStr); err != nil {
			return err
		}
	}

	if req.APIKey != "" {
		if err := upsertSetting("provider_"+req.Key+"_api_key", req.APIKey); err != nil {
			return err
		}
	}

	if req.BaseURL != "" {
		if err := upsertSetting("provider_"+req.Key+"_base_url", req.BaseURL); err != nil {
			return err
		}
	}

	if req.MarginPercent != nil {
		mStr := fmt.Sprintf("%.2f", *req.MarginPercent)
		if err := upsertSetting("provider_"+req.Key+"_margin", mStr); err != nil {
			return err
		}
	}

	return nil
}

func (l *DeliveryEngineLogic) TestBotProviderConnection(req *types.TestBotProviderConnectionRequest) (*types.TestBotProviderConnectionResponse, error) {
	if req.Key == "" {
		return nil, xerr.NewErrCode(xerr.InvalidParams)
	}

	effectiveRate := l.getEffectiveRate()
	adapter, err := l.svcCtx.FulfillmentManager.GetAdapter(req.Key)
	if err != nil {
		return &types.TestBotProviderConnectionResponse{
			Success:     false,
			IsConnected: false,
			Message:     fmt.Sprintf("تامین‌کننده ربات %s یافت نشد", req.Key),
		}, nil
	}

	apiKey := req.APIKey
	if apiKey == "" {
		var setting model.SystemSetting
		if err := l.svcCtx.DB.WithContext(l.ctx).Where("`key` = ?", "provider_"+req.Key+"_api_key").First(&setting).Error; err == nil {
			apiKey = setting.Value
		}
	}

	if apiKey == "" {
		return &types.TestBotProviderConnectionResponse{
			Success:     false,
			IsConnected: false,
			Message:     "کلید API وارد نشده است. لطفاً ابتدا کلید را وارد نمایید.",
		}, nil
	}

	baseURL := req.BaseURL
	if baseURL == "" {
		baseURL = adapter.GetDefaultBaseURL()
	}

	cfg := fulfillment.ProviderConfig{
		Key:     req.Key,
		APIKey:  apiKey,
		BaseURL: baseURL,
	}

	bal, err := adapter.GetBalance(http.DefaultClient, cfg)
	if err != nil {
		return &types.TestBotProviderConnectionResponse{
			Success:     false,
			IsConnected: false,
			Message:     fmt.Sprintf("خطا در اتصال به ربات: %v", err),
		}, nil
	}

	balanceToman := int64(bal.BalanceUSD * float64(effectiveRate))
	return &types.TestBotProviderConnectionResponse{
		Success:      true,
		IsConnected:  true,
		BalanceUSD:   bal.BalanceUSD,
		BalanceToman: balanceToman,
		Message:      fmt.Sprintf("اتصال موفق و فعال! موجودی زنده: $%.2f USD (%d تومان)", bal.BalanceUSD, balanceToman),
	}, nil
}

func (l *DeliveryEngineLogic) GetBotCatalog(providerKey string, forceRefresh bool) ([]fulfillment.ProviderProductDTO, error) {
	effectiveRate := l.getEffectiveRate()

	var settings []model.SystemSetting
	_ = l.svcCtx.DB.WithContext(l.ctx).Where("category = ?", "bot_providers").Find(&settings).Error

	settingMap := make(map[string]string)
	for _, s := range settings {
		settingMap[s.Key] = s.Value
	}

	apiKey := settingMap["provider_"+providerKey+"_api_key"]
	baseURL := settingMap["provider_"+providerKey+"_base_url"]
	margin := 25.0
	if mStr, ok := settingMap["provider_"+providerKey+"_margin"]; ok && mStr != "" {
		if mVal, err := strconv.ParseFloat(mStr, 64); err == nil && mVal >= 0 {
			margin = mVal
		}
	}

	if providerKey == fulfillment.ProviderInsightX && apiKey == "" {
		_, legacyKey, legacyMargin := l.getInstantDeliveryConfig()
		if legacyKey != "" {
			apiKey = legacyKey
			margin = legacyMargin
		}
	}

	if apiKey == "" {
		return nil, fmt.Errorf("کلید API برای تامین‌کننده %s تنظیم نشده است", providerKey)
	}

	cfg := fulfillment.ProviderConfig{
		Key:           providerKey,
		APIKey:        apiKey,
		BaseURL:       baseURL,
		MarginPercent: margin,
	}

	items, err := l.svcCtx.FulfillmentManager.GetCatalog(providerKey, cfg, forceRefresh)
	if err != nil {
		return nil, err
	}

	// Fetch existing products in Anahix to mark which items are already linked
	var products []model.Product
	_ = l.svcCtx.DB.WithContext(l.ctx).Select("id, title_fa, external_provider, external_product_id, external_product_id_str").Find(&products).Error

	linkedMap := make(map[string][]string)
	for _, prod := range products {
		if prod.ExternalProvider == providerKey || (prod.ExternalProvider == "third_party_api" && providerKey == fulfillment.ProviderInsightX) {
			targetKey := prod.ExternalProductIDStr
			if targetKey == "" && prod.ExternalProductID > 0 {
				targetKey = strconv.FormatInt(prod.ExternalProductID, 10)
			}
			if targetKey != "" {
				linkedMap[targetKey] = append(linkedMap[targetKey], prod.TitleFA)
			}
		}
	}

	for i := range items {
		// Calculate suggested Toman price: CostUSD * Rate * (1 + Margin/100) rounded to 500
		costToman := float64(items[i].PriceUSD) * float64(effectiveRate) * (1.0 + margin/100.0)
		tomanRounded := int64(math.Ceil(costToman/500.0) * 500.0)
		items[i].PriceToman = tomanRounded

		if titles, linked := linkedMap[items[i].ID]; linked {
			items[i].IsLinked = true
			items[i].LinkedProducts = titles
		}
	}

	return items, nil
}

func (l *DeliveryEngineLogic) ImportBotProduct(req *types.ImportBotProductRequest) (*model.Product, error) {
	if req.TitleFA == "" || req.Slug == "" || req.ProviderKey == "" || req.ExternalProductID == "" {
		return nil, xerr.NewErrCode(xerr.InvalidParams)
	}

	// Check if slug already exists
	var existing model.Product
	err := l.svcCtx.DB.WithContext(l.ctx).Where("slug = ?", req.Slug).First(&existing).Error

	numericID, _ := strconv.ParseInt(req.ExternalProductID, 10, 64)

	deliveryType := model.DeliveryInstant
	if req.DeliveryStrategy == "manual_operator" {
		deliveryType = model.DeliveryManualTimed
	}

	strategy := req.DeliveryStrategy
	if strategy == "" {
		strategy = "third_party_instant"
	}

	stockDisplayMode := req.StockDisplayMode
	if stockDisplayMode == "" {
		stockDisplayMode = "show_exact"
	}

	// Calculate initial stock mode and count from supplier data
	stockMode := req.StockMode
	manualStock := req.ManualStock
	if stockMode == "" {
		if req.Unlimited {
			stockMode = "always_in_stock"
			manualStock = 999
		} else if req.InStock != nil && !*req.InStock {
			stockMode = "out_of_stock"
			manualStock = 0
		} else if req.Stock > 0 {
			stockMode = "manual"
			manualStock = int(req.Stock)
		} else if req.Stock == 0 && (req.InStock == nil || !*req.InStock) {
			stockMode = "out_of_stock"
			manualStock = 0
		} else {
			stockMode = "auto"
		}
	}
	outOfStockText := req.OutOfStockText
	if outOfStockText == "" || outOfStockText == "در حال تهیه و تامین مجدد" {
		outOfStockText = "ناموجود"
	}

	reqInput := req.RequiredCustomerInput
	if reqInput == "" {
		reqInput = "none"
	}

	payloadType := req.DeliveredPayloadType
	if payloadType == "" {
		payloadType = "license_code"
	}

	snoozeDays := req.PrePurchaseSnoozeDays
	if snoozeDays <= 0 {
		snoozeDays = 7
	}

	prod := model.Product{
		Title:                 req.TitleEN,
		TitleFA:               req.TitleFA,
		Slug:                  req.Slug,
		Description:           req.Description,
		Category:              model.ProductCategory(req.Category),
		DeliveryType:          deliveryType,
		PriceToman:            req.PriceToman,
		BaseCostUSD:           req.BaseCostUSD,
		IsActive:              false, // Default to inactive per user specification so admin can review and enrich media before publishing
		ImageURL:              req.ImageURL,
		IconURL:               req.IconURL,
		GroupTag:              req.GroupTag,
		GroupTitleFA:          req.GroupTitleFA,
		DeliveryStrategy:      strategy,
		ExternalProvider:      req.ProviderKey,
		ExternalProductID:     numericID,
		ExternalProductIDStr:  req.ExternalProductID,
		ExternalProductName:   req.ExternalProductName,
		RequiredCustomerInput: reqInput,
		DeliveredPayloadType:  payloadType,
		PrePurchaseNotice:     req.PrePurchaseNotice,
		PrePurchaseNoticeMode: req.PrePurchaseNoticeMode,
		PrePurchaseSnoozeDays: snoozeDays,
		StockDisplayMode:      stockDisplayMode,
		StockMode:             stockMode,
		ManualStock:           manualStock,
		OutOfStockText:        outOfStockText,
	}

	if err == nil {
		// Update existing
		prod.ID = existing.ID
		if err := l.svcCtx.DB.WithContext(l.ctx).Save(&prod).Error; err != nil {
			return nil, err
		}
		return &prod, nil
	}

	// Create new
	prod.ID = uuid.New()
	if err := l.svcCtx.DB.WithContext(l.ctx).Create(&prod).Error; err != nil {
		return nil, err
	}

	return &prod, nil
}

// Store Categories logic
func (l *DeliveryEngineLogic) ListStoreCategories() ([]model.StoreCategory, error) {
	var categories []model.StoreCategory
	err := l.svcCtx.DB.WithContext(l.ctx).
		Preload("Children").
		Where("parent_id IS NULL").
		Order("sort_order asc, created_at asc").
		Find(&categories).Error
	return categories, err
}

func (l *DeliveryEngineLogic) CreateStoreCategory(req *types.CreateStoreCategoryRequest) error {
	cat := model.StoreCategory{
		ID:          uuid.New(),
		TitleFA:     req.TitleFA,
		TitleEN:     req.TitleEN,
		Slug:        req.Slug,
		Description: req.Description,
		IconURL:     req.IconURL,
		SortOrder:   req.SortOrder,
		IsActive:    req.IsActive,
	}
	if req.ParentID != nil && *req.ParentID != "" {
		if pid, err := uuid.Parse(*req.ParentID); err == nil {
			cat.ParentID = &pid
		}
	}
	return l.svcCtx.DB.WithContext(l.ctx).Create(&cat).Error
}

func (l *DeliveryEngineLogic) UpdateStoreCategory(id string, req *types.UpdateStoreCategoryRequest) error {
	catUUID, err := uuid.Parse(id)
	if err != nil {
		return xerr.NewErrCode(xerr.InvalidParams)
	}

	updates := make(map[string]interface{})
	if req.TitleFA != "" {
		updates["title_fa"] = req.TitleFA
	}
	if req.TitleEN != "" {
		updates["title_en"] = req.TitleEN
	}
	if req.Slug != "" {
		updates["slug"] = req.Slug
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.IconURL != "" {
		updates["icon_url"] = req.IconURL
	}
	updates["sort_order"] = req.SortOrder
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	if req.ParentID != nil && *req.ParentID != "" {
		if pid, err := uuid.Parse(*req.ParentID); err == nil {
			updates["parent_id"] = pid
		}
	} else if req.ParentID != nil && *req.ParentID == "" {
		updates["parent_id"] = nil
	}

	return l.svcCtx.DB.WithContext(l.ctx).Model(&model.StoreCategory{}).Where("id = ?", catUUID).Updates(updates).Error
}

func (l *DeliveryEngineLogic) DeleteStoreCategory(id string) error {
	catUUID, err := uuid.Parse(id)
	if err != nil {
		return xerr.NewErrCode(xerr.InvalidParams)
	}
	return l.svcCtx.DB.WithContext(l.ctx).Where("id = ? OR parent_id = ?", catUUID, catUUID).Delete(&model.StoreCategory{}).Error
}
