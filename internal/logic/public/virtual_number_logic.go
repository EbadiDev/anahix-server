package public

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/EbadiDev/anahix-server/internal/logic/audit"
	"github.com/EbadiDev/anahix-server/internal/logic/provider/sms"
	"github.com/EbadiDev/anahix-server/internal/model"
	"github.com/EbadiDev/anahix-server/internal/svc"
	"github.com/EbadiDev/anahix-server/internal/types"
	"github.com/EbadiDev/anahix-server/pkg/exchangerate"
	"github.com/google/uuid"
)

type VirtualNumberLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	client *sms.Client
}

func NewVirtualNumberLogic(ctx context.Context, svcCtx *svc.ServiceContext) *VirtualNumberLogic {
	return &VirtualNumberLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		client: sms.NewClient(),
	}
}

func (l *VirtualNumberLogic) getSettings() (enabled bool, apiKey string, margin float64, usdRate int64, rubRate int64, defaultOp string) {
	var settings []model.SystemSetting
	l.svcCtx.DB.WithContext(l.ctx).Where("category IN (?, ?)", "virtual_numbers", "fivesim").Find(&settings)

	enabled = true
	margin = 25.0
	manualUSDRate := int64(95000)
	rubRate = 1050
	defaultOp = "any"
	ratePriority := "manual"

	for _, s := range settings {
		switch s.Key {
		case "vn_enabled", "fivesim_enabled":
			enabled = (s.Value == "true")
		case "vn_api_key", "fivesim_api_key":
			if apiKey == "" {
				apiKey = strings.TrimSpace(s.Value)
			}
		case "vn_margin_percent", "fivesim_margin_percent":
			if m, err := strconv.ParseFloat(s.Value, 64); err == nil && m > 0 {
				margin = m
			}
		case "vn_usd_rate", "fivesim_usd_rate":
			if u, err := strconv.ParseInt(s.Value, 10, 64); err == nil && u > 0 {
				manualUSDRate = u
			}
		case "vn_rub_rate", "fivesim_rub_rate":
			if r, err := strconv.ParseInt(s.Value, 10, 64); err == nil && r > 0 {
				rubRate = r
			}
		case "vn_default_operator", "fivesim_default_operator":
			if s.Value != "" {
				defaultOp = s.Value
			}
		case "vn_rate_priority", "fivesim_rate_priority":
			if s.Value != "" {
				ratePriority = s.Value
			}
		}
	}

	// Resolve dynamic rate based on priority (Option 1: Manual, Option 2: SwapWallet Live USDT)
	effectiveRate, _, _ := exchangerate.GetDefaultClient().ResolveRate(ratePriority, manualUSDRate)
	usdRate = effectiveRate

	return
}

var popularServiceSlugs = []string{
	"telegram", "openai", "whatsapp", "google", "claude", "discord",
	"instagram", "twitter", "apple", "microsoft", "steam", "tiktok",
	"netflix", "spotify", "binance", "uber", "paypal", "amazon",
	"facebook", "snapchat", "vk",
}

var ServiceMeta = map[string]struct {
	Name   string
	NameFA string
	Icon   string
}{
	"telegram":  {"Telegram", "تلگرام", "telegram"},
	"openai":    {"ChatGPT & OpenAI", "اوپن‌ای‌آی و چت‌جی‌پی‌تی", "openai"},
	"whatsapp":  {"WhatsApp", "واتساپ", "whatsapp"},
	"google":    {"Google / Gmail", "گوگل و جیمیل", "google"},
	"claude":    {"Claude AI", "کلود آنتروپیک", "claude"},
	"discord":   {"Discord", "دیسکورد", "discord"},
	"instagram": {"Instagram", "اینستاگرام", "instagram"},
	"twitter":   {"X (Twitter)", "توییتر و اکس", "twitter"},
	"apple":     {"Apple ID", "اپل آیدی", "apple"},
	"microsoft": {"Microsoft", "مایکروسافت", "microsoft"},
	"steam":     {"Steam", "استیم", "steam"},
	"tiktok":    {"TikTok", "تیک‌تاک", "tiktok"},
	"netflix":   {"Netflix", "نتفلیکس", "netflix"},
	"spotify":   {"Spotify", "اسپاتیفای", "spotify"},
	"binance":   {"Binance", "بایننس", "binance"},
	"uber":      {"Uber", "اوبر", "uber"},
	"paypal":    {"PayPal", "پی‌پال", "paypal"},
	"amazon":    {"Amazon", "آمازون", "amazon"},
	"facebook":  {"Facebook", "فیسبوک", "facebook"},
	"snapchat":  {"Snapchat", "اسنپ‌چت", "snapchat"},
	"vk":        {"VKontakte", "وی‌کی", "vk"},
	"line":      {"LINE", "لاین", "line"},
	"viber":     {"Viber", "وایبر", "viber"},
	"wechat":    {"WeChat", "وی‌چت", "wechat"},
	"yahoo":     {"Yahoo", "یاهو", "yahoo"},
}

