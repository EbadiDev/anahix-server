package exchangerate

import (
	"testing"
)

func TestSwapWalletGetUSDTPrice(t *testing.T) {
	client := NewClient()
	rate, err := client.GetUSDTPrice()
	if err != nil {
		t.Logf("SwapWallet network request warning (expected in isolated CI): %v", err)
		return
	}

	if rate <= 0 {
		t.Fatalf("expected positive USDT rate, got %d", rate)
	}
	t.Logf("Live USDT/IRT rate from SwapWallet: %d Toman", rate)
}

func TestResolveRateManualPriority(t *testing.T) {
	client := NewClient()
	effectiveRate, source, _ := client.ResolveRate("manual", 95000)
	if effectiveRate != 95000 {
		t.Fatalf("expected manual rate 95000, got %d", effectiveRate)
	}
	if source != "manual" {
		t.Fatalf("expected source 'manual', got %s", source)
	}
}

func TestResolveRateSwapWalletPriority(t *testing.T) {
	client := NewClient()
	effectiveRate, source, liveUSDT := client.ResolveRate("swapwallet", 95000)
	if liveUSDT > 0 {
		if effectiveRate != liveUSDT {
			t.Fatalf("expected live rate %d, got %d", liveUSDT, effectiveRate)
		}
		if source != "swapwallet" {
			t.Fatalf("expected source 'swapwallet', got %s", source)
		}
	} else {
		// Fallback scenario
		if effectiveRate != 95000 {
			t.Fatalf("expected fallback manual rate 95000, got %d", effectiveRate)
		}
	}
}
