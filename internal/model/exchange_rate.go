package model

import (
	"time"

	"gorm.io/gorm"
)

type ExchangeRateSetting struct {
	ID uint `gorm:"primaryKey" json:"id"`

	// Currency conversion rate: 1 USD / USDT to Iranian Toman
	USDtoTomanRate int64 `gorm:"not null;default:94000" json:"usd_to_toman_rate"`

	// Default reseller profit margin in percentage (e.g. 18.5 for 18.5%)
	DefaultProfitMarginPercent float64 `gorm:"type:decimal(5,2);default:15.00" json:"default_profit_margin_percent"`

	// Fixed processing / operational fee in Toman
	FixedFeeToman int64 `gorm:"default:20000" json:"fixed_fee_toman"`

	// Auto-fetch toggle (e.g., from Nobitex/Wallex API)
	AutoFetchEnabled bool `gorm:"default:false" json:"auto_fetch_enabled"`

	LastUpdatedAt time.Time      `json:"last_updated_at"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

func (s *ExchangeRateSetting) CalculatePriceToman(baseCostUSD float64, customMarginPercent *float64) int64 {
	margin := s.DefaultProfitMarginPercent
	if customMarginPercent != nil {
		margin = *customMarginPercent
	}

	costInToman := float64(s.USDtoTomanRate) * baseCostUSD
	profitMultiplier := 1.0 + (margin / 100.0)
	finalToman := (costInToman * profitMultiplier) + float64(s.FixedFeeToman)

	// Round to nearest 1,000 Toman for clean Iranian pricing
	rounded := (int64(finalToman) + 500) / 1000 * 1000
	return rounded
}
