package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TransactionStatus string

const (
	TxStatusSuccess  TransactionStatus = "SUCCESS"
	TxStatusPending  TransactionStatus = "PENDING"
	TxStatusFailed   TransactionStatus = "FAILED"
	TxStatusRefunded TransactionStatus = "REFUNDED"
)

type TransactionType string

const (
	TxTypeStoreOrder    TransactionType = "store_order"
	TxTypeVirtualNumber TransactionType = "virtual_number"
	TxTypeDeposit       TransactionType = "deposit"
	TxTypeRefund        TransactionType = "refund"
)

// Transaction represents a financial or digital asset transaction record
type Transaction struct {
	ID                uuid.UUID         `gorm:"type:uuid;primaryKey" json:"id"`
	TransactionNumber string            `gorm:"size:64;uniqueIndex;not null" json:"transaction_number"`
	OrderID           *uuid.UUID        `gorm:"type:uuid;index" json:"order_id,omitempty"`
	OrderType         TransactionType   `gorm:"size:32;not null;index" json:"order_type"`
	CustomerPhone     string            `gorm:"size:32;index" json:"customer_phone"`
	CustomerEmail     string            `gorm:"size:120;index" json:"customer_email"`
	AmountToman       int64             `gorm:"not null" json:"amount_toman"`
	AmountUSD         float64           `gorm:"type:decimal(10,2);default:0" json:"amount_usd"`
	PaymentPlatform   PaymentPlatform   `gorm:"size:32;not null;index" json:"payment_platform"`
	PaymentRefID      string            `gorm:"size:120" json:"payment_ref_id,omitempty"`
	PaymentAuthority  string            `gorm:"size:120;index" json:"payment_authority,omitempty"`
	Status            TransactionStatus `gorm:"size:24;not null;default:'SUCCESS';index" json:"status"`
	Description       string            `gorm:"type:text" json:"description"`
	Metadata          string            `gorm:"type:text;default:'{}'" json:"metadata"` // JSON metadata
	CreatedAt         time.Time         `gorm:"index" json:"created_at"`
	UpdatedAt         time.Time         `json:"updated_at"`
	DeletedAt         gorm.DeletedAt    `gorm:"index" json:"-"`
}

func (t *Transaction) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

func (Transaction) TableName() string {
	return "transactions"
}

// AdminActionLog represents an immutable record of an administrative action
type AdminActionLog struct {
	ID            uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	AdminID       *uuid.UUID     `gorm:"type:uuid;index" json:"admin_id,omitempty"`
	AdminUsername string         `gorm:"size:60;not null;index" json:"admin_username"`
	Action        string         `gorm:"size:60;not null;index" json:"action"`
	Resource      string         `gorm:"size:60;not null;index" json:"resource"`
	ResourceID    string         `gorm:"size:120;index" json:"resource_id,omitempty"`
	Description   string         `gorm:"type:text;not null" json:"description"`
	Details       string         `gorm:"type:text;default:'{}'" json:"details"` // JSON payload / changes diff
	IPAddress     string         `gorm:"size:64" json:"ip_address"`
	UserAgent     string         `gorm:"size:255" json:"user_agent"`
	CreatedAt     time.Time      `gorm:"index" json:"created_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

func (a *AdminActionLog) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}

func (AdminActionLog) TableName() string {
	return "admin_action_logs"
}
