package types

import (
	"time"

	"github.com/EbadiDev/anahix-server/internal/model"
)

// Public Storefront DTOs

type ProductListItem struct {
	ID             string  `json:"id"`
	Title          string  `json:"title"`
	TitleFA        string  `json:"title_fa"`
	Slug           string  `json:"slug"`
	Category       string  `json:"category"`
	DeliveryType   string  `json:"delivery_type"`
	PriceToman     int64   `json:"price_toman"`
	BaseCostUSD    float64 `json:"base_cost_usd,omitempty"`
	SLAMinMinutes  int     `json:"sla_min_minutes"`
	SLAMaxMinutes  int     `json:"sla_max_minutes"`
	SLADisplayText string  `json:"sla_display_text"`
	OperatingHours string  `json:"operating_hours"`
	IconURL        string  `json:"icon_url"`
	BannerURL      string  `json:"banner_url"`
	AvailableStock int64   `json:"available_stock"`
	IsActive       bool    `json:"is_active"`
}

type ProductDetail struct {
	ProductListItem
	Description  string `json:"description"`
	Instructions string `json:"instructions"`
}

type PaymentMethodItem struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	NameFA      string `json:"name_fa"`
	Platform    string `json:"platform"`
	Icon        string `json:"icon"`
	Description string `json:"description"`
}

type CreateOrderRequest struct {
	ProductID     string `json:"product_id" binding:"required"`
	CustomerPhone string `json:"customer_phone" binding:"required"`
	CustomerEmail string `json:"customer_email"`
	PaymentMethod string `json:"payment_method" binding:"required"`

	// Optional fields for Timed / On-Account Upgrade (Mode C)
	CustomerAccountEmail    string `json:"customer_account_email"`
	CustomerAccountPassword string `json:"customer_account_password"`
	CustomerNotes           string `json:"customer_notes"`
}

type CreateOrderResponse struct {
	OrderID          string    `json:"order_id"`
	OrderNumber      string    `json:"order_number"`
	AccessToken      string    `json:"access_token"`
	TotalAmountToman int64     `json:"total_amount_toman"`
	PaymentMethod    string    `json:"payment_method"`
	PaymentURL       string    `json:"payment_url"`
	ExpiresAt        time.Time `json:"expires_at"`
}

type DeliveredItemDTO struct {
	ProductTitle     string `json:"product_title"`
	ProductSlug      string `json:"product_slug"`
	DeliveryType     string `json:"delivery_type"`
	Payload          string `json:"payload"` // Decrypted activation link or credentials
	Instructions     string `json:"instructions"`
}

type OrderDetailResponse struct {
	OrderID          string             `json:"order_id"`
	OrderNumber      string             `json:"order_number"`
	Status           string             `json:"status"`
	DeliveryType     string             `json:"delivery_type"`
	TotalAmountToman int64              `json:"total_amount_toman"`
	CustomerPhone    string             `json:"customer_phone"`
	PaymentMethod    string             `json:"payment_method"`
	CreatedAt        time.Time          `json:"created_at"`
	PaidAt           *time.Time         `json:"paid_at,omitempty"`
	CompletedAt      *time.Time         `json:"completed_at,omitempty"`
	DeliveredItems   []DeliveredItemDTO `json:"delivered_items,omitempty"`
	
	// Live 2FA Negotiation for timed personal upgrades
	TwoFactorPrompt      string     `json:"two_factor_prompt,omitempty"`
	TwoFactorStatus      string     `json:"two_factor_status"`
	TwoFactorRequestedAt *time.Time `json:"two_factor_requested_at,omitempty"`
}

type SubmitTwoFactorRequest struct {
	TwoFactorCode string `json:"two_factor_code"`
	Code          string `json:"code"`
}

// Admin / Operator DTOs

type AdminLoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type AdminLoginResponse struct {
	Token    string `json:"token"`
	Username string `json:"username"`
	FullName string `json:"full_name"`
	Role     string `json:"role"`
}

type BulkImportInventoryRequest struct {
	ProductID string   `json:"product_id" binding:"required"`
	ItemType  string   `json:"item_type" binding:"required"` // "activation_url", "credentials_pair", "license_key"
	Items     []string `json:"items" binding:"required,min=1"`
	BatchTag  string   `json:"batch_tag"`
}

type BulkImportInventoryResponse struct {
	ImportedCount int `json:"imported_count"`
}

type RequestTwoFactorRequest struct {
	PromptMessage string `json:"prompt_message" binding:"required"`
}

type CompleteOrderRequest struct {
	OperatorNotes string `json:"operator_notes"`
}

type UpdateExchangeRateRequest struct {
	USDtoTomanRate             int64   `json:"usd_to_toman_rate" binding:"required"`
	DefaultProfitMarginPercent float64 `json:"default_profit_margin_percent" binding:"required"`
	FixedFeeToman              int64   `json:"fixed_fee_toman"`
	AutoFetchEnabled           bool    `json:"auto_fetch_enabled"`
}

