package public

import (
	"fmt"
	"strconv"
	"time"

	"github.com/EbadiDev/anahix-server/internal/logic/public"
	"github.com/EbadiDev/anahix-server/internal/model"
	"github.com/EbadiDev/anahix-server/internal/svc"
	"github.com/EbadiDev/anahix-server/internal/types"
	"github.com/EbadiDev/anahix-server/pkg/result"
	"github.com/EbadiDev/anahix-server/pkg/xerr"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type PublicHandler struct {
	svcCtx *svc.ServiceContext
}

func NewPublicHandler(svcCtx *svc.ServiceContext) *PublicHandler {
	return &PublicHandler{svcCtx: svcCtx}
}

func (h *PublicHandler) ListProducts(c *gin.Context) {
	q := c.Query("search")
	l := public.NewCatalogLogic(c.Request.Context(), h.svcCtx)
	resp, err := l.ListProducts(q)
	result.HttpResult(c, resp, err)
}

func (h *PublicHandler) GetProductDetail(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		result.ParamErrorResult(c, xerr.NewErrCode(xerr.InvalidParams))
		return
	}

	l := public.NewCatalogLogic(c.Request.Context(), h.svcCtx)
	resp, err := l.GetProductDetail(slug)
	result.HttpResult(c, resp, err)
}

func (h *PublicHandler) CreateOrder(c *gin.Context) {
	var req types.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		result.ParamErrorResult(c, err)
		return
	}

	l := public.NewOrderLogic(c.Request.Context(), h.svcCtx)
	resp, err := l.CreateOrder(&req)
	result.HttpResult(c, resp, err)
}

func (h *PublicHandler) GetOrderDetail(c *gin.Context) {
	orderID := c.Param("order_id")
	token := c.Query("token")
	if token == "" {
		token = c.GetHeader("X-Order-Token")
	}

	if orderID == "" {
		result.ParamErrorResult(c, xerr.NewErrCode(xerr.InvalidParams))
		return
	}

	l := public.NewOrderLogic(c.Request.Context(), h.svcCtx)
	resp, err := l.GetOrderDetail(orderID, token)
	result.HttpResult(c, resp, err)
}

func (h *PublicHandler) SubmitTwoFactor(c *gin.Context) {
	orderID := c.Param("order_id")
	token := c.Query("token")
	if token == "" {
		token = c.GetHeader("X-Order-Token")
	}

	var req types.SubmitTwoFactorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		result.ParamErrorResult(c, err)
		return
	}

	code := req.TwoFactorCode
	if code == "" {
		code = req.Code
	}

	l := public.NewOrderLogic(c.Request.Context(), h.svcCtx)
	err := l.SubmitTwoFactor(orderID, token, code)
	result.HttpResult(c, gin.H{"status": "submitted"}, err)
}

func (h *PublicHandler) ListPaymentMethods(c *gin.Context) {
	l := public.NewPaymentLogic(c.Request.Context(), h.svcCtx)
	resp, err := l.ListPaymentMethods()
	result.HttpResult(c, resp, err)
}

func (h *PublicHandler) ValidateCoupon(c *gin.Context) {
	var req types.ValidateCouponRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		result.ParamErrorResult(c, err)
		return
	}

	l := public.NewCouponLogic(c.Request.Context(), h.svcCtx)
	resp, err := l.ValidateCoupon(&req)
	result.HttpResult(c, resp, err)
}

func (h *PublicHandler) ListPublicCoupons(c *gin.Context) {
	l := public.NewCouponLogic(c.Request.Context(), h.svcCtx)
	resp, err := l.ListPublicPromotions()
	result.HttpResult(c, resp, err)
}

func (h *PublicHandler) GetReferralStats(c *gin.Context) {
	phone := c.GetString("phone")
	if phone == "" {
		phone = c.Query("phone")
	}

	l := public.NewReferralLogic(c.Request.Context(), h.svcCtx)
	resp, err := l.GetReferralStats(phone)
	result.HttpResult(c, resp, err)
}

func (h *PublicHandler) TransferReferralCommission(c *gin.Context) {
	phone := c.GetString("phone")
	if phone == "" {
		result.ParamErrorResult(c, xerr.NewErrMsg("احراز هویت الزامی است"))
		return
	}

	l := public.NewReferralLogic(c.Request.Context(), h.svcCtx)
	resp, err := l.TransferToWallet(phone)
	result.HttpResult(c, resp, err)
}

