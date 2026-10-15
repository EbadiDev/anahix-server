package model

import (
	"encoding/json"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ContentCategory represents taxonomy for both Blog articles and Help guides
type ContentCategory struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	Name        string         `gorm:"size:120;not null" json:"name"`
	Slug        string         `gorm:"size:120;uniqueIndex;not null" json:"slug"`
	Type        string         `gorm:"size:30;not null;index" json:"type"` // "blog" | "help"
	Icon        string         `gorm:"size:100;default:'BookOpen'" json:"icon"`
	Description string         `gorm:"type:text" json:"description"`
	SortOrder   int            `gorm:"default:0" json:"sort_order"`
	IsActive    bool           `gorm:"default:true;index" json:"is_active"`
	ArticleCount int64         `gorm:"-" json:"article_count,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (c *ContentCategory) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

// TOCItem represents an entry in the Table of Contents for quick navigation
type TOCItem struct {
	ID    string `json:"id"`
	Text  string `json:"text"`
	Level int    `json:"level"` // 2 for H2, 3 for H3
}

// Article represents blog posts, news, tutorials and AI comparisons
type Article struct {
	ID                  uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	Title               string         `gorm:"size:255;not null" json:"title"`
	Slug                string         `gorm:"size:150;uniqueIndex;not null" json:"slug"`
	Excerpt             string         `gorm:"type:text" json:"excerpt"`
	Content             string         `gorm:"type:longtext" json:"content"` // Markdown / Rich HTML
	CoverImage          string         `gorm:"size:512" json:"cover_image"`
	CategorySlug        string         `gorm:"size:100;index" json:"category_slug"`
	CategoryName        string         `gorm:"size:120" json:"category_name"`
	Tags                string         `gorm:"type:text" json:"tags"` // JSON array string
	AuthorName          string         `gorm:"size:120;default:'تحریریه آناهیکس'" json:"author_name"`
	AuthorRole          string         `gorm:"size:120;default:'کارشناس هوش مصنوعی و امنیت'" json:"author_role"`
	AuthorAvatar        string         `gorm:"size:512" json:"author_avatar"`
	ReadingTimeMinutes  int            `gorm:"default:5" json:"reading_time_minutes"`
	ViewCount           int64          `gorm:"default:0" json:"view_count"`
	LikeCount           int64          `gorm:"default:0" json:"like_count"`
	IsFeatured          bool           `gorm:"default:false;index" json:"is_featured"`
	IsPublished         bool           `gorm:"default:true;index" json:"is_published"`
	PublishedAt         *time.Time     `json:"published_at"`

	// Store Integration (CTA to purchase related account or virtual number)
	RelatedProductID    *uuid.UUID     `gorm:"type:uuid;index" json:"related_product_id,omitempty"`
	RelatedProductTitle string         `gorm:"size:255" json:"related_product_title,omitempty"`
	RelatedProductSlug  string         `gorm:"size:120" json:"related_product_slug,omitempty"`
	RelatedProductPrice int64          `gorm:"default:0" json:"related_product_price,omitempty"`

	// SEO Meta Tags & Rich Snippets
	MetaTitle           string         `gorm:"size:255" json:"meta_title"`
	MetaDescription     string         `gorm:"type:text" json:"meta_description"`
	MetaKeywords        string         `gorm:"size:500" json:"meta_keywords"`
	CanonicalURL        string         `gorm:"size:512" json:"canonical_url"`

	// Table of contents stored as JSON string
	TableOfContentsJSON string         `gorm:"type:text" json:"table_of_contents_json"`

	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	DeletedAt           gorm.DeletedAt `gorm:"index" json:"-"`
}

func (a *Article) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	if a.ReadingTimeMinutes <= 0 {
		a.ReadingTimeMinutes = CalculateReadingTime(a.Content)
	}
	if a.PublishedAt == nil && a.IsPublished {
		now := time.Now()
		a.PublishedAt = &now
	}
	return nil
}

func (a *Article) BeforeUpdate(tx *gorm.DB) error {
	if a.ReadingTimeMinutes <= 0 {
		a.ReadingTimeMinutes = CalculateReadingTime(a.Content)
	}
	return nil
}

// GetTagsSlice deserializes tags JSON array string
func (a *Article) GetTagsSlice() []string {
	if a.Tags == "" {
		return []string{}
	}
	var list []string
	if err := json.Unmarshal([]byte(a.Tags), &list); err != nil {
		// Fallback to comma separated
		parts := strings.Split(a.Tags, ",")
		res := make([]string, 0, len(parts))
		for _, p := range parts {
			trimmed := strings.TrimSpace(p)
			if trimmed != "" {
				res = append(res, trimmed)
			}
		}
		return res
	}
	return list
}

// HelpGuide represents Knowledge Base articles, error troubleshooting, and service guides
type HelpGuide struct {
	ID                  uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	Title               string         `gorm:"size:255;not null" json:"title"`
	Slug                string         `gorm:"size:150;uniqueIndex;not null" json:"slug"`
	Excerpt             string         `gorm:"type:text" json:"excerpt"`
	Content             string         `gorm:"type:longtext" json:"content"`
	CategorySlug        string         `gorm:"size:100;index" json:"category_slug"`
	CategoryName        string         `gorm:"size:120" json:"category_name"`
	Icon                string         `gorm:"size:100;default:'HelpCircle'" json:"icon"`
	Difficulty          string         `gorm:"size:50;default:'ساده'" json:"difficulty"` // "ساده", "متوسط", "پیشرفته"
	ServiceTarget       string         `gorm:"size:60;index;default:'general'" json:"service_target"` // "gemini", "chatgpt", "claude", "virtual_number", "general"
	ReadingTimeMinutes  int            `gorm:"default:3" json:"reading_time_minutes"`
	ViewCount           int64          `gorm:"default:0" json:"view_count"`
	HelpfulYes          int64          `gorm:"default:0" json:"helpful_yes"`
	HelpfulNo           int64          `gorm:"default:0" json:"helpful_no"`
	SortOrder           int            `gorm:"default:0" json:"sort_order"`
	IsFeatured          bool           `gorm:"default:false;index" json:"is_featured"`
	IsPublished         bool           `gorm:"default:true;index" json:"is_published"`
	PublishedAt         *time.Time     `json:"published_at"`

	// Store Integration (CTA to purchase related account or virtual number)
	RelatedProductID    *uuid.UUID     `gorm:"type:uuid;index" json:"related_product_id,omitempty"`
	RelatedProductTitle string         `gorm:"size:255" json:"related_product_title,omitempty"`
	RelatedProductSlug  string         `gorm:"size:120" json:"related_product_slug,omitempty"`
	RelatedProductPrice int64          `gorm:"default:0" json:"related_product_price,omitempty"`

	// SEO Meta Tags
	MetaTitle           string         `gorm:"size:255" json:"meta_title"`
	MetaDescription     string         `gorm:"type:text" json:"meta_description"`
	MetaKeywords        string         `gorm:"size:500" json:"meta_keywords"`

	TableOfContentsJSON string         `gorm:"type:text" json:"table_of_contents_json"`

	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	DeletedAt           gorm.DeletedAt `gorm:"index" json:"-"`
}

func (h *HelpGuide) BeforeCreate(tx *gorm.DB) error {
	if h.ID == uuid.Nil {
		h.ID = uuid.New()
	}
	if h.ReadingTimeMinutes <= 0 {
		h.ReadingTimeMinutes = CalculateReadingTime(h.Content)
	}
	if h.PublishedAt == nil && h.IsPublished {
		now := time.Now()
		h.PublishedAt = &now
	}
	return nil
}

func (h *HelpGuide) BeforeUpdate(tx *gorm.DB) error {
	if h.ReadingTimeMinutes <= 0 {
		h.ReadingTimeMinutes = CalculateReadingTime(h.Content)
	}
	return nil
}

// ArticleComment represents reader comments on blog posts
type ArticleComment struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	ArticleID   uuid.UUID      `gorm:"type:uuid;not null;index" json:"article_id"`
	CustomerID  *uuid.UUID     `gorm:"type:uuid;index" json:"customer_id,omitempty"`
	AuthorName  string         `gorm:"size:120;not null" json:"author_name"`
	Email       string         `gorm:"size:120" json:"email,omitempty"`
	Content     string         `gorm:"type:text;not null" json:"content"`
	Status      string         `gorm:"size:30;not null;default:'pending';index" json:"status"` // "pending", "approved", "rejected"
	AdminReply  string         `gorm:"type:text" json:"admin_reply,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (c *ArticleComment) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	if c.Status == "" {
		c.Status = "pending"
	}
	return nil
}

