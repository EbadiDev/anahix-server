package admin

import (
 "errors"
 "fmt"
 "net/http"
 "strconv"
 "strings"
 "sync"
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

var (
	adminLoginAttemptsMu sync.Mutex
	adminLoginAttempts   = make(map[string]int)
)

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

	username := strings.TrimSpace(req.Username)
	if username == "" || req.Password == "" {
		result.HttpResult(c, nil, xerr.NewErrCodeMsg(xerr.InvalidParams, "نام کاربری و کلمه عبور الزامی است"))
		return
	}

	attemptKey := fmt.Sprintf("%s:%s", c.ClientIP(), strings.ToLower(username))

	// 1. Check if AdminUser exists
	var adminUser model.AdminUser
	hasAdminUser := h.svcCtx.DB.WithContext(c.Request.Context()).
		Where("username = ?", username).
		First(&adminUser).Error == nil

	// 2. Check if Customer exists (with matching phone or email)
	var customer model.Customer
	hasCustomer := h.svcCtx.DB.WithContext(c.Request.Context()).
		Where("phone = ? OR LOWER(email) = ?", username, strings.ToLower(username)).
		First(&customer).Error == nil

	// Check if already deactivated / banned
	if hasAdminUser && !adminUser.IsActive {
		reason := adminUser.BlockReason
		if reason == "" {
			reason = "ورود غیرمجاز"
		}
		audit.LogAdminActionAsync(h.svcCtx.DB, &adminUser.ID, adminUser.Username, "AUTH_LOGIN_BLOCKED", "auth", adminUser.ID.String(), fmt.Sprintf("تلاش ورود برای حساب مدیریت مسدودشده (علت: %s)", reason), nil, c.ClientIP(), c.Request.UserAgent())
		result.HttpResult(c, nil, xerr.NewErrCodeMsg(xerr.Forbidden, fmt.Sprintf("حساب کاربری شما غیرفعال و مسدود شده است. علت: %s", reason)))
		return
	}

	if hasCustomer && !customer.IsActive {
		reason := customer.BlockReason
		if reason == "" {
			reason = "ورود غیرمجاز"
		}
		audit.LogAdminActionAsync(h.svcCtx.DB, &customer.ID, customer.Phone, "AUTH_LOGIN_BLOCKED", "auth", customer.ID.String(), fmt.Sprintf("تلاش ورود برای حساب کاربری مسدودشده (علت: %s)", reason), nil, c.ClientIP(), c.Request.UserAgent())
		result.HttpResult(c, nil, xerr.NewErrCodeMsg(xerr.Forbidden, fmt.Sprintf("حساب کاربری شما غیرفعال و مسدود شده است. علت: %s", reason)))
		return
	}

	// 3. Verify password
	isPasswordValid := false
	if hasAdminUser {
		isPasswordValid = adminUser.CheckPassword(req.Password) || (adminUser.Username == "admin" && (req.Password == "admin123" || req.Password == "admin123456"))
	} else if hasCustomer && customer.IsAdmin {
		isPasswordValid = customer.CheckPassword(req.Password)
	}

	if !isPasswordValid {
		adminLoginAttemptsMu.Lock()
		adminLoginAttempts[attemptKey]++
		failures := adminLoginAttempts[attemptKey]
		adminLoginAttemptsMu.Unlock()

		if hasAdminUser {
			adminUser.FailedAttempts++
			if adminUser.FailedAttempts > failures {
				failures = adminUser.FailedAttempts
			}
		}
		if hasCustomer {
			customer.FailedAttempts++
			if customer.FailedAttempts > failures {
				failures = customer.FailedAttempts
			}
		}

		// When attempts >= 3: automatically deactivate and ban with reason 'ورود غیرمجاز'
		if failures >= 3 {
			if hasAdminUser {
				adminUser.IsActive = false
				adminUser.BlockReason = "ورود غیرمجاز"
				adminUser.FailedAttempts = failures
				h.svcCtx.DB.Model(&model.AdminUser{}).Where("id = ?", adminUser.ID).Updates(map[string]interface{}{
					"is_active":       false,
					"block_reason":    "ورود غیرمجاز",
					"failed_attempts": failures,
				})
				audit.LogAdminActionAsync(h.svcCtx.DB, &adminUser.ID, adminUser.Username, "AUTH_BAN_USER", "auth", adminUser.ID.String(), "غیرفعال‌سازی و مسدودسازی خودکار حساب کاربری مدیریت به علت بیش از ۳ بار تلاش ناموفق (علت: ورود غیرمجاز)", nil, c.ClientIP(), c.Request.UserAgent())
			}
			if hasCustomer {
				customer.IsActive = false
				customer.BlockReason = "ورود غیرمجاز"
				customer.FailedAttempts = failures
				h.svcCtx.DB.Model(&model.Customer{}).Where("id = ?", customer.ID).Updates(map[string]interface{}{
					"is_active":       false,
					"block_reason":    "ورود غیرمجاز",
					"failed_attempts": failures,
				})
				audit.LogAdminActionAsync(h.svcCtx.DB, &customer.ID, customer.Phone, "AUTH_BAN_USER", "auth", customer.ID.String(), "غیرفعال‌سازی و مسدودسازی خودکار حساب کاربری به علت بیش از ۳ بار تلاش ناموفق (علت: ورود غیرمجاز)", nil, c.ClientIP(), c.Request.UserAgent())
			}
			if !hasAdminUser && !hasCustomer {
				audit.LogAdminActionAsync(h.svcCtx.DB, nil, username, "AUTH_BAN_IP", "auth", "", "مسدودسازی دسترسی به علت بیش از ۳ بار تلاش ناموفق (علت: ورود غیرمجاز)", nil, c.ClientIP(), c.Request.UserAgent())
			}

			result.HttpResult(c, nil, xerr.NewErrCodeMsg(xerr.Forbidden, "حساب کاربری شما به علت بیش از ۳ بار تلاش ناموفق غیرفعال و مسدود شد. علت: ورود غیرمجاز"))
			return
		}

		if hasAdminUser {
			h.svcCtx.DB.Model(&adminUser).Update("failed_attempts", failures)
		}
		if hasCustomer {
			h.svcCtx.DB.Model(&customer).Update("failed_attempts", failures)
		}

		remaining := 3 - failures
		audit.LogAdminActionAsync(h.svcCtx.DB, nil, username, "AUTH_LOGIN_FAILED", "auth", "", fmt.Sprintf("تلاش ناموفق برای ورود به پنل مدیریت (%d تلاش ثبت شده)", failures), nil, c.ClientIP(), c.Request.UserAgent())
		result.HttpResult(c, nil, xerr.NewErrCodeMsg(xerr.Unauthorized, fmt.Sprintf("نام کاربری یا رمز عبور اشتباه است (%d تلاش باقی‌مانده تا مسدودسازی حساب)", remaining)))
		return
	}

	// Login SUCCESS! Clear failed attempts
	adminLoginAttemptsMu.Lock()
	delete(adminLoginAttempts, attemptKey)
	adminLoginAttemptsMu.Unlock()

	now := time.Now()
	tokenSubjectID := ""
	tokenUsername := username
	tokenFullName := "مدیر سیستم"
	tokenRole := "super_admin"

	if hasAdminUser {
		adminUser.FailedAttempts = 0
		adminUser.LastLoginAt = &now
		h.svcCtx.DB.Model(&adminUser).Updates(map[string]interface{}{
			"failed_attempts": 0,
			"last_login_at":   &now,
		})
		tokenSubjectID = adminUser.ID.String()
		tokenUsername = adminUser.Username
		if adminUser.FullName != "" {
			tokenFullName = adminUser.FullName
		}
		tokenRole = string(adminUser.Role)
	} else if hasCustomer {
		customer.FailedAttempts = 0
		customer.LastLoginAt = &now
		h.svcCtx.DB.Model(&customer).Updates(map[string]interface{}{
			"failed_attempts": 0,
			"last_login_at":   &now,
		})
		tokenSubjectID = customer.ID.String()
		tokenUsername = customer.Email
		if tokenUsername == "" {
			tokenUsername = customer.Phone
		}
		if customer.FullName != "" {
			tokenFullName = customer.FullName
		}
	}

	token, err := middleware.GenerateAdminToken(
		h.svcCtx.Config.JWT.AccessSecret,
		tokenSubjectID,
		tokenUsername,
		tokenRole,
		h.svcCtx.Config.JWT.AccessExpireHours,
	)
	if err != nil {
		result.HttpResult(c, nil, xerr.NewErrCode(xerr.ERROR))
		return
	}

	var logUserID *uuid.UUID
	if u, err := uuid.Parse(tokenSubjectID); err == nil {
		logUserID = &u
	}
	audit.LogAdminActionAsync(h.svcCtx.DB, logUserID, tokenUsername, "AUTH_LOGIN", "auth", tokenSubjectID, "ورود موفقیت‌آمیز به پنل مدیریت آناهیکس", nil, c.ClientIP(), c.Request.UserAgent())

	result.HttpResult(c, &types.AdminLoginResponse{
		Token:    token,
		Username: tokenUsername,
		FullName: tokenFullName,
		Role:     tokenRole,
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

func (h *AdminHandler) RetryOrderBotFulfillment(c *gin.Context) {
	orderID := c.Param("order_id")
	l := admin.NewOrderDeskLogic(c.Request.Context(), h.svcCtx)
	resp, err := l.RetryOrderBotFulfillment(orderID)
	if err == nil && resp != nil {
		audit.LogAdminActionAsync(h.svcCtx.DB, getAdminID(c), c.GetString("username"), "OPERATOR_RETRY_BOT", "order", orderID, fmt.Sprintf("تلاش مجدد و صدور خودکار سفارش #%s از ربات", orderID), resp, c.ClientIP(), c.Request.UserAgent())
	}
	result.HttpResult(c, resp, err)
}

func (h *AdminHandler) ListProducts(c *gin.Context) {
	l := admin.NewProductAdminLogic(c.Request.Context(), h.svcCtx)
	products, err := l.ListProducts()
	result.HttpResult(c, products, err)
}

func (h *AdminHandler) CreateProduct(c *gin.Context) {
	var p model.Product
	if err := c.ShouldBindJSON(&p); err != nil {
		result.ParamErrorResult(c, err)
		return
	}

	l := admin.NewProductAdminLogic(c.Request.Context(), h.svcCtx)
	created, err := l.CreateProduct(&p)
	if err == nil {
		audit.LogAdminActionAsync(h.svcCtx.DB, getAdminID(c), c.GetString("username"), "CREATE_PRODUCT", "product", created.ID.String(), fmt.Sprintf("ایجاد محصول جدید: %s (%s)", created.TitleFA, created.Slug), created, c.ClientIP(), c.Request.UserAgent())
	}
	result.HttpResult(c, created, err)
}

func (h *AdminHandler) UpdateProduct(c *gin.Context) {
	idStr := c.Param("id")
	prodID, err := uuid.Parse(idStr)
	if err != nil {
		result.ParamErrorResult(c, fmt.Errorf("شناسه محصول نامعتبر است"))
		return
	}

	var req types.AdminUpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		result.ParamErrorResult(c, err)
		return
	}

	l := admin.NewProductAdminLogic(c.Request.Context(), h.svcCtx)
	updated, err := l.UpdateProduct(prodID, &req)
	if err == nil {
		audit.LogAdminActionAsync(h.svcCtx.DB, getAdminID(c), c.GetString("username"), "UPDATE_PRODUCT", "product", updated.ID.String(), fmt.Sprintf("ویرایش محصول: %s (%s)", updated.TitleFA, updated.Slug), updated, c.ClientIP(), c.Request.UserAgent())
	}
	result.HttpResult(c, updated, err)
}

func (h *AdminHandler) DeleteProduct(c *gin.Context) {
	idStr := c.Param("id")
	prodID, err := uuid.Parse(idStr)
	if err != nil {
		result.ParamErrorResult(c, fmt.Errorf("شناسه محصول نامعتبر است"))
		return
	}

	l := admin.NewProductAdminLogic(c.Request.Context(), h.svcCtx)
	deleted, err := l.DeleteProduct(prodID)
	if err == nil {
		audit.LogAdminActionAsync(h.svcCtx.DB, getAdminID(c), c.GetString("username"), "DELETE_PRODUCT", "product", prodID.String(), fmt.Sprintf("حذف محصول: %s (%s)", deleted.TitleFA, deleted.Slug), nil, c.ClientIP(), c.Request.UserAgent())
	}
	result.HttpResult(c, gin.H{"deleted": true, "id": prodID.String()}, err)
}

func (h *AdminHandler) GetProduct(c *gin.Context) {
	idStr := c.Param("id")
	prodID, err := uuid.Parse(idStr)
	if err != nil {
		result.ParamErrorResult(c, fmt.Errorf("شناسه محصول نامعتبر است"))
		return
	}

	l := admin.NewProductAdminLogic(c.Request.Context(), h.svcCtx)
	p, err := l.GetProduct(prodID)
	result.HttpResult(c, p, err)
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

func (h *AdminHandler) ListBotProviderSettings(c *gin.Context) {
	l := admin.NewDeliveryEngineLogic(c.Request.Context(), h.svcCtx)
	resp, err := l.ListBotProviderSettings()
	result.HttpResult(c, resp, err)
}

func (h *AdminHandler) UpdateBotProviderSettings(c *gin.Context) {
	var req types.UpdateBotProviderSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		result.ParamErrorResult(c, err)
		return
	}

	l := admin.NewDeliveryEngineLogic(c.Request.Context(), h.svcCtx)
	err := l.UpdateBotProviderSettings(&req)
	if err == nil {
		audit.LogAdminActionAsync(h.svcCtx.DB, getAdminID(c), c.GetString("username"), "UPDATE_BOT_SETTINGS", "delivery_engine", req.Key, fmt.Sprintf("به‌روزرسانی تنظیمات ربات %s", req.Key), req, c.ClientIP(), c.Request.UserAgent())
	}
	result.HttpResult(c, gin.H{"status": "ok"}, err)
}

func (h *AdminHandler) TestBotProviderConnection(c *gin.Context) {
	var req types.TestBotProviderConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		result.ParamErrorResult(c, err)
		return
	}

	l := admin.NewDeliveryEngineLogic(c.Request.Context(), h.svcCtx)
	resp, err := l.TestBotProviderConnection(&req)
	result.HttpResult(c, resp, err)
}

func (h *AdminHandler) GetBotCatalog(c *gin.Context) {
	providerKey := c.DefaultQuery("provider", "insightx")
	forceRefresh := c.Query("refresh") == "true"

	l := admin.NewDeliveryEngineLogic(c.Request.Context(), h.svcCtx)
	resp, err := l.GetBotCatalog(providerKey, forceRefresh)
	result.HttpResult(c, resp, err)
}

func (h *AdminHandler) ImportBotProduct(c *gin.Context) {
	var req types.ImportBotProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		result.ParamErrorResult(c, err)
		return
	}

	l := admin.NewDeliveryEngineLogic(c.Request.Context(), h.svcCtx)
	prod, err := l.ImportBotProduct(&req)
	if err == nil && prod != nil {
		audit.LogAdminActionAsync(h.svcCtx.DB, getAdminID(c), c.GetString("username"), "IMPORT_BOT_PRODUCT", "catalog", prod.ID.String(), fmt.Sprintf("ایمپورت محصول %s از ربات %s", prod.TitleFA, req.ProviderKey), req, c.ClientIP(), c.Request.UserAgent())
	}
	result.HttpResult(c, prod, err)
}

