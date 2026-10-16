package admin

import (
	"github.com/EbadiDev/anahix-server/internal/logic/admin"
	"github.com/EbadiDev/anahix-server/internal/svc"
	"github.com/EbadiDev/anahix-server/internal/types"
	"github.com/EbadiDev/anahix-server/pkg/result"
	"github.com/gin-gonic/gin"
)

type LoyaltyAdminHandler struct {
	svcCtx *svc.ServiceContext
}

func NewLoyaltyAdminHandler(svcCtx *svc.ServiceContext) *LoyaltyAdminHandler {
	return &LoyaltyAdminHandler{svcCtx: svcCtx}
}

// ListTiers godoc
// @Summary      List all loyalty tiers with customer counts
// @Tags         Admin Loyalty
// @Security     BearerAuth
// @Router       /admin/loyalty/tiers [get]
func (h *LoyaltyAdminHandler) ListTiers(c *gin.Context) {
	l := admin.NewLoyaltyAdminLogic(c.Request.Context(), h.svcCtx)
	resp, err := l.ListTiers()
	result.HttpResult(c, resp, err)
}

// CreateTier godoc
// @Summary      Create a new dynamic loyalty tier
// @Tags         Admin Loyalty
// @Security     BearerAuth
// @Router       /admin/loyalty/tiers [post]
func (h *LoyaltyAdminHandler) CreateTier(c *gin.Context) {
	var req types.CreateOrUpdateLoyaltyTierRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		result.ParamErrorResult(c, err)
		return
	}
	l := admin.NewLoyaltyAdminLogic(c.Request.Context(), h.svcCtx)
	resp, err := l.CreateTier(&req)
	result.HttpResult(c, resp, err)
}

// UpdateTier godoc
// @Summary      Update an existing loyalty tier
// @Tags         Admin Loyalty
// @Security     BearerAuth
// @Router       /admin/loyalty/tiers/:id [put]
func (h *LoyaltyAdminHandler) UpdateTier(c *gin.Context) {
	id := c.Param("id")
	var req types.CreateOrUpdateLoyaltyTierRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		result.ParamErrorResult(c, err)
		return
	}
	l := admin.NewLoyaltyAdminLogic(c.Request.Context(), h.svcCtx)
	resp, err := l.UpdateTier(id, &req)
	result.HttpResult(c, resp, err)
}

// DeleteTier godoc
// @Summary      Delete a loyalty tier
// @Tags         Admin Loyalty
// @Security     BearerAuth
// @Router       /admin/loyalty/tiers/:id [delete]
func (h *LoyaltyAdminHandler) DeleteTier(c *gin.Context) {
	id := c.Param("id")
	l := admin.NewLoyaltyAdminLogic(c.Request.Context(), h.svcCtx)
	err := l.DeleteTier(id)
	result.HttpResult(c, gin.H{"deleted": true}, err)
}

// GetSettings godoc
// @Summary      Get system loyalty settings
// @Tags         Admin Loyalty
// @Security     BearerAuth
// @Router       /admin/loyalty/settings [get]
func (h *LoyaltyAdminHandler) GetSettings(c *gin.Context) {
	l := admin.NewLoyaltyAdminLogic(c.Request.Context(), h.svcCtx)
	resp, err := l.GetLoyaltySettings()
	result.HttpResult(c, resp, err)
}

// UpdateSettings godoc
// @Summary      Update system loyalty settings
// @Tags         Admin Loyalty
// @Security     BearerAuth
// @Router       /admin/loyalty/settings [post]
func (h *LoyaltyAdminHandler) UpdateSettings(c *gin.Context) {
	var body map[string]string
	if err := c.ShouldBindJSON(&body); err != nil {
		result.ParamErrorResult(c, err)
		return
	}
	l := admin.NewLoyaltyAdminLogic(c.Request.Context(), h.svcCtx)
	err := l.UpdateLoyaltySettings(body)
	result.HttpResult(c, gin.H{"updated": true}, err)
}
