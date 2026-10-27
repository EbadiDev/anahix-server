package fulfillment

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

type JeroAccountsAdapter struct{}

func (a *JeroAccountsAdapter) GetProviderKey() string {
	return ProviderJeroAccounts
}

func (a *JeroAccountsAdapter) GetProviderNameFA() string {
	return "جرو اکانتس (Jero Accounts)"
}

func (a *JeroAccountsAdapter) GetDefaultBaseURL() string {
	return "https://api.jeroaccounts.com/api/reseller"
}

func (a *JeroAccountsAdapter) GetBalance(client *http.Client, cfg ProviderConfig) (*ProviderBalanceDTO, error) {
	if cfg.APIKey == "" {
		return nil, errors.New("کلید API جرو اکانتس تنظیم نشده است")
	}
	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = a.GetDefaultBaseURL()
	}

	req, err := http.NewRequest(http.MethodGet, baseURL+"?action=balance", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+cfg.APIKey)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("خطا در اتصال به جرو اکانتس: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, errors.New("کلید API جرو اکانتس نامعتبر است (401)")
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("کد خطای HTTP جرو اکانتس: %d (%s)", resp.StatusCode, string(body))
	}

	var data struct {
		Balance float64 `json:"balance"`
		Success *bool   `json:"success,omitempty"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	return &ProviderBalanceDTO{
		ProviderKey:    ProviderJeroAccounts,
		ProviderNameFA: a.GetProviderNameFA(),
		BalanceUSD:     data.Balance,
		Status:         "active",
		Raw:            data,
	}, nil
}

func (a *JeroAccountsAdapter) GetCatalog(client *http.Client, cfg ProviderConfig) ([]ProviderProductDTO, error) {
	if cfg.APIKey == "" {
		return nil, errors.New("کلید API جرو اکانتس تنظیم نشده است")
	}
	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = a.GetDefaultBaseURL()
	}

	req, err := http.NewRequest(http.MethodGet, baseURL+"?action=products", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+cfg.APIKey)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("خطا در دریافت کاتالوگ جرو اکانتس: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("کد خطای کاتالوگ جرو اکانتس: %d (%s)", resp.StatusCode, string(body))
	}

	var products []struct {
		ID    int64       `json:"id"`
		Title string      `json:"title"`
		Price float64     `json:"price"`
		Stock interface{} `json:"stock"`
		Tiers interface{} `json:"tiers"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&products); err != nil {
		return nil, err
	}

	var list []ProviderProductDTO
	for _, p := range products {
		var stockCount int64 = 100
		if s, ok := p.Stock.(float64); ok {
			stockCount = int64(s)
		}

		requiredInput := "none"
		payloadType := "credentials_pair"
		titleLower := strings.ToLower(p.Title)

		if p.ID == 900001 || p.ID == 900003 || strings.Contains(titleLower, "elementor") || strings.Contains(titleLower, "rank math") {
			requiredInput = "wordpress_admin"
			payloadType = "direct_service"
		} else if p.ID == 900002 || strings.Contains(titleLower, "ui8") {
			requiredInput = "custom_links"
			payloadType = "files_or_downloads"
		} else if strings.Contains(titleLower, "key") || strings.Contains(titleLower, "license") {
			payloadType = "license_code"
		} else if strings.Contains(titleLower, "invite") || strings.Contains(titleLower, "link") {
			payloadType = "activation_link"
		}

		list = append(list, ProviderProductDTO{
			ID:                   strconv.FormatInt(p.ID, 10),
			NumericID:            p.ID,
			Name:                 p.Title,
			PriceUSD:             p.Price,
			Stock:                stockCount,
			InStock:              stockCount > 0,
			ProviderKey:          ProviderJeroAccounts,
			ProviderNameFA:       a.GetProviderNameFA(),
			RequiredInput:        requiredInput,
			DeliveredPayloadType: payloadType,
		})
	}

	return list, nil
}

func (a *JeroAccountsAdapter) CreateOrder(client *http.Client, cfg ProviderConfig, req ProviderOrderRequest) (*ProviderOrderResultDTO, error) {
	if cfg.APIKey == "" {
		return nil, errors.New("کلید API جرو اکانتس تنظیم نشده است")
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

	qty := req.Quantity
	if qty <= 0 {
		qty = 1
	}

	externalID := req.IdempotencyKey
	if externalID == "" {
		externalID = fmt.Sprintf("ANAHIX_%d", prodID)
	}

	payload := map[string]interface{}{
		"product_id":        prodID,
		"quantity":          qty,
		"external_order_id": externalID,
	}

	// WordPress Auto-Install product (Elementor Pro 900001, Rank Math 900003)
	if req.WordPressSite != "" {
		payload["site"] = req.WordPressSite
		payload["username"] = req.WordPressUser
		payload["password"] = req.WordPressPass
	}

	// UI8 Download product (900002)
	if len(req.CustomLinks) > 0 {
		payload["links"] = req.CustomLinks
		payload["quantity"] = len(req.CustomLinks)
	}

	rawPayload, _ := json.Marshal(payload)
	httpReq, err := http.NewRequest(http.MethodPost, baseURL+"?action=order", bytes.NewReader(rawPayload))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+cfg.APIKey)

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("خطا در ایجاد سفارش جرو اکانتس: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, errors.New("کلید API جرو اکانتس نامعتبر است (401)")
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("خطای ثبت سفارش جرو اکانتس (%d): %s", resp.StatusCode, string(body))
	}

	var data struct {
		Success *bool       `json:"success,omitempty"`
		OrderID interface{} `json:"order_id"`
		Items   []string    `json:"items"`
		Total   float64     `json:"total"`
		Balance float64     `json:"balance"`
		Status  string      `json:"status,omitempty"`
		Error   *struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error,omitempty"`
	}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("خطا در پردازش پاسخ جرو اکانتس: %w", err)
	}

	if data.Error != nil && data.Error.Message != "" {
		return nil, fmt.Errorf("خطای ارسالی جرو: %s (%s)", data.Error.Message, data.Error.Code)
	}

	delivered := ""
	payloadType := "license_code"
	if len(data.Items) > 0 {
		delivered = data.Items[0]
		if strings.HasPrefix(delivered, "http") {
			payloadType = "activation_link"
		} else if strings.Contains(delivered, ":") {
			payloadType = "credentials_pair"
		}
	} else if req.WordPressSite != "" {
		delivered = fmt.Sprintf("نصب و لایسنس خودکار روی سایت %s با موفقیت انجام شد", req.WordPressSite)
		payloadType = "direct_service"
	}

	return &ProviderOrderResultDTO{
		Success:          true,
		ProviderKey:      ProviderJeroAccounts,
		OrderID:          fmt.Sprintf("%v", data.OrderID),
		Codes:            data.Items,
		DeliveredPayload: delivered,
		PayloadType:      payloadType,
		TotalCostUSD:     data.Total,
		RawJSON:          string(body),
	}, nil
}
