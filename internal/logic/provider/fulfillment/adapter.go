package fulfillment

import "net/http"

type FulfillmentAdapter interface {
	GetProviderKey() string
	GetProviderNameFA() string
	GetDefaultBaseURL() string
	GetBalance(client *http.Client, cfg ProviderConfig) (*ProviderBalanceDTO, error)
	GetCatalog(client *http.Client, cfg ProviderConfig) ([]ProviderProductDTO, error)
	CreateOrder(client *http.Client, cfg ProviderConfig, req ProviderOrderRequest) (*ProviderOrderResultDTO, error)
}
