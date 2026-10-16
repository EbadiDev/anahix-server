package types

type BannerItemResponse struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	Subtitle   string `json:"subtitle"`
	BadgeText  string `json:"badge_text"`
	ImageURL   string `json:"image_url"`
	LinkURL    string `json:"link_url"`
	Position   string `json:"position"`
	BgGradient string `json:"bg_gradient"`
	ButtonText string `json:"button_text"`
	SortOrder  int    `json:"sort_order"`
	IsActive   bool   `json:"is_active"`
}

type CreateBannerRequest struct {
	Title      string `json:"title" binding:"required"`
	Subtitle   string `json:"subtitle"`
	BadgeText  string `json:"badge_text"`
	ImageURL   string `json:"image_url" binding:"required"`
	LinkURL    string `json:"link_url"`
	Position   string `json:"position"`
	BgGradient string `json:"bg_gradient"`
	ButtonText string `json:"button_text"`
	SortOrder  int    `json:"sort_order"`
	IsActive   *bool  `json:"is_active"`
}

type UpdateBannerRequest struct {
	Title      *string `json:"title"`
	Subtitle   *string `json:"subtitle"`
	BadgeText  *string `json:"badge_text"`
	ImageURL   *string `json:"image_url"`
	LinkURL    *string `json:"link_url"`
	Position   *string `json:"position"`
	BgGradient *string `json:"bg_gradient"`
	ButtonText *string `json:"button_text"`
	SortOrder  *int    `json:"sort_order"`
	IsActive   *bool   `json:"is_active"`
}
