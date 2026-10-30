package public

import (
	"github.com/EbadiDev/anahix-server/internal/logic/public"
	"github.com/EbadiDev/anahix-server/internal/svc"
	"github.com/EbadiDev/anahix-server/internal/types"
	"github.com/EbadiDev/anahix-server/pkg/result"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type VirtualNumberHandler struct {
	svcCtx *svc.ServiceContext
}

func NewVirtualNumberHandler(svcCtx *svc.ServiceContext) *VirtualNumberHandler {
	return &VirtualNumberHandler{svcCtx: svcCtx}
}

func (h *VirtualNumberHandler) ListServices(c *gin.Context) {
	l := public.NewVirtualNumberLogic(c.Request.Context(), h.svcCtx)
	services, err := l.ListServices()
	result.HttpResult(c, services, err)
}

func (h *VirtualNumberHandler) ListCountriesForService(c *gin.Context) {
	service := c.Query("service")
	l := public.NewVirtualNumberLogic(c.Request.Context(), h.svcCtx)
	countries, err := l.ListCountriesForService(service)
	result.HttpResult(c, countries, err)
}

func (h *VirtualNumberHandler) ListCatalog(c *gin.Context) {
	l := public.NewVirtualNumberLogic(c.Request.Context(), h.svcCtx)
	items, err := l.GetCatalog()
	result.HttpResult(c, items, err)
}

func (h *VirtualNumberHandler) CreateOrder(c *gin.Context) {
	var req types.BuyVirtualNumberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		result.ParamErrorResult(c, err)
		return
	}

	custIDVal, hasAuth := c.Get("customer_id")
	if hasAuth {
		if uid, err := uuid.Parse(custIDVal.(string)); err == nil {
			req.AuthCustomerID = &uid
		}
	}

	l := public.NewVirtualNumberLogic(c.Request.Context(), h.svcCtx)
	resp, err := l.CreateOrder(&req)
	result.HttpResult(c, resp, err)
}

func (h *VirtualNumberHandler) GetOrderDetail(c *gin.Context) {
	orderID := c.Param("order_id")
	l := public.NewVirtualNumberLogic(c.Request.Context(), h.svcCtx)
	resp, err := l.GetOrderDetail(orderID)
	result.HttpResult(c, resp, err)
}

func (h *VirtualNumberHandler) CancelOrder(c *gin.Context) {
	orderID := c.Param("order_id")
	l := public.NewVirtualNumberLogic(c.Request.Context(), h.svcCtx)
	err := l.CancelOrder(orderID)
	result.HttpResult(c, gin.H{"status": "canceled", "message": "سفارش لغو شد و وجه استرداد گردید"}, err)
}

func (h *VirtualNumberHandler) FinishOrder(c *gin.Context) {
	orderID := c.Param("order_id")
	l := public.NewVirtualNumberLogic(c.Request.Context(), h.svcCtx)
	err := l.FinishOrder(orderID)
	result.HttpResult(c, gin.H{"status": "finished", "message": "فعال‌سازی با موفقیت نهایی شد"}, err)
}
