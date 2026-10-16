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
	AuthOTPRequired     bool   `json:"auth_otp_required"`     // require mobile otp
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
	AuthGoogleEnabled   bool   `json:"auth_google_enabled"`
	AuthGithubEnabled   bool   `json:"auth_github_enabled"`
	AuthPasswordEnabled bool   `json:"auth_password_enabled"`
	AuthOTPRequired     bool   `json:"auth_otp_required"`
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
