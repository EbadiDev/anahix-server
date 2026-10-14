package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Referral struct {
	ID                 uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	ReferrerCustomerID uuid.UUID `gorm:"type:uuid;not null;index" json:"referrer_customer_id"`
	ReferredCustomerID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex" json:"referred_customer_id"`
	ReferralCode       string    `gorm:"size:32;not null;index" json:"referral_code"`
	CreatedAt          time.Time `json:"created_at"`
}

func (r *Referral) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}

func (Referral) TableName() string {
	return "referrals"
}

type ReferralCommissionStatus string

const (
	CommissionConfirmed           ReferralCommissionStatus = "CONFIRMED"
	CommissionTransferredToWallet ReferralCommissionStatus = "TRANSFERRED_TO_WALLET"
)

type ReferralCommission struct {
	ID                 uuid.UUID                `gorm:"type:uuid;primaryKey" json:"id"`
	ReferrerCustomerID uuid.UUID                `gorm:"type:uuid;not null;index" json:"referrer_customer_id"`
	ReferredCustomerID uuid.UUID                `gorm:"type:uuid;not null;index" json:"referred_customer_id"`
	OrderID            uuid.UUID                `gorm:"type:uuid;not null;index" json:"order_id"`
	OrderAmountToman   int64                    `gorm:"not null" json:"order_amount_toman"`
	CommissionToman    int64                    `gorm:"not null" json:"commission_toman"`
	CommissionRate     float64                  `gorm:"type:decimal(5,2);default:0.10" json:"commission_rate"`
	Status             ReferralCommissionStatus `gorm:"size:32;not null;default:'CONFIRMED';index" json:"status"`
	CreatedAt          time.Time                `json:"created_at"`
	UpdatedAt          time.Time                `json:"updated_at"`
}

func (rc *ReferralCommission) BeforeCreate(tx *gorm.DB) error {
	if rc.ID == uuid.Nil {
		rc.ID = uuid.New()
	}
	return nil
}

func (ReferralCommission) TableName() string {
	return "referral_commissions"
}
