package public

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

func setupTestDB(t *testing.T) *svc.ServiceContext {
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

	_ = db.Create(&model.SystemSetting{
		Category: "sms",
		Key:      "smsir_template_id",
		Value:    "100000",
	})

	cfg := &config.Config{
		JWT: config.JWTConfig{
			AccessSecret:      "test-secret-key-1234567890",
			AccessExpireHours: 720,
		},
	}

	svcCtx := &svc.ServiceContext{
		DB:          db,
		Config:      cfg,
		SMSIRClient: smsir.NewClient(),
	}

	return svcCtx
}

func TestCustomerAuthFlow(t *testing.T) {
	ctx := context.Background()
	svcCtx := setupTestDB(t)
	authLogic := NewCustomerAuthLogic(ctx, svcCtx)

	mobile := "09123456789"

	// 1. Send OTP
	sendResp, err := authLogic.SendOTP(&types.SendOTPRequest{
		Mobile: mobile,
	})
	if err != nil {
		t.Fatalf("SendOTP failed: %v", err)
	}
	if !sendResp.Sent {
		t.Fatalf("expected sendResp.Sent == true")
	}

	// 2. Cooldown check: immediate second send should fail
	_, err = authLogic.SendOTP(&types.SendOTPRequest{
		Mobile: mobile,
	})
	if err == nil {
		t.Fatalf("expected cooldown error on rapid resend, got nil")
	}

	// 3. Find OTP code generated in DB
	var otp model.OTPVerification
	err = svcCtx.DB.Where("phone = ? AND used = ?", mobile, false).First(&otp).Error
	if err != nil {
		t.Fatalf("failed to find otp record: %v", err)
	}
	if len(otp.Code) != 5 {
		t.Fatalf("expected 5-digit code, got %s", otp.Code)
	}

	// 4. Verify OTP with wrong code
	_, err = authLogic.VerifyOTP(&types.VerifyOTPRequest{
		Mobile: mobile,
		Code:   "00000",
	})
	if err == nil {
		t.Fatalf("expected error on invalid code, got nil")
	}

	// 5. Verify OTP with correct code
	verifyResp, err := authLogic.VerifyOTP(&types.VerifyOTPRequest{
		Mobile: mobile,
		Code:   otp.Code,
	})
	if err != nil {
		t.Fatalf("VerifyOTP failed with correct code: %v", err)
	}
	if verifyResp.Token == "" {
		t.Fatalf("expected non-empty JWT token")
	}
	if verifyResp.Customer.Phone != mobile {
		t.Fatalf("expected customer phone %s, got %s", mobile, verifyResp.Customer.Phone)
	}

	// 6. Check GetMe
	meResp, err := authLogic.GetMe(verifyResp.Customer.ID)
	if err != nil {
		t.Fatalf("GetMe failed: %v", err)
	}
	if meResp.Phone != mobile {
		t.Fatalf("expected %s, got %s", mobile, meResp.Phone)
	}

	// 7. Used OTP cannot be reused
	_, err = authLogic.VerifyOTP(&types.VerifyOTPRequest{
		Mobile: mobile,
		Code:   otp.Code,
	})
	if err == nil {
		t.Fatalf("expected error reusing already consumed OTP code")
	}
}

func TestMobileNormalization(t *testing.T) {
	ctx := context.Background()
	svcCtx := setupTestDB(t)
	authLogic := NewCustomerAuthLogic(ctx, svcCtx)

	// Send using +98 format
	sendResp, err := authLogic.SendOTP(&types.SendOTPRequest{
		Mobile: "+989351112233",
	})
	if err != nil {
		t.Fatalf("SendOTP failed for +98 format: %v", err)
	}
	if !sendResp.Sent {
		t.Fatalf("expected sendResp.Sent == true")
	}

	// Must be stored as 09351112233
	var otp model.OTPVerification
	err = svcCtx.DB.Where("phone = ?", "09351112233").First(&otp).Error
	if err != nil {
		t.Fatalf("expected normalized phone 09351112233 in db: %v", err)
	}
}
