package cmd

import (
	"log"
	"time"

	"github.com/EbadiDev/anahix-server/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func seedContentData(db *gorm.DB) {
	// 1. Seed Categories
	var catCount int64
	db.Model(&model.ContentCategory{}).Count(&catCount)
	if catCount == 0 {
		categories := []model.ContentCategory{
			// Blog Categories
			{
				ID:          uuid.New(),
				Name:        "آموزش و ترفندها",
				Slug:        "tutorials",
				Type:        "blog",
				Icon:        "GraduationCap",
				Description: "آموزش گام به گام استفاده از ابزارهای هوش مصنوعی، پرامپت‌نویسی و فعال‌سازی سرویس‌ها",
				SortOrder:   1,
				IsActive:    true,
			},
			{
				ID:          uuid.New(),
				Name:        "مقایسه مدل‌ها و بنچمارک",
				Slug:        "comparisons",
				Type:        "blog",
				Icon:        "Scale",
				Description: "مقایسه تخصصی GPT-4o، مدل‌های o3/o1، Claude 3.7 و Gemini Advanced",
				SortOrder:   2,
				IsActive:    true,
			},
			{
				ID:          uuid.New(),
				Name:        "اخبار و تازه‌های هوش مصنوعی",
				Slug:        "ai-news",
				Type:        "blog",
				Icon:        "Sparkles",
				Description: "جدیدترین آپدیت‌ها و رونمایی‌های شرکت‌های OpenAI، Anthropic، Google و DeepSeek",
				SortOrder:   3,
				IsActive:    true,
			},
			{
				ID:          uuid.New(),
				Name:        "امنیت و ضدتحریم",
				Slug:        "security-access",
				Type:        "blog",
				Icon:        "ShieldCheck",
				Description: "روش‌های امن دور زدن محدودیت‌های منطقه‌ای و محافظت از اکانت‌ها در برابر بلاک شدن",
				SortOrder:   4,
				IsActive:    true,
			},

			// Help Guides Categories
			{
				ID:          uuid.New(),
				Name:        "رفع ارورها و خطاهای تحریم",
				Slug:        "troubleshooting",
				Type:        "help",
				Icon:        "AlertTriangle",
				Description: "حل خطاهای Access Denied، Cloudflare Turnstile، شماره تماس و لو رفتن لوکیشن",
				SortOrder:   1,
				IsActive:    true,
			},
			{
				ID:          uuid.New(),
				Name:        "راهنمای فعال‌سازی و ورود",
				Slug:        "activation-setup",
				Type:        "help",
				Icon:        "KeyRound",
				Description: "نحوه ورود به اکانت‌های اشتراکی و اختصاصی، لاگین و دریافت کد ۲FA",
				SortOrder:   2,
				IsActive:    true,
			},
			{
				ID:          uuid.New(),
				Name:        "شماره مجازی و پیامک",
				Slug:        "virtual-numbers",
				Type:        "help",
				Icon:        "Radio",
				Description: "نکات مهم خرید شماره مجازی تلگرام، گوگل و اوپن ای‌آی و استرداد وجه",
				SortOrder:   3,
				IsActive:    true,
			},
			{
				ID:          uuid.New(),
				Name:        "گارانتی و قوانین تعویض",
				Slug:        "warranty-policy",
				Type:        "help",
				Icon:        "FileCheck",
				Description: "شرایط گارانتی تا آخرین روز اشتراک، جایگزینی لایسنس و پشتیبانی",
				SortOrder:   4,
				IsActive:    true,
			},
		}

		for _, c := range categories {
			db.Create(&c)
		}
		log.Println("🌱 Seeded content categories for Blog and Knowledge Base")
	}

	// Ensure Product Categories (Type: "product") are present in content_categories table
	var prodCatCount int64
	db.Model(&model.ContentCategory{}).Where("type = ?", "product").Count(&prodCatCount)
	if prodCatCount == 0 {
		productCategories := []model.ContentCategory{
			{
				ID:          uuid.New(),
				Name:        "هوش مصنوعی و چت‌بات",
				Slug:        "ai_chat",
				Type:        "product",
				Icon:        "Bot",
				Description: "اشتراک قانونی ChatGPT Plus، Claude 3.7 Sonnet و Google Gemini Advanced",
				SortOrder:   1,
				IsActive:    true,
			},
			{
				ID:          uuid.New(),
				Name:        "دستیارهای کدنویسی",
				Slug:        "coding_ai",
				Type:        "product",
				Icon:        "Code2",
				Description: "اکانت‌های تخصصی برنامه‌نویسی Cursor AI Pro، GitHub Copilot و Replit",
				SortOrder:   2,
				IsActive:    true,
			},
			{
				ID:          uuid.New(),
				Name:        "ارتقای اکانت شخصی",
				Slug:        "personal_upgrade",
				Type:        "product",
				Icon:        "Zap",
				Description: "فعال‌سازی بدون پسورد و ۱۰۰٪ پایدار روی ایمیل و اکانت شخصی خریدار",
				SortOrder:   3,
				IsActive:    true,
			},
			{
				ID:          uuid.New(),
				Name:        "فضای ابری و گوگل وان",
				Slug:        "cloud_storage",
				Type:        "product",
				Icon:        "HardDrive",
				Description: "ارتقای حجم جیمیل و درایو اختصاصی ۲ ترابایت قانونی با هوش مصنوعی",
				SortOrder:   4,
				IsActive:    true,
			},
			{
				ID:          uuid.New(),
				Name:        "فیلم، سریال و استریم",
				Slug:        "film_series",
				Type:        "product",
				Icon:        "Tv",
				Description: "اشتراک‌های پریمیوم نتفلیکس، اسپاتیفای و یوتیوب پریمیوم بدون قطعی",
				SortOrder:   5,
				IsActive:    true,
			},
			{
				ID:          uuid.New(),
				Name:        "شماره‌های مجازی بین‌المللی",
				Slug:        "virtual_numbers",
				Type:        "product",
				Icon:        "Radio",
				Description: "دریافت آنی پیامک تایید تلگرام، گوگل، اوپن‌ای‌آی و پلتفرم‌های خارجی",
				SortOrder:   6,
				IsActive:    true,
			},
		}

		for _, pc := range productCategories {
			db.Create(&pc)
		}
		log.Println("🌱 Seeded dynamic product categories in content_categories table")
	}

	// Ensure Subcategories are seeded for rich nested navigation
	var subCatCount int64
	db.Model(&model.ContentCategory{}).Where("parent_id IS NOT NULL").Count(&subCatCount)
	if subCatCount == 0 {
		var aiChatCat, codingCat, filmCat model.ContentCategory
		db.Where("slug = ? AND type = ?", "ai_chat", "product").First(&aiChatCat)
		db.Where("slug = ? AND type = ?", "coding_ai", "product").First(&codingCat)
		db.Where("slug = ? AND type = ?", "film_series", "product").First(&filmCat)

		var subCategories []model.ContentCategory

		if aiChatCat.ID != uuid.Nil {
			pID := aiChatCat.ID
			subCategories = append(subCategories,
				model.ContentCategory{
					ID:          uuid.New(),
					ParentID:    &pID,
					Name:        "چت‌جی‌پی‌تی پلاس (ChatGPT Plus & o3)",
					Slug:        "chatgpt",
					Type:        "product",
					Icon:        "Bot",
					Description: "اکانت قانونی چت‌جی‌پی‌تی با دسترسی نامحدود به مدل‌های GPT-4o و استدلالی o3",
					SortOrder:   1,
					IsActive:    true,
				},
				model.ContentCategory{
					ID:          uuid.New(),
					ParentID:    &pID,
					Name:        "انتروپیک کلود پرو (Claude 3.7 Sonnet)",
					Slug:        "claude",
					Type:        "product",
					Icon:        "Bot",
					Description: "قدرتمندترین مدل استدلال و کدنویسی جهان با پنجره کانتکست ۲۰۰ هزار توکن",
					SortOrder:   2,
					IsActive:    true,
				},
				model.ContentCategory{
					ID:          uuid.New(),
					ParentID:    &pID,
					Name:        "گوگل جمینای ادونس (Gemini Advanced)",
					Slug:        "gemini",
					Type:        "product",
					Icon:        "Sparkles",
					Description: "اشتراک ۱۸ ماهه جمینای با ۲ ترابایت حافظه ابری گوگل درایو و جیمیل",
					SortOrder:   3,
					IsActive:    true,
				},
			)
		}

		if codingCat.ID != uuid.Nil {
			pID := codingCat.ID
			subCategories = append(subCategories,
				model.ContentCategory{
					ID:          uuid.New(),
					ParentID:    &pID,
					Name:        "کرسر پرو هوش مصنوعی (Cursor Pro)",
					Slug:        "cursor",
					Type:        "product",
					Icon:        "Code2",
					Description: "ادیتور انقلابی هوش مصنوعی بر پایه VS Code با ۵۰۰ درخواست پرمیوم ماهانه",
					SortOrder:   1,
					IsActive:    true,
				},
				model.ContentCategory{
					ID:          uuid.New(),
					ParentID:    &pID,
					Name:        "گیت‌هاب کوپایلوت (GitHub Copilot)",
					Slug:        "github_copilot",
					Type:        "product",
					Icon:        "Code2",
					Description: "دستیار رسمی برنامه‌نویسی مایکروسافت با تکمیل خودکار هوشمند چندخطی",
					SortOrder:   2,
					IsActive:    true,
				},
			)
		}

		if filmCat.ID != uuid.Nil {
			pID := filmCat.ID
			subCategories = append(subCategories,
				model.ContentCategory{
					ID:          uuid.New(),
					ParentID:    &pID,
					Name:        "نتفلیکس پریمیوم (Netflix 4K UHD)",
					Slug:        "netflix",
					Type:        "product",
					Icon:        "Tv",
					Description: "اکانت اختصاصی بدون قطعی با کیفیت اولترا اچ‌دی و پشتیبانی ۴ کاربر همزمان",
					SortOrder:   1,
					IsActive:    true,
				},
				model.ContentCategory{
					ID:          uuid.New(),
					ParentID:    &pID,
					Name:        "اسپاتیفای پریمیوم (Spotify Premium)",
					Slug:        "spotify",
					Type:        "product",
					Icon:        "Radio",
					Description: "موسیقی بدون تبلیغات، دانلود آفلاین با بالاترین کیفیت صدای ۳۲۰kbps",
					SortOrder:   2,
					IsActive:    true,
				},
				model.ContentCategory{
					ID:          uuid.New(),
					ParentID:    &pID,
					Name:        "یوتیوب پریمیوم (YouTube Premium)",
					Slug:        "youtube",
					Type:        "product",
					Icon:        "Tv",
					Description: "حذف کامل تبلیغات یوتیوب، پخش در پس‌زمینه و دسترسی رایگان به YouTube Music",
					SortOrder:   3,
					IsActive:    true,
				},
			)
		}

		for _, sub := range subCategories {
			db.Create(&sub)
		}
		log.Println("🌱 Seeded rich nested subcategories in content_categories table")
	}

	// Find some product slugs to link
	var geminiProduct model.Product
	db.Where("slug LIKE ?", "%gemini%").First(&geminiProduct)

	var chatgptProduct model.Product
	db.Where("slug LIKE ?", "%chatgpt%").First(&chatgptProduct)

	var claudeProduct model.Product
	db.Where("slug LIKE ?", "%claude%").First(&claudeProduct)

	// 2. Seed High-Value SEO Blog Articles
	var articleCount int64
	db.Model(&model.Article{}).Count(&articleCount)
	if articleCount == 0 {
		now := time.Now()

		articles := []model.Article{
			{
				ID:           uuid.New(),
				Title:        "آموزش خرید جمینای ادونسد (Gemini Advanced) با فضای ۲ ترابایت ابری گوگل",
				Slug:         "how-to-buy-gemini-advanced-iran",
				Excerpt:      "بررسی کامل قابلیت‌های مدل Gemini 1.5 Pro و Ultra، نحوه فعال‌سازی روی ایمیل شخصی یا اکانت آماده بدون ریسک بن، و ادغام عمیق با گوگل درایو و داکس.",
				CategorySlug: "tutorials",
				CategoryName: "آموزش و ترفندها",
				CoverImage:   "https://images.unsplash.com/photo-1677442136019-21780ecad995?auto=format&fit=crop&w=1200&q=80",
				Tags:         `["جمینای ادونسد", "گوگل وان", "خرید هوش مصنوعی", "Gemini 1.5 Pro", "آموزش هوش مصنوعی"]`,
				AuthorName:   "تیم فنی آناهیکس",
				AuthorRole:   "متخصص راهکارهای کلود و هوش مصنوعی",
				AuthorAvatar: "https://images.unsplash.com/photo-1534528741775-53994a69daeb?auto=format&fit=crop&w=200&q=80",
				ReadingTimeMinutes: 6,
				ViewCount:          1420,
				LikeCount:          89,
				IsFeatured:         true,
				IsPublished:        true,
				PublishedAt:        &now,
				RelatedProductTitle: "اشتراک جمینای ادونسد (Gemini Advanced 18 Months)",
				RelatedProductSlug:  "gemini-advanced-18m",
				RelatedProductPrice: 1450000,
				MetaTitle:           "راهنمای خرید جمینای ادونسد در ایران | فعال‌سازی ۱۸ ماهه بدون ریسک",
				MetaDescription:     "آموزش گام به گام و سریع خرید اکانت Gemini Advanced با لایسنس معتبر گوگل و ۲ ترابایت فضای ابری گوگل وان در آناهیکس.",
				MetaKeywords:        "آموزش خرید جمینای ادونسد, خرید Gemini Advanced, فعالسازی جمینای در ایران, گوگل وان ۲ ترابایت",
				Content: `## مقدمه: چرا جمینای ادونسد به انتخاب اول کاربران حرفه‌ای تبدیل شده؟

شرکت گوگل با معرفی مدل‌های پیشرفته **Gemini 1.5 Pro** و قابلیت پردازش شگفت‌انگیز **۱ تا ۲ میلیون توکن کانتکست (Context Window)**، استانداردهای دنیای هوش مصنوعی را بازتعریف کرده است. با این حجم کانتکست، شما می‌توانید یک کتاب کامل ۵۰۰ صفحه‌ای، یک مخزن کد بزرگ یا چندین ساعت ویدیوی آموزشی را مستقیماً آپلود کرده و تحلیلی بی‌نقص دریافت کنید.

علاوه بر این، با خرید این اشتراک، **۲ ترابایت فضای ذخیره‌سازی ابری Google One** به همراه یکپارچگی اختصاصی در Google Docs، Gmail و Google Sheets به شما تعلق می‌گیرد.

---

## تفاوت نسخه رایگان و نسخه Gemini Advanced

| ویژگی | جمینای رایگان (Flash) | جمینای ادونسد (Pro / Ultra) |
| :--- | :--- | :--- |
| **مدل محاسباتی** | Gemini 1.5 Flash | Gemini 1.5 Pro با استدلال عمیق |
| **پنجره کانتکست** | ۳۲,۰۰۰ توکن | تا ۲,۰۰۰,۰۰۰ توکن همزمان |
| **فضای ابری گوگل وان** | ۱۵ گیگابایت پایه | **۲,۰۰۰ گیگابایت (۲ ترابایت)** |
| **تحلیل فایل‌های سنگین** | محدود | فایل‌های متنی، صوتی، ویدیو و کد |
| **ادغام در جیمیل و داکس** | ندارد | فعال به عنوان دستیار هوشمند |

---

## روش‌های تهیه اکانت در ایران و چالش‌های تحریم

با توجه به محدودیت‌های پرداخت بین‌المللی با کارت‌های عضو شبکه شتاب و حساسیت بالای گوگل روی لوکیشن ایران، فعال‌سازی مستقیم با کارت بانکی معمولی غیرممکن است. دو شیوه متداول در مارکت ایران وجود دارد:

1. **اکانت اختصاصی آماده (پیشنهادی)**: یک حساب جیمیل تایید شده و تازه ساخته شده که اشتراک ۱۸ ماهه قانونی جمینای ادونسد از قبل روی آن شارژ شده و تحویل آنی می‌شود.
2. **ارتقای ایمیل شخصی از طریق فمیلی پلن**: نیاز به دعوت‌نامه خانوادگی دارد اما ممکن است در صورت داشتن سابقه عضویت در فمیلی‌های قبلی با ارور ۱۲ ماهه تغییر فمیلی مواجه شوید.

### مراحل امن لاگین به حساب جمینای ادونسد:
1. ابتدا مرورگر خود را باز کرده و یک Profile جدید (مثلاً در Chrome یا Edge) بسازید.
2. از یک ابزار تغییر IP پایدار با لوکیشن کشورهای معتبر (آلمان، سوئیس، انگلیس یا آمریکا) استفاده کنید.
3. به آدرس **gemini.google.com** مراجعه کرده و با ایمیل و رمز عبور دریافتی از پنل آناهیکس وارد شوید.
4. در بالای صفحه نشان بنفش‌رنگ **Advanced** را مشاهده خواهید کرد که تاییدکننده فعال بودن پلن است.

---

## نکات طلایی برای جلوگیری از قطع شدن دسترسی

- **عدم استفاده از DNSها یا پروکسی‌های رایگان بی‌کیفیت**: ترجیحاً از کانکشن‌های استاتیک استفاده فرمایید.
- **فعال‌سازی تایید دو مرحله‌ای (2FA)**: پس از تحویل اکانت اختصاصی، شماره پشتیبان خود را جهت امنیت کامل ثبت کنید.
- **استفاده از هوش مصنوعی در موبایل**: اپلیکیشن رسمی Gemini را از گوگل پلی یا اپ استور دانلود کنید و بدون محدودیت از وویس مود تعاملی لذت ببرید.`,
			},
			{
				ID:           uuid.New(),
				Title:        "رفع قطعی خطای Access Denied و ۴۰۳ در کلود (Claude 3.7 Sonnet)",
				Slug:         "fix-claude-access-denied-error-403",
				Excerpt:      "راهنمای گام به گام رفع خطاهای آزاردهنده 'App unavailable in your region' و 'Access Denied 403' در پلتفرم کلاود آنتروپیک همراه با بررسی تنظیمات مرورگر و WebRTC.",
				CategorySlug: "security-access",
				CategoryName: "امنیت و ضدتحریم",
				CoverImage:   "https://images.unsplash.com/photo-1618005182384-a83a8bd57fbe?auto=format&fit=crop&w=1200&q=80",
				Tags:         `["کلود", "Claude 3.7", "Access Denied", "رفع ارور ۴۰۳", "آموزش هوش مصنوعی"]`,
				AuthorName:   "تیم فنی آناهیکس",
				AuthorRole:   "کارشناس امنیت سایبری و شبکه",
				AuthorAvatar: "https://images.unsplash.com/photo-1507003211169-0a1dd7228f2d?auto=format&fit=crop&w=200&q=80",
				ReadingTimeMinutes: 5,
				ViewCount:          2180,
				LikeCount:          143,
				IsFeatured:         true,
				IsPublished:        true,
				PublishedAt:        &now,
				RelatedProductTitle: "اکانت Claude Pro / Team اختصاصی",
				RelatedProductSlug:  "claude-pro-upgrade",
				RelatedProductPrice: 1980000,
				MetaTitle:           "رفع خطای Access Denied در کلود (Claude) | راهنمای ۱۰۰٪ تضمینی",
				MetaDescription:     "آموزش کامل حل ارور Access Denied و ۴۰۳ در سایت Claude.ai با رفع نشت WebRTC و تنظیمات کش مرورگر.",
				MetaKeywords:        "رفع خطای access denied در کلود, ارور ۴۰۳ claude, مشکل ورود به claude ai, تحریم کلود در ایران",
				Content: `## چرا سایت Claude.ai به کاربران ایرانی خطای Access Denied می‌دهد؟

کمپانی **Anthropic** (توسعه‌دهنده مدل‌های هوش مصنوعی Claude) یکی از سخت‌گیرانه‌ترین سیاست‌های ژئوبلاک و فیلترینگ منطقه‌ای را پیاده‌سازی کرده است. سیستم‌های حفاظتی Cloudflare در لبه شبکه Claude نه تنها آدرس IP شما را چک می‌کنند، بلکه پروتکل‌های لایه ۷، هدرهای مرورگر، تایم‌زون سیستم، زبان سیستم‌عامل و نشت **WebRTC** را به طور پیوسته مانیتور می‌نمایند.

اگر حتی برای یک میلی‌ثانیه لوکیشن واقعی شما لو برود، با خطای زیر روبه‌رو خواهید شد:
> **"App unavailable in your country or region"** یا **"Access Denied (Error 1020 / 403)"**

---

## گام اول: جلوگیری از نشت آدرس IP واقعی از طریق WebRTC

پروتکل WebRTC در مرورگرها به سایت‌ها اجازه می‌دهد آدرس IP محلی و حتی IP عمومی پشت پروکسی را بخوانند. برای بستن این حفره امنیتی:

1. در مرورگر فایرفاکس عبارت **about:config** را در آدرس‌بار بنویسید.
2. عبارت **media.peerconnection.enabled** را سرچ کرده و مقدار آن را به **false** تغییر دهید.
3. در مرورگر کروم، اکستنشن معتبر **WebRTC Control** را از Chrome Web Store نصب کرده و حالت آن را روی دکمه آبی (Block WebRTC) بگذارید.
4. با مراجعه به سایت **browserleaks.com/webrtc** مطمئن شوید که هیچ IP ایرانی نمایش داده نمی‌شود.

---

## گام دوم: حل تعارض منطقه زمانی (Time Zone) و زبان مرورگر

یکی از شایع‌ترین نشانه‌هایی که باعث بلاک فوری در کلاود می‌شود، تضاد بین IP سرور (مثلاً آلمان) با ساعت سیستم شما (تهران +03:30) است.

- ساعت سیستم ویندوز یا مک خود را روی **Set time automatically** بگذارید یا موقتاً با لوکیشن سرور خود همگام کنید.
- در تنظیمات مرورگر، زبان اول را روی **English (United States)** قرار دهید.

---

## گام سوم: پاکسازی کامل کوکی‌ها و استفاده از سشن تمیز

اگر قبلاً خطای Access Denied را دیده‌اید، این خطا در LocalStorage و کوکی مرورگر شما ذخیره شده است. برای حل آن:
1. در صفحه claude.ai کلیدهای ترکیبی **Ctrl + Shift + Delete** را بزنید.
2. بازه زمانی را روی **All time** گذاشته و کوکی‌ها و کش را کاملاً پاک کنید.
3. مرورگر را ببندید، ابزار تغییر IP را متصل کنید و ترجیحاً از یک تب Incognito یا پروفایل تازه استفاده نمایید.

---

## جمع‌بندی و راهکار پایدار
برای توسعه‌دهندگان و نویسندگانی که کار روزمره‌شان وابسته به کدنویسی و تحلیل با **Claude 3.7 Sonnet** است، استفاده از اشتراک‌های آماده قانونی آناهیکس همراه با گارانتی تعویض و سرورهای ضدتحریم اختصاصی پایدارترین گزینه بدون دغدغه قطعی است.`,
			},
			{
				ID:           uuid.New(),
				Title:        "تفاوت GPT-4o با مدل‌های استدلالی OpenAI o3 و o1: کدام برای شما مناسب‌تر است؟",
				Slug:         "gpt-4o-vs-o3-vs-o1-benchmark-comparison",
				Excerpt:      "مقایسه جامع سرعت پاسخ‌دهی، توانایی استدلال زنجیره‌ای تفکر (Chain-of-Thought)، حل مسائل ریاضی و المپیادی، و مهندسی معکوس کد بین GPT-4o و نسل جدید مدل‌های o3/o1.",
				CategorySlug: "comparisons",
				CategoryName: "مقایسه مدل‌ها و بنچمارک",
				CoverImage:   "https://images.unsplash.com/photo-1620712943543-bcc4688e7485?auto=format&fit=crop&w=1200&q=80",
				Tags:         `["GPT-4o", "OpenAI o3", "OpenAI o1", "مقایسه هوش مصنوعی", "کدنویسی AI"]`,
				AuthorName:   "تحریریه آناهیکس",
				AuthorRole:   "پژوهشگر هوش مصنوعی و داده",
				AuthorAvatar: "https://images.unsplash.com/photo-1535713875002-d1d0cf377fde?auto=format&fit=crop&w=200&q=80",
				ReadingTimeMinutes: 7,
				ViewCount:          3140,
				LikeCount:          210,
				IsFeatured:         true,
				IsPublished:        true,
				PublishedAt:        &now,
				RelatedProductTitle: "اکانت ChatGPT Plus با پشتیبانی از مدل‌های o1 و o3",
				RelatedProductSlug:  "chatgpt-plus-personal",
				RelatedProductPrice: 1890000,
				MetaTitle:           "تفاوت GPT-4o با OpenAI o3 و o1 | بنچمارک و راهنمای انتخاب",
				MetaDescription:     "مقایسه کامل قابلیت‌ها و بنچمارک‌های GPT-4o و مدل استدلالی OpenAI o3 برای برنامه‌نویسان، محققان و کاربران حرفه‌ای.",
				MetaKeywords:        "تفاوت GPT-4o با o3, مدل o1 اوپن ای آی, مقایسه چت جی پی تی, خرید اکانت chatgpt plus",
				Content: `## انقلاب مدل‌های استدلالی (Reasoning Models) در OpenAI

برای سال‌ها، مدل‌های زبان بزرگ (LLMs) بر پایه پیش‌بینی توکن بعدی با سرعت بالا کار می‌کردند. اما انتشار خانواده مدل‌های **OpenAI o1** و نسخه ارتقایافته **o3** تحولی بنیادین به نام **زنجیره افکار پنهان (Hidden Chain-of-Thought)** ایجاد کرد. این مدل‌ها به جای پاسخ شتاب‌زده، چند ثانیه تا چند ده ثانیه «فکر می‌کنند»، فرضیه‌ها را اعتبارسنجی می‌کنند و سپس پاسخی دقیق تولید می‌نمایند.

---

## جدول مقایسه جامع: GPT-4o در برابر OpenAI o1 و OpenAI o3

| شاخص | GPT-4o (Omni) | OpenAI o1 | OpenAI o3 (پیشرفته‌ترین) |
| :--- | :--- | :--- | :--- |
| **سرعت پاسخ‌دهی** | بسیار سریع (آنی) | کندتر (۱۰ تا ۳۰ ثانیه تفکر) | متوسط تا کنترل‌شده |
| **ورودی چندوجهی** | متن، صوت و تصویر زنده | فقط متنی و تصویری کد | متن، کد پیشرفته و بینایی ماشین |
| **مهارت در ریاضیات و فیزیک** | خوب (سطح دبیرستان/دانشگاه) | عالی (۸۴٪ در مسابقات ریاضی AIME) | **فوق‌العاده (۹۶.۷٪ مدال طلای المپیاد)** |
| **تولید محتوا و نگارش متون** | روان، لحن طبیعی و غنی | خشک و بسیار منطقی | دقیق و چندلایه |
| **دیباگ کدهای پیچیده نرم‌افزاری** | مناسب کدهای متوسط | عالی در عیب‌یابی معماری | **شاهکار در بازسازی الگوها و ریفکتور** |
| **سقف مصرف در اکانت پلاس** | بالا (۸۰ پیام در ۳ ساعت) | محدود (حدود ۵۰ پیام در هفته) | به صورت دسترسی اختصاصی Pro |

---

## کدام مدل برای کاربرد شما مناسب است؟

### سناریو ۱: برای چت روزمره، خلاصه‌سازی و نویسندگی (انتخاب: GPT-4o)
اگر به دنبال پاسخ سریع برای نوشتن ایمیل، تولید پست‌های شبکه‌های اجتماعی، خلاصه‌سازی مقالات طولانی یا مکالمه صوتی طبیعی هستید، **GPT-4o** با فاصله بهترین و ارزان‌ترین گزینه است.

### سناریو ۲: برای تحلیل الگوریتم‌ها و حل باگ‌های غیرقابل حل (انتخاب: o1 و o3)
زمانی که یک الگوریتم پیچیده توزیع‌شده، اثبات روابط ریاضی، مسائل رمزنگاری یا معماری نرم‌افزار بزرگ دارید که مدل‌های معمولی در آن دچار توهم (Hallucination) می‌شوند، مدل **o3** مانند یک مهندس ارشد نرم‌افزار در کنار شما می‌اندیشد.

---

## نحوه دسترسی به این مدل‌ها در ایران

کاربران نسخه رایگان تنها به تعداد محدودی از پرسش‌های GPT-4o دسترسی دارند و مدل‌های o1 و o3 در نسخه رایگان فعال نیستند. با خرید **اشتراک ChatGPT Plus** در پلتفرم آناهیکس، دسترسی کامل به تمامی این مدل‌ها همراه با قابلیت ساخت GPTهای سفارشی برای شما آنی فراهم می‌شود.`,
			},
		}

		for _, a := range articles {
			if geminiProduct.ID != uuid.Nil && a.Slug == "how-to-buy-gemini-advanced-iran" {
				a.RelatedProductID = &geminiProduct.ID
			}
			if claudeProduct.ID != uuid.Nil && a.Slug == "fix-claude-access-denied-error-403" {
				a.RelatedProductID = &claudeProduct.ID
			}
			if chatgptProduct.ID != uuid.Nil && a.Slug == "gpt-4o-vs-o3-vs-o1-benchmark-comparison" {
				a.RelatedProductID = &chatgptProduct.ID
			}
			db.Create(&a)
		}
		log.Println("🌱 Seeded high-value SEO blog articles")
	}

	// 3. Seed Help Guides (Knowledge Base)
	var guideCount int64
	db.Model(&model.HelpGuide{}).Count(&guideCount)
	if guideCount == 0 {
		now := time.Now()

		guides := []model.HelpGuide{
			{
				ID:           uuid.New(),
				Title:        "راهنمای رفع مشکل عدم دریافت پیامک در شماره‌های مجازی ۵SIM",
				Slug:         "troubleshoot-virtual-number-sms-delivery",
				Excerpt:      "اگر بعد از خرید شماره مجازی پیامک کد تایید برای تلگرام یا گوگل ارسال نشد چه کاری باید انجام دهیم؟ راهنمای انصراف خودکار و بازگشت ۱۰۰٪ وجه.",
				CategorySlug: "virtual-numbers",
				CategoryName: "شماره مجازی و پیامک",
				Icon:         "Radio",
				Difficulty:   "ساده",
				ServiceTarget: "virtual_number",
				ReadingTimeMinutes: 3,
				ViewCount:          850,
				HelpfulYes:         184,
				HelpfulNo:          6,
				SortOrder:          1,
				IsFeatured:         true,
				IsPublished:        true,
				PublishedAt:        &now,
				MetaTitle:           "راهنمای عدم دریافت کد در شماره مجازی ۵SIM | بازگشت وجه آنی",
				MetaDescription:     "بررسی دلایل عدم ارسال SMS در شماره‌های مجازی تلگرام و اینستاگرام و نحوه لغو و استرداد موجودی کیف پول.",
				MetaKeywords:        "عدم دریافت پیامک شماره مجازی, کد تلگرام شماره مجازی, بازگشت وجه 5sim, خرید شماره مجازی",
				Content: `## چرا در برخی مواقع کد تایید SMS ارسال نمی‌شود؟

در سامانه‌های بین‌المللی ارائه‌دهنده شماره مجازی (نظیر ۵SIM)، ارتباط بین اپراتور مخابراتی کشور مقصد و سرویس مقصد (تلگرام، گوگل، واتساپ، OpenAI) به صورت زنده برقرار می‌شود. در شرایط زیر ممکن است پیامک در بازه ۱۵ دقیقه‌ای نرسد:

1. **فشار بار روی سرورهای تلگرام یا گوگل**: در ساعات شلوغی، سیستم ضداسپم تلگرام ممکن است ارسال کد را تا چند دقیقه به تعویق بیندازد.
2. **استفاده از نسخه‌های غیررسمی تلگرام**: اگر از تلگرام‌های غیررسمی استفاده کنید، تلگرام کد را به جای پیامک به عنوان اعلان روی دستگاه دیگر می‌فرستد.
3. **تغییر سرفصل اپراتور**: برخی اپراتورهای ارزان‌قیمت در برخی کشورها بلاک هستند.

---

## گارانتی بازگشت ۱۰۰٪ وجه در آناهیکس

**مهم‌ترین مزیت آناهیکس تضمین بدون ریسک است**:
- تایمر دریافت پیامک برای هر شماره **۱۵ دقیقه** است.
- اگر تا پایان این زمان کدی دریافت نکنید یا دکمه **انصراف و لغو شماره** را بزنید، **تمام مبلغ پرداختی در همان ثانیه به کیف پول شما برگشت داده می‌شود**.
- هیچ‌گونه کارمزدی برای شماره‌های لغو شده کسر نخواهد شد!

---

## ترفندهای کاربردی برای دریافت ۱۰۰٪ سریع کد

- **تغییر اپراتور شماره**: هنگام خرید شماره در تب اپراتورها، به جای گزینه ارزان‌ترین اپراتور، اپراتورهای معتبر مانند **virtual66** یا **virtual34** را تست کنید.
- **انتخاب کشورهای کم‌ترافیک**: کشورهایی مانند فنلاند، لهستان، برزیل، ویتنام و مالزی درصد دریافت کد بالاتری نسبت به شماره‌های شلوغ دارند.
- **استفاده از نسخه رسمی Telegram یا Telegram X**: همیشه از اپلیکیشن رسمی برای دریافت کد استفاده فرمایید.`,
			},
			{
				ID:           uuid.New(),
				Title:        "نحوه ورود به اکانت آماده جمینای ادونسد و استفاده از ۲ ترابایت ابری",
				Slug:         "setup-and-login-gemini-advanced-ready-account",
				Excerpt:      "آموزش تصویری لاگین به جیمیل اختصاصی تحویل داده شده توسط ربات آناهیکس و دسترسی به Google One و Gemini Pro بدون ارور تایید هویت.",
				CategorySlug: "activation-setup",
				CategoryName: "راهنمای فعال‌سازی و ورود",
				Icon:         "KeyRound",
				Difficulty:   "ساده",
				ServiceTarget: "gemini",
				ReadingTimeMinutes: 4,
				ViewCount:          1120,
				HelpfulYes:         240,
				HelpfulNo:          8,
				SortOrder:          2,
				IsFeatured:         true,
				IsPublished:        true,
				PublishedAt:        &now,
				MetaTitle:           "آموزش ورود به اکانت آماده جمینای ادونسد | گام به گام",
				MetaDescription:     "راهنمای سریع و مطمئن ورود به ایمیل و رمز اکانت جمینای ادونسد خریداری شده از آناهیکس.",
				MetaKeywords:        "ورود به اکانت جمینای, لاگین gemini advanced, گوگل وان ۲ ترابایت, آموزش gemini",
				Content: `## مشخصات تحویلی در پنل سفارشات

بلافاصله پس از تکمیل پرداخت سفارش در آناهیکس، اطلاعات اکانت در بخش **پیگیری سفارش** و **پنل کاربری** به فرمت زیر نمایش داده می‌شود:
- **ایمیل (Gmail اختصاصی)**: example@gmail.com
- **رمز عبور (Password)**: Password123
- **ایمیل یا کد بازیابی (Recovery Email)**: backup@domain.com

---

## مراحل گام‌به‌گام ورود ایمن

### گام اول: ساخت پروفایل مجزا در مرورگر
برای جلوگیری از تداخل جیمیل‌های قبلی شما با اکانت جدید:
1. در بالای مرورگر کروم روی تصویر پروفایل خود کلیک کرده و گزینه **Add profile** را بزنید.
2. گزینه **Continue without an account** را انتخاب کرده و یک نام دلخواه بگذارید.

### گام دوم: اتصال به IP معتبر
قبل از وارد کردن مشخصات، مطمئن شوید ابزار تغییر IP شما روشن و پایدار است.

### گام سوم: لاگین در Accounts.google.com
1. به سایت **accounts.google.com** بروید و ایمیل و پسورد را وارد کنید.
2. اگر گوگل از شما سوال امنیتی پرسید، گزینه **Confirm your recovery email** را انتخاب کنید و ایمیل بازیابی مندرج در فاکتور سفارش را کپی نمایید (نیازی به ورود به ایمیل بازیابی نیست، فقط تایپ آدرس آن کافی است).

### گام چهارم: لذت از جمینای ادونسد و گوگل وان
به آدرس **gemini.google.com** بروید و از محیط هوشمند مدل Gemini 1.5 Pro با سقف ۲ میلیون توکن کانتکست استفاده کنید!`,
			},
			{
				ID:           uuid.New(),
				Title:        "شرایط گارانتی تعویض و پشتیبانی ۲۴ ساعته در آناهیکس",
				Slug:         "warranty-and-account-replacement-policy",
				Excerpt:      "توضیحات شفاف درباره گارانتی کامل اکانت‌ها تا آخرین روز اشتراک، تعهد زمان تحویل (SLA) و نحوه ثبت تیکت پشتیبانی فوری.",
				CategorySlug: "warranty-policy",
				CategoryName: "گارانتی و قوانین تعویض",
				Icon:         "ShieldCheck",
				Difficulty:   "ساده",
				ServiceTarget: "general",
				ReadingTimeMinutes: 3,
				ViewCount:          740,
				HelpfulYes:         195,
				HelpfulNo:          3,
				SortOrder:          3,
				IsFeatured:         false,
				IsPublished:        true,
				PublishedAt:        &now,
				MetaTitle:           "قوانین گارانتی و تعویض اکانت‌های هوش مصنوعی در آناهیکس",
				MetaDescription:     "تعهدات آناهیکس در ارائه گارانتی کامل، جایگزینی لایسنس و پشتیبانی اختصاصی برای تمامی خریداران.",
				MetaKeywords:        "گارانتی اکانت هوش مصنوعی, قوانین آناهیکس, پشتیبانی چت جی پی تی, تعویض اکانت کلود",
				Content: `## تعهد قطعی آناهیکس به مشتریان گرامی

یکی از دغدغه‌های اصلی کاربران ایرانی در تهیه اکانت‌های هوش مصنوعی، احتمال لغو اشتراک یا قفل شدن حساب توسط کمپانی‌های خارجی است. در آناهیکس تمامی لایسنس‌ها و اکانت‌ها تحت پوشش **گارانتی تعویض تا آخرین روز اشتراک** هستند.

---

## موارد تحت پوشش گارانتی

- هرگونه خروج حساب از حالت پرمیوم به دلیل تغییر سیاست‌های پلتفرم مقصد.
- خطاهای سیستمی روی اکانت که توسط پشتیبان قابل رفع نباشد.
- عدم تطابق لایسنس یا زمان اشتراک با فاکتور صادر شده.

در صورت بروز هر کدام از این موارد، اکانت جایگزین یا لایسنس تازه در سریع‌ترین زمان ممکن به صورت کاملاً رایگان در اختیار شما قرار می‌گیرد.

---

## نحوه ارتباط با بخش پشتیبانی
برای پیگیری تیکت‌ها و دریافت راهنمایی:
1. در پنل کاربری به تب **پشتیبانی و تیکت‌ها** مراجعه فرمایید.
2. تیکت جدید همراه با اسکرین‌شات از خطای مربوطه ارسال کنید.
3. در صورت نیاز به راهنمایی آنی می‌توانید با تلگرام پشتیبانی به نشانی **@AnahixSupport** در ارتباط باشید.`,
			},
		}

		for _, g := range guides {
			db.Create(&g)
		}
		log.Println("🌱 Seeded knowledge base help guides")
	}

	// 4. Seed Dynamic Content Writing Templates (Database-driven Templates)
	var templateCount int64
	db.Model(&model.ContentTemplate{}).Count(&templateCount)
	if templateCount == 0 {
		templates := []model.ContentTemplate{
			{
				ID:          uuid.New(),
				Title:       "قالب رفع ارور و عیب‌یابی (Troubleshooting)",
				Slug:        "troubleshoot-guide",
				Type:        "help",
				Icon:        "AlertTriangle",
				Description: "ساختار استاندارد برای تشریح ارور، پیش‌نیازها، راه‌حل گام به گام و گارانتی",
				SortOrder:   1,
				IsActive:    true,
				TemplateContent: `## شرح ارور و علت به وجود آمدن

هنگام استفاده از این سرویس، به دلیل تغییرات اخیر آی‌پی یا محدودیت‌های تحریمی، ممکن است کاربران با این پیام خطا مواجه شوند:
> **"Access Denied (Error 403) / Please try again later"**

---

## بررسی پیش‌نیازها قبل از حل ارور

- مطمئن شوید که ساعت سیستم عامل با لوکیشن سرور همگام است.
- از اتصال اینترنت بدون نشت DNS و WebRTC اطمینان حاصل کنید.

---

## گام اول: پاکسازی کش و کوکی‌های مرورگر

1. کلیدهای ترکیبی **Ctrl + Shift + Delete** را فشار دهید.
2. بازه زمانی را روی **All time** قرار داده و کش و کوکی‌ها را پاک کنید.
3. مرورگر را به طور کامل ببندید و مجدداً باز نمایید.

---

## گام دوم: تغییر روش اتصال و تنظیمات DNS

> [!TIP]
> استفاده از DNSهای اختصاصی ضدتحریم یا پروفایل مجزای کاربری در کروم تا ۹۵٪ این اختلال را برطرف می‌سازد.

---

## راهکار نهایی و جایگزین در آناهیکس

در صورتی که مشکل همچنان پابرجا بود، اکانت شما تحت **گارانتی تعویض آناهیکس** قرار دارد و می‌توانید فوراً از طریق ثبت تیکت یک اکانت جایگزین دریافت کنید.`,
			},
			{
				ID:          uuid.New(),
				Title:       "قالب مقایسه و بنچمارک هوش مصنوعی",
				Slug:        "ai-model-comparison",
				Type:        "blog",
				Icon:        "Columns",
				Description: "قالب اختصاصی جدول مقایسه، نقاط قوت و ضعف و راهنمای انتخاب",
				SortOrder:   2,
				IsActive:    true,
				TemplateContent: `## مقایسه اجمالی و تفاوت‌های کلیدی

در این مقاله قصد داریم دو مدل محبوب و پرکاربرد هوش مصنوعی را از ابعاد مختلف فنی، سرعت و دقت مقایسه کنیم.

---

## جدول مقایسه مشخصات و بنچمارک‌ها

| ویژگی و شاخص | سرویس اول | سرویس دوم |
| :--- | :--- | :--- |
| **سرعت پاسخ‌دهی** | بسیار سریع | با تامل و دقت بالاتر |
| **سقف توکن کانتکست** | ۱۲۸,۰۰۰ توکن | ۲,۰۰۰,۰۰۰ توکن |
| **هزینه اشتراک ماهانه** | مقرون‌به‌صرفه | حرفه‌ای |
| **پشتیبانی از تحلیل کد** | عالی | استثنایی با استدلال |

---

## نقاط قوت و سناریوهای کاربردی

- **برای چه کسانی مناسب است؟**: دانشجویان، برنامه‌نویسان و پژوهشگران داده.
- **نتیجه‌گیری خرید**: اگر اولویت شما سرعت است گزینه اول و اگر دقت استدلال اولویت دارد گزینه دوم پیشنهاد می‌شود.`,
			},
			{
				ID:          uuid.New(),
				Title:       "قالب آموزش گام‌به‌گام و راه‌اندازی",
				Slug:        "step-by-step-tutorial",
				Type:        "all",
				Icon:        "Lightbulb",
				Description: "آموزش گام‌به‌گام فعال‌سازی، پیش‌نیازها و نکات امنیتی",
				SortOrder:   3,
				IsActive:    true,
				TemplateContent: `## هدف از این آموزش

در این راهنمای جامع، فرآیند فعال‌سازی و راه‌اندازی این ابزار را از نقطه صفر تا استفاده حرفه‌ای بررسی می‌کنیم.

---

## پیش‌نیازهای شروع کار

1. یک مرورگر به‌روز (ترجیحاً Google Chrome یا Edge).
2. دریافت مشخصات لاگین اختصاصی از پنل کاربری آناهیکس.

---

## مراحل گام به گام فعال‌سازی

### مرحله اول: ورود به پورتال اصلی
ابتدا به صفحه ورود مراجعه کرده و اطلاعات دریافتی شامل ایمیل و رمزعبور را وارد کنید.

### مرحله دوم: تایید کد ۲FA در صورت نیاز
کد یک‌بار مصرف ارسالی به ایمیل یا کلید TOTP را وارد کنید تا هویت شما تایید شود.

> [!NOTE]
> نیازی به وارد کردن شماره تلفن شخصی ایران در هیچ‌کدام از مراحل فعال‌سازی وجود ندارد.

---

## جمع‌بندی و نکات امنیتی
همیشه پس از اتمام کار، از مرورگرهای مشترک خارج شوید و اطلاعات حساب کاربری خود را در اختیار دیگران قرار ندهید.`,
			},
			{
				ID:          uuid.New(),
				Title:       "قالب نقد و بررسی عمیق سرویس AI",
				Slug:        "ai-deep-dive-review",
				Type:        "blog",
				Icon:        "Sparkles",
				Description: "بررسی موشکافانه، معماری مدل، نقاط ضعف و سناریوهای واقعی تست",
				SortOrder:   4,
				IsActive:    true,
				TemplateContent: `## بررسی موشکافانه سرویس و جایگاه آن در بازار

در این مقاله تخصصی، به بررسی عمیق معماری مدل، محدودیت‌های نرخ مصرف (Rate Limits) و تجربیات کاربری می‌پردازیم.

---

## نقاط قوت برجسته
- رابط کاربری روان و بهینه‌سازی‌شده
- پایداری بالا در زمان بار ترافیکی سنگین
- ادغام با افزونه‌ها و ابزارهای توسعه‌دهنده

---

## محدودیت‌ها و چالش‌ها
- نیاز به آی‌پی پایدار جهت جلوگیری از بلاک
- محدودیت کانتکست در پرامپت‌های طولانی

> [!TIP]
> برای استفاده نامحدود بدون دغدغه قطعی، پیشنهاد می‌کنیم از اکانت‌های اختصاصی با پشتیبانی ۲۴ ساعته آناهیکس استفاده فرمایید.`,
			},
		}

		for _, tmpl := range templates {
			db.Create(&tmpl)
		}
		log.Println("🌱 Seeded dynamic content writing templates into DB")
	}

	// 5. Seed Dynamic Service Targets (Platform Targets)
	var targetCount int64
	db.Model(&model.ServiceTarget{}).Count(&targetCount)
	if targetCount == 0 {
		targets := []model.ServiceTarget{
			{ID: uuid.New(), Key: "gemini", Label: "جمینای (Gemini)", Icon: "Sparkles", SortOrder: 1, IsActive: true},
			{ID: uuid.New(), Key: "chatgpt", Label: "چت‌جی‌پی‌تی (ChatGPT)", Icon: "Bot", SortOrder: 2, IsActive: true},
			{ID: uuid.New(), Key: "claude", Label: "کلود (Claude)", Icon: "Brain", SortOrder: 3, IsActive: true},
			{ID: uuid.New(), Key: "virtual_number", Label: "شماره مجازی (5SIM)", Icon: "Radio", SortOrder: 4, IsActive: true},
			{ID: uuid.New(), Key: "midjourney", Label: "میدجورنی (Midjourney)", Icon: "Palette", SortOrder: 5, IsActive: true},
			{ID: uuid.New(), Key: "cursor", Label: "کرسر و ابزارهای کدنویسی", Icon: "Code", SortOrder: 6, IsActive: true},
			{ID: uuid.New(), Key: "general", Label: "عمومی و حساب کاربری", Icon: "LifeBuoy", SortOrder: 7, IsActive: true},
		}

		for _, tgt := range targets {
			db.Create(&tgt)
		}
		log.Println("🌱 Seeded dynamic service platform targets into DB")
	}

	// 6. Seed Sample Reader Comments
	var commentCount int64
	db.Model(&model.ArticleComment{}).Count(&commentCount)
	if commentCount == 0 {
		var firstArticle model.Article
		if err := db.Where("slug = ?", "how-to-buy-gemini-advanced-iran").First(&firstArticle).Error; err == nil {
			comments := []model.ArticleComment{
				{
					ID:         uuid.New(),
					ArticleID:  firstArticle.ID,
					AuthorName: "امیرحسین رضوی",
					Email:      "amir@example.com",
					Content:    "سلام، من اکانت ۱۸ ماهه جمینای رو از آناهیکس گرفتم و کمتر از ۱ دقیقه تحویل شد. گوگل وان ۲ ترابایت هم فوراً روی درایو فعال شد. ممنون از تیم خوبتون.",
					Status:     "approved",
					AdminReply: "درود بر شما امیرحسین عزیز؛ بسیار خوشحالیم که تجربه رضایت‌بخشی داشتید. گارانتی اشتراک شما تا آخرین روز فعال خواهد بود.",
				},
				{
					ID:         uuid.New(),
					ArticleID:  firstArticle.ID,
					AuthorName: "سارا ملکی",
					Email:      "sara.m@gmail.com",
					Content:    "آیا برای استفاده در گوشی آیفون هم نیاز به لاگین جداگانه در اپلیکیشن گوگل هست؟",
					Status:     "approved",
					AdminReply: "بله سارای گرامی، کافی است در مرورگر Safari یا اپلیکیشن Google جیمیل ارائه‌شده را وارد نمایید و از بخش Gemini استفاده کنید.",
				},
			}
			for _, c := range comments {
				db.Create(&c)
			}
			log.Println("🌱 Seeded verified comments and admin replies for blog articles")
		}
	}
}
