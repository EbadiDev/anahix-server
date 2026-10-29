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

type StackVaultAdapter struct{}

func (a *StackVaultAdapter) GetProviderKey() string {
	return ProviderStackVault
}

func (a *StackVaultAdapter) GetProviderNameFA() string {
	return "StackVault Digital"
}

func (a *StackVaultAdapter) GetDefaultBaseURL() string {
	return "https://decohomz.com/reseller-api/v1"
}

func (a *StackVaultAdapter) GetBalance(client *http.Client, cfg ProviderConfig) (*ProviderBalanceDTO, error) {
	if cfg.APIKey == "" {
		return nil, errors.New("کلید API استک‌ولت تنظیم نشده است")
	}
	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = a.GetDefaultBaseURL()
	}

	req, err := http.NewRequest(http.MethodGet, baseURL+"/balance", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+cfg.APIKey)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("خطا در اتصال به استک‌ولت: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, errors.New("کلید API استک‌ولت نامعتبر است (401)")
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("کد خطای HTTP استک‌ولت: %d (%s)", resp.StatusCode, string(body))
	}

	var data struct {
		OK      bool    `json:"ok"`
		Balance float64 `json:"balance"`
		Name    string  `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	return &ProviderBalanceDTO{
		ProviderKey:    ProviderStackVault,
		ProviderNameFA: a.GetProviderNameFA(),
		BalanceUSD:     data.Balance,
		Status:         "active",
		Raw:            data,
	}, nil
}

func (a *StackVaultAdapter) GetCatalog(client *http.Client, cfg ProviderConfig) ([]ProviderProductDTO, error) {
	if cfg.APIKey == "" {
		return nil, errors.New("کلید API استک‌ولت تنظیم نشده است")
	}
	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = a.GetDefaultBaseURL()
	}

	req, err := http.NewRequest(http.MethodGet, baseURL+"/products", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+cfg.APIKey)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("خطا در دریافت کاتالوگ استک‌ولت: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("کد خطای کاتالوگ استک‌ولت: %d (%s)", resp.StatusCode, string(body))
	}

	var data struct {
		OK       bool `json:"ok"`
		Products []struct {
			ID      string  `json:"id"`
			Name    string  `json:"name"`
			Price   float64 `json:"price"`
			Stock   int64   `json:"stock"`
			InStock bool    `json:"in_stock"`
		} `json:"products"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	var list []ProviderProductDTO
	for _, p := range data.Products {
		payloadType := "license_code"
		nameLower := strings.ToLower(p.Name)
		if strings.Contains(nameLower, "gift") || strings.Contains(nameLower, "invite") || strings.Contains(nameLower, "link") {
			payloadType = "activation_link"
		} else if strings.Contains(nameLower, "account") || strings.Contains(nameLower, "aged") {
			payloadType = "credentials_pair"
		}

		requiredInput := "none"
		if strings.Contains(nameLower, "stars") || strings.Contains(nameLower, "premium") {
			requiredInput = "telegram_username"
			payloadType = "direct_service"
		}

		list = append(list, ProviderProductDTO{
			ID:                   p.ID,
			Name:                 p.Name,
			PriceUSD:             p.Price,
			Stock:                p.Stock,
			InStock:              p.InStock || p.Stock > 0,
			ProviderKey:          ProviderStackVault,
			ProviderNameFA:       a.GetProviderNameFA(),
			RequiredInput:        requiredInput,
			DeliveredPayloadType: payloadType,
		})
	}

	return list, nil
}

func (a *StackVaultAdapter) CreateOrder(client *http.Client, cfg ProviderConfig, req ProviderOrderRequest) (*ProviderOrderResultDTO, error) {
	if cfg.APIKey == "" {
		return nil, errors.New("کلید API استک‌ولت تنظیم نشده است")
	}
	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = a.GetDefaultBaseURL()
	}

	qty := req.Quantity
	if qty <= 0 {
		qty = 1
	}

	endpoint := baseURL + "/orders"
	payload := map[string]interface{}{
		"product_id": req.ProductID,
		"quantity":   qty,
	}
	if req.IdempotencyKey != "" {
		payload["idempotency_key"] = req.IdempotencyKey
	}

	// Handle Telegram Stars / Premium special orders if requested
	if req.TelegramUsername != "" {
		cleanUsername := strings.TrimPrefix(req.TelegramUsername, "@")
		if strings.Contains(strings.ToLower(req.ProductID), "stars") {
			endpoint = baseURL + "/stars/order"
			payload = map[string]interface{}{
				"username": cleanUsername,
				"amount":   qty,
			}
			if req.IdempotencyKey != "" {
				payload["idempotency_key"] = req.IdempotencyKey
			}
		} else if strings.Contains(strings.ToLower(req.ProductID), "premium") {
			endpoint = baseURL + "/premium/order"
			months := 3
			if strings.Contains(req.ProductID, "6") {
				months = 6
			} else if strings.Contains(req.ProductID, "12") {
				months = 12
			}
			payload = map[string]interface{}{
				"username": cleanUsername,
				"months":   months,
			}
			if req.IdempotencyKey != "" {
				payload["idempotency_key"] = req.IdempotencyKey
			}
		}
	}

	rawPayload, _ := json.Marshal(payload)
	httpReq, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(rawPayload))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+cfg.APIKey)

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("خطا در ایجاد سفارش استک‌ولت: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, errors.New("کلید API استک‌ولت نامعتبر است (401)")
	}
	if resp.StatusCode == http.StatusPaymentRequired {
		return nil, errors.New("موجودی حساب استک‌ولت ناکافی است (402)")
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("خطای ثبت سفارش استک‌ولت (%d): %s", resp.StatusCode, string(body))
	}

	var data struct {
		OK                  bool        `json:"ok"`
		OrderID             interface{} `json:"order_id"`
		Delivery            []string    `json:"delivery"`
		DeliveryInstruction string      `json:"delivery_instruction"`
		UnitPrice           float64     `json:"unit_price"`
		Total               float64     `json:"total"`
	}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("خطا در پردازش پاسخ استک‌ولت: %w", err)
	}

	delivered := ""
	payloadType := "license_code"
	if len(data.Delivery) > 0 {
		delivered = data.Delivery[0]
		if strings.HasPrefix(delivered, "http") {
			payloadType = "activation_link"
		}
	} else if req.TelegramUsername != "" {
		delivered = "اعمال موفق روی اکانت تلگرام @" + strings.TrimPrefix(req.TelegramUsername, "@")
		payloadType = "direct_service"
	}

	return &ProviderOrderResultDTO{
		Success:          true,
		ProviderKey:      ProviderStackVault,
		OrderID:          fmt.Sprintf("%v", data.OrderID),
		Links:            data.Delivery,
		DeliveredPayload: delivered,
		PayloadType:      payloadType,
		Instructions:     data.DeliveryInstruction,
		UnitCostUSD:      data.UnitPrice,
		TotalCostUSD:     data.Total,
		RawJSON:          string(body),
	}, nil
}
