package public

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/EbadiDev/anahix-server/internal/model"
	"github.com/EbadiDev/anahix-server/internal/svc"
	"github.com/EbadiDev/anahix-server/internal/types"
	"github.com/EbadiDev/anahix-server/pkg/xerr"
)

type CatalogLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCatalogLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CatalogLogic {
	return &CatalogLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CatalogLogic) ListProducts() ([]types.ProductListItem, error) {
	var products []model.Product
	err := l.svcCtx.DB.WithContext(l.ctx).
		Where("is_active = ?", true).
		Order("created_at desc").
		Find(&products).Error
	if err != nil {
		return nil, xerr.NewErrCode(xerr.DatabaseError)
	}

	result := make([]types.ProductListItem, len(products))
	for i, p := range products {
		var availableStock int64 = 0
		if p.StockMode == "always_in_stock" {
			availableStock = 999
		} else if p.StockMode == "manual" {
			availableStock = int64(p.ManualStock)
		} else if p.StockMode == "out_of_stock" {
			availableStock = 0
		} else if p.DeliveryType == model.DeliveryInstant {
			l.svcCtx.DB.WithContext(l.ctx).
				Model(&model.InventoryItem{}).
				Where("product_id = ? AND status = ?", p.ID, model.InventoryAvailable).
				Count(&availableStock)
		} else {
			availableStock = 99
		}

		var features []string
		if p.FeaturesJSON != "" {
			_ = json.Unmarshal([]byte(p.FeaturesJSON), &features)
		}

		var flashSaleEndStr *string
		if p.FlashSaleEndAt != nil {
			s := p.FlashSaleEndAt.Format(time.RFC3339)
			flashSaleEndStr = &s
		}
		var discountPercent int = 0
		var origPrice int64 = p.PriceToman
		if p.IsFlashSale && p.FlashSalePriceToman > 0 && p.PriceToman > p.FlashSalePriceToman {
			discountPercent = int(float64(p.PriceToman-p.FlashSalePriceToman) / float64(p.PriceToman) * 100)
			origPrice = p.PriceToman
		}

		result[i] = types.ProductListItem{
			ID:                  p.ID.String(),
			Title:               p.Title,
			TitleFA:             p.TitleFA,
			Slug:                p.Slug,
			Category:            string(p.Category),
			DeliveryType:        string(p.DeliveryType),
			PriceToman:          p.PriceToman,
			BaseCostUSD:         p.BaseCostUSD,
			SLAMinMinutes:       p.SLAMinMinutes,
			SLAMaxMinutes:       p.SLAMaxMinutes,
			SLADisplayText:      p.SLADisplayText,
			OperatingHours:      p.OperatingHours,
			IconURL:             p.IconURL,
			BannerURL:           p.BannerURL,
			ImageURL:            p.ImageURL,
			IsFlashSale:         p.IsFlashSale,
			FlashSalePriceToman: p.FlashSalePriceToman,
			FlashSaleEndAt:      flashSaleEndStr,
			FlashSaleStock:      p.FlashSaleStock,
			FlashSaleSold:       p.FlashSaleSold,
			OriginalPriceToman:  origPrice,
			DiscountPercent:     discountPercent,
			IsBestSeller:        p.IsBestSeller,
			IsFeatured:          p.IsFeatured,
			RatingAverage:       p.RatingAverage,
			SalesCount:          p.SalesCount,
			BadgesJSON:          p.BadgesJSON,
			AvailableStock:      availableStock,
			IsActive:            p.IsActive,
			GroupTag:            p.GroupTag,
			GroupTitleFA:        p.GroupTitleFA,
			StockDisplayMode:    p.StockDisplayMode,
			StockMode:           p.StockMode,
			OutOfStockText:      p.OutOfStockText,
			Features:            features,
		}
	}

	return result, nil
}