// Virtual Numbers Admin Settings DTOs

type VirtualNumberSettingsResponse struct {
	Enabled          bool    `json:"enabled"`
	APIKeyMasked     string  `json:"api_key_masked"`
	HasAPIKey        bool    `json:"has_api_key"`
	MarginPercent    float64 `json:"margin_percent"`
	USDRate          int64   `json:"usd_rate"`
	RUBRate          int64   `json:"rub_rate"`
	DefaultOperator  string  `json:"default_operator"`
	RatePriority     string  `json:"rate_priority"`
	LiveUSDRate      int64   `json:"live_usd_rate"`
	EffectiveUSDRate int64   `json:"effective_usd_rate"`
	ActiveRateSource string  `json:"active_rate_source"`
	Balance          float64 `json:"balance"`
	Rating           float64 `json:"rating"`
	StatusMessage    string  `json:"status_message"`
	IsMock           bool    `json:"is_mock"`
}

type FiveSimSettingsResponse = VirtualNumberSettingsResponse

type UpdateVirtualNumberSettingsRequest struct {
	Enabled         *bool   `json:"enabled"`
	APIKey          string  `json:"api_key"`
	MarginPercent   float64 `json:"margin_percent"`
	USDRate         int64   `json:"usd_rate"`
	RUBRate         int64   `json:"rub_rate"`
	DefaultOperator string  `json:"default_operator"`
	RatePriority    string  `json:"rate_priority"`
}

type UpdateFiveSimSettingsRequest = UpdateVirtualNumberSettingsRequest

// Public Virtual Numbers DTOs

type VirtualNumberServiceItem struct {
	Slug          string  `json:"slug"`
	Name          string  `json:"name"`
	NameFA        string  `json:"name_fa"`
	Category      string  `json:"category"`
	TotalQty      int     `json:"total_qty"`
	MinPriceUSD   float64 `json:"min_price_usd"`
	MinPriceToman int64   `json:"min_price_toman"`
	Icon          string  `json:"icon"`
	IsPopular     bool    `json:"is_popular"`
}

type VirtualNumberOperatorItem struct {
	Operator   string  `json:"operator"`
	CostUSD    float64 `json:"cost_usd"`
	PriceToman int64   `json:"price_toman"`
	Count      int     `json:"count"`
	Rate       float64 `json:"rate"`
}

type VirtualNumberCountryItem struct {
	Country            string                      `json:"country"`
	CountryNameFA      string                      `json:"country_name_fa"`
	CountryFlag        string                      `json:"country_flag"`
	TotalStock         int                         `json:"total_stock"`
	BestRate           float64                     `json:"best_rate"`
	CheapestPriceToman int64                       `json:"cheapest_price_toman"`
	CheapestPriceUSD   float64                     `json:"cheapest_price_usd"`
	DefaultOperator    string                      `json:"default_operator"`
	Operators          []VirtualNumberOperatorItem `json:"operators"`
}

type VirtualNumberCatalogItem struct {
	Service            string  `json:"service"`
	ServiceNameFA      string  `json:"service_name_fa"`
	Country            string  `json:"country"`
	CountryNameFA      string  `json:"country_name_fa"`
	CountryFlag        string  `json:"country_flag"`
	BaseCost           float64 `json:"base_cost"`
	PriceToman         int64   `json:"price_toman"`
	OriginalPriceToman int64   `json:"original_price_toman"`
	AvailableStock     int     `json:"available_stock"`
	SuccessRatePercent int     `json:"success_rate_percent"`
	Icon               string  `json:"icon"`
}

type BuyVirtualNumberRequest struct {
	Country       string `json:"country" binding:"required"`
	Service       string `json:"service" binding:"required"`
	Operator      string `json:"operator"`
	CustomerPhone string `json:"customer_phone" binding:"required"`
	PaymentMethod string `json:"payment_method"`
}

type VirtualNumberOrderResponse struct {
	ID               string    `json:"id"`
	OrderNo          string    `json:"order_no"`
	Phone            string    `json:"phone"`
	Country          string    `json:"country"`
	CountryNameFA    string    `json:"country_name_fa"`
	CountryFlag      string    `json:"country_flag"`
	Service          string    `json:"service"`
	ServiceNameFA    string    `json:"service_name_fa"`
	PriceToman       int64     `json:"price_toman"`
	Status           string    `json:"status"`
	SMSCode          string    `json:"sms_code"`
	SMSText          string    `json:"sms_text"`
	ExpiresAt        time.Time `json:"expires_at"`
	SecondsRemaining int       `json:"seconds_remaining"`
	CreatedAt        time.Time `json:"created_at"`
}

// Delivery Engine & Product Group DTOs

type ProductGroupItemDTO struct {
	ID                  string  `json:"id"`
	Title               string  `json:"title"`
	TitleFA             string  `json:"title_fa"`
	Slug                string  `json:"slug"`
	PriceToman          int64   `json:"price_toman"`
	DeliveryType        string  `json:"delivery_type"`
	DeliveryStrategy    string  `json:"delivery_strategy"`
	ExternalProvider    string  `json:"external_provider"`
	ExternalProductID   int64   `json:"external_product_id"`
	ExternalProductName string  `json:"external_product_name"`
	IsActive            bool    `json:"is_active"`
}

