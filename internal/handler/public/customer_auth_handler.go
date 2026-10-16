package public

import (
	"github.com/EbadiDev/anahix-server/internal/logic/public"
	"github.com/EbadiDev/anahix-server/internal/svc"
	"github.com/EbadiDev/anahix-server/internal/types"
	"github.com/EbadiDev/anahix-server/pkg/result"
	"github.com/EbadiDev/anahix-server/pkg/xerr"
	"github.com/gin-gonic/gin"
)

type CustomerAuthHandler struct {
	svcCtx *svc.ServiceContext
}

func NewCustomerAuthHandler(svcCtx *svc.ServiceContext) *CustomerAuthHandler {
	return &CustomerAuthHandler{svcCtx: svcCtx}
}

// SendOTP triggers SMS.ir verification code dispatch
func (h *CustomerAuthHandler) SendOTP(c *gin.Context) {
	var req types.SendOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		result.ParamErrorResult(c, err)
		return
	}

	l := public.NewCustomerAuthLogic(c.Request.Context(), h.svcCtx)
	resp, err := l.SendOTP(&req)
	result.HttpResult(c, resp, err)
}

// VerifyOTP validates the OTP code and creates or logs in the customer
func (h *CustomerAuthHandler) VerifyOTP(c *gin.Context) {
	var req types.VerifyOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		result.ParamErrorResult(c, err)
		return
	}

	l := public.NewCustomerAuthLogic(c.Request.Context(), h.svcCtx)
	resp, err := l.VerifyOTP(&req)
	result.HttpResult(c, resp, err)
}

// GetMe returns the authenticated customer profile
func (h *CustomerAuthHandler) GetMe(c *gin.Context) {
	customerID := c.GetString("customer_id")
	if customerID == "" {
		result.ParamErrorResult(c, xerr.NewErrCode(xerr.Unauthorized))
		return
	}

	l := public.NewCustomerAuthLogic(c.Request.Context(), h.svcCtx)
	resp, err := l.GetMe(customerID)
	result.HttpResult(c, resp, err)
}

// UpdateMe modifies customer profile info
func (h *CustomerAuthHandler) UpdateMe(c *gin.Context) {
	customerID := c.GetString("customer_id")
	if customerID == "" {
		result.ParamErrorResult(c, xerr.NewErrCode(xerr.Unauthorized))
		return
	}

	var req types.UpdateCustomerProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		result.ParamErrorResult(c, err)
		return
	}

	l := public.NewCustomerAuthLogic(c.Request.Context(), h.svcCtx)
	resp, err := l.UpdateProfile(customerID, &req)
	result.HttpResult(c, resp, err)
}

// LoginWithPassword allows customer login via phone/email and password
func (h *CustomerAuthHandler) LoginWithPassword(c *gin.Context) {
	var req types.PasswordLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		result.ParamErrorResult(c, err)
		return
	}

	l := public.NewCustomerAuthLogic(c.Request.Context(), h.svcCtx)
	resp, err := l.LoginWithPassword(&req)
	result.HttpResult(c, resp, err)
}

// OAuthSimulate handles Google and GitHub social logins and profile synchronization
func (h *CustomerAuthHandler) OAuthSimulate(c *gin.Context) {
	var req types.OAuthSimulateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		result.ParamErrorResult(c, err)
		return
	}

	l := public.NewCustomerAuthLogic(c.Request.Context(), h.svcCtx)
	resp, err := l.OAuthSimulate(&req)
	result.HttpResult(c, resp, err)
}

// SetPassword sets or changes password for authenticated customer
func (h *CustomerAuthHandler) SetPassword(c *gin.Context) {
	customerID := c.GetString("customer_id")
	if customerID == "" {
		result.ParamErrorResult(c, xerr.NewErrCode(xerr.Unauthorized))
		return
	}

	var req types.SetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		result.ParamErrorResult(c, err)
		return
	}

	l := public.NewCustomerAuthLogic(c.Request.Context(), h.svcCtx)
	err := l.SetPassword(customerID, &req)
	result.HttpResult(c, gin.H{"updated": true}, err)
}
