package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// StoreCategory represents hierarchical product categories (e.g. AI -> Chatbots -> Gemini)
type StoreCategory struct {
	ID          uuid.UUID       `gorm:"type:uuid;primaryKey" json:"id"`
	ParentID    *uuid.UUID      `gorm:"type:uuid;index" json:"parent_id,omitempty"`
	TitleFA     string          `gorm:"size:120;not null" json:"title_fa"`
	TitleEN     string          `gorm:"size:120" json:"title_en"`
	Slug        string          `gorm:"size:100;uniqueIndex;not null" json:"slug"`
	Description string          `gorm:"type:text" json:"description"`
	IconURL     string          `gorm:"size:512" json:"icon_url"`
	SortOrder   int             `gorm:"default:0" json:"sort_order"`
	IsActive    bool            `gorm:"default:true;index" json:"is_active"`
	Children    []StoreCategory `gorm:"foreignKey:ParentID" json:"children,omitempty"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (c *StoreCategory) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}
