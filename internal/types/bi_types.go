package types

import "time"

// AdminBIOverviewRequest filters for the BI dashboard
type AdminBIOverviewRequest struct {
	Range     string `form:"range" json:"range"`           // 7d, 30d, 90d, 1y, all
	GroupBy   string `form:"group_by" json:"group_by"`     // daily, weekly, monthly
	StartDate string `form:"start_date" json:"start_date"` // Optional ISO date
	EndDate   string `form:"end_date" json:"end_date"`     // Optional ISO date
}

// AdminBIKPIs represents headline commercial metrics
type AdminBIKPIs struct {
	TotalRevenueToman      int64   `json:"total_revenue_toman"`
	TotalRevenueUSDT       float64 `json:"total_revenue_usdt"`
	TotalCOGSToman         int64   `json:"total_cogs_toman"`
	TotalCOGSUSDT          float64 `json:"total_cogs_usdt"`
	GrossProfitToman       int64   `json:"gross_profit_toman"`
	GrossProfitUSDT        float64 `json:"gross_profit_usdt"`
	GrossMarginPercent     float64 `json:"gross_margin_percent"`
	TotalOrders            int64   `json:"total_orders"`
	PaidOrders             int64   `json:"paid_orders"`
	ConversionRatePercent  float64 `json:"conversion_rate_percent"`
	AverageOrderValueToman int64   `json:"average_order_value_toman"`
	AverageOrderValueUSDT  float64 `json:"average_order_value_usdt"`
	TotalCustomers         int64   `json:"total_customers"`
	VirtualNumbersRevenue  int64   `json:"virtual_numbers_revenue"`
	VirtualNumbersCount    int64   `json:"virtual_numbers_count"`
	StoreProductsRevenue   int64   `json:"store_products_revenue"`
	StoreProductsCount     int64   `json:"store_products_count"`
}

// AdminBITimeSeriesPoint represents one data point on line/bar charts
type AdminBITimeSeriesPoint struct {
	PeriodLabel      string  `json:"period_label"` // e.g. "۱۴۰۲/۰۷/۱۰" or "روز دوشنبه"
	Date             string  `json:"date"`         // ISO format YYYY-MM-DD
	RevenueToman     int64   `json:"revenue_toman"`
	RevenueUSDT      float64 `json:"revenue_usdt"`
	COGSToman        int64   `json:"cogs_toman"`
	COGSUSDT         float64 `json:"cogs_usdt"`
	GrossProfitToman int64   `json:"gross_profit_toman"`
	GrossProfitUSDT  float64 `json:"gross_profit_usdt"`
	GrossMargin      float64 `json:"gross_margin"`
	TotalOrders      int64   `json:"total_orders"`
	PaidOrders       int64   `json:"paid_orders"`
	ConversionRate   float64 `json:"conversion_rate"`
}

// AdminBITopProduct represents product sales performance ranking
type AdminBITopProduct struct {
	ProductID        string  `json:"product_id"`
	TitleFA          string  `json:"title_fa"`
	Slug             string  `json:"slug"`
	Category         string  `json:"category"`
	UnitsSold        int64   `json:"units_sold"`
	RevenueToman     int64   `json:"revenue_toman"`
	COGSToman        int64   `json:"cogs_toman"`
	GrossProfitToman int64   `json:"gross_profit_toman"`
	MarginPercent    float64 `json:"margin_percent"`
}

// AdminBITopCountry represents 5SIM virtual number performance by country
type AdminBITopCountry struct {
	Country       string `json:"country"`
	CountryNameFA string `json:"country_name_fa"`
	CountryFlag   string `json:"country_flag"`
	OrdersCount   int64  `json:"orders_count"`
	RevenueToman  int64  `json:"revenue_toman"`
	TopService    string `json:"top_service"`
}

// AdminBILoyaltyDistribution represents customer count & spend per VIP tier
type AdminBILoyaltyDistribution struct {
	TierCode          string  `json:"tier_code"`
	TierName          string  `json:"tier_name"`
	TierNameEN        string  `json:"tier_name_en"`
	BadgeColor        string  `json:"badge_color"`
	DiscountPercent   float64 `json:"discount_percent"`
	CustomerCount     int64   `json:"customer_count"`
	TotalSpentToman   int64   `json:"total_spent_toman"`
	PercentageOfUsers float64 `json:"percentage_of_users"`
	PercentageOfSales float64 `json:"percentage_of_sales"`
}

// AdminBIOverviewResponse wraps the entire analytics data
type AdminBIOverviewResponse struct {
	KPIs          AdminBIKPIs                  `json:"kpis"`
	TimeSeries    []AdminBITimeSeriesPoint     `json:"time_series"`
	TopProducts   []AdminBITopProduct          `json:"top_products"`
	TopCountries  []AdminBITopCountry          `json:"top_countries"`
	LoyaltyStats  []AdminBILoyaltyDistribution `json:"loyalty_stats"`
	EffectiveUSDRate float64                  `json:"effective_usd_rate"`
	GeneratedAt   string                       `json:"generated_at"`
}

// LoyaltyTierDTO represents a tier entity exposed to frontend
type LoyaltyTierDTO struct {
	ID              string    `json:"id"`
	Code            string    `json:"code"`
	Name            string    `json:"name"`
	NameEN          string    `json:"name_en"`
	MinSpendToman   int64     `json:"min_spend_toman"`
	DiscountPercent float64   `json:"discount_percent"`
	QueuePriority   int       `json:"queue_priority"`
	BadgeColor      string    `json:"badge_color"`
	BadgeGradient   string    `json:"badge_gradient"`
	Icon            string    `json:"icon"`
	Description     string    `json:"description"`
	Perks           []string  `json:"perks"`
	OrderIndex      int       `json:"order_index"`
	IsActive        bool      `json:"is_active"`
	UserCount       int64     `json:"user_count,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// CreateOrUpdateLoyaltyTierRequest
type CreateOrUpdateLoyaltyTierRequest struct {
	Code            string   `json:"code" binding:"required"`
	Name            string   `json:"name" binding:"required"`
	NameEN          string   `json:"name_en"`
	MinSpendToman   int64    `json:"min_spend_toman"`
	DiscountPercent float64  `json:"discount_percent"`
	QueuePriority   int      `json:"queue_priority"`
	BadgeColor      string   `json:"badge_color"`
	BadgeGradient   string   `json:"badge_gradient"`
	Icon            string   `json:"icon"`
	Description     string   `json:"description"`
	Perks           []string `json:"perks"`
	OrderIndex      int      `json:"order_index"`
	IsActive        *bool    `json:"is_active"`
}

// CustomerLoyaltyProfileResponse represents the user's loyalty status and gamified progress
type CustomerLoyaltyProfileResponse struct {
	CurrentTier       LoyaltyTierDTO   `json:"current_tier"`
	NextTier          *LoyaltyTierDTO  `json:"next_tier,omitempty"`
	TotalSpentToman   int64            `json:"total_spent_toman"`
	MonthlySpentToman int64            `json:"monthly_spent_toman"`
	Period            string           `json:"period"` // "monthly_30d" or "all_time"
	ActiveSpentToman  int64            `json:"active_spent_toman"`
	RequiredForNext   int64            `json:"required_for_next"`
	ProgressPercent   float64          `json:"progress_percent"`
	AllTiers          []LoyaltyTierDTO `json:"all_tiers"`
	ActiveDiscountPct float64          `json:"active_discount_pct"`
	TotalSavedToman   int64            `json:"total_saved_toman"`
}