func (h *PublicHandler) GetWalletLedger(c *gin.Context) {
	phone := c.GetString("phone")
	if phone == "" {
		phone = c.Query("phone")
	}

	if phone == "" {
		result.ParamErrorResult(c, xerr.NewErrMsg("شماره همراه معتبر نیست"))
		return
	}

	l := public.NewWalletLogic(c.Request.Context(), h.svcCtx)
	resp, err := l.GetCustomerLedger(phone)
	result.HttpResult(c, resp, err)
}

// Review Endpoints

func (h *PublicHandler) GetProductReviews(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		result.ParamErrorResult(c, xerr.NewErrCode(xerr.InvalidParams))
		return
	}

	ratingFilter := 0
	if rStr := c.Query("rating"); rStr != "" {
		if r, err := strconv.Atoi(rStr); err == nil {
			ratingFilter = r
		}
	}

	sortOrder := c.DefaultQuery("sort", "newest")
	onlyWithComments := c.Query("has_comment") == "true"
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	l := public.NewReviewLogic(c.Request.Context(), h.svcCtx)
	resp, err := l.GetProductReviews(slug, ratingFilter, sortOrder, onlyWithComments, page, pageSize)
	result.HttpResult(c, resp, err)
}

func (h *PublicHandler) SubmitReview(c *gin.Context) {
	var req types.SubmitReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		result.ParamErrorResult(c, err)
		return
	}

	customerID := c.GetString("customer_id")
	customerMobile := c.GetString("mobile")
	if customerMobile == "" {
		customerMobile = c.GetString("phone")
	}
	clientIP := c.ClientIP()

	l := public.NewReviewLogic(c.Request.Context(), h.svcCtx)
	resp, err := l.SubmitReview(&req, customerID, customerMobile, clientIP)
	result.HttpResult(c, resp, err)
}

func (h *PublicHandler) VoteReviewHelpful(c *gin.Context) {
	reviewID := c.Param("id")
	if reviewID == "" {
		result.ParamErrorResult(c, xerr.NewErrCode(xerr.InvalidParams))
		return
	}

	voterKey := c.GetString("customer_id")
	if voterKey == "" {
		voterKey = c.GetString("mobile")
	}
	if voterKey == "" {
		voterKey = c.ClientIP()
	}

	l := public.NewReviewLogic(c.Request.Context(), h.svcCtx)
	newCount, err := l.VoteHelpful(reviewID, voterKey)
	result.HttpResult(c, gin.H{"helpful_count": newCount}, err)
}

func (h *PublicHandler) CheckReviewEligibility(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		result.ParamErrorResult(c, xerr.NewErrCode(xerr.InvalidParams))
		return
	}

	customerID := c.GetString("customer_id")
	customerMobile := c.GetString("mobile")
	if customerMobile == "" {
		customerMobile = c.GetString("phone")
	}

	l := public.NewReviewLogic(c.Request.Context(), h.svcCtx)
	resp, err := l.CheckEligibility(slug, customerID, customerMobile)
	result.HttpResult(c, resp, err)
}

func (h *PublicHandler) GetMyReviews(c *gin.Context) {
	customerID := c.GetString("customer_id")
	customerMobile := c.GetString("mobile")
	if customerMobile == "" {
		customerMobile = c.GetString("phone")
	}

	l := public.NewReviewLogic(c.Request.Context(), h.svcCtx)
	resp, err := l.GetCustomerReviews(customerID, customerMobile)
	result.HttpResult(c, resp, err)
}

func (h *PublicHandler) GetOrderInvoice(c *gin.Context) {
	orderIDOrNo := c.Param("order_id")
	token := c.Query("token")

	var currentCustUUID *uuid.UUID
	if cidVal, exists := c.Get("customer_id"); exists {
		if cidStr, ok := cidVal.(string); ok && cidStr != "" {
			if u, err := uuid.Parse(cidStr); err == nil {
				currentCustUUID = &u
			}
		}
	}

	l := public.NewInvoiceLogic(c.Request.Context(), h.svcCtx)
	resp, err := l.GetOrderInvoice(orderIDOrNo, token, currentCustUUID, false)
	result.HttpResult(c, resp, err)
}

