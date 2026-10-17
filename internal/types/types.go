package types

import (
	"time"

	"github.com/EbadiDev/anahix-server/internal/model"
)

// Public Storefront DTOs

type ProductListItem struct {
	ID                  string   `json:"id"`
	Title               string   `json:"title"`
	TitleFA             string   `json:"title_fa"`
	Slug                string   `json:"slug"`
	Category            string   `json:"category"`
	DeliveryType        string   `json:"delivery_type"`
	PriceToman          int64    `json:"price_toman"`
	BaseCostUSD         float64  `json:"base_cost_usd,omitempty"`
	SLAMinMinutes       int      `json:"sla_min_minutes"`
	SLAMaxMinutes       int      `json:"sla_max_minutes"`
	SLADisplayText      string   `json:"sla_display_text"`
	OperatingHours      string   `json:"operating_hours"`
	IconURL             string   `json:"icon_url"`
	BannerURL           string   `json:"banner_url"`
	ImageURL            string   `json:"image_url,omitempty"`
	IsFlashSale         bool     `json:"is_flash_sale"`
	FlashSalePriceToman int64    `json:"flash_sale_price_toman"`
	FlashSaleStock      int      `json:"flash_sale_stock"`
	FlashSaleSold       int      `json:"flash_sale_sold"`
	IsBestSeller        bool     `json:"is_best_seller"`
	IsFeatured          bool     `json:"is_featured"`
	RatingAverage       float64  `json:"rating_average"`
	SalesCount          int      `json:"sales_count"`
	BadgesJSON          string   `json:"badges_json,omitempty"`
	AvailableStock      int64    `json:"available_stock"`
	IsActive            bool     `json:"is_active"`
	GroupTag            string   `json:"group_tag,omitempty"`
	GroupTitleFA        string   `json:"group_title_fa,omitempty"`
	Features            []string `json:"features,omitempty"`
}

type ProductPlan struct {
	ID                 string   `json:"id"`
	Title              string   `json:"title"`
	DurationDays       int      `json:"duration_days"`
	DurationText       string   `json:"duration_text"`
	PriceToman         int64    `json:"price_toman"`
	OriginalPriceToman int64    `json:"original_price_toman,omitempty"`
	BaseCostUSD        float64  `json:"base_cost_usd,omitempty"`
	Badge              string   `json:"badge,omitempty"`
	IsDefault          bool     `json:"is_default"`
	DeliveryType       string   `json:"delivery_type,omitempty"`
	Features           []string `json:"features,omitempty"`
}

type ComparisonColumn struct {
	Key       string `json:"key"`
	Title     string `json:"title"`
	Subtitle  string `json:"subtitle,omitempty"`
	Badge     string `json:"badge,omitempty"`
	Highlight bool   `json:"highlight"`
}

type ComparisonFeature struct {
	Name     string            `json:"name"`
	Category string            `json:"category,omitempty"`
	Values   map[string]string `json:"values"`
}

type ProductComparison struct {
	Columns  []ComparisonColumn  `json:"columns"`
	Features []ComparisonFeature `json:"features"`
}

