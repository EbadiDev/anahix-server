package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Customer struct {
	ID           uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	Phone        string         `gorm:"size:30;uniqueIndex;not null" json:"phone"`
	FirstName    string         `gorm:"size:60" json:"first_name,omitempty"`
	LastName     string         `gorm:"size:60" json:"last_name,omitempty"`
	FullName     string         `gorm:"size:120" json:"full_name"`
	Email        string         `gorm:"size:120" json:"email"`
	BalanceToman         int64          `gorm:"not null;default:0" json:"balance_toman"`
	ReferralCode         string         `gorm:"size:32;uniqueIndex" json:"referral_code"`
	ReferralBalanceToman int64          `gorm:"not null;default:0" json:"referral_balance_toman"`
	IsActive             bool           `gorm:"default:true;index" json:"is_active"`
	IsAdmin              bool           `gorm:"default:false;index" json:"is_admin"`
	Tier                 string         `gorm:"size:30;default:'standard';index" json:"tier"` // standard, colleague, vip
	BlockReason          string         `gorm:"size:255" json:"block_reason,omitempty"`
	InternalNotes        string         `gorm:"type:text" json:"internal_notes,omitempty"`
	LastLoginAt          *time.Time     `json:"last_login_at,omitempty"`

	// Corporate / Legal Entity fields for official tax invoices
	IsCorporate    bool   `gorm:"default:false" json:"is_corporate"`
	CompanyName    string `gorm:"size:200" json:"company_name,omitempty"`
	NationalID     string `gorm:"size:30" json:"national_id,omitempty"`
	EconomicCode   string `gorm:"size:30" json:"economic_code,omitempty"`
	RegistrationNo string `gorm:"size:30" json:"registration_no,omitempty"`
	PostalCode     string `gorm:"size:20" json:"postal_code,omitempty"`
	CompanyAddress string `gorm:"size:300" json:"company_address,omitempty"`
	CompanyPhone   string `gorm:"size:30" json:"company_phone,omitempty"`

	CreatedAt            time.Time      `json:"created_at"`
	UpdatedAt            time.Time      `json:"updated_at"`
	DeletedAt            gorm.DeletedAt `gorm:"index" json:"-"`
}

func (c *Customer) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	if c.ReferralCode == "" {
		// Generate clean referral code e.g. ANX-A8F2
		u := uuid.New().String()
		c.ReferralCode = "ANX-" + u[:4]
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
