package cmd

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/EbadiDev/anahix-server/internal/config"
	"github.com/EbadiDev/anahix-server/internal/handler"
	"github.com/EbadiDev/anahix-server/internal/model"
	"github.com/EbadiDev/anahix-server/internal/svc"
	"github.com/EbadiDev/anahix-server/pkg/orm"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"gorm.io/gorm"
)

var configPath string

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Start the Anahix API server",
	Run: func(cmd *cobra.Command, args []string) {
		runServer()
	},
}

func init() {
	runCmd.Flags().StringVarP(&configPath, "config", "c", "etc/config.yaml", "Path to config file")
}

func runServer() {
	fmt.Println("🚀 Starting Anahix Server (آناهیکس)...")

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load config from %s: %v", configPath, err)
	}

	// Initialize DB (Postgres connection with fallback to SQLite for local development if Postgres not reachable)
	db, err := orm.InitPostgres(cfg.Postgres, cfg.Server.Debug)
	if err != nil {
		log.Printf("⚠️ Postgres connection failed (%v). Falling back to persistent local SQLite 'anahix.db' for development...", err)
		localDB, localErr := orm.InitLocalSQLite("anahix.db")
		if localErr != nil {
			log.Fatalf("Fatal: Database initialization failed: %v", localErr)
		}
		db = localDB
	} else {
		// Run auto migrations on Postgres
		if err := orm.AutoMigrate(db); err != nil {
			log.Fatalf("Database migration failed: %v", err)
		}
	}

	// Seed default admin and sample products if table is empty
	seedInitialData(db)

	svcCtx, err := svc.NewServiceContext(cfg, db)
	if err != nil {
		log.Fatalf("Failed to initialize ServiceContext: %v", err)
	}

	if !cfg.Server.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()
	handler.RegisterRoutes(router, svcCtx)

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	go func() {
		log.Printf("✅ Anahix API server listening on http://%s\n", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down Anahix server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}
	log.Println("👋 Anahix server exited cleanly.")
}

