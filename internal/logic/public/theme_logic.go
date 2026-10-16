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
		Where("category IN ?", []string{"theme", "auth"}).
		Find(&settings).Error; err != nil {
		return nil, err
	}

	res := &types.ThemeSettingsResponse{
		EventName:           "عادی",
		EventActive:         false,
		PrimaryColor:        "#2563eb",
		AccentColor:         "#3b82f6",
		HeaderAnnouncement:  "تمامی سرویس‌ها و شماره‌های مجازی فعال و تحویل آنی برقرار است",
		EventEffect:         "none",
		EventIcon:           "sparkles",
		AuthGoogleEnabled:   true,
		AuthGithubEnabled:   true,
		AuthPasswordEnabled: true,
		AuthOTPRequired:     false,
	}

	for _, s := range settings {
		switch s.Key {
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
		}
	}

	return res, nil
}
