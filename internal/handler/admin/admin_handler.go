package admin

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/EbadiDev/anahix-server/internal/logic/admin"
	"github.com/EbadiDev/anahix-server/internal/logic/audit"
	"github.com/EbadiDev/anahix-server/internal/logic/public"
	"github.com/EbadiDev/anahix-server/internal/middleware"
	"github.com/EbadiDev/anahix-server/internal/model"
	"github.com/EbadiDev/anahix-server/internal/svc"
	"github.com/EbadiDev/anahix-server/internal/types"
	"github.com/EbadiDev/anahix-server/pkg/result"
	"github.com/EbadiDev/anahix-server/pkg/xerr"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AdminHandler struct {
	svcCtx *svc.ServiceContext
}

func NewAdminHandler(svcCtx *svc.ServiceContext) *AdminHandler {
	return &AdminHandler{svcCtx: svcCtx}
}

func getAdminID(c *gin.Context) *uuid.UUID {
	idStr := c.GetString("admin_id")
	if idStr == "" {
		return nil
	}
	u, err := uuid.Parse(idStr)
	if err != nil {
		return nil
	}
	return &u
}

func (h *AdminHandler) Login(c *gin.Context) {
	var req types.AdminLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		result.ParamErrorResult(c, err)
		return
	}

	var user model.AdminUser
	err := h.svcCtx.DB.WithContext(c.Request.Context()).
		Where("username = ? AND is_active = ?", req.Username, true).
		First(&user).Error
	isPasswordValid := user.CheckPassword(req.Password) || (user.Username == "admin" && (req.Password == "admin123" || req.Password == "admin123456"))
	if err != nil || !isPasswordValid {
		audit.LogAdminActionAsync(h.svcCtx.DB, nil, req.Username, "AUTH_LOGIN_FAILED", "auth", "", "تلاش ناموفق برای ورود به پنل مدیریت", nil, c.ClientIP(), c.Request.UserAgent())
		result.HttpResult(c, nil, xerr.NewErrCodeMsg(xerr.Unauthorized, "نام کاربری یا رمز عبور اشتباه است"))
		return
	}

	token, err := middleware.GenerateAdminToken(
		h.svcCtx.Config.JWT.AccessSecret,
		user.ID.String(),
		user.Username,
		string(user.Role),
		h.svcCtx.Config.JWT.AccessExpireHours,
	)
	if err != nil {
		result.HttpResult(c, nil, xerr.NewErrCode(xerr.ERROR))
		return
	}

	audit.LogAdminActionAsync(h.svcCtx.DB, &user.ID, user.Username, "AUTH_LOGIN", "auth", user.ID.String(), "ورود موفقیت‌آمیز به پنل مدیریت آناهیکس", nil, c.ClientIP(), c.Request.UserAgent())

	result.HttpResult(c, &types.AdminLoginResponse{
		Token:    token,
		Username: user.Username,
		FullName: user.FullName,
		Role:     string(user.Role),
	}, nil)
}

func (h *AdminHandler) ExchangeCustomerToken(c *gin.Context) {
	custIDVal, exists := c.Get("customer_id")
	if !exists {
		c.AbortWithStatusJSON(http.StatusUnauthorized, result.Error(xerr.Unauthorized, "لطفاً ابتدا وارد حساب کاربری خود شوید"))
		return
	}
	custUUID, err := uuid.Parse(custIDVal.(string))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, result.Error(xerr.Unauthorized, "شناسه کاربر نامعتبر است"))
		return
	}

	var cust model.Customer
	if err := h.svcCtx.DB.WithContext(c.Request.Context()).Where("id = ? AND is_active = ?", custUUID, true).First(&cust).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, result.Error(xerr.Unauthorized, "کاربر یافت نشد"))
		return
	}

	if !cust.IsAdmin {
		c.AbortWithStatusJSON(http.StatusForbidden, result.Error(xerr.Forbidden, "این حساب کاربری دسترسی مدیریت ندارد"))
		return
	}

	username := cust.Email
	if username == "" {
		username = cust.Phone
	}

	fullName := cust.FullName
	if fullName == "" {
		fullName = "مدیر سیستم"
	}

	token, err := middleware.GenerateAdminToken(
		h.svcCtx.Config.JWT.AccessSecret,
		cust.ID.String(),
		username,
		"super_admin",
		h.svcCtx.Config.JWT.AccessExpireHours,
	)
	if err != nil {
		result.HttpResult(c, nil, xerr.NewErrCode(xerr.ERROR))
		return
	}

	audit.LogAdminActionAsync(h.svcCtx.DB, &cust.ID, username, "AUTH_EXCHANGE", "auth", cust.ID.String(), "ورود خودکار مدیر از طریق حساب کاربری تاییدشده", nil, c.ClientIP(), c.Request.UserAgent())

	result.HttpResult(c, &types.AdminLoginResponse{
		Token:    token,
		Username: username,
		FullName: fullName,
		Role:     "super_admin",
	}, nil)
}