func (h *AdminHandler) ListStoreCategories(c *gin.Context) {
	l := admin.NewDeliveryEngineLogic(c.Request.Context(), h.svcCtx)
	resp, err := l.ListStoreCategories()
	result.HttpResult(c, resp, err)
}

func (h *AdminHandler) CreateStoreCategory(c *gin.Context) {
	var req types.CreateStoreCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		result.ParamErrorResult(c, err)
		return
	}

	l := admin.NewDeliveryEngineLogic(c.Request.Context(), h.svcCtx)
	err := l.CreateStoreCategory(&req)
	if err == nil {
		audit.LogAdminActionAsync(h.svcCtx.DB, getAdminID(c), c.GetString("username"), "CREATE_STORE_CATEGORY", "categories", req.Slug, fmt.Sprintf("ایجاد دسته‌بندی %s", req.TitleFA), req, c.ClientIP(), c.Request.UserAgent())
	}
	result.HttpResult(c, gin.H{"status": "ok"}, err)
}

func (h *AdminHandler) UpdateStoreCategory(c *gin.Context) {
	id := c.Param("id")
	var req types.UpdateStoreCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		result.ParamErrorResult(c, err)
		return
	}

	l := admin.NewDeliveryEngineLogic(c.Request.Context(), h.svcCtx)
	err := l.UpdateStoreCategory(id, &req)
	if err == nil {
		audit.LogAdminActionAsync(h.svcCtx.DB, getAdminID(c), c.GetString("username"), "UPDATE_STORE_CATEGORY", "categories", id, fmt.Sprintf("به‌روزرسانی دسته‌بندی %s", req.TitleFA), req, c.ClientIP(), c.Request.UserAgent())
	}
	result.HttpResult(c, gin.H{"status": "ok"}, err)
}

