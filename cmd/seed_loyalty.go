package cmd

import (
	"log"

	"github.com/EbadiDev/anahix-server/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func seedLoyaltyData(db *gorm.DB) {
	var count int64
	db.Model(&model.LoyaltyTier{}).Count(&count)
	if count == 0 {
		tiers := []model.LoyaltyTier{
			{
				ID:              uuid.New(),
				Code:            "bronze",
				Name:            "برنزی (عادی)",
				NameEN:          "Bronze",
				MinSpendToman:   0,
				DiscountPercent: 0.0,
				QueuePriority:   0,
				BadgeColor:      "#cd7f32",
				BadgeGradient:   "from-amber-700/20 via-amber-800/10 to-amber-900/30 border-amber-600/30 text-amber-500",
				Icon:            "medal",
				Description:     "سطح پایه کاربری با دسترسی به کلیه امکانات استاندارد، تخفیف‌های مناسبتی و تحویل آنی",
				PerksJSON:       `["قیمت‌های رقابتی و مصوب", "تحویل خودکار لایسنس‌ها و اکانت‌ها", "پشتیبانی تیکتی استاندارد", "امکان استفاده از کدهای تخفیف عمومی"]`,
				OrderIndex:      1,
				IsActive:        true,
			},
			{
				ID:              uuid.New(),
				Code:            "silver",
				Name:            "نقره‌ای (ویژه)",
				NameEN:          "Silver",
				MinSpendToman:   5000000, // 5 Million Tomans
				DiscountPercent: 2.0,     // 2% permanent discount
				QueuePriority:   10,
				BadgeColor:      "#94a3b8",
				BadgeGradient:   "from-slate-400/20 via-zinc-400/10 to-slate-500/30 border-slate-400/40 text-slate-300",
				Icon:            "star",
				Description:     "ویژه خریداران فعال با ۲٪ تخفیف خودکار روی تمامی لایسنس‌ها و اولویت در پردازش دستی",
				PerksJSON:       `["۲٪ تخفیف خودکار روی تمامی لایسنس‌ها و اکانت‌ها", "اولویت در صف پردازش اپراتور Mode C", "نشان کاربری نقره‌ای در پروفایل و تیکت‌ها", "اولویت پاسخگویی در پشتیبانی فنی"]`,
				OrderIndex:      2,
				IsActive:        true,
			},
			{
				ID:              uuid.New(),
				Code:            "gold",
				Name:            "طلایی (ریسلرها و همکاران)",
				NameEN:          "Gold",
				MinSpendToman:   20000000, // 20 Million Tomans
				DiscountPercent: 5.0,      // 5% permanent discount
				QueuePriority:   50,
				BadgeColor:      "#eab308",
				BadgeGradient:   "from-yellow-500/25 via-amber-500/15 to-amber-600/35 border-yellow-500/50 text-yellow-400 shadow-yellow-500/10",
				Icon:            "crown",
				Description:     "مخصوص همکاران، آژانس‌های تبلیغاتی و خریداران عمده با ۵٪ تخفیف خودکار و بالاترین اولویت در صف",
				PerksJSON:       `["۵٪ تخفیف خودکار روی کلیه محصولات، اکانت‌ها و شماره‌های مجازی", "بالاترین اولویت VIP در صف اپراتور (پردازش بدون معطلی)", "کانال ارتباطی اختصاصی با سوپروایزر", "امکان رزرو لایسنس‌های اختصاصی", "نشان زرین VIP کلاب"]`,
				OrderIndex:      3,
				IsActive:        true,
			},
			{
				ID:              uuid.New(),
				Code:            "diamond",
				Name:            "الماسی (سازمانی و پارتنر کلیدی)",
				NameEN:          "Diamond",
				MinSpendToman:   50000000, // 50 Million Tomans
				DiscountPercent: 8.0,      // 8% permanent discount
				QueuePriority:   100,
				BadgeColor:      "#38bdf8",
				BadgeGradient:   "from-cyan-500/25 via-sky-500/15 to-blue-600/35 border-cyan-400/50 text-cyan-400 shadow-cyan-500/15",
				Icon:            "diamond",
				Description:     "سطح انحصاری سازمان‌ها و شرکای کلیدی با بیشترین درصد تخفیف، پشتیبانی تلفنی و مدیر حساب اختصاصی",
				PerksJSON:       `["۸٪ تخفیف سازمانی روی کلیه خدمات", "تضمین تحویل فوق سریع سفارش‌ها زیر ۱۰ دقیقه", "مدیر حساب اختصاصی ۲۴ ساعته", "امکان خرید اعتباری و تسویه دوره‌ای", "صدور آنی فاکتور رسمی سازمانی"]`,
				OrderIndex:      4,
				IsActive:        true,
			},
		}

		for _, t := range tiers {
			db.Create(&t)
		}
		log.Println("🌱 Seeded default Loyalty Tiers (bronze, silver, gold, diamond)")
	}

	// Seed loyalty and referral system settings if not exists
	var settingCount int64
	db.Model(&model.SystemSetting{}).Where("category = ?", "loyalty").Count(&settingCount)
	if settingCount == 0 {
		settings := []model.SystemSetting{
			{
				Category: "loyalty",
				Key:      "loyalty_enabled",
				Value:    "true",
				Type:     "bool",
				Desc:     "فعال بودن باشگاه مشتریان و تخفیف وفاداری VIP",
			},
			{
				Category: "loyalty",
				Key:      "loyalty_calculation_period",
				Value:    "monthly_30d",
				Type:     "string",
				Desc:     "دوره محاسبه حجم خرید برای ارتقای سطح (monthly_30d یا all_time)",
			},
			{
				Category: "loyalty",
				Key:      "loyalty_auto_upgrade",
				Value:    "true",
				Type:     "bool",
				Desc:     "ارتقای خودکار سطح کاربر پس از تکمیل یا پرداخت سفارش",
			},
			{
				Category: "loyalty",
				Key:      "referral_enabled",
				Value:    "active",
				Type:     "string",
				Desc:     "وضعیت همکاری در فروش (active, coming_soon, disabled)",
			},
			{
				Category: "loyalty",
				Key:      "referral_commission_rate",
				Value:    "10",
				Type:     "number",
				Desc:     "درصد پورسانت معرف از هر سفارش",
			},
			{
				Category: "loyalty",
				Key:      "referral_bonus_amount",
				Value:    "20000",
				Type:     "number",
				Desc:     "پاداش شارژ هدیه برای خریدار جدید (تومان)",
			},
			{
				Category: "loyalty",
				Key:      "referral_min_withdraw",
				Value:    "50000",
				Type:     "number",
				Desc:     "حداقل موجودی جهت تسویه‌حساب نقدی (تومان)",
			},
		}

		for _, s := range settings {
			db.Create(&s)
		}
		log.Println("🌱 Seeded Loyalty Club & Referral system settings (category: 'loyalty')")
	}
}
