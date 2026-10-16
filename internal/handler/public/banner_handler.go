package public

import (
	"github.com/EbadiDev/anahix-server/internal/logic/public"
	"github.com/EbadiDev/anahix-server/internal/svc"
	"github.com/EbadiDev/anahix-server/pkg/result"
	"github.com/gin-gonic/gin"
)

type BannerHandler struct {
	svcCtx *svc.ServiceContext
}

func NewBannerHandler(svcCtx *svc.ServiceContext) *BannerHandler {
	return &BannerHandler{svcCtx: svcCtx}
}

func (h *BannerHandler) ListBanners(c *gin.Context) {
	position := c.Query("position")
	l := public.NewBannerLogic(c.Request.Context(), h.svcCtx)
	resp, err := l.ListPublicBanners(position)
	result.HttpResult(c, resp, err)
}
