package fulfillment

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"
	"testing"
)

const testAPIKey = "isk_live_jexXb9akFkevDfAGpybP9AM4eI0fNqhE"

func setupMockFulfillmentServer(t *testing.T) (*httptest.Server, *int32) {
	var productsCallCount int32

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/health":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(HealthResponse{
				OK:      true,
				Service: "insightx-fulfillment",
				Version: "1.0.0",
			})

		case "/api/v1/balance":
			auth := r.Header.Get("Authorization")
			if !strings.HasPrefix(auth, "Bearer ") || strings.TrimPrefix(auth, "Bearer ") != testAPIKey {
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte(`{"status":"error","message":"invalid api key"}`))
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(BalanceResponse{
				BalanceUSDT:     9.85,
				KeyPrefix:       "isk_live",
				Status:          "active",
				RateLimitPerMin: 60,
			})

		case "/api/v1/products":
			auth := r.Header.Get("Authorization")
			if !strings.HasPrefix(auth, "Bearer ") || strings.TrimPrefix(auth, "Bearer ") != testAPIKey {
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte(`{"status":"error","message":"invalid api key"}`))
				return
			}
			atomic.AddInt32(&productsCallCount, 1)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(ProductsListResponse{
				Products: []CatalogItem{
					{
						ID:              14,
						Name:            "Gemini Pro 18 Months",
						Emoji:           "✨",
						Description:     "Google Gemini Advanced subscription",
						DescriptionText: "Activation link",
						PriceUSDT:       12.50,
						Price:           12.50,
						BasePriceUSDT:   10.00,
						Stock:           42,
						Available:       true,
						Unlimited:       false,
					},
				},
			})

		case "/api/v1/orders":
			auth := r.Header.Get("Authorization")
			if !strings.HasPrefix(auth, "Bearer ") || strings.TrimPrefix(auth, "Bearer ") != testAPIKey {
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte(`{"status":"error","message":"invalid api key"}`))
				return
			}

			var req CreateOrderRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			if req.ProductID == 9999 {
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write([]byte(`{"status":"error","message":"product not found"}`))
				return
			}
			if req.ProductID == 8888 {
				w.WriteHeader(http.StatusPaymentRequired)
				_, _ = w.Write([]byte(`{"status":"error","message":"insufficient funds"}`))
				return
			}

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(OrderResult{
				OrderID:   1001,
				ID:        1001,
				Status:    "completed",
				ProductID: req.ProductID,
				Quantity:  req.Quantity,
				Codes:     []string{"https://gemini.google.com/redeem?code=TEST_PROMO_CODE"},
				Code:      "https://gemini.google.com/redeem?code=TEST_PROMO_CODE",
			})

		default:
			http.NotFound(w, r)
		}
	})

	server := httptest.NewServer(handler)
	return server, &productsCallCount
}

func TestNewClientDefaults(t *testing.T) {
	c := NewClient("")
	if c.baseURL != DefaultBaseURL {
		t.Errorf("expected default base URL %q, got %q", DefaultBaseURL, c.baseURL)
	}
}

func TestFulfillmentGetHealth(t *testing.T) {
	server, _ := setupMockFulfillmentServer(t)
	defer server.Close()

	c := NewClient(server.URL)
	h, err := c.GetHealth()
	if err != nil {
		t.Fatalf("GetHealth failed: %v", err)
	}
	if !h.OK {
		t.Errorf("expected health ok=true, got %v", h.OK)
	}
	if h.Service != "insightx-fulfillment" {
		t.Errorf("expected service 'insightx-fulfillment', got %q", h.Service)
	}

	// Test health failure on error response
	errorServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte(`{"ok":false,"error":"maintenance"}`))
	}))
	defer errorServer.Close()

	errClient := NewClient(errorServer.URL)
	_, err = errClient.GetHealth()
	if err == nil {
		t.Error("expected error on non-200 health response, got nil")
	}
}

func TestFulfillmentGetBalance(t *testing.T) {
	server, _ := setupMockFulfillmentServer(t)
	defer server.Close()

	c := NewClient(server.URL)
	bal, err := c.GetBalance(testAPIKey)
	if err != nil {
		t.Fatalf("GetBalance failed: %v", err)
	}
	if bal.BalanceUSDT <= 0 {
		t.Errorf("expected positive balance, got %f", bal.BalanceUSDT)
	}
	if bal.RateLimitPerMin != 60 {
		t.Errorf("expected rate limit 60, got %d", bal.RateLimitPerMin)
	}
	t.Logf("Live provider balance: %f USDT", bal.BalanceUSDT)

	// Empty API key validation
	_, err = c.GetBalance("")
	if err != ErrInvalidAPIKey {
		t.Errorf("expected ErrInvalidAPIKey, got: %v", err)
	}

	// Unauthorized API key
	_, err = c.GetBalance("invalid-key")
	if err != ErrInvalidAPIKey {
		t.Errorf("expected ErrInvalidAPIKey on 401, got: %v", err)
	}
}

