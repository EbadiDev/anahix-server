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
	ImageURL     string `gorm:"size:512" json:"image_url"`
	GalleryJSON  string `gorm:"type:text" json:"gallery_json"`

	// Flash Sale & Promotions
	IsFlashSale         bool       `gorm:"default:false;index" json:"is_flash_sale"`
	FlashSalePriceToman int64      `gorm:"default:0" json:"flash_sale_price_toman"`
	FlashSaleEndAt      *time.Time `json:"flash_sale_end_at,omitempty"`
	FlashSaleStock      int        `gorm:"default:0" json:"flash_sale_stock"`
	FlashSaleSold       int        `gorm:"default:0" json:"flash_sale_sold"`

	// Social proof & Merchandising
	IsBestSeller  bool    `gorm:"default:false;index" json:"is_best_seller"`
	IsFeatured    bool    `gorm:"default:false;index" json:"is_featured"`
	RatingAverage float64 `gorm:"type:decimal(3,2);default:5.0" json:"rating_average"`
	SalesCount    int     `gorm:"default:0" json:"sales_count"`
	BadgesJSON    string  `gorm:"type:text" json:"badges_json"`

	// Delivery Strategy & Group Routing Engine
	GroupTag            string `gorm:"size:80;index;default:'general'" json:"group_tag"`
	GroupTitleFA        string `gorm:"size:120" json:"group_title_fa"`
	DeliveryStrategy    string `gorm:"size:50;default:'auto'" json:"delivery_strategy"`
	ExternalProvider    string `gorm:"size:50;default:'third_party_api'" json:"external_provider"`
	ExternalProductID   int64  `gorm:"default:0;index" json:"external_product_id"`
	ExternalProductIDStr string `gorm:"size:120;index" json:"external_product_id_str"`
	ExternalProductName string `gorm:"size:255" json:"external_product_name"`

	// Input & Delivery Payload Modes
	RequiredCustomerInput string `gorm:"size:50;default:'none'" json:"required_customer_input"`
	DeliveredPayloadType  string `gorm:"size:50;default:'license_code'" json:"delivered_payload_type"`

	// Pre-Purchase Notice & Rich Content Interceptor
	PrePurchaseNotice     string `gorm:"type:text" json:"pre_purchase_notice"`
	PrePurchaseNoticeMode string `gorm:"size:30;default:'none'" json:"pre_purchase_notice_mode"`
	PrePurchaseSnoozeDays int    `gorm:"default:7" json:"pre_purchase_snooze_days"`

	// Merchandising & Stock Display Control
	StockDisplayMode      string `gorm:"size:30;default:'show_exact'" json:"stock_display_mode"`
	ParentCategorySlug    string `gorm:"size:100;index" json:"parent_category_slug"`
	CategorySlug          string `gorm:"size:100;index" json:"category_slug"`
	ShowSalesCount        bool   `gorm:"default:true" json:"show_sales_count"`
	ShowProgress          bool   `gorm:"default:true" json:"show_progress"`

	// PDP & SEO Enhancements
	MetaTitle       string `gorm:"size:255" json:"meta_title"`
	MetaDescription string `gorm:"type:text" json:"meta_description"`
	OgImageURL      string `gorm:"size:512" json:"og_image_url"`
	PlansJSON       string `gorm:"type:text" json:"plans_json"`
	ComparisonJSON  string `gorm:"type:text" json:"comparison_json"`
	GuideJSON       string `gorm:"type:text" json:"guide_json"`
	FaqJSON         string `gorm:"type:text" json:"faq_json"`
	FeaturesJSON    string `gorm:"type:text" json:"features_json"`

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
