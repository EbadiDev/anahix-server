package admin

import (
	"github.com/EbadiDev/anahix-server/internal/logic/admin"
	"github.com/EbadiDev/anahix-server/internal/svc"
	"github.com/EbadiDev/anahix-server/internal/types"
	"github.com/EbadiDev/anahix-server/pkg/result"
	"github.com/gin-gonic/gin"
)

type BannerAdminHandler struct {
	svcCtx *svc.ServiceContext
}

func NewBannerAdminHandler(svcCtx *svc.ServiceContext) *BannerAdminHandler {
	return &BannerAdminHandler{svcCtx: svcCtx}
}

func (h *BannerAdminHandler) ListBanners(c *gin.Context) {
	l := admin.NewBannerAdminLogic(c.Request.Context(), h.svcCtx)
	resp, err := l.ListBanners()
	result.HttpResult(c, resp, err)
}

func (h *BannerAdminHandler) CreateBanner(c *gin.Context) {
	var req types.CreateBannerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		result.ParamErrorResult(c, err)
		return
	}

	l := admin.NewBannerAdminLogic(c.Request.Context(), h.svcCtx)
	resp, err := l.CreateBanner(&req)
	result.HttpResult(c, resp, err)
}

func (h *BannerAdminHandler) UpdateBanner(c *gin.Context) {
	id := c.Param("id")
	var req types.UpdateBannerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		result.ParamErrorResult(c, err)
		return
	}

	l := admin.NewBannerAdminLogic(c.Request.Context(), h.svcCtx)
	resp, err := l.UpdateBanner(id, &req)
	result.HttpResult(c, resp, err)
}

func (h *BannerAdminHandler) DeleteBanner(c *gin.Context) {
	id := c.Param("id")
	l := admin.NewBannerAdminLogic(c.Request.Context(), h.svcCtx)
	err := l.DeleteBanner(id)
	result.HttpResult(c, gin.H{"deleted": true}, err)
}