var CountryMeta = map[string]struct {
	NameFA string
	Flag   string
	Code   string
}{
	"england":      {"انگلستان", "🇬🇧", "gb"},
	"usa":          {"ایالات متحده", "🇺🇸", "us"},
	"germany":      {"آلمان", "🇩🇪", "de"},
	"netherlands":  {"هلند", "🇳🇱", "nl"},
	"france":       {"فرانسه", "🇫🇷", "fr"},
	"sweden":       {"سوئد", "🇸🇪", "se"},
	"canada":       {"کانادا", "🇨🇦", "ca"},
	"spain":        {"اسپانیا", "🇪🇸", "es"},
	"indonesia":    {"اندونزی", "🇮🇩", "id"},
	"malaysia":     {"مالزی", "🇲🇾", "my"},
	"thailand":     {"تایلند", "🇹🇭", "th"},
	"brazil":       {"برزیل", "🇧🇷", "br"},
	"india":        {"هند", "🇮🇳", "in"},
	"turkey":       {"ترکیه", "🇹🇷", "tr"},
	"uae":          {"امارات", "🇦🇪", "ae"},
	"poland":       {"لهستان", "🇵🇱", "pl"},
	"georgia":      {"گرجستان", "🇬🇪", "ge"},
	"romania":      {"رومانی", "🇷🇴", "ro"},
	"russia":       {"روسیه", "🇷🇺", "ru"},
	"ukraine":      {"اوکراین", "🇺🇦", "ua"},
	"philippines":  {"فیلیپین", "🇵🇭", "ph"},
	"vietnam":      {"ویتنام", "🇻🇳", "vn"},
	"singapore":    {"سنگاپور", "🇸🇬", "sg"},
	"australia":    {"استرالیا", "🇦🇺", "au"},
	"argentina":    {"آرژانتین", "🇦🇷", "ar"},
	"colombia":     {"کلمبیا", "🇨🇴", "co"},
	"southafrica":  {"آفریقای جنوبی", "🇿🇦", "za"},
	"egypt":        {"مصر", "🇪🇬", "eg"},
	"pakistan":     {"پاکستان", "🇵🇰", "pk"},
	"bangladesh":   {"بنگلادش", "🇧🇩", "bd"},
	"mexico":       {"مکزیک", "🇲🇽", "mx"},
	"italy":        {"ایتالیا", "🇮🇹", "it"},
	"portugal":     {"پرتغال", "🇵🇹", "pt"},
	"czech":        {"جمهوری چک", "🇨🇿", "cz"},
	"austria":      {"اتریش", "🇦🇹", "at"},
	"belgium":      {"بلژیک", "🇧🇪", "be"},
	"denmark":      {"دانمارک", "🇩🇰", "dk"},
	"finland":      {"فنلاند", "🇫🇮", "fi"},
	"norway":       {"نروژ", "🇳🇴", "no"},
	"greece":       {"یونان", "🇬🇷", "gr"},
	"switzerland":  {"سوئیس", "🇨🇭", "ch"},
	"ireland":      {"ایرلند", "🇮🇪", "ie"},
	"newzealand":   {"نیوزیلند", "🇳🇿", "nz"},
	"japan":        {"ژاپن", "🇯🇵", "jp"},
	"korea":        {"کره جنوبی", "🇰🇷", "kr"},
	"china":        {"چین", "🇨🇳", "cn"},
	"hongkong":     {"هنگ‌کنگ", "🇭🇰", "hk"},
	"taiwan":       {"تایوان", "🇹🇼", "tw"},
	"israel":       {"اسرائیل", "🇮🇱", "il"},
	"saudiarabia":  {"عربستان سعودی", "🇸🇦", "sa"},
	"qatar":        {"قطر", "🇶🇦", "qa"},
	"kuwait":       {"کویت", "🇰🇼", "kw"},
	"oman":         {"عمان", "🇴🇲", "om"},
	"morocco":      {"مراکش", "🇲🇦", "ma"},
	"algeria":      {"الجزایر", "🇩🇿", "dz"},
	"tunisia":      {"تونس", "🇹🇳", "tn"},
	"nigeria":      {"نیجریه", "🇳🇬", "ng"},
	"kenya":        {"کنیا", "🇰🇪", "ke"},
	"ghana":        {"غنا", "🇬🇭", "gh"},
	"kazakhstan":   {"قزاقستان", "🇰🇿", "kz"},
	"uzbekistan":   {"ازبکستان", "🇺🇿", "uz"},
	"kyrgyzstan":   {"قرقیزستان", "🇰🇬", "kg"},
	"tajikistan":   {"تاجیکستان", "🇹🇯", "tj"},
	"armenia":      {"ارمنستان", "🇦🇲", "am"},
	"azerbaijan":   {"آذربایجان", "🇦🇿", "az"},
	"moldova":      {"مولداوی", "🇲🇩", "md"},
	"lithuania":    {"لیتوانی", "🇱🇹", "lt"},
	"latvia":       {"لتونی", "🇱🇻", "lv"},
	"estonia":      {"استونی", "🇪🇪", "ee"},
	"hungary":      {"مجارستان", "🇭🇺", "hu"},
	"slovakia":     {"اسلواکی", "🇸🇰", "sk"},
	"slovenia":     {"اسلوونی", "🇸🇮", "si"},
	"croatia":      {"کرواسی", "🇭🇷", "hr"},
	"serbia":       {"صربستان", "🇷🇸", "rs"},
	"bulgaria":     {"بلغارستان", "🇧🇬", "bg"},
	"cyprus":       {"قبرس", "🇨🇾", "cy"},
	"chile":        {"شیلی", "🇨🇱", "cl"},
	"peru":         {"پرو", "🇵🇪", "pe"},
	"ecuador":      {"اکوادور", "🇪🇨", "ec"},
	"venezuela":    {"ونزوئلا", "🇻🇪", "ve"},
	"cambodia":     {"کامبوج", "🇰🇭", "kh"},
	"myanmar":      {"میانمار", "🇲🇲", "mm"},
	"laos":         {"لائوس", "🇱🇦", "la"},
	"nepal":        {"نپال", "🇳🇵", "np"},
	"srilanka":     {"سری‌لانکا", "🇱🇰", "lk"},
}

