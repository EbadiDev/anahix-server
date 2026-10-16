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

	"github.com/EbadiDev/anahix-server/initialize"
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

	// Run ArchNet-style migrations and initial database seeders
	initialize.StartInitDatabase(db)

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

	// Ensure system owner phone is granted admin status
	db.Model(&model.Customer{}).Where("phone = ?", "09928376377").Update("is_admin", true)

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
	smsTemplateDefaults := []model.SystemSetting{
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
			Key:      "smsir_template_otp",
			Value:    "100000",
			Type:     "string",
			Desc:     "شناسه الگوی خوش‌آمدگویی و ورود OTP",
		},
		{
			Category: "sms",
			Key:      "smsir_template_2fa",
			Value:    "100001",
			Type:     "string",
			Desc:     "شناسه الگوی اعلام درخواست کد ۲FA به مشتری",
		},
		{
			Category: "sms",
			Key:      "smsir_template_order_complete",
			Value:    "100002",
			Type:     "string",
			Desc:     "شناسه الگوی تحویل لایسنس و تکمیل سفارش",
		},
		{
			Category: "sms",
			Key:      "smsir_template_renewal",
			Value:    "100003",
			Type:     "string",
			Desc:     "شناسه الگوی یادآور تمدید اشتراک",
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

	for _, s := range smsTemplateDefaults {
		var existing model.SystemSetting
		if err := db.Where("category = ? AND `key` = ?", s.Category, s.Key).First(&existing).Error; err != nil {
			db.Create(&s)
		}
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

	// Seed Coupons if empty
	var couponCount int64
	db.Model(&model.Coupon{}).Count(&couponCount)
	if couponCount == 0 {
		coupons := []model.Coupon{
			{
				ID:                  uuid.New(),
				Code:                "WELCOME10",
				DiscountType:        model.DiscountTypePercentage,
				DiscountValue:       10, // 10%
				MinOrderAmountToman: 100000,
				MaxDiscountToman:    250000,
				PerUserLimit:        1,
				MaxUsage:            1000,
				Category:            "all",
				IsActive:            true,
			},
			{
				ID:                  uuid.New(),
				Code:                "ANAHIX50",
				DiscountType:        model.DiscountTypeFixed,
				DiscountValue:       50000, // 50,000 Toman
				MinOrderAmountToman: 300000,
				PerUserLimit:        1,
				MaxUsage:            500,
				Category:            "all",
				IsActive:            true,
			},
			{
				ID:                  uuid.New(),
				Code:                "TELEGRAM20",
				DiscountType:        model.DiscountTypePercentage,
				DiscountValue:       20, // 20%
				MinOrderAmountToman: 50000,
				MaxDiscountToman:    100000,
				PerUserLimit:        1,
				MaxUsage:            2000,
				Category:            "virtual_number",
				IsActive:            true,
			},
		}
		for _, c := range coupons {
			db.Create(&c)
		}
		log.Println("🌱 Seeded default promotional coupons ('WELCOME10', 'ANAHIX50', 'TELEGRAM20')")
	}

	// Seed Official Tax Billing & Corporate Seller Configurations
	var billingCount int64
	db.Model(&model.SystemSetting{}).Where("category = ?", "billing").Count(&billingCount)
	if billingCount == 0 {
		billingSettings := []model.SystemSetting{
			{
				Category: "billing",
				Key:      "billing_seller_company_name",
				Value:    "شرکت نوآوران داده‌پرداز آناهیکس (با مسئولیت محدود)",
				Type:     "string",
				Desc:     "نام رسمی شرکت / فروشگاه جهت صدور فاکتور رسمی",
			},
			{
				Category: "billing",
				Key:      "billing_seller_brand_fa",
				Value:    "فروشگاه تخصصی هوش مصنوعی آناهیکس",
				Type:     "string",
				Desc:     "نام تجاری یا برند فارسی فروشگاه",
			},
			{
				Category: "billing",
				Key:      "billing_seller_national_id",
				Value:    "14012345678",
				Type:     "string",
				Desc:     "شناسه ملی شرکت فروشنده (۱۱ رقمی)",
			},
			{
				Category: "billing",
				Key:      "billing_seller_economic_code",
				Value:    "411567891234",
				Type:     "string",
				Desc:     "کد اقتصادی ۱۲ رقمی سازمان امور مالیاتی",
			},
			{
				Category: "billing",
				Key:      "billing_seller_registration_no",
				Value:    "589234",
				Type:     "string",
				Desc:     "شماره ثبت شرکت در اداره ثبت شرکت‌ها",
			},
			{
				Category: "billing",
				Key:      "billing_seller_postal_code",
				Value:    "1997834561",
				Type:     "string",
				Desc:     "کد پستی ۱۰ رقمی اقامتگاه قانونی فروشنده",
			},
			{
				Category: "billing",
				Key:      "billing_seller_province",
				Value:    "تهران",
				Type:     "string",
				Desc:     "استان محل اقامتگاه قانونی",
			},
			{
				Category: "billing",
				Key:      "billing_seller_city",
				Value:    "تهران",
				Type:     "string",
				Desc:     "شهر محل اقامتگاه قانونی",
			},
			{
				Category: "billing",
				Key:      "billing_seller_address",
				Value:    "تهران، خیابان ولیعصر، تقاطع میرداماد، برج فناوری اطلاعات، طبقه ۵، واحد ۱۵",
				Type:     "string",
				Desc:     "نشانی دقیق پستی دفتر مرکزی فروشنده",
			},
			{
				Category: "billing",
				Key:      "billing_seller_phone",
				Value:    "021-88776655",
				Type:     "string",
				Desc:     "تلفن تماس ثابت و پشتیبانی صدور فاکتور",
			},
			{
				Category: "billing",
				Key:      "billing_seller_email",
				Value:    "billing@anahix.com",
				Type:     "string",
				Desc:     "ایمیل رسمی واحد مالی و صدور فاکتور",
			},
			{
				Category: "billing",
				Key:      "billing_seller_website",
				Value:    "https://anahix.com",
				Type:     "string",
				Desc:     "آدرس وب‌سایت رسمی فروشنده",
			},
			{
				Category: "billing",
				Key:      "billing_vat_percent",
				Value:    "10",
				Type:     "number",
				Desc:     "درصد قانونی مالیات بر ارزش افزوده (پیش‌فرض ۱۰ درصد)",
			},
			{
				Category: "billing",
				Key:      "billing_vat_enabled",
				Value:    "true",
				Type:     "bool",
				Desc:     "فعال بودن محاسبه مالیات بر ارزش افزوده در فاکتور رسمی",
			},
			{
				Category: "billing",
				Key:      "billing_invoice_notes",
				Value:    "این صورتحساب مطابق با ماده ۱۹ قانون مالیات بر ارزش افزوده تنظیم گردیده و مشمول خرید رسمی شرکتی و سازمانی می‌باشد.",
				Type:     "string",
				Desc:     "متن توضیحات و ضوابط حقوقی ذیل فاکتور رسمی",
			},
			{
				Category: "billing",
				Key:      "billing_tax_memory_id",
				Value:    "ANX-TAX-9824",
				Type:     "string",
				Desc:     "شناسه یکتای حافظه مالیاتی یا شماره مجوز رسمی",
			},
			{
				Category: "billing",
				Key:      "billing_stamp_signature_url",
				Value:    "",
				Type:     "string",
				Desc:     "آدرس تصویر مهر و امضای دیجیتال رسمی فروشگاه",
			},
		}

		for _, s := range billingSettings {
			db.Create(&s)
		}
		log.Println("🌱 Seeded official tax billing configurations (category: 'billing')")
	}

	// Seed PDP rich configurations (SEO, Plans, Comparison Matrix, Guides, FAQs)
	seedPDPData(db)

	// Seed Verified Buyer Reviews & Social Proof
	seedReviewsData(db)
	seedTicketsData(db)

	// Seed Blog Articles, Knowledge Base Guides, and Content Taxonomies
	seedContentData(db)

	// Seed Loyalty Tiers and VIP Club Configurations
	seedLoyaltyData(db)

	// Seed Promotional Banners, Holiday Theme Settings, and Flash Deals
	seedBannersAndThemeData(db)
}

