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
		log.Printf("⚠️ Postgres connection failed (%v). Falling back to local SQLite 'anahix.db' for development...", err)
		testDB, testErr := orm.InitTestDB()
		if testErr != nil {
			log.Fatalf("Fatal: Database initialization failed: %v", testErr)
		}
		db = testDB
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
}