// ListServices returns all active services dynamically fetched from 5SIM
func (l *VirtualNumberLogic) ListServices() ([]types.VirtualNumberServiceItem, error) {
	_, _, margin, usdRate, _, _ := l.getSettings()

	rawProducts, err := l.client.GetProducts()
	if err != nil || len(rawProducts) == 0 {
		// Fallback minimal list if 5sim guest API temporarily fails
		var fallback []types.VirtualNumberServiceItem
		for _, slug := range popularServiceSlugs {
			meta, ok := ServiceMeta[slug]
			if !ok {
				continue
			}
			priceUSD := 0.2
			priceToman := int64(math.Round(priceUSD * float64(usdRate) * (1.0 + margin/100.0)))
			priceToman = ((priceToman + 499) / 500) * 500
			fallback = append(fallback, types.VirtualNumberServiceItem{
				Slug:          slug,
				Name:          meta.Name,
				NameFA:        meta.NameFA,
				Category:      "activation",
				TotalQty:      100000,
				MinPriceUSD:   priceUSD,
				MinPriceToman: priceToman,
				Icon:          meta.Icon,
				IsPopular:     true,
			})
		}
		return fallback, nil
	}

	popularMap := make(map[string]int)
	for idx, s := range popularServiceSlugs {
		popularMap[s] = idx + 1
	}

	result := make([]types.VirtualNumberServiceItem, 0, len(rawProducts))
	for slug, p := range rawProducts {
		if p.Qty <= 0 && popularMap[slug] == 0 {
			continue // skip empty non-popular services
		}

		name := strings.Title(strings.ReplaceAll(slug, "_", " "))
		nameFA := name
		icon := "message-square"
		isPopular := false

		if meta, ok := ServiceMeta[slug]; ok {
			name = meta.Name
			nameFA = meta.NameFA
			icon = meta.Icon
			isPopular = true
		}

		priceUSD := p.Price
		if priceUSD <= 0 {
			priceUSD = 0.05
		}

		priceToman := int64(math.Round(priceUSD * float64(usdRate) * (1.0 + margin/100.0)))
		priceToman = ((priceToman + 499) / 500) * 500
		if priceToman < 5000 {
			priceToman = 5000
		}

		result = append(result, types.VirtualNumberServiceItem{
			Slug:          slug,
			Name:          name,
			NameFA:        nameFA,
			Category:      p.Category,
			TotalQty:      p.Qty,
			MinPriceUSD:   priceUSD,
			MinPriceToman: priceToman,
			Icon:          icon,
			IsPopular:     isPopular,
		})
	}

	// Sort popular first, then by quantity
	sort.Slice(result, func(i, j int) bool {
		pi := popularMap[result[i].Slug]
		pj := popularMap[result[j].Slug]
		if pi > 0 && pj > 0 {
			return pi < pj
		}
		if pi > 0 {
			return true
		}
		if pj > 0 {
			return false
		}
		if result[i].TotalQty != result[j].TotalQty {
			return result[i].TotalQty > result[j].TotalQty
		}
		return result[i].Slug < result[j].Slug
	})

	return result, nil
}

