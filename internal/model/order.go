package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OrderStatus string

const (
	OrderPendingPayment         OrderStatus = "pending_payment"
	OrderPaid                   OrderStatus = "paid"
	OrderInQueue                OrderStatus = "in_queue"
	OrderProcessing             OrderStatus = "processing"
	OrderAwaitingCustomerAction OrderStatus = "awaiting_customer_action"
	OrderCompleted              OrderStatus = "completed"
	OrderCancelled              OrderStatus = "cancelled"
	OrderRefunded               OrderStatus = "refunded"
	OrderExpired                OrderStatus = "expired"
)

type TwoFactorStatus string

const (
	TwoFactorIdle      TwoFactorStatus = "idle"
	TwoFactorRequested TwoFactorStatus = "requested"
	TwoFactorSubmitted TwoFactorStatus = "submitted"
	TwoFactorVerified  TwoFactorStatus = "verified"
	TwoFactorExpired   TwoFactorStatus = "expired"
)

type PaymentMethod string

const (
	PaymentMethodBalance    PaymentMethod = "balance"      // Default wallet balance
	PaymentMethodZarinpal   PaymentMethod = "zarinpal"     // Zarinpal gateway
	PaymentMethodCardToCard PaymentMethod = "card_to_card" // Manual card to card transfer
	PaymentMethodCrypto     PaymentMethod = "crypto"       // USDT / TRX crypto
)

type Order struct {
	ID          uuid.UUID    `gorm:"type:uuid;primaryKey" json:"id"`
	OrderNumber string       `gorm:"size:64;uniqueIndex;not null" json:"order_number"`
	
	CustomerPhone string     `gorm:"size:30;index" json:"customer_phone"`
	CustomerEmail string     `gorm:"size:120;index" json:"customer_email"`
	CustomerID    *uuid.UUID `gorm:"type:uuid;index" json:"customer_id,omitempty"`
	AccessToken   string     `gorm:"size:64;uniqueIndex;not null" json:"access_token"`

	Status       OrderStatus  `gorm:"size:35;not null;default:'pending_payment';index" json:"status"`
	DeliveryType DeliveryType `gorm:"size:30;not null" json:"delivery_type"`

	TotalAmountToman    int64         `gorm:"not null" json:"total_amount_toman"`
	SubtotalAmountToman int64         `gorm:"not null;default:0" json:"subtotal_amount_toman"`
	DiscountAmountToman int64         `gorm:"not null;default:0" json:"discount_amount_toman"`
	LoyaltyDiscountToman int64        `gorm:"not null;default:0" json:"loyalty_discount_toman"`
	CustomerTier        string        `gorm:"size:40;default:'bronze';index" json:"customer_tier"`
	PriorityScore       int           `gorm:"default:0;index" json:"priority_score"`
	CouponCode          string        `gorm:"size:64" json:"coupon_code,omitempty"`
	ReferralCode        string        `gorm:"size:32;index" json:"referral_code,omitempty"`
	PaymentMethod       PaymentMethod `gorm:"size:30;not null" json:"payment_method"`
	PaymentRefID        string        `gorm:"size:120" json:"payment_ref_id,omitempty"`
	PaymentAuthority    string        `gorm:"size:120;index" json:"payment_authority,omitempty"`

	PaidAt      *time.Time `json:"paid_at,omitempty"`
	ExpiresAt   time.Time  `gorm:"not null;index" json:"expires_at"` // Payment expiration (e.g. 20 min)
	CompletedAt *time.Time `json:"completed_at,omitempty"`

	// On-Account Upgrade (Mode C) Fields
	// Encrypted user account credentials (e.g., user's personal OpenAI email/password)
	CustomerCredentialsEncrypted string `gorm:"type:text" json:"-"`
	
	// Live 2FA Negotiation Engine
	TwoFactorPrompt      string          `gorm:"size:255" json:"two_factor_prompt,omitempty"`
	TwoFactorCode        string          `gorm:"size:30" json:"two_factor_code,omitempty"`
	TwoFactorStatus      TwoFactorStatus `gorm:"size:30;default:'idle'" json:"two_factor_status"`
	TwoFactorRequestedAt *time.Time      `json:"two_factor_requested_at,omitempty"`

	// Operator SLA and internal notes
	AssignedOperatorID *uuid.UUID `gorm:"type:uuid" json:"assigned_operator_id,omitempty"`
	OperatorNotes      string     `gorm:"type:text" json:"operator_notes,omitempty"`
	CardReceiptURL     string     `gorm:"size:512" json:"card_receipt_url,omitempty"`

	// Official Tax Invoice & Corporate Billing Fields
	IsCorporate    bool   `gorm:"default:false" json:"is_corporate"`
	CompanyName    string `gorm:"size:200" json:"company_name,omitempty"`
	NationalID     string `gorm:"size:30" json:"national_id,omitempty"`
	EconomicCode   string `gorm:"size:30" json:"economic_code,omitempty"`
	RegistrationNo string `gorm:"size:30" json:"registration_no,omitempty"`
	PostalCode     string `gorm:"size:20" json:"postal_code,omitempty"`
	BuyerAddress   string `gorm:"size:300" json:"buyer_address,omitempty"`
	BuyerPhone     string `gorm:"size:30" json:"buyer_phone,omitempty"`
	TaxAmountToman int64  `gorm:"default:0" json:"tax_amount_toman"`

	Items []OrderItem `gorm:"foreignKey:OrderID" json:"items"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (o *Order) BeforeCreate(tx *gorm.DB) error {
	if o.ID == uuid.Nil {
		o.ID = uuid.New()
	}
	return nil
}

type OrderItem struct {
	ID              uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	OrderID         uuid.UUID  `gorm:"type:uuid;not null;index" json:"order_id"`
	ProductID       uuid.UUID  `gorm:"type:uuid;not null;index" json:"product_id"`
	InventoryItemID *uuid.UUID `gorm:"type:uuid" json:"inventory_item_id,omitempty"`
	
	ProductTitle    string     `gorm:"size:255;not null" json:"product_title"`
	ProductSlug     string     `gorm:"size:120;not null" json:"product_slug"`
	DeliveryType    DeliveryType `gorm:"size:30;not null" json:"delivery_type"`
	UnitPriceToman  int64        `gorm:"not null" json:"unit_price_toman"`
	Quantity        int          `gorm:"default:1;not null" json:"quantity"`
	SubtotalToman   int64        `gorm:"default:0" json:"subtotal_toman"`
	CostUSD         float64      `gorm:"type:decimal(10,2);default:0" json:"cost_usd"`
	CostToman       int64        `gorm:"default:0" json:"cost_toman"`
	
	// Delivered payload & external tracking
	DeliveredPayload     string `gorm:"type:text" json:"delivered_payload,omitempty"`
	DeliveredPayloadType string `gorm:"size:50;default:'license_code'" json:"delivered_payload_type,omitempty"`
	ExternalOrderID      string `gorm:"size:120" json:"external_order_id,omitempty"`
	ExternalProvider     string `gorm:"size:50" json:"external_provider,omitempty"`
	CustomerInputData    string `gorm:"type:text" json:"customer_input_data,omitempty"`
	Instructions         string `gorm:"type:text" json:"instructions,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (oi *OrderItem) BeforeCreate(tx *gorm.DB) error {
	if oi.ID == uuid.Nil {
		oi.ID = uuid.New()
	}
	return nil
}