type ActivationStep struct {
	StepNumber  int    `json:"step_number"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Icon        string `json:"icon,omitempty"`
	Tip         string `json:"tip,omitempty"`
	ImageURL    string `json:"image_url,omitempty"`
}

type ProductFAQ struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

type ProductSEO struct {
	MetaTitle       string `json:"meta_title"`
	MetaDescription string `json:"meta_description"`
	OgImageURL      string `json:"og_image_url,omitempty"`
}

type ProductDetail struct {
	ProductListItem
	Description      string             `json:"description"`
	Instructions     string             `json:"instructions"`
	DeliveryStrategy string             `json:"delivery_strategy,omitempty"`
	SEO              ProductSEO         `json:"seo"`
	Plans            []ProductPlan      `json:"plans"`
	ComparisonTable  *ProductComparison `json:"comparison_table,omitempty"`
	ActivationGuide  []ActivationStep   `json:"activation_guide"`
	Faqs             []ProductFAQ       `json:"faqs"`
	Features         []string           `json:"features"`

	// Raw JSON strings for Admin editing
	MetaTitle       string `json:"meta_title,omitempty"`
	MetaDescription string `json:"meta_description,omitempty"`
	OgImageURL      string `json:"og_image_url,omitempty"`
	PlansJSON       string `json:"plans_json,omitempty"`
	ComparisonJSON  string `json:"comparison_json,omitempty"`
	GuideJSON       string `json:"guide_json,omitempty"`
	FaqJSON         string `json:"faq_json,omitempty"`
	FeaturesJSON    string `json:"features_json,omitempty"`
}

type PaymentMethodItem struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	NameFA      string `json:"name_fa"`
	Platform    string `json:"platform"`
	Icon        string `json:"icon"`
	Description string `json:"description"`
}

type CartItemInput struct {
	ProductID string `json:"product_id"`
	PlanID    string `json:"plan_id,omitempty"`
	Quantity  int    `json:"quantity"`
}

type CreateOrderRequest struct {
	ProductID     string          `json:"product_id"`
	PlanID        string          `json:"plan_id,omitempty"`
	Items         []CartItemInput `json:"items"`
	CustomerPhone string          `json:"customer_phone" binding:"required"`
	CustomerEmail string          `json:"customer_email"`
	PaymentMethod string          `json:"payment_method" binding:"required"`
	CouponCode    string          `json:"coupon_code"`
	ReferralCode  string          `json:"referral_code"`

	// Optional fields for Timed / On-Account Upgrade (Mode C)
	CustomerAccountEmail    string `json:"customer_account_email"`
	CustomerAccountPassword string `json:"customer_account_password"`
	CustomerNotes           string `json:"customer_notes"`

	// Corporate / Legal entity invoice request
	IsCorporate    *bool  `json:"is_corporate,omitempty"`
	CompanyName    string `json:"company_name,omitempty"`
	NationalID     string `json:"national_id,omitempty"`
	EconomicCode   string `json:"economic_code,omitempty"`
	RegistrationNo string `json:"registration_no,omitempty"`
	PostalCode     string `json:"postal_code,omitempty"`
	BuyerAddress   string `json:"buyer_address,omitempty"`
	BuyerPhone     string `json:"buyer_phone,omitempty"`
}

type CreateOrderResponse struct {
	OrderID             string    `json:"order_id"`
	OrderNumber         string    `json:"order_number"`
	AccessToken         string    `json:"access_token"`
	SubtotalAmountToman int64     `json:"subtotal_amount_toman"`
	DiscountAmountToman int64     `json:"discount_amount_toman"`
	TotalAmountToman    int64     `json:"total_amount_toman"`
	PaymentMethod       string    `json:"payment_method"`
	PaymentURL          string    `json:"payment_url"`
	ExpiresAt           time.Time `json:"expires_at"`
}

type DeliveredItemDTO struct {
	ProductTitle     string `json:"product_title"`
	ProductSlug      string `json:"product_slug"`
	DeliveryType     string `json:"delivery_type"`
	Payload          string `json:"payload"` // Decrypted activation link or credentials
	Instructions     string `json:"instructions"`
}

type OrderDetailResponse struct {
	OrderID             string             `json:"order_id"`
	OrderNumber         string             `json:"order_number"`
	Status              string             `json:"status"`
	DeliveryType        string             `json:"delivery_type"`
	SubtotalAmountToman int64              `json:"subtotal_amount_toman"`
	DiscountAmountToman int64              `json:"discount_amount_toman"`
	CouponCode          string             `json:"coupon_code,omitempty"`
	TotalAmountToman    int64              `json:"total_amount_toman"`
	CustomerPhone       string             `json:"customer_phone"`
	PaymentMethod       string             `json:"payment_method"`
	CreatedAt           time.Time          `json:"created_at"`
	PaidAt              *time.Time         `json:"paid_at,omitempty"`
	CompletedAt         *time.Time         `json:"completed_at,omitempty"`
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
	CountryCode        string                      `json:"country_code,omitempty"`
	FlagURL            string                      `json:"flag_url,omitempty"`
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
	IsNew    bool                    `json:"is_new"`
	Customer CustomerProfileResponse `json:"customer"`
}

type CustomerProfileResponse struct {
	ID                   string    `json:"id"`
	Phone                string    `json:"phone"`
	Mobile               string    `json:"mobile"`
	FirstName            string    `json:"first_name,omitempty"`
	LastName             string    `json:"last_name,omitempty"`
	FullName             string    `json:"full_name"`
	Email                string    `json:"email"`
	AvatarURL            string    `json:"avatar_url,omitempty"`
	BalanceToman         int64     `json:"balance_toman"`
	Balance              int64     `json:"balance"`
	ReferralCode         string    `json:"referral_code"`
	ReferralBalanceToman int64     `json:"referral_balance_toman"`
	IsActive             bool      `json:"is_active"`
	IsAdmin              bool      `json:"is_admin"`

	// Corporate / Legal entity profile
	IsCorporate    bool   `json:"is_corporate"`
	CompanyName    string `json:"company_name,omitempty"`
	NationalID     string `json:"national_id,omitempty"`
	EconomicCode   string `json:"economic_code,omitempty"`
	RegistrationNo string `json:"registration_no,omitempty"`
	PostalCode     string `json:"postal_code,omitempty"`
	CompanyAddress string `json:"company_address,omitempty"`
	CompanyPhone   string `json:"company_phone,omitempty"`

	CreatedAt time.Time `json:"created_at"`
}

type UpdateCustomerProfileRequest struct {
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
	FullName  string `json:"full_name,omitempty"`
	Email     string `json:"email,omitempty"`
	Phone     string `json:"phone,omitempty"`
	AvatarURL string `json:"avatar_url,omitempty"`

	// Corporate / Legal entity update
	IsCorporate    *bool  `json:"is_corporate,omitempty"`
	CompanyName    string `json:"company_name,omitempty"`
	NationalID     string `json:"national_id,omitempty"`
	EconomicCode   string `json:"economic_code,omitempty"`
	RegistrationNo string `json:"registration_no,omitempty"`
	PostalCode     string `json:"postal_code,omitempty"`
	CompanyAddress string `json:"company_address,omitempty"`
	CompanyPhone   string `json:"company_phone,omitempty"`
}

// SMS Admin Settings DTOs

type SMSTemplateItem struct {
	Key         string   `json:"key"`          // "otp", "two_factor", "order_complete", "renewal"
	Title       string   `json:"title"`        // e.g. "الگوی خوش‌آمدگویی و ورود OTP"
	Description string   `json:"description"`  // e.g. "ارسال کد یکبار مصرف ورود و ثبت‌نام مشتریان"
	TemplateID  string   `json:"template_id"`  // e.g. "100000"
	Variables   []string `json:"variables"`    // e.g. ["Code"], ["Order", "Prompt"]
	SampleText  string   `json:"sample_text"`  // e.g. "کد ورود به آناهیکس: #Code#"
	IsEnabled   bool     `json:"is_enabled"`
}

type SMSSettingsResponse struct {
	Provider              string            `json:"provider"`
	APIKeyMask            string            `json:"api_key_mask"`
	LineNumber            string            `json:"line_number"`
	TemplateID            string            `json:"template_id"`
	TemplateOTP           string            `json:"template_otp"`
	Template2FA           string            `json:"template_2fa"`
	TemplateOrderComplete string            `json:"template_order_complete"`
	TemplateRenewal       string            `json:"template_renewal"`
	Templates             []SMSTemplateItem `json:"templates"`
	IsSandbox             bool              `json:"is_sandbox"`
	LiveCredit            float64           `json:"live_credit"`
	Lines                 []string          `json:"lines"`
}

type UpdateSMSSettingsRequest struct {
	APIKey                *string `json:"api_key"`
	LineNumber            *string `json:"line_number"`
	TemplateID            *string `json:"template_id"`
	TemplateOTP           *string `json:"template_otp"`
	Template2FA           *string `json:"template_2fa"`
	TemplateOrderComplete *string `json:"template_order_complete"`
	TemplateRenewal       *string `json:"template_renewal"`
	IsSandbox             *bool   `json:"is_sandbox"`
}

type SendTestSMSRequest struct {
	Mobile       string            `json:"mobile" binding:"required"`
	TemplateType string            `json:"template_type"` // "otp", "two_factor", "order_complete", "renewal"
	TemplateID   *int              `json:"template_id"`
	Parameters   map[string]string `json:"parameters"`
	Message      string            `json:"message"`
}

// Coupon DTOs

type ValidateCouponRequest struct {
	Code          string `json:"code" binding:"required"`
	SubtotalToman int64  `json:"subtotal_toman" binding:"required"`
	CustomerPhone string `json:"customer_phone"`
	Category      string `json:"category"`
}

type ValidateCouponResponse struct {
	Valid            bool   `json:"valid"`
	Code             string `json:"code"`
	DiscountType     string `json:"discount_type"`
	DiscountValue    int64  `json:"discount_value"`
	DiscountToman    int64  `json:"discount_toman"`
	FinalAmountToman int64  `json:"final_amount_toman"`
	Message          string `json:"message"`
}

type CouponListItem struct {
	ID                  string     `json:"id"`
	Code                string     `json:"code"`
	DiscountType        string     `json:"discount_type"`
	DiscountValue       int64      `json:"discount_value"`
	MinOrderAmountToman int64      `json:"min_order_amount_toman"`
	MaxDiscountToman    int64      `json:"max_discount_toman"`
	PerUserLimit        int        `json:"per_user_limit"`
	MaxUsage            int        `json:"max_usage"`
	CurrentUsage        int        `json:"current_usage"`
	Category            string     `json:"category"`
	ExpiresAt           *time.Time `json:"expires_at,omitempty"`
	IsActive            bool       `json:"is_active"`
	CreatedAt           time.Time  `json:"created_at"`
}

type CreateCouponRequest struct {
	Code                string     `json:"code" binding:"required"`
	DiscountType        string     `json:"discount_type" binding:"required"`
	DiscountValue       int64      `json:"discount_value" binding:"required"`
	MinOrderAmountToman int64      `json:"min_order_amount_toman"`
	MaxDiscountToman    int64      `json:"max_discount_toman"`
	PerUserLimit        int        `json:"per_user_limit"`
	MaxUsage            int        `json:"max_usage"`
	Category            string     `json:"category"`
	ExpiresAt           *time.Time `json:"expires_at"`
	IsActive            bool       `json:"is_active"`
}

// Referral DTOs

type ReferralStatsResponse struct {
	ReferralCode               string  `json:"referral_code"`
	ReferralURL                string  `json:"referral_url"`
	CommissionRatePercent      float64 `json:"commission_rate_percent"`
	TotalReferrals             int     `json:"total_referrals"`
	SuccessfulOrders           int     `json:"successful_orders"`
	TotalCommissionToman       int64   `json:"total_commission_toman"`
	WithdrawableCommissionToman int64  `json:"withdrawable_commission_toman"`
}

type TransferCommissionResponse struct {
	TransferredToman int64  `json:"transferred_toman"`
	NewBalanceToman  int64  `json:"new_balance_toman"`
	Message          string `json:"message"`
}

// Inventory Admin Management DTOs

type InventoryItemDTO struct {
	ID              string     `json:"id"`
	ProductID       string     `json:"product_id"`
	ProductTitle    string     `json:"product_title"`
	ItemType        string     `json:"item_type"`
	Status          string     `json:"status"`
	CostUSD         float64    `json:"cost_usd"`
	CostToman       int64      `json:"cost_toman"`
	ExpirationDate  *time.Time `json:"expiration_date,omitempty"`
	BatchTag        string     `json:"batch_tag,omitempty"`
	ReservedOrderID *string    `json:"reserved_order_id,omitempty"`
	DeliveredAt     *time.Time `json:"delivered_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	PayloadPreview  string     `json:"payload_preview,omitempty"`
}

