package model

import (
	"time"

	"gorm.io/gorm"
)

type SystemSetting struct {
	ID        int64          `gorm:"primaryKey" json:"id"`
	Category  string         `gorm:"size:60;not null;index:idx_cat_key" json:"category"`       // e.g. "sms", "site", "telegram"
	Key       string         `gorm:"size:100;not null;uniqueIndex;index:idx_cat_key" json:"key"` // e.g. "sms_provider", "sms_config"
	Value     string         `gorm:"type:text;not null" json:"value"`                          // JSON or plain string
	Type      string         `gorm:"size:30;not null;default:'string'" json:"type"`             // "string", "json", "bool", "number"
	Desc      string         `gorm:"size:255" json:"desc"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (*SystemSetting) TableName() string {
	return "system"
}