// GuideFeedback prevents repeated helpful/unhelpful votes from the same IP/voter
type GuideFeedback struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	GuideID   uuid.UUID `gorm:"type:uuid;not null;index" json:"guide_id"`
	VoterKey  string    `gorm:"size:120;not null;index" json:"voter_key"` // IP or customer phone
	IsHelpful bool      `gorm:"not null" json:"is_helpful"`
	CreatedAt time.Time `json:"created_at"`
}

func (f *GuideFeedback) BeforeCreate(tx *gorm.DB) error {
	if f.ID == uuid.Nil {
		f.ID = uuid.New()
	}
	return nil
}

// ContentTemplate stores pre-built reusable drafting templates for articles and help guides
type ContentTemplate struct {
	ID              uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	Title           string         `gorm:"size:150;not null" json:"title"`
	Slug            string         `gorm:"size:120;uniqueIndex;not null" json:"slug"`
	Type            string         `gorm:"size:30;not null;default:'all';index" json:"type"` // "all" | "blog" | "help"
	Icon            string         `gorm:"size:60;default:'FileText'" json:"icon"`
	Description     string         `gorm:"size:255" json:"description"`
	TemplateContent string         `gorm:"type:longtext;not null" json:"template_content"`
	SortOrder       int            `gorm:"default:0" json:"sort_order"`
	IsActive        bool           `gorm:"default:true;index" json:"is_active"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

func (t *ContentTemplate) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

// ServiceTarget represents AI services and platform targets for guides and products
type ServiceTarget struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Key       string    `gorm:"size:60;uniqueIndex;not null" json:"key"` // "gemini", "chatgpt", "claude", "virtual_number", "general"
	Label     string    `gorm:"size:120;not null" json:"label"`
	Icon      string    `gorm:"size:60;default:'Sparkles'" json:"icon"`
	SortOrder int       `gorm:"default:0" json:"sort_order"`
	IsActive  bool      `gorm:"default:true;index" json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (s *ServiceTarget) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

// CalculateReadingTime calculates estimated reading time in minutes based on Persian & English word count
func CalculateReadingTime(text string) int {
	clean := strings.TrimSpace(text)
	if clean == "" {
		return 1
	}
	words := strings.Fields(clean)
	wordCount := len(words)
	// Average reading speed: ~180 words per minute for mixed Persian/English tech articles
	minutes := int(math.Ceil(float64(wordCount) / 180.0))
	if minutes < 1 {
		return 1
	}
	return minutes
}
