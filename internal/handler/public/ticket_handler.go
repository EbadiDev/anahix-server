package public

import (
	"errors"

	"github.com/EbadiDev/anahix-server/internal/logic/public"
	"github.com/EbadiDev/anahix-server/internal/svc"
	"github.com/EbadiDev/anahix-server/internal/types"
	"github.com/EbadiDev/anahix-server/pkg/result"
	"github.com/EbadiDev/anahix-server/pkg/xerr"
	"github.com/gin-gonic/gin"
)

type TicketHandler struct {
	svcCtx *svc.ServiceContext
}

func NewTicketHandler(svcCtx *svc.ServiceContext) *TicketHandler {
	return &TicketHandler{svcCtx: svcCtx}
}

// CreateTicket initiates a new customer ticket
func (h *TicketHandler) CreateTicket(c *gin.Context) {
	customerID := c.GetString("customer_id")
	if customerID == "" {
		result.ParamErrorResult(c, xerr.NewErrCode(xerr.Unauthorized))
		return
	}

	var req types.CreateTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		result.ParamErrorResult(c, err)
		return
	}

	l := public.NewTicketLogic(c.Request.Context(), h.svcCtx)
	resp, err := l.CreateTicket(&req, customerID)
	result.HttpResult(c, resp, err)
}

// ListMyTickets returns all tickets for current customer
func (h *TicketHandler) ListMyTickets(c *gin.Context) {
	customerID := c.GetString("customer_id")
	if customerID == "" {
		result.ParamErrorResult(c, xerr.NewErrCode(xerr.Unauthorized))
		return
	}

	l := public.NewTicketLogic(c.Request.Context(), h.svcCtx)
	resp, err := l.ListMyTickets(customerID)
	result.HttpResult(c, resp, err)
}

// GetTicketDetail returns conversation details for one ticket
func (h *TicketHandler) GetTicketDetail(c *gin.Context) {
	customerID := c.GetString("customer_id")
	if customerID == "" {
		result.ParamErrorResult(c, xerr.NewErrCode(xerr.Unauthorized))
		return
	}

	ticketID := c.Param("id")
	if ticketID == "" {
		result.ParamErrorResult(c, errors.New("شناسه تیکت مشخص نشده است"))
		return
	}

	l := public.NewTicketLogic(c.Request.Context(), h.svcCtx)
	resp, err := l.GetTicketDetail(ticketID, customerID)
	result.HttpResult(c, resp, err)
}

// ReplyTicket sends customer reply in ticket thread
func (h *TicketHandler) ReplyTicket(c *gin.Context) {
	customerID := c.GetString("customer_id")
	if customerID == "" {
		result.ParamErrorResult(c, xerr.NewErrCode(xerr.Unauthorized))
		return
	}

	ticketID := c.Param("id")
	if ticketID == "" {
		result.ParamErrorResult(c, errors.New("شناسه تیکت مشخص نشده است"))
		return
	}

	var req types.ReplyTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		result.ParamErrorResult(c, err)
		return
	}

	l := public.NewTicketLogic(c.Request.Context(), h.svcCtx)
	resp, err := l.ReplyTicket(ticketID, &req, customerID)
	result.HttpResult(c, resp, err)
}

// CloseTicket closes a customer ticket
func (h *TicketHandler) CloseTicket(c *gin.Context) {
	customerID := c.GetString("customer_id")
	if customerID == "" {
		result.ParamErrorResult(c, xerr.NewErrCode(xerr.Unauthorized))
		return
	}

	ticketID := c.Param("id")
	if ticketID == "" {
		result.ParamErrorResult(c, errors.New("شناسه تیکت مشخص نشده است"))
		return
	}

	l := public.NewTicketLogic(c.Request.Context(), h.svcCtx)
	err := l.CloseTicket(ticketID, customerID)
	result.HttpResult(c, gin.H{"closed": true}, err)
}

// RateTicket registers satisfaction feedback
func (h *TicketHandler) RateTicket(c *gin.Context) {
	customerID := c.GetString("customer_id")
	if customerID == "" {
		result.ParamErrorResult(c, xerr.NewErrCode(xerr.Unauthorized))
		return
	}

	ticketID := c.Param("id")
	if ticketID == "" {
		result.ParamErrorResult(c, errors.New("شناسه تیکت مشخص نشده است"))
		return
	}

	var req types.RateTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		result.ParamErrorResult(c, err)
		return
	}

	l := public.NewTicketLogic(c.Request.Context(), h.svcCtx)
	err := l.RateTicket(ticketID, &req, customerID)
	result.HttpResult(c, gin.H{"rated": true}, err)
}

// UploadAttachment handles file upload for ticket attachments
func (h *TicketHandler) UploadAttachment(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		result.ParamErrorResult(c, errors.New("فایل ارسال نشده است یا فیلد file خالی است"))
		return
	}

	l := public.NewTicketLogic(c.Request.Context(), h.svcCtx)
	resp, err := l.UploadAttachment(file)
	result.HttpResult(c, resp, err)
}
