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
	templateOTP := strings.Trim(configMap["smsir_template_otp"], "'\" \t\r\n")
	if templateOTP == "" {
		templateOTP = strings.Trim(configMap["smsir_template_id"], "'\" \t\r\n")
	}
	if templateOTP == "" {
		templateOTP = "100000"
	}

	template2FA := strings.Trim(configMap["smsir_template_2fa"], "'\" \t\r\n")
	if template2FA == "" {
		template2FA = "100001"
	}

	templateOrderComplete := strings.Trim(configMap["smsir_template_order_complete"], "'\" \t\r\n")
	if templateOrderComplete == "" {
		templateOrderComplete = "100002"
	}

	templateRenewal := strings.Trim(configMap["smsir_template_renewal"], "'\" \t\r\n")
	if templateRenewal == "" {
		templateRenewal = "100003"
	}
	isSandbox := configMap["smsir_is_sandbox"] == "true"
	ticketSMSMode := configMap["ticket_sms_mode"]
	if ticketSMSMode == "" {
		ticketSMSMode = "default_on"
	}

	templates := []types.SMSTemplateItem{
		{
			Key:         "otp",
			Title:       "الگوی خوش‌آمدگویی و ورود OTP",
			Description: "ارسال کد یکبار مصرف ورود و ثبت‌نام سریع مشتریان (عبور از بلک‌لیست تبلیغاتی)",
			TemplateID:  templateOTP,
			Variables:   []string{"Code"},
			SampleText:  "کد ورود به آناهیکس: #Code#\nاین کد محرمانه است.",
			IsEnabled:   templateOTP != "",
		},
		{
			Key:         "two_factor",
			Title:       "الگوی اعلام درخواست کد ۲FA به مشتری",
			Description: "اطلاع‌رسانی فوری به خریدار جهت ورود به پنل و ارسال کد تایید دومرحله‌ای سفارش",
			TemplateID:  template2FA,
			Variables:   []string{"Order", "Prompt"},
			SampleText:  "مشتری گرامی، جهت تکمیل سفارش #Order# نیاز به تایید ۲FA است: #Prompt#\nورود: anahix.ir",
			IsEnabled:   template2FA != "",
		},
		{
			Key:         "order_complete",
			Title:       "الگوی تحویل لایسنس و تکمیل سفارش",
			Description: "ارسال لایسنس، اطلاعات اکانت و تایید تحویل سفارش به شماره همراه مشتری",
			TemplateID:  templateOrderComplete,
			Variables:   []string{"Order", "License"},
			SampleText:  "سفارش #Order# تکمیل شد. لایسنس: #License#\nمشاهده جزییات: anahix.ir/profile?tab=orders",
			IsEnabled:   templateOrderComplete != "",
		},
		{
			Key:         "renewal",
			Title:       "الگوی یادآور تمدید اشتراک",
			Description: "اطلاع‌رسانی نزدیک شدن به پایان مهلت اشتراک و دعوت به تمدید آنلاین",
			TemplateID:  templateRenewal,
			Variables:   []string{"Service", "Days", "Link"},
			SampleText:  "کاربر گرامی، اشتراک #Service# شما #Days# روز دیگر منقضی می‌شود.\nتمدید سریع: #Link#",
			IsEnabled:   templateRenewal != "",
		},
	}

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
		Provider:              provider,
		APIKeyMask:            maskedKey,
		LineNumber:            lineNumber,
		TemplateID:            templateOTP,
		TemplateOTP:           templateOTP,
		Template2FA:           template2FA,
		TemplateOrderComplete: templateOrderComplete,
		TemplateRenewal:       templateRenewal,
		Templates:             templates,
		IsSandbox:             isSandbox,
		LiveCredit:            credit,
		Lines:                 linesStr,
		TicketSMSMode:         ticketSMSMode,
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
		tid := strings.TrimSpace(*req.TemplateID)
		_ = upsertSetting("smsir_template_id", tid, "شناسه قالب ارسال سریع پیامک اعتبارسنجی (Verify Template)", "string")
		_ = upsertSetting("smsir_template_otp", tid, "شناسه الگوی خوش‌آمدگویی و ورود OTP", "string")
	}
	if req.TemplateOTP != nil && *req.TemplateOTP != "" {
		tid := strings.TrimSpace(*req.TemplateOTP)
		_ = upsertSetting("smsir_template_otp", tid, "شناسه الگوی خوش‌آمدگویی و ورود OTP", "string")
		_ = upsertSetting("smsir_template_id", tid, "شناسه قالب ارسال سریع پیامک اعتبارسنجی (Verify Template)", "string")
	}
	if req.Template2FA != nil && *req.Template2FA != "" {
		_ = upsertSetting("smsir_template_2fa", strings.TrimSpace(*req.Template2FA), "شناسه الگوی اعلام درخواست کد ۲FA به مشتری", "string")
	}
	if req.TemplateOrderComplete != nil && *req.TemplateOrderComplete != "" {
		_ = upsertSetting("smsir_template_order_complete", strings.TrimSpace(*req.TemplateOrderComplete), "شناسه الگوی تحویل لایسنس و تکمیل سفارش", "string")
	}
	if req.TemplateRenewal != nil && *req.TemplateRenewal != "" {
		_ = upsertSetting("smsir_template_renewal", strings.TrimSpace(*req.TemplateRenewal), "شناسه الگوی یادآور تمدید اشتراک", "string")
	}
	if req.IsSandbox != nil {
		val := "false"
		if *req.IsSandbox {
			val = "true"
		}
		_ = upsertSetting("smsir_is_sandbox", val, "حالت محیط آزمایشی (سندباکس)", "bool")
	}
	if req.TicketSMSMode != nil && *req.TicketSMSMode != "" {
		_ = upsertSetting("ticket_sms_mode", strings.TrimSpace(*req.TicketSMSMode), "حالت پیش‌فرض اطلاع‌رسانی پیامک تیکت‌ها", "string")
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
	templateType := strings.ToLower(strings.TrimSpace(req.TemplateType))
	if templateType == "" {
		templateType = "otp"
	}

	templateID := 0
	if req.TemplateID != nil && *req.TemplateID > 0 {
		templateID = *req.TemplateID
	}

	var params []smsir.VerifyParameter
	detailsDesc := ""

	switch templateType {
	case "two_factor", "2fa":
		if templateID == 0 {
			cleaned := strings.Trim(configMap["smsir_template_2fa"], "'\" \t\r\n")
			if tid, err := strconv.Atoi(cleaned); err == nil && tid > 0 {
				templateID = tid
			}
		}
		if templateID <= 0 {
			templateID = 100001
		}
		orderVal := "ORD-TEST-99"
		promptVal := "لطفاً کد تایید ۲FA را ارسال نمایید"
		if req.Parameters != nil {
			if v, ok := req.Parameters["Order"]; ok && v != "" {
				orderVal = v
			}
			if v, ok := req.Parameters["Prompt"]; ok && v != "" {
				promptVal = v
			}
		}
		params = []smsir.VerifyParameter{
			{Name: "Order", Value: orderVal},
			{Name: "Prompt", Value: promptVal},
		}
		detailsDesc = fmt.Sprintf("تست الگوی اعلام ۲FA به شماره %s (الگو: %d)", phone, templateID)

	case "order_complete", "complete":
		if templateID == 0 {
			cleaned := strings.Trim(configMap["smsir_template_order_complete"], "'\" \t\r\n")
			if tid, err := strconv.Atoi(cleaned); err == nil && tid > 0 {
				templateID = tid
			}
		}
		if templateID <= 0 {
			templateID = 100002
		}
		orderVal := "ORD-TEST-99"
		licenseVal := "ANHX-KEY-XXXX-YYYY"
		if req.Parameters != nil {
			if v, ok := req.Parameters["Order"]; ok && v != "" {
				orderVal = v
			}
			if v, ok := req.Parameters["License"]; ok && v != "" {
				licenseVal = v
			}
		}
		params = []smsir.VerifyParameter{
			{Name: "Order", Value: orderVal},
			{Name: "License", Value: licenseVal},
		}
		detailsDesc = fmt.Sprintf("تست الگوی تکمیل سفارش و لایسنس به شماره %s (الگو: %d)", phone, templateID)

	case "renewal":
		if templateID == 0 {
			cleaned := strings.Trim(configMap["smsir_template_renewal"], "'\" \t\r\n")
			if tid, err := strconv.Atoi(cleaned); err == nil && tid > 0 {
				templateID = tid
			}
		}
		if templateID <= 0 {
			templateID = 100003
		}
		serviceVal := "Google One 2TB"
		daysVal := "۳"
		linkVal := "anahix.ir/renew"
		if req.Parameters != nil {
			if v, ok := req.Parameters["Service"]; ok && v != "" {
				serviceVal = v
			}
			if v, ok := req.Parameters["Days"]; ok && v != "" {
				daysVal = v
			}
			if v, ok := req.Parameters["Link"]; ok && v != "" {
				linkVal = v
			}
		}
		params = []smsir.VerifyParameter{
			{Name: "Service", Value: serviceVal},
			{Name: "Days", Value: daysVal},
			{Name: "Link", Value: linkVal},
		}
		detailsDesc = fmt.Sprintf("تست الگوی یادآور تمدید اشتراک به شماره %s (الگو: %d)", phone, templateID)

	default: // "otp"
		if templateID == 0 {
			cleaned := strings.Trim(configMap["smsir_template_otp"], "'\" \t\r\n")
			if cleaned == "" {
				cleaned = strings.Trim(configMap["smsir_template_id"], "'\" \t\r\n")
			}
			if tid, err := strconv.Atoi(cleaned); err == nil && tid > 0 {
				templateID = tid
			}
		}
		if templateID <= 0 {
			templateID = 100000
		}
		codeVal := "12345"
		if req.Parameters != nil {
			if v, ok := req.Parameters["Code"]; ok && v != "" {
				codeVal = v
			}
		}
		params = []smsir.VerifyParameter{
			{Name: "Code", Value: codeVal},
		}
		detailsDesc = fmt.Sprintf("تست الگوی خوش‌آمدگویی و ورود OTP به شماره %s (الگو: %d)", phone, templateID)
	}

	// Dispatch test pattern SMS
	_, err := l.svcCtx.SMSIRClient.SendPattern(l.ctx, apiKey, phone, templateID, params)
	if err != nil {
		return fmt.Errorf("خطا در ارسال پیامک آزمایشی: %w", err)
	}

	// Record admin audit log
	adminAction := model.AdminActionLog{
		AdminUsername: adminUsername,
		Action:        "send_test_sms",
		Resource:      phone,
		Details:       detailsDesc,
	}
	_ = l.svcCtx.DB.WithContext(l.ctx).Create(&adminAction)

	return nil
}
