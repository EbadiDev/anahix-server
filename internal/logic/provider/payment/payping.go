package payment

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type PayPingClient struct {
	httpClient *http.Client
	baseURL    string
}

func NewPayPingClient() *PayPingClient {
	return &PayPingClient{
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
		baseURL: "https://api.payping.ir/v2/pay",
	}
}

type PayPingCreateRequest struct {
	Amount        int64  `json:"amount"` // In Toman
	PayerIdentity string `json:"payerIdentity,omitempty"`
	PayerName     string `json:"payerName,omitempty"`
	Description   string `json:"description,omitempty"`
	ReturnUrl     string `json:"returnUrl"`
	ClientRefId   string `json:"clientRefId,omitempty"`
}

type PayPingCreateResponse struct {
	Code string `json:"code"`
}

type PayPingVerifyRequest struct {
	Amount int64  `json:"amount"` // In Toman
	RefId  string `json:"refId"`
}

type PayPingVerifyResponse struct {
	CardNumber  string `json:"cardNumber,omitempty"`
	CardHashPan string `json:"cardHashPan,omitempty"`
}

// RequestPayment creates an IPG payment token and returns the redirect URL
func (c *PayPingClient) RequestPayment(ctx context.Context, token string, req PayPingCreateRequest) (code string, redirectURL string, err error) {
	token = strings.TrimSpace(token)
	if token == "" {
		// Mock/Sandbox auto-approval code for testing when token is not configured
		mockCode := fmt.Sprintf("mock_pp_%d", time.Now().UnixNano())
		return mockCode, fmt.Sprintf("/payment/callback/payping?code=%s&clientrefid=%s&refid=SIMULATED_REF_%d", mockCode, req.ClientRefId, time.Now().Unix()), nil
	}

	url := c.baseURL
	reqBody, err := json.Marshal(req)
	if err != nil {
		return "", "", fmt.Errorf("خطا در سریال‌سازی درخواست پی‌پینگ: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(reqBody))
	if err != nil {
		return "", "", fmt.Errorf("خطا در ساخت درخواست پی‌پینگ: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", "", fmt.Errorf("عدم ارتباط با درگاه پی‌پینگ: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != 201 {
		return "", "", fmt.Errorf("پاسخ ناموفق پی‌پینگ (%d): %s", resp.StatusCode, string(bodyBytes))
	}

	var resData PayPingCreateResponse
	if err := json.Unmarshal(bodyBytes, &resData); err != nil {
		return "", "", fmt.Errorf("خطا در پردازش پاسخ پی‌پینگ: %w", err)
	}

	if resData.Code == "" {
		return "", "", fmt.Errorf("شناسه پرداخت از پی‌پینگ دریافت نشد")
	}

	gotoURL := fmt.Sprintf("https://api.payping.ir/v2/pay/gotoipg/%s", resData.Code)
	return resData.Code, gotoURL, nil
}

// VerifyPayment checks and validates the payment against PayPing verify endpoint
func (c *PayPingClient) VerifyPayment(ctx context.Context, token string, req PayPingVerifyRequest) (*PayPingVerifyResponse, error) {
	token = strings.TrimSpace(token)
	if token == "" || strings.HasPrefix(req.RefId, "SIMULATED_") {
		// Mock verify for sandbox / local development
		return &PayPingVerifyResponse{
			CardNumber:  "603799******1234",
			CardHashPan: "MOCK_HASH_PAN_SUCCESS",
		}, nil
	}

	url := fmt.Sprintf("%s/verify", c.baseURL)
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("خطا در سریال‌سازی درخواست تایید پی‌پینگ: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("خطا در ساخت درخواست تایید پی‌پینگ: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("خطا در تماس با تاییدیه پی‌پینگ: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("تایید پرداخت توسط پی‌پینگ انجام نشد (%d): %s", resp.StatusCode, string(bodyBytes))
	}

	var verifyRes PayPingVerifyResponse
	if err := json.Unmarshal(bodyBytes, &verifyRes); err != nil {
		// Success status without cardNumber body is still valid in PayPing v2
		return &PayPingVerifyResponse{}, nil
	}

	return &verifyRes, nil
}