// ListCountriesForService returns all countries with operators for the selected service
func (l *VirtualNumberLogic) ListCountriesForService(service string) ([]types.VirtualNumberCountryItem, error) {
	service = strings.ToLower(strings.TrimSpace(service))
	if service == "" {
		return nil, errors.New("سرویس انتخابی نامعتبر است")
	}

	_, _, margin, usdRate, _, defaultOp := l.getSettings()

	countriesMap, err := l.client.GetPrices(service)
	if err != nil {
		return nil, err
	}

	var result []types.VirtualNumberCountryItem
	for country, ops := range countriesMap {
		if len(ops) == 0 {
			continue
		}

		var opItems []types.VirtualNumberOperatorItem
		totalStock := 0
		var bestRate float64
		minPriceToman := int64(math.MaxInt64)
		var minPriceUSD float64
		var cheapestOp string

		for opName, opInfo := range ops {
			costUSD := opInfo.Cost
			if costUSD <= 0 {
				costUSD = 0.1
			}

			priceToman := int64(math.Round(costUSD * float64(usdRate) * (1.0 + margin/100.0)))
			priceToman = ((priceToman + 499) / 500) * 500
			if priceToman < 5000 {
				priceToman = 5000
			}

			totalStock += opInfo.Count
			if opInfo.Rate > bestRate {
				bestRate = opInfo.Rate
			}

			if opInfo.Count > 0 && (cheapestOp == "" || priceToman < minPriceToman) {
				minPriceToman = priceToman
				minPriceUSD = costUSD
				cheapestOp = opName
			}

			opItems = append(opItems, types.VirtualNumberOperatorItem{
				Operator:   opName,
				CostUSD:    costUSD,
				PriceToman: priceToman,
				Count:      opInfo.Count,
				Rate:       opInfo.Rate,
			})
		}

		if minPriceToman == math.MaxInt64 {
			if len(opItems) > 0 {
				minPriceToman = opItems[0].PriceToman
				minPriceUSD = opItems[0].CostUSD
				cheapestOp = opItems[0].Operator
			} else {
				minPriceToman = 15000
				minPriceUSD = 0.15
				cheapestOp = defaultOp
			}
		}

		// Sort operators: stock > 0 first, then by price ascending
		sort.Slice(opItems, func(i, j int) bool {
			if opItems[i].Count > 0 && opItems[j].Count == 0 {
				return true
			}
			if opItems[i].Count == 0 && opItems[j].Count > 0 {
				return false
			}
			return opItems[i].PriceToman < opItems[j].PriceToman
		})

		countryNameFA := strings.Title(country)
		countryFlag := "🌐"
		countryCode := ""
		flagURL := ""
		if meta, ok := CountryMeta[country]; ok {
			countryNameFA = meta.NameFA
			countryFlag = meta.Flag
			countryCode = meta.Code
			if meta.Code != "" {
				flagURL = fmt.Sprintf("https://flagcdn.com/w40/%s.png", meta.Code)
			}
		}

		result = append(result, types.VirtualNumberCountryItem{
			Country:            country,
			CountryNameFA:      countryNameFA,
			CountryFlag:        countryFlag,
			CountryCode:        countryCode,
			FlagURL:            flagURL,
			TotalStock:         totalStock,
			BestRate:           bestRate,
			CheapestPriceToman: minPriceToman,
			CheapestPriceUSD:   minPriceUSD,
			DefaultOperator:    cheapestOp,
			Operators:          opItems,
		})
	}

	// Sort countries: available stock > 0 first, then by cheapest price ascending
	sort.Slice(result, func(i, j int) bool {
		if result[i].TotalStock > 0 && result[j].TotalStock == 0 {
			return true
		}
		if result[i].TotalStock == 0 && result[j].TotalStock > 0 {
			return false
		}
		return result[i].CheapestPriceToman < result[j].CheapestPriceToman
	})

	return result, nil
}