type InventoryStatsResponse struct {
	TotalCount          int64 `json:"total_count"`
	AvailableCount      int64 `json:"available_count"`
	ReservedCount       int64 `json:"reserved_count"`
	SoldDeliveredCount  int64 `json:"sold_delivered_count"`
	VoidCount           int64 `json:"void_count"`
}

type ListInventoryResponse struct {
	Total int64              `json:"total"`
	Items []InventoryItemDTO `json:"items"`
}

type ReplaceKeyRequest struct {
	OrderID            string `json:"order_id" binding:"required"`
	OldInventoryItemID string `json:"old_inventory_item_id" binding:"required"`
	Reason             string `json:"reason"`
}

type ReplaceKeyResponse struct {
	Success            bool   `json:"success"`
	NewInventoryItemID string `json:"new_inventory_item_id"`
	DeliveredPayload   string `json:"delivered_payload"`
	Message            string `json:"message"`
}

// Customer Wallet Ledger DTO

type WalletLedgerItemDTO struct {
	ID                string    `json:"id"`
	TransactionNumber string    `json:"transaction_number"`
	Title             string    `json:"title"`
	Type              string    `json:"type"` // deposit, store_order, virtual_number, refund, referral_bonus
	AmountToman       int64     `json:"amount_toman"`
	Status            string    `json:"status"`
	PaymentMethod     string    `json:"payment_method"`
	PaymentRefID      string    `json:"payment_ref_id,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
}

// Review and Social Proof DTOs

type ReviewItemDTO struct {
	ID               string     `json:"id"`
	ProductID        string     `json:"product_id"`
	ProductSlug      string     `json:"product_slug"`
	ProductTitle     string     `json:"product_title"`
	OrderID          string     `json:"order_id,omitempty"`
	OrderNumber      string     `json:"order_number,omitempty"`
	CustomerName     string     `json:"customer_name"`
	CustomerPhone    string     `json:"customer_phone,omitempty"` // Masked for public
	Rating           int        `json:"rating"`
	Comment          string     `json:"comment"`
	ExperienceTags   []string   `json:"experience_tags"`
	Status           string     `json:"status"` // pending, approved, rejected
	IsVerifiedBuyer  bool       `json:"is_verified_buyer"`
	IsFeatured       bool       `json:"is_featured"`
	HelpfulCount     int        `json:"helpful_count"`
	AdminReply       string     `json:"admin_reply,omitempty"`
	AdminReplyAuthor string     `json:"admin_reply_author,omitempty"`
	AdminRepliedAt   *time.Time `json:"admin_replied_at,omitempty"`
	RejectionReason  string     `json:"rejection_reason,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
}

