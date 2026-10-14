package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Customer struct {
	ID           uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	Phone        string         `gorm:"size:30;uniqueIndex;not null" json:"phone"`
	FullName     string         `gorm:"size:120" json:"full_name"`
	Email        string         `gorm:"size:120" json:"email"`
	BalanceToman int64          `gorm:"not null;default:0" json:"balance_toman"`
	IsActive     bool           `gorm:"default:true;index" json:"is_active"`
	LastLoginAt  *time.Time     `json:"last_login_at,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

func (c *Customer) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

type OTPVerification struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Phone     string    `gorm:"size:30;index;not null" json:"phone"`
	Code      string    `gorm:"size:10;not null" json:"code"`
	ExpiresAt time.Time `gorm:"index;not null" json:"expires_at"`
	Attempts  int       `gorm:"default:0" json:"attempts"`
	Used      bool      `gorm:"default:false;index" json:"used"`
	CreatedAt time.Time `json:"created_at"`
}

func (o *OTPVerification) BeforeCreate(tx *gorm.DB) error {
	if o.ID == uuid.Nil {
		o.ID = uuid.New()
	}
	return nil
}
