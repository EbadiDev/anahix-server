package handler

import (
	"github.com/EbadiDev/anahix-server/internal/handler/admin"
	"github.com/EbadiDev/anahix-server/internal/handler/public"
	"github.com/EbadiDev/anahix-server/internal/middleware"
	"github.com/EbadiDev/anahix-server/internal/svc"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, svcCtx *svc.ServiceContext) {
	// Global Middlewares
	r.Use(middleware.CorsMiddleware())
	r.Use(gin.Recovery())

	pubHandler := public.NewPublicHandler(svcCtx)
	custAuthHandler := public.NewCustomerAuthHandler(svcCtx)
	pubTicketHandler := public.NewTicketHandler(svcCtx)
	vnumHandler := public.NewVirtualNumberHandler(svcCtx)
	pubContentHandler := public.NewContentHandler(svcCtx)
	admHandler := admin.NewAdminHandler(svcCtx)
	admContentHandler := admin.NewContentAdminHandler(svcCtx)
	admBIHandler := admin.NewBIAnalyticsHandler(svcCtx)
	admLoyaltyHandler := admin.NewLoyaltyAdminHandler(svcCtx)
	pubLoyaltyHandler := public.NewLoyaltyHandler(svcCtx)
	pubBannerHandler := public.NewBannerHandler(svcCtx)
	pubThemeHandler := public.NewThemeHandler(svcCtx)
	admBannerHandler := admin.NewBannerAdminHandler(svcCtx)
	admThemeHandler := admin.NewThemeAdminHandler(svcCtx)
	admUploadHandler := admin.NewUploadHandler(svcCtx)
	pubNotifHandler := public.NewNotificationHandler(svcCtx)

	// Static assets serving (Uploaded ticket screenshots and attachments)
	r.Static("/uploads", "./uploads")

	api := r.Group("/api/v1")
	{
		// Customer Authentication (SMS.ir OTP & Multi-Method)
		authGroup := api.Group("/auth")
		{
			authGroup.POST("/otp/send", custAuthHandler.SendOTP)
			authGroup.POST("/otp/verify", custAuthHandler.VerifyOTP)
			authGroup.POST("/password/login", custAuthHandler.LoginWithPassword)
			authGroup.POST("/register", custAuthHandler.RegisterWithPassword)
			authGroup.POST("/social/simulate", custAuthHandler.OAuthSimulate)
			authGroup.POST("/password/set", middleware.CustomerAuthMiddleware(svcCtx.Config.JWT.AccessSecret), custAuthHandler.SetPassword)
			authGroup.GET("/me", middleware.CustomerAuthMiddleware(svcCtx.Config.JWT.AccessSecret), custAuthHandler.GetMe)
			authGroup.PUT("/me", middleware.CustomerAuthMiddleware(svcCtx.Config.JWT.AccessSecret), custAuthHandler.UpdateMe)
			authGroup.GET("/my-reviews", middleware.CustomerAuthMiddleware(svcCtx.Config.JWT.AccessSecret), pubHandler.GetMyReviews)
		}

		// Promotional Banners & Flash Deals
		banners := api.Group("/banners")
		{
			banners.GET("", pubBannerHandler.ListBanners)
		}

		// Active Event Theme & Dynamic Customization
		system := api.Group("/system")
		{
			system.GET("/theme", pubThemeHandler.GetActiveTheme)
		}

		// Public Storefront Endpoints
		products := api.Group("/products")
		{
			products.GET("", pubHandler.ListProducts)
			products.GET("/:slug", pubHandler.GetProductDetail)
			products.GET("/:slug/reviews", pubHandler.GetProductReviews)
			products.GET("/:slug/review-eligibility", middleware.CustomerOptionalAuthMiddleware(svcCtx.Config.JWT.AccessSecret), pubHandler.CheckReviewEligibility)
		}

		// Public Hierarchical Store Categories
		api.GET("/store/categories", pubHandler.ListStoreCategories)

		// Social Proof & Verified Buyer Reviews
		reviews := api.Group("/reviews")
		{
			reviews.POST("", middleware.CustomerOptionalAuthMiddleware(svcCtx.Config.JWT.AccessSecret), pubHandler.SubmitReview)
			reviews.POST("/:id/helpful", pubHandler.VoteReviewHelpful)
		}

		orders := api.Group("/orders")
		{
			orders.POST("", pubHandler.CreateOrder)
			orders.GET("/:order_id", pubHandler.GetOrderDetail)
			orders.POST("/:order_id/2fa", pubHandler.SubmitTwoFactor)
			orders.GET("/:order_id/invoice", middleware.CustomerOptionalAuthMiddleware(svcCtx.Config.JWT.AccessSecret), pubHandler.GetOrderInvoice)
			orders.PUT("/:order_id/invoice-buyer", middleware.CustomerOptionalAuthMiddleware(svcCtx.Config.JWT.AccessSecret), pubHandler.UpdateOrderBuyerInvoice)
		}

		payments := api.Group("/payments")
		{
		 payments.GET("", pubHandler.ListPaymentMethods)
		}

		// System Announcements & Stock Alerts
		api.GET("/announcements", pubHandler.ListAnnouncements)
		api.POST("/stock-alerts", pubHandler.CreateStockAlert)

		// Coupons
		coupons := api.Group("/coupons")
		{
			coupons.GET("", pubHandler.ListPublicCoupons)
			coupons.POST("/validate", pubHandler.ValidateCoupon)
		}

		// VIP Club & Loyalty Tiers
		loyalty := api.Group("/loyalty")
		{
			loyalty.GET("/profile", middleware.CustomerOptionalAuthMiddleware(svcCtx.Config.JWT.AccessSecret), pubLoyaltyHandler.GetMyLoyaltyProfile)
			loyalty.GET("/tiers", pubLoyaltyHandler.ListTiersPublic)
		}

		// Referrals & Affiliate
		referrals := api.Group("/referrals")
		{
			referrals.GET("/stats", pubHandler.GetReferralStats)
			referrals.POST("/transfer-to-wallet", middleware.CustomerAuthMiddleware(svcCtx.Config.JWT.AccessSecret), pubHandler.TransferReferralCommission)
		}

		// Customer Wallet Ledger
		wallet := api.Group("/wallet")
		{
			wallet.GET("/ledger", pubHandler.GetWalletLedger)
		}

		// Customer Support & Real-time Ticketing (with attachments & SMS notifications)
		tickets := api.Group("/tickets")
		tickets.Use(middleware.CustomerAuthMiddleware(svcCtx.Config.JWT.AccessSecret))
		{
			tickets.POST("", pubTicketHandler.CreateTicket)
			tickets.GET("", pubTicketHandler.ListMyTickets)
			tickets.GET("/:id", pubTicketHandler.GetTicketDetail)
			tickets.POST("/:id/reply", pubTicketHandler.ReplyTicket)
			tickets.POST("/:id/close", pubTicketHandler.CloseTicket)
			tickets.POST("/:id/rate", pubTicketHandler.RateTicket)
			tickets.PATCH("/:id/notifications", pubTicketHandler.UpdateNotificationSettings)
			tickets.PUT("/:id/notifications", pubTicketHandler.UpdateNotificationSettings)
			tickets.POST("/upload", pubTicketHandler.UploadAttachment)
		}

		// In-App Customer Notifications
		notifs := api.Group("/notifications")
		notifs.Use(middleware.CustomerAuthMiddleware(svcCtx.Config.JWT.AccessSecret))
		{
			notifs.GET("", pubNotifHandler.ListNotifications)
			notifs.POST("/:id/read", pubNotifHandler.MarkAsRead)
			notifs.POST("/read-all", pubNotifHandler.MarkAllAsRead)
		}

		// Customer Favorites & Wishlist
		favorites := api.Group("/favorites")
		favorites.Use(middleware.CustomerAuthMiddleware(svcCtx.Config.JWT.AccessSecret))
		{
			favorites.POST("/toggle", pubHandler.ToggleFavorite)
			favorites.GET("", pubHandler.ListFavorites)
		}

		// Content Taxonomies (Categories, Templates & Service Targets)
		content := api.Group("/content")
		{
			content.GET("/categories", pubContentHandler.ListCategories)
			content.GET("/templates", pubContentHandler.ListTemplates)
			content.GET("/services", pubContentHandler.ListServices)
		}

		// Blog Articles (SEO & Tutorials)
		articles := api.Group("/articles")
		{
			articles.GET("", pubContentHandler.ListArticles)
			articles.GET("/:slug", pubContentHandler.GetArticleDetail)
			articles.POST("/:slug/like", middleware.CustomerOptionalAuthMiddleware(svcCtx.Config.JWT.AccessSecret), pubContentHandler.LikeArticle)
			articles.POST("/:slug/comments", middleware.CustomerOptionalAuthMiddleware(svcCtx.Config.JWT.AccessSecret), pubContentHandler.SubmitComment)
		}

		// Knowledge Base / Help Center (Guides, Troubleshooting & SLAs)
		help := api.Group("/help")
		{
			help.GET("/guides", pubContentHandler.ListHelpGuides)
			help.GET("/guides/:slug", pubContentHandler.GetHelpGuideDetail)
			help.POST("/guides/:slug/vote", middleware.CustomerOptionalAuthMiddleware(svcCtx.Config.JWT.AccessSecret), pubContentHandler.VoteHelpGuide)
		}

		// Virtual Numbers (5SIM integration)
		vnums := api.Group("/virtual-numbers")
		{
			vnums.GET("/services", vnumHandler.ListServices)
			vnums.GET("/countries", vnumHandler.ListCountriesForService)
			vnums.GET("/catalog", vnumHandler.ListCatalog)
			vnums.POST("/orders", vnumHandler.CreateOrder)
			vnums.GET("/orders/:order_id", vnumHandler.GetOrderDetail)
			vnums.POST("/orders/:order_id/cancel", vnumHandler.CancelOrder)
			vnums.POST("/orders/:order_id/finish", vnumHandler.FinishOrder)
		}

		// Admin Endpoints
		adminGroup := api.Group("/admin")
		{
			// Auth
			adminGroup.POST("/auth/login", admHandler.Login)
			adminGroup.POST("/auth/exchange", middleware.CustomerAuthMiddleware(svcCtx.Config.JWT.AccessSecret), admHandler.ExchangeCustomerToken)

			// Protected Admin Routes
			protected := adminGroup.Group("")
			protected.Use(middleware.AdminAuthMiddleware(svcCtx.Config.JWT.AccessSecret, svcCtx.DB))
			{
				// Products & Inventory
				protected.GET("/products", admHandler.ListProducts)
				protected.GET("/products/:id", admHandler.GetProduct)
				protected.POST("/products", admHandler.CreateProduct)
				protected.PUT("/products/:id", admHandler.UpdateProduct)
				protected.DELETE("/products/:id", admHandler.DeleteProduct)
				protected.POST("/inventory/bulk-import", admHandler.BulkImportInventory)
				protected.GET("/inventory", admHandler.ListInventory)
				protected.GET("/inventory/stats", admHandler.GetInventoryStats)
				protected.POST("/inventory/replace-key", admHandler.ReplaceKey)
				protected.GET("/inventory/export", admHandler.ExportKeys)

				// Coupons Management
				protected.GET("/coupons", admHandler.ListCoupons)
				protected.POST("/coupons", admHandler.CreateCoupon)
				protected.PUT("/coupons/:id", admHandler.UpdateCoupon)
				protected.PUT("/coupons/:id/toggle", admHandler.ToggleCoupon)
				protected.DELETE("/coupons/:id", admHandler.DeleteCoupon)

				// Buyer Reviews & Social Proof Moderation
				protected.GET("/reviews", admHandler.ListReviews)
				protected.PUT("/reviews/:id/status", admHandler.UpdateReviewStatus)
				protected.POST("/reviews/:id/reply", admHandler.ReplyReview)
				protected.PUT("/reviews/:id/feature", admHandler.FeatureReview)
				protected.DELETE("/reviews/:id", admHandler.DeleteReview)

				// Support & Real Ticketing Engine Management
				protected.GET("/tickets", admHandler.ListTickets)
				protected.GET("/tickets/:id", admHandler.GetTicketDetail)
				protected.POST("/tickets/:id/reply", admHandler.ReplyTicket)
				protected.PUT("/tickets/:id", admHandler.UpdateTicket)
				protected.DELETE("/tickets/:id", admHandler.DeleteTicket)
				protected.POST("/tickets/upload", admHandler.UploadTicketAttachment)

				// 360-Degree Customer Management
				protected.GET("/customers", admHandler.ListCustomers)
				protected.GET("/customers/:id", admHandler.GetCustomer360)
				protected.POST("/customers/:id/wallet-adjust", admHandler.AdjustCustomerWallet)
				protected.PUT("/customers/:id", admHandler.UpdateCustomer)
				protected.POST("/customers/:id/sms", admHandler.SendCustomerSMS)
				protected.POST("/customers/:id/replace-key", admHandler.ReplaceCustomerOrderKey)

				// Operator Order Queue Workbench & Enhanced Orders Management
				protected.GET("/orders", admHandler.ListQueueOrders)
				protected.GET("/orders/all", admHandler.ListAllOrders)
				protected.PUT("/orders/:order_id/status", admHandler.UpdateOrderStatus)
				protected.GET("/orders/:order_id", admHandler.GetOrderDetail)
				protected.POST("/orders/:order_id/2fa-request", admHandler.RequestTwoFactor)
				protected.POST("/orders/:order_id/complete", admHandler.CompleteOrder)
				protected.POST("/orders/:order_id/retry-bot", admHandler.RetryOrderBotFulfillment)
				protected.GET("/orders/:order_id/invoice", admHandler.GetOrderInvoice)

				// System Announcements Management
				protected.GET("/announcements", admHandler.ListAnnouncements)
				protected.POST("/announcements", admHandler.CreateAnnouncement)
				protected.PUT("/announcements/:id", admHandler.UpdateAnnouncement)
				protected.DELETE("/announcements/:id", admHandler.DeleteAnnouncement)

				// Notification Metrics & BI
				protected.GET("/notifications/stats", pubNotifHandler.GetAdminNotificationStats)

				// Official Corporate Billing & Tax Settings
				protected.GET("/settings/billing", admHandler.GetBillingSettings)
				protected.POST("/settings/billing", admHandler.UpdateBillingSettings)

				// Payment Gateways Management
				protected.GET("/payments", admHandler.ListPaymentGateways)
				protected.PUT("/payments/:id", admHandler.UpdatePaymentGateway)

				// SMS Gateway (SMS.ir) Settings & Testing
				protected.GET("/settings/sms", admHandler.GetSMSSettings)
				protected.POST("/settings/sms", admHandler.UpdateSMSSettings)
				protected.POST("/settings/sms/test", admHandler.SendTestSMS)

				// Virtual Numbers Settings & Management
				protected.GET("/settings/virtual-numbers", admHandler.GetVirtualNumberSettings)
				protected.POST("/settings/virtual-numbers", admHandler.UpdateVirtualNumberSettings)
				protected.GET("/virtual-numbers/orders", admHandler.ListVirtualOrders)

				// Delivery Engine & Instant Fulfillment Provider Management
				protected.GET("/settings/instant-delivery", admHandler.GetInstantDeliverySettings)
				protected.POST("/settings/instant-delivery", admHandler.UpdateInstantDeliverySettings)
				protected.GET("/delivery/provider-catalog", admHandler.ListProviderCatalog)
				protected.GET("/delivery/groups", admHandler.ListProductGroups)
				protected.POST("/delivery/groups/strategy", admHandler.UpdateGroupDeliveryStrategy)
				protected.POST("/delivery/products/link", admHandler.LinkProductProvider)

				// Multi-Bot Providers & Fulfillment Hub
				protected.GET("/providers/settings", admHandler.ListBotProviderSettings)
				protected.POST("/providers/settings", admHandler.UpdateBotProviderSettings)
				protected.POST("/providers/test", admHandler.TestBotProviderConnection)
				protected.GET("/providers/catalog", admHandler.GetBotCatalog)
				protected.POST("/providers/import", admHandler.ImportBotProduct)

				// Hierarchical Store Categories Management
				protected.GET("/store/categories", admHandler.ListStoreCategories)
				protected.POST("/store/categories", admHandler.CreateStoreCategory)
				protected.PUT("/store/categories/:id", admHandler.UpdateStoreCategory)
				protected.DELETE("/store/categories/:id", admHandler.DeleteStoreCategory)

				// Audit & Transaction Logs
				protected.GET("/logs/transactions", admHandler.ListTransactions)
				protected.PUT("/logs/transactions/:id/status", admHandler.UpdateTransactionStatus)
				protected.GET("/logs/actions", admHandler.ListAdminActions)

				// Blog Articles & Knowledge Base Management
				protected.GET("/articles", admContentHandler.ListArticles)
				protected.GET("/articles/:id", admContentHandler.GetArticle)
				protected.POST("/articles", admContentHandler.CreateArticle)
				protected.PUT("/articles/:id", admContentHandler.UpdateArticle)
				protected.DELETE("/articles/:id", admContentHandler.DeleteArticle)

				protected.GET("/help/guides", admContentHandler.ListHelpGuides)
				protected.GET("/help/guides/:id", admContentHandler.GetHelpGuide)
				protected.POST("/help/guides", admContentHandler.CreateHelpGuide)
				protected.PUT("/help/guides/:id", admContentHandler.UpdateHelpGuide)
				protected.DELETE("/help/guides/:id", admContentHandler.DeleteHelpGuide)

				protected.GET("/content/categories", admContentHandler.ListCategories)
				protected.POST("/content/categories", admContentHandler.CreateCategory)
				protected.PUT("/content/categories/:id", admContentHandler.UpdateCategory)
				protected.DELETE("/content/categories/:id", admContentHandler.DeleteCategory)

				protected.GET("/content/comments", admContentHandler.ListComments)
				protected.PUT("/content/comments/:id", admContentHandler.UpdateCommentStatus)
				protected.DELETE("/content/comments/:id", admContentHandler.DeleteComment)

				protected.GET("/content/templates", admContentHandler.ListTemplates)
				protected.POST("/content/templates", admContentHandler.CreateTemplate)
				protected.PUT("/content/templates/:id", admContentHandler.UpdateTemplate)
				protected.DELETE("/content/templates/:id", admContentHandler.DeleteTemplate)

				protected.GET("/content/services", admContentHandler.ListServices)
				protected.POST("/content/services", admContentHandler.CreateServiceTarget)
				protected.PUT("/content/services/:id", admContentHandler.UpdateServiceTarget)
				protected.DELETE("/content/services/:id", admContentHandler.DeleteServiceTarget)

				// Business Intelligence & Commercial Analytics
				protected.GET("/bi/overview", admBIHandler.GetOverview)

				// VIP Club & Loyalty Tiers Management
				protected.GET("/loyalty/tiers", admLoyaltyHandler.ListTiers)
				protected.POST("/loyalty/tiers", admLoyaltyHandler.CreateTier)
				protected.PUT("/loyalty/tiers/:id", admLoyaltyHandler.UpdateTier)
				protected.DELETE("/loyalty/tiers/:id", admLoyaltyHandler.DeleteTier)
				protected.GET("/loyalty/settings", admLoyaltyHandler.GetSettings)
				protected.POST("/loyalty/settings", admLoyaltyHandler.UpdateSettings)

				// Banners & Promotional Campaigns Management
				protected.GET("/banners", admBannerHandler.ListBanners)
				protected.POST("/banners", admBannerHandler.CreateBanner)
				protected.PUT("/banners/:id", admBannerHandler.UpdateBanner)
				protected.DELETE("/banners/:id", admBannerHandler.DeleteBanner)

				// Theme, Holiday Events & Auth Methods Management
				protected.GET("/settings/theme", admThemeHandler.GetSettings)
				protected.POST("/settings/theme", admThemeHandler.UpdateSettings)

				// Media & Product Images Upload
				protected.POST("/upload/media", admUploadHandler.UploadMedia)

				// Backward-compatibility aliases
				protected.GET("/settings/5sim", admHandler.GetVirtualNumberSettings)
				protected.POST("/settings/5sim", admHandler.UpdateVirtualNumberSettings)
				protected.GET("/settings/insightx", admHandler.GetInstantDeliverySettings)
				protected.POST("/settings/insightx", admHandler.UpdateInstantDeliverySettings)
				protected.GET("/insightx/products", admHandler.ListProviderCatalog)
			}
		}
	}

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"brand":   svcCtx.Config.Site.BrandFA,
			"service": "Anahix AI Accounts Engine",
		})
	})
}
