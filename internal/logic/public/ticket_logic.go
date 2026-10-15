package public

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"strings"
	"time"

	"github.com/EbadiDev/anahix-server/internal/model"
	"github.com/EbadiDev/anahix-server/internal/svc"
	"github.com/EbadiDev/anahix-server/internal/types"
	"github.com/EbadiDev/anahix-server/pkg/upload"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TicketLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewTicketLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TicketLogic {
	return &TicketLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// CreateTicket initiates a new support ticket and records the opening message
func (l *TicketLogic) CreateTicket(req *types.CreateTicketRequest, customerIDStr string) (*types.TicketDetailResponse, error) {
	if strings.TrimSpace(req.Subject) == "" {
		return nil, errors.New("موضوع تیکت الزامی است")
	}
	if strings.TrimSpace(req.Message) == "" {
		return nil, errors.New("متن پیام تیکت الزامی است")
	}

	custUUID, err := uuid.Parse(customerIDStr)
	if err != nil {
		return nil, errors.New("شناسه مشتری نامعتبر است")
	}

	var customer model.Customer
	if err := l.svcCtx.DB.WithContext(l.ctx).Where("id = ?", custUUID).First(&customer).Error; err != nil {
		return nil, errors.New("حساب کاربری یافت نشد")
	}

	customerName := customer.FullName
	if strings.TrimSpace(customerName) == "" {
		customerName = customer.Phone
	}

	priority := model.TicketPriorityMedium
	if req.Priority != "" {
		priority = model.TicketPriority(req.Priority)
	}

	notifySMS := true
	if req.NotifyViaSMS != nil {
		notifySMS = *req.NotifyViaSMS
	}

	notifyEmail := true
	if req.NotifyViaEmail != nil {
		notifyEmail = *req.NotifyViaEmail
	}

	now := time.Now()
	ticket := model.Ticket{
		ID:             uuid.New(),
		CustomerID:     customer.ID,
		CustomerPhone:  customer.Phone,
		CustomerName:   customerName,
		CustomerEmail:  customer.Email,
		Subject:        strings.TrimSpace(req.Subject),
		Department:     strings.TrimSpace(req.Department),
		Priority:       priority,
		Status:         model.TicketStatusAwaitingAdmin,
		RelatedOrderNo: strings.TrimSpace(req.RelatedOrderNo),
		NotifyViaSMS:   notifySMS,
		NotifyViaEmail: notifyEmail,
		LastReplyAt:    &now,
		LastReplyBy:    "customer",
	}

	// Prepare attachments
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

	firstMessage := model.TicketMessage{
		ID:         uuid.New(),
		TicketID:   ticket.ID,
		SenderType: "customer",
		SenderID:   customer.ID.String(),
		SenderName: customerName,
		Message:    strings.TrimSpace(req.Message),
		IsInternal: false,
	}
	firstMessage.SetAttachments(atts)

	// Save within transaction
	txErr := l.svcCtx.DB.WithContext(l.ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&ticket).Error; err != nil {
			return err
		}
		if err := tx.Create(&firstMessage).Error; err != nil {
			return err
		}
		return nil
	})
	if txErr != nil {
		return nil, fmt.Errorf("خطا در ثبت تیکت: %w", txErr)
	}

	return &types.TicketDetailResponse{
		Ticket:   toTicketItemDTO(&ticket, 1),
		Messages: []types.TicketMessageDTO{toTicketMessageDTO(&firstMessage)},
	}, nil
}

// ListMyTickets returns all tickets submitted by the authenticated customer
func (l *TicketLogic) ListMyTickets(customerIDStr string) ([]types.TicketItemDTO, error) {
	custUUID, err := uuid.Parse(customerIDStr)
	if err != nil {
		return nil, errors.New("شناسه مشتری نامعتبر است")
	}

	var tickets []model.Ticket
	if err := l.svcCtx.DB.WithContext(l.ctx).
		Where("customer_id = ?", custUUID).
		Order("updated_at DESC").
		Find(&tickets).Error; err != nil {
		return nil, err
	}

	dtos := make([]types.TicketItemDTO, 0, len(tickets))
	for _, t := range tickets {
		var msgCount int64
		l.svcCtx.DB.WithContext(l.ctx).
			Model(&model.TicketMessage{}).
			Where("ticket_id = ? AND is_internal = ?", t.ID, false).
			Count(&msgCount)

		dtos = append(dtos, toTicketItemDTO(&t, int(msgCount)))
	}

	return dtos, nil
}

// GetTicketDetail retrieves a customer's specific ticket and customer-visible messages
func (l *TicketLogic) GetTicketDetail(ticketIDStr string, customerIDStr string) (*types.TicketDetailResponse, error) {
	tUUID, err := uuid.Parse(ticketIDStr)
	if err != nil {
		return nil, errors.New("شناسه تیکت نامعتبر است")
	}
	cUUID, err := uuid.Parse(customerIDStr)
	if err != nil {
		return nil, errors.New("شناسه مشتری نامعتبر است")
	}

	var ticket model.Ticket
	if err := l.svcCtx.DB.WithContext(l.ctx).
		Where("id = ? AND customer_id = ?", tUUID, cUUID).
		First(&ticket).Error; err != nil {
		return nil, errors.New("تیکت مورد نظر یافت نشد")
	}

	var messages []model.TicketMessage
	if err := l.svcCtx.DB.WithContext(l.ctx).
		Where("ticket_id = ? AND is_internal = ?", ticket.ID, false).
		Order("created_at ASC").
		Find(&messages).Error; err != nil {
		return nil, err
	}

	msgDTOs := make([]types.TicketMessageDTO, 0, len(messages))
	for _, m := range messages {
		msgDTOs = append(msgDTOs, toTicketMessageDTO(&m))
	}

	return &types.TicketDetailResponse{
		Ticket:   toTicketItemDTO(&ticket, len(messages)),
		Messages: msgDTOs,
	}, nil
}

