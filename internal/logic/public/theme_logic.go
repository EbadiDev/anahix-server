package public

import (
	"context"

	"github.com/EbadiDev/anahix-server/internal/model"
	"github.com/EbadiDev/anahix-server/internal/svc"
	"github.com/EbadiDev/anahix-server/internal/types"
)

type ThemeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewThemeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ThemeLogic {
	return &ThemeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ThemeLogic) GetActiveTheme() (*types.ThemeSettingsResponse, error) {
	var settings []model.SystemSetting
	if err := l.svcCtx.DB.WithContext(l.ctx).
		Where("category IN ?", []string{"theme", "auth", "checkout"}).
		Find(&settings).Error; err != nil {
		return nil, err
	}

	res := &types.ThemeSettingsResponse{
		DefaultTheme:                  "graphite",
		EventName:                     "عادی",
		EventActive:                   false,
		PrimaryColor:                  "#2563eb",
		AccentColor:                   "#3b82f6",
		HeaderAnnouncement:            "تمامی سرویس‌ها و شماره‌های مجازی فعال و تحویل آنی برقرار است",
		EventEffect:                   "none",
		EventIcon:                     "sparkles",
		AuthGoogleEnabled:             true,
		AuthGithubEnabled:             true,
		AuthPasswordEnabled:           true,
		AuthOTPRequired:               false,
		CheckoutChannelSectionEnabled: true,
		CheckoutChannelSMSEnabled:     true,
		CheckoutChannelEmailEnabled:   true,
		CheckoutChannelTitle:          "دریافت فاکتور و لایسنس سفارش:",
		CheckoutChannelSMSLabel:       "پیامک به موبایل",
		CheckoutChannelEmailLabel:     "ارسال به ایمیل",
	}

	for _, s := range settings {
		switch s.Key {
		case "default_theme":
			if s.Value != "" {
				res.DefaultTheme = s.Value
			}
		case "theme_event_name":
			if s.Value != "" {
				res.EventName = s.Value
			}
		case "theme_event_active":
			res.EventActive = s.Value == "true" || s.Value == "1"
		case "theme_primary_color":
			if s.Value != "" {
				res.PrimaryColor = s.Value
			}
		case "theme_accent_color":
			if s.Value != "" {
				res.AccentColor = s.Value
			}
		case "theme_header_announcement":
			if s.Value != "" {
				res.HeaderAnnouncement = s.Value
			}
		case "theme_event_effect":
			if s.Value != "" {
				res.EventEffect = s.Value
			}
		case "theme_event_icon":
			if s.Value != "" {
				res.EventIcon = s.Value
			}
		case "auth_google_enabled":
			res.AuthGoogleEnabled = s.Value != "false" && s.Value != "0"
		case "auth_github_enabled":
			res.AuthGithubEnabled = s.Value != "false" && s.Value != "0"
		case "auth_password_enabled":
			res.AuthPasswordEnabled = s.Value != "false" && s.Value != "0"
		case "auth_otp_required":
			res.AuthOTPRequired = s.Value == "true" || s.Value == "1"
		case "checkout_channel_section_enabled":
			res.CheckoutChannelSectionEnabled = s.Value != "false" && s.Value != "0"
		case "checkout_channel_sms_enabled":
			res.CheckoutChannelSMSEnabled = s.Value != "false" && s.Value != "0"
		case "checkout_channel_email_enabled":
			res.CheckoutChannelEmailEnabled = s.Value != "false" && s.Value != "0"
		case "checkout_channel_title":
			if s.Value != "" {
				res.CheckoutChannelTitle = s.Value
			}
		case "checkout_channel_sms_label":
			if s.Value != "" {
				res.CheckoutChannelSMSLabel = s.Value
			}
		case "checkout_channel_email_label":
			if s.Value != "" {
				res.CheckoutChannelEmailLabel = s.Value
			}
		}
	}

	return res, nil
}
