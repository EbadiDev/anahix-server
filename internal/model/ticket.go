package model

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TicketStatus string

const (
	TicketStatusAwaitingAdmin TicketStatus = "awaiting_admin" // در انتظار پاسخ پشتیبان
	TicketStatusAwaitingUser  TicketStatus = "awaiting_user"  // در انتظار پاسخ کاربر
	TicketStatusInProgress    TicketStatus = "in_progress"    // در حال بررسی توسط کارشناس فنی
	TicketStatusResolved      TicketStatus = "resolved"       // حل شده / پاسخ داده شده
	TicketStatusClosed        TicketStatus = "closed"         // بسته شده
)

type TicketPriority string

const (
	TicketPriorityLow    TicketPriority = "low"    // کم
	TicketPriorityMedium TicketPriority = "medium" // متوسط
	TicketPriorityHigh   TicketPriority = "high"   // زیاد
	TicketPriorityUrgent TicketPriority = "urgent" // فوری / بحرانی
)

type TicketDepartment string

const (
	DepartmentVirtualNumbers TicketDepartment = "virtual_numbers" // شماره‌های مجازی (5SIM)
	DepartmentAIAccounts     TicketDepartment = "ai_accounts"     // اکانت‌های هوش مصنوعی
	DepartmentLicenseSupport TicketDepartment = "license_support" // لایسنس و خطای فعال‌سازی
	DepartmentBilling        TicketDepartment = "billing"         // امور مالی، واریز و کیف پول
	DepartmentGeneral        TicketDepartment = "general"         // عمومی و مشاوره
)

// Ticket represents a customer support ticket
type Ticket struct {
	ID                  uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	TicketNumber        string         `gorm:"size:64;uniqueIndex;not null" json:"ticket_number"`
	CustomerID          uuid.UUID      `gorm:"type:uuid;index;not null" json:"customer_id"`
	CustomerPhone       string         `gorm:"size:30;index" json:"customer_phone"`
	CustomerName        string         `gorm:"size:120" json:"customer_name"`
	CustomerEmail       string         `gorm:"size:120;index" json:"customer_email"`
	Subject             string         `gorm:"size:255;not null" json:"subject"`
	Department          string         `gorm:"size:60;not null;index" json:"department"`
	Priority            TicketPriority `gorm:"size:20;not null;default:'medium';index" json:"priority"`
	Status              TicketStatus   `gorm:"size:30;not null;default:'awaiting_admin';index" json:"status"`
	RelatedOrderNo      string         `gorm:"size:64;index" json:"related_order_no,omitempty"`
	NotifyViaSMS        bool           `gorm:"default:true" json:"notify_via_sms"`
	NotifyViaEmail      bool           `gorm:"default:true" json:"notify_via_email"`
	AssignedAdminID     *uuid.UUID     `gorm:"type:uuid;index" json:"assigned_admin_id,omitempty"`
	AssignedAdminName   string         `gorm:"size:120" json:"assigned_admin_name,omitempty"`
	SatisfactionRating  int            `gorm:"default:0" json:"satisfaction_rating,omitempty"` // 1 to 5
	SatisfactionComment string         `gorm:"size:500" json:"satisfaction_comment,omitempty"`
	LastReplyAt         *time.Time     `json:"last_reply_at,omitempty"`
	LastReplyBy         string         `gorm:"size:30;default:'customer'" json:"last_reply_by"` // "customer" | "admin"
	ClosedAt            *time.Time     `json:"closed_at,omitempty"`

	Messages []TicketMessage `gorm:"foreignKey:TicketID" json:"messages,omitempty"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (t *Ticket) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	if t.TicketNumber == "" {
		b := make([]byte, 3)
		_, _ = rand.Read(b)
		t.TicketNumber = fmt.Sprintf("TK-%s", hex.EncodeToString(b))
	}
	if t.Priority == "" {
		t.Priority = TicketPriorityMedium
	}
	if t.Status == "" {
		t.Status = TicketStatusAwaitingAdmin
	}
	now := time.Now()
	if t.LastReplyAt == nil {
		t.LastReplyAt = &now
	}
	return nil
}

// TicketAttachment represents a file or image uploaded to a ticket message
type TicketAttachment struct {
	ID       string `json:"id"`
	FileName string `json:"file_name"`
	FileURL  string `json:"file_url"`
	FileSize int64  `json:"file_size"`
	MimeType string `json:"mime_type"`
	IsImage  bool   `json:"is_image"`
}

// TicketMessage represents one message in a ticket thread
type TicketMessage struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	TicketID        uuid.UUID `gorm:"type:uuid;not null;index" json:"ticket_id"`
	SenderType      string    `gorm:"size:30;not null" json:"sender_type"` // "customer", "admin", "system"
	SenderID        string    `gorm:"size:64;not null" json:"sender_id"`
	SenderName      string    `gorm:"size:120" json:"sender_name"`
	Message         string    `gorm:"type:text;not null" json:"message"`
	AttachmentsJSON string    `gorm:"type:text" json:"attachments_json,omitempty"`
	IsInternal      bool      `gorm:"default:false;index" json:"is_internal"` // internal admin note invisible to user

	Attachments []TicketAttachment `gorm:"-" json:"attachments,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (m *TicketMessage) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}

func (m *TicketMessage) AfterFind(tx *gorm.DB) error {
	if m.AttachmentsJSON != "" {
		var atts []TicketAttachment
		if err := json.Unmarshal([]byte(m.AttachmentsJSON), &atts); err == nil {
			m.Attachments = atts
		}
	}
	return nil
}

func (m *TicketMessage) SetAttachments(atts []TicketAttachment) {
	m.Attachments = atts
	if len(atts) == 0 {
		m.AttachmentsJSON = ""
		return
	}
	if b, err := json.Marshal(atts); err == nil {
		m.AttachmentsJSON = string(b)
	}
}