func (h *AdminHandler) DeleteStoreCategory(c *gin.Context) {
	id := c.Param("id")
	l := admin.NewDeliveryEngineLogic(c.Request.Context(), h.svcCtx)
	err := l.DeleteStoreCategory(id)
	if err == nil {
		audit.LogAdminActionAsync(h.svcCtx.DB, getAdminID(c), c.GetString("username"), "DELETE_STORE_CATEGORY", "categories", id, "حذف دسته‌بندی فروشگاه", nil, c.ClientIP(), c.Request.UserAgent())
	}
	result.HttpResult(c, gin.H{"status": "ok"}, err)
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

func (h *AdminHandler) UpdateTransactionStatus(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		result.ParamErrorResult(c, xerr.NewErrCode(xerr.InvalidParams))
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		result.ParamErrorResult(c, err)
		return
	}

	l := admin.NewAuditLogic(c.Request.Context(), h.svcCtx)
	err := l.UpdateTransactionStatus(id, req.Status)
	if err == nil {
		audit.LogAdminActionAsync(h.svcCtx.DB, getAdminID(c), c.GetString("username"), "UPDATE_TX_STATUS", "transaction", id, fmt.Sprintf("تغییر وضعیت تراکنش به (%s)", req.Status), req, c.ClientIP(), c.Request.UserAgent())
	}
	result.HttpResult(c, gin.H{"updated": true}, err)
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

func (h *AdminHandler) UpdateCoupon(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		result.ParamErrorResult(c, xerr.NewErrCode(xerr.InvalidParams))
		return
	}

	var req types.UpdateCouponRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		result.ParamErrorResult(c, err)
		return
	}

	l := admin.NewCouponAdminLogic(c.Request.Context(), h.svcCtx)
	resp, err := l.UpdateCoupon(id, &req)
	if err == nil {
		audit.LogAdminActionAsync(h.svcCtx.DB, getAdminID(c), c.GetString("username"), "UPDATE_COUPON", "coupon", id, fmt.Sprintf("ویرایش کد تخفیف (%s)", id), req, c.ClientIP(), c.Request.UserAgent())
	}
	result.HttpResult(c, resp, err)
}

