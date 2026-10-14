package exchangerate

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type SwapWalletResponse struct {
	Status string            `json:"status"`
	Result map[string]string `json:"result"`
}

type Client struct {
	httpClient *http.Client
	mu         sync.RWMutex
	cachedRate int64
	cacheTime  time.Time
}

var (
	defaultClient *Client
	once          sync.Once
)

func GetDefaultClient() *Client {
	once.Do(func() {
		defaultClient = NewClient()
	})
	return defaultClient
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 6 * time.Second,
		},
	}
}

// GetUSDTPrice fetches real-time USDT/IRT price from SwapWallet API (cached for 3 minutes)
func (c *Client) GetUSDTPrice() (int64, error) {
	c.mu.RLock()
	if c.cachedRate > 0 && time.Since(c.cacheTime) < 3*time.Minute {
		rate := c.cachedRate
		c.mu.RUnlock()
		return rate, nil
	}
	c.mu.RUnlock()

	resp, err := c.httpClient.Get("https://swapwallet.app/api/v1/market/prices")
	if err != nil {
		c.mu.RLock()
		if c.cachedRate > 0 {
			defer c.mu.RUnlock()
			return c.cachedRate, nil
		}
		c.mu.RUnlock()
		return 0, fmt.Errorf("خطا در اتصال به SwapWallet: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.mu.RLock()
		if c.cachedRate > 0 {
			defer c.mu.RUnlock()
			return c.cachedRate, nil
		}
		c.mu.RUnlock()
		return 0, fmt.Errorf("کد خطای HTTP صرافی SwapWallet: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	var data SwapWalletResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return 0, fmt.Errorf("خطا در پارس پاسخ SwapWallet: %w", err)
	}

	valStr, ok := data.Result["USDT/IRT"]
	if !ok || strings.TrimSpace(valStr) == "" {
		valStr = data.Result["USD/IRT"]
	}
	if strings.TrimSpace(valStr) == "" {
		return 0, errors.New("نرخ USDT/IRT در پاسخ SwapWallet یافت نشد")
	}

	rateFloat, err := strconv.ParseFloat(valStr, 64)
	if err != nil || rateFloat <= 0 {
		return 0, fmt.Errorf("نرخ نامعتبر USDT از SwapWallet: %s", valStr)
	}

	rate := int64(rateFloat)

	c.mu.Lock()
	c.cachedRate = rate
	c.cacheTime = time.Now()
	c.mu.Unlock()

	return rate, nil
}

// ResolveRate evaluates priority between manual rate and SwapWallet live USDT converter
func (c *Client) ResolveRate(priority string, manualRate int64) (effectiveRate int64, activeSource string, liveUSDT int64) {
	liveRate, err := c.GetUSDTPrice()
	if err == nil && liveRate > 0 {
		liveUSDT = liveRate
	}

	priority = strings.ToLower(strings.TrimSpace(priority))
	if priority == "swapwallet" || priority == "usdt" {
		if liveUSDT > 0 {
			return liveUSDT, "swapwallet", liveUSDT
		}
		// Fallback to manual if SwapWallet is temporarily unreachable
		if manualRate > 0 {
			return manualRate, "manual (پشتیبان به دلیل خطای اتصال)", liveUSDT
		}
		return 95000, "default (پیش‌فرض)", liveUSDT
	}

	// Priority 1: Manual rate
	if manualRate > 0 {
		return manualRate, "manual", liveUSDT
	}
	if liveUSDT > 0 {
		return liveUSDT, "swapwallet (پشتیبان نرخ دستی)", liveUSDT
	}
	return 95000, "default (پیش‌فرض)", liveUSDT
}
