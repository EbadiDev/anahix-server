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
		Where("category IN ?", []string{"theme", "auth", "checkout", "branding", "seo", "menus", "sms", "hero"}).
		Find(&settings).Error; err != nil {
		return nil, err
	}

	res := &types.ThemeSettingsResponse{
		DefaultTheme:                  "graphite",
		SiteNameFA:                    "",
		SiteNameEN:                    "",
		SiteSubtitle:                  "",
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
		AuthEmailEnabled:              true,
		AuthOTPRequired:               false,
		TrustTickerEnabled:            true,
		ScrollingAnnouncementEnabled:  true,
		ScrollingAnnouncementText:     "🔥 تحویل آنی و خودکار ۰ ثانیه‌ای لایسنس‌ها برقرار است • پشتیبانی ۲۴ ساعته متخصصین هوش مصنوعی در تلگرام",
		TicketSMSMode:                 "default_on",
		TicketEmailMode:               "default_on",
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
		case "footer_services_json":
			res.FooterServicesJSON = s.Value
		case "footer_badges_json":
			res.FooterBadgesJSON = s.Value
		case "footer_description":
			res.FooterDescription = s.Value
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

	return res, nil
}
