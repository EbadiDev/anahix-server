package svc

import (
	"github.com/EbadiDev/anahix-server/internal/config"
	"github.com/EbadiDev/anahix-server/pkg/crypto"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config *config.Config
	DB     *gorm.DB
	Vault  *crypto.Vault
}

func NewServiceContext(cfg *config.Config, db *gorm.DB) (*ServiceContext, error) {
	vault, err := crypto.NewVault(cfg.Vault.MasterKey)
	if err != nil {
		return nil, err
	}

	return &ServiceContext{
		Config: cfg,
		DB:     db,
		Vault:  vault,
	}, nil
}
