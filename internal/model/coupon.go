package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DiscountType string

const (
	DiscountTypePercentage DiscountType = "percentage"
	DiscountTypeFixed      DiscountType = "fixed"
)

type Coupon struct {
	ID                  uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	Code                string         `gorm:"size:64;uniqueIndex;not null" json:"code"`
	DiscountType        DiscountType   `gorm:"size:20;not null;default:'percentage'" json:"discount_type"`
	DiscountValue       int64          `gorm:"not null" json:"discount_value"`
	MinOrderAmountToman int64          `gorm:"default:0" json:"min_order_amount_toman"`
	MaxDiscountToman    int64          `gorm:"default:0" json:"max_discount_toman"`
	PerUserLimit        int            `gorm:"default:1" json:"per_user_limit"`
	MaxUsage            int            `gorm:"default:0" json:"max_usage"`
	CurrentUsage        int            `gorm:"default:0" json:"current_usage"`
	Category            string         `gorm:"size:50;default:'all'" json:"category"`
	ExpiresAt           *time.Time     `json:"expires_at,omitempty"`
	IsActive            bool           `gorm:"default:true;index" json:"is_active"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	DeletedAt           gorm.DeletedAt `gorm:"index" json:"-"`
}

func (c *Coupon) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

func (Coupon) TableName() string {
	return "coupons"
}

// CouponUsage tracks which user consumed which coupon
type CouponUsage struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	CouponID      uuid.UUID `gorm:"type:uuid;not null;index" json:"coupon_id"`
	CustomerPhone string    `gorm:"size:32;not null;index" json:"customer_phone"`
	OrderID       uuid.UUID `gorm:"type:uuid;not null;index" json:"order_id"`
	DiscountToman int64     `gorm:"not null" json:"discount_toman"`
	CreatedAt     time.Time `json:"created_at"`
}

func (u *CouponUsage) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

func (CouponUsage) TableName() string {
	return "coupon_usages"
}