type RatingBreakdown struct {
	Star5        int64 `json:"star_5"`
	Star4        int64 `json:"star_4"`
	Star3        int64 `json:"star_3"`
	Star2        int64 `json:"star_2"`
	Star1        int64 `json:"star_1"`
	Star5Percent int   `json:"star_5_percent"`
	Star4Percent int   `json:"star_4_percent"`
	Star3Percent int   `json:"star_3_percent"`
	Star2Percent int   `json:"star_2_percent"`
	Star1Percent int   `json:"star_1_percent"`
}

type ExperienceTagSummary struct {
	Tag     string `json:"tag"`
	Count   int64  `json:"count"`
	Percent int    `json:"percent"`
}

type ProductReviewsSummary struct {
	AverageRating       float64                `json:"average_rating"`
	TotalReviews        int64                  `json:"total_reviews"`
	RatingBreakdown     RatingBreakdown        `json:"rating_breakdown"`
	TopTags             []ExperienceTagSummary `json:"top_tags"`
	VerifiedBuyersCount int64                  `json:"verified_buyers_count"`
	SatisfactionRate    int                    `json:"satisfaction_rate"` // e.g. 98%
}

type ProductReviewsResponse struct {
	Summary ProductReviewsSummary `json:"summary"`
	Reviews []ReviewItemDTO       `json:"reviews"`
	Total   int64                 `json:"total"`
}

