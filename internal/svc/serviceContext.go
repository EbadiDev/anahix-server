package svc

import (
	"github.com/EbadiDev/anahix-server/internal/config"
	"github.com/EbadiDev/anahix-server/internal/logic/provider/fulfillment"
	"github.com/EbadiDev/anahix-server/internal/logic/provider/sms"
	"github.com/EbadiDev/anahix-server/internal/logic/provider/smsir"
	"github.com/EbadiDev/anahix-server/pkg/crypto"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config              *config.Config
	DB                  *gorm.DB
	Vault               *crypto.Vault
	SMSProvider         *sms.Client
	FulfillmentProvider *fulfillment.Client
	FulfillmentManager  *fulfillment.Manager
	SMSIRClient         *smsir.Client
}

func NewServiceContext(cfg *config.Config, db *gorm.DB) (*ServiceContext, error) {
	vault, err := crypto.NewVault(cfg.Vault.MasterKey)
	if err != nil {
		return nil, err
	}

	return &ServiceContext{
		Config:              cfg,
		DB:                  db,
		Vault:               vault,
		SMSProvider:         sms.NewClient(),
		FulfillmentProvider: fulfillment.NewClient(""),
		FulfillmentManager:  fulfillment.NewManager(),
		SMSIRClient:         smsir.NewClient(),
	}, nil
}
