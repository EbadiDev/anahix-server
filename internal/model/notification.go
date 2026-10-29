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

type SystemAnnouncement struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	Title       string         `gorm:"size:255;not null" json:"title"`
	Content     string         `gorm:"type:text;not null" json:"content"`
	ImageURL    string         `gorm:"size:512" json:"image_url,omitempty"`
	TargetPages string         `gorm:"size:255;default:'all'" json:"target_pages"` // "all", "/", "/category", "/products", "/profile"
	DisplayType string         `gorm:"size:50;default:'bell_and_popup'" json:"display_type"` // "bell_only", "popup_first_load", "banner_top", "bell_and_popup"
	Severity    string         `gorm:"size:30;default:'info'" json:"severity"` // "info", "discount", "warning", "success"
	IsActive    bool           `gorm:"default:true;index" json:"is_active"`
	ExpiresAt   *time.Time     `json:"expires_at,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (a *SystemAnnouncement) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}

type StockAlert struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	ProductID string         `gorm:"size:100;index;not null" json:"product_id"`
	Mobile    string         `gorm:"size:30;not null" json:"mobile"`
	Email     string         `gorm:"size:255" json:"email"`
	Notified  bool           `gorm:"default:false;index" json:"notified"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (s *StockAlert) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

type UserFavorite struct {
	ID         uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	CustomerID uuid.UUID      `gorm:"type:uuid;index;not null" json:"customer_id"`
	TargetType string         `gorm:"size:30;not null" json:"target_type"` // "product", "article", "guide"
	TargetID   string         `gorm:"size:255;not null" json:"target_id"` // slug or id
	Title      string         `gorm:"size:255" json:"title"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

func (f *UserFavorite) BeforeCreate(tx *gorm.DB) error {
	if f.ID == uuid.Nil {
		f.ID = uuid.New()
	}
	return nil
}
