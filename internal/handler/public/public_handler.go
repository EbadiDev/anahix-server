package public

import (
	"github.com/EbadiDev/anahix-server/internal/logic/public"
	"github.com/EbadiDev/anahix-server/internal/svc"
	"github.com/EbadiDev/anahix-server/internal/types"
	"github.com/EbadiDev/anahix-server/pkg/result"
	"github.com/EbadiDev/anahix-server/pkg/xerr"
	"github.com/gin-gonic/gin"
)

type PublicHandler struct {
	svcCtx *svc.ServiceContext
}

func NewPublicHandler(svcCtx *svc.ServiceContext) *PublicHandler {
	return &PublicHandler{svcCtx: svcCtx}
}

func (h *PublicHandler) ListProducts(c *gin.Context) {
	l := public.NewCatalogLogic(c.Request.Context(), h.svcCtx)
	resp, err := l.ListProducts()
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

	if orderID == "" || token == "" {
		result.ParamErrorResult(c, xerr.NewErrCode(xerr.InvalidAccessToken))
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

	l := public.NewOrderLogic(c.Request.Context(), h.svcCtx)
	err := l.SubmitTwoFactor(orderID, token, req.TwoFactorCode)
	result.HttpResult(c, gin.H{"status": "submitted"}, err)
}

func (h *PublicHandler) ListPaymentMethods(c *gin.Context) {
	l := public.NewPaymentLogic(c.Request.Context(), h.svcCtx)
	resp, err := l.ListPaymentMethods()
	result.HttpResult(c, resp, err)
}
