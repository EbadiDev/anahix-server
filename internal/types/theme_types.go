package types

type ThemeSettingsResponse struct {
	DefaultTheme        string `json:"default_theme"`         // "graphite", "lapis", "turquoise", "saffron", "pomegranate", "custom"
	EventName           string `json:"event_name"`            // e.g. "نوروز ۱۴۰۵", "یلدا", "بلک فرایدی", "عادی"
	EventActive         bool   `json:"event_active"`          // true or false
	PrimaryColor        string `json:"primary_color"`         // hex or HSL e.g. "#2563eb"
	AccentColor         string `json:"accent_color"`          // hex or HSL
	HeaderAnnouncement  string `json:"header_announcement"`   // top bar message
	EventEffect         string `json:"event_effect"`          // "none", "snow", "confetti", "stars", "lanterns"
	EventIcon           string `json:"event_icon"`            // e.g. "sparkles", "gift", "tree", "fire"
	AuthGoogleEnabled   bool   `json:"auth_google_enabled"`   // toggle google oauth
	AuthGithubEnabled   bool   `json:"auth_github_enabled"`   // toggle github oauth
	AuthPasswordEnabled bool   `json:"auth_password_enabled"` // toggle password login
	AuthEmailEnabled    bool   `json:"auth_email_enabled"`    // toggle email registration & login
	AuthOTPRequired     bool   `json:"auth_otp_required"`     // require mobile otp
	OffersBadgeText     string `json:"offers_badge_text,omitempty"`      // custom badge for /offers header
	OffersTitle                   string `json:"offers_title,omitempty"`           // custom title for /offers header
	OffersSubtitle                string `json:"offers_subtitle,omitempty"`        // custom subtitle for /offers header
	SiteNameFA                    string `json:"site_name_fa,omitempty"`
	SiteNameEN                    string `json:"site_name_en,omitempty"`
	SiteSubtitle                  string `json:"site_subtitle,omitempty"`
	SiteFaviconURL                string `json:"site_favicon_url,omitempty"`
	SiteLogoURL                   string `json:"site_logo_url,omitempty"`
	MetaTitle                     string `json:"meta_title,omitempty"`
	MetaDescription               string `json:"meta_description,omitempty"`
	MetaKeywords                  string `json:"meta_keywords,omitempty"`
	OGImageURL                    string `json:"og_image_url,omitempty"`
	SupportPhone                  string `json:"support_phone,omitempty"`
	SupportTelegram               string `json:"support_telegram,omitempty"`
	HeaderMenuJSON                string `json:"header_menu_json,omitempty"`
	FooterMenuJSON                string `json:"footer_menu_json,omitempty"`
	FooterServicesJSON            string `json:"footer_services_json,omitempty"`
	FooterBadgesJSON              string `json:"footer_badges_json,omitempty"`
	FooterDescription             string `json:"footer_description,omitempty"`
	SMSApiKey                     string `json:"sms_api_key,omitempty"`
	SMSLineNumber                 string `json:"sms_line_number,omitempty"`
	ZarinpalMerchantID            string `json:"zarinpal_merchant_id,omitempty"`
	GoogleClientID                string `json:"google_client_id,omitempty"`
	GithubClientID                string `json:"github_client_id,omitempty"`
	FivesimApiKey                 string `json:"fivesim_api_key,omitempty"`
	TicketSMSMode                 string `json:"ticket_sms_mode,omitempty"`
	TicketEmailMode               string `json:"ticket_email_mode,omitempty"`
	TrustTickerEnabled            bool   `json:"trust_ticker_enabled"`
	TrustTickerJSON               string `json:"trust_ticker_json,omitempty"`
	ScrollingAnnouncementEnabled  bool   `json:"scrolling_announcement_enabled"`
	ScrollingAnnouncementText     string `json:"scrolling_announcement_text,omitempty"`
	ScrollingAnnouncementIcon     string `json:"scrolling_announcement_icon,omitempty"`
	ScrollingAnnouncementBadge    string `json:"scrolling_announcement_badge,omitempty"`
	HeroBadgeText                 string `json:"hero_badge_text,omitempty"`
	HeroTitle                     string `json:"hero_title,omitempty"`
	HeroTypewriterWords           string `json:"hero_typewriter_words,omitempty"`
	HeroDescription               string `json:"hero_description,omitempty"`
	HeroBtn1Text                  string `json:"hero_btn1_text,omitempty"`
	HeroBtn1Link                  string `json:"hero_btn1_link,omitempty"`
	HeroBtn2Text                  string `json:"hero_btn2_text,omitempty"`
	HeroBtn2Link                  string `json:"hero_btn2_link,omitempty"`
}