func (l *CatalogLogic) GetProductDetail(slug string) (*types.ProductDetail, error) {
	var p model.Product
	err := l.svcCtx.DB.WithContext(l.ctx).
		Where("slug = ? AND is_active = ?", slug, true).
		First(&p).Error
	if err != nil {
		return nil, xerr.NewErrCode(xerr.ProductNotFound)
	}

	var availableStock int64 = 0
	if p.StockMode == "always_in_stock" {
		availableStock = 999
	} else if p.StockMode == "manual" {
		availableStock = int64(p.ManualStock)
	} else if p.StockMode == "out_of_stock" {
		availableStock = 0
	} else if p.DeliveryType == model.DeliveryInstant {
		l.svcCtx.DB.WithContext(l.ctx).
			Model(&model.InventoryItem{}).
			Where("product_id = ? AND status = ?", p.ID, model.InventoryAvailable).
			Count(&availableStock)
	} else {
		availableStock = 99
	}

	// 1. Features
	var features []string
	if p.FeaturesJSON != "" {
		_ = json.Unmarshal([]byte(p.FeaturesJSON), &features)
	}
	if len(features) == 0 {
		features = defaultFeaturesForProduct(&p)
	}

	// 2. Plans / Variants
	var plans []types.ProductPlan
	if p.PlansJSON != "" {
		_ = json.Unmarshal([]byte(p.PlansJSON), &plans)
	}
	if len(plans) == 0 {
		plans = defaultPlansForProduct(&p)
	}

	// 3. Comparison Table
	var comparison *types.ProductComparison
	if p.ComparisonJSON != "" {
		var comp types.ProductComparison
		if err := json.Unmarshal([]byte(p.ComparisonJSON), &comp); err == nil && len(comp.Columns) > 0 {
			comparison = &comp
		}
	}
	if comparison == nil {
		comparison = defaultComparisonForProduct(&p)
	}

	// 4. Activation Guide
	var guide []types.ActivationStep
	if p.GuideJSON != "" {
		_ = json.Unmarshal([]byte(p.GuideJSON), &guide)
	}
	if len(guide) == 0 {
		guide = defaultGuideForProduct(&p)
	}

	// 5. FAQs
	var faqs []types.ProductFAQ
	if p.FaqJSON != "" {
		_ = json.Unmarshal([]byte(p.FaqJSON), &faqs)
	}
	if len(faqs) == 0 {
		faqs = defaultFaqsForProduct(&p)
	}

	// 6. SEO
	metaTitle := p.MetaTitle
	if metaTitle == "" {
		metaTitle = fmt.Sprintf("خرید %s | تحویل قانونی و آنی", p.TitleFA)
	}
	metaDesc := p.MetaDescription
	if metaDesc == "" {
		if p.Description != "" {
			metaDesc = p.Description
		} else {
			metaDesc = fmt.Sprintf("خرید اشتراک قانونی %s با بهترین قیمت تومانی، تحویل فوری، گارانتی عدم قطعی و پشتیبانی ۲۴ ساعته.", p.TitleFA)
		}
	}
	ogImg := p.OgImageURL
	if ogImg == "" {
		if p.BannerURL != "" {
			ogImg = p.BannerURL
		} else if p.IconURL != "" {
			ogImg = p.IconURL
		} else {
			ogImg = "/assets/hero.png"
		}
	}

	var detailFlashSaleEnd *string
	if p.FlashSaleEndAt != nil {
		s := p.FlashSaleEndAt.Format(time.RFC3339)
		detailFlashSaleEnd = &s
	}
	var detailDiscountPercent int = 0
	var detailOrigPrice int64 = p.PriceToman
	if p.IsFlashSale && p.FlashSalePriceToman > 0 && p.PriceToman > p.FlashSalePriceToman {
		detailDiscountPercent = int(float64(p.PriceToman-p.FlashSalePriceToman) / float64(p.PriceToman) * 100)
		detailOrigPrice = p.PriceToman
	}

	detail := &types.ProductDetail{

		ProductListItem: types.ProductListItem{
			ID:                  p.ID.String(),
			Title:               p.Title,
			TitleFA:             p.TitleFA,
			Slug:                p.Slug,
			Category:            string(p.Category),
			DeliveryType:        string(p.DeliveryType),
			PriceToman:          p.PriceToman,
			BaseCostUSD:         p.BaseCostUSD,
			SLAMinMinutes:       p.SLAMinMinutes,
			SLAMaxMinutes:       p.SLAMaxMinutes,
			SLADisplayText:      p.SLADisplayText,
			OperatingHours:      p.OperatingHours,
			IconURL:             p.IconURL,
			BannerURL:           p.BannerURL,
			ImageURL:            p.ImageURL,
			IsFlashSale:         p.IsFlashSale,
			FlashSalePriceToman: p.FlashSalePriceToman,
			FlashSaleEndAt:      detailFlashSaleEnd,
			FlashSaleStock:      p.FlashSaleStock,
			FlashSaleSold:       p.FlashSaleSold,
			OriginalPriceToman:  detailOrigPrice,
			DiscountPercent:     detailDiscountPercent,
			IsBestSeller:        p.IsBestSeller,
			IsFeatured:          p.IsFeatured,
			RatingAverage:       p.RatingAverage,
			SalesCount:          p.SalesCount,
			BadgesJSON:          p.BadgesJSON,
			AvailableStock:      availableStock,
			IsActive:            p.IsActive,
			GroupTag:            p.GroupTag,
			GroupTitleFA:        p.GroupTitleFA,
			Features:            features,
		},
		Description:      p.Description,
		Instructions:     p.Instructions,
		DeliveryStrategy: p.DeliveryStrategy,
		SEO: types.ProductSEO{
			MetaTitle:       metaTitle,
			MetaDescription: metaDesc,
			OgImageURL:      ogImg,
		},
		Plans:           plans,
		ComparisonTable: comparison,
		ActivationGuide: guide,
		Faqs:            faqs,
		Features:        features,

		// Raw strings for Admin editing
		MetaTitle:       p.MetaTitle,
		MetaDescription: p.MetaDescription,
		OgImageURL:      p.OgImageURL,
		PlansJSON:       p.PlansJSON,
		ComparisonJSON:  p.ComparisonJSON,
		GuideJSON:       p.GuideJSON,
		FaqJSON:         p.FaqJSON,
		FeaturesJSON:    p.FeaturesJSON,
	}

	return detail, nil
}