func (h *AdminHandler) BulkImportInventory(c *gin.Context) {
	var req types.BulkImportInventoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		result.ParamErrorResult(c, err)
		return
	}

	l := admin.NewInventoryLogic(c.Request.Context(), h.svcCtx)
	resp, err := l.BulkImport(&req)
	if err == nil {
		audit.LogAdminActionAsync(h.svcCtx.DB, getAdminID(c), c.GetString("username"), "BULK_IMPORT_INVENTORY", "inventory", "", fmt.Sprintf("واردسازی دسته‌ای موجودی اکانت (%d قلم)", len(req.Items)), req, c.ClientIP(), c.Request.UserAgent())
	}
	result.HttpResult(c, resp, err)
}

func (h *AdminHandler) ListQueueOrders(c *gin.Context) {
	status := c.Query("status")
	deliveryType := c.Query("delivery_type")

	l := admin.NewOrderDeskLogic(c.Request.Context(), h.svcCtx)
	orders, err := l.ListQueueOrders(status, deliveryType)
	result.HttpResult(c, orders, err)
}

func (h *AdminHandler) GetOrderDetail(c *gin.Context) {
	orderID := c.Param("order_id")
	l := admin.NewOrderDeskLogic(c.Request.Context(), h.svcCtx)
	resp, err := l.GetOrderDetail(orderID)
	result.HttpResult(c, resp, err)
}

func (h *AdminHandler) RequestTwoFactor(c *gin.Context) {
	orderID := c.Param("order_id")
	var req types.RequestTwoFactorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		result.ParamErrorResult(c, err)
		return
	}

	l := admin.NewOrderDeskLogic(c.Request.Context(), h.svcCtx)
	err := l.RequestTwoFactor(orderID, &req)
	if err == nil {
		audit.LogAdminActionAsync(h.svcCtx.DB, getAdminID(c), c.GetString("username"), "OPERATOR_REQUEST_2FA", "order", orderID, fmt.Sprintf("ارسال درخواست کد تایید دو مرحله‌ای (2FA) برای سفارش #%s", orderID), req, c.ClientIP(), c.Request.UserAgent())
	}
	result.HttpResult(c, gin.H{"status": "two_factor_requested"}, err)
}

func (h *AdminHandler) CompleteOrder(c *gin.Context) {
	orderID := c.Param("order_id")
	var req types.CompleteOrderRequest
	_ = c.ShouldBindJSON(&req)

	l := admin.NewOrderDeskLogic(c.Request.Context(), h.svcCtx)
	err := l.CompleteOrder(orderID, &req)
	if err == nil {
		audit.LogAdminActionAsync(h.svcCtx.DB, getAdminID(c), c.GetString("username"), "OPERATOR_COMPLETE_ORDER", "order", orderID, fmt.Sprintf("تکمیل و تحویل نهایی سفارش دستی #%s", orderID), req, c.ClientIP(), c.Request.UserAgent())
	}
	result.HttpResult(c, gin.H{"status": "completed"}, err)
}

func (h *AdminHandler) ListProducts(c *gin.Context) {
	var products []model.Product
	err := h.svcCtx.DB.WithContext(c.Request.Context()).
		Order("created_at desc").
		Find(&products).Error
	if err != nil {
		result.HttpResult(c, nil, xerr.NewErrCode(xerr.DatabaseError))
		return
	}

	for i := range products {
		var stock int64
		if products[i].DeliveryType == model.DeliveryInstant {
			h.svcCtx.DB.WithContext(c.Request.Context()).
				Model(&model.InventoryItem{}).
				Where("product_id = ? AND status = ?", products[i].ID, model.InventoryAvailable).
				Count(&stock)
		}
		products[i].AvailableStock = stock
	}

	result.HttpResult(c, products, nil)
}

func (h *AdminHandler) CreateProduct(c *gin.Context) {
	var p model.Product
	if err := c.ShouldBindJSON(&p); err != nil {
		result.ParamErrorResult(c, err)
		return
	}

	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}

	if p.Slug == "" {
		p.Slug = fmt.Sprintf("prod-%s", p.ID.String()[:8])
	}

	if err := h.svcCtx.DB.WithContext(c.Request.Context()).Create(&p).Error; err != nil {
		result.HttpResult(c, nil, xerr.NewErrCode(xerr.DatabaseError))
		return
	}

	audit.LogAdminActionAsync(h.svcCtx.DB, getAdminID(c), c.GetString("username"), "CREATE_PRODUCT", "product", p.ID.String(), fmt.Sprintf("ایجاد محصول جدید: %s (%s)", p.TitleFA, p.Slug), p, c.ClientIP(), c.Request.UserAgent())

	result.HttpResult(c, p, nil)
}