type SubmitReviewRequest struct {
	ProductID        string   `json:"product_id"`
	ProductSlug      string   `json:"product_slug"`
	OrderID          string   `json:"order_id"`
	OrderNumber      string   `json:"order_number"`
	OrderAccessToken string   `json:"order_access_token"`
	CustomerPhone    string   `json:"customer_phone"`
	CustomerName     string   `json:"customer_name"`
	Rating           int      `json:"rating" binding:"required,min=1,max=5"`
	Comment          string   `json:"comment"`
	ExperienceTags   []string `json:"experience_tags"`
}

type SubmitReviewResponse struct {
	ReviewID         string `json:"review_id"`
	Status           string `json:"status"`
	Message          string `json:"message"`
	RequiresApproval bool   `json:"requires_approval"`
}

type ReviewEligibilityResponse struct {
	Eligible       bool           `json:"eligible"`
	Reason         string         `json:"reason,omitempty"`
	OrderID        string         `json:"order_id,omitempty"`
	OrderNumber    string         `json:"order_number,omitempty"`
	ProductTitle   string         `json:"product_title,omitempty"`
	ExistingReview *ReviewItemDTO `json:"existing_review,omitempty"`
}

type AdminReviewListRequest struct {
	Status    string `form:"status"` // all, pending, approved, rejected
	ProductID string `form:"product_id"`
	Search    string `form:"search"`
	Page      int    `form:"page,default=1"`
	PageSize  int    `form:"page_size,default=20"`
}

type AdminReviewStats struct {
	TotalReviews    int64   `json:"total_reviews"`
	PendingReviews  int64   `json:"pending_reviews"`
	ApprovedReviews int64   `json:"approved_reviews"`
	RejectedReviews int64   `json:"rejected_reviews"`
	AverageRating   float64 `json:"average_rating"`
}

type AdminReviewListResponse struct {
	Items []ReviewItemDTO  `json:"items"`
	Total int64            `json:"total"`
	Stats AdminReviewStats `json:"stats"`
}

type AdminUpdateReviewStatusRequest struct {
	Status          string `json:"status" binding:"required"` // approved, rejected
	RejectionReason string `json:"rejection_reason"`
}

type AdminReplyReviewRequest struct {
	Reply  string `json:"reply" binding:"required"`
	Author string `json:"author"`
}

type AdminFeatureReviewRequest struct {
	IsFeatured bool `json:"is_featured"`
}

// Support Ticket DTOs

type TicketAttachmentDTO struct {
	ID       string `json:"id"`
	FileName string `json:"file_name"`
	FileURL  string `json:"file_url"`
	FileSize int64  `json:"file_size"`
	MimeType string `json:"mime_type"`
	IsImage  bool   `json:"is_image"`
}

type TicketMessageDTO struct {
	ID          string                `json:"id"`
	TicketID    string                `json:"ticket_id"`
	SenderType  string                `json:"sender_type"` // customer, admin, system
	SenderID    string                `json:"sender_id"`
	SenderName  string                `json:"sender_name"`
	Message     string                `json:"message"`
	Attachments []TicketAttachmentDTO `json:"attachments"`
	IsInternal  bool                  `json:"is_internal"`
	CreatedAt   time.Time             `json:"created_at"`
}