func defaultFeaturesForProduct(p *model.Product) []string {
	if p.DeliveryType == model.DeliveryInstant {
		return []string{
			"تحویل آنی بلافاصله پس از پرداخت",
			"فعال‌سازی درجا و بدون نیاز به ارسال پسورد",
			"تضمین ۱۰۰٪ اصالت لایسنس و عدم قطعی",
			"پشتیبانی شبانه‌روزی آناهیکس",
		}
	}
	return []string{
		"فعال‌سازی مستقیم روی اکانت کاربری شخصی شما",
		"حفظ کامل تاریخچه چت‌ها و اطلاعات قبلی",
		"تبادل امن کد دو مرحله‌ای (2FA) در میز کار",
		"پروتکل امنیتی پاکسازی کلمه عبور (Zeroize)",
		"پرداخت قانونی با ویزا/مسترکارت بین‌المللی",
	}
}

func defaultPlansForProduct(p *model.Product) []types.ProductPlan {
	basePrice := p.PriceToman
	if basePrice <= 0 {
		basePrice = 1500000
	}
	originalBase := int64(float64(basePrice) * 1.18 / 10000) * 10000

	threeMonthPrice := int64(float64(basePrice)*2.8/10000) * 10000
	threeMonthOrig := int64(float64(basePrice)*3.4/10000) * 10000

	oneYearPrice := int64(float64(basePrice)*9.5/10000) * 10000
	oneYearOrig := int64(float64(basePrice)*12.0/10000) * 10000

	deliveryStr := string(p.DeliveryType)

	return []types.ProductPlan{
		{
			ID:                 "plan-1m",
			Title:              "اشتراک ۱ ماهه اختصاصی",
			DurationDays:       30,
			DurationText:       "۱ ماهه",
			PriceToman:         basePrice,
			OriginalPriceToman: originalBase,
			BaseCostUSD:        p.BaseCostUSD,
			Badge:              "محبوب‌ترین",
			IsDefault:          true,
			DeliveryType:       deliveryStr,
			Features: []string{
				"دسترسی کامل به تمامی امکانات نسخه پرو",
				"گارانتی ۳۰ روزه تعویض و عدم قطعی",
				"پشتیبانی ۲۴ ساعته تلگرام",
			},
		},
		{
			ID:                 "plan-3m",
			Title:              "اشتراک ۳ ماهه اقتصادی",
			DurationDays:       90,
			DurationText:       "۳ ماهه",
			PriceToman:         threeMonthPrice,
			OriginalPriceToman: threeMonthOrig,
			BaseCostUSD:        p.BaseCostUSD * 2.8,
			Badge:              "تخفیف ویژه",
			IsDefault:          false,
			DeliveryType:       deliveryStr,
			Features: []string{
				"صرفه‌جویی قابل توجه در هزینه نسبت به تمدید ماهانه",
				"گارانتی کامل ۹۰ روزه اختصاصی",
				"اولویت بالا در صف پردازش سفارشات",
			},
		},
		{
			ID:                 "plan-1y",
			Title:              "اشتراک ۱ ساله ویژه",
			DurationDays:       365,
			DurationText:       "۱ ساله",
			PriceToman:         oneYearPrice,
			OriginalPriceToman: oneYearOrig,
			BaseCostUSD:        p.BaseCostUSD * 9.5,
			Badge:              "بیشترین تخفیف",
			IsDefault:          false,
			DeliveryType:       deliveryStr,
			Features: []string{
				"بیشترین میزان صرفه‌جویی و راحتی بدون نیاز به تمدید مداوم",
				"پشتیبانی VIP و گارانتی ۳۶۵ روزه",
				"دسترسی زودهنگام به مدل‌ها و قابلیت‌های جدید",
			},
		},
	}
}

