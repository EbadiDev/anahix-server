package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// LoyaltyTier represents a dynamic VIP & customer loyalty tier stored in database
type LoyaltyTier struct {
	ID              uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	Code            string         `gorm:"size:40;uniqueIndex;not null" json:"code"` // bronze, silver, gold, diamond, etc.
	Name            string         `gorm:"size:80;not null" json:"name"`             // برنزی (عادی), نقره‌ای, طلایی (ریسلرها)
	NameEN          string         `gorm:"size:80" json:"name_en"`                  // Bronze, Silver, Gold
	MinSpendToman   int64          `gorm:"not null;default:0" json:"min_spend_toman"` // Minimum purchase spend in Toman required
	DiscountPercent float64        `gorm:"type:decimal(5,2);not null;default:0" json:"discount_percent"` // 0.0, 2.0, 5.0, etc.
	QueuePriority   int            `gorm:"default:0;index" json:"queue_priority"`    // 0: normal, 10: high, 50: urgent VIP
	BadgeColor      string         `gorm:"size:30;default:'#94a3b8'" json:"badge_color"` // HEX color (e.g. #cd7f32, #94a3b8, #eab308)
	BadgeGradient   string         `gorm:"size:120" json:"badge_gradient"`          // CSS gradient or styling class
	Icon            string         `gorm:"size:40;default:'medal'" json:"icon"`      // medal, star, crown, diamond, zap
	Description     string         `gorm:"type:text" json:"description"`
	PerksJSON       string         `gorm:"type:text" json:"perks_json"`              // JSON array of perk strings e.g. ["تخفیف ۲٪", "اولویت صف"]
	OrderIndex      int            `gorm:"default:0" json:"order_index"`
	IsActive        bool           `gorm:"default:true;index" json:"is_active"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

func (t *LoyaltyTier) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

func (LoyaltyTier) TableName() string {
	return "loyalty_tiers"
}
