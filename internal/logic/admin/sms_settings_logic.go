package admin

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/EbadiDev/anahix-server/internal/logic/provider/smsir"
	"github.com/EbadiDev/anahix-server/internal/model"
	"github.com/EbadiDev/anahix-server/internal/svc"
	"github.com/EbadiDev/anahix-server/internal/types"
	"github.com/EbadiDev/anahix-server/pkg/xerr"
	"gorm.io/gorm/clause"
)

type SMSSettingsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSMSSettingsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SMSSettingsLogic {
	return &SMSSettingsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SMSSettingsLogic) GetSMSSettings() (*types.SMSSettingsResponse, error) {
	var settings []model.SystemSetting
	_ = l.svcCtx.DB.WithContext(l.ctx).Where("category = ?", "sms").Find(&settings).Error

	configMap := make(map[string]string)
	for _, s := range settings {
		configMap[s.Key] = s.Value
	}

	provider := configMap["sms_provider"]
	if provider == "" {
		provider = "smsir"
	}
	apiKey := strings.TrimSpace(configMap["smsir_api_key"])
	lineNumber := strings.Trim(configMap["smsir_line_number"], "'\" \t\r\n")
	templateID := strings.Trim(configMap["smsir_template_id"], "'\" \t\r\n")
	isSandbox := configMap["smsir_is_sandbox"] == "true"

	maskedKey := ""
	if len(apiKey) > 8 {
		maskedKey = apiKey[:4] + "••••••••" + apiKey[len(apiKey)-4:]
	} else if apiKey != "" {
		maskedKey = "••••••••"
	}

	credit, _ := l.svcCtx.SMSIRClient.GetCredit(l.ctx, apiKey)
	linesInt, _ := l.svcCtx.SMSIRClient.GetLines(l.ctx, apiKey)
	linesStr := make([]string, 0, len(linesInt))
	for _, ln := range linesInt {
		linesStr = append(linesStr, strconv.FormatInt(ln, 10))
	}

	return &types.SMSSettingsResponse{
		Provider:   provider,
		APIKeyMask: maskedKey,
		LineNumber: lineNumber,
		TemplateID: templateID,
		IsSandbox:  isSandbox,
		LiveCredit: credit,
		Lines:      linesStr,
	}, nil
}

func (l *SMSSettingsLogic) UpdateSMSSettings(req *types.UpdateSMSSettingsRequest, adminUsername string) error {
	upsertSetting := func(key, value, desc, valType string) error {
		setting := model.SystemSetting{
			Category: "sms",
			Key:      key,
			Value:    value,
			Desc:     desc,
			Type:     valType,
		}
		return l.svcCtx.DB.WithContext(l.ctx).Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "key"}},
			DoUpdates: clause.AssignmentColumns([]string{"value", "updated_at"}),
		}).Create(&setting).Error
	}

	_ = upsertSetting("sms_provider", "smsir", "سامانه پیامک فعال", "string")

	if req.APIKey != nil && *req.APIKey != "" {
		_ = upsertSetting("smsir_api_key", strings.TrimSpace(*req.APIKey), "کلید API سامانه پیامک", "string")
	}
	if req.LineNumber != nil {
		_ = upsertSetting("smsir_line_number", strings.TrimSpace(*req.LineNumber), "شماره خط اختصاصی ارسال پیامک", "string")
	}
	if req.TemplateID != nil && *req.TemplateID != "" {
		_ = upsertSetting("smsir_template_id", strings.TrimSpace(*req.TemplateID), "شناسه قالب ارسال سریع پیامک اعتبارسنجی (Verify Template)", "string")
	}
	if req.IsSandbox != nil {
		val := "false"
		if *req.IsSandbox {
			val = "true"
		}
		_ = upsertSetting("smsir_is_sandbox", val, "حالت محیط آزمایشی (سندباکس)", "bool")
	}

	// Record admin audit log
	adminAction := model.AdminActionLog{
		AdminUsername: adminUsername,
		Action:        "update_sms_settings",
		Resource:      "system_settings/sms",
		Details:       "تنظیمات وب‌سرویس سامانه پیامک به‌روزرسانی شد",
	}
	_ = l.svcCtx.DB.WithContext(l.ctx).Create(&adminAction)

	return nil
}

func (l *SMSSettingsLogic) SendTestSMS(req *types.SendTestSMSRequest, adminUsername string) error {
	phone := smsir.NormalizeMobile(req.Mobile)
	if phone == "" {
		return xerr.NewErrMsg("شماره موبایل وارد شده نامعتبر است")
	}

	var settings []model.SystemSetting
	l.svcCtx.DB.WithContext(l.ctx).Where("category = ?", "sms").Find(&settings)
	configMap := make(map[string]string)
	for _, s := range settings {
		configMap[s.Key] = s.Value
	}

	apiKey := configMap["smsir_api_key"]
	templateID := 0
	cleanedTid := strings.Trim(configMap["smsir_template_id"], "'\" \t\r\n")
	if tid, err := strconv.Atoi(cleanedTid); err == nil && tid > 0 {
		templateID = tid
	}
	if templateID <= 0 {
		return fmt.Errorf("شناسه قالب پیامک در سیستم پیکربندی نشده است")
	}

	// Dispatch test verify code
	testCode := "12345"
	_, err := l.svcCtx.SMSIRClient.SendVerify(l.ctx, apiKey, phone, templateID, testCode)
	if err != nil {
		return fmt.Errorf("خطا در ارسال پیامک آزمایشی: %w", err)
	}

	// Record admin audit log
	adminAction := model.AdminActionLog{
		AdminUsername: adminUsername,
		Action:        "send_test_sms",
		Resource:      phone,
		Details:       fmt.Sprintf("پیامک آزمایشی با موفقیت به شماره %s ارسال شد", phone),
	}
	_ = l.svcCtx.DB.WithContext(l.ctx).Create(&adminAction)

	return nil
}