func defaultComparisonForProduct(p *model.Product) *types.ProductComparison {
	brand := strings.ToLower(p.Title + " " + p.Slug)
	productName := p.TitleFA
	if strings.Contains(brand, "chatgpt") {
		return &types.ProductComparison{
			Columns: []types.ComparisonColumn{
				{Key: "free", Title: "نسخه رایگان (Free)", Subtitle: "کاربری عمومی روزمره", Highlight: false},
				{Key: "plus", Title: "نسخه پلاس (ChatGPT Plus)", Subtitle: "ارائه شده توسط آناهیکس", Badge: "پیشنهاد ما", Highlight: true},
				{Key: "team", Title: "نسخه تیمی (Team Plan)", Subtitle: "سازمان‌ها و کسب‌وکارها", Highlight: false},
			},
			Features: []types.ComparisonFeature{
				{Name: "دسترسی به قدرتمندترین مدل‌ها (GPT-4o و o1)", Category: "مدل‌های پردازشی", Values: map[string]string{"free": "بسیار محدود", "plus": "نامحدود و کامل", "team": "نامحدود و بدون سقف"}},
				{Name: "سقف ارسال پیام در هر ۳ ساعت", Category: "محدودیت استفاده", Values: map[string]string{"free": "۱۰ پیام در روز", "plus": "۸۰ پیام در هر ۳ ساعت", "team": "۱۰۰+ پیام در هر ۳ ساعت"}},
				{Name: "دستیار صوتی زنده (Advanced Voice Mode)", Category: "امکانات نسل جدید", Values: map[string]string{"free": "غیرفعال", "plus": "فعال با مکالمه آنی و طبیعی", "team": "فعال با پهنای باند اختصاصی"}},
				{Name: "ساخت و شخصی‌سازی GPT اختصاصی (Custom GPTs)", Category: "ابزارهای سفارشی", Values: map[string]string{"free": "فقط مشاهده", "plus": "ساخت نامحدود + انتشار عمومی", "team": "اشتراک‌گذاری درون‌سازمانی"}},
				{Name: "تحلیل پیشرفته داده‌ها، جداول و کد پایتون", Category: "تحلیل و پردازش", Values: map[string]string{"free": "محدود به فایل‌های سبک", "plus": "کاملاً نامحدود", "team": "نامحدود با امنیت محرمانه داده"}},
				{Name: "سرعت پاسخ‌دهی در ساعات اوج مصرف جهانی", Category: "سرعت و پایداری", Values: map[string]string{"free": "معمولی با صف انتظار", "plus": "حداکثر سرعت و اولویت سرور", "team": "اولویت اختصاصی دیتاسنتر"}},
				{Name: "ضمانت عدم قطعی و پشتیبانی مستقیم ریالی", Category: "گارانتی آناهیکس", Values: map[string]string{"free": "ندارد", "plus": "ضمانت ۳۰ روزه بازگشت وجه", "team": "پشتیبانی سازمانی"}},
			},
		}
	}

	if strings.Contains(brand, "gemini") {
		return &types.ProductComparison{
			Columns: []types.ComparisonColumn{
				{Key: "free", Title: "گوگل جمینای استاندارد", Subtitle: "حساب رایگان جیمیل", Highlight: false},
				{Key: "advanced", Title: "جمینای ادونس (Gemini Advanced)", Subtitle: "لایسنس ۱۸ ماهه آناهیکس", Badge: "ارزش خرید بالا", Highlight: true},
			},
			Features: []types.ComparisonFeature{
				{Name: "مدل پیشرفته هوش مصنوعی", Category: "هوش و استدلال", Values: map[string]string{"free": "Gemini 1.5 Flash", "advanced": "Gemini 1.5 Pro با استدلال نسل بعد"}},
				{Name: "حجم کانتکست (پنجره متنی)", Category: "حافظه و کانتکست", Values: map[string]string{"free": "۳۲ هزار توکن", "advanced": "۱ تا ۲ میلیون توکن (تحلیل کتاب و ویدیو)"}},
				{Name: "فضای ذخیره‌سازی ابری Google One", Category: "خدمات گوگل", Values: map[string]string{"free": "۱۵ گیگابایت اشتراکی", "advanced": "۲ ترابایت (۲,۰۰۰ گیگابایت) ابری"}},
				{Name: "یکپارچگی با جیمیل، داکس و درایو", Category: "بهره‌وری Workspace", Values: map[string]string{"free": "پایه‌ای", "advanced": "دستیار هوشمند مستقیم در Docs و Gmail"}},
				{Name: "نحوه فعال‌سازی و مدت زمان", Category: "شرایط تحویل", Values: map[string]string{"free": "محدود", "advanced": "لینک قانونی ۱۸ ماهه بدون نیاز به پسورد"}},
			},
		}
	}

	if strings.Contains(brand, "claude") {
		return &types.ProductComparison{
			Columns: []types.ComparisonColumn{
				{Key: "free", Title: "نسخه رایگان کلود", Subtitle: "استفاده ساده", Highlight: false},
				{Key: "pro", Title: "کلود پرو (Claude Pro)", Subtitle: "سرویس ویژه آناهیکس", Badge: "پیشنهاد برنامه‌نویسان", Highlight: true},
			},
			Features: []types.ComparisonFeature{
				{Name: "دسترسی به Claude 3.5 Sonnet و Opus", Category: "کیفیت مدل", Values: map[string]string{"free": "با سقف محدود روزانه", "pro": "دسترسی کامل با بالاترین سقف مجاز"}},
				{Name: "سقف ارسال پیام در مقایسه با رایگان", Category: "محدودیت", Values: map[string]string{"free": "استاندارد روزانه", "pro": "حداقل ۵ برابر پیام بیشتر"}},
				{Name: "قابلیت شبیه‌سازی زنده کدهای Artifacts", Category: "توسعه و کدنویسی", Values: map[string]string{"free": "پایه‌ای", "pro": "نامحدود با سرعت بالا و رندر لحظه‌ای"}},
				{Name: "دسترسی به پروژه‌ها (Projects) و اسناد سفارشی", Category: "سازماندهی", Values: map[string]string{"free": "غیرفعال", "pro": "فعال با حافظه پروژه‌محور"}},
				{Name: "پایداری هنگام ترافیک سنگین", Category: "پایداری", Values: map[string]string{"free": "خطای Capacity", "pro": "اولویت دسترسی بدون معطلی"}},
			},
		}
	}

	// General fallback comparison
	return &types.ProductComparison{
		Columns: []types.ComparisonColumn{
			{Key: "standard", Title: "پلن پایه / رایگان", Subtitle: "امکانات محدود", Highlight: false},
			{Key: "premium", Title: productName, Subtitle: "ارائه شده توسط آناهیکس", Badge: "بهترین انتخاب", Highlight: true},
		},
		Features: []types.ComparisonFeature{
			{Name: "دسترسی نامحدود و بدون فیلتر", Category: "دسترسی", Values: map[string]string{"standard": "با محدودیت و وقفه", "premium": "دسترسی کامل بدون افت کیفیت"}},
			{Name: "سرعت پردازش و اولویت سرور", Category: "سرعت", Values: map[string]string{"standard": "عادی", "premium": "حداکثر سرعت و بدون صف انتظار"}},
			{Name: "امنیت پرداخت و حساب کاربری", Category: "امنیت", Values: map[string]string{"standard": "نامشخص", "premium": "۱۰۰٪ قانونی با ضمانت عدم مسدودی"}},
			{Name: "پشتیبانی و گارانتی بازگشت وجه", Category: "پشتیبانی", Values: map[string]string{"standard": "ندارد", "premium": "گارانتی طلایی آناهیکس"}},
		},
	}
}

