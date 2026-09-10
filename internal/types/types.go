package types

import "time"

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
	TwoFactorCode string `json:"two_factor_code" binding:"required"`
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
