package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ReviewStatus string

const (
	ReviewStatusPending  ReviewStatus = "pending"  // در انتظار تایید مدیریت
	ReviewStatusApproved ReviewStatus = "approved" // تایید شده و عمومی
	ReviewStatusRejected ReviewStatus = "rejected" // رد شده
)

// Review represents buyer social proof, rating, comment and tags
type Review struct {
	ID               uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	ProductID        uuid.UUID      `gorm:"type:uuid;not null;index" json:"product_id"`
	ProductSlug      string         `gorm:"size:120;index" json:"product_slug"`
	ProductTitle     string         `gorm:"size:255" json:"product_title"`
	OrderID          *uuid.UUID     `gorm:"type:uuid;index" json:"order_id,omitempty"`
	OrderNumber      string         `gorm:"size:64;index" json:"order_number,omitempty"`
	CustomerID       *uuid.UUID     `gorm:"type:uuid;index" json:"customer_id,omitempty"`
	CustomerPhone    string         `gorm:"size:30;index" json:"customer_phone"`
	CustomerName     string         `gorm:"size:120" json:"customer_name"`
	Rating           int            `gorm:"not null;index" json:"rating"` // 1 to 5 stars
	Comment          string         `gorm:"type:text" json:"comment"`
	ExperienceTags   string         `gorm:"type:text" json:"experience_tags"` // JSON array string
	Status           ReviewStatus   `gorm:"size:30;not null;default:'pending';index" json:"status"`
	IsVerifiedBuyer  bool           `gorm:"default:true;index" json:"is_verified_buyer"`
	IsFeatured       bool           `gorm:"default:false;index" json:"is_featured"` // Pinned by admin
	HelpfulCount     int            `gorm:"default:0" json:"helpful_count"`
	AdminReply       string         `gorm:"type:text" json:"admin_reply,omitempty"`
	AdminReplyAuthor string         `gorm:"size:120" json:"admin_reply_author,omitempty"`
	AdminRepliedAt   *time.Time     `json:"admin_replied_at,omitempty"`
	RejectionReason  string         `gorm:"size:255" json:"rejection_reason,omitempty"`

	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
}

func (r *Review) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}

// GetTagsSlice deserializes the JSON experience tags string
func (r *Review) GetTagsSlice() []string {
	if r.ExperienceTags == "" {
		return []string{}
	}
	var tags []string
	if err := json.Unmarshal([]byte(r.ExperienceTags), &tags); err != nil {
		return []string{}
	}
	return tags
}

// ReviewHelpfulVote tracks votes to prevent duplicate upvotes
type ReviewHelpfulVote struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	ReviewID  uuid.UUID `gorm:"type:uuid;not null;index" json:"review_id"`
	VoterKey  string    `gorm:"size:120;not null;index" json:"voter_key"` // IP or Customer Phone/ID
	CreatedAt time.Time `json:"created_at"`
}

func (v *ReviewHelpfulVote) BeforeCreate(tx *gorm.DB) error {
	if v.ID == uuid.Nil {
		v.ID = uuid.New()
	}
	return nil
}
