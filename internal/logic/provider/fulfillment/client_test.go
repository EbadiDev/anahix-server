package fulfillment

import (
	"testing"
)

const testAPIKey = "isk_live_jexXb9akFkevDfAGpybP9AM4eI0fNqhE"

func TestFulfillmentGetHealth(t *testing.T) {
	c := NewClient("")
	h, err := c.GetHealth()
	if err != nil {
		t.Fatalf("GetHealth failed: %v", err)
	}
	if !h.OK {
		t.Errorf("expected health ok=true, got %v", h.OK)
	}
}

func TestFulfillmentGetBalance(t *testing.T) {
	c := NewClient("")
	bal, err := c.GetBalance(testAPIKey)
	if err != nil {
		t.Fatalf("GetBalance failed: %v", err)
	}
	if bal.BalanceUSDT <= 0 {
		t.Errorf("expected positive balance, got %f", bal.BalanceUSDT)
	}
	t.Logf("Live provider balance: %f USDT", bal.BalanceUSDT)
}

func TestFulfillmentGetProducts(t *testing.T) {
	c := NewClient("")
	prods, err := c.GetProducts(testAPIKey, false)
	if err != nil {
		t.Fatalf("GetProducts failed: %v", err)
	}
	if len(prods) == 0 {
		t.Fatalf("expected non-empty catalog, got 0")
	}
	t.Logf("Retrieved %d products from fulfillment provider", len(prods))
}