func defaultGuideForProduct(p *model.Product) []types.ActivationStep {
	if p.DeliveryType == model.DeliveryInstant {
		return []types.ActivationStep{
			{
				StepNumber:  1,
				Title:       "انتخاب پلن و پرداخت ریالی",
				Description: "مدت زمان مد نظر خود را انتخاب کرده و هزینه را از طریق درگاه‌های امن شتابی یا تتر پرداخت فرمایید.",
				Icon:        "credit-card",
				Tip:         "پس از پرداخت موفق، فوراً به صفحه پیگیری سفارش هدایت می‌شوید.",
			},
			{
				StepNumber:  2,
				Title:       "دریافت آنی لینک یا اطلاعات حساب",
				Description: "در کسری از ثانیه لینک فعال‌سازی اختصاصی یا مشخصات اکانت در پنل شما قرار می‌گیرد.",
				Icon:        "key",
				Tip:         "یک نسخه از اطلاعات سفارش نیز از طریق پیامک برای شما ارسال می‌گردد.",
			},
			{
				StepNumber:  3,
				Title:       "تنظیم DNS یا رفع تحریم اولیه",
				Description: "برای ورود بدون دردسر، اطمینان حاصل فرمایید از اینترنت پایدار و در صورت لزوم DNSهای رفع تحریم نظیر شکن یا الکترو استفاده می‌کنید.",
				Icon:        "globe",
				Tip:         "دی ان اس پیشنهادی شکن: 178.22.122.100 و 185.51.200.2",
			},
			{
				StepNumber:  4,
				Title:       "ورود و استفاده نامحدود",
				Description: "روی لینک کلیک کرده یا مشخصات را در وب‌سایت مرجع وارد کنید و بلافاصله از تمامی امکانات پریمیوم لذت ببرید.",
				Icon:        "sparkles",
				Tip:         "در صورت داشتن هرگونه سوال، تیم پشتیبانی فنی آماده راهنمایی و پاسخگویی است.",
			},
		}
	}

	return []types.ActivationStep{
		{
			StepNumber:  1,
			Title:       "ثبت مشخصات اکانت و پرداخت",
			Description: "پلن مورد نظر را انتخاب و ایمیل و پسورد اکانت خود را در کادرهای امن ثبت سفارش وارد کنید.",
			Icon:        "credit-card",
			Tip:         "اطلاعات شما با الگوریتم AES-256 رمزنگاری شده و غیرقابل نفوذ است.",
		},
		{
			StepNumber:  2,
			Title:       "آغاز فعال‌سازی توسط اپراتور",
			Description: "سفارش شما وارد میز کار اپراتور شده و فرآیند پرداخت ارزی قانونی با کارت معتبر بین‌المللی شروع می‌شود.",
			Icon:        "user-check",
			Tip:         "زمان معمول انجام این مرحله بین ۱۵ الی ۳۰ دقیقه در ساعات کاری است.",
		},
		{
			StepNumber:  3,
			Title:       "تبادل زنده کد دو مرحله‌ای (2FA)",
			Description: "در صورت فعال بودن کد تایید دو مرحله‌ای، نوتیفیکیشن لایو در صفحه پیگیری به شما نمایش داده می‌شود تا کد پیامک شده را درج کنید.",
			Icon:        "shield-check",
			Tip:         "این تبادل بدون نیاز به تماس تلفنی و کاملاً اتوماتیک در مرورگر شما انجام می‌گیرد.",
		},
		{
			StepNumber:  4,
			Title:       "اتمام کار و پاکسازی کامل رمزها (Zeroize)",
			Description: "اشتراک شما ارتقا یافته و رمز عبور وارد شده بلافاصله برای همیشه از سیستم امحا می‌شود.",
			Icon:        "sparkles",
			Tip:         "اکانت شخصی شما اکنون آماده استفاده با دسترسی‌های حرفه‌ای است.",
		},
	}
}

