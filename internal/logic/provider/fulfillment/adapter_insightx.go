package fulfillment

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

type InsightXAdapter struct{}

func (a *InsightXAdapter) GetProviderKey() string {
	return ProviderInsightX
}

func (a *InsightXAdapter) GetProviderNameFA() string {
	return "InsightX Store"
}

func (a *InsightXAdapter) GetDefaultBaseURL() string {
	return "https://api.insightxpro.store"
}

func (a *InsightXAdapter) GetBalance(client *http.Client, cfg ProviderConfig) (*ProviderBalanceDTO, error) {
	if cfg.APIKey == "" {
		return nil, errors.New("کلید API اینسایت استور تنظیم نشده است")
	}
	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = a.GetDefaultBaseURL()
	}

	req, err := http.NewRequest(http.MethodGet, baseURL+"/api/v1/balance", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
	req.Header.Set("X-API-Key", cfg.APIKey)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("خطا در اتصال به اینسایت استور: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, errors.New("کلید API اینسایت استور نامعتبر یا منقضی است (401)")
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("کد خطای HTTP اینسایت استور: %d (%s)", resp.StatusCode, string(body))
	}

	var data struct {
		BalanceUSDT     float64 `json:"balance_usdt"`
		Status          string  `json:"status"`
		RateLimitPerMin int     `json:"rate_limit_per_min"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	return &ProviderBalanceDTO{
		ProviderKey:     ProviderInsightX,
		ProviderNameFA:  a.GetProviderNameFA(),
		BalanceUSD:      data.BalanceUSDT,
		Status:          data.Status,
		RateLimitPerMin: data.RateLimitPerMin,
		Raw:             data,
	}, nil
}

func (a *InsightXAdapter) GetCatalog(client *http.Client, cfg ProviderConfig) ([]ProviderProductDTO, error) {
	if cfg.APIKey == "" {
		return nil, errors.New("کلید API اینسایت استور تنظیم نشده است")
	}
	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = a.GetDefaultBaseURL()
	}

	req, err := http.NewRequest(http.MethodGet, baseURL+"/api/v1/products", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
	req.Header.Set("X-API-Key", cfg.APIKey)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("خطا در دریافت کاتالوگ اینسایت استور: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("کد خطای کاتالوگ اینسایت استور: %d (%s)", resp.StatusCode, string(body))
	}

	var data struct {
		Products []struct {
			ID              int64   `json:"id"`
			Name            string  `json:"name"`
			Emoji           string  `json:"emoji"`
			DescriptionText string  `json:"description_text"`
			PriceUSDT       float64 `json:"price_usdt"`
			Price           float64 `json:"price"`
			BasePriceUSDT   float64 `json:"base_price_usdt"`
			Stock           int64   `json:"stock"`
			Available       bool    `json:"available"`
			Unlimited       bool    `json:"unlimited"`
		} `json:"products"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	var list []ProviderProductDTO
	for _, p := range data.Products {
		price := p.PriceUSDT
		if price <= 0 {
			price = p.Price
		}
		if price <= 0 {
			price = p.BasePriceUSDT
		}
		list = append(list, ProviderProductDTO{
			ID:                   strconv.FormatInt(p.ID, 10),
			NumericID:            p.ID,
			Name:                 p.Name,
			Description:          p.DescriptionText,
			PriceUSD:             price,
			Stock:                p.Stock,
			InStock:              p.Available || p.Stock > 0 || p.Unlimited,
			Unlimited:            p.Unlimited,
			Emoji:                p.Emoji,
			ProviderKey:          ProviderInsightX,
			ProviderNameFA:       a.GetProviderNameFA(),
			RequiredInput:        "none",
			DeliveredPayloadType: "license_code",
		})
	}

	return list, nil
}

func (a *InsightXAdapter) CreateOrder(client *http.Client, cfg ProviderConfig, req ProviderOrderRequest) (*ProviderOrderResultDTO, error) {
	if cfg.APIKey == "" {
		return nil, errors.New("کلید API اینسایت استور تنظیم نشده است")
	}
	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = a.GetDefaultBaseURL()
	}

	prodID := req.NumericID
	if prodID <= 0 {
		if id, err := strconv.ParseInt(req.ProductID, 10, 64); err == nil {
			prodID = id
		}
	}
	if prodID <= 0 {
		return nil, fmt.Errorf("شناسه نامعتبر محصول برای اینسایت استور: %s", req.ProductID)
	}

	qty := req.Quantity
	if qty <= 0 {
		qty = 1
	}

	payload := map[string]interface{}{
		"product_id": prodID,
		"quantity":   qty,
	}
	if req.IdempotencyKey != "" {
		payload["idempotency_key"] = req.IdempotencyKey
	}

	rawPayload, _ := json.Marshal(payload)
	httpReq, err := http.NewRequest(http.MethodPost, baseURL+"/api/v1/orders", bytes.NewReader(rawPayload))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+cfg.APIKey)
	httpReq.Header.Set("X-API-Key", cfg.APIKey)
	if req.IdempotencyKey != "" {
		httpReq.Header.Set("Idempotency-Key", req.IdempotencyKey)
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("خطا در ایجاد سفارش اینسایت استور: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, errors.New("کلید API نامعتبر است (401)")
	}
	if resp.StatusCode == http.StatusPaymentRequired {
		return nil, errors.New("موجودی حساب اینسایت استور ناکافی است (402)")
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("خطای ثبت سفارش اینسایت استور (%d): %s", resp.StatusCode, string(body))
	}

	var data struct {
		Order struct {
			ID             interface{} `json:"id"`
			OrderID        interface{} `json:"order_id"`
			Status         string      `json:"status"`
			Codes          []string    `json:"codes"`
			DeliveredCodes []string    `json:"delivered_codes"`
			UnitCost       float64     `json:"unit_price_usdt"`
			TotalCost      float64     `json:"total_usdt"`
		} `json:"order"`
	}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("خطا در خواندن پاسخ سفارش: %w", err)
	}

	codes := data.Order.Codes
	if len(codes) == 0 {
		codes = data.Order.DeliveredCodes
	}

	delivered := ""
	if len(codes) > 0 {
		delivered = codes[0]
	}

	oid := fmt.Sprintf("%v", data.Order.OrderID)
	if oid == "" || oid == "<nil>" {
		oid = fmt.Sprintf("%v", data.Order.ID)
	}

	return &ProviderOrderResultDTO{
		Success:          true,
		ProviderKey:      ProviderInsightX,
		OrderID:          oid,
		Codes:            codes,
		DeliveredPayload: delivered,
		PayloadType:      "license_code",
		UnitCostUSD:      data.Order.UnitCost,
		TotalCostUSD:     data.Order.TotalCost,
		RawJSON:          string(body),
	}, nil
}
