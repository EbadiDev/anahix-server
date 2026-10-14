package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type VirtualNumberStatus string

const (
	VNumStatusPending  VirtualNumberStatus = "PENDING"
	VNumStatusReceived VirtualNumberStatus = "RECEIVED"
	VNumStatusFinished VirtualNumberStatus = "FINISHED"
	VNumStatusCanceled VirtualNumberStatus = "CANCELED"
	VNumStatusTimeout  VirtualNumberStatus = "TIMEOUT"
)

type VirtualNumberOrder struct {
	ID             uuid.UUID           `gorm:"type:uuid;primary_key" json:"id"`
	OrderNo        string              `gorm:"size:64;uniqueIndex;not null" json:"order_no"`
	ProviderOrderID int64               `gorm:"index" json:"provider_order_id"`
	FiveSimOrderID int64               `gorm:"index" json:"fivesim_order_id,omitempty"`
	Phone          string              `gorm:"size:32;not null" json:"phone"`
	Country        string              `gorm:"size:64;not null" json:"country"`
	CountryNameFA  string              `gorm:"size:64" json:"country_name_fa"`
	CountryFlag    string              `gorm:"size:16" json:"country_flag"`
	Service        string              `gorm:"size:64;not null" json:"service"`
	ServiceNameFA  string              `gorm:"size:64" json:"service_name_fa"`
	Operator       string              `gorm:"size:64;default:'any'" json:"operator"`
	BaseCost       float64             `gorm:"type:decimal(10,2);default:0" json:"base_cost"`
	PriceToman     int64               `gorm:"not null" json:"price_toman"`
	Status         VirtualNumberStatus `gorm:"size:32;not null;default:'PENDING'" json:"status"`
	SMSCode        string              `gorm:"size:32" json:"sms_code"`
	SMSText        string              `gorm:"type:text" json:"sms_text"`
	CustomerPhone  string              `gorm:"size:32" json:"customer_phone"`
	ExpiresAt      time.Time           `json:"expires_at"`
	CreatedAt      time.Time           `json:"created_at"`
	UpdatedAt      time.Time           `json:"updated_at"`
	DeletedAt      gorm.DeletedAt      `gorm:"index" json:"-"`
}

func (VirtualNumberOrder) TableName() string {
	return "virtual_number_orders"
}
