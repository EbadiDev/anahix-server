package admin

import (
	"context"
	"testing"

	"github.com/EbadiDev/anahix-server/internal/config"
	"github.com/EbadiDev/anahix-server/internal/logic/provider/smsir"
	"github.com/EbadiDev/anahix-server/internal/model"
	"github.com/EbadiDev/anahix-server/internal/svc"
	"github.com/EbadiDev/anahix-server/internal/types"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupTestAdminDB(t *testing.T) *svc.ServiceContext {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite memory db: %v", err)
	}

	err = db.AutoMigrate(
		&model.Customer{},
		&model.OTPVerification{},
		&model.SystemSetting{},
		&model.AdminActionLog{},
	)
	if err != nil {
		t.Fatalf("failed to auto migrate tables: %v", err)
	}

	cfg := &config.Config{
		JWT: config.JWTConfig{
			AccessSecret:      "test-secret-key-1234567890",
			AccessExpireHours: 720,
		},
	}

	return &svc.ServiceContext{
		DB:          db,
		Config:      cfg,
		SMSIRClient: smsir.NewClient(),
	}
}

func TestSMSSettingsLogic(t *testing.T) {
	ctx := context.Background()
	svcCtx := setupTestAdminDB(t)
	logic := NewSMSSettingsLogic(ctx, svcCtx)

	// 1. Initial settings check
	resp, err := logic.GetSMSSettings()
	if err != nil {
		t.Fatalf("GetSMSSettings failed: %v", err)
	}
	if resp.Provider != "smsir" {
		t.Fatalf("expected provider 'smsir', got '%s'", resp.Provider)
	}

	// 2. Update settings
	apiKey := ""
	lineNum := "30004505000017"
	templateID := "654321"
	template2FA := "200001"
	templateOrderComplete := "200002"
	templateRenewal := "200003"
	isSandbox := true

	err = logic.UpdateSMSSettings(&types.UpdateSMSSettingsRequest{
		APIKey:                &apiKey,
		LineNumber:            &lineNum,
		TemplateID:            &templateID,
		Template2FA:           &template2FA,
		TemplateOrderComplete: &templateOrderComplete,
		TemplateRenewal:       &templateRenewal,
		IsSandbox:             &isSandbox,
	}, "admin")
	if err != nil {
		t.Fatalf("UpdateSMSSettings failed: %v", err)
	}

	// 3. Verify settings updated & templates populated
	respUpdated, err := logic.GetSMSSettings()
	if err != nil {
		t.Fatalf("GetSMSSettings after update failed: %v", err)
	}
	if respUpdated.TemplateID != templateID {
		t.Fatalf("expected templateID %s, got %s", templateID, respUpdated.TemplateID)
	}
	if respUpdated.Template2FA != template2FA {
		t.Fatalf("expected template2FA %s, got %s", template2FA, respUpdated.Template2FA)
	}
	if respUpdated.TemplateOrderComplete != templateOrderComplete {
		t.Fatalf("expected templateOrderComplete %s, got %s", templateOrderComplete, respUpdated.TemplateOrderComplete)
	}
	if respUpdated.TemplateRenewal != templateRenewal {
		t.Fatalf("expected templateRenewal %s, got %s", templateRenewal, respUpdated.TemplateRenewal)
	}
	if len(respUpdated.Templates) != 4 {
		t.Fatalf("expected 4 templates, got %d", len(respUpdated.Templates))
	}
	if respUpdated.LineNumber != lineNum {
		t.Fatalf("expected lineNum %s, got %s", lineNum, respUpdated.LineNumber)
	}

	// 4. Send test SMS for different templates
	err = logic.SendTestSMS(&types.SendTestSMSRequest{
		Mobile:       "09121112233",
		TemplateType: "otp",
	}, "admin")
	if err != nil {
		t.Fatalf("SendTestSMS (otp) failed: %v", err)
	}

	err = logic.SendTestSMS(&types.SendTestSMSRequest{
		Mobile:       "09121112233",
		TemplateType: "two_factor",
	}, "admin")
	if err != nil {
		t.Fatalf("SendTestSMS (two_factor) failed: %v", err)
	}

	err = logic.SendTestSMS(&types.SendTestSMSRequest{
		Mobile:       "09121112233",
		TemplateType: "order_complete",
	}, "admin")
	if err != nil {
		t.Fatalf("SendTestSMS (order_complete) failed: %v", err)
	}

	err = logic.SendTestSMS(&types.SendTestSMSRequest{
		Mobile:       "09121112233",
		TemplateType: "renewal",
	}, "admin")
	if err != nil {
		t.Fatalf("SendTestSMS (renewal) failed: %v", err)
	}

	// 5. Verify action log
	var logCount int64
	svcCtx.DB.Model(&model.AdminActionLog{}).Count(&logCount)
	if logCount < 2 {
		t.Fatalf("expected at least 2 admin action logs (update and test), got %d", logCount)
	}
}
