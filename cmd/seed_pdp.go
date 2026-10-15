package cmd

import (
	"encoding/json"
	"log"

	"github.com/EbadiDev/anahix-server/internal/model"
	"github.com/EbadiDev/anahix-server/internal/types"
	"gorm.io/gorm"
)

func seedPDPData(db *gorm.DB) {
	// 1. ChatGPT Plus Personal
	chatgptPlans := []types.ProductPlan{
		{
			ID:                 "chatgpt-1m-personal",
			Title:              "اشتراک ۱ ماهه ارتقا روی اکانت شخصی",
			DurationDays:       30,
			DurationText:       "۱ ماهه اختصاصی",
			PriceToman:         1950000,
			OriginalPriceToman: 2350000,
			BaseCostUSD:        20.0,
			Badge:              "محبوب‌ترین",
			IsDefault:          true,
			DeliveryType:       "manual_timed",
			Features: []string{
				"تمدید مستقیم با مسترکارت معتبر بین‌المللی",
				"حفظ کامل سوابق، چت‌ها و کاستوم GPTها",
				"تبادل زنده کد دو مرحله‌ای (2FA) در میز کار",
				"پروتکل امنیتی پاکسازی کلمه عبور (Zeroize)",
				"ضمانت کامل ۳۰ روزه عدم قطعی",
			},
		},
		{
			ID:                 "chatgpt-3m-personal",
			Title:              "اشتراک ۳ ماهه اقتصادی روی اکانت شخصی",
			DurationDays:       90,
			DurationText:       "۳ ماهه اقتصادی",
			PriceToman:         5490000,
			OriginalPriceToman: 7050000,
			BaseCostUSD:        56.0,
			Badge:              "بیشترین تخفیف",
			IsDefault:          false,
			DeliveryType:       "manual_timed",
			Features: []string{
				"صرفه‌جویی بیش از ۳۶۰ هزار تومان نسبت به خرید ماهانه",
				"گارانتی و بیمه کامل ۹۰ روزه آناهیکس",
				"اولویت طلایی در صف بررسی اپراتور",
				"پشتیبانی اختصاصی VIP",
			},
		},
		{
			ID:                 "chatgpt-1m-ready",
			Title:              "اکانت آماده اختصاصی تحویل آنی",
			DurationDays:       30,
			DurationText:       "۱ ماهه آماده",
			PriceToman:         1890000,
			OriginalPriceToman: 2200000,
			BaseCostUSD:        19.0,
			Badge:              "تحویل ۰ ثانیه‌ای",
			IsDefault:          false,
			DeliveryType:       "instant",
			Features: []string{
				"تحویل در لحظه (ایمیل و پسورد اختصاصی)",
				"امکان تغییر رمز عبور و ایمیل بازیابی توسط شما",
				"بدون معطلی و بدون نیاز به تبادل کد ۲FA",
			},
		},
	}
	chatgptPlansJSON, _ := json.Marshal(chatgptPlans)

	chatgptComp := types.ProductComparison{
		Columns: []types.ComparisonColumn{
			{Key: "free", Title: "نسخه رایگان (Free)", Subtitle: "کاربری عمومی روزمره", Highlight: false},
			{Key: "plus", Title: "نسخه پلاس (ChatGPT Plus)", Subtitle: "ارائه شده توسط آناهیکس", Badge: "پیشنهاد ما", Highlight: true},
			{Key: "team", Title: "نسخه تیمی (Team Plan)", Subtitle: "سازمان‌ها و گروه‌ها", Highlight: false},
		},
		Features: []types.ComparisonFeature{
			{Name: "دسترسی به قدرتمندترین مدل‌ها (GPT-4o و مدل‌های استدلالی o1)", Category: "مدل‌های پردازشی", Values: map[string]string{"free": "بسیار محدود", "plus": "نامحدود و کامل", "team": "نامحدود و با سقف دوبرابر"}},
			{Name: "سقف ارسال پیام در هر ۳ ساعت", Category: "محدودیت استفاده", Values: map[string]string{"free": "۱۰ پیام در روز", "plus": "۸۰ پیام در هر ۳ ساعت", "team": "۱۰۰+ پیام در هر ۳ ساعت"}},
			{Name: "دستیار صوتی زنده با زبان فارسی (Advanced Voice)", Category: "امکانات نسل جدید", Values: map[string]string{"free": "غیرفعال", "plus": "فعال با مکالمه آنی و لحن طبیعی", "team": "فعال با پهنای باند اختصاصی"}},
			{Name: "ساخت و شخصی‌سازی GPT اختصاصی (Custom GPTs)", Category: "ابزارهای سفارشی", Values: map[string]string{"free": "فقط مشاهده", "plus": "ساخت نامحدود + انتشار عمومی و خصوصی", "team": "اشتراک‌گذاری درون‌سازمانی"}},
			{Name: "تحلیل پیشرفته داده‌ها، رسم نمودار و اجرای کد پایتون", Category: "تحلیل و پردازش", Values: map[string]string{"free": "محدود به فایل‌های سبک", "plus": "کاملاً نامحدود برای دیتاست‌های سنگین", "team": "نامحدود با حفظ محرمانگی داده"}},
			{Name: "تولید تصاویر فوق‌العاده با DALL·E 3", Category: "مولتی‌مدیال", Values: map[string]string{"free": "۳ تصویر در روز", "plus": "تولید با بالاترین کیفیت بدون محدودیت روزانه", "team": "تولید همزمان چند تصویری"}},
			{Name: "پشتیبانی، ضمانت بازگشت وجه و پرداخت ریالی", Category: "گارانتی آناهیکس", Values: map[string]string{"free": "ندارد", "plus": "ضمانت کامل ۳۰ روزه تعویض یا بازگشت وجه", "team": "پشتیبانی قرارداد تجاری"}},
		},
	}
	chatgptCompJSON, _ := json.Marshal(chatgptComp)

	chatgptGuide := []types.ActivationStep{
		{
			StepNumber:  1,
			Title:       "ثبت سفارش و پرداخت امن ریالی",
			Description: "پلن مد نظر خود را انتخاب کرده و هزینه را از طریق درگاه امن شاپرک یا موجودی کیف پول پرداخت فرمایید.",
			Icon:        "credit-card",
			Tip:         "پس از پرداخت موفق، فوراً به صفحه پیگیری سفارش هدایت می‌شوید.",
		},
		{
			StepNumber:  2,
			Title:       "آغاز فعال‌سازی توسط اپراتور مجرب",
			Description: "سفارش شما وارد میز کار اپراتور شده و پرداخت ارزی قانونی با کارت بین‌المللی اختصاصی آغاز می‌گردد.",
			Icon:        "user-check",
			Tip:         "میانگین زمان پردازش ۱۵ الی ۳۰ دقیقه در ساعات کاری است.",
		},
		{
			StepNumber:  3,
			Title:       "تبادل هوشمند کد دو مرحله‌ای (2FA)",
			Description: "در صورت دریافت کد ورود از سوی OpenAI، اعلان لایو در صفحه پیگیری به شما نشان داده می‌شود تا کد پیامک یا ایمیل شده را وارد کنید.",
			Icon:        "shield-check",
			Tip:         "نیازی به تماس تلفنی نیست؛ تبادل کد کاملاً در وب‌سایت انجام می‌گیرد.",
		},
		{
			StepNumber:  4,
			Title:       "تکمیل اشتراک و امحای رمز (Zeroize)",
			Description: "اکانت شما پلاس شده و کلمه عبور وارد شده بلافاصله با متد کریپتوگرافیک برای همیشه از سرور پاکسازی (Zeroize) می‌شود.",
			Icon:        "sparkles",
			Tip:         "برای اتصال پایدار و بدون ارور ۴۰۳، استفاده از DNS شکن (178.22.122.100) یا الکترو توصیه می‌شود.",
		},
	}
	chatgptGuideJSON, _ := json.Marshal(chatgptGuide)

	chatgptFaqs := []types.ProductFAQ{
		{
			Question: "آیا در ارتقای اکانت شخصی، گفتگوها و تاریخچه قبلی من حفظ می‌شود؟",
			Answer:   "بله صد در صد. ارتقا دقیقاً روی همان جیمیل یا ایمیل کاربری OpenAI شما انجام می‌شود و تمام پروژه‌ها، چت‌ها و کاستوم جی‌پی‌تی‌های شما دست‌نخورده باقی می‌مانند.",
		},
		{
			Question: "آیا احتمال قطعی یا بن شدن اکانت در این روش وجود دارد؟",
			Answer:   "خیر، تمامی پرداخت‌ها با کارت‌های مسترکارت اختصاصی و با ارز قانونی انجام می‌شود. آناهیکس کل مدت اشتراک را با گارانتی کتبی پوشش می‌دهد.",
		},
		{
			Question: "پروتکل امنیتی Zeroize پسوردهای آناهیکس چیست؟",
			Answer:   "رمز عبور وارد شده ابتدا با کلید AES-256 رمزنگاری شده و در والت امن قرار می‌گیرد. به محض تایید موفق پرداخت توسط اپراتور، کلید و داده رمز برای همیشه صفر (Zeroize) می‌شوند.",
		},
		{
			Question: "در صورت بروز مشکل در فرآیند فعال‌سازی با چه کسی در ارتباط باشم؟",
			Answer:   "پشتیبانی آنلاین تلگرام آناهیکس (@AnahixSupport) و بخش تیکت‌ها به صورت ۲۴ ساعته در خدمت شما هستند.",
		},
	}
	chatgptFaqsJSON, _ := json.Marshal(chatgptFaqs)

	db.Model(&model.Product{}).Where("slug = ?", "chatgpt-plus-personal").Updates(map[string]interface{}{
		"title_fa":         "ارتقای قانونی چت‌جی‌پی‌تی پلاس روی اکانت شخصی (ChatGPT Plus)",
		"is_active":        true,
		"meta_title":       "خرید اکانت قانونی ChatGPT Plus (ارتقای اکانت شخصی) | آناهیکس",
		"meta_description": "خرید اشتراک قانونی ChatGPT Plus روی ایمیل شخصی شما با تضمین عدم قطعی، تبادل آنلاین ۲FA، پروتکل امنیتی Zeroize و پشتیبانی شبانه‌روزی در آناهیکس.",
		"og_image_url":     "/assets/hero.png",
		"plans_json":       string(chatgptPlansJSON),
		"comparison_json":  string(chatgptCompJSON),
		"guide_json":       string(chatgptGuideJSON),
		"faq_json":         string(chatgptFaqsJSON),
	})

	// 2. Gemini Advanced 18 Months
	geminiPlans := []types.ProductPlan{
		{
			ID:                 "gemini-18m-link",
			Title:              "لینک اختصاصی ۱۸ ماهه گوگل جمینای و Google One",
			DurationDays:       540,
			DurationText:       "۱۸ ماهه (بیشترین ارزش)",
			PriceToman:         1480000,
			OriginalPriceToman: 2100000,
			BaseCostUSD:        12.0,
			Badge:              "ارزش خرید فوق‌العاده",
			IsDefault:          true,
			DeliveryType:       "instant",
			Features: []string{
				"۲ ترابایت فضای ابری Google One (درایو، جیمیل، فوتوز)",
				"دسترسی به Gemini 1.5 Pro با پنجره متنی ۲ میلیون توکن",
				"تحویل ۰ ثانیه‌ای بلافاصله پس از پرداخت",
				"فعال‌سازی با یک کلیک روی جیمیل شخصی بدون پسورد",
			},
		},
		{
			ID:                 "gemini-12m-direct",
			Title:              "اشتراک ۱ ساله مستقیم گوگل جمینای",
			DurationDays:       365,
			DurationText:       "۱ ساله",
			PriceToman:         1150000,
			OriginalPriceToman: 1600000,
			BaseCostUSD:        9.5,
			Badge:              "پلن اقتصادی",
			IsDefault:          false,
			DeliveryType:       "instant",
			Features: []string{
				"تحویل فوری خودکار",
				"دسترسی کامل به مدل‌های هوش مصنوعی پیشرفته گوگل",
				"گارانتی کامل ۳۶۵ روزه",
			},
		},
	}
	geminiPlansJSON, _ := json.Marshal(geminiPlans)

	geminiComp := types.ProductComparison{
		Columns: []types.ComparisonColumn{
			{Key: "free", Title: "گوگل جمینای استاندارد", Subtitle: "حساب معمولی گوگل", Highlight: false},
			{Key: "advanced", Title: "جمینای ادونس (Gemini Advanced)", Subtitle: "لایسنس ۱۸ ماهه آناهیکس", Badge: "ارزش خرید بالا", Highlight: true},
		},
		Features: []types.ComparisonFeature{
			{Name: "مدل هوش مصنوعی پایه", Category: "مدل و هوش", Values: map[string]string{"free": "Gemini 1.5 Flash", "advanced": "Gemini 1.5 Pro با عمیق‌ترین سطح استدلال"}},
			{Name: "کانتکست ویندو (حافظه تحلیلی متن)", Category: "حافظه متنی", Values: map[string]string{"free": "۳۲ هزار توکن", "advanced": "۱ الی ۲ میلیون توکن (تحلیل کتاب‌های چندصد صفحه‌ای و ویدیو)"}},
			{Name: "فضای ذخیره‌سازی ابری Google One", Category: "سرویس‌های ابری", Values: map[string]string{"free": "۱۵ گیگابایت اشتراکی", "advanced": "۲ ترابایت (۲,۰۰۰ گیگابایت) ابری گوگل درایو"}},
			{Name: "یکپارچگی هوشمند با Google Docs و Gmail", Category: "بهره‌وری Workspace", Values: map[string]string{"free": "غیرفعال", "advanced": "دستیار نگارش و خلاصه‌سازی هوشمند در جیمیل و داکس"}},
			{Name: "تحویل و شیوه فعال‌سازی", Category: "نحوه دریافت", Values: map[string]string{"free": "رایگان", "advanced": "تحویل ۰ ثانیه‌ای لینک بدون نیاز به ارائه پسورد"}},
		},
	}
	geminiCompJSON, _ := json.Marshal(geminiComp)

	geminiGuide := []types.ActivationStep{
		{
			StepNumber:  1,
			Title:       "پرداخت و دریافت ۰ ثانیه‌ای لینک",
			Description: "پس از تکمیل خرید در درگاه بانکی، لینک دعوت اختصاصی گوگل فوراً در صفحه پیگیری به شما نمایش داده می‌شود.",
			Icon:        "key",
			Tip:         "یک کپی از لینک به شماره همراه شما نیز پیامک خواهد شد.",
		},
		{
			StepNumber:  2,
			Title:       "ورود به جیمیل شخصی",
			Description: "مرورگر خود را باز کرده و مطمئن شوید که با همان اکانت گوگل (Gmail) که می‌خواهید اشتراک روی آن فعال شود لاگین هستید.",
			Icon:        "user-check",
			Tip:         "هیچ نیازی به دادن پسورد اکانت به آناهیکس نیست.",
		},
		{
			StepNumber:  3,
			Title:       "کلیک بر روی لینک و تایید عضویت",
			Description: "روی لینک کلیک کرده و دکمه Join Family یا پذیرش دعوت را بزنید تا ۲ ترابایت فضا و جمینای ادونس فعال گردد.",
			Icon:        "link",
			Tip:         "در صورت مشاهده خطای تحریم، DNS شکن را روی شبکه خود ست فرمایید.",
		},
		{
			StepNumber:  4,
			Title:       "شروع استفاده از Gemini Advanced",
			Description: "وارد gemini.google.com شوید و از کار با قدرتمندترین مدل هوش مصنوعی گوگل لذت ببرید.",
			Icon:        "sparkles",
			Tip:         "فضای ۲ ترابایت ابری به صورت خودکار در Google Drive و Google Photos اعمال می‌شود.",
		},
	}
	geminiGuideJSON, _ := json.Marshal(geminiGuide)

	db.Model(&model.Product{}).Where("slug = ?", "gemini-advanced-18m").Updates(map[string]interface{}{
		"is_active":        true,
		"meta_title":       "خرید لینک فعال‌سازی ۱۸ ماهه گوگل جمینای ادونس و گوگل وان | آناهیکس",
		"meta_description": "خرید اشتراک ۱۸ ماهه Google Gemini Advanced با ۲ ترابایت فضای ابری، تحویل ۰ ثانیه‌ای، فعال‌سازی با یک کلیک روی جیمیل شخصی بدون پسورد در آناهیکس.",
		"og_image_url":     "/assets/hero.png",
		"plans_json":       string(geminiPlansJSON),
		"comparison_json":  string(geminiCompJSON),
		"guide_json":       string(geminiGuideJSON),
	})

	// 3. Claude Pro Ready
	claudePlans := []types.ProductPlan{
		{
			ID:                 "claude-1m-ready",
			Title:              "اکانت اختصاصی آماده تحویل آنی ۱ ماهه",
			DurationDays:       30,
			DurationText:       "۱ ماهه آماده",
			PriceToman:         1890000,
			OriginalPriceToman: 2250000,
			BaseCostUSD:        20.0,
			Badge:              "پیشنهاد برنامه‌نویسان",
			IsDefault:          true,
			DeliveryType:       "instant",
			Features: []string{
				"دسترسی کامل به مدل شگفت‌انگیز Claude 3.5 Sonnet",
				"قابلیت اجرای زنده کدهای Artifacts",
				"حداقل ۵ برابر پیام بیشتر نسبت به نسخه رایگان",
				"ایمیل و پسورد اختصاصی با قابلیت تغییر مشخصات",
			},
		},
		{
			ID:                 "claude-1m-personal",
			Title:              "ارتقای مستقیم روی ایمیل شخصی شما",
			DurationDays:       30,
			DurationText:       "۱ ماهه شخصی",
			PriceToman:         1980000,
			OriginalPriceToman: 2400000,
			BaseCostUSD:        21.0,
			Badge:              "اکانت شخصی",
			IsDefault:          false,
			DeliveryType:       "manual_timed",
			Features: []string{
				"فعال‌سازی روی ایمیل شخصی توسط اپراتور",
				"گارانتی ۳۰ روزه تعویض و عدم قطعی",
				"پرداخت قانونی با کارت بین‌المللی",
			},
		},
	}
	claudePlansJSON, _ := json.Marshal(claudePlans)

	claudeComp := types.ProductComparison{
		Columns: []types.ComparisonColumn{
			{Key: "free", Title: "نسخه رایگان کلود", Subtitle: "استفاده محدود روزمره", Highlight: false},
			{Key: "pro", Title: "کلود پرو (Claude Pro)", Subtitle: "ارائه اختصاصی آناهیکس", Badge: "بهترین برای کدنویسی", Highlight: true},
		},
		Features: []types.ComparisonFeature{
			{Name: "دسترسی به Claude 3.5 Sonnet و Claude 3 Opus", Category: "مدل‌ها", Values: map[string]string{"free": "سقف محدود در ساعات کاری", "pro": "دسترسی کامل با اولویت پردازشی بالا"}},
			{Name: "سقف ارسال پیام در هر ۵ ساعت", Category: "محدودیت استفاده", Values: map[string]string{"free": "۱۰ الی ۱۵ پیام", "pro": "حداقل ۵ برابر بیشتر با پایداری کامل"}},
			{Name: "قابلیت شبیه‌سازی زنده نرم‌افزار (Artifacts)", Category: "توسعه فرانت‌اند", Values: map[string]string{"free": "بسیار پایه", "pro": "رندر بلادرنگ پروژه‌های React، HTML و SVG"}},
			{Name: "امکان ساخت پروژه‌ها و بارگذاری داکیومنت‌ها (Projects)", Category: "مدیریت حافظه", Values: map[string]string{"free": "غیرفعال", "pro": "فعال با حافظه پروژه‌محور اختصاصی"}},
			{Name: "پایداری هنگام ترافیک سنگین جهانی", Category: "کیفیت سرویس", Values: map[string]string{"free": "نمایش ارور Over Capacity", "pro": "پاسخ‌دهی سریع بدون معطلی"}},
		},
	}
	claudeCompJSON, _ := json.Marshal(claudeComp)

	db.Model(&model.Product{}).Where("slug = ?", "claude-pro-ready").Updates(map[string]interface{}{
		"is_active":        true,
		"meta_title":       "خرید اکانت کلود پرو اختصاصی (Claude 3.5 Sonnet) | آناهیکس",
		"meta_description": "خرید اکانت اختصاصی و آماده Claude Pro آنتروپیک با دسترسی کامل به Claude 3.5 Sonnet، قابلیت Artifacts و تحویل آنی در آناهیکس.",
		"og_image_url":     "/assets/hero.png",
		"plans_json":       string(claudePlansJSON),
		"comparison_json":  string(claudeCompJSON),
	})

	// 4. Cursor Pro
	cursorPlans := []types.ProductPlan{
		{
			ID:                 "cursor-1m-pro",
			Title:              "اشتراک ۱ ماهه Cursor Pro",
			DurationDays:       30,
			DurationText:       "۱ ماهه ویژه توسعه‌دهندگان",
			PriceToman:         1750000,
			OriginalPriceToman: 2150000,
			BaseCostUSD:        20.0,
			Badge:              "انتخاب برنامه‌نویسان",
			IsDefault:          true,
			DeliveryType:       "manual_timed",
			Features: []string{
				"۵۰۰ درخواست پریمیوم سریع در ماه (Fast Requests)",
				"دسترسی به تب جادویی Composer و حالت Agentic",
				"استفاده نامحدود از درخواست‌های معمولی (Slow Requests)",
				"دسترسی به قدرتمندترین مدل‌ها مانند Claude 3.5 Sonnet",
			},
		},
	}
	cursorPlansJSON, _ := json.Marshal(cursorPlans)

	db.Model(&model.Product{}).Where("slug = ?", "cursor-pro-1m").Updates(map[string]interface{}{
		"is_active":        true,
		"meta_title":       "خرید اشتراک ادیتور کرسر پرو (Cursor Pro AI) | آناهیکس",
		"meta_description": "خرید اشتراک ادیتور هوش مصنوعی Cursor Pro با دسترسی نامحدود، Composer، پشتیبانی کامل از Claude 3.5 Sonnet و فعال‌سازی قانونی در آناهیکس.",
		"og_image_url":     "/assets/hero.png",
		"plans_json":       string(cursorPlansJSON),
	})

	log.Println("🌱 Seeded rich PDP metadata, plan options, comparison matrices, and activation guides into products table")
}
