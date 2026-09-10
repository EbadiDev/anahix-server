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
	admHandler := admin.NewAdminHandler(svcCtx)

	api := r.Group("/api/v1")
	{
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
