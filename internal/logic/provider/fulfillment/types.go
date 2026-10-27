package fulfillment

import "time"

const (
	ProviderInsightX      = "insightx"
	ProviderStackVault    = "stackvault"
	ProviderLootPaglu     = "paglu"
	ProviderJeroAccounts  = "jero"
	ProviderSupremeLeader = "supreme"
)

type ProviderConfig struct {
	Key           string  `json:"key"`
	NameFA        string  `json:"name_fa"`
	Enabled       bool    `json:"enabled"`
	APIKey        string  `json:"api_key"`
	BaseURL       string  `json:"base_url"`
	MarginPercent float64 `json:"margin_percent"`
}

type ProviderBalanceDTO struct {
	ProviderKey     string      `json:"provider_key"`
	ProviderNameFA  string      `json:"provider_name_fa"`
	BalanceUSD      float64     `json:"balance_usd"`
	BalanceToman    int64       `json:"balance_toman"`
	Status          string      `json:"status"`
	RateLimitPerMin int         `json:"rate_limit_per_min,omitempty"`
	Raw             interface{} `json:"raw,omitempty"`
}

type ProviderProductDTO struct {
	ID                  string  `json:"id"`
	NumericID           int64   `json:"numeric_id,omitempty"`
	Name                string  `json:"name"`
	Description         string  `json:"description,omitempty"`
	PriceUSD            float64 `json:"price_usd"`
	PriceToman          int64   `json:"price_toman"`
	Stock               int64   `json:"stock"`
	InStock             bool    `json:"in_stock"`
	Unlimited           bool    `json:"unlimited"`
	Emoji               string  `json:"emoji,omitempty"`
	ProviderKey         string  `json:"provider_key"`
	ProviderNameFA      string  `json:"provider_name_fa"`
	RequiredInput       string  `json:"required_input"`       // none, account_creds, telegram_username, wordpress_admin, custom_links
	DeliveredPayloadType string  `json:"delivered_payload_type"` // license_code, activation_link, credentials_pair, direct_service, files_or_downloads
	IsLinked            bool    `json:"is_linked"`
	LinkedProducts      []string `json:"linked_products,omitempty"`
}

type ProviderOrderRequest struct {
	ProductID        string `json:"product_id"`
	NumericID        int64  `json:"numeric_id,omitempty"`
	Quantity         int    `json:"quantity"`
	IdempotencyKey   string `json:"idempotency_key"`
	CustomerEmail    string `json:"customer_email,omitempty"`
	CustomerPassword string `json:"customer_password,omitempty"`
	TelegramUsername string `json:"telegram_username,omitempty"`
	WordPressSite    string `json:"wordpress_site,omitempty"`
	WordPressUser    string `json:"wordpress_user,omitempty"`
	WordPressPass    string `json:"wordpress_pass,omitempty"`
	CustomLinks      []string `json:"custom_links,omitempty"`
}

type ProviderOrderResultDTO struct {
	Success          bool                   `json:"success"`
	ProviderKey      string                 `json:"provider_key"`
	OrderID          string                 `json:"order_id"`
	Codes            []string               `json:"codes,omitempty"`
	Links            []string               `json:"links,omitempty"`
	Credentials      map[string]string      `json:"credentials,omitempty"`
	DeliveredPayload string                 `json:"delivered_payload"`
	PayloadType      string                 `json:"payload_type"`
	Instructions     string                 `json:"instructions,omitempty"`
	UnitCostUSD      float64                `json:"unit_cost_usd"`
	TotalCostUSD     float64                `json:"total_cost_usd"`
	RawJSON          string                 `json:"raw_json,omitempty"`
	Error            string                 `json:"error,omitempty"`
}

type CachedCatalog struct {
	Items     []ProviderProductDTO
	UpdatedAt time.Time
}
