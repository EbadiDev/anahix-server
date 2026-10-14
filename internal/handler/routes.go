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
		}

		// Public Storefront Endpoints
		products := api.Group("/products")
		{
			products.GET("", pubHandler.ListProducts)
			products.GET("/:slug", pubHandler.GetProductDetail)
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
				protected.POST("/products", admHandler.CreateProduct)
				protected.POST("/inventory/bulk-import", admHandler.BulkImportInventory)

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
