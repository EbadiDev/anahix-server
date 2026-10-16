package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BannerPosition string

const (
	BannerPositionMainSlider BannerPosition = "main_slider" // Main hero carousel
	BannerPositionHeroSide   BannerPosition = "hero_side"   // Alongside hero
	BannerPositionMiddle     BannerPosition = "middle_strip" // Middle promotional strip
	BannerPositionFlashSale  BannerPosition = "flash_sale"  // Flash sale promotion banner
)

type Banner struct {
	ID         uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	Title      string         `gorm:"size:255;not null" json:"title"`
	Subtitle   string         `gorm:"size:255" json:"subtitle"`
	BadgeText  string         `gorm:"size:100" json:"badge_text"`
	ImageURL   string         `gorm:"size:512;not null" json:"image_url"`
	LinkURL    string         `gorm:"size:512" json:"link_url"`
	Position   BannerPosition `gorm:"size:50;default:'main_slider';index" json:"position"`
	BgGradient string         `gorm:"size:100;default:'from-blue-600 via-indigo-600 to-purple-700'" json:"bg_gradient"`
	ButtonText string         `gorm:"size:100;default:'مشاهده و خرید'" json:"button_text"`
	SortOrder  int            `gorm:"default:0;index" json:"sort_order"`
	IsActive   bool           `gorm:"default:true;index" json:"is_active"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

func (b *Banner) BeforeCreate(tx *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}
