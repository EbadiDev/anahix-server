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
		Where("category IN ?", []string{"theme", "auth", "checkout", "branding", "seo", "menus", "apikeys", "fivesim"}).
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
		case "offers_badge_text":
			if s.Value != "" {
				res.OffersBadgeText = s.Value
			}
		case "offers_title":
			if s.Value != "" {
				res.OffersTitle = s.Value
			}
		case "offers_subtitle":
			if s.Value != "" {
				res.OffersSubtitle = s.Value
			}
		case "site_name_fa":
			if s.Value != "" {
				res.SiteNameFA = s.Value
			}
		case "site_name_en":
			if s.Value != "" {
				res.SiteNameEN = s.Value
			}
		case "site_subtitle":
			if s.Value != "" {
				res.SiteSubtitle = s.Value
			}
		case "site_favicon_url":
			if s.Value != "" {
				res.SiteFaviconURL = s.Value
			}
		case "site_logo_url":
			if s.Value != "" {
				res.SiteLogoURL = s.Value
			}
		case "meta_title":
			if s.Value != "" {
				res.MetaTitle = s.Value
			}
		case "meta_description":
			if s.Value != "" {
				res.MetaDescription = s.Value
			}
		case "meta_keywords":
			if s.Value != "" {
				res.MetaKeywords = s.Value
			}
		case "og_image_url":
			if s.Value != "" {
				res.OGImageURL = s.Value
			}
		case "support_phone":
			if s.Value != "" {
				res.SupportPhone = s.Value
			}
		case "support_telegram":
			if s.Value != "" {
				res.SupportTelegram = s.Value
			}
		case "header_menu_json":
			if s.Value != "" {
				res.HeaderMenuJSON = s.Value
			}
		case "footer_menu_json":
			if s.Value != "" {
				res.FooterMenuJSON = s.Value
			}
		case "sms_api_key":
			if s.Value != "" {
				res.SMSApiKey = s.Value
			}
		case "sms_line_number":
			if s.Value != "" {
				res.SMSLineNumber = s.Value
			}
		case "zarinpal_merchant_id":
			if s.Value != "" {
				res.ZarinpalMerchantID = s.Value
			}
		case "google_client_id":
			if s.Value != "" {
				res.GoogleClientID = s.Value
			}
		case "github_client_id":
			if s.Value != "" {
				res.GithubClientID = s.Value
			}
		case "fivesim_api_key", "vn_api_key":
			if s.Value != "" {
				res.FivesimApiKey = s.Value
			}
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

	var settingsToUpsert []model.SystemSetting

	if req.DefaultTheme != "" {
		settingsToUpsert = append(settingsToUpsert, model.SystemSetting{
			Category: "theme", Key: "default_theme", Value: req.DefaultTheme, Type: "string", Desc: "تم سیستم طراحی پیش‌فرض سراسری سایت",
		})
	}
	if req.EventName != "" {
		settingsToUpsert = append(settingsToUpsert, model.SystemSetting{
			Category: "theme", Key: "theme_event_name", Value: req.EventName, Type: "string", Desc: "عنوان رویداد یا مناسبت فعال",
		})
	}
	settingsToUpsert = append(settingsToUpsert, model.SystemSetting{
		Category: "theme", Key: "theme_event_active", Value: boolToStr(req.EventActive), Type: "bool", Desc: "فعال‌سازی حالت رویداد و مناسبت ویژه",
	})
	if req.PrimaryColor != "" {
		settingsToUpsert = append(settingsToUpsert, model.SystemSetting{
			Category: "theme", Key: "theme_primary_color", Value: req.PrimaryColor, Type: "string", Desc: "رنگ اصلی سازمانی و رویداد",
		})
	}
	if req.AccentColor != "" {
		settingsToUpsert = append(settingsToUpsert, model.SystemSetting{
			Category: "theme", Key: "theme_accent_color", Value: req.AccentColor, Type: "string", Desc: "رنگ ثانویه یا تاکیدی رویداد",
		})
	}
	if req.HeaderAnnouncement != "" {
		settingsToUpsert = append(settingsToUpsert, model.SystemSetting{
			Category: "theme", Key: "theme_header_announcement", Value: req.HeaderAnnouncement, Type: "string", Desc: "متن پیام نوار اعلان بالای سایت",
		})
	}
	if req.EventEffect != "" {
		settingsToUpsert = append(settingsToUpsert, model.SystemSetting{
			Category: "theme", Key: "theme_event_effect", Value: req.EventEffect, Type: "string", Desc: "افکت گرافیکی وایب‌فارسی",
		})
	}
	if req.EventIcon != "" {
		settingsToUpsert = append(settingsToUpsert, model.SystemSetting{
			Category: "theme", Key: "theme_event_icon", Value: req.EventIcon, Type: "string", Desc: "آیکون سه بعدی یا گرافیکی رویداد",
		})
	}
	// Only update auth settings if they were actually configured/sent
	if req.AuthPasswordEnabled || req.AuthGoogleEnabled || req.AuthGithubEnabled || req.AuthOTPRequired {
		settingsToUpsert = append(settingsToUpsert,
			model.SystemSetting{Category: "auth", Key: "auth_google_enabled", Value: boolToStr(req.AuthGoogleEnabled), Type: "bool", Desc: "امکان ورود با حساب گوگل"},
			model.SystemSetting{Category: "auth", Key: "auth_github_enabled", Value: boolToStr(req.AuthGithubEnabled), Type: "bool", Desc: "امکان ورود با حساب گیت‌هاب"},
			model.SystemSetting{Category: "auth", Key: "auth_password_enabled", Value: boolToStr(req.AuthPasswordEnabled), Type: "bool", Desc: "امکان ورود با نام کاربری و کلمه عبور"},
			model.SystemSetting{Category: "auth", Key: "auth_otp_required", Value: boolToStr(req.AuthOTPRequired), Type: "bool", Desc: "الزام تایید دومرحله‌ای پیامکی OTP"},
		)
	}

	if req.CheckoutChannelSectionEnabled != nil {
		settingsToUpsert = append(settingsToUpsert, model.SystemSetting{
			Category: "checkout",
			Key:      "checkout_channel_section_enabled",
			Value:    boolToStr(*req.CheckoutChannelSectionEnabled),
			Type:     "bool",
			Desc:     "نمایش بخش انتخاب کانال دریافت فاکتور و لایسنس در تسویه‌حساب",
		})
	}
	if req.CheckoutChannelSMSEnabled != nil {
		settingsToUpsert = append(settingsToUpsert, model.SystemSetting{
			Category: "checkout",
			Key:      "checkout_channel_sms_enabled",
			Value:    boolToStr(*req.CheckoutChannelSMSEnabled),
			Type:     "bool",
			Desc:     "فعال‌سازی کانال پیامک جهت دریافت اطلاعات سفارش",
		})
	}
	if req.CheckoutChannelEmailEnabled != nil {
		settingsToUpsert = append(settingsToUpsert, model.SystemSetting{
			Category: "checkout",
			Key:      "checkout_channel_email_enabled",
			Value:    boolToStr(*req.CheckoutChannelEmailEnabled),
			Type:     "bool",
			Desc:     "فعال‌سازی کانال ایمیل جهت دریافت اطلاعات سفارش",
		})
	}
	if req.CheckoutChannelTitle != "" {
		settingsToUpsert = append(settingsToUpsert, model.SystemSetting{
			Category: "checkout",
			Key:      "checkout_channel_title",
			Value:    req.CheckoutChannelTitle,
			Type:     "string",
			Desc:     "عنوان بخش کانال دریافت فاکتور و لایسنس",
		})
	}
	if req.CheckoutChannelSMSLabel != "" {
		settingsToUpsert = append(settingsToUpsert, model.SystemSetting{
			Category: "checkout",
			Key:      "checkout_channel_sms_label",
			Value:    req.CheckoutChannelSMSLabel,
			Type:     "string",
			Desc:     "عنوان دکمه پیامک به موبایل",
		})
	}
	if req.CheckoutChannelEmailLabel != "" {
		settingsToUpsert = append(settingsToUpsert, model.SystemSetting{
			Category: "checkout",
			Key:      "checkout_channel_email_label",
			Value:    req.CheckoutChannelEmailLabel,
			Type:     "string",
			Desc:     "عنوان دکمه ارسال به ایمیل",
		})
	}
	if req.OffersBadgeText != "" {
		settingsToUpsert = append(settingsToUpsert, model.SystemSetting{
			Category: "theme",
			Key:      "offers_badge_text",
			Value:    req.OffersBadgeText,
			Type:     "string",
			Desc:     "عنوان نشان هدر صفحه آفرها",
		})
	}
	if req.OffersTitle != "" {
		settingsToUpsert = append(settingsToUpsert, model.SystemSetting{
			Category: "theme",
			Key:      "offers_title",
			Value:    req.OffersTitle,
			Type:     "string",
			Desc:     "تیتر اصلی هدر صفحه آفرها",
		})
	}
	if req.OffersSubtitle != "" {
		settingsToUpsert = append(settingsToUpsert, model.SystemSetting{
			Category: "theme",
			Key:      "offers_subtitle",
			Value:    req.OffersSubtitle,
			Type:     "string",
			Desc:     "زیرعنوان توضیحات صفحه آفرها",
		})
	}
	if req.SiteNameFA != "" {
		settingsToUpsert = append(settingsToUpsert, model.SystemSetting{
			Category: "branding", Key: "site_name_fa", Value: req.SiteNameFA, Type: "string", Desc: "نام فارسی وبسایت",
		})
	}
	if req.SiteNameEN != "" {
		settingsToUpsert = append(settingsToUpsert, model.SystemSetting{
			Category: "branding", Key: "site_name_en", Value: req.SiteNameEN, Type: "string", Desc: "نام انگلیسی وبسایت",
		})
	}
	if req.SiteSubtitle != "" {
		settingsToUpsert = append(settingsToUpsert, model.SystemSetting{
			Category: "branding", Key: "site_subtitle", Value: req.SiteSubtitle, Type: "string", Desc: "زیرعنوان یا شعار برند",
		})
	}
	if req.SiteFaviconURL != "" {
		settingsToUpsert = append(settingsToUpsert, model.SystemSetting{
			Category: "branding", Key: "site_favicon_url", Value: req.SiteFaviconURL, Type: "string", Desc: "آیکون فاوآیکون وبسایت (FavIcon)",
		})
	}
	if req.SiteLogoURL != "" {
		settingsToUpsert = append(settingsToUpsert, model.SystemSetting{
			Category: "branding", Key: "site_logo_url", Value: req.SiteLogoURL, Type: "string", Desc: "لوگوی رسمی وبسایت",
		})
	}
	if req.MetaTitle != "" {
		settingsToUpsert = append(settingsToUpsert, model.SystemSetting{
			Category: "seo", Key: "meta_title", Value: req.MetaTitle, Type: "string", Desc: "عنوان متا سئو صفحه اصلی",
		})
	}
	if req.MetaDescription != "" {
		settingsToUpsert = append(settingsToUpsert, model.SystemSetting{
			Category: "seo", Key: "meta_description", Value: req.MetaDescription, Type: "string", Desc: "توضیحات متا سئو",
		})
	}
	if req.MetaKeywords != "" {
		settingsToUpsert = append(settingsToUpsert, model.SystemSetting{
			Category: "seo", Key: "meta_keywords", Value: req.MetaKeywords, Type: "string", Desc: "کلمات کلیدی سئو",
		})
	}
	if req.OGImageURL != "" {
		settingsToUpsert = append(settingsToUpsert, model.SystemSetting{
			Category: "seo", Key: "og_image_url", Value: req.OGImageURL, Type: "string", Desc: "تصویر شبکه‌های اجتماعی OpenGraph",
		})
	}
	if req.SupportPhone != "" {
		settingsToUpsert = append(settingsToUpsert, model.SystemSetting{
			Category: "branding", Key: "support_phone", Value: req.SupportPhone, Type: "string", Desc: "شماره تماس پشتیبانی",
		})
	}
	if req.SupportTelegram != "" {
		settingsToUpsert = append(settingsToUpsert, model.SystemSetting{
			Category: "branding", Key: "support_telegram", Value: req.SupportTelegram, Type: "string", Desc: "آیدی تلگرام پشتیبانی",
		})
	}
	if req.HeaderMenuJSON != "" {
		settingsToUpsert = append(settingsToUpsert, model.SystemSetting{
			Category: "menus", Key: "header_menu_json", Value: req.HeaderMenuJSON, Type: "string", Desc: "پیکربندی منوی هدر سایت",
		})
	}
	if req.FooterMenuJSON != "" {
		settingsToUpsert = append(settingsToUpsert, model.SystemSetting{
			Category: "menus", Key: "footer_menu_json", Value: req.FooterMenuJSON, Type: "string", Desc: "پیکربندی منوی فوتر سایت",
		})
	}
	if req.SMSApiKey != "" {
		settingsToUpsert = append(settingsToUpsert, model.SystemSetting{
			Category: "apikeys", Key: "sms_api_key", Value: req.SMSApiKey, Type: "string", Desc: "کلید وب‌سرویس سامانه پیامک",
		})
	}
	if req.SMSLineNumber != "" {
		settingsToUpsert = append(settingsToUpsert, model.SystemSetting{
			Category: "apikeys", Key: "sms_line_number", Value: req.SMSLineNumber, Type: "string", Desc: "شماره خط فرستنده پیامک",
		})
	}
	if req.ZarinpalMerchantID != "" {
		settingsToUpsert = append(settingsToUpsert, model.SystemSetting{
			Category: "apikeys", Key: "zarinpal_merchant_id", Value: req.ZarinpalMerchantID, Type: "string", Desc: "مرچنت کد درگاه بانکی زرین‌پال",
		})
	}
	if req.GoogleClientID != "" {
		settingsToUpsert = append(settingsToUpsert, model.SystemSetting{
			Category: "apikeys", Key: "google_client_id", Value: req.GoogleClientID, Type: "string", Desc: "شناسه کلاینت گوگل OAuth",
		})
	}
	if req.GithubClientID != "" {
		settingsToUpsert = append(settingsToUpsert, model.SystemSetting{
			Category: "apikeys", Key: "github_client_id", Value: req.GithubClientID, Type: "string", Desc: "شناسه کلاینت گیت‌هاب OAuth",
		})
	}
	if req.FivesimApiKey != "" {
		settingsToUpsert = append(settingsToUpsert, model.SystemSetting{
			Category: "fivesim", Key: "fivesim_api_key", Value: req.FivesimApiKey, Type: "string", Desc: "کلید API سامانه شماره مجازی ۵SIM",
		})
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