func (h *PublicHandler) UpdateOrderBuyerInvoice(c *gin.Context) {
	orderIDOrNo := c.Param("order_id")
	token := c.Query("token")

	var req types.UpdateBuyerInvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		result.ParamErrorResult(c, err)
		return
	}

	var currentCustUUID *uuid.UUID
	if cidVal, exists := c.Get("customer_id"); exists {
		if cidStr, ok := cidVal.(string); ok && cidStr != "" {
			if u, err := uuid.Parse(cidStr); err == nil {
				currentCustUUID = &u
			}
		}
	}

	l := public.NewInvoiceLogic(c.Request.Context(), h.svcCtx)
	err := l.UpdateOrderBuyerInvoice(orderIDOrNo, token, currentCustUUID, false, &req)
	result.HttpResult(c, gin.H{"status": "ok", "message": "اطلاعات خریدار فاکتور رسمی با موفقیت به‌روزرسانی شد"}, err)
}

func (h *PublicHandler) ListStoreCategories(c *gin.Context) {
	var categories []model.StoreCategory
	err := h.svcCtx.DB.WithContext(c.Request.Context()).
		Preload("Children").
		Where("parent_id IS NULL AND is_active = ?", true).
		Order("sort_order asc, created_at asc").
		Find(&categories).Error
	result.HttpResult(c, categories, err)
}

func (h *PublicHandler) ListAnnouncements(c *gin.Context) {
	now := time.Now()
	var list []model.SystemAnnouncement
	err := h.svcCtx.DB.WithContext(c.Request.Context()).
		Where("is_active = ? AND (expires_at IS NULL OR expires_at > ?)", true, now).
		Order("created_at desc").
		Find(&list).Error
	result.HttpResult(c, list, err)
}

func (h *PublicHandler) CreateStockAlert(c *gin.Context) {
	var req struct {
		ProductID string `json:"product_id" binding:"required"`
		Mobile    string `json:"mobile"`
		Email     string `json:"email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		result.ParamErrorResult(c, err)
		return
	}
	alert := model.StockAlert{
		ID:        uuid.New(),
		ProductID: req.ProductID,
		Mobile:    req.Mobile,
		Email:     req.Email,
		Notified:  false,
	}
	err := h.svcCtx.DB.WithContext(c.Request.Context()).Create(&alert).Error
	result.HttpResult(c, gin.H{"success": true, "message": "درخواست اطلاع‌رسانی با موفقیت ثبت شد."}, err)
}

func (h *PublicHandler) ToggleFavorite(c *gin.Context) {
	cidStr := c.GetString("customer_id")
	custID, err := uuid.Parse(cidStr)
	if err != nil {
		result.ParamErrorResult(c, fmt.Errorf("کاربر احراز هویت نشده است"))
		return
	}

	var req struct {
		TargetType string `json:"target_type" binding:"required"`
		TargetID   string `json:"target_id" binding:"required"`
		Title      string `json:"title"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		result.ParamErrorResult(c, err)
		return
	}

	var existing model.UserFavorite
	db := h.svcCtx.DB.WithContext(c.Request.Context())
	findErr := db.Where("customer_id = ? AND target_type = ? AND target_id = ?", custID, req.TargetType, req.TargetID).First(&existing).Error
	if findErr == nil {
		// Already favorited, remove it
		db.Delete(&existing)
		result.HttpResult(c, gin.H{"is_favorite": false, "message": "از نشان‌شده‌ها حذف شد"}, nil)
		return
	}

	// Add new favorite
	fav := model.UserFavorite{
		ID:         uuid.New(),
		CustomerID: custID,
		TargetType: req.TargetType,
		TargetID:   req.TargetID,
		Title:      req.Title,
	}
	createErr := db.Create(&fav).Error
	result.HttpResult(c, gin.H{"is_favorite": true, "message": "به نشان‌شده‌ها اضافه شد"}, createErr)
}

func (h *PublicHandler) ListFavorites(c *gin.Context) {
	cidStr := c.GetString("customer_id")
	custID, err := uuid.Parse(cidStr)
	if err != nil {
		result.ParamErrorResult(c, fmt.Errorf("کاربر احراز هویت نشده است"))
		return
	}

	var list []model.UserFavorite
	db := h.svcCtx.DB.WithContext(c.Request.Context())
	err = db.Where("customer_id = ? AND target_type IN ('article', 'guide')", custID).Order("created_at desc").Find(&list).Error
	result.HttpResult(c, list, err)
}