// GetCatalog maintains backward compatibility for existing consumers
func (l *VirtualNumberLogic) GetCatalog() ([]types.VirtualNumberCatalogItem, error) {
	services, err := l.ListServices()
	if err != nil {
		return nil, err
	}

	// Take top 6 popular services
	var topServices []types.VirtualNumberServiceItem
	for _, s := range services {
		if s.IsPopular {
			topServices = append(topServices, s)
		}
		if len(topServices) >= 6 {
			break
		}
	}

	topCountries := []string{"england", "usa", "germany", "netherlands", "indonesia", "france"}
	var catalog []types.VirtualNumberCatalogItem

	for _, cCode := range topCountries {
		cMeta, ok := CountryMeta[cCode]
		if !ok {
			cMeta = struct{ NameFA, Flag, Code string }{strings.Title(cCode), "🌐", ""}
		}
		for _, s := range topServices {
			catalog = append(catalog, types.VirtualNumberCatalogItem{
				Service:            s.Slug,
				ServiceNameFA:      s.NameFA,
				Country:            cCode,
				CountryNameFA:      cMeta.NameFA,
				CountryFlag:        cMeta.Flag,
				BaseCost:           s.MinPriceUSD,
				PriceToman:         s.MinPriceToman,
				OriginalPriceToman: int64(float64(s.MinPriceToman) * 1.25),
				AvailableStock:     150,
				SuccessRatePercent: 98,
				Icon:               s.Icon,
			})
		}
	}

	return catalog, nil
}

