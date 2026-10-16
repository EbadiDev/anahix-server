package admin

import (
	"github.com/EbadiDev/anahix-server/internal/logic/admin"
	"github.com/EbadiDev/anahix-server/internal/svc"
	"github.com/EbadiDev/anahix-server/internal/types"
	"github.com/EbadiDev/anahix-server/pkg/result"
	"github.com/gin-gonic/gin"
)

type BIAnalyticsHandler struct {
	svcCtx *svc.ServiceContext
}

func NewBIAnalyticsHandler(svcCtx *svc.ServiceContext) *BIAnalyticsHandler {
	return &BIAnalyticsHandler{svcCtx: svcCtx}
}

// GetOverview godoc
// @Summary      Get Business Intelligence & Analytics Overview
// @Description  Compiles comprehensive sales, gross profit COGS, conversion rate, AOV, top products & virtual numbers
// @Tags         Admin BI & Analytics
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        range query string false "Date range: 7d, 30d, 90d, 1y, all"
// @Param        group_by query string false "Group by: daily, weekly, monthly"
// @Success      200 {object} result.Response{data=types.AdminBIOverviewResponse}
// @Router       /admin/bi/overview [get]
func (h *BIAnalyticsHandler) GetOverview(c *gin.Context) {
	var req types.AdminBIOverviewRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		result.ParamErrorResult(c, err)
		return
	}

	l := admin.NewBIAnalyticsLogic(c.Request.Context(), h.svcCtx)
	resp, err := l.GetOverview(&req)
	result.HttpResult(c, resp, err)
}
