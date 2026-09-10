package admin

import (
	"github.com/EbadiDev/anahix-server/internal/logic/admin"
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
	if err != nil || !user.CheckPassword(req.Password) {
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
	result.HttpResult(c, gin.H{"status": "two_factor_requested"}, err)
}

func (h *AdminHandler) CompleteOrder(c *gin.Context) {
	orderID := c.Param("order_id")
	var req types.CompleteOrderRequest
	_ = c.ShouldBindJSON(&req)

	l := admin.NewOrderDeskLogic(c.Request.Context(), h.svcCtx)
	err := l.CompleteOrder(orderID, &req)
	result.HttpResult(c, gin.H{"status": "completed"}, err)
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

	if err := h.svcCtx.DB.WithContext(c.Request.Context()).Create(&p).Error; err != nil {
		result.HttpResult(c, nil, xerr.NewErrCode(xerr.DatabaseError))
		return
	}

	result.HttpResult(c, p, nil)
}
