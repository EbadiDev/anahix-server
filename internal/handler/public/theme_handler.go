package public

import (
	"github.com/EbadiDev/anahix-server/internal/logic/public"
	"github.com/EbadiDev/anahix-server/internal/svc"
	"github.com/EbadiDev/anahix-server/pkg/result"
	"github.com/gin-gonic/gin"
)

type ThemeHandler struct {
	svcCtx *svc.ServiceContext
}

func NewThemeHandler(svcCtx *svc.ServiceContext) *ThemeHandler {
	return &ThemeHandler{svcCtx: svcCtx}
}

func (h *ThemeHandler) GetActiveTheme(c *gin.Context) {
	l := public.NewThemeLogic(c.Request.Context(), h.svcCtx)
	resp, err := l.GetActiveTheme()
	result.HttpResult(c, resp, err)
}