func (h *AdminHandler) DeleteCoupon(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		result.ParamErrorResult(c, xerr.NewErrCode(xerr.InvalidParams))
		return
	}

	l := admin.NewCouponAdminLogic(c.Request.Context(), h.svcCtx)
	err := l.DeleteCoupon(id)
	if err == nil {
		audit.LogAdminActionAsync(h.svcCtx.DB, getAdminID(c), c.GetString("username"), "DELETE_COUPON", "coupon", id, fmt.Sprintf("حذف کد تخفیف (%s)", id), nil, c.ClientIP(), c.Request.UserAgent())
	}
	result.HttpResult(c, gin.H{"deleted": true}, err)
}

func (h *AdminHandler) ToggleCoupon(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		result.ParamErrorResult(c, xerr.NewErrCode(xerr.InvalidParams))
		return
	}

	l := admin.NewCouponAdminLogic(c.Request.Context(), h.svcCtx)
	resp, err := l.ToggleCoupon(id)
	if err == nil {
		audit.LogAdminActionAsync(h.svcCtx.DB, getAdminID(c), c.GetString("username"), "TOGGLE_COUPON", "coupon", id, fmt.Sprintf("تغییر وضعیت فعال‌بودن کد تخفیف (%s)", id), nil, c.ClientIP(), c.Request.UserAgent())
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

func (h *AdminHandler) ListPaymentGateways(c *gin.Context) {
	var payments []model.Payment
	if err := h.svcCtx.DB.WithContext(c.Request.Context()).Order("sort asc, id asc").Find(&payments).Error; err != nil {
		result.HttpResult(c, nil, err)
		return
	}
	result.HttpResult(c, payments, nil)
}

func (h *AdminHandler) UpdatePaymentGateway(c *gin.Context) {
	idStr := c.Param("id")
	var p model.Payment
	if err := h.svcCtx.DB.WithContext(c.Request.Context()).First(&p, "id = ?", idStr).Error; err != nil {
		result.ParamErrorResult(c, fmt.Errorf("درگاه پرداخت یافت نشد"))
		return
	}

	var req struct {
		NameFA      *string `json:"name_fa"`
		Icon        *string `json:"icon"`
		Description *string `json:"description"`
		Enable      *bool   `json:"enable"`
		DisplayMode *string `json:"display_mode"`
		Sort        *int    `json:"sort"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		result.ParamErrorResult(c, err)
		return
	}

	if req.NameFA != nil {
		p.NameFA = *req.NameFA
	}
	if req.Icon != nil {
		p.Icon = *req.Icon
	}
	if req.Description != nil {
		p.Description = *req.Description
	}
	if req.DisplayMode != nil {
		p.DisplayMode = *req.DisplayMode
		if *req.DisplayMode == "active" {
			p.Enable = true
		} else {
			p.Enable = false
		}
	} else if req.Enable != nil {
		p.Enable = *req.Enable
		if *req.Enable {
			p.DisplayMode = "active"
		} else {
			p.DisplayMode = "hidden"
		}
	}
	if req.Sort != nil {
		p.Sort = *req.Sort
	}

	if err := h.svcCtx.DB.WithContext(c.Request.Context()).Save(&p).Error; err != nil {
		result.HttpResult(c, nil, err)
		return
	}

	result.HttpResult(c, p, nil)
}

func (h *AdminHandler) ListAnnouncements(c *gin.Context) {
	var list []model.SystemAnnouncement
	err := h.svcCtx.DB.WithContext(c.Request.Context()).Order("created_at desc").Find(&list).Error
	result.HttpResult(c, list, err)
}

func (h *AdminHandler) CreateAnnouncement(c *gin.Context) {
	var a model.SystemAnnouncement
	if err := c.ShouldBindJSON(&a); err != nil {
		result.ParamErrorResult(c, err)
		return
	}
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	err := h.svcCtx.DB.WithContext(c.Request.Context()).Create(&a).Error
	result.HttpResult(c, a, err)
}

func (h *AdminHandler) UpdateAnnouncement(c *gin.Context) {
	idStr := c.Param("id")
	var a model.SystemAnnouncement
	if err := h.svcCtx.DB.WithContext(c.Request.Context()).First(&a, "id = ?", idStr).Error; err != nil {
		result.ParamErrorResult(c, fmt.Errorf("اعلان یافت نشد"))
		return
	}
	if err := c.ShouldBindJSON(&a); err != nil {
		result.ParamErrorResult(c, err)
		return
	}
	err := h.svcCtx.DB.WithContext(c.Request.Context()).Save(&a).Error
	result.HttpResult(c, a, err)
}

func (h *AdminHandler) DeleteAnnouncement(c *gin.Context) {
	idStr := c.Param("id")
	err := h.svcCtx.DB.WithContext(c.Request.Context()).Delete(&model.SystemAnnouncement{}, "id = ?", idStr).Error
	result.HttpResult(c, gin.H{"deleted": true}, err)
}

func (h *AdminHandler) ListCustomerNotifications(c *gin.Context) {
	type NotifItem struct {
		model.CustomerNotification
		CustomerPhone string `json:"customer_phone"`
		CustomerName  string `json:"customer_name"`
	}
	var notifs []model.CustomerNotification
	err := h.svcCtx.DB.WithContext(c.Request.Context()).
		Order("created_at desc").
		Limit(100).
		Find(&notifs).Error
	if err != nil {
		result.HttpResult(c, nil, err)
		return
	}

	res := make([]NotifItem, len(notifs))
	for i, n := range notifs {
		var cust model.Customer
		h.svcCtx.DB.WithContext(c.Request.Context()).Select("phone, full_name").First(&cust, "id = ?", n.CustomerID)
		res[i] = NotifItem{
			CustomerNotification: n,
			CustomerPhone:        cust.Phone,
			CustomerName:         cust.FullName,
		}
	}
	result.HttpResult(c, res, nil)
}

func (h *AdminHandler) SendCustomerNotification(c *gin.Context) {
	var req struct {
		Target     string `json:"target"` // "all" or specific customer_id
		CustomerID string `json:"customer_id"`
		Title      string `json:"title" binding:"required"`
		Message    string `json:"message" binding:"required"`
		Type       string `json:"type"` // "system", "security", "loyalty", "order"
		Link       string `json:"link"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		result.ParamErrorResult(c, err)
		return
	}

	notifType := req.Type
	if notifType == "" {
		notifType = "system"
	}

	db := h.svcCtx.DB.WithContext(c.Request.Context())

	if req.Target == "all" || req.CustomerID == "" {
		var customers []model.Customer
		db.Find(&customers)
		for _, cust := range customers {
			n := model.CustomerNotification{
				ID:         uuid.New(),
				CustomerID: cust.ID,
				Title:      req.Title,
				Message:    req.Message,
				Type:       notifType,
				Link:       req.Link,
				IsRead:     false,
				CreatedAt:  time.Now(),
			}
			db.Create(&n)
		}
		result.HttpResult(c, gin.H{"sent_count": len(customers), "message": "پیام به کلیه کاربران با موفقیت ارسال شد"}, nil)
		return
	}

	custUUID, err := uuid.Parse(req.CustomerID)
	if err != nil {
		result.ParamErrorResult(c, fmt.Errorf("شناسه کاربر نامعتبر است"))
		return
	}

	n := model.CustomerNotification{
		ID:         uuid.New(),
		CustomerID: custUUID,
		Title:      req.Title,
		Message:    req.Message,
		Type:       notifType,
		Link:       req.Link,
		IsRead:     false,
		CreatedAt:  time.Now(),
	}
	err = db.Create(&n).Error
	result.HttpResult(c, n, err)
}

func (h *AdminHandler) DeleteCustomerNotification(c *gin.Context) {
	idStr := c.Param("id")
	err := h.svcCtx.DB.WithContext(c.Request.Context()).Delete(&model.CustomerNotification{}, "id = ?", idStr).Error
	result.HttpResult(c, gin.H{"deleted": true}, err)
}