type ProductGroupDTO struct {
	GroupTag                string                `json:"group_tag"`
	GroupTitleFA            string                `json:"group_title_fa"`
	DeliveryStrategy        string                `json:"delivery_strategy"` // "manual_operator" | "third_party_instant"
	LinkedProviderProductID int64                 `json:"linked_provider_product_id,omitempty"`
	LinkedProviderName      string                `json:"linked_provider_name,omitempty"`
	ProductsCount           int                   `json:"products_count"`
	Products                []ProductGroupItemDTO `json:"products"`
}

type UpdateGroupStrategyRequest struct {
	GroupTag         string `json:"group_tag" binding:"required"`
	DeliveryStrategy string `json:"delivery_strategy" binding:"required"`
}

type LinkProductProviderRequest struct {
	ProductID         string `json:"product_id" binding:"required"`
	ExternalProductID int64  `json:"external_product_id"`
	DeliveryStrategy  string `json:"delivery_strategy"`
}

type LinkProductInsightXRequest = LinkProductProviderRequest

type InstantDeliverySettingsResponse struct {
	Enabled          bool    `json:"enabled"`
	APIKeyMasked     string  `json:"api_key_masked"`
	HasAPIKey        bool    `json:"has_api_key"`
	BalanceUSDT      float64 `json:"balance_usdt"`
	BalanceToman     int64   `json:"balance_toman"`
	EffectiveUSDRate int64   `json:"effective_usd_rate"`
	MarginPercent    float64 `json:"margin_percent"`
	RateLimit        int     `json:"rate_limit"`
	StatusMessage    string  `json:"status_message"`
}

type InsightXSettingsResponse = InstantDeliverySettingsResponse

type UpdateInstantDeliverySettingsRequest struct {
	Enabled       *bool    `json:"enabled"`
	APIKey        string   `json:"api_key"`
	MarginPercent *float64 `json:"margin_percent"`
}

type UpdateInsightXSettingsRequest = UpdateInstantDeliverySettingsRequest

type ProviderCatalogItemDTO struct {
	ID             int64    `json:"id"`
	Name           string   `json:"name"`
	Emoji          string   `json:"emoji"`
	Description    string   `json:"description"`
	PriceUSDT      float64  `json:"price_usdt"`
	PriceToman     int64    `json:"price_toman"`
	Stock          int64    `json:"stock"`
	Available      bool     `json:"available"`
	Unlimited      bool     `json:"unlimited"`
	IsLinked       bool     `json:"is_linked"`
	LinkedProducts []string `json:"linked_products,omitempty"`
}

type InsightXProductDTO = ProviderCatalogItemDTO

// Audit and Transaction Logs DTOs

type TransactionListResponse struct {
	Total int64               `json:"total"`
	Items []model.Transaction `json:"items"`
}

type AdminActionLogListResponse struct {
	Total int64                  `json:"total"`
	Items []model.AdminActionLog `json:"items"`
}

// Customer Auth DTOs

type SendOTPRequest struct {
	Mobile string `json:"mobile" binding:"required"`
}

type SendOTPResponse struct {
	Sent      bool   `json:"sent"`
	Cooldown  int    `json:"cooldown"`
	ExpiresIn int    `json:"expires_in"`
	Message   string `json:"message"`
}

type VerifyOTPRequest struct {
	Mobile string `json:"mobile" binding:"required"`
	Code   string `json:"code" binding:"required"`
}

type VerifyOTPResponse struct {
	Token    string                  `json:"token"`
	Customer CustomerProfileResponse `json:"customer"`
}

type CustomerProfileResponse struct {
	ID           string    `json:"id"`
	Phone        string    `json:"phone"`
	Mobile       string    `json:"mobile"`
	FullName     string    `json:"full_name"`
	Email        string    `json:"email"`
	BalanceToman int64     `json:"balance_toman"`
	Balance      int64     `json:"balance"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
}

// SMS Admin Settings DTOs

type SMSSettingsResponse struct {
	Provider    string   `json:"provider"`
	APIKeyMask  string   `json:"api_key_mask"`
	LineNumber  string   `json:"line_number"`
	TemplateID  string   `json:"template_id"`
	IsSandbox   bool     `json:"is_sandbox"`
	LiveCredit  float64  `json:"live_credit"`
	Lines       []string `json:"lines"`
}

type UpdateSMSSettingsRequest struct {
	APIKey     *string `json:"api_key"`
	LineNumber *string `json:"line_number"`
	TemplateID *string `json:"template_id"`
	IsSandbox  *bool   `json:"is_sandbox"`
}

type SendTestSMSRequest struct {
	Mobile  string `json:"mobile" binding:"required"`
	Message string `json:"message"`
}

