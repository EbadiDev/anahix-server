package admin

import (
	"context"
	"fmt"

	"github.com/EbadiDev/anahix-server/internal/model"
	"github.com/EbadiDev/anahix-server/internal/svc"
	"github.com/EbadiDev/anahix-server/internal/types"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ThemeAdminLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewThemeAdminLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ThemeAdminLogic {
	return &ThemeAdminLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ThemeAdminLogic) GetSettings() (*types.ThemeSettingsResponse, error) {
	var settings []model.SystemSetting
	if err := l.svcCtx.DB.WithContext(l.ctx).
		Where("category IN ?", []string{"theme", "auth"}).
		Find(&settings).Error; err != nil {
		return nil, err
	}

	res := &types.ThemeSettingsResponse{
		DefaultTheme:        "graphite",
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
		case "default_theme":
			if s.Value != "" {
				res.DefaultTheme = s.Value
			}
		case "theme_event_name":
			res.EventName = s.Value
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

func (l *ThemeAdminLogic) UpdateSettings(req *types.UpdateThemeSettingsRequest) error {
	boolToStr := func(b bool) string {
		if b {
			return "true"
		}
		return "false"
	}

	defaultTheme := req.DefaultTheme
	if defaultTheme == "" {
		defaultTheme = "graphite"
	}

	settingsToUpsert := []model.SystemSetting{
		{Category: "theme", Key: "default_theme", Value: defaultTheme, Type: "string", Desc: "تم سیستم طراحی پیش‌فرض سراسری سایت"},
		{Category: "theme", Key: "theme_event_name", Value: req.EventName, Type: "string", Desc: "عنوان رویداد یا مناسبت فعال"},
		{Category: "theme", Key: "theme_event_active", Value: boolToStr(req.EventActive), Type: "bool", Desc: "فعال‌سازی حالت رویداد و مناسبت ویژه"},
		{Category: "theme", Key: "theme_primary_color", Value: req.PrimaryColor, Type: "string", Desc: "رنگ اصلی سازمانی و رویداد"},
		{Category: "theme", Key: "theme_accent_color", Value: req.AccentColor, Type: "string", Desc: "رنگ ثانویه یا تاکیدی رویداد"},
		{Category: "theme", Key: "theme_header_announcement", Value: req.HeaderAnnouncement, Type: "string", Desc: "متن پیام نوار اعلان بالای سایت"},
		{Category: "theme", Key: "theme_event_effect", Value: req.EventEffect, Type: "string", Desc: "افکت گرافیکی وایب‌فارسی (برف، ستاره، بادکنک و...)"},
		{Category: "theme", Key: "theme_event_icon", Value: req.EventIcon, Type: "string", Desc: "آیکون سه بعدی یا گرافیکی رویداد"},
		{Category: "auth", Key: "auth_google_enabled", Value: boolToStr(req.AuthGoogleEnabled), Type: "bool", Desc: "امکان ورود با حساب گوگل"},
		{Category: "auth", Key: "auth_github_enabled", Value: boolToStr(req.AuthGithubEnabled), Type: "bool", Desc: "امکان ورود با حساب گیت‌هاب"},
		{Category: "auth", Key: "auth_password_enabled", Value: boolToStr(req.AuthPasswordEnabled), Type: "bool", Desc: "امکان ورود با نام کاربری و کلمه عبور"},
		{Category: "auth", Key: "auth_otp_required", Value: boolToStr(req.AuthOTPRequired), Type: "bool", Desc: "الزام تایید دومرحله‌ای پیامکی OTP"},
	}

	return l.svcCtx.DB.WithContext(l.ctx).Transaction(func(tx *gorm.DB) error {
		for _, s := range settingsToUpsert {
			err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "key"}},
				DoUpdates: clause.AssignmentColumns([]string{"value", "desc", "updated_at"}),
			}).Create(&s).Error
			if err != nil {
				return fmt.Errorf("خطا در ذخیره تنظیم %s: %w", s.Key, err)
			}
		}
		return nil
	})
}
