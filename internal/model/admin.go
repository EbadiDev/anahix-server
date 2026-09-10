package model

import (
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AdminRole string

const (
	RoleSuperAdmin AdminRole = "superadmin"
	RoleOperator   AdminRole = "operator"
	RoleSupport    AdminRole = "support"
)

type AdminUser struct {
	ID           uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	Username     string         `gorm:"size:60;uniqueIndex;not null" json:"username"`
	PasswordHash string         `gorm:"size:255;not null" json:"-"`
	FullName     string         `gorm:"size:120" json:"full_name"`
	Role         AdminRole      `gorm:"size:30;not null;default:'operator'" json:"role"`
	IsActive     bool           `gorm:"default:true;index" json:"is_active"`
	LastLoginAt  *time.Time     `json:"last_login_at,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

func (a *AdminUser) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}

func (a *AdminUser) SetPassword(plain string) error {
	bytes, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	a.PasswordHash = string(bytes)
	return nil
}

func (a *AdminUser) CheckPassword(plain string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(a.PasswordHash), []byte(plain))
	return err == nil
}
