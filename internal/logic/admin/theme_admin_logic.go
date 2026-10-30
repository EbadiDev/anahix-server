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
		SiteNameFA:          "",
		SiteNameEN:          "",
		SiteSubtitle:        "",
		EventName:           "عادی",
		EventActive:         false,
		PrimaryColor:        "#2563eb",
		AccentColor:         "#3b82f6",
		HeaderAnnouncement:  "تمامی سرویس‌ها و شماره‌های مجازی فعال و تحویل آنی برقرار است",
		EventEffect:         "none",
		EventIcon:           "sparkles",
		AuthGoogleEnabled:   true,
		AuthGithubEnabled:   true,
		AuthPasswordEnabled:          true,
		AuthEmailEnabled:             true,
		AuthOTPRequired:              false,
		TrustTickerEnabled:           true,
		ScrollingAnnouncementEnabled: true,
		ScrollingAnnouncementText:    "🔥 تحویل آنی و خودکار ۰ ثانیه‌ای لایسنس‌ها برقرار است • پشتیبانی ۲۴ ساعته متخصصین هوش مصنوعی در تلگرام",
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
		case "auth_email_enabled":
			res.AuthEmailEnabled = s.Value != "false" && s.Value != "0"
		case "auth_otp_required":
			res.AuthOTPRequired = s.Value == "true" || s.Value == "1"
		case "trust_ticker_enabled":
			res.TrustTickerEnabled = s.Value != "false" && s.Value != "0"
		case "trust_ticker_json":
			res.TrustTickerJSON = s.Value
		case "scrolling_announcement_enabled":
			res.ScrollingAnnouncementEnabled = s.Value != "false" && s.Value != "0"
		case "scrolling_announcement_text":
			if s.Value != "" {
				res.ScrollingAnnouncementText = s.Value
			}
		case "scrolling_announcement_icon":
			if s.Value != "" {
				res.ScrollingAnnouncementIcon = s.Value
			}
		case "scrolling_announcement_badge":
			if s.Value != "" {
				res.ScrollingAnnouncementBadge = s.Value
			}
		case "hero_badge_text":
			if s.Value != "" {
				res.HeroBadgeText = s.Value
			}
		case "hero_title":
			if s.Value != "" {
				res.HeroTitle = s.Value
			}
		case "hero_typewriter_words":
			if s.Value != "" {
				res.HeroTypewriterWords = s.Value
			}
		case "hero_description":
			if s.Value != "" {
				res.HeroDescription = s.Value
			}
		case "hero_btn1_text":
			if s.Value != "" {
				res.HeroBtn1Text = s.Value
			}
		case "hero_btn1_link":
			if s.Value != "" {
				res.HeroBtn1Link = s.Value
			}
		case "hero_btn2_text":
			if s.Value != "" {
				res.HeroBtn2Text = s.Value
			}
		case "hero_btn2_link":
			if s.Value != "" {
				res.HeroBtn2Link = s.Value
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
			res.SiteNameFA = s.Value
		case "site_name_en":
			res.SiteNameEN = s.Value
		case "site_subtitle":
			res.SiteSubtitle = s.Value
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
		case "footer_badges_json":
			res.FooterBadgesJSON = s.Value
		case "footer_description":
			res.FooterDescription = s.Value
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
		case "ticket_sms_mode":
			if s.Value != "" {
				res.TicketSMSMode = s.Value
			}
		case "ticket_email_mode":
			if s.Value != "" {
				res.TicketEmailMode = s.Value
			}
		}
	}

	if res.TicketSMSMode == "" {
		res.TicketSMSMode = "default_on"
	}
	if res.TicketEmailMode == "" {
		res.TicketEmailMode = "default_on"
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
	// Individually update auth settings if provided
	if req.AuthGoogleEnabled != nil {
		settingsToUpsert = append(settingsToUpsert, model.SystemSetting{
			Category: "auth", Key: "auth_google_enabled", Value: boolToStr(*req.AuthGoogleEnabled), Type: "bool", Desc: "امکان ورود با حساب گوگل",
		})
	}
	if req.AuthGithubEnabled != nil {
		settingsToUpsert = append(settingsToUpsert, model.SystemSetting{
			Category: "auth", Key: "auth_github_enabled", Value: boolToStr(*req.AuthGithubEnabled), Type: "bool", Desc: "امکان ورود با حساب گیت‌هاب",
		})
	}
	if req.AuthPasswordEnabled != nil {
		settingsToUpsert = append(settingsToUpsert, model.SystemSetting{
			Category: "auth", Key: "auth_password_enabled", Value: boolToStr(*req.AuthPasswordEnabled), Type: "bool", Desc: "امکان ورود با نام کاربری و کلمه عبور",
		})
	}
	if req.AuthEmailEnabled != nil {
		settingsToUpsert = append(settingsToUpsert, model.SystemSetting{
			Category: "auth", Key: "auth_email_enabled", Value: boolToStr(*req.AuthEmailEnabled), Type: "bool", Desc: "امکان ثبت‌نام و ورود با ایمیل",
		})
	}
	if req.AuthOTPRequired != nil {
		settingsToUpsert = append(settingsToUpsert, model.SystemSetting{
			Category: "auth", Key: "auth_otp_required", Value: boolToStr(*req.AuthOTPRequired), Type: "bool", Desc: "الزام تایید دومرحله‌ای پیامکی OTP",
		})
	}
	if req.TrustTickerEnabled != nil {
		settingsToUpsert = append(settingsToUpsert, model.SystemSetting{
			Category: "theme", Key: "trust_ticker_enabled", Value: boolToStr(*req.TrustTickerEnabled), Type: "bool", Desc: "نمایش نوار تضمین خدمات در صفحه اصلی",
		})
	}
	if req.TrustTickerJSON != "" {
		settingsToUpsert = append(settingsToUpsert, model.SystemSetting{
			Category: "theme", Key: "trust_ticker_json", Value: req.TrustTickerJSON, Type: "string", Desc: "داده‌های متنی و آیکون‌های نوار تضمین خدمات",
		})
	}
	if req.ScrollingAnnouncementEnabled != nil {
		settingsToUpsert = append(settingsToUpsert, model.SystemSetting{
			Category: "theme", Key: "scrolling_announcement_enabled", Value: boolToStr(*req.ScrollingAnnouncementEnabled), Type: "bool", Desc: "نمایش نوار اسکرولی متحرک در صفحه اصلی",
		})
	}
	if req.ScrollingAnnouncementText != "" {
		settingsToUpsert = append(settingsToUpsert, model.SystemSetting{
			Category: "theme", Key: "scrolling_announcement_text", Value: req.ScrollingAnnouncementText, Type: "string", Desc: "متن نوار متحرک اسکرولی جشنواره",
		})
	}
	if req.ScrollingAnnouncementIcon != "" {
		settingsToUpsert = append(settingsToUpsert, model.SystemSetting{
			Category: "theme", Key: "scrolling_announcement_icon", Value: req.ScrollingAnnouncementIcon, Type: "string", Desc: "آیکون نوار متحرک اسکرولی",
		})
	}
	if req.ScrollingAnnouncementBadge != "" {
		settingsToUpsert = append(settingsToUpsert, model.SystemSetting{
			Category: "theme", Key: "scrolling_announcement_badge", Value: req.ScrollingAnnouncementBadge, Type: "string", Desc: "متن نشان نوار متحرک اسکرولی",
		})
	}
	if req.HeroBadgeText != "" {
		settingsToUpsert = append(settingsToUpsert, model.SystemSetting{
			Category: "hero", Key: "hero_badge_text", Value: req.HeroBadgeText, Type: "string", Desc: "متن نشان بالای هیرو",
		})
	}
	if req.HeroTitle != "" {
		settingsToUpsert = append(settingsToUpsert, model.SystemSetting{
			Category: "hero", Key: "hero_title", Value: req.HeroTitle, Type: "string", Desc: "تیتر اصلی هیرو سکشن",
		})
	}
	if req.HeroTypewriterWords != "" {
		settingsToUpsert = append(settingsToUpsert, model.SystemSetting{
			Category: "hero", Key: "hero_typewriter_words", Value: req.HeroTypewriterWords, Type: "string", Desc: "کلمات متحرک تایپ‌رایتر هیرو",
		})
	}
	if req.HeroDescription != "" {
		settingsToUpsert = append(settingsToUpsert, model.SystemSetting{
			Category: "hero", Key: "hero_description", Value: req.HeroDescription, Type: "string", Desc: "متن توضیحات هیرو سکشن",
		})
	}
	if req.HeroBtn1Text != "" {
		settingsToUpsert = append(settingsToUpsert, model.SystemSetting{
			Category: "hero", Key: "hero_btn1_text", Value: req.HeroBtn1Text, Type: "string", Desc: "متن دکمه اول هیرو",
		})
	}
	if req.HeroBtn1Link != "" {
		settingsToUpsert = append(settingsToUpsert, model.SystemSetting{
			Category: "hero", Key: "hero_btn1_link", Value: req.HeroBtn1Link, Type: "string", Desc: "لینک دکمه اول هیرو",
		})
	}
	if req.HeroBtn2Text != "" {
		settingsToUpsert = append(settingsToUpsert, model.SystemSetting{
			Category: "hero", Key: "hero_btn2_text", Value: req.HeroBtn2Text, Type: "string", Desc: "متن دکمه دوم هیرو",
		})
	}
	if req.HeroBtn2Link != "" {
		settingsToUpsert = append(settingsToUpsert, model.SystemSetting{
			Category: "hero", Key: "hero_btn2_link", Value: req.HeroBtn2Link, Type: "string", Desc: "لینک دکمه دوم هیرو",
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
	if req.SiteNameFA != nil {
		settingsToUpsert = append(settingsToUpsert, model.SystemSetting{
			Category: "branding", Key: "site_name_fa", Value: *req.SiteNameFA, Type: "string", Desc: "نام فارسی وبسایت",
		})
	}
	if req.SiteNameEN != nil {
		settingsToUpsert = append(settingsToUpsert, model.SystemSetting{
			Category: "branding", Key: "site_name_en", Value: *req.SiteNameEN, Type: "string", Desc: "نام انگلیسی وبسایت",
		})
	}
	if req.SiteSubtitle != nil {
		settingsToUpsert = append(settingsToUpsert, model.SystemSetting{
			Category: "branding", Key: "site_subtitle", Value: *req.SiteSubtitle, Type: "string", Desc: "زیرعنوان یا شعار برند",
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
	if req.FooterBadgesJSON != "" {
		settingsToUpsert = append(settingsToUpsert, model.SystemSetting{
			Category: "menus", Key: "footer_badges_json", Value: req.FooterBadgesJSON, Type: "json", Desc: "نمادهای اعتماد و مجوزهای فوتر",
		})
	}
	if req.FooterDescription != "" {
		settingsToUpsert = append(settingsToUpsert, model.SystemSetting{
			Category: "branding", Key: "footer_description", Value: req.FooterDescription, Type: "string", Desc: "متن توضیحات و معرفی فوتر",
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
	if req.TicketSMSMode != "" {
		settingsToUpsert = append(settingsToUpsert, model.SystemSetting{
			Category: "theme", Key: "ticket_sms_mode", Value: req.TicketSMSMode, Type: "string", Desc: "حالت پیش‌فرض اطلاع‌رسانی پیامک تیکت‌ها",
		})
	}
	if req.TicketEmailMode != "" {
		settingsToUpsert = append(settingsToUpsert, model.SystemSetting{
			Category: "theme", Key: "ticket_email_mode", Value: req.TicketEmailMode, Type: "string", Desc: "حالت پیش‌فرض اطلاع‌رسانی ایمیل تیکت‌ها",
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
