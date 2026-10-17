package public

import (
	"github.com/EbadiDev/anahix-server/internal/logic/public"
	"github.com/EbadiDev/anahix-server/internal/svc"
	"github.com/EbadiDev/anahix-server/pkg/result"
	"github.com/EbadiDev/anahix-server/pkg/xerr"
	"github.com/gin-gonic/gin"
)

type NotificationHandler struct {
	svcCtx *svc.ServiceContext
}

func NewNotificationHandler(svcCtx *svc.ServiceContext) *NotificationHandler {
	return &NotificationHandler{svcCtx: svcCtx}
}

// ListNotifications returns customer's in-app notification list
func (h *NotificationHandler) ListNotifications(c *gin.Context) {
	customerID := c.GetString("customer_id")
	if customerID == "" {
		result.ParamErrorResult(c, xerr.NewErrCode(xerr.Unauthorized))
		return
	}

	l := public.NewNotificationLogic(c.Request.Context(), h.svcCtx)
	resp, err := l.ListNotifications(customerID)
	result.HttpResult(c, resp, err)
}

// MarkAsRead marks single notification as read
func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	customerID := c.GetString("customer_id")
	notifID := c.Param("id")
	if customerID == "" || notifID == "" {
		result.ParamErrorResult(c, xerr.NewErrMsg("پارامترهای درخواست نامعتبر است"))
		return
	}

	l := public.NewNotificationLogic(c.Request.Context(), h.svcCtx)
	err := l.MarkAsRead(customerID, notifID)
	result.HttpResult(c, gin.H{"updated": true}, err)
}

// MarkAllAsRead marks all customer notifications as read
func (h *NotificationHandler) MarkAllAsRead(c *gin.Context) {
	customerID := c.GetString("customer_id")
	if customerID == "" {
		result.ParamErrorResult(c, xerr.NewErrCode(xerr.Unauthorized))
		return
	}

	l := public.NewNotificationLogic(c.Request.Context(), h.svcCtx)
	err := l.MarkAllAsRead(customerID)
	result.HttpResult(c, gin.H{"updated": true}, err)
}

// GetAdminNotificationStats returns platform notification statistics
func (h *NotificationHandler) GetAdminNotificationStats(c *gin.Context) {
	l := public.NewNotificationLogic(c.Request.Context(), h.svcCtx)
	resp, err := l.GetAdminNotificationStats()
	result.HttpResult(c, resp, err)
}