type TicketItemDTO struct {
	ID                  string     `json:"id"`
	TicketNumber        string     `json:"ticket_number"`
	CustomerID          string     `json:"customer_id"`
	CustomerPhone       string     `json:"customer_phone"`
	CustomerName        string     `json:"customer_name"`
	CustomerEmail       string     `json:"customer_email"`
	Subject             string     `json:"subject"`
	Department          string     `json:"department"`
	Priority            string     `json:"priority"`
	Status              string     `json:"status"`
	RelatedOrderNo      string     `json:"related_order_no,omitempty"`
	NotifyViaSMS        bool       `json:"notify_via_sms"`
	NotifyViaEmail      bool       `json:"notify_via_email"`
	AssignedAdminName   string     `json:"assigned_admin_name,omitempty"`
	SatisfactionRating  int        `json:"satisfaction_rating,omitempty"`
	SatisfactionComment string     `json:"satisfaction_comment,omitempty"`
	LastReplyAt         *time.Time `json:"last_reply_at,omitempty"`
	LastReplyBy         string     `json:"last_reply_by"`
	MessagesCount       int        `json:"messages_count"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

type TicketDetailResponse struct {
	Ticket   TicketItemDTO      `json:"ticket"`
	Messages []TicketMessageDTO `json:"messages"`
}

type CreateTicketRequest struct {
	Subject        string                `json:"subject" binding:"required"`
	Department     string                `json:"department" binding:"required"`
	Priority       string                `json:"priority"`
	RelatedOrderNo string                `json:"related_order_no"`
	NotifyViaSMS   *bool                 `json:"notify_via_sms"`
	NotifyViaEmail *bool                 `json:"notify_via_email"`
	Message        string                `json:"message" binding:"required"`
	Attachments    []TicketAttachmentDTO `json:"attachments"`
}

type ReplyTicketRequest struct {
	Message        string                `json:"message" binding:"required"`
	Attachments    []TicketAttachmentDTO `json:"attachments"`
	NotifyViaSMS   *bool                 `json:"notify_via_sms"`
	NotifyViaEmail *bool                 `json:"notify_via_email"`
}

type AdminReplyTicketRequest struct {
	Message        string                `json:"message" binding:"required"`
	IsInternal     bool                  `json:"is_internal"`
	Attachments    []TicketAttachmentDTO `json:"attachments"`
	Status         string                `json:"status"` // optional updated status
	NotifySMS      bool                  `json:"notify_sms"`
	NotifyEmail    bool                  `json:"notify_email"`
}

type AdminUpdateTicketRequest struct {
	Status            *string `json:"status"`
	Priority          *string `json:"priority"`
	Department        *string `json:"department"`
	AssignedAdminName *string `json:"assigned_admin_name"`
	NotifySMS         bool    `json:"notify_sms"`
	NotifyEmail       bool    `json:"notify_email"`
}

type AdminTicketListRequest struct {
	Status     string `form:"status"` // all, awaiting_admin, awaiting_user, in_progress, resolved, closed
	Department string `form:"department"`
	Priority   string `form:"priority"`
	Search     string `form:"search"`
	Page       int    `form:"page,default=1"`
	PageSize   int    `form:"page_size,default=20"`
}

type AdminTicketStats struct {
	Total         int64 `json:"total"`
	AwaitingAdmin int64 `json:"awaiting_admin"`
	AwaitingUser  int64 `json:"awaiting_user"`
	InProgress    int64 `json:"in_progress"`
	Resolved      int64 `json:"resolved"`
	Closed        int64 `json:"closed"`
}

type AdminTicketListResponse struct {
	Items []TicketItemDTO  `json:"items"`
	Total int64            `json:"total"`
	Stats AdminTicketStats `json:"stats"`
}

type RateTicketRequest struct {
	Rating  int    `json:"rating" binding:"required,min=1,max=5"`
	Comment string `json:"comment"`
}

type UploadAttachmentResponse struct {
	ID       string `json:"id"`
	FileName string `json:"file_name"`
	FileURL  string `json:"file_url"`
	FileSize int64  `json:"file_size"`
	MimeType string `json:"mime_type"`
	IsImage  bool   `json:"is_image"`
}

// ==========================================
// Official Tax Invoice & Billing DTOs
// ==========================================

type SellerBillingInfo struct {
	CompanyName       string `json:"company_name"`
	BrandFA           string `json:"brand_fa"`
	NationalID        string `json:"national_id"`
	EconomicCode      string `json:"economic_code"`
	RegistrationNo    string `json:"registration_no"`
	PostalCode        string `json:"postal_code"`
	Province          string `json:"province"`
	City              string `json:"city"`
	Address           string `json:"address"`
	Phone             string `json:"phone"`
	Email             string `json:"email"`
	Website           string `json:"website"`
	TaxMemoryID       string `json:"tax_memory_id"`
	StampSignatureURL string `json:"stamp_signature_url"`
}

type BuyerBillingInfo struct {
	IsCorporate    bool   `json:"is_corporate"`
	CompanyName    string `json:"company_name"`
	FullName       string `json:"full_name"`
	NationalID     string `json:"national_id"`
	EconomicCode   string `json:"economic_code"`
	RegistrationNo string `json:"registration_no"`
	PostalCode     string `json:"postal_code"`
	Address        string `json:"address"`
	Phone          string `json:"phone"`
	Email          string `json:"email"`
}

type InvoiceItemDTO struct {
	RowIndex                int    `json:"row_index"`
	ProductID               string `json:"product_id"`
	ProductTitle            string `json:"product_title"`
	ProductSlug             string `json:"product_slug"`
	DeliveryType            string `json:"delivery_type"`
	Quantity                int    `json:"quantity"`
	UnitPriceToman          int64  `json:"unit_price_toman"`
	SubtotalToman           int64  `json:"subtotal_toman"`
	DiscountToman           int64  `json:"discount_toman"`
	TotalAfterDiscountToman int64  `json:"total_after_discount_toman"`
	TaxRatePercent          int    `json:"tax_rate_percent"`
	TaxToman                int64  `json:"tax_toman"`
	FinalAmountToman        int64  `json:"final_amount_toman"`
}

type OfficialInvoiceResponse struct {
	InvoiceNumber           string            `json:"invoice_number"`
	OrderNumber             string            `json:"order_number"`
	OrderID                 string            `json:"order_id"`
	InvoiceType             string            `json:"invoice_type"`    // "proforma" or "official"
	InvoiceTypeFA           string            `json:"invoice_type_fa"` // "پیش‌فاکتور رسمی" or "صورتحساب رسمی فروش کالا و خدمات"
	IssueDate               time.Time         `json:"issue_date"`
	IssueDateJalali         string            `json:"issue_date_jalali"`
	PaidAt                  *time.Time        `json:"paid_at,omitempty"`
	PaidAtJalali            string            `json:"paid_at_jalali,omitempty"`
	PaymentMethod           string            `json:"payment_method"`
	PaymentMethodFA         string            `json:"payment_method_fa"`
	PaymentRefID            string            `json:"payment_ref_id,omitempty"`
	PaymentAuthority        string            `json:"payment_authority,omitempty"`
	PaymentStatus           string            `json:"payment_status"`
	PaymentStatusFA         string            `json:"payment_status_fa"`
	Seller                  SellerBillingInfo `json:"seller"`
	Buyer                   BuyerBillingInfo  `json:"buyer"`
	Items                   []InvoiceItemDTO  `json:"items"`
	SubtotalToman           int64             `json:"subtotal_toman"`
	TotalDiscountToman      int64             `json:"total_discount_toman"`
	AmountAfterDiscountToman int64            `json:"amount_after_discount_toman"`
	TaxRatePercent          int               `json:"tax_rate_percent"`
	TotalTaxToman           int64             `json:"total_tax_toman"`
	FinalTotalToman         int64             `json:"final_total_toman"`
	FinalTotalTomanWords    string            `json:"final_total_toman_words"`
	FinalTotalRials         int64             `json:"final_total_rials"`
	Notes                   string            `json:"notes"`
}

type UpdateBuyerInvoiceRequest struct {
	IsCorporate    bool   `json:"is_corporate"`
	CompanyName    string `json:"company_name"`
	FullName       string `json:"full_name"`
	NationalID     string `json:"national_id"`
	EconomicCode   string `json:"economic_code"`
	RegistrationNo string `json:"registration_no"`
	PostalCode     string `json:"postal_code"`
	Address        string `json:"address"`
	Phone          string `json:"phone"`
	Email          string `json:"email"`
}

type BillingSettingsDTO struct {
	SellerCompanyName string `json:"seller_company_name"`
	SellerBrandFA     string `json:"seller_brand_fa"`
	SellerNationalID  string `json:"seller_national_id"`
	SellerEconomicCode string `json:"seller_economic_code"`
	SellerRegistrationNo string `json:"seller_registration_no"`
	SellerPostalCode  string `json:"seller_postal_code"`
	SellerProvince    string `json:"seller_province"`
	SellerCity        string `json:"seller_city"`
	SellerAddress     string `json:"seller_address"`
	SellerPhone       string `json:"seller_phone"`
	SellerEmail       string `json:"seller_email"`
	SellerWebsite     string `json:"seller_website"`
	VATPercent        int    `json:"vat_percent"`
	VATEnabled        bool   `json:"vat_enabled"`
	InvoiceNotes      string `json:"invoice_notes"`
	TaxMemoryID       string `json:"tax_memory_id"`
	StampSignatureURL string `json:"stamp_signature_url"`
}

type UpdateBillingSettingsRequest struct {
	SellerCompanyName    *string `json:"seller_company_name"`
	SellerBrandFA        *string `json:"seller_brand_fa"`
	SellerNationalID     *string `json:"seller_national_id"`
	SellerEconomicCode   *string `json:"seller_economic_code"`
	SellerRegistrationNo *string `json:"seller_registration_no"`
	SellerPostalCode     *string `json:"seller_postal_code"`
	SellerProvince       *string `json:"seller_province"`
	SellerCity           *string `json:"seller_city"`
	SellerAddress        *string `json:"seller_address"`
	SellerPhone          *string `json:"seller_phone"`
	SellerEmail          *string `json:"seller_email"`
	SellerWebsite        *string `json:"seller_website"`
	VATPercent           *int    `json:"vat_percent"`
	VATEnabled           *bool   `json:"vat_enabled"`
	InvoiceNotes         *string `json:"invoice_notes"`
	TaxMemoryID          *string `json:"tax_memory_id"`
	StampSignatureURL    *string `json:"stamp_signature_url"`
}

// -----------------------------------------------------------------------------
// Admin 360-Degree Customer Management & Order Enhancement DTOs
// -----------------------------------------------------------------------------

type AdminCustomerListItem struct {
	ID                   string     `json:"id"`
	Phone                string     `json:"phone"`
	FirstName            string     `json:"first_name,omitempty"`
	LastName             string     `json:"last_name,omitempty"`
	FullName             string     `json:"full_name"`
	Email                string     `json:"email"`
	Tier                 string     `json:"tier"` // "standard", "colleague", "vip"
	BalanceToman         int64      `json:"balance_toman"`
	ReferralCode         string     `json:"referral_code"`
	ReferralBalanceToman int64      `json:"referral_balance_toman"`
	TotalOrdersCount     int64      `json:"total_orders_count"`
	StoreOrdersCount     int64      `json:"store_orders_count"`
	VirtualOrdersCount   int64      `json:"virtual_orders_count"`
	TotalSpentToman      int64      `json:"total_spent_toman"`
	IsActive             bool       `json:"is_active"`
	IsAdmin              bool       `json:"is_admin"`
	BlockReason          string     `json:"block_reason,omitempty"`
	InternalNotes        string     `json:"internal_notes,omitempty"`
	LastLoginAt          *time.Time `json:"last_login_at,omitempty"`
	IsOnline             bool       `json:"is_online"`
	CreatedAt            time.Time  `json:"created_at"`
}

type AdminListCustomersRequest struct {
	Search    string `form:"search" json:"search"`
	Status    string `form:"status" json:"status"` // "all", "active", "blocked"
	HasBalance *bool `form:"has_balance" json:"has_balance"`
	Tier      string `form:"tier" json:"tier"`     // "all", "standard", "colleague", "vip"
	SortBy    string `form:"sort_by" json:"sort_by"` // "created_at", "last_login_at", "total_spent", "balance"
	SortOrder string `form:"sort_order" json:"sort_order"` // "desc", "asc"
	Page      int    `form:"page,default=1" json:"page"`
	PageSize  int    `form:"page_size,default=15" json:"page_size"`
}

type AdminListCustomersResponse struct {
	Customers         []AdminCustomerListItem `json:"customers"`
	Total             int64                   `json:"total"`
	Page              int                     `json:"page"`
	PageSize          int                     `json:"page_size"`
	TotalActive       int64                   `json:"total_active"`
	TotalBlocked      int64                   `json:"total_blocked"`
	TotalVIP          int64                   `json:"total_vip"`
	TotalBalanceToman int64                   `json:"total_balance_toman"`
}

type AdminCustomerOrderBrief struct {
	ID               string              `json:"id"`
	OrderNumber      string              `json:"order_number"`
	Status           model.OrderStatus   `json:"status"`
	DeliveryType     model.DeliveryType  `json:"delivery_type"`
	TotalAmountToman int64               `json:"total_amount_toman"`
	PaymentMethod    model.PaymentMethod `json:"payment_method"`
	ProductTitle     string              `json:"product_title"`
	DeliveredPayload string              `json:"delivered_payload,omitempty"`
	Items            []model.OrderItem   `json:"items"`
	CreatedAt        time.Time           `json:"created_at"`
	PaidAt           *time.Time          `json:"paid_at,omitempty"`
}

type AdminReferredUserItem struct {
	ID        string    `json:"id"`
	Phone     string    `json:"phone"`
	FullName  string    `json:"full_name"`
	CreatedAt time.Time `json:"created_at"`
}

type AdminCustomerAffiliateStats struct {
	ReferralCode               string                    `json:"referral_code"`
	ReferredCount              int64                     `json:"referred_count"`
	TotalCommissionToman       int64                     `json:"total_commission_toman"`
	TransferredCommissionToman int64                     `json:"transferred_commission_toman"`
	AvailableCommissionToman   int64                     `json:"available_commission_toman"`
	ReferredUsers              []AdminReferredUserItem   `json:"referred_users"`
	Commissions                []model.ReferralCommission `json:"commissions"`
}

type AdminCustomerDetail360Response struct {
	Customer          AdminCustomerListItem       `json:"customer"`
	WalletTransactions []model.Transaction        `json:"wallet_transactions"`
	Orders            []AdminCustomerOrderBrief   `json:"orders"`
	VirtualNumbers    []model.VirtualNumberOrder  `json:"virtual_numbers"`
	AffiliateStats    AdminCustomerAffiliateStats `json:"affiliate_stats"`
	AuditLogs         []model.AdminActionLog      `json:"audit_logs"`
}

type AdminAdjustCustomerWalletRequest struct {
	AmountToman int64  `json:"amount_toman" binding:"required,gt=0"`
	Action      string `json:"action" binding:"required"` // "credit" (شارژ), "debit" (کسر)
	Reason      string `json:"reason" binding:"required"` // Mandatory علت تراکنش
}

type AdminUpdateCustomerStatusRequest struct {
	IsActive      *bool   `json:"is_active"`
	BlockReason   *string `json:"block_reason"`
	Tier          *string `json:"tier"`
	InternalNotes *string `json:"internal_notes"`
	FullName      *string `json:"full_name"`
	Email         *string `json:"email"`
}

type AdminSendCustomerSMSRequest struct {
	Message    string `json:"message" binding:"required"`
	TemplateID *int   `json:"template_id"`
}

type AdminReplaceCustomerKeyRequest struct {
	OrderItemID string `json:"order_item_id" binding:"required"`
	NewPayload  string `json:"new_payload" binding:"required"`
	Reason      string `json:"reason"`
}

type AdminUpdateOrderStatusRequest struct {
	Status         model.OrderStatus `json:"status" binding:"required"`
	OperatorNotes  string            `json:"operator_notes"`
	RefundToWallet bool              `json:"refund_to_wallet"`
}






