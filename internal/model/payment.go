package model

import (
	"time"

	"gorm.io/gorm"
)

type PaymentPlatform string

const (
	PlatformBalance    PaymentPlatform = "balance"      // Default wallet balance payment
	PlatformZarinpal   PaymentPlatform = "zarinpal"     // ZarinPal Iranian PSP
	PlatformCardToCard PaymentPlatform = "card_to_card" // Manual bank transfer with receipt
	PlatformCrypto     PaymentPlatform = "crypto"       // USDT / TRX / TON
)

type Payment struct {
	Id          int64           `gorm:"primaryKey" json:"id"`
	Name        string          `gorm:"size:100;not null" json:"name"`
	NameFA      string          `gorm:"size:100;not null" json:"name_fa"`
	Platform    PaymentPlatform `gorm:"size:50;not null;index" json:"platform"`
	Icon        string          `gorm:"size:255" json:"icon"`
	Config      string          `gorm:"type:text;not null;default:'{}'" json:"config"` // JSON configuration (merchant_id, card_number, etc.)
	Description string          `gorm:"type:text" json:"description"`
	Enable      bool            `gorm:"not null;default:false;index" json:"enable"`
	Sort        int             `gorm:"default:0" json:"sort"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
	DeletedAt   gorm.DeletedAt  `gorm:"index" json:"-"`
}

func (*Payment) TableName() string {
	return "payment"
}
