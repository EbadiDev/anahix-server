package sms

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

var (
	ErrNoAPIKey        = errors.New("SMS provider API key is not configured")
	ErrOrderNotFound   = errors.New("order not found on SMS provider")
	ErrInsufficientBal = errors.New("insufficient balance on SMS provider")
	ErrNoNumbersAvail  = errors.New("no virtual numbers currently available for this service/country")
)

type ProfileResponse struct {
	ID                      int64   `json:"id"`
	Email                   string  `json:"email"`
	Vendor                  string  `json:"vendor"`
	DefaultForwardingNumber string  `json:"default_forwarding_number"`
	Balance                 float64 `json:"balance"`
	Rating                  float64 `json:"rating"`
	IsMock                  bool    `json:"is_mock,omitempty"`
}

type SMSItem struct {
	ID        int64  `json:"id"`
	CreatedAt string `json:"created_at"`
	Date      string `json:"date"`
	Sender    string `json:"sender"`
	Text      string `json:"text"`
	Code      string `json:"code"`
}

type OrderResponse struct {
	ID               int64     `json:"id"`
	Phone            string    `json:"phone"`
	Operator         string    `json:"operator"`
	Product          string    `json:"product"`
	Price            float64   `json:"price"`
	Status           string    `json:"status"` // PENDING, RECEIVED, FINISHED, CANCELED, TIMEOUT
	Expires          string    `json:"expires"`
	SMS              []SMSItem `json:"sms"`
	CreatedAt        string    `json:"created_at"`
	Forwarding       bool      `json:"forwarding"`
	ForwardingNumber string    `json:"forwarding_number"`
	Country          string    `json:"country"`
	IsMock           bool      `json:"is_mock,omitempty"`
}

type ProductRawInfo struct {
	Category string  `json:"Category"`
	Qty      int     `json:"Qty"`
	Price    float64 `json:"Price"`
}

type OperatorRawInfo struct {
	Cost  float64 `json:"cost"`
	Count int     `json:"count"`
	Rate  float64 `json:"rate"`
}

type Client struct {
	baseURL      string
	httpClient   *http.Client
	mu           sync.RWMutex
	productCache map[string]ProductRawInfo
	productTime  time.Time
	priceCache   map[string]map[string]map[string]OperatorRawInfo
	priceTime    map[string]time.Time
}

func NewClient() *Client {
	return &Client{
		baseURL: "https://5sim.net/v1",
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
		productCache: make(map[string]ProductRawInfo),
		priceCache:   make(map[string]map[string]map[string]OperatorRawInfo),
		priceTime:    make(map[string]time.Time),
	}
}

// GetProducts fetches the list of all available services from the upstream SMS provider
func (c *Client) GetProducts() (map[string]ProductRawInfo, error) {
	c.mu.RLock()
	if len(c.productCache) > 0 && time.Since(c.productTime) < 10*time.Minute {
		data := c.productCache
		c.mu.RUnlock()
		return data, nil
	}
	c.mu.RUnlock()

	url := fmt.Sprintf("%s/guest/products/any/any", c.baseURL)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		c.mu.RLock()
		if len(c.productCache) > 0 {
			defer c.mu.RUnlock()
			return c.productCache, nil
		}
		c.mu.RUnlock()
		return nil, fmt.Errorf("خطا در ارتباط با زیرساخت شماره‌های مجازی: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result map[string]ProductRawInfo
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("خطا در پردازش لیست سرویس‌های سامانه شماره مجازی: %w", err)
	}

	c.mu.Lock()
	c.productCache = result
	c.productTime = time.Now()
	c.mu.Unlock()

	return result, nil
}

// GetPrices fetches live prices and operator stock for a specific service
func (c *Client) GetPrices(product string) (map[string]map[string]OperatorRawInfo, error) {
	product = strings.ToLower(strings.TrimSpace(product))
	if product == "" {
		return nil, errors.New("نام سرویس برای دریافت قیمت‌ها مشخص نشده است")
	}

	c.mu.RLock()
	if cached, ok := c.priceCache[product]; ok && time.Since(c.priceTime[product]) < 2*time.Minute {
		c.mu.RUnlock()
		return cached, nil
	}
	c.mu.RUnlock()

	url := fmt.Sprintf("%s/guest/prices?product=%s", c.baseURL, product)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		c.mu.RLock()
		if cached, ok := c.priceCache[product]; ok {
			defer c.mu.RUnlock()
			return cached, nil
		}
		c.mu.RUnlock()
		return nil, fmt.Errorf("خطا در استعلام اطلاعات سرویس %s: %w", product, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var fullResult map[string]map[string]map[string]OperatorRawInfo
	if err := json.Unmarshal(body, &fullResult); err != nil {
		return nil, fmt.Errorf("خطا در دریافت قیمت‌های سامانه شماره مجازی: %w", err)
	}

	countriesMap := fullResult[product]
	if countriesMap == nil {
		countriesMap = make(map[string]map[string]OperatorRawInfo)
	}

	c.mu.Lock()
	c.priceCache[product] = countriesMap
	c.priceTime[product] = time.Now()
	c.mu.Unlock()

	return countriesMap, nil
}

// GetProfile checks SMS provider account balance and validity of the API Key
func (c *Client) GetProfile(apiKey string) (*ProfileResponse, error) {
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return &ProfileResponse{
			ID:      99999,
			Email:   "demo@anahix.ai",
			Balance: 0.0,
			Rating:  100,
			IsMock:  true,
		}, nil
	}

	url := fmt.Sprintf("%s/user/profile", c.baseURL)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("SMS provider network error: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return nil, errors.New("کلید اتصال زیرساخت شماره نامعتبر است (Unauthorized)")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("SMS provider error status %d: %s", resp.StatusCode, string(body))
	}

	var prof ProfileResponse
	if err := json.Unmarshal(body, &prof); err != nil {
		return nil, fmt.Errorf("failed to parse profile: %w", err)
	}
	return &prof, nil
}

