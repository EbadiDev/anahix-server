package smsir

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

const (
	BaseURL = "https://api.sms.ir/v1"
	DefaultTimeout = 10 * time.Second
)

var (
	ErrEmptyAPIKey     = errors.New("کلید سامانه پیامک تنظیم نشده است")
	ErrInvalidMobile   = errors.New("شماره موبایل وارد شده نامعتبر است")
	ErrTemplateMissing = errors.New("شناسه قالب پیامک (Template ID) مشخص نشده است")
)

type Client struct {
	httpClient *http.Client
	baseURL    string
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
		baseURL: BaseURL,
	}
}

// SMS.ir General API Response Wrapper
type BaseResponse struct {
	Status  int             `json:"status"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

type VerifyParameter struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type VerifyRequest struct {
	Mobile     string            `json:"mobile"`
	TemplateID int               `json:"templateId"`
	Parameters []VerifyParameter `json:"parameters"`
}

type VerifyResultData struct {
	MessageID int64   `json:"messageId"`
	Cost      float64 `json:"cost"`
}

type BulkRequest struct {
	LineNumber   int64    `json:"lineNumber"`
	MessageText  string   `json:"messageText"`
	Mobiles      []string `json:"mobiles"`
	SendDateTime *int64   `json:"sendDateTime,omitempty"`
}

type BulkResultData struct {
	PackID     string  `json:"packId"`
	MessageIDs []int64 `json:"messageIds"`
	Cost       float64 `json:"cost"`
}

type MessageStatusData struct {
	MessageID        int64    `json:"messageId"`
	Mobile           int64    `json:"mobile"`
	MessageText      string   `json:"messageText"`
	SendDateTime     int64    `json:"sendDateTime"`
	LineNumber       int64    `json:"lineNumber"`
	Cost             float64  `json:"cost"`
	DeliveryState    *int     `json:"deliveryState"`
	DeliveryDateTime *int64   `json:"deliveryDateTime"`
}

// NormalizeMobile ensures mobile starts with 09 or standard 11-digit Iranian format
func NormalizeMobile(mobile string) string {
	m := strings.TrimSpace(mobile)
	m = strings.ReplaceAll(m, " ", "")
	m = strings.ReplaceAll(m, "-", "")
	if strings.HasPrefix(m, "+98") {
		m = "0" + strings.TrimPrefix(m, "+98")
	} else if strings.HasPrefix(m, "98") && len(m) == 12 {
		m = "0" + strings.TrimPrefix(m, "98")
	} else if len(m) == 10 && strings.HasPrefix(m, "9") {
		m = "0" + m
	}
	return m
}

// SendVerify dispatches high-priority OTP code using SMS.ir Verify endpoint
func (c *Client) SendVerify(ctx context.Context, apiKey string, mobile string, templateID int, code string) (*VerifyResultData, error) {
	normMobile := NormalizeMobile(mobile)
	if normMobile == "" {
		return nil, ErrInvalidMobile
	}

	if templateID <= 0 {
		return nil, ErrTemplateMissing
	}

	// Development / Mock fallback if no API key is provided
	if strings.TrimSpace(apiKey) == "" {
		log.Printf("📱 [SMS.ir SIMULATOR] Sent OTP Code [%s] to [%s] (Template: %d)", code, normMobile, templateID)
		return &VerifyResultData{
			MessageID: time.Now().Unix(),
			Cost:      1.0,
		}, nil
	}

	payload := VerifyRequest{
		Mobile:     normMobile,
		TemplateID: templateID,
		Parameters: []VerifyParameter{
			{Name: "Code", Value: code},
		},
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("خطا در سریال‌سازی داده‌های پیامک: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/send/verify", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("خطا در ساخت درخواست وب‌سرویس: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-API-KEY", apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("خطا در برقراری ارتباط با سامانه پیامک: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("خطا در دریافت پاسخ از سامانه پیامک: %w", err)
	}

	var baseResp BaseResponse
	if err := json.Unmarshal(respBody, &baseResp); err != nil {
		return nil, fmt.Errorf("پاسخ نامعتبر از سامانه پیامک (کد وضعیت HTTP: %d): %s", resp.StatusCode, string(respBody))
	}

	if baseResp.Status != 1 {
		return nil, fmt.Errorf("خطای سامانه پیامک (کد %d): %s", baseResp.Status, baseResp.Message)
	}

	var resultData VerifyResultData
	if err := json.Unmarshal(baseResp.Data, &resultData); err != nil {
		return nil, fmt.Errorf("خطا در تحلیل دیتای پیامک تایید: %w", err)
	}

	log.Printf("✅ [SMS.ir] Verify OTP sent successfully to %s (MsgID: %d, Cost: %.1f)", normMobile, resultData.MessageID, resultData.Cost)
	return &resultData, nil
}

// SendBulk dispatches text SMS to multiple recipients
func (c *Client) SendBulk(ctx context.Context, apiKey string, lineNumber int64, message string, mobiles []string) (*BulkResultData, error) {
	if strings.TrimSpace(apiKey) == "" {
		log.Printf("📱 [SMS.ir SIMULATOR] Sent Bulk Text to %d numbers: %s", len(mobiles), message)
		return &BulkResultData{
			PackID:     "simulated-pack-id",
			MessageIDs: []int64{time.Now().Unix()},
			Cost:       float64(len(mobiles)),
		}, nil
	}

	normalizedMobiles := make([]string, 0, len(mobiles))
	for _, m := range mobiles {
		if nm := NormalizeMobile(m); nm != "" {
			normalizedMobiles = append(normalizedMobiles, nm)
		}
	}

	payload := BulkRequest{
		LineNumber:  lineNumber,
		MessageText: message,
		Mobiles:     normalizedMobiles,
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/send/bulk", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-API-KEY", apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var baseResp BaseResponse
	if err := json.Unmarshal(respBody, &baseResp); err != nil {
		return nil, fmt.Errorf("پاسخ نامعتبر از سامانه پیامک: %s", string(respBody))
	}

	if baseResp.Status != 1 {
		return nil, fmt.Errorf("خطای سامانه پیامک (%d): %s", baseResp.Status, baseResp.Message)
	}

	var bulkData BulkResultData
	if err := json.Unmarshal(baseResp.Data, &bulkData); err != nil {
		return nil, err
	}

	return &bulkData, nil
}

// GetCredit checks remaining account balance
func (c *Client) GetCredit(ctx context.Context, apiKey string) (float64, error) {
	if strings.TrimSpace(apiKey) == "" {
		return 1000.0, nil // Mock balance
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/credit", nil)
	if err != nil {
		return 0, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-API-KEY", apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("خطا در استعلام موجودی سامانه پیامک: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var baseResp BaseResponse
	if err := json.Unmarshal(respBody, &baseResp); err != nil {
		return 0, fmt.Errorf("پاسخ نامعتبر از سامانه پیامک: %s", string(respBody))
	}

	if baseResp.Status != 1 {
		return 0, fmt.Errorf("خطای سامانه پیامک (%d): %s", baseResp.Status, baseResp.Message)
	}

	var credit float64
	if err := json.Unmarshal(baseResp.Data, &credit); err != nil {
		return 0, fmt.Errorf("خطا در تبدیل موجودی اعتباری: %w", err)
	}

	return credit, nil
}

// GetLines lists available sender lines
func (c *Client) GetLines(ctx context.Context, apiKey string) ([]int64, error) {
	if strings.TrimSpace(apiKey) == "" {
		return []int64{30004505000017, 10002155613464}, nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/line", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-API-KEY", apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var baseResp BaseResponse
	if err := json.Unmarshal(respBody, &baseResp); err != nil {
		return nil, err
	}

	if baseResp.Status != 1 {
		return nil, fmt.Errorf("خطای SMS.ir (%d): %s", baseResp.Status, baseResp.Message)
	}

	var lines []int64
	if err := json.Unmarshal(baseResp.Data, &lines); err != nil {
		return nil, err
	}

	return lines, nil
}