// ReplyTicket adds a customer reply to the ticket
func (l *TicketLogic) ReplyTicket(ticketIDStr string, req *types.ReplyTicketRequest, customerIDStr string) (*types.TicketMessageDTO, error) {
	if strings.TrimSpace(req.Message) == "" {
		return nil, errors.New("متن پاسخ نمی‌تواند خالی باشد")
	}

	tUUID, err := uuid.Parse(ticketIDStr)
	if err != nil {
		return nil, errors.New("شناسه تیکت نامعتبر است")
	}
	cUUID, err := uuid.Parse(customerIDStr)
	if err != nil {
		return nil, errors.New("شناسه مشتری نامعتبر است")
	}

	var ticket model.Ticket
	if err := l.svcCtx.DB.WithContext(l.ctx).
		Where("id = ? AND customer_id = ?", tUUID, cUUID).
		First(&ticket).Error; err != nil {
		return nil, errors.New("تیکت مورد نظر یافت نشد")
	}

	customerName := ticket.CustomerName
	if strings.TrimSpace(customerName) == "" {
		customerName = ticket.CustomerPhone
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
	msg := model.TicketMessage{
		ID:         uuid.New(),
		TicketID:   ticket.ID,
		SenderType: "customer",
		SenderID:   customerIDStr,
		SenderName: customerName,
		Message:    strings.TrimSpace(req.Message),
		IsInternal: false,
	}
	msg.SetAttachments(atts)

	updates := map[string]interface{}{
		"status":        model.TicketStatusAwaitingAdmin,
		"last_reply_at": &now,
		"last_reply_by": "customer",
		"updated_at":    now,
	}
	if req.NotifyViaSMS != nil {
		updates["notify_via_sms"] = *req.NotifyViaSMS
	}
	if req.NotifyViaEmail != nil {
		updates["notify_via_email"] = *req.NotifyViaEmail
	}

	txErr := l.svcCtx.DB.WithContext(l.ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&msg).Error; err != nil {
			return err
		}
		if err := tx.Model(&ticket).Updates(updates).Error; err != nil {
			return err
		}
		return nil
	})
	if txErr != nil {
		return nil, fmt.Errorf("خطا در ارسال پاسخ: %w", txErr)
	}

	dto := toTicketMessageDTO(&msg)
	return &dto, nil
}

// CloseTicket closes a customer ticket
func (l *TicketLogic) CloseTicket(ticketIDStr string, customerIDStr string) error {
	tUUID, err := uuid.Parse(ticketIDStr)
	if err != nil {
		return errors.New("شناسه تیکت نامعتبر است")
	}
	cUUID, err := uuid.Parse(customerIDStr)
	if err != nil {
		return errors.New("شناسه مشتری نامعتبر است")
	}

	now := time.Now()
	res := l.svcCtx.DB.WithContext(l.ctx).
		Model(&model.Ticket{}).
		Where("id = ? AND customer_id = ?", tUUID, cUUID).
		Updates(map[string]interface{}{
			"status":     model.TicketStatusClosed,
			"closed_at":  &now,
			"updated_at": now,
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("تیکت یافت نشد")
	}
	return nil
}

// RateTicket registers customer satisfaction rating and comment for a ticket
func (l *TicketLogic) RateTicket(ticketIDStr string, req *types.RateTicketRequest, customerIDStr string) error {
	if req.Rating < 1 || req.Rating > 5 {
		return errors.New("امتیاز باید بین ۱ تا ۵ ستاره باشد")
	}

	tUUID, err := uuid.Parse(ticketIDStr)
	if err != nil {
		return errors.New("شناسه تیکت نامعتبر است")
	}
	cUUID, err := uuid.Parse(customerIDStr)
	if err != nil {
		return errors.New("شناسه مشتری نامعتبر است")
	}

	res := l.svcCtx.DB.WithContext(l.ctx).
		Model(&model.Ticket{}).
		Where("id = ? AND customer_id = ?", tUUID, cUUID).
		Updates(map[string]interface{}{
			"satisfaction_rating":  req.Rating,
			"satisfaction_comment": strings.TrimSpace(req.Comment),
			"updated_at":           time.Now(),
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("تیکت یافت نشد")
	}
	return nil
}

// UploadAttachment handles file uploads for tickets
func (l *TicketLogic) UploadAttachment(fileHeader *multipart.FileHeader) (*types.UploadAttachmentResponse, error) {
	return upload.SaveTicketAttachment(fileHeader)
}

// Helper converters

func toTicketItemDTO(t *model.Ticket, msgCount int) types.TicketItemDTO {
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

func toTicketMessageDTO(m *model.TicketMessage) types.TicketMessageDTO {
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
