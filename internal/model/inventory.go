package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type InventoryItemStatus string

const (
	InventoryAvailable InventoryItemStatus = "available"
	InventoryReserved  InventoryItemStatus = "reserved"
	InventorySold      InventoryItemStatus = "sold"
	InventoryVoid      InventoryItemStatus = "void"
)

type ItemType string

const (
	ItemTypeActivationURL   ItemType = "activation_url"   // One-time activation link (Gemini 18m, etc.)
	ItemTypeCredentialsPair ItemType = "credentials_pair" // Username / Password / 2FA secret
	ItemTypeLicenseKey      ItemType = "license_key"      // Promo code or key
)

type InventoryItem struct {
	ID        uuid.UUID           `gorm:"type:uuid;primaryKey" json:"id"`
	ProductID uuid.UUID           `gorm:"type:uuid;not null;index:idx_product_status" json:"product_id"`
	ItemType  ItemType            `gorm:"size:40;not null;default:'activation_url'" json:"item_type"`
	
	// Stored encrypted via AES-256-GCM
	PayloadEncrypted string              `gorm:"type:text;not null" json:"-"`
	
	Status           InventoryItemStatus `gorm:"size:30;not null;default:'available';index:idx_product_status" json:"status"`
	ReservedOrderID  *uuid.UUID          `gorm:"type:uuid;index" json:"reserved_order_id,omitempty"`
	ReservedUntil    *time.Time          `json:"reserved_until,omitempty"`
	SoldAt           *time.Time          `json:"sold_at,omitempty"`

	// Note or batch identifier for suppliers
	BatchTag string `gorm:"size:100" json:"batch_tag,omitempty"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (i *InventoryItem) BeforeCreate(tx *gorm.DB) error {
	if i.ID == uuid.Nil {
		i.ID = uuid.New()
	}
	return nil
}