// BuyActivation purchases a virtual number for receiving SMS
func (c *Client) BuyActivation(apiKey, country, operator, product string) (*OrderResponse, error) {
	apiKey = strings.TrimSpace(apiKey)
	country = strings.ToLower(strings.TrimSpace(country))
	operator = strings.ToLower(strings.TrimSpace(operator))
	product = strings.ToLower(strings.TrimSpace(product))

	if operator == "" {
		operator = "any"
	}

	// Mock mode fallback when admin has not set API key yet
	if apiKey == "" {
		mockID := time.Now().UnixNano() % 10000000
		phonePrefix := map[string]string{
			"england":     "+447911",
			"germany":     "+49151",
			"france":      "+33612",
			"netherlands": "+31628",
			"sweden":      "+46701",
			"usa":         "+12025",
			"canada":      "+14165",
			"indonesia":   "+62812",
			"malaysia":    "+6012",
		}
		prefix := phonePrefix[country]
		if prefix == "" {
			prefix = "+447900"
		}

		phone := fmt.Sprintf("%s%05d", prefix, mockID%100000)
		expires := time.Now().Add(15 * time.Minute).Format(time.RFC3339)
		return &OrderResponse{
			ID:        mockID,
			Phone:     phone,
			Operator:  "smart",
			Product:   product,
			Price:     18.5,
			Status:    "PENDING",
			Expires:   expires,
			SMS:       []SMSItem{},
			CreatedAt: time.Now().Format(time.RFC3339),
			Country:   country,
			IsMock:    true,
		}, nil
	}

	url := fmt.Sprintf("%s/user/buy/activation/%s/%s/%s", c.baseURL, country, operator, product)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("SMS provider network error: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	bodyStr := strings.TrimSpace(string(body))
	if strings.Contains(bodyStr, "no free phones") || strings.Contains(bodyStr, "not enough user balance") {
		if strings.Contains(bodyStr, "not enough") {
			return nil, ErrInsufficientBal
		}
		return nil, ErrNoNumbersAvail
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("SMS provider error (%d): %s", resp.StatusCode, bodyStr)
	}

	var ord OrderResponse
	if err := json.Unmarshal(body, &ord); err != nil {
		return nil, fmt.Errorf("failed to parse order response: %w", err)
	}
	return &ord, nil
}

// CheckOrder gets updated status and incoming SMS list
func (c *Client) CheckOrder(apiKey string, orderID int64, createdAt time.Time, product string) (*OrderResponse, error) {
	apiKey = strings.TrimSpace(apiKey)

	// Mock mode behavior
	if apiKey == "" {
		elapsed := time.Since(createdAt)
		status := "PENDING"
		var sms []SMSItem

		if elapsed > 12*time.Second {
			status = "RECEIVED"
			otpCode := fmt.Sprintf("%05d", (orderID*7)%90000+10000)
			sms = append(sms, SMSItem{
				ID:        orderID + 100,
				CreatedAt: time.Now().Format(time.RFC3339),
				Date:      time.Now().Format("2006-01-02 15:04:05"),
				Sender:    strings.ToUpper(product),
				Text:      fmt.Sprintf("Your %s verification code is: %s. Valid for 5 minutes.", strings.ToUpper(product), otpCode),
				Code:      otpCode,
			})
		}

		return &OrderResponse{
			ID:      orderID,
			Status:  status,
			SMS:     sms,
			IsMock:  true,
			Expires: createdAt.Add(15 * time.Minute).Format(time.RFC3339),
		}, nil
	}

	url := fmt.Sprintf("%s/user/check/%d", c.baseURL, orderID)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("SMS provider network error: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrOrderNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("SMS provider error (%d): %s", resp.StatusCode, string(body))
	}

	var ord OrderResponse
	if err := json.Unmarshal(body, &ord); err != nil {
		return nil, fmt.Errorf("failed to parse check order response: %w", err)
	}
	return &ord, nil
}

// FinishOrder finalizes activation after code is received
func (c *Client) FinishOrder(apiKey string, orderID int64) error {
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return nil
	}

	url := fmt.Sprintf("%s/user/finish/%d", c.baseURL, orderID)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

// CancelOrder cancels activation and refunds balance
func (c *Client) CancelOrder(apiKey string, orderID int64) error {
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return nil
	}

	url := fmt.Sprintf("%s/user/cancel/%d", c.baseURL, orderID)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}
