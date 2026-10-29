package fulfillment

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type LootPagluAdapter struct{}

func (a *LootPagluAdapter) GetProviderKey() string {
	return ProviderLootPaglu
}

func (a *LootPagluAdapter) GetProviderNameFA() string {
	return "Loot Paglu"
}

func (a *LootPagluAdapter) GetDefaultBaseURL() string {
	return "https://lootpaglu.in"
}

func (a *LootPagluAdapter) GetBalance(client *http.Client, cfg ProviderConfig) (*ProviderBalanceDTO, error) {
	if cfg.APIKey == "" {
		return nil, errors.New("کلید API لوت پاگلو تنظیم نشده است")
	}
	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = a.GetDefaultBaseURL()
	}

	req, err := http.NewRequest(http.MethodGet, baseURL+"/api/v1/me", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-API-Key", cfg.APIKey)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("خطا در اتصال به لوت پاگلو: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, errors.New("کلید API لوت پاگلو نامعتبر است (401)")
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("کد خطای HTTP لوت پاگلو: %d (%s)", resp.StatusCode, string(body))
	}

	var data struct {
		UserID       int64   `json:"userId"`
		WalletINR    float64 `json:"wallet_inr"`
		WalletCrypto float64 `json:"wallet_crypto"`
		TotalSpent   float64 `json:"total_spent"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	return &ProviderBalanceDTO{
		ProviderKey:    ProviderLootPaglu,
		ProviderNameFA: a.GetProviderNameFA(),
		BalanceUSD:     data.WalletCrypto,
		Status:         "active",
		Raw:            data,
	}, nil
}

func (a *LootPagluAdapter) GetCatalog(client *http.Client, cfg ProviderConfig) ([]ProviderProductDTO, error) {
	if cfg.APIKey == "" {
		return nil, errors.New("کلید API لوت پاگلو تنظیم نشده است")
	}
	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = a.GetDefaultBaseURL()
	}

	req, err := http.NewRequest(http.MethodGet, baseURL+"/api/v1/products", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-API-Key", cfg.APIKey)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("خطا در دریافت کاتالوگ لوت پاگلو: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("کد خطای کاتالوگ لوت پاگلو: %d (%s)", resp.StatusCode, string(body))
	}

	var data struct {
		Status   string `json:"status"`
		Services []struct {
			ServiceID string `json:"service_id"`
			Name      string `json:"name"`
			Stock     int64  `json:"available_stock"`
			Prices    struct {
				UPIPrice    float64 `json:"upiPrice"`
				CryptoPrice float64 `json:"cryptoPrice"`
			} `json:"prices"`
		} `json:"services"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	var list []ProviderProductDTO
	for _, p := range data.Services {
		price := p.Prices.CryptoPrice
		if price <= 0 {
			price = p.Prices.UPIPrice / 85.0 // fallback conversion if needed
		}

		payloadType := "license_code"
		nameLower := strings.ToLower(p.Name)
		if strings.Contains(nameLower, "link") || strings.Contains(nameLower, "invite") {
			payloadType = "activation_link"
		} else if strings.Contains(nameLower, "account") {
			payloadType = "credentials_pair"
		}

		list = append(list, ProviderProductDTO{
			ID:                   p.ServiceID,
			Name:                 p.Name,
			PriceUSD:             price,
			Stock:                p.Stock,
			InStock:              p.Stock > 0,
			ProviderKey:          ProviderLootPaglu,
			ProviderNameFA:       a.GetProviderNameFA(),
			RequiredInput:        "none",
			DeliveredPayloadType: payloadType,
		})
	}

	return list, nil
}

func (a *LootPagluAdapter) CreateOrder(client *http.Client, cfg ProviderConfig, req ProviderOrderRequest) (*ProviderOrderResultDTO, error) {
	if cfg.APIKey == "" {
		return nil, errors.New("کلید API لوت پاگلو تنظیم نشده است")
	}
	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = a.GetDefaultBaseURL()
	}

	qty := req.Quantity
	if qty <= 0 {
		qty = 1
	}

	payload := map[string]interface{}{
		"service_id": req.ProductID,
		"quantity":   qty,
		"currency":   "crypto",
	}

	rawPayload, _ := json.Marshal(payload)
	httpReq, err := http.NewRequest(http.MethodPost, baseURL+"/api/v1/order", bytes.NewReader(rawPayload))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-API-Key", cfg.APIKey)
	if req.IdempotencyKey != "" {
		httpReq.Header.Set("Idempotency-Key", req.IdempotencyKey)
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("خطا در ایجاد سفارش لوت پاگلو: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, errors.New("کلید API لوت پاگلو نامعتبر است (401)")
	}
	if resp.StatusCode == http.StatusBadRequest {
		return nil, fmt.Errorf("خطای درخواست لوت پاگلو (400): %s", string(body))
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("خطای ثبت سفارش لوت پاگلو (%d): %s", resp.StatusCode, string(body))
	}

	var data struct {
		Status    string   `json:"status"`
		Message   string   `json:"message"`
		Success   bool     `json:"success"`
		OrderID   string   `json:"order_id"`
		TotalCost float64  `json:"total_cost"`
		Products  []string `json:"products"`
	}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("خطا در پارس پاسخ لوت پاگلو: %w", err)
	}

	delivered := ""
	payloadType := "license_code"
	if len(data.Products) > 0 {
		delivered = data.Products[0]
		if strings.HasPrefix(delivered, "http") {
			payloadType = "activation_link"
		}
	}

	return &ProviderOrderResultDTO{
		Success:          true,
		ProviderKey:      ProviderLootPaglu,
		OrderID:          data.OrderID,
		Codes:            data.Products,
		DeliveredPayload: delivered,
		PayloadType:      payloadType,
		UnitCostUSD:      data.TotalCost / float64(qty),
		TotalCostUSD:     data.TotalCost,
		RawJSON:          string(body),
	}, nil
}
