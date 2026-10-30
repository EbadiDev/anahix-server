package admin

import (
	"github.com/EbadiDev/anahix-server/internal/logic/admin"
	"github.com/EbadiDev/anahix-server/internal/svc"
	"github.com/EbadiDev/anahix-server/internal/types"
	"github.com/EbadiDev/anahix-server/pkg/result"
	"github.com/gin-gonic/gin"
)

type ThemeAdminHandler struct {
	svcCtx *svc.ServiceContext
}

func NewThemeAdminHandler(svcCtx *svc.ServiceContext) *ThemeAdminHandler {
	return &ThemeAdminHandler{svcCtx: svcCtx}
}

func (h *ThemeAdminHandler) GetSettings(c *gin.Context) {
	c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	l := admin.NewThemeAdminLogic(c.Request.Context(), h.svcCtx)
	resp, err := l.GetSettings()
	result.HttpResult(c, resp, err)
}

func (h *ThemeAdminHandler) UpdateSettings(c *gin.Context) {
	var req types.UpdateThemeSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		result.ParamErrorResult(c, err)
		return
	}

	l := admin.NewThemeAdminLogic(c.Request.Context(), h.svcCtx)
	err := l.UpdateSettings(&req)
	result.HttpResult(c, gin.H{"updated": true}, err)
}