func seedInitialData(db *gorm.DB) {
	// Seed Admin
	var adminCount int64
	db.Model(&model.AdminUser{}).Count(&adminCount)
	if adminCount == 0 {
		admin := model.AdminUser{
			ID:       uuid.New(),
			Username: "admin",
			FullName: "مدیر کل آناهیکس",
			Role:     model.RoleSuperAdmin,
			IsActive: true,
		}
		_ = admin.SetPassword("admin123456")
		db.Create(&admin)
		log.Println("🌱 Created initial admin user: 'admin' (password: 'admin123456')")
	}

	// Seed Initial AI Products if empty
	var prodCount int64
	db.Model(&model.Product{}).Count(&prodCount)
	if prodCount == 0 {
		products := []model.Product{
			{
				ID:             uuid.New(),
				Title:          "Gemini Advanced 18 Months Activation Link",
				TitleFA:        "لینک فعال‌سازی ۱۸ ماهه گوگل جمینای ادونس (Gemini)",
				Slug:           "gemini-advanced-18m",
				Category:       model.CategoryAILink,
				DeliveryType:   model.DeliveryInstant,
				PriceToman:     1480000,
				BaseCostUSD:    12.0,
				IsActive:       true,
				OperatingHours: "۲۴ ساعته (تحویل آنی)",
				Description:    "فعال‌سازی قانونی اکانت Google Gemini Advanced و Google One به مدت ۱۸ ماه با امکان فعال‌سازی روی جیمیل شخصی خودتان بدون نیاز به پسورد.",
				Instructions:   "۱. روی لینک تحویل داده شده کلیک کنید.\n۲. با اکانت گوگل خود لاگین کنید.\n۳. بر روی دکمه Redeem کلیک کنید تا ۱۸ ماه فعال شود.",
			},
			{
				ID:             uuid.New(),
				Title:          "ChatGPT Plus Personal Account Upgrade",
				TitleFA:        "خرید و فعال‌سازی چت جی‌پی‌تی پلاس روی اکانت شخصی (ChatGPT Plus)",
				Slug:           "chatgpt-plus-personal",
				Category:       model.CategoryPersonalUpgrade,
				DeliveryType:   model.DeliveryManualTimed,
				SLAMinMinutes:  30,
				SLAMaxMinutes:  180,
				SLADisplayText: "تحویل بین ۳۰ دقیقه تا ۳ ساعت کاری",
				PriceToman:     1950000,
				BaseCostUSD:    20.0,
				IsActive:       true,
				OperatingHours: "۰۹:۰۰ الی ۲۳:۰۰",
				Description:    "پرداخت مستقیم اشتراک ChatGPT Plus با کارت‌های بین‌المللی معتبر روی ایمیل شخصی شما با حفظ تمام چت‌ها و هیستوری قبلی.",
				Instructions:   "اطلاعات اکانت خود را در مرحله خرید ثبت کنید. در صورت نیاز به کد تایید دومرحله‌ای، اپراتور از طریق همین صفحه از شما درخواست کد خواهد کرد.",
			},
			{
				ID:             uuid.New(),
				Title:          "Claude Pro Ready Account",
				TitleFA:        "اکانت آماده و اختصاصی کلود پرو (Claude Pro)",
				Slug:           "claude-pro-ready",
				Category:       model.CategoryAIAccount,
				DeliveryType:   model.DeliveryInstant,
				PriceToman:     1890000,
				BaseCostUSD:    20.0,
				IsActive:       true,
				OperatingHours: "۲۴ ساعته (تحویل آنی)",
				Description:    "اکانت کاملاً اختصاصی Claude Pro ساخته شده روی ایمیل معتبر با دسترسی به Claude 3.5 Sonnet و سقف پیام بالا.",
				Instructions:   "ایمیل و رمز عبور دریافتی را در سایت claude.ai وارد کرده و استفاده نمایید. رمز عبور قابل تغییر است.",
			},
		}

		for _, p := range products {
			db.Create(&p)
		}
		log.Println("🌱 Seeded initial AI products: Gemini 18m (Instant), ChatGPT Plus (Timed), Claude Pro (Instant)")
	}

	// Ensure group tags and titles are populated for seeded products
	db.Model(&model.Product{}).Where("slug LIKE ? AND (group_tag = '' OR group_tag = 'general')", "%gemini%").Updates(map[string]interface{}{
		"group_tag":             "gemini",
		"group_title_fa":        "سرویس‌های گوگل جمینای (Google Gemini)",
		"delivery_strategy":     "third_party_instant",
		"external_provider":     "insightx",
		"external_product_id":   14,
		"external_product_name": "Gemini Pro 18 Months",
	})
	db.Model(&model.Product{}).Where("slug LIKE ? AND (group_tag = '' OR group_tag = 'general')", "%chatgpt%").Updates(map[string]interface{}{
		"group_tag":         "chatgpt",
		"group_title_fa":    "محصولات اوپن‌ای‌آی و چت‌جی‌پی‌تی (OpenAI ChatGPT)",
		"delivery_strategy": "manual_operator",
	})
	db.Model(&model.Product{}).Where("slug LIKE ? AND (group_tag = '' OR group_tag = 'general')", "%claude%").Updates(map[string]interface{}{
		"group_tag":         "claude",
		"group_title_fa":    "سرویس‌های انتروپیک کلود (Anthropic Claude)",
		"delivery_strategy": "manual_operator",
	})

	// Seed Payment Methods (ArchNet pattern: Balance as default ID = -1)
	var paymentCount int64
	db.Model(&model.Payment{}).Count(&paymentCount)
	if paymentCount == 0 {
		payments := []model.Payment{
			{
				Id:          -1,
				Name:        "Balance",
				NameFA:      "کیف پول / موجودی حساب",
				Platform:    model.PlatformBalance,
				Icon:        "wallet",
				Description: "پرداخت مستقیم و آنی از طریق شارژ کیف پول حساب کاربری",
				Config:      "{}",
				Enable:      true,
				Sort:        0,
			},
			{
				Id:          1,
				Name:        "ZarinPal",
				NameFA:      "درگاه پرداخت زرین‌پال (کارت‌های شتاب)",
				Platform:    model.PlatformZarinpal,
				Icon:        "zarinpal",
				Description: "پرداخت امن اینترنتی با تمامی کارت‌های بانکی عضو شبکه شتاب",
				Config:      `{"merchant_id":"00000000-0000-0000-0000-000000000000","sandbox":true}`,
				Enable:      true,
				Sort:        1,
			},
			{
				Id:          2,
				Name:        "CardToCard",
				NameFA:      "کارت به کارت (واریز مستقیم بانکی)",
				Platform:    model.PlatformCardToCard,
				Icon:        "credit-card",
				Description: "انتقال وجه کارت به کارت و ارسال رسید و شماره پیگیری برای تایید اپراتور",
				Config:      `{"card_number":"6037-9918-0000-0000","card_holder":"آناهیکس","bank_name":"بانک ملی"}`,
				Enable:      true,
				Sort:        2,
			},
			{
				Id:          3,
				Name:        "Cryptocurrency",
				NameFA:      "ارز دیجیتال (تتر USDT / ترون TRX)",
				Platform:    model.PlatformCrypto,
				Icon:        "bitcoin",
				Description: "پرداخت ناشناس و بین‌المللی با تتر شبکه TRC20 یا تون",
				Config:      `{"network":"TRC20","wallet_address":""}`,
				Enable:      false,
				Sort:        3,
			},
		}

		for _, pay := range payments {
			db.Create(&pay)
		}
		log.Println("🌱 Seeded payment methods with Balance (id: -1) as default, ZarinPal, CardToCard, and Crypto")
	}

	// Seed System Settings (SMS Gateway & Notification configurations in DB)
	var sysCount int64
	db.Model(&model.SystemSetting{}).Where("category = ?", "sms").Count(&sysCount)
	if sysCount == 0 {
		smsSettings := []model.SystemSetting{
			{
				Category: "sms",
				Key:      "sms_provider",
				Value:    "smsir",
				Type:     "string",
				Desc:     "سرویس‌دهنده پیامک اصلی (smsir, mock)",
			},
			{
				Category: "sms",
				Key:      "smsir_api_key",
				Value:    "",
				Type:     "string",
				Desc:     "کلید وب‌سرویس اختصاصی سامانه پیامکی SMS.ir",
			},
			{
				Category: "sms",
				Key:      "smsir_line_number",
				Value:    "30004505000017",
				Type:     "string",
				Desc:     "شماره خط فرستنده اختصاصی پیامک",
			},
			{
				Category: "sms",
				Key:      "smsir_template_id",
				Value:    "100000",
				Type:     "string",
				Desc:     "شناسه قالب پیامک تایید OTP در سامانه پیامک",
			},
			{
				Category: "sms",
				Key:      "smsir_is_sandbox",
				Value:    "true",
				Type:     "bool",
				Desc:     "حالت آزمایشی سامانه پیامک (بدون کسر اعتبار)",
			},
			{
				Category: "sms",
				Key:      "sms_enabled",
				Value:    "true",
				Type:     "bool",
				Desc:     "فعال بودن ارسال پیامک و کد تایید احراز هویت",
			},
		}

		for _, s := range smsSettings {
			db.Create(&s)
		}
		log.Println("🌱 Seeded SMS.ir gateway configurations into database (category: 'sms')")
	}

	// Seed Virtual Numbers Settings in DB
	var vnCount int64
	db.Model(&model.SystemSetting{}).Where("category IN (?, ?)", "virtual_numbers", "fivesim").Count(&vnCount)
	if vnCount == 0 {
		vnSettings := []model.SystemSetting{
			{
				Category: "virtual_numbers",
				Key:      "vn_enabled",
				Value:    "true",
				Type:     "bool",
				Desc:     "فعال بودن فروش شماره مجازی دریافت پیامک",
			},
			{
				Category: "virtual_numbers",
				Key:      "vn_api_key",
				Value:    "",
				Type:     "string",
				Desc:     "کلید دسترسی زیرساخت شماره‌های مجازی آناهیکس",
			},
			{
				Category: "virtual_numbers",
				Key:      "vn_margin_percent",
				Value:    "25",
				Type:     "number",
				Desc:     "درصد سود فروش روی هر شماره مجازی (پیش‌فرض ۲۵ درصد)",
			},
			{
				Category: "virtual_numbers",
				Key:      "vn_usd_rate",
				Value:    "95000",
				Type:     "number",
				Desc:     "نرخ تبدیل دلار به تومان برای محاسبه قیمت شماره‌ها",
			},
			{
				Category: "virtual_numbers",
				Key:      "vn_rub_rate",
				Value:    "1050",
				Type:     "number",
				Desc:     "نرخ تبدیل روبل به تومان",
			},
			{
				Category: "virtual_numbers",
				Key:      "vn_default_operator",
				Value:    "any",
				Type:     "string",
				Desc:     "اپراتور پیش‌فرض خرید شماره (any برای بهترین درصد موفقیت)",
			},
			{
				Category: "virtual_numbers",
				Key:      "vn_rate_priority",
				Value:    "manual",
				Type:     "string",
				Desc:     "اولویت محاسبه نرخ ارز",
			},
		}

		for _, s := range vnSettings {
			db.Create(&s)
		}
		log.Println("🌱 Seeded dynamic virtual number provider configurations (category: 'virtual_numbers')")
	}

	// Seed Instant Delivery Provider Settings in DB
	var instantCount int64
	db.Model(&model.SystemSetting{}).Where("category IN (?, ?)", "instant_delivery", "insightx").Count(&instantCount)
	if instantCount == 0 {
		instantSettings := []model.SystemSetting{
			{
				Category: "instant_delivery",
				Key:      "provider_enabled",
				Value:    "true",
				Type:     "bool",
				Desc:     "فعال بودن تحویل آنی و خودکار محصولات دیجیتال آناهیکس",
			},
			{
				Category: "instant_delivery",
				Key:      "provider_api_key",
				Value:    "isk_live_jexXb9akFkevDfAGpybP9AM4eI0fNqhE",
				Type:     "string",
				Desc:     "کلید اتصال به موتور تحویل آنی و لایسنس‌های آناهیکس",
			},
			{
				Category: "instant_delivery",
				Key:      "provider_margin_percent",
				Value:    "25",
				Type:     "number",
				Desc:     "درصد مارجین سود محاسبه قیمت تومانی لایسنس‌ها",
			},
		}

		for _, s := range instantSettings {
			db.Create(&s)
		}
		log.Println("🌱 Seeded dynamic instant delivery provider configurations (category: 'instant_delivery')")
	}

	// Update existing sample products with Group Tags and provider mappings
	db.Model(&model.Product{}).Where("slug LIKE ?", "%gemini%").Updates(map[string]interface{}{
		"group_tag":             "gemini",
		"group_title_fa":        "گوگل جمینای (Google Gemini)",
		"delivery_strategy":     "third_party_instant",
		"external_provider":     "third_party_api",
		"external_product_id":   14,
		"external_product_name": "Gemini Pro 18 Months",
	})
	db.Model(&model.Product{}).Where("slug LIKE ?", "%chatgpt%").Updates(map[string]interface{}{
		"group_tag":         "chatgpt",
		"group_title_fa":    "اوپن‌ای‌آی و چت‌جی‌پی‌تی (ChatGPT)",
		"delivery_strategy": "manual_operator",
		"external_provider": "third_party_api",
	})
	db.Model(&model.Product{}).Where("slug LIKE ?", "%claude%").Updates(map[string]interface{}{
		"group_tag":         "claude",
		"group_title_fa":    "سرویس‌های انتروپیک کلود (Anthropic Claude)",
		"delivery_strategy": "manual_operator",
		"external_provider": "third_party_api",
	})
}

