package fulfillment

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type SupremeLeaderAdapter struct{}

func (a *SupremeLeaderAdapter) GetProviderKey() string {
	return ProviderSupremeLeader
}

func (a *SupremeLeaderAdapter) GetProviderNameFA() string {
	return "Supreme Leader Shop"
}

func (a *SupremeLeaderAdapter) GetDefaultBaseURL() string {
	return "https://telegram-shop-bot-production-48fa.up.railway.app"
}

func (a *SupremeLeaderAdapter) GetBalance(client *http.Client, cfg ProviderConfig) (*ProviderBalanceDTO, error) {
	if cfg.APIKey == "" {
		return nil, errors.New("کلید API سوپریم لیدر شاپ تنظیم نشده است")
	}
	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = a.GetDefaultBaseURL()
	}

	targetURL := fmt.Sprintf("%s/api/telegram-buyer/balance?key=%s", baseURL, url.QueryEscape(cfg.APIKey))
	req, err := http.NewRequest(http.MethodGet, targetURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("خطا در اتصال به سوپریم لیدر: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, errors.New("کلید API سوپریم لیدر شاپ نامعتبر است (401)")
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("کد خطای HTTP سوپریم لیدر: %d (%s)", resp.StatusCode, string(body))
	}

	var data struct {
		Balance float64 `json:"balance"`
		Status  string  `json:"status"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	return &ProviderBalanceDTO{
		ProviderKey:    ProviderSupremeLeader,
		ProviderNameFA: a.GetProviderNameFA(),
		BalanceUSD:     data.Balance,
		Status:         data.Status,
		Raw:            data,
	}, nil
}

func (a *SupremeLeaderAdapter) GetCatalog(client *http.Client, cfg ProviderConfig) ([]ProviderProductDTO, error) {
	if cfg.APIKey == "" {
		return nil, errors.New("کلید API سوپریم لیدر شاپ تنظیم نشده است")
	}
	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = a.GetDefaultBaseURL()
	}

	targetURL := fmt.Sprintf("%s/api/telegram-buyer/products?key=%s", baseURL, url.QueryEscape(cfg.APIKey))
	req, err := http.NewRequest(http.MethodGet, targetURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("خطا در دریافت کاتالوگ سوپریم لیدر: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("کد خطای کاتالوگ سوپریم لیدر: %d (%s)", resp.StatusCode, string(body))
	}

	var data struct {
		Products []struct {
			ID          interface{} `json:"id"`
			ProductID   interface{} `json:"product_id"`
			Name        string      `json:"name"`
			Title       string      `json:"title"`
			Price       float64     `json:"price"`
			Stock       int64       `json:"stock"`
			Description string      `json:"description"`
		} `json:"products"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	var list []ProviderProductDTO
	for _, p := range data.Products {
		idStr := fmt.Sprintf("%v", p.ProductID)
		if idStr == "" || idStr == "<nil>" {
			idStr = fmt.Sprintf("%v", p.ID)
		}
		title := p.Name
		if title == "" {
			title = p.Title
		}

		payloadType := "license_code"
		nameLower := strings.ToLower(title)
		if strings.Contains(nameLower, "invite") || strings.Contains(nameLower, "link") {
			payloadType = "activation_link"
		} else if strings.Contains(nameLower, "account") {
			payloadType = "credentials_pair"
		}

		list = append(list, ProviderProductDTO{
			ID:                   idStr,
			Name:                 title,
			Description:          p.Description,
			PriceUSD:             p.Price,
			Stock:                p.Stock,
			InStock:              p.Stock > 0,
			ProviderKey:          ProviderSupremeLeader,
			ProviderNameFA:       a.GetProviderNameFA(),
			RequiredInput:        "none",
			DeliveredPayloadType: payloadType,
		})
	}

	return list, nil
}

func (a *SupremeLeaderAdapter) CreateOrder(client *http.Client, cfg ProviderConfig, req ProviderOrderRequest) (*ProviderOrderResultDTO, error) {
	if cfg.APIKey == "" {
		return nil, errors.New("کلید API سوپریم لیدر شاپ تنظیم نشده است")
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
		"key":        cfg.APIKey,
		"product_id": req.ProductID,
		"quantity":   qty,
		"lang":       "en",
	}

	rawPayload, _ := json.Marshal(payload)
	httpReq, err := http.NewRequest(http.MethodPost, baseURL+"/api/telegram-buyer/purchase", bytes.NewReader(rawPayload))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("خطا در ایجاد سفارش سوپریم لیدر: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, errors.New("کلید API سوپریم لیدر نامعتبر است (401)")
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("خطای ثبت سفارش سوپریم لیدر (%d): %s", resp.StatusCode, string(body))
	}

	var data struct {
		Success bool        `json:"success"`
		OrderID interface{} `json:"order_id"`
		Item    string      `json:"item"`
		Code    string      `json:"code"`
		Account string      `json:"account"`
		Error   string      `json:"error"`
	}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("خطا در پردازش پاسخ سوپریم لیدر: %w", err)
	}

	delivered := data.Code
	if delivered == "" {
		delivered = data.Item
	}
	if delivered == "" {
		delivered = data.Account
	}

	payloadType := "license_code"
	if strings.HasPrefix(delivered, "http") {
		payloadType = "activation_link"
	} else if strings.Contains(delivered, ":") {
		payloadType = "credentials_pair"
	}

	return &ProviderOrderResultDTO{
		Success:          true,
		ProviderKey:      ProviderSupremeLeader,
		OrderID:          fmt.Sprintf("%v", data.OrderID),
		Codes:            []string{delivered},
		DeliveredPayload: delivered,
		PayloadType:      payloadType,
		RawJSON:          string(body),
	}, nil
}
