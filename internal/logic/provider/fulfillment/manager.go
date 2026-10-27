package fulfillment

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

type Manager struct {
	httpClient *http.Client
	adapters   map[string]FulfillmentAdapter

	mu           sync.RWMutex
	catalogCache map[string]CachedCatalog
}

func NewManager() *Manager {
	m := &Manager{
		httpClient: &http.Client{
			Timeout: 20 * time.Second,
		},
		adapters:     make(map[string]FulfillmentAdapter),
		catalogCache: make(map[string]CachedCatalog),
	}

	// Register the 5 primary bot adapters
	m.RegisterAdapter(&InsightXAdapter{})
	m.RegisterAdapter(&StackVaultAdapter{})
	m.RegisterAdapter(&LootPagluAdapter{})
	m.RegisterAdapter(&JeroAccountsAdapter{})
	m.RegisterAdapter(&SupremeLeaderAdapter{})

	return m
}

func (m *Manager) RegisterAdapter(a FulfillmentAdapter) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.adapters[a.GetProviderKey()] = a
}

func (m *Manager) GetAdapter(providerKey string) (FulfillmentAdapter, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if providerKey == "" || providerKey == "third_party_api" {
		providerKey = ProviderInsightX
	}
	adapter, exists := m.adapters[providerKey]
	if !exists {
		return nil, fmt.Errorf("تامین‌کننده ربات با شناسه %s یافت نشد", providerKey)
	}
	return adapter, nil
}

func (m *Manager) ListAdapters() []FulfillmentAdapter {
	m.mu.RLock()
	defer m.mu.RUnlock()
	list := make([]FulfillmentAdapter, 0, len(m.adapters))
	// Return in consistent order
	keys := []string{ProviderInsightX, ProviderStackVault, ProviderLootPaglu, ProviderJeroAccounts, ProviderSupremeLeader}
	for _, k := range keys {
		if a, ok := m.adapters[k]; ok {
			list = append(list, a)
		}
	}
	return list
}

func (m *Manager) GetBalance(providerKey string, cfg ProviderConfig) (*ProviderBalanceDTO, error) {
	adapter, err := m.GetAdapter(providerKey)
	if err != nil {
		return nil, err
	}
	return adapter.GetBalance(m.httpClient, cfg)
}

func (m *Manager) GetCatalog(providerKey string, cfg ProviderConfig, forceRefresh bool) ([]ProviderProductDTO, error) {
	adapter, err := m.GetAdapter(providerKey)
	if err != nil {
		return nil, err
	}

	cacheKey := providerKey + "_" + cfg.APIKey
	if !forceRefresh {
		m.mu.RLock()
		cached, exists := m.catalogCache[cacheKey]
		if exists && time.Since(cached.UpdatedAt) < 3*time.Minute {
			items := cached.Items
			m.mu.RUnlock()
			return items, nil
		}
		m.mu.RUnlock()
	}

	items, err := adapter.GetCatalog(m.httpClient, cfg)
	if err != nil {
		return nil, err
	}

	m.mu.Lock()
	m.catalogCache[cacheKey] = CachedCatalog{
		Items:     items,
		UpdatedAt: time.Now(),
	}
	m.mu.Unlock()

	return items, nil
}

func (m *Manager) CreateOrder(providerKey string, cfg ProviderConfig, req ProviderOrderRequest) (*ProviderOrderResultDTO, error) {
	adapter, err := m.GetAdapter(providerKey)
	if err != nil {
		return nil, err
	}
	return adapter.CreateOrder(m.httpClient, cfg, req)
}

// Backward-compatibility methods for legacy Client interface
func (m *Manager) GetHealth() (*HealthResponse, error) {
	return &HealthResponse{
		OK:      true,
		Service: "fulfillment-manager",
		Version: "3.0.0",
	}, nil
}

func (m *Manager) GetLegacyBalance(apiKey string) (*BalanceResponse, error) {
	cfg := ProviderConfig{
		Key:    ProviderInsightX,
		APIKey: apiKey,
	}
	dto, err := m.GetBalance(ProviderInsightX, cfg)
	if err != nil {
		return nil, err
	}
	return &BalanceResponse{
		BalanceUSDT:     dto.BalanceUSD,
		Status:          dto.Status,
		RateLimitPerMin: dto.RateLimitPerMin,
	}, nil
}
