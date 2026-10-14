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

func (h *PublicHandler) GetReferralStats(c *gin.Context) {
	phone := c.GetString("phone")
	if phone == "" {
		phone = c.Query("phone")
	}

	if phone == "" {
		result.ParamErrorResult(c, xerr.NewErrMsg("شماره همراه معتبر نیست"))
		return
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

