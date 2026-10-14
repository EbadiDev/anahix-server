package fulfillment

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

const (
	DefaultBaseURL = "https://insightxstore-bot-production.up.railway.app"
	CacheTTL       = 5 * time.Minute
)

var (
	ErrInvalidAPIKey     = errors.New("invalid or missing fulfillment provider API key")
	ErrInsufficientFunds = errors.New("insufficient balance in fulfillment provider account")
	ErrProductNotFound   = errors.New("product not found on fulfillment provider catalog")
)

type Client struct {
	baseURL    string
	httpClient *http.Client

	mu          sync.RWMutex
	cachedProds []CatalogItem
	cacheTime   time.Time
}

func NewClient(baseURL string) *Client {
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

type HealthResponse struct {
	OK      bool   `json:"ok"`
	Service string `json:"service"`
	Version string `json:"version"`
}

type BalanceResponse struct {
	BalanceUSDT        float64       `json:"balance_usdt"`
	KeyPrefix          string        `json:"key_prefix"`
	Status             string        `json:"status"`
	RateLimitPerMin    int           `json:"rate_limit_per_min"`
	RecentTransactions []interface{} `json:"recent_transactions"`
}

type CatalogItem struct {
	ID              int64   `json:"id"`
	Name            string  `json:"name"`
	Emoji           string  `json:"emoji"`
	Description     string  `json:"description"`
	DescriptionText string  `json:"description_text"`
	PriceUSDT       float64 `json:"price_usdt"`
	Price           float64 `json:"price"`
	BasePriceUSDT   float64 `json:"base_price_usdt"`
	Stock           int64   `json:"stock"`
	Available       bool    `json:"available"`
	Unlimited       bool    `json:"unlimited"`
}

type ProductsListResponse struct {
	Products []CatalogItem `json:"products"`
}

type CreateOrderRequest struct {
	ProductID int64 `json:"product_id"`
	Quantity  int   `json:"quantity"`
}

type OrderResult struct {
	OrderID   interface{} `json:"order_id"`
	ID        interface{} `json:"id"`
	Status    string      `json:"status"`
	ProductID int64       `json:"product_id"`
	Quantity  int         `json:"quantity"`
	Codes     []string    `json:"codes"`
	Code      string      `json:"code"`
	Data      interface{} `json:"data"`
	RawJSON   string      `json:"raw_json,omitempty"`
}

func (c *Client) GetHealth() (*HealthResponse, error) {
	resp, err := c.httpClient.Get(c.baseURL + "/api/v1/health")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var h HealthResponse
	if err := json.NewDecoder(resp.Body).Decode(&h); err != nil {
		return nil, err
	}
	return &h, nil
}

func (c *Client) GetBalance(apiKey string) (*BalanceResponse, error) {
	if apiKey == "" {
		return nil, ErrInvalidAPIKey
	}

	req, err := http.NewRequest(http.MethodGet, c.baseURL+"/api/v1/balance", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fulfillment provider balance request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, ErrInvalidAPIKey
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("fulfillment provider balance error (status %d): %s", resp.StatusCode, string(body))
	}

	var b BalanceResponse
	if err := json.NewDecoder(resp.Body).Decode(&b); err != nil {
		return nil, err
	}
	return &b, nil
}

func (c *Client) GetProducts(apiKey string, forceRefresh bool) ([]CatalogItem, error) {
	if apiKey == "" {
		return nil, ErrInvalidAPIKey
	}

	c.mu.RLock()
	if !forceRefresh && len(c.cachedProds) > 0 && time.Since(c.cacheTime) < CacheTTL {
		prods := c.cachedProds
		c.mu.RUnlock()
		return prods, nil
	}
	c.mu.RUnlock()

	req, err := http.NewRequest(http.MethodGet, c.baseURL+"/api/v1/products", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fulfillment provider catalog request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, ErrInvalidAPIKey
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("fulfillment provider catalog error (status %d): %s", resp.StatusCode, string(body))
	}

	var pResp ProductsListResponse
	if err := json.NewDecoder(resp.Body).Decode(&pResp); err != nil {
		return nil, fmt.Errorf("failed to decode catalog response: %w", err)
	}

	c.mu.Lock()
	c.cachedProds = pResp.Products
	c.cacheTime = time.Now()
	c.mu.Unlock()

	return pResp.Products, nil
}

func (c *Client) CreateOrder(apiKey string, idempotencyKey string, productID int64, quantity int) (*OrderResult, error) {
	if apiKey == "" {
		return nil, ErrInvalidAPIKey
	}
	if quantity <= 0 {
		quantity = 1
	}

	payload := CreateOrderRequest{
		ProductID: productID,
		Quantity:  quantity,
	}

	rawPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/api/v1/orders", bytes.NewReader(rawPayload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	if idempotencyKey != "" {
		req.Header.Set("Idempotency-Key", idempotencyKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fulfillment provider order creation failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, ErrInvalidAPIKey
	}
	if resp.StatusCode == http.StatusPaymentRequired {
		return nil, ErrInsufficientFunds
	}
	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrProductNotFound
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("order failed with status %d: %s", resp.StatusCode, string(body))
	}

	var res OrderResult
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, fmt.Errorf("failed to parse order response: %w", err)
	}
	res.RawJSON = string(body)

	return &res, nil
}