func (l *VirtualNumberLogic) CreateOrder(req *types.BuyVirtualNumberRequest) (*types.VirtualNumberOrderResponse, error) {
	enabled, apiKey, margin, usdRate, _, defaultOp := l.getSettings()
	if !enabled {
		return nil, errors.New("فروش شماره مجازی موقتاً توسط مدیریت غیرفعال شده است")
	}

	country := strings.ToLower(strings.TrimSpace(req.Country))
	service := strings.ToLower(strings.TrimSpace(req.Service))
	operator := strings.ToLower(strings.TrimSpace(req.Operator))
	if operator == "" {
		operator = defaultOp
	}

	// Call 5SIM user API to buy activation
	ord5Sim, err := l.client.BuyActivation(apiKey, country, operator, service)
	if err != nil {
		return nil, fmt.Errorf("خطا در خرید شماره از سرویس: %w", err)
	}

	// Calculate customer price in Toman using live 5SIM USD price
	baseUSD := ord5Sim.Price
	if baseUSD <= 0 {
		baseUSD = 0.2
	}
	costToman := int64(math.Round(baseUSD * float64(usdRate)))
	priceToman := int64(math.Round(float64(costToman) * (1.0 + margin/100.0)))
	priceToman = ((priceToman + 499) / 500) * 500
	if priceToman < 5000 {
		priceToman = 5000
	}

	orderID := uuid.New()
	orderNo := fmt.Sprintf("VNUM-%d", time.Now().UnixNano()%900000+100000)

	cMeta, hasC := CountryMeta[country]
	if !hasC {
		cMeta = struct{ NameFA, Flag, Code string }{strings.Title(country), "🌐", ""}
	}

	sMeta, hasS := ServiceMeta[service]
	sNameFA := strings.Title(service)
	if hasS {
		sNameFA = sMeta.NameFA
	}

	expiresAt := time.Now().Add(15 * time.Minute)
	if ord5Sim.Expires != "" {
		if t, err := time.Parse(time.RFC3339, ord5Sim.Expires); err == nil {
			expiresAt = t
		}
	}

	order := model.VirtualNumberOrder{
		ID:              orderID,
		OrderNo:         orderNo,
		ProviderOrderID: ord5Sim.ID,
		FiveSimOrderID:  ord5Sim.ID,
		Phone:           ord5Sim.Phone,
		Country:        country,
		CountryNameFA:  cMeta.NameFA,
		CountryFlag:    cMeta.Flag,
		Service:        service,
		ServiceNameFA:  sNameFA,
		Operator:       ord5Sim.Operator,
		BaseCost:       baseUSD,
		PriceToman:     priceToman,
		Status:         model.VNumStatusPending,
		CustomerPhone:  req.CustomerPhone,
		ExpiresAt:      expiresAt,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := l.svcCtx.DB.WithContext(l.ctx).Create(&order).Error; err != nil {
		return nil, fmt.Errorf("خطا در ثبت سفارش: %w", err)
	}

	// Record transaction log
	_ = audit.RecordTransaction(l.ctx, l.svcCtx.DB, &model.Transaction{
		OrderID:         &order.ID,
		OrderType:       model.TxTypeVirtualNumber,
		CustomerPhone:   order.CustomerPhone,
		AmountToman:     order.PriceToman,
		AmountUSD:       order.BaseCost,
		PaymentPlatform: model.PlatformBalance,
		Status:          model.TxStatusSuccess,
		Description:     fmt.Sprintf("خرید شماره مجازی %s (%s) برای %s", order.CountryNameFA, order.Phone, order.ServiceNameFA),
	})

	remSeconds := int(time.Until(order.ExpiresAt).Seconds())
	if remSeconds < 0 {
		remSeconds = 0
	}

	return &types.VirtualNumberOrderResponse{
		ID:               order.ID.String(),
		OrderNo:          order.OrderNo,
		Phone:            order.Phone,
		Country:          order.Country,
		CountryNameFA:    order.CountryNameFA,
		CountryFlag:      order.CountryFlag,
		Service:          order.Service,
		ServiceNameFA:    order.ServiceNameFA,
		PriceToman:       order.PriceToman,
		Status:           string(order.Status),
		SMSCode:          order.SMSCode,
		SMSText:          order.SMSText,
		ExpiresAt:        order.ExpiresAt,
		SecondsRemaining: remSeconds,
		CreatedAt:        order.CreatedAt,
	}, nil
}

func (l *VirtualNumberLogic) GetOrderDetail(orderIDStr string) (*types.VirtualNumberOrderResponse, error) {
	_, apiKey, _, _, _, _ := l.getSettings()

	var order model.VirtualNumberOrder
	var err error
	if parsedUUID, parseErr := uuid.Parse(orderIDStr); parseErr == nil {
		err = l.svcCtx.DB.WithContext(l.ctx).Where("id = ?", parsedUUID).First(&order).Error
	} else {
		err = l.svcCtx.DB.WithContext(l.ctx).Where("order_no = ?", orderIDStr).First(&order).Error
	}
	if err != nil {
		return nil, errors.New("سفارش شماره مجازی یافت نشد")
	}

	// If pending and not expired, poll 5SIM for SMS
	if order.Status == model.VNumStatusPending {
		if time.Now().After(order.ExpiresAt) {
			order.Status = model.VNumStatusTimeout
			l.svcCtx.DB.WithContext(l.ctx).Model(&order).Update("status", model.VNumStatusTimeout)
		} else {
			pID := order.ProviderOrderID
			if pID == 0 {
				pID = order.FiveSimOrderID
			}
			chk, chkErr := l.client.CheckOrder(apiKey, pID, order.CreatedAt, order.Service)
			if chkErr == nil && chk != nil {
				if len(chk.SMS) > 0 {
					order.Status = model.VNumStatusReceived
					order.SMSCode = chk.SMS[0].Code
					order.SMSText = chk.SMS[0].Text
					l.svcCtx.DB.WithContext(l.ctx).Model(&order).Updates(map[string]interface{}{
						"status":   model.VNumStatusReceived,
						"sms_code": order.SMSCode,
						"sms_text": order.SMSText,
					})
				} else if chk.Status == "CANCELED" || chk.Status == "TIMEOUT" {
					order.Status = model.VirtualNumberStatus(chk.Status)
					l.svcCtx.DB.WithContext(l.ctx).Model(&order).Update("status", order.Status)
				}
			}
		}
	}

	remSeconds := int(time.Until(order.ExpiresAt).Seconds())
	if remSeconds < 0 {
		remSeconds = 0
	}

	return &types.VirtualNumberOrderResponse{
		ID:               order.ID.String(),
		OrderNo:          order.OrderNo,
		Phone:            order.Phone,
		Country:          order.Country,
		CountryNameFA:    order.CountryNameFA,
		CountryFlag:      order.CountryFlag,
		Service:          order.Service,
		ServiceNameFA:    order.ServiceNameFA,
		PriceToman:       order.PriceToman,
		Status:           string(order.Status),
		SMSCode:          order.SMSCode,
		SMSText:          order.SMSText,
		ExpiresAt:        order.ExpiresAt,
		SecondsRemaining: remSeconds,
		CreatedAt:        order.CreatedAt,
	}, nil
}

func (l *VirtualNumberLogic) CancelOrder(orderIDStr string) error {
	_, apiKey, _, _, _, _ := l.getSettings()

	var order model.VirtualNumberOrder
	var err error
	if parsedUUID, parseErr := uuid.Parse(orderIDStr); parseErr == nil {
		err = l.svcCtx.DB.WithContext(l.ctx).Where("id = ?", parsedUUID).First(&order).Error
	} else {
		err = l.svcCtx.DB.WithContext(l.ctx).Where("order_no = ?", orderIDStr).First(&order).Error
	}
	if err != nil {
		return errors.New("سفارش یافت نشد")
	}

	if order.Status != model.VNumStatusPending {
		return errors.New("فقط سفارش‌های در انتظار دریافت پیامک قابل لغو و استرداد هستند")
	}

	pID := order.ProviderOrderID
	if pID == 0 {
		pID = order.FiveSimOrderID
	}
	_ = l.client.CancelOrder(apiKey, pID)
	order.Status = model.VNumStatusCanceled
	err = l.svcCtx.DB.WithContext(l.ctx).Model(&order).Update("status", model.VNumStatusCanceled).Error
	if err == nil {
		_ = audit.RecordTransaction(l.ctx, l.svcCtx.DB, &model.Transaction{
			OrderID:         &order.ID,
			OrderType:       model.TxTypeRefund,
			CustomerPhone:   order.CustomerPhone,
			AmountToman:     order.PriceToman,
			AmountUSD:       order.BaseCost,
			PaymentPlatform: model.PlatformBalance,
			Status:          model.TxStatusRefunded,
			Description:     fmt.Sprintf("استرداد وجه سفارش شماره مجازی #%s (%s)", order.OrderNo, order.ServiceNameFA),
		})
	}
	return err
}

func (l *VirtualNumberLogic) FinishOrder(orderIDStr string) error {
	_, apiKey, _, _, _, _ := l.getSettings()

	var order model.VirtualNumberOrder
	var err error
	if parsedUUID, parseErr := uuid.Parse(orderIDStr); parseErr == nil {
		err = l.svcCtx.DB.WithContext(l.ctx).Where("id = ?", parsedUUID).First(&order).Error
	} else {
		err = l.svcCtx.DB.WithContext(l.ctx).Where("order_no = ?", orderIDStr).First(&order).Error
	}
	if err != nil {
		return errors.New("سفارش یافت نشد")
	}

	pID := order.ProviderOrderID
	if pID == 0 {
		pID = order.FiveSimOrderID
	}
	_ = l.client.FinishOrder(apiKey, pID)
	order.Status = model.VNumStatusFinished
	return l.svcCtx.DB.WithContext(l.ctx).Model(&order).Update("status", model.VNumStatusFinished).Error
}
