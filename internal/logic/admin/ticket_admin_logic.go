package admin

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"strings"
	"time"

	"github.com/EbadiDev/anahix-server/internal/logic/notification"
	"github.com/EbadiDev/anahix-server/internal/logic/public"
	"github.com/EbadiDev/anahix-server/internal/model"
	"github.com/EbadiDev/anahix-server/internal/svc"
	"github.com/EbadiDev/anahix-server/internal/types"
	"github.com/EbadiDev/anahix-server/pkg/upload"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TicketAdminLogic struct {
	ctx      context.Context
	svcCtx   *svc.ServiceContext
	notifier *notification.TicketNotifier
}

func NewTicketAdminLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TicketAdminLogic {
	return &TicketAdminLogic{
		ctx:      ctx,
		svcCtx:   svcCtx,
		notifier: notification.NewTicketNotifier(svcCtx),
	}
}

// ListTickets returns paginated support tickets with filtering and status counts
func (l *TicketAdminLogic) ListTickets(req *types.AdminTicketListRequest) (*types.AdminTicketListResponse, error) {
	public.AutoCloseExpiredAwaitingUserTickets(l.svcCtx.DB.WithContext(l.ctx))

	db := l.svcCtx.DB.WithContext(l.ctx).Model(&model.Ticket{})

	// 1. Calculate status statistics
	var stats types.AdminTicketStats
	l.svcCtx.DB.WithContext(l.ctx).Model(&model.Ticket{}).Count(&stats.Total)
	l.svcCtx.DB.WithContext(l.ctx).Model(&model.Ticket{}).Where("status = ?", model.TicketStatusAwaitingAdmin).Count(&stats.AwaitingAdmin)
	l.svcCtx.DB.WithContext(l.ctx).Model(&model.Ticket{}).Where("status = ?", model.TicketStatusAwaitingUser).Count(&stats.AwaitingUser)
	l.svcCtx.DB.WithContext(l.ctx).Model(&model.Ticket{}).Where("status = ?", model.TicketStatusInProgress).Count(&stats.InProgress)
	l.svcCtx.DB.WithContext(l.ctx).Model(&model.Ticket{}).Where("status = ?", model.TicketStatusResolved).Count(&stats.Resolved)
	l.svcCtx.DB.WithContext(l.ctx).Model(&model.Ticket{}).Where("status = ?", model.TicketStatusClosed).Count(&stats.Closed)

	// 2. Apply filters
	if req.Status != "" && req.Status != "all" {
		db = db.Where("status = ?", req.Status)
	}
	if req.Department != "" && req.Department != "all" {
		db = db.Where("department = ?", req.Department)
	}
	if req.Priority != "" && req.Priority != "all" {
		db = db.Where("priority = ?", req.Priority)
	}
	if req.Search != "" {
		searchTerm := "%" + strings.TrimSpace(req.Search) + "%"
		db = db.Where("ticket_number LIKE ? OR subject LIKE ? OR customer_phone LIKE ? OR customer_name LIKE ? OR customer_email LIKE ? OR related_order_no LIKE ?",
			searchTerm, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm)
	}

	var totalFiltered int64
	if err := db.Count(&totalFiltered).Error; err != nil {
		return nil, err
	}

	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	var tickets []model.Ticket
	if err := db.Order("last_reply_at DESC, created_at DESC").
		Offset(offset).Limit(pageSize).
		Find(&tickets).Error; err != nil {
		return nil, err
	}

	items := make([]types.TicketItemDTO, 0, len(tickets))
	for _, t := range tickets {
		var msgCount int64
		l.svcCtx.DB.WithContext(l.ctx).
			Model(&model.TicketMessage{}).
			Where("ticket_id = ?", t.ID).
			Count(&msgCount)

		items = append(items, toAdminTicketItemDTO(&t, int(msgCount)))
	}

	return &types.AdminTicketListResponse{
		Items: items,
		Total: totalFiltered,
		Stats: stats,
	}, nil
}

// GetTicketDetail returns ticket data and the complete message history (including internal staff notes)
func (l *TicketAdminLogic) GetTicketDetail(ticketIDStr string) (*types.TicketDetailResponse, error) {
	public.AutoCloseExpiredAwaitingUserTickets(l.svcCtx.DB.WithContext(l.ctx))

	tUUID, err := uuid.Parse(ticketIDStr)
	if err != nil {
		return nil, errors.New("شناسه تیکت نامعتبر است")
	}

	var ticket model.Ticket
	if err := l.svcCtx.DB.WithContext(l.ctx).Where("id = ?", tUUID).First(&ticket).Error; err != nil {
		return nil, errors.New("تیکت مورد نظر یافت نشد")
	}

	var messages []model.TicketMessage
	if err := l.svcCtx.DB.WithContext(l.ctx).
		Where("ticket_id = ?", ticket.ID).
		Order("created_at ASC").
		Find(&messages).Error; err != nil {
		return nil, err
	}

	msgDTOs := make([]types.TicketMessageDTO, 0, len(messages))
	for _, m := range messages {
		msgDTOs = append(msgDTOs, toAdminTicketMessageDTO(&m))
	}

	return &types.TicketDetailResponse{
		Ticket:   toAdminTicketItemDTO(&ticket, len(messages)),
		Messages: msgDTOs,
	}, nil
}

