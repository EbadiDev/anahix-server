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
	vnumHandler := public.NewVirtualNumberHandler(svcCtx)
	admHandler := admin.NewAdminHandler(svcCtx)

	api := r.Group("/api/v1")
	{
		// Customer Authentication (SMS.ir OTP)
		authGroup := api.Group("/auth")
		{
			authGroup.POST("/otp/send", custAuthHandler.SendOTP)
			authGroup.POST("/otp/verify", custAuthHandler.VerifyOTP)
			authGroup.GET("/me", middleware.CustomerAuthMiddleware(svcCtx.Config.JWT.AccessSecret), custAuthHandler.GetMe)
			authGroup.PUT("/me", middleware.CustomerAuthMiddleware(svcCtx.Config.JWT.AccessSecret), custAuthHandler.UpdateMe)
			authGroup.GET("/my-reviews", middleware.CustomerAuthMiddleware(svcCtx.Config.JWT.AccessSecret), pubHandler.GetMyReviews)
		}

		// Public Storefront Endpoints
		products := api.Group("/products")
		{
			products.GET("", pubHandler.ListProducts)
			products.GET("/:slug", pubHandler.GetProductDetail)
			products.GET("/:slug/reviews", pubHandler.GetProductReviews)
			products.GET("/:slug/review-eligibility", middleware.CustomerOptionalAuthMiddleware(svcCtx.Config.JWT.AccessSecret), pubHandler.CheckReviewEligibility)
		}

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
		}

		payments := api.Group("/payments")
		{
			payments.GET("", pubHandler.ListPaymentMethods)
		}

		// Coupons
		coupons := api.Group("/coupons")
		{
			coupons.POST("/validate", pubHandler.ValidateCoupon)
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

			// Protected Admin Routes
			protected := adminGroup.Group("")
			protected.Use(middleware.AdminAuthMiddleware(svcCtx.Config.JWT.AccessSecret))
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

				// Buyer Reviews & Social Proof Moderation
				protected.GET("/reviews", admHandler.ListReviews)
				protected.PUT("/reviews/:id/status", admHandler.UpdateReviewStatus)
				protected.POST("/reviews/:id/reply", admHandler.ReplyReview)
				protected.PUT("/reviews/:id/feature", admHandler.FeatureReview)
				protected.DELETE("/reviews/:id", admHandler.DeleteReview)

				// Operator Order Queue Workbench
				protected.GET("/orders", admHandler.ListQueueOrders)
				protected.GET("/orders/:order_id", admHandler.GetOrderDetail)
				protected.POST("/orders/:order_id/2fa-request", admHandler.RequestTwoFactor)
				protected.POST("/orders/:order_id/complete", admHandler.CompleteOrder)

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

				// Audit & Transaction Logs
				protected.GET("/logs/transactions", admHandler.ListTransactions)
				protected.GET("/logs/actions", admHandler.ListAdminActions)

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