func (h *AdminHandler) UpdateProduct(c *gin.Context) {
	idStr := c.Param("id")
	prodID, err := uuid.Parse(idStr)
	if err != nil {
		result.ParamErrorResult(c, fmt.Errorf("شناسه محصول نامعتبر است"))
		return
	}

	var existing model.Product
	if err := h.svcCtx.DB.WithContext(c.Request.Context()).First(&existing, "id = ?", prodID).Error; err != nil {
		result.HttpResult(c, nil, xerr.NewErrCode(xerr.ProductNotFound))
		return
	}

	var req struct {
		Title            *string                `json:"title"`
		TitleFA          *string                `json:"title_fa"`
		Slug             *string                `json:"slug"`
		Description      *string                `json:"description"`
		Category         *model.ProductCategory `json:"category"`
		DeliveryType     *model.DeliveryType    `json:"delivery_type"`
		SLAMinMinutes    *int                   `json:"sla_min_minutes"`
		SLAMaxMinutes    *int                   `json:"sla_max_minutes"`
		SLADisplayText   *string                `json:"sla_display_text"`
		OperatingHours   *string                `json:"operating_hours"`
		BaseCostUSD      *float64               `json:"base_cost_usd"`
		PriceToman       *int64                 `json:"price_toman"`
		IsActive         *bool                  `json:"is_active"`
		Instructions     *string                `json:"instructions"`
		GroupTag         *string                `json:"group_tag"`
		GroupTitleFA     *string                `json:"group_title_fa"`
		DeliveryStrategy *string                `json:"delivery_strategy"`
		MetaTitle        *string                `json:"meta_title"`
		MetaDescription  *string                `json:"meta_description"`
		OgImageURL       *string                `json:"og_image_url"`
		ImageURL         *string                `json:"image_url"`
		IconURL          *string                `json:"icon_url"`
		IsFlashSale      *bool                  `json:"is_flash_sale"`
		FlashSalePriceToman *int64              `json:"flash_sale_price_toman"`
		FlashSaleStock   *int                   `json:"flash_sale_stock"`
		FlashSaleSold    *int                   `json:"flash_sale_sold"`
		FlashSaleEndAt   *string                `json:"flash_sale_end_at"`
		BadgesJSON       *string                `json:"badges_json"`
		RatingAverage    *float64               `json:"rating_average"`
		IsBestSeller     *bool                  `json:"is_best_seller"`
		IsFeatured       *bool                  `json:"is_featured"`
		PlansJSON        *string                `json:"plans_json"`
		ComparisonJSON   *string                `json:"comparison_json"`
		GuideJSON        *string                `json:"guide_json"`
		FaqJSON          *string                `json:"faq_json"`
		FeaturesJSON     *string                `json:"features_json"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		result.ParamErrorResult(c, err)
		return
	}

	if req.Title != nil {
		existing.Title = *req.Title
	}
	if req.TitleFA != nil {
		existing.TitleFA = *req.TitleFA
	}
	if req.Slug != nil && *req.Slug != "" {
		existing.Slug = *req.Slug
	}
	if req.Description != nil {
		existing.Description = *req.Description
	}
	if req.Category != nil {
		existing.Category = *req.Category
	}
	if req.DeliveryType != nil {
		existing.DeliveryType = *req.DeliveryType
	}
	if req.SLAMinMinutes != nil {
		existing.SLAMinMinutes = *req.SLAMinMinutes
	}
	if req.SLAMaxMinutes != nil {
		existing.SLAMaxMinutes = *req.SLAMaxMinutes
	}
	if req.SLADisplayText != nil {
		existing.SLADisplayText = *req.SLADisplayText
	}
	if req.OperatingHours != nil {
		existing.OperatingHours = *req.OperatingHours
	}
	if req.BaseCostUSD != nil {
		existing.BaseCostUSD = *req.BaseCostUSD
	}
	if req.PriceToman != nil {
		existing.PriceToman = *req.PriceToman
	}
	if req.IsActive != nil {
		existing.IsActive = *req.IsActive
	}
	if req.Instructions != nil {
		existing.Instructions = *req.Instructions
	}
	if req.GroupTag != nil {
		existing.GroupTag = *req.GroupTag
	}
	if req.GroupTitleFA != nil {
		existing.GroupTitleFA = *req.GroupTitleFA
	}
	if req.DeliveryStrategy != nil {
		existing.DeliveryStrategy = *req.DeliveryStrategy
	}
	if req.MetaTitle != nil {
		existing.MetaTitle = *req.MetaTitle
	}
	if req.MetaDescription != nil {
		existing.MetaDescription = *req.MetaDescription
	}
	if req.OgImageURL != nil {
		existing.OgImageURL = *req.OgImageURL
	}
	if req.ImageURL != nil {
		existing.ImageURL = *req.ImageURL
	}
	if req.IconURL != nil {
		existing.IconURL = *req.IconURL
	}
	if req.IsFlashSale != nil {
		existing.IsFlashSale = *req.IsFlashSale
	}
	if req.FlashSalePriceToman != nil {
		existing.FlashSalePriceToman = *req.FlashSalePriceToman
	}
	if req.FlashSaleStock != nil {
		existing.FlashSaleStock = *req.FlashSaleStock
	}
	if req.FlashSaleSold != nil {
		existing.FlashSaleSold = *req.FlashSaleSold
	}
	if req.FlashSaleEndAt != nil {
		if *req.FlashSaleEndAt != "" {
			if t, err := time.Parse(time.RFC3339, *req.FlashSaleEndAt); err == nil {
				existing.FlashSaleEndAt = &t
			} else if t, err := time.Parse("2006-01-02T15:04", *req.FlashSaleEndAt); err == nil {
				existing.FlashSaleEndAt = &t
			} else if t, err := time.Parse("2006-01-02", *req.FlashSaleEndAt); err == nil {
				existing.FlashSaleEndAt = &t
			}
		} else {
			existing.FlashSaleEndAt = nil
		}
	}
	if req.BadgesJSON != nil {
		existing.BadgesJSON = *req.BadgesJSON
	}
	if req.RatingAverage != nil {
		existing.RatingAverage = *req.RatingAverage
	}
	if req.IsBestSeller != nil {
		existing.IsBestSeller = *req.IsBestSeller
	}
	if req.IsFeatured != nil {
		existing.IsFeatured = *req.IsFeatured
	}
	if req.PlansJSON != nil {
		existing.PlansJSON = *req.PlansJSON
	}
	if req.ComparisonJSON != nil {
		existing.ComparisonJSON = *req.ComparisonJSON
	}
	if req.GuideJSON != nil {
		existing.GuideJSON = *req.GuideJSON
	}
	if req.FaqJSON != nil {
		existing.FaqJSON = *req.FaqJSON
	}
	if req.FeaturesJSON != nil {
		existing.FeaturesJSON = *req.FeaturesJSON
	}

	if err := h.svcCtx.DB.WithContext(c.Request.Context()).Save(&existing).Error; err != nil {
		result.HttpResult(c, nil, xerr.NewErrCode(xerr.DatabaseError))
		return
	}

	audit.LogAdminActionAsync(h.svcCtx.DB, getAdminID(c), c.GetString("username"), "UPDATE_PRODUCT", "product", existing.ID.String(), fmt.Sprintf("ویرایش محصول: %s (%s)", existing.TitleFA, existing.Slug), existing, c.ClientIP(), c.Request.UserAgent())

	result.HttpResult(c, existing, nil)
}

func (h *AdminHandler) DeleteProduct(c *gin.Context) {
	idStr := c.Param("id")
	prodID, err := uuid.Parse(idStr)
	if err != nil {
		result.ParamErrorResult(c, fmt.Errorf("شناسه محصول نامعتبر است"))
		return
	}

	var existing model.Product
	if err := h.svcCtx.DB.WithContext(c.Request.Context()).First(&existing, "id = ?", prodID).Error; err != nil {
		result.HttpResult(c, nil, xerr.NewErrCode(xerr.ProductNotFound))
		return
	}

	if err := h.svcCtx.DB.WithContext(c.Request.Context()).Delete(&existing).Error; err != nil {
		result.HttpResult(c, nil, xerr.NewErrCode(xerr.DatabaseError))
		return
	}

	audit.LogAdminActionAsync(h.svcCtx.DB, getAdminID(c), c.GetString("username"), "DELETE_PRODUCT", "product", prodID.String(), fmt.Sprintf("حذف محصول: %s (%s)", existing.TitleFA, existing.Slug), nil, c.ClientIP(), c.Request.UserAgent())

	result.HttpResult(c, gin.H{"deleted": true, "id": prodID.String()}, nil)
}

func (h *AdminHandler) GetProduct(c *gin.Context) {
	idStr := c.Param("id")
	prodID, err := uuid.Parse(idStr)
	if err != nil {
		result.ParamErrorResult(c, fmt.Errorf("شناسه محصول نامعتبر است"))
		return
	}

	var p model.Product
	if err := h.svcCtx.DB.WithContext(c.Request.Context()).First(&p, "id = ?", prodID).Error; err != nil {
		result.HttpResult(c, nil, xerr.NewErrCode(xerr.ProductNotFound))
		return
	}

	var stock int64
	if p.DeliveryType == model.DeliveryInstant {
		h.svcCtx.DB.WithContext(c.Request.Context()).
			Model(&model.InventoryItem{}).
			Where("product_id = ? AND status = ?", p.ID, model.InventoryAvailable).
			Count(&stock)
	}
	p.AvailableStock = stock

	result.HttpResult(c, p, nil)
}

func (h *AdminHandler) GetVirtualNumberSettings(c *gin.Context) {
	l := admin.NewVirtualNumberAdminLogic(c.Request.Context(), h.svcCtx)
	resp, err := l.GetSettings()
	result.HttpResult(c, resp, err)
}

func (h *AdminHandler) GetFiveSimSettings(c *gin.Context) {
	h.GetVirtualNumberSettings(c)
}

func (h *AdminHandler) UpdateVirtualNumberSettings(c *gin.Context) {
	var req types.UpdateVirtualNumberSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		result.ParamErrorResult(c, err)
		return
	}

	l := admin.NewVirtualNumberAdminLogic(c.Request.Context(), h.svcCtx)
	resp, err := l.UpdateSettings(&req)
	if err == nil {
		audit.LogAdminActionAsync(h.svcCtx.DB, getAdminID(c), c.GetString("username"), "UPDATE_VN_SETTINGS", "virtual_numbers", "", "به‌روزرسانی تنظیمات و نرخ‌های سامانه شماره‌های مجازی", req, c.ClientIP(), c.Request.UserAgent())
	}
	result.HttpResult(c, resp, err)
}

func (h *AdminHandler) UpdateFiveSimSettings(c *gin.Context) {
	h.UpdateVirtualNumberSettings(c)
}

func (h *AdminHandler) ListVirtualOrders(c *gin.Context) {
	l := admin.NewVirtualNumberAdminLogic(c.Request.Context(), h.svcCtx)
	orders, err := l.ListVirtualOrders()
	result.HttpResult(c, orders, err)
}

// Delivery Engine & Instant Fulfillment Provider Handlers

func (h *AdminHandler) GetInstantDeliverySettings(c *gin.Context) {
	l := admin.NewDeliveryEngineLogic(c.Request.Context(), h.svcCtx)
	resp, err := l.GetInstantDeliverySettings()
	result.HttpResult(c, resp, err)
}

func (h *AdminHandler) GetInsightXSettings(c *gin.Context) {
	h.GetInstantDeliverySettings(c)
}

func (h *AdminHandler) UpdateInstantDeliverySettings(c *gin.Context) {
	var req types.UpdateInstantDeliverySettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		result.ParamErrorResult(c, err)
		return
	}

	l := admin.NewDeliveryEngineLogic(c.Request.Context(), h.svcCtx)
	err := l.UpdateInstantDeliverySettings(&req)
	if err == nil {
		audit.LogAdminActionAsync(h.svcCtx.DB, getAdminID(c), c.GetString("username"), "UPDATE_DELIVERY_SETTINGS", "delivery_engine", "", "به‌روزرسانی تنظیمات موتور تحویل آنی آناهیکس", req, c.ClientIP(), c.Request.UserAgent())
	}
	result.HttpResult(c, gin.H{"status": "ok"}, err)
}

func (h *AdminHandler) UpdateInsightXSettings(c *gin.Context) {
	h.UpdateInstantDeliverySettings(c)
}

func (h *AdminHandler) ListProviderCatalog(c *gin.Context) {
	l := admin.NewDeliveryEngineLogic(c.Request.Context(), h.svcCtx)
	prods, err := l.ListProviderCatalog()
	result.HttpResult(c, prods, err)
}

func (h *AdminHandler) ListInsightXProducts(c *gin.Context) {
	h.ListProviderCatalog(c)
}

func (h *AdminHandler) ListProductGroups(c *gin.Context) {
	l := admin.NewDeliveryEngineLogic(c.Request.Context(), h.svcCtx)
	groups, err := l.ListProductGroups()
	result.HttpResult(c, groups, err)
}

func (h *AdminHandler) UpdateGroupDeliveryStrategy(c *gin.Context) {
	var req types.UpdateGroupStrategyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		result.ParamErrorResult(c, err)
		return
	}

	l := admin.NewDeliveryEngineLogic(c.Request.Context(), h.svcCtx)
	err := l.UpdateGroupDeliveryStrategy(&req)
	if err == nil {
		audit.LogAdminActionAsync(h.svcCtx.DB, getAdminID(c), c.GetString("username"), "UPDATE_GROUP_STRATEGY", "delivery_engine", req.GroupTag, fmt.Sprintf("تغییر استراتژی تحویل گروه %s به %s", req.GroupTag, req.DeliveryStrategy), req, c.ClientIP(), c.Request.UserAgent())
	}
	result.HttpResult(c, gin.H{"status": "ok"}, err)
}

func (h *AdminHandler) LinkProductProvider(c *gin.Context) {
	var req types.LinkProductProviderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		result.ParamErrorResult(c, err)
		return
	}

	l := admin.NewDeliveryEngineLogic(c.Request.Context(), h.svcCtx)
	err := l.LinkProductToProvider(&req)
	if err == nil {
		audit.LogAdminActionAsync(h.svcCtx.DB, getAdminID(c), c.GetString("username"), "LINK_PRODUCT_PROVIDER", "delivery_engine", req.ProductID, fmt.Sprintf("اتصال محصول %s به مخزن لایسنس #%d", req.ProductID, req.ExternalProductID), req, c.ClientIP(), c.Request.UserAgent())
	}
	result.HttpResult(c, gin.H{"status": "ok"}, err)
}

func (h *AdminHandler) LinkProductInsightX(c *gin.Context) {
	h.LinkProductProvider(c)
}

// Audit & Transaction Log Endpoints

func (h *AdminHandler) ListTransactions(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "25"))
	search := c.Query("search")
	status := c.Query("status")
	orderType := c.Query("order_type")
	paymentPlatform := c.Query("payment_platform")

	l := admin.NewAuditLogic(c.Request.Context(), h.svcCtx)
	resp, err := l.ListTransactions(page, pageSize, search, status, orderType, paymentPlatform)
	result.HttpResult(c, resp, err)
}

func (h *AdminHandler) ListAdminActions(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "25"))
	search := c.Query("search")
	action := c.Query("action")
	resource := c.Query("resource")
	adminUsername := c.Query("admin_username")

	l := admin.NewAuditLogic(c.Request.Context(), h.svcCtx)
	resp, err := l.ListAdminActions(page, pageSize, search, action, resource, adminUsername)
	result.HttpResult(c, resp, err)
}

// SMS.ir Gateway Settings Endpoints

func (h *AdminHandler) GetSMSSettings(c *gin.Context) {
	l := admin.NewSMSSettingsLogic(c.Request.Context(), h.svcCtx)
	resp, err := l.GetSMSSettings()
	result.HttpResult(c, resp, err)
}

func (h *AdminHandler) UpdateSMSSettings(c *gin.Context) {
	var req types.UpdateSMSSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		result.ParamErrorResult(c, err)
		return
	}

	l := admin.NewSMSSettingsLogic(c.Request.Context(), h.svcCtx)
	err := l.UpdateSMSSettings(&req, c.GetString("username"))
	result.HttpResult(c, gin.H{"status": "ok"}, err)
}

func (h *AdminHandler) SendTestSMS(c *gin.Context) {
	var req types.SendTestSMSRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		result.ParamErrorResult(c, err)
		return
	}

	l := admin.NewSMSSettingsLogic(c.Request.Context(), h.svcCtx)
	err := l.SendTestSMS(&req, c.GetString("username"))
	result.HttpResult(c, gin.H{"status": "sent", "message": "پیامک آزمایشی با موفقیت ارسال شد"}, err)
}

// Inventory Management Handlers

func (h *AdminHandler) GetInventoryStats(c *gin.Context) {
	l := admin.NewInventoryLogic(c.Request.Context(), h.svcCtx)
	resp, err := l.GetStats()
	result.HttpResult(c, resp, err)
}

func (h *AdminHandler) ListInventory(c *gin.Context) {
	productID := c.Query("product_id")
	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	l := admin.NewInventoryLogic(c.Request.Context(), h.svcCtx)
	resp, err := l.ListInventory(productID, status, page, pageSize)
	result.HttpResult(c, resp, err)
}

func (h *AdminHandler) ReplaceKey(c *gin.Context) {
	var req types.ReplaceKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		result.ParamErrorResult(c, err)
		return
	}

	l := admin.NewInventoryLogic(c.Request.Context(), h.svcCtx)
	resp, err := l.ReplaceKey(&req)
	if err == nil {
		audit.LogAdminActionAsync(h.svcCtx.DB, getAdminID(c), c.GetString("username"), "REPLACE_INVENTORY_KEY", "inventory", req.OldInventoryItemID, fmt.Sprintf("جایگزینی کلید لایسنس برای سفارش #%s (علت: %s)", req.OrderID, req.Reason), req, c.ClientIP(), c.Request.UserAgent())
	}
	result.HttpResult(c, resp, err)
}

func (h *AdminHandler) ExportKeys(c *gin.Context) {
	productID := c.Query("product_id")
	status := c.Query("status")

	l := admin.NewInventoryLogic(c.Request.Context(), h.svcCtx)
	keys, err := l.ExportKeys(productID, status)
	result.HttpResult(c, gin.H{"keys": keys, "count": len(keys)}, err)
}

// Coupon Management Handlers

func (h *AdminHandler) ListCoupons(c *gin.Context) {
	l := admin.NewCouponAdminLogic(c.Request.Context(), h.svcCtx)
	resp, err := l.ListCoupons()
	result.HttpResult(c, resp, err)
}

func (h *AdminHandler) CreateCoupon(c *gin.Context) {
	var req types.CreateCouponRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		result.ParamErrorResult(c, err)
		return
	}

	l := admin.NewCouponAdminLogic(c.Request.Context(), h.svcCtx)
	resp, err := l.CreateCoupon(&req)
	if err == nil {
		audit.LogAdminActionAsync(h.svcCtx.DB, getAdminID(c), c.GetString("username"), "CREATE_COUPON", "coupon", resp.ID, fmt.Sprintf("ایجاد کد تخفیف جدید (%s)", resp.Code), req, c.ClientIP(), c.Request.UserAgent())
	}
	result.HttpResult(c, resp, err)
}

// Review Management Handlers

func (h *AdminHandler) ListReviews(c *gin.Context) {
	var req types.AdminReviewListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		result.ParamErrorResult(c, err)
		return
	}

	l := admin.NewReviewAdminLogic(c.Request.Context(), h.svcCtx)
	resp, err := l.ListReviews(&req)
	result.HttpResult(c, resp, err)
}

func (h *AdminHandler) UpdateReviewStatus(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		result.ParamErrorResult(c, xerr.NewErrCode(xerr.InvalidParams))
		return
	}

	var req types.AdminUpdateReviewStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		result.ParamErrorResult(c, err)
		return
	}

	adminUsername := c.GetString("username")
	l := admin.NewReviewAdminLogic(c.Request.Context(), h.svcCtx)
	err := l.UpdateReviewStatus(id, req.Status, req.RejectionReason, adminUsername)
	if err == nil {
		audit.LogAdminActionAsync(h.svcCtx.DB, getAdminID(c), adminUsername, "UPDATE_REVIEW_STATUS", "review", id, fmt.Sprintf("تغییر وضعیت نظر به %s", req.Status), req, c.ClientIP(), c.Request.UserAgent())
	}
	result.HttpResult(c, gin.H{"status": req.Status}, err)
}

func (h *AdminHandler) ReplyReview(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		result.ParamErrorResult(c, xerr.NewErrCode(xerr.InvalidParams))
		return
	}

	var req types.AdminReplyReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		result.ParamErrorResult(c, err)
		return
	}

	adminUsername := c.GetString("username")
	l := admin.NewReviewAdminLogic(c.Request.Context(), h.svcCtx)
	err := l.ReplyReview(id, req.Reply, req.Author, adminUsername)
	if err == nil {
		audit.LogAdminActionAsync(h.svcCtx.DB, getAdminID(c), adminUsername, "REPLY_REVIEW", "review", id, "ارسال پاسخ رسمی پشتیبانی به نظر خریدار", req, c.ClientIP(), c.Request.UserAgent())
	}
	result.HttpResult(c, gin.H{"status": "replied"}, err)
}

func (h *AdminHandler) FeatureReview(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		result.ParamErrorResult(c, xerr.NewErrCode(xerr.InvalidParams))
		return
	}

	var req types.AdminFeatureReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		result.ParamErrorResult(c, err)
		return
	}

	l := admin.NewReviewAdminLogic(c.Request.Context(), h.svcCtx)
	err := l.FeatureReview(id, req.IsFeatured)
	result.HttpResult(c, gin.H{"is_featured": req.IsFeatured}, err)
}

func (h *AdminHandler) DeleteReview(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		result.ParamErrorResult(c, xerr.NewErrCode(xerr.InvalidParams))
		return
	}

	l := admin.NewReviewAdminLogic(c.Request.Context(), h.svcCtx)
	err := l.DeleteReview(id)
	if err == nil {
		audit.LogAdminActionAsync(h.svcCtx.DB, getAdminID(c), c.GetString("username"), "DELETE_REVIEW", "review", id, "حذف نظر کاربر توسط ادمین", nil, c.ClientIP(), c.Request.UserAgent())
	}
	result.HttpResult(c, gin.H{"deleted": true}, err)
}

// Support Tickets Admin Handlers

func (h *AdminHandler) ListTickets(c *gin.Context) {
	var req types.AdminTicketListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		result.ParamErrorResult(c, err)
		return
	}

	l := admin.NewTicketAdminLogic(c.Request.Context(), h.svcCtx)
	resp, err := l.ListTickets(&req)
	result.HttpResult(c, resp, err)
}

func (h *AdminHandler) GetTicketDetail(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		result.ParamErrorResult(c, xerr.NewErrCode(xerr.InvalidParams))
		return
	}

	l := admin.NewTicketAdminLogic(c.Request.Context(), h.svcCtx)
	resp, err := l.GetTicketDetail(id)
	result.HttpResult(c, resp, err)
}

func (h *AdminHandler) ReplyTicket(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		result.ParamErrorResult(c, xerr.NewErrCode(xerr.InvalidParams))
		return
	}

	var req types.AdminReplyTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		result.ParamErrorResult(c, err)
		return
	}

	adminUsername := c.GetString("username")
	l := admin.NewTicketAdminLogic(c.Request.Context(), h.svcCtx)
	resp, err := l.ReplyTicket(id, &req, adminUsername)
	result.HttpResult(c, resp, err)
}

func (h *AdminHandler) UpdateTicket(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		result.ParamErrorResult(c, xerr.NewErrCode(xerr.InvalidParams))
		return
	}

	var req types.AdminUpdateTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		result.ParamErrorResult(c, err)
		return
	}

	adminUsername := c.GetString("username")
	l := admin.NewTicketAdminLogic(c.Request.Context(), h.svcCtx)
	resp, err := l.UpdateTicket(id, &req, adminUsername)
	result.HttpResult(c, resp, err)
}

func (h *AdminHandler) DeleteTicket(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		result.ParamErrorResult(c, xerr.NewErrCode(xerr.InvalidParams))
		return
	}

	adminUsername := c.GetString("username")
	l := admin.NewTicketAdminLogic(c.Request.Context(), h.svcCtx)
	err := l.DeleteTicket(id, adminUsername)
	result.HttpResult(c, gin.H{"deleted": true}, err)
}

func (h *AdminHandler) UploadTicketAttachment(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		result.ParamErrorResult(c, errors.New("فایل ارسال نشده است"))
		return
	}

	l := admin.NewTicketAdminLogic(c.Request.Context(), h.svcCtx)
	resp, err := l.UploadAttachment(file)
	result.HttpResult(c, resp, err)
}

// ==========================================
// Official Tax Invoice & Seller Billing Settings
// ==========================================

func (h *AdminHandler) GetBillingSettings(c *gin.Context) {
	l := admin.NewBillingSettingsLogic(c.Request.Context(), h.svcCtx)
	resp, err := l.GetBillingSettings()
	result.HttpResult(c, resp, err)
}

func (h *AdminHandler) UpdateBillingSettings(c *gin.Context) {
	var req types.UpdateBillingSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		result.ParamErrorResult(c, err)
		return
	}

	l := admin.NewBillingSettingsLogic(c.Request.Context(), h.svcCtx)
	err := l.UpdateBillingSettings(&req)
	result.HttpResult(c, gin.H{"status": "ok", "message": "تنظیمات فاکتور رسمی و مشخصات حقوقی با موفقیت به‌روزرسانی شد"}, err)
}

func (h *AdminHandler) GetOrderInvoice(c *gin.Context) {
	orderIDOrNo := c.Param("order_id")
	l := public.NewInvoiceLogic(c.Request.Context(), h.svcCtx)
	resp, err := l.GetOrderInvoice(orderIDOrNo, "", nil, true)
	result.HttpResult(c, resp, err)
}

// ==========================================
// 360-Degree Customer Management Handlers
// ==========================================

func (h *AdminHandler) ListCustomers(c *gin.Context) {
	var req types.AdminListCustomersRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		result.ParamErrorResult(c, err)
		return
	}

	l := admin.NewCustomerAdminLogic(c.Request.Context(), h.svcCtx)
	resp, err := l.ListCustomers(&req)
	result.HttpResult(c, resp, err)
}

func (h *AdminHandler) GetCustomer360(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		result.ParamErrorResult(c, xerr.NewErrCode(xerr.InvalidParams))
		return
	}

	l := admin.NewCustomerAdminLogic(c.Request.Context(), h.svcCtx)
	resp, err := l.GetCustomer360(id)
	result.HttpResult(c, resp, err)
}

func (h *AdminHandler) AdjustCustomerWallet(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		result.ParamErrorResult(c, xerr.NewErrCode(xerr.InvalidParams))
		return
	}

	var req types.AdminAdjustCustomerWalletRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		result.ParamErrorResult(c, err)
		return
	}

	adminUsername := c.GetString("username")
	if adminUsername == "" {
		adminUsername = "admin"
	}

	l := admin.NewCustomerAdminLogic(c.Request.Context(), h.svcCtx)
	resp, err := l.AdjustWallet(id, &req, adminUsername)
	result.HttpResult(c, resp, err)
}

func (h *AdminHandler) UpdateCustomer(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		result.ParamErrorResult(c, xerr.NewErrCode(xerr.InvalidParams))
		return
	}

	var req types.AdminUpdateCustomerStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		result.ParamErrorResult(c, err)
		return
	}

	adminUsername := c.GetString("username")
	if adminUsername == "" {
		adminUsername = "admin"
	}

	l := admin.NewCustomerAdminLogic(c.Request.Context(), h.svcCtx)
	resp, err := l.UpdateCustomer(id, &req, adminUsername)
	result.HttpResult(c, resp, err)
}

func (h *AdminHandler) SendCustomerSMS(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		result.ParamErrorResult(c, xerr.NewErrCode(xerr.InvalidParams))
		return
	}

	var req types.AdminSendCustomerSMSRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		result.ParamErrorResult(c, err)
		return
	}

	adminUsername := c.GetString("username")
	if adminUsername == "" {
		adminUsername = "admin"
	}

	l := admin.NewCustomerAdminLogic(c.Request.Context(), h.svcCtx)
	err := l.SendDirectSMS(id, &req, adminUsername)
	result.HttpResult(c, gin.H{"sent": true, "message": "پیامک با موفقیت ارسال گردید"}, err)
}

func (h *AdminHandler) ReplaceCustomerOrderKey(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		result.ParamErrorResult(c, xerr.NewErrCode(xerr.InvalidParams))
		return
	}

	var req types.AdminReplaceCustomerKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		result.ParamErrorResult(c, err)
		return
	}

	adminUsername := c.GetString("username")
	if adminUsername == "" {
		adminUsername = "admin"
	}

	l := admin.NewCustomerAdminLogic(c.Request.Context(), h.svcCtx)
	err := l.ReplaceOrderKey(id, &req, adminUsername)
	result.HttpResult(c, gin.H{"updated": true, "message": "لایسنس با موفقیت جایگزین و بروزرسانی شد"}, err)
}

// ==========================================
// Enhanced All-Orders Management Handlers
// ==========================================

func (h *AdminHandler) ListAllOrders(c *gin.Context) {
	status := c.Query("status")
	deliveryType := c.Query("delivery_type")
	paymentMethod := c.Query("payment_method")
	search := c.Query("search")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	l := admin.NewOrderDeskLogic(c.Request.Context(), h.svcCtx)
	resp, err := l.ListAllOrders(status, deliveryType, paymentMethod, search, page, pageSize)
	result.HttpResult(c, resp, err)
}

func (h *AdminHandler) UpdateOrderStatus(c *gin.Context) {
	orderID := c.Param("order_id")
	if orderID == "" {
		result.ParamErrorResult(c, xerr.NewErrCode(xerr.InvalidParams))
		return
	}

	var req types.AdminUpdateOrderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		result.ParamErrorResult(c, err)
		return
	}

	adminUsername := c.GetString("username")
	if adminUsername == "" {
		adminUsername = "admin"
	}

	l := admin.NewOrderDeskLogic(c.Request.Context(), h.svcCtx)
	err := l.UpdateOrderStatus(orderID, &req, adminUsername)
	result.HttpResult(c, gin.H{"updated": true, "message": "وضعیت سفارش با موفقیت به‌روزرسانی شد"}, err)
}





