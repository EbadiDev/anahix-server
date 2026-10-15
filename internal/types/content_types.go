package types

import "time"

// Content Categories DTOs
type CategoryDTO struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Slug         string `json:"slug"`
	Type         string `json:"type"` // "blog" | "help"
	Icon         string `json:"icon"`
	Description  string `json:"description"`
	SortOrder    int    `json:"sort_order"`
	IsActive     bool   `json:"is_active"`
	ArticleCount int64  `json:"article_count"`
}

type CreateCategoryRequest struct {
	Name        string `json:"name" binding:"required"`
	Slug        string `json:"slug" binding:"required"`
	Type        string `json:"type" binding:"required"` // "blog" | "help"
	Icon        string `json:"icon"`
	Description string `json:"description"`
	SortOrder   int    `json:"sort_order"`
	IsActive    *bool  `json:"is_active"`
}

type UpdateCategoryRequest struct {
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Type        string `json:"type"`
	Icon        string `json:"icon"`
	Description string `json:"description"`
	SortOrder   int    `json:"sort_order"`
	IsActive    *bool  `json:"is_active"`
}

// Blog Article DTOs
type ArticleListItemDTO struct {
	ID                  string    `json:"id"`
	Title               string    `json:"title"`
	Slug                string    `json:"slug"`
	Excerpt             string    `json:"excerpt"`
	CoverImage          string    `json:"cover_image"`
	CategorySlug        string    `json:"category_slug"`
	CategoryName        string    `json:"category_name"`
	Tags                []string  `json:"tags"`
	AuthorName          string    `json:"author_name"`
	AuthorRole          string    `json:"author_role"`
	AuthorAvatar        string    `json:"author_avatar"`
	ReadingTimeMinutes  int       `json:"reading_time_minutes"`
	ViewCount           int64     `json:"view_count"`
	LikeCount           int64     `json:"like_count"`
	IsFeatured          bool      `json:"is_featured"`
	IsPublished         bool      `json:"is_published"`
	PublishedAt         *time.Time `json:"published_at"`
	RelatedProductTitle string    `json:"related_product_title,omitempty"`
	RelatedProductSlug  string    `json:"related_product_slug,omitempty"`
	CreatedAt           time.Time `json:"created_at"`
}

type ArticleDetailDTO struct {
	ID                  string            `json:"id"`
	Title               string            `json:"title"`
	Slug                string            `json:"slug"`
	Excerpt             string            `json:"excerpt"`
	Content             string            `json:"content"`
	CoverImage          string            `json:"cover_image"`
	CategorySlug        string            `json:"category_slug"`
	CategoryName        string            `json:"category_name"`
	Tags                []string          `json:"tags"`
	AuthorName          string            `json:"author_name"`
	AuthorRole          string            `json:"author_role"`
	AuthorAvatar        string            `json:"author_avatar"`
	ReadingTimeMinutes  int               `json:"reading_time_minutes"`
	ViewCount           int64             `json:"view_count"`
	LikeCount           int64             `json:"like_count"`
	IsFeatured          bool              `json:"is_featured"`
	IsPublished         bool              `json:"is_published"`
	PublishedAt         *time.Time        `json:"published_at"`
	RelatedProductID    string            `json:"related_product_id,omitempty"`
	RelatedProductTitle string            `json:"related_product_title,omitempty"`
	RelatedProductSlug  string            `json:"related_product_slug,omitempty"`
	RelatedProductPrice int64             `json:"related_product_price,omitempty"`
	MetaTitle           string            `json:"meta_title"`
	MetaDescription     string            `json:"meta_description"`
	MetaKeywords        string            `json:"meta_keywords"`
	CanonicalURL        string            `json:"canonical_url"`
	TableOfContents     []TOCItemDTO      `json:"table_of_contents"`
	Comments            []ArticleCommentDTO `json:"comments"`
	RelatedArticles     []ArticleListItemDTO `json:"related_articles"`
	CreatedAt           time.Time         `json:"created_at"`
	UpdatedAt           time.Time         `json:"updated_at"`
}

type TOCItemDTO struct {
	ID    string `json:"id"`
	Text  string `json:"text"`
	Level int    `json:"level"`
}