// ReplyTicket publishes a support agent reply or records an internal staff note
func (l *TicketAdminLogic) ReplyTicket(ticketIDStr string, req *types.AdminReplyTicketRequest, adminUsername string) (*types.TicketMessageDTO, error) {
	if strings.TrimSpace(req.Message) == "" {
		return nil, errors.New("متن پاسخ الزامی است")
	}

	tUUID, err := uuid.Parse(ticketIDStr)
	if err != nil {
		return nil, errors.New("شناسه تیکت نامعتبر است")
	}

	var ticket model.Ticket
	if err := l.svcCtx.DB.WithContext(l.ctx).Where("id = ?", tUUID).First(&ticket).Error; err != nil {
		return nil, errors.New("تیکت مورد نظر یافت نشد")
	}

	var atts []model.TicketAttachment
	for _, a := range req.Attachments {
		atts = append(atts, model.TicketAttachment{
			ID:       a.ID,
			FileName: a.FileName,
			FileURL:  a.FileURL,
			FileSize: a.FileSize,
			MimeType: a.MimeType,
			IsImage:  a.IsImage,
		})
	}

	now := time.Now()
	senderName := "پشتیبانی آناهیکس"
	if req.IsInternal {
		senderName = "یادداشت محرمانه اپراتور (" + adminUsername + ")"
	}

	msg := model.TicketMessage{
		ID:         uuid.New(),
		TicketID:   ticket.ID,
		SenderType: "admin",
		SenderID:   adminUsername,
		SenderName: senderName,
		Message:    strings.TrimSpace(req.Message),
		IsInternal: req.IsInternal,
	}
	msg.SetAttachments(atts)

	txErr := l.svcCtx.DB.WithContext(l.ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&msg).Error; err != nil {
			return err
		}

		// Only public agent replies affect ticket status and trigger customer alerts
		if !req.IsInternal {
			newStatus := model.TicketStatusAwaitingUser
			if req.Status != "" {
				newStatus = model.TicketStatus(req.Status)
			}

			updates := map[string]interface{}{
				"status":              newStatus,
				"last_reply_at":       &now,
				"last_reply_by":       "admin",
				"assigned_admin_name": adminUsername,
				"updated_at":          now,
			}
			if newStatus == model.TicketStatusClosed {
				updates["closed_at"] = &now
			}

			if err := tx.Model(&ticket).Updates(updates).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if txErr != nil {
		return nil, fmt.Errorf("خطا در ثبت پاسخ: %w", txErr)
	}

	// Trigger notifications only for non-internal responses
	if !req.IsInternal {
		l.notifier.NotifyCustomerOnReply(l.ctx, &ticket, req.Message, adminUsername, req.NotifySMS, req.NotifyEmail)
	}

	// Audit log
	actionType := "ticket_public_reply"
	if req.IsInternal {
		actionType = "ticket_internal_note"
	}
	_ = l.svcCtx.DB.WithContext(l.ctx).Create(&model.AdminActionLog{
		ID:            uuid.New(),
		AdminUsername: adminUsername,
		Action:        actionType,
		Resource:      fmt.Sprintf("ticket/%s", ticket.TicketNumber),
		Details:       fmt.Sprintf("Reply sent. NotifySMS: %v, NotifyEmail: %v", req.NotifySMS, req.NotifyEmail),
		CreatedAt:     now,
	})

	dto := toAdminTicketMessageDTO(&msg)
	return &dto, nil
}

// UpdateTicket alters status, priority, department or assigned agent
func (l *TicketAdminLogic) UpdateTicket(ticketIDStr string, req *types.AdminUpdateTicketRequest, adminUsername string) (*types.TicketItemDTO, error) {
	tUUID, err := uuid.Parse(ticketIDStr)
	if err != nil {
		return nil, errors.New("شناسه تیکت نامعتبر است")
	}

	var ticket model.Ticket
	if err := l.svcCtx.DB.WithContext(l.ctx).Where("id = ?", tUUID).First(&ticket).Error; err != nil {
		return nil, errors.New("تیکت مورد نظر یافت نشد")
	}

	now := time.Now()
	updates := map[string]interface{}{
		"updated_at": now,
	}

	if req.Status != nil && *req.Status != "" {
		updates["status"] = model.TicketStatus(*req.Status)
		if model.TicketStatus(*req.Status) == model.TicketStatusClosed {
			updates["closed_at"] = &now
		}
	}
	if req.Priority != nil && *req.Priority != "" {
		updates["priority"] = model.TicketPriority(*req.Priority)
	}
	if req.Department != nil && *req.Department != "" {
		updates["department"] = *req.Department
	}
	if req.AssignedAdminName != nil {
		updates["assigned_admin_name"] = *req.AssignedAdminName
	}
	if req.NotifyViaSMS != nil {
		updates["notify_via_sms"] = *req.NotifyViaSMS
	}

	if err := l.svcCtx.DB.WithContext(l.ctx).Model(&ticket).Updates(updates).Error; err != nil {
		return nil, err
	}

	// Reload updated ticket
	_ = l.svcCtx.DB.WithContext(l.ctx).Where("id = ?", tUUID).First(&ticket)

	if req.NotifySMS || req.NotifyEmail {
		statusLabel := string(ticket.Status)
		l.notifier.NotifyCustomerOnReply(l.ctx, &ticket, fmt.Sprintf("وضعیت تیکت شما به '%s' تغییر یافت.", statusLabel), adminUsername, req.NotifySMS, req.NotifyEmail)
	}

	_ = l.svcCtx.DB.WithContext(l.ctx).Create(&model.AdminActionLog{
		ID:            uuid.New(),
		AdminUsername: adminUsername,
		Action:        "update_ticket_status",
		Resource:      fmt.Sprintf("ticket/%s", ticket.TicketNumber),
		Details:       fmt.Sprintf("Updated to status: %s, priority: %s", ticket.Status, ticket.Priority),
		CreatedAt:     now,
	})

	var msgCount int64
	l.svcCtx.DB.WithContext(l.ctx).Model(&model.TicketMessage{}).Where("ticket_id = ?", ticket.ID).Count(&msgCount)
	dto := toAdminTicketItemDTO(&ticket, int(msgCount))
	return &dto, nil
}

// DeleteTicket soft-deletes a ticket
func (l *TicketAdminLogic) DeleteTicket(ticketIDStr string, adminUsername string) error {
	tUUID, err := uuid.Parse(ticketIDStr)
	if err != nil {
		return errors.New("شناسه تیکت نامعتبر است")
	}

	var ticket model.Ticket
	if err := l.svcCtx.DB.WithContext(l.ctx).Where("id = ?", tUUID).First(&ticket).Error; err != nil {
		return errors.New("تیکت یافت نشد")
	}

	if err := l.svcCtx.DB.WithContext(l.ctx).Delete(&ticket).Error; err != nil {
		return err
	}

	_ = l.svcCtx.DB.WithContext(l.ctx).Create(&model.AdminActionLog{
		ID:            uuid.New(),
		AdminUsername: adminUsername,
		Action:        "delete_ticket",
		Resource:      fmt.Sprintf("ticket/%s", ticket.TicketNumber),
		Details:       "Ticket deleted by admin",
		CreatedAt:     time.Now(),
	})

	return nil
}

// UploadAttachment handles file uploads for admin replies
func (l *TicketAdminLogic) UploadAttachment(fileHeader *multipart.FileHeader) (*types.UploadAttachmentResponse, error) {
	return upload.SaveTicketAttachment(fileHeader)
}

// Helpers

func toAdminTicketItemDTO(t *model.Ticket, msgCount int) types.TicketItemDTO {
	return types.TicketItemDTO{
		ID:                  t.ID.String(),
		TicketNumber:        t.TicketNumber,
		CustomerID:          t.CustomerID.String(),
		CustomerPhone:       t.CustomerPhone,
		CustomerName:        t.CustomerName,
		CustomerEmail:       t.CustomerEmail,
		Subject:             t.Subject,
		Department:          t.Department,
		Priority:            string(t.Priority),
		Status:              string(t.Status),
		RelatedOrderNo:      t.RelatedOrderNo,
		NotifyViaSMS:        t.NotifyViaSMS,
		NotifyViaEmail:      t.NotifyViaEmail,
		AssignedAdminName:   t.AssignedAdminName,
		SatisfactionRating:  t.SatisfactionRating,
		SatisfactionComment: t.SatisfactionComment,
		LastReplyAt:         t.LastReplyAt,
		LastReplyBy:         t.LastReplyBy,
		MessagesCount:       msgCount,
		CreatedAt:           t.CreatedAt,
		UpdatedAt:           t.UpdatedAt,
	}
}

func toAdminTicketMessageDTO(m *model.TicketMessage) types.TicketMessageDTO {
	var atts []types.TicketAttachmentDTO
	for _, a := range m.Attachments {
		atts = append(atts, types.TicketAttachmentDTO{
			ID:       a.ID,
			FileName: a.FileName,
			FileURL:  a.FileURL,
			FileSize: a.FileSize,
			MimeType: a.MimeType,
			IsImage:  a.IsImage,
		})
	}
	if atts == nil {
		atts = []types.TicketAttachmentDTO{}
	}

	return types.TicketMessageDTO{
		ID:          m.ID.String(),
		TicketID:    m.TicketID.String(),
		SenderType:  m.SenderType,
		SenderID:    m.SenderID,
		SenderName:  m.SenderName,
		Message:     m.Message,
		Attachments: atts,
		IsInternal:  m.IsInternal,
		CreatedAt:   m.CreatedAt,
	}
}
