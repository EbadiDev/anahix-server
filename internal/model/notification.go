package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CustomerNotification struct {
	ID         uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	CustomerID uuid.UUID      `gorm:"type:uuid;index;not null" json:"customer_id"`
	Title      string         `gorm:"size:255;not null" json:"title"`
	Message    string         `gorm:"type:text;not null" json:"message"`
	Type       string         `gorm:"size:50;default:'system'" json:"type"` // "order", "security", "loyalty", "system"
	Link       string         `gorm:"size:255" json:"link,omitempty"`
	IsRead     bool           `gorm:"default:false;index" json:"is_read"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}
