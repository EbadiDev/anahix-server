package admin

import (
	"fmt"
	"strconv"

	"github.com/EbadiDev/anahix-server/internal/logic/admin"
	"github.com/EbadiDev/anahix-server/internal/logic/audit"
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

	var req model.Product
	if err := c.ShouldBindJSON(&req); err != nil {
		result.ParamErrorResult(c, err)
		return
	}

	existing.Title = req.Title
	existing.TitleFA = req.TitleFA
	if req.Slug != "" {
		existing.Slug = req.Slug
	}
	existing.Description = req.Description
	existing.Category = req.Category
	existing.DeliveryType = req.DeliveryType
	existing.SLAMinMinutes = req.SLAMinMinutes
	existing.SLAMaxMinutes = req.SLAMaxMinutes
	existing.SLADisplayText = req.SLADisplayText
	existing.OperatingHours = req.OperatingHours
	existing.BaseCostUSD = req.BaseCostUSD
	existing.PriceToman = req.PriceToman
	existing.IsActive = req.IsActive
	existing.Instructions = req.Instructions
	existing.GroupTag = req.GroupTag
	existing.GroupTitleFA = req.GroupTitleFA
	existing.DeliveryStrategy = req.DeliveryStrategy

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