type UpdateThemeSettingsRequest struct {
	DefaultTheme        string `json:"default_theme"`
	EventName           string `json:"event_name"`
	EventActive         bool   `json:"event_active"`
	PrimaryColor        string `json:"primary_color"`
	AccentColor         string `json:"accent_color"`
	HeaderAnnouncement  string `json:"header_announcement"`
	EventEffect         string `json:"event_effect"`
	EventIcon           string `json:"event_icon"`
	AuthGoogleEnabled   *bool  `json:"auth_google_enabled,omitempty"`
	AuthGithubEnabled   *bool  `json:"auth_github_enabled,omitempty"`
	AuthPasswordEnabled *bool  `json:"auth_password_enabled,omitempty"`
	AuthEmailEnabled    *bool  `json:"auth_email_enabled,omitempty"`
	AuthOTPRequired     *bool  `json:"auth_otp_required,omitempty"`
	OffersBadgeText     string `json:"offers_badge_text,omitempty"`
	OffersTitle                   string `json:"offers_title,omitempty"`
	OffersSubtitle                string `json:"offers_subtitle,omitempty"`
	SiteNameFA                    *string `json:"site_name_fa"`
	SiteNameEN                    *string `json:"site_name_en"`
	SiteSubtitle                  *string `json:"site_subtitle"`
	SiteFaviconURL                string `json:"site_favicon_url,omitempty"`
	SiteLogoURL                   string `json:"site_logo_url,omitempty"`
	MetaTitle                     string `json:"meta_title,omitempty"`
	MetaDescription               string `json:"meta_description,omitempty"`
	MetaKeywords                  string `json:"meta_keywords,omitempty"`
	OGImageURL                    string `json:"og_image_url,omitempty"`
	SupportPhone                  string `json:"support_phone,omitempty"`
	SupportTelegram               string `json:"support_telegram,omitempty"`
	HeaderMenuJSON                string `json:"header_menu_json,omitempty"`
	FooterMenuJSON                string `json:"footer_menu_json,omitempty"`
	FooterServicesJSON            string `json:"footer_services_json,omitempty"`
	FooterBadgesJSON              string `json:"footer_badges_json,omitempty"`
	FooterDescription             string `json:"footer_description,omitempty"`
	SMSApiKey                     string `json:"sms_api_key,omitempty"`
	SMSLineNumber                 string `json:"sms_line_number,omitempty"`
	ZarinpalMerchantID            string `json:"zarinpal_merchant_id,omitempty"`
	GoogleClientID                string `json:"google_client_id,omitempty"`
	GithubClientID                string `json:"github_client_id,omitempty"`
	FivesimApiKey                 string `json:"fivesim_api_key,omitempty"`
	TicketSMSMode                 string `json:"ticket_sms_mode,omitempty"`
	TicketEmailMode               string `json:"ticket_email_mode,omitempty"`
	TrustTickerEnabled            *bool  `json:"trust_ticker_enabled,omitempty"`
	TrustTickerJSON               string `json:"trust_ticker_json,omitempty"`
	ScrollingAnnouncementEnabled  *bool  `json:"scrolling_announcement_enabled,omitempty"`
	ScrollingAnnouncementText     string `json:"scrolling_announcement_text,omitempty"`
	ScrollingAnnouncementIcon     string `json:"scrolling_announcement_icon,omitempty"`
	ScrollingAnnouncementBadge    string `json:"scrolling_announcement_badge,omitempty"`
	HeroBadgeText                 string `json:"hero_badge_text,omitempty"`
	HeroTitle                     string `json:"hero_title,omitempty"`
	HeroTypewriterWords           string `json:"hero_typewriter_words,omitempty"`
	HeroDescription               string `json:"hero_description,omitempty"`
	HeroBtn1Text                  string `json:"hero_btn1_text,omitempty"`
	HeroBtn1Link                  string `json:"hero_btn1_link,omitempty"`
	HeroBtn2Text                  string `json:"hero_btn2_text,omitempty"`
	HeroBtn2Link                  string `json:"hero_btn2_link,omitempty"`
}

type PasswordLoginRequest struct {
	Identifier string `json:"identifier" binding:"required"` // phone or email
	Password   string `json:"password" binding:"required"`
}

type OAuthSimulateRequest struct {
	Provider string `json:"provider" binding:"required"` // "google" or "github"
	ID       string `json:"id" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Name     string `json:"name"`
	Avatar   string `json:"avatar"`
}

type SetPasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password" binding:"required,min=6"`
}

type RegisterWithPasswordRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
	FullName string `json:"full_name"`
	Phone    string `json:"phone"`
}