type CreateArticleRequest struct {
	Title               string   `json:"title" binding:"required"`
	Slug                string   `json:"slug" binding:"required"`
	Excerpt             string   `json:"excerpt"`
	Content             string   `json:"content" binding:"required"`
	CoverImage          string   `json:"cover_image"`
	CategorySlug        string   `json:"category_slug"`
	CategoryName        string   `json:"category_name"`
	Tags                []string `json:"tags"`
	AuthorName          string   `json:"author_name"`
	AuthorRole          string   `json:"author_role"`
	AuthorAvatar        string   `json:"author_avatar"`
	ReadingTimeMinutes  int      `json:"reading_time_minutes"`
	IsFeatured          bool     `json:"is_featured"`
	IsPublished         *bool    `json:"is_published"`
	RelatedProductID    string   `json:"related_product_id"`
	RelatedProductTitle string   `json:"related_product_title"`
	RelatedProductSlug  string   `json:"related_product_slug"`
	RelatedProductPrice int64    `json:"related_product_price"`
	MetaTitle           string   `json:"meta_title"`
	MetaDescription     string   `json:"meta_description"`
	MetaKeywords        string   `json:"meta_keywords"`
	CanonicalURL        string   `json:"canonical_url"`
	TableOfContentsJSON string   `json:"table_of_contents_json"`
}

type UpdateArticleRequest struct {
	Title               string   `json:"title"`
	Slug                string   `json:"slug"`
	Excerpt             string   `json:"excerpt"`
	Content             string   `json:"content"`
	CoverImage          string   `json:"cover_image"`
	CategorySlug        string   `json:"category_slug"`
	CategoryName        string   `json:"category_name"`
	Tags                []string `json:"tags"`
	AuthorName          string   `json:"author_name"`
	AuthorRole          string   `json:"author_role"`
	AuthorAvatar        string   `json:"author_avatar"`
	ReadingTimeMinutes  int      `json:"reading_time_minutes"`
	IsFeatured          *bool    `json:"is_featured"`
	IsPublished         *bool    `json:"is_published"`
	RelatedProductID    string   `json:"related_product_id"`
	RelatedProductTitle string   `json:"related_product_title"`
	RelatedProductSlug  string   `json:"related_product_slug"`
	RelatedProductPrice int64    `json:"related_product_price"`
	MetaTitle           string   `json:"meta_title"`
	MetaDescription     string   `json:"meta_description"`
	MetaKeywords        string   `json:"meta_keywords"`
	CanonicalURL        string   `json:"canonical_url"`
	TableOfContentsJSON string   `json:"table_of_contents_json"`
}

// Help Guide DTOs
type HelpGuideListItemDTO struct {
	ID                  string     `json:"id"`
	Title               string     `json:"title"`
	Slug                string     `json:"slug"`
	Excerpt             string     `json:"excerpt"`
	CategorySlug        string     `json:"category_slug"`
	CategoryName        string     `json:"category_name"`
	Icon                string     `json:"icon"`
	Difficulty          string     `json:"difficulty"`
	ServiceTarget       string     `json:"service_target"`
	ReadingTimeMinutes  int        `json:"reading_time_minutes"`
	ViewCount           int64      `json:"view_count"`
	HelpfulYes          int64      `json:"helpful_yes"`
	HelpfulNo           int64      `json:"helpful_no"`
	IsFeatured          bool       `json:"is_featured"`
	IsPublished         bool       `json:"is_published"`
	PublishedAt         *time.Time `json:"published_at"`
	RelatedProductTitle string     `json:"related_product_title,omitempty"`
	RelatedProductSlug  string     `json:"related_product_slug,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`
}

type HelpGuideDetailDTO struct {
	ID                  string                `json:"id"`
	Title               string                `json:"title"`
	Slug                string                `json:"slug"`
	Excerpt             string                `json:"excerpt"`
	Content             string                `json:"content"`
	CategorySlug        string                `json:"category_slug"`
	CategoryName        string                `json:"category_name"`
	Icon                string                `json:"icon"`
	Difficulty          string                `json:"difficulty"`
	ServiceTarget       string                `json:"service_target"`
	ReadingTimeMinutes  int                   `json:"reading_time_minutes"`
	ViewCount           int64                 `json:"view_count"`
	HelpfulYes          int64                 `json:"helpful_yes"`
	HelpfulNo           int64                 `json:"helpful_no"`
	IsFeatured          bool                  `json:"is_featured"`
	IsPublished         bool                  `json:"is_published"`
	PublishedAt         *time.Time            `json:"published_at"`
	RelatedProductID    string                `json:"related_product_id,omitempty"`
	RelatedProductTitle string                `json:"related_product_title,omitempty"`
	RelatedProductSlug  string                `json:"related_product_slug,omitempty"`
	RelatedProductPrice int64                 `json:"related_product_price,omitempty"`
	MetaTitle           string                `json:"meta_title"`
	MetaDescription     string                `json:"meta_description"`
	MetaKeywords        string                `json:"meta_keywords"`
	TableOfContents     []TOCItemDTO          `json:"table_of_contents"`
	RelatedGuides       []HelpGuideListItemDTO `json:"related_guides"`
	CreatedAt           time.Time             `json:"created_at"`
	UpdatedAt           time.Time             `json:"updated_at"`
}

