package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DeliveryType string

const (
	DeliveryInstant     DeliveryType = "instant"      // 0-second auto stock allocation
	DeliveryManualTimed DeliveryType = "manual_timed" // Admin/operator manual activation
)

type ProductCategory string

const (
	CategoryAILink        ProductCategory = "ai_link"        // Activation links (Gemini 18m, etc.)
	CategoryAIAccount     ProductCategory = "ai_account"     // Ready-made accounts
	CategoryPersonalUpgrade ProductCategory = "personal_upgrade" // On-account upgrade (ChatGPT Plus on user email)
	CategoryAPIKey        ProductCategory = "api_key"        // API keys / Credits
	CategoryDigitalCode   ProductCategory = "digital_code"   // License / Gift cards
)

type Product struct {
	ID             uuid.UUID       `gorm:"type:uuid;primaryKey" json:"id"`
	Title          string          `gorm:"size:255;not null" json:"title"`
	TitleFA        string          `gorm:"size:255;not null" json:"title_fa"`
	Slug           string          `gorm:"size:120;uniqueIndex;not null" json:"slug"`
	Description    string          `gorm:"type:text" json:"description"`
	Category       ProductCategory `gorm:"size:50;not null;index" json:"category"`
	DeliveryType   DeliveryType    `gorm:"size:30;not null;default:'instant'" json:"delivery_type"`
	SLAMinMinutes  int             `gorm:"default:0" json:"sla_min_minutes"`
	SLAMaxMinutes  int             `gorm:"default:0" json:"sla_max_minutes"`
	SLADisplayText string          `gorm:"size:255" json:"sla_display_text"`
	OperatingHours string          `gorm:"size:100;default:'09:00 - 23:00'" json:"operating_hours"`

	// Reseller pricing
	BaseCostUSD float64 `gorm:"type:decimal(10,2);default:0" json:"base_cost_usd"`
	PriceToman  int64   `gorm:"not null" json:"price_toman"` // In Iranian Toman (e.g. 1,450,000)

	IsActive     bool   `gorm:"default:true;index" json:"is_active"`
	Instructions string `gorm:"type:text" json:"instructions"`
	IconURL      string `gorm:"size:512" json:"icon_url"`
	BannerURL    string `gorm:"size:512" json:"banner_url"`

	// Computed stock count for instant delivery items
	AvailableStock int64 `gorm:"-" json:"available_stock"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (p *Product) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}
