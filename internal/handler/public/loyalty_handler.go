package public

import (
	"strings"

	"github.com/EbadiDev/anahix-server/internal/logic/public"
	"github.com/EbadiDev/anahix-server/internal/svc"
	"github.com/EbadiDev/anahix-server/pkg/result"
	"github.com/gin-gonic/gin"
)

type LoyaltyHandler struct {
	svcCtx *svc.ServiceContext
}

func NewLoyaltyHandler(svcCtx *svc.ServiceContext) *LoyaltyHandler {
	return &LoyaltyHandler{svcCtx: svcCtx}
}

// GetMyLoyaltyProfile godoc
// @Summary      Get user's VIP loyalty profile, current perks, and next tier progression
// @Tags         Customer Loyalty
// @Security     BearerAuth
// @Router       /loyalty/profile [get]
func (h *LoyaltyHandler) GetMyLoyaltyProfile(c *gin.Context) {
	phone := c.Query("phone")
	if phone == "" {
		// Try resolving phone from authenticated customer claims if available
		if val, exists := c.Get("claims"); exists {
			if claims, ok := val.(map[string]interface{}); ok {
				if p, ok := claims["phone"].(string); ok {
					phone = p
				}
			}
		}
	}

	phone = strings.TrimSpace(phone)

	l := public.NewLoyaltyPublicLogic(c.Request.Context(), h.svcCtx)
	resp, err := l.GetCustomerLoyaltyProfile(phone)
	result.HttpResult(c, resp, err)
}

// ListTiersPublic godoc
// @Summary      Get all available VIP tiers and their perks for marketing & transparency
// @Tags         Customer Loyalty
// @Router       /loyalty/tiers [get]
func (h *LoyaltyHandler) ListTiersPublic(c *gin.Context) {
	l := public.NewLoyaltyPublicLogic(c.Request.Context(), h.svcCtx)
	resp, err := l.GetCustomerLoyaltyProfile("")
	if err != nil {
		result.HttpResult(c, nil, err)
		return
	}
	result.HttpResult(c, resp.AllTiers, nil)
}