type CreateHelpGuideRequest struct {
	Title               string `json:"title" binding:"required"`
	Slug                string `json:"slug" binding:"required"`
	Excerpt             string `json:"excerpt"`
	Content             string `json:"content" binding:"required"`
	CategorySlug        string `json:"category_slug"`
	CategoryName        string `json:"category_name"`
	Icon                string `json:"icon"`
	Difficulty          string `json:"difficulty"`
	ServiceTarget       string `json:"service_target"`
	ReadingTimeMinutes  int    `json:"reading_time_minutes"`
	SortOrder           int    `json:"sort_order"`
	IsFeatured          bool   `json:"is_featured"`
	IsPublished         *bool  `json:"is_published"`
	RelatedProductID    string `json:"related_product_id"`
	RelatedProductTitle string `json:"related_product_title"`
	RelatedProductSlug  string `json:"related_product_slug"`
	RelatedProductPrice int64  `json:"related_product_price"`
	MetaTitle           string `json:"meta_title"`
	MetaDescription     string `json:"meta_description"`
	MetaKeywords        string `json:"meta_keywords"`
	TableOfContentsJSON string `json:"table_of_contents_json"`
}

type UpdateHelpGuideRequest struct {
	Title               string `json:"title"`
	Slug                string `json:"slug"`
	Excerpt             string `json:"excerpt"`
	Content             string `json:"content"`
	CategorySlug        string `json:"category_slug"`
	CategoryName        string `json:"category_name"`
	Icon                string `json:"icon"`
	Difficulty          string `json:"difficulty"`
	ServiceTarget       string `json:"service_target"`
	ReadingTimeMinutes  int    `json:"reading_time_minutes"`
	SortOrder           int    `json:"sort_order"`
	IsFeatured          *bool  `json:"is_featured"`
	IsPublished         *bool  `json:"is_published"`
	RelatedProductID    string `json:"related_product_id"`
	RelatedProductTitle string `json:"related_product_title"`
	RelatedProductSlug  string `json:"related_product_slug"`
	RelatedProductPrice int64  `json:"related_product_price"`
	MetaTitle           string `json:"meta_title"`
	MetaDescription     string `json:"meta_description"`
	MetaKeywords        string `json:"meta_keywords"`
	TableOfContentsJSON string `json:"table_of_contents_json"`
}

// Comments & Feedback DTOs
type ArticleCommentDTO struct {
	ID         string    `json:"id"`
	ArticleID  string    `json:"article_id"`
	AuthorName string    `json:"author_name"`
	Content    string    `json:"content"`
	Status     string    `json:"status"`
	AdminReply string    `json:"admin_reply,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

type SubmitArticleCommentRequest struct {
	AuthorName string `json:"author_name" binding:"required"`
	Email      string `json:"email"`
	Content    string `json:"content" binding:"required"`
}

type GuideFeedbackRequest struct {
	IsHelpful bool `json:"is_helpful"`
}

// Content Templates DTOs (Dynamic DB-driven writing templates)
type ContentTemplateDTO struct {
	ID              string `json:"id"`
	Title           string `json:"title"`
	Slug            string `json:"slug"`
	Type            string `json:"type"` // "all" | "blog" | "help"
	Icon            string `json:"icon"`
	Description     string `json:"description"`
	TemplateContent string `json:"template_content"`
	SortOrder       int    `json:"sort_order"`
	IsActive        bool   `json:"is_active"`
}

type CreateTemplateRequest struct {
	Title           string `json:"title" binding:"required"`
	Slug            string `json:"slug" binding:"required"`
	Type            string `json:"type"`
	Icon            string `json:"icon"`
	Description     string `json:"description"`
	TemplateContent string `json:"template_content" binding:"required"`
	SortOrder       int    `json:"sort_order"`
	IsActive        *bool  `json:"is_active"`
}

type UpdateTemplateRequest struct {
	Title           string `json:"title"`
	Slug            string `json:"slug"`
	Type            string `json:"type"`
	Icon            string `json:"icon"`
	Description     string `json:"description"`
	TemplateContent string `json:"template_content"`
	SortOrder       int    `json:"sort_order"`
	IsActive        *bool  `json:"is_active"`
}

// Service Target DTOs (Dynamic AI platform targets)
type ServiceTargetDTO struct {
	ID        string `json:"id"`
	Key       string `json:"key"`
	Label     string `json:"label"`
	Icon      string `json:"icon"`
	SortOrder int    `json:"sort_order"`
	IsActive  bool   `json:"is_active"`
}

type CreateServiceTargetRequest struct {
	Key       string `json:"key" binding:"required"`
	Label     string `json:"label" binding:"required"`
	Icon      string `json:"icon"`
	SortOrder int    `json:"sort_order"`
	IsActive  *bool  `json:"is_active"`
}

type ContentListResponse[T any] struct {
	Items      []T   `json:"items"`
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalPages int   `json:"total_pages"`
}
