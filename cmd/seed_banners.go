package cmd

import (
	"log"
	"strings"
	"time"

	"github.com/EbadiDev/anahix-server/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func seedBannersAndThemeData(db *gorm.DB) {
	// 1. Seed Banners
	var bannerCount int64
	db.Model(&model.Banner{}).Count(&bannerCount)
	if bannerCount == 0 {
		now := time.Now()
		banners := []model.Banner{
			{
				ID:         uuid.New(),
				Title:      "جشنواره تخفیف ویژه هوش مصنوعی",
				Subtitle:   "تا ۴۰٪ تخفیف روی اکانت‌های قانونی ChatGPT Plus و Claude 3.7 Sonnet با تحویل فوری",
				BadgeText:  "جشنواره شگفت‌انگیز",
				ImageURL:   "https://images.unsplash.com/photo-1618005182384-a83a8bd57fbe?q=80&w=1200&auto=format&fit=crop",
				LinkURL:    "/#flash-sale",
				Position:   model.BannerPositionMainSlider,
				BgGradient: "from-blue-600 via-indigo-700 to-slate-900",
				ButtonText: "مشاهده تخفیف‌های زمان‌دار",
				SortOrder:  1,
				IsActive:   true,
				CreatedAt:  now,
				UpdatedAt:  now,
			},
			{
				ID:         uuid.New(),
				Title:      "اشتراک اختصاصی Google Gemini Advanced",
				Subtitle:   "ارتقای ۱۸ ماهه بدون واسطه و پایدار همراه با ۲ ترابایت فضای ابری گوگل وان",
				BadgeText:  "پرفروش‌ترین سرویس ماه",
				ImageURL:   "https://images.unsplash.com/photo-1620712943543-bcc4688e7485?q=80&w=1200&auto=format&fit=crop",
				LinkURL:    "/products/gemini-advanced-18m",
				Position:   model.BannerPositionMainSlider,
				BgGradient: "from-indigo-600 via-purple-700 to-slate-900",
				ButtonText: "خرید فوری ۱۸ ماهه",
				SortOrder:  2,
				IsActive:   true,
				CreatedAt:  now,
				UpdatedAt:  now,
			},
			{
				ID:         uuid.New(),
				Title:      "شماره‌های مجازی ۱۸۰ کشور با گارانتی تعویض",
				Subtitle:   "دریافت آنی پیامک تایید تلگرام، اوپن‌ای‌آی، گوگل و اینستاگرام با عودت وجه تضمینی",
				BadgeText:  "سرویس جدید",
				ImageURL:   "https://images.unsplash.com/photo-1526374965328-7f61d4dc18c5?q=80&w=1200&auto=format&fit=crop",
				LinkURL:    "/virtual-numbers",
				Position:   model.BannerPositionMiddle,
				BgGradient: "from-cyan-600 via-blue-700 to-slate-900",
				ButtonText: "دریافت شماره مجازی",
				SortOrder:  3,
				IsActive:   true,
				CreatedAt:  now,
				UpdatedAt:  now,
			},
		}

		for _, b := range banners {
			db.Create(&b)
		}
		log.Println("🌱 Seeded initial promotional banners (table: 'banners')")
	}

	// Ensure all banners have valid timestamps
	var allBanners []model.Banner
	db.Find(&allBanners)
	nowTime := time.Now()
	for _, b := range allBanners {
		if b.CreatedAt.IsZero() {
			db.Model(&model.Banner{}).Where("id = ?", b.ID).Updates(map[string]interface{}{
				"created_at": nowTime.Add(-12 * time.Hour),
				"updated_at": nowTime,
			})
		}
	}

	// 2. Seed Default Theme Settings if not present
	var themeCount int64
	db.Model(&model.SystemSetting{}).Where("category IN ?", []string{"theme", "auth"}).Count(&themeCount)
	if themeCount == 0 {
		themeSettings := []model.SystemSetting{
			{Category: "theme", Key: "theme_event_name", Value: "رویداد پیش‌فرض آناهیکس", Type: "string", Desc: "عنوان رویداد فعال"},
			{Category: "theme", Key: "theme_event_active", Value: "false", Type: "bool", Desc: "فعال‌سازی رویداد و تغییرات بصری"},
			{Category: "theme", Key: "theme_primary_color", Value: "#2563eb", Type: "string", Desc: "رنگ اصلی سازمانی (رویال بلو)"},
			{Category: "theme", Key: "theme_accent_color", Value: "#3b82f6", Type: "string", Desc: "رنگ ثانویه"},
			{Category: "theme", Key: "theme_header_announcement", Value: "تمامی سرویس‌ها و شماره‌های مجازی فعال و تحویل آنی ۰ ثانیه‌ای برقرار است", Type: "string", Desc: "متن نوار اعلان هدر"},
			{Category: "theme", Key: "theme_event_effect", Value: "none", Type: "string", Desc: "افکت بصری انیمیشنی"},
			{Category: "theme", Key: "theme_event_icon", Value: "sparkles", Type: "string", Desc: "آیکون رویداد"},
			{Category: "auth", Key: "auth_google_enabled", Value: "true", Type: "bool", Desc: "ورود با گوگل"},
			{Category: "auth", Key: "auth_github_enabled", Value: "true", Type: "bool", Desc: "ورود با گیت‌هاب"},
			{Category: "auth", Key: "auth_password_enabled", Value: "true", Type: "bool", Desc: "ورود با کلمه عبور"},
			{Category: "auth", Key: "auth_otp_required", Value: "false", Type: "bool", Desc: "اجباری بودن OTP"},
		}

		for _, s := range themeSettings {
			db.Create(&s)
		}
		log.Println("🌱 Seeded theme and multi-auth system configurations")
	}

	// 3. Update sample products with Flash Sale and high-res Images
	var products []model.Product
	db.Find(&products)
	futureDate := time.Now().Add(72 * time.Hour)

	for _, p := range products {
		updates := map[string]interface{}{}
		slug := strings.ToLower(p.Slug)

		if strings.Contains(slug, "gemini") {
			updates["image_url"] = "https://images.unsplash.com/photo-1620712943543-bcc4688e7485?q=80&w=800&auto=format&fit=crop"
			updates["is_flash_sale"] = true
			updates["flash_sale_price_toman"] = 990000 // ~33% off 1,480,000
			updates["flash_sale_end_at"] = futureDate
			updates["flash_sale_stock"] = 60
			updates["flash_sale_sold"] = 41
			updates["is_best_seller"] = true
			updates["is_featured"] = true
			updates["rating_average"] = 4.95
			updates["sales_count"] = 1240
			updates["badges_json"] = `["۳۳٪ تخفیف شگفت‌انگیز", "۲ ترابایت فضای ابری", "۱۸ ماهه قانونی"]`
		} else if strings.Contains(slug, "chatgpt") {
			updates["image_url"] = "https://images.unsplash.com/photo-1677442136019-21780ecad995?q=80&w=800&auto=format&fit=crop"
			updates["is_flash_sale"] = true
			updates["flash_sale_price_toman"] = 1560000 // 20% off 1,950,000
			updates["flash_sale_end_at"] = time.Now().Add(96 * time.Hour)
			updates["flash_sale_stock"] = 50
			updates["flash_sale_sold"] = 34
			updates["is_best_seller"] = true
			updates["is_featured"] = true
			updates["rating_average"] = 4.98
			updates["sales_count"] = 2410
			updates["badges_json"] = `["۲۰٪ تخفیف ویژه", "ارتقای قانونی ایمیل شخصی", "پروتکل امنیتی Zeroize"]`
		} else if strings.Contains(slug, "claude") {
			updates["image_url"] = "https://images.unsplash.com/photo-1618005182384-a83a8bd57fbe?q=80&w=800&auto=format&fit=crop"
			updates["is_flash_sale"] = true
			updates["flash_sale_price_toman"] = 1417000 // 25% off 1,890,000
			updates["flash_sale_end_at"] = time.Now().Add(48 * time.Hour)
			updates["flash_sale_stock"] = 40
			updates["flash_sale_sold"] = 28
			updates["is_best_seller"] = true
			updates["is_featured"] = true
			updates["rating_average"] = 4.92
			updates["sales_count"] = 890
			updates["badges_json"] = `["۲۵٪ تخفیف زمان‌دار", "مدل Sonnet 3.7", "اکانت آماده اختصاصی"]`
		} else if strings.Contains(slug, "cursor") {
			updates["image_url"] = "https://images.unsplash.com/photo-1550751827-4bd374c3f58b?q=80&w=800&auto=format&fit=crop"
			updates["is_flash_sale"] = true
			updates["flash_sale_price_toman"] = 1225000 // 30% off 1,750,000
			updates["flash_sale_end_at"] = time.Now().Add(120 * time.Hour)
			updates["flash_sale_stock"] = 45
			updates["flash_sale_sold"] = 31
			updates["is_best_seller"] = true
			updates["is_featured"] = true
			updates["rating_average"] = 4.94
			updates["sales_count"] = 680
			updates["badges_json"] = `["۳۰٪ تخفیف ویژه برنامه‌نویسان", "دستیار کدنویسی AI", "تحویل فوری"]`
		} else {
			if p.ImageURL == "" {
				updates["image_url"] = "https://images.unsplash.com/photo-1550751827-4bd374c3f58b?q=80&w=800&auto=format&fit=crop"
			}
			updates["rating_average"] = 4.88
			updates["sales_count"] = 420
			updates["badges_json"] = `["تحویل قانونی", "ضمانت بازگشت وجه"]`
		}

		if len(updates) > 0 {
			db.Model(&model.Product{}).Where("id = ?", p.ID).Updates(updates)
		}
	}

	// 4. Ensure Midjourney test offer product exists
	var midjourneyCount int64
	db.Model(&model.Product{}).Where("slug = ?", "midjourney-pro-1m").Count(&midjourneyCount)
	if midjourneyCount == 0 {
		endAt := time.Now().Add(72 * time.Hour)
		mj := model.Product{
			ID:                  uuid.New(),
			Title:               "Midjourney Pro AI 1 Month Subscription",
			TitleFA:             "اشتراک ۱ ماهه هوش مصنوعی میدجورنی پرو (Midjourney Pro)",
			Slug:                "midjourney-pro-1m",
			Description:         "دسترسی اختصاصی و نامحدود به سرورهای تولید تصویر فوق پیشرفته Midjourney v6 با حالت مخفی (Stealth Mode) و تولید موازی.",
			Category:            model.CategoryAIAccount,
			DeliveryType:        model.DeliveryInstant,
			SLAMinMinutes:       1,
			SLAMaxMinutes:       5,
			SLADisplayText:      "تحویل ۰ ثانیه‌ای (آنی)",
			OperatingHours:      "۲۴ ساعته (تحویل خودکار)",
			BaseCostUSD:         25,
			PriceToman:          2400000,
			IsActive:            true,
			ImageURL:            "https://images.unsplash.com/photo-1618005182384-a83a8bd57fbe?q=80&w=800&auto=format&fit=crop",
			IsFlashSale:         true,
			FlashSalePriceToman: 1680000, // 30% discount
			FlashSaleEndAt:      &endAt,
			FlashSaleStock:      30,
			FlashSaleSold:       19,
			IsBestSeller:        true,
			IsFeatured:          true,
			RatingAverage:       4.96,
			SalesCount:          510,
			BadgesJSON:          `["۳۰٪ تخفیف جشنواره", "تولید نامحدود تصویر", "Stealth Mode"]`,
			GroupTag:            "midjourney",
			GroupTitleFA:        "میدجورنی (Midjourney AI)",
			CreatedAt:           nowTime,
			UpdatedAt:           nowTime,
		}
		db.Create(&mj)
		log.Println("🌱 Seeded Midjourney Pro test offer product (table: 'products')")
	}
}