func defaultFaqsForProduct(p *model.Product) []types.ProductFAQ {
	brand := strings.ToLower(p.Title + " " + p.Slug)
	if strings.Contains(brand, "chatgpt") {
		return []types.ProductFAQ{
			{
				Question: "آیا در ارتقای روی اکانت شخصی، تاریخچه چت‌ها و پروژه‌های من حفظ می‌شود؟",
				Answer:   "بله کاملاً. در این روش اشتراک مستقیم روی همان ایمیل و اکانت OpenAI فعلی شما فعال می‌شود و هیچ‌یک از گفتگوها، کاستوم GPTها و سوابق شما تغییر نخواهد کرد.",
			},
			{
				Question: "آیا در طول دوره اشتراک احتمال مسدودی یا قطعی اکانت وجود دارد؟",
				Answer:   "خیر، تمامی پرداخت‌ها با کارت‌های اعتباری معتبر بین‌المللی و حساب‌های بانکی قانونی انجام می‌پذیرد. آناهیکس تمام ۳۰ روز اشتراک را تضمین و گارانتی کامل می‌نماید.",
			},
			{
				Question: "برای ورود به ChatGPT چه نوع دی‌ان‌اس یا فیلترشکنی پیشنهاد می‌شود؟",
				Answer:   "برای عبور از تحریم‌های منطقه‌ای شرکت OpenAI، استفاده از DNS شکن (178.22.122.100) یا الکترو و همچنین آی‌پی‌های ثابت غیر از ایران پیشنهاد می‌شود.",
			},
			{
				Question: "پروتکل نابودسازی رمز (Zeroize) چیست و چطور امنیت من حفظ می‌شود؟",
				Answer:   "کلمه عبور وارد شده در سیستم رمزنگاری شده و اپراتور تنها جهت انجام تراکنش ارتقا به آن دسترسی دارد. به محض فشردن کلید تایید نهایی سفارش، رمز برای همیشه از پایگاه داده حذف می‌گردد.",
			},
		}
	}

	if strings.Contains(brand, "gemini") {
		return []types.ProductFAQ{
			{
				Question: "اشتراک ۱۸ ماهه گوگل جمینای چگونه فعال می‌شود؟",
				Answer:   "پس از پرداخت، یک لینک فعال‌سازی اختصاصی به شما داده می‌شود. با باز کردن لینک در مرورگر و کلیک روی دکمه فعال‌سازی روی جیمیل شخصی‌تان، اکانت شما بلافاصله به اشتراک ۱۸ ماهه گوگل وان و جمینای ادونس متصل می‌شود.",
			},
			{
				Question: "آیا نیاز است رمز جیمیل خود را به کسی بدهم؟",
				Answer:   "خیر! این مزیت بزرگ لینک اختصاصی آناهیکس است. هیچ نیازی به ارائه ایمیل، پسورد یا دسترسی به جیمیل شما نیست و فعال‌سازی توسط خودتان با یک کلیک انجام می‌گردد.",
			},
			{
				Question: "آیا فضای ۲ ترابایت ابری گوگل درایو هم فعال می‌شود؟",
				Answer:   "بله، این پلن علاوه بر هوش مصنوعی Gemini Advanced، شامل ۲ ترابایت فضای ابری مشترک در Google One (درایو، جیمیل و گوگل فوتوز) به مدت ۱۸ ماه کامل است.",
			},
			{
				Question: "اگر سقف اعضای فمیلی گوگل من پر باشد چه کنم؟",
				Answer:   "گوگل امکان تغییر فمیلی را هر ۱۲ ماه یک‌بار می‌دهد. اگر جیمیل شما قبلاً عضو فمیلی دیگری بوده است، می‌توانید یک جیمیل تازه یا تمیز برای فعال‌سازی استفاده کنید.",
			},
		}
	}

	return []types.ProductFAQ{
		{
			Question: "ضمانت این محصول به چه صورت است؟",
			Answer:   "تمامی محصولات ارائه شده در آناهیکس دارای ضمانت کامل اصالت و سلامت در طول کل مدت اشتراک هستند و در صورت بروز هرگونه اختلال، پشتیبانی سریعاً تعویض یا رفع اشکال می‌نماید.",
		},
		{
			Question: "تحویل سفارش چقدر زمان می‌برد؟",
			Answer:   "محصولات با برچسب تحویل آنی ظرف ۰ ثانیه تحویل می‌شوند. سفارشات نیازمند مداخله اپراتور معمولاً بین ۱۵ تا ۳۰ دقیقه در ساعات کاری فعال می‌گردند.",
		},
		{
			Question: "آیا پس از اتمام دوره امکان تمدید مجدد وجود دارد؟",
			Answer:   "بله، شما می‌توانید پیش از اتمام مهلت اشتراک، سفارش تمدید خود را در سایت ثبت نمایید تا سرویس بدون قطعی ادامه یابد.",
		},
	}
}