func TestFulfillmentGetProducts(t *testing.T) {
	server, callCount := setupMockFulfillmentServer(t)
	defer server.Close()

	c := NewClient(server.URL)

	// Empty API key validation
	_, err := c.GetProducts("", false)
	if err != ErrInvalidAPIKey {
		t.Errorf("expected ErrInvalidAPIKey on empty key, got: %v", err)
	}

	// Initial fetch (populates cache)
	prods, err := c.GetProducts(testAPIKey, false)
	if err != nil {
		t.Fatalf("GetProducts failed: %v", err)
	}
	if len(prods) == 0 {
		t.Fatalf("expected non-empty catalog, got 0")
	}
	if prods[0].ID != 14 {
		t.Errorf("expected product ID 14, got %d", prods[0].ID)
	}
	if atomic.LoadInt32(callCount) != 1 {
		t.Errorf("expected 1 server call, got %d", atomic.LoadInt32(callCount))
	}

	// Second fetch without forceRefresh (must hit cache)
	cachedProds, err := c.GetProducts(testAPIKey, false)
	if err != nil {
		t.Fatalf("GetProducts cached fetch failed: %v", err)
	}
	if len(cachedProds) != len(prods) {
		t.Errorf("expected cached products match, got %d", len(cachedProds))
	}
	if atomic.LoadInt32(callCount) != 1 {
		t.Errorf("expected call count to remain 1 (cache hit), got %d", atomic.LoadInt32(callCount))
	}

	// Third fetch with forceRefresh (must bypass cache)
	refreshedProds, err := c.GetProducts(testAPIKey, true)
	if err != nil {
		t.Fatalf("GetProducts forceRefresh failed: %v", err)
	}
	if len(refreshedProds) == 0 {
		t.Fatalf("expected non-empty catalog after refresh, got 0")
	}
	if atomic.LoadInt32(callCount) != 2 {
		t.Errorf("expected call count to be 2 after forceRefresh, got %d", atomic.LoadInt32(callCount))
	}

	t.Logf("Retrieved %d products from fulfillment provider", len(prods))
}

func TestFulfillmentCreateOrder(t *testing.T) {
	server, _ := setupMockFulfillmentServer(t)
	defer server.Close()

	c := NewClient(server.URL)

	// Empty API key
	_, err := c.CreateOrder("", "idempotency-1", 14, 1)
	if err != ErrInvalidAPIKey {
		t.Errorf("expected ErrInvalidAPIKey, got: %v", err)
	}

	// Success order
	res, err := c.CreateOrder(testAPIKey, "idempotency-1", 14, 1)
	if err != nil {
		t.Fatalf("CreateOrder failed: %v", err)
	}
	if res.Status != "completed" {
		t.Errorf("expected status 'completed', got %q", res.Status)
	}
	if len(res.Codes) != 1 || res.Codes[0] != "https://gemini.google.com/redeem?code=TEST_PROMO_CODE" {
		t.Errorf("unexpected delivery codes: %+v", res.Codes)
	}

	// Insufficient funds (402)
	_, err = c.CreateOrder(testAPIKey, "idempotency-2", 8888, 1)
	if err != ErrInsufficientFunds {
		t.Errorf("expected ErrInsufficientFunds, got: %v", err)
	}

	// Product not found (404)
	_, err = c.CreateOrder(testAPIKey, "idempotency-3", 9999, 1)
	if err != ErrProductNotFound {
		t.Errorf("expected ErrProductNotFound, got: %v", err)
	}
}

func TestFulfillmentLiveIntegration(t *testing.T) {
	if os.Getenv("INSIGHTX_LIVE_TEST") != "true" {
		t.Skip("skipping live network test; set INSIGHTX_LIVE_TEST=true to run")
	}

	c := NewClient("")
	h, err := c.GetHealth()
	if err != nil {
		t.Logf("Live fulfillment provider health check failed (service may be offline): %v", err)
		return
	}
	t.Logf("Live health: ok=%v, service=%s", h.OK, h.Service)
}
