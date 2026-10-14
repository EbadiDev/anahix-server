package smsir

import (
	"context"
	"testing"
)

func TestNormalizeMobile(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"09121234567", "09121234567"},
		{"+989121234567", "09121234567"},
		{"989121234567", "09121234567"},
		{"9121234567", "09121234567"},
		{" 0912-123-4567 ", "09121234567"},
	}

	for _, tt := range tests {
		got := NormalizeMobile(tt.input)
		if got != tt.expected {
			t.Errorf("NormalizeMobile(%q) = %q; want %q", tt.input, got, tt.expected)
		}
	}
}

func TestSimulatorSendVerify(t *testing.T) {
	c := NewClient()
	ctx := context.Background()

	// Empty apiKey should use simulator mode without failing
	res, err := c.SendVerify(ctx, "", "09121234567", 123456, "54321")
	if err != nil {
		t.Fatalf("expected simulator to succeed, got: %v", err)
	}
	if res.MessageID == 0 {
		t.Errorf("expected non-zero MessageID in simulator")
	}
}

func TestSimulatorGetCredit(t *testing.T) {
	c := NewClient()
	ctx := context.Background()

	cred, err := c.GetCredit(ctx, "")
	if err != nil {
		t.Fatalf("expected mock credit to succeed, got: %v", err)
	}
	if cred <= 0 {
		t.Errorf("expected positive mock credit, got %f", cred)
	}
}
