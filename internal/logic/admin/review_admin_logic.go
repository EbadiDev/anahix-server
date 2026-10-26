package admin

import (
	"context"
	"math"
	"strings"
	"time"

	"github.com/EbadiDev/anahix-server/internal/model"
	"github.com/EbadiDev/anahix-server/internal/svc"
	"github.com/EbadiDev/anahix-server/internal/types"
	"github.com/EbadiDev/anahix-server/pkg/xerr"
	"github.com/google/uuid"
)

type ReviewAdminLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewReviewAdminLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReviewAdminLogic {
	return &ReviewAdminLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// ListReviews retrieves all reviews with filtering, statistics, and pagination for admin moderation
func (l *ReviewAdminLogic) ListReviews(req *types.AdminReviewListRequest) (*types.AdminReviewListResponse, error) {
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	page := req.Page
	if page <= 0 {
		page = 1
	}

	// 1. Calculate general stats across all reviews
	var totalReviews int64
	var pendingReviews int64
	var approvedReviews int64
	var rejectedReviews int64
	var sumRating int64
	var ratingCount int64

	l.svcCtx.DB.WithContext(l.ctx).Model(&model.Review{}).Count(&totalReviews)
	l.svcCtx.DB.WithContext(l.ctx).Model(&model.Review{}).Where("status = ?", model.ReviewStatusPending).Count(&pendingReviews)
	l.svcCtx.DB.WithContext(l.ctx).Model(&model.Review{}).Where("status = ?", model.ReviewStatusApproved).Count(&approvedReviews)
	l.svcCtx.DB.WithContext(l.ctx).Model(&model.Review{}).Where("status = ?", model.ReviewStatusRejected).Count(&rejectedReviews)

	type RatingSum struct {
		Sum   int64
		Count int64
	}
	var rSum RatingSum
	l.svcCtx.DB.WithContext(l.ctx).Model(&model.Review{}).
		Select("COALESCE(SUM(rating), 0) as sum, COUNT(id) as count").
		Scan(&rSum)

	sumRating = rSum.Sum
	ratingCount = rSum.Count

	var avgRating float64 = 5.0
	if ratingCount > 0 {
		avgRating = math.Round((float64(sumRating)/float64(ratingCount))*10) / 10
	}

	stats := types.AdminReviewStats{
		TotalReviews:    totalReviews,
		PendingReviews:  pendingReviews,
		ApprovedReviews: approvedReviews,
		RejectedReviews: rejectedReviews,
		AverageRating:   avgRating,
	}

	// 2. Query filtered reviews
	query := l.svcCtx.DB.WithContext(l.ctx).Model(&model.Review{})

	if req.Status != "" && req.Status != "all" {
		query = query.Where("status = ?", req.Status)
	}

	if req.ProductID != "" {
		if pUUID, err := uuid.Parse(req.ProductID); err == nil {
			query = query.Where("product_id = ?", pUUID)
		}
	}

	if strings.TrimSpace(req.Search) != "" {
		search := "%" + strings.TrimSpace(req.Search) + "%"
		query = query.Where("customer_name LIKE ? OR customer_phone LIKE ? OR comment LIKE ? OR order_number LIKE ? OR product_title LIKE ?",
			search, search, search, search, search)
	}

	var filteredTotal int64
	query.Count(&filteredTotal)

	offset := (page - 1) * pageSize
	var reviews []model.Review
	// Prioritize pending reviews, then newest
	if err := query.Order("CASE WHEN status = 'pending' THEN 0 ELSE 1 END ASC, created_at DESC").
		Offset(offset).Limit(pageSize).Find(&reviews).Error; err != nil {
		return nil, xerr.NewErrCode(xerr.DatabaseError)
	}

	custMap := make(map[uuid.UUID]model.Customer)
	var custIDs []uuid.UUID
	for _, r := range reviews {
		if r.CustomerID != nil && *r.CustomerID != uuid.Nil {
			custIDs = append(custIDs, *r.CustomerID)
		}
	}
	if len(custIDs) > 0 {
		var customers []model.Customer
		_ = l.svcCtx.DB.WithContext(l.ctx).
			Select("id, avatar_url, email").
			Where("id IN ?", custIDs).
			Find(&customers).Error
		for _, c := range customers {
			custMap[c.ID] = c
		}
	}

	var dtos []types.ReviewItemDTO
	for _, r := range reviews {
		var oID string
		if r.OrderID != nil {
			oID = r.OrderID.String()
		}

		var avatar string
		var email string
		if r.CustomerID != nil {
			if c, ok := custMap[*r.CustomerID]; ok {
				avatar = c.AvatarURL
				email = c.Email
			}
		}

		dtos = append(dtos, types.ReviewItemDTO{
			ID:               r.ID.String(),
			ProductID:        r.ProductID.String(),
			ProductSlug:      r.ProductSlug,
			ProductTitle:     r.ProductTitle,
			OrderID:          oID,
			OrderNumber:      r.OrderNumber,
			CustomerName:     r.CustomerName,
			CustomerPhone:    r.CustomerPhone,
			CustomerAvatar:   avatar,
			CustomerEmail:    email,
			Rating:           r.Rating,
			Comment:          r.Comment,
			ExperienceTags:   r.GetTagsSlice(),
			Status:           string(r.Status),
			IsVerifiedBuyer:  r.IsVerifiedBuyer,
			IsFeatured:       r.IsFeatured,
			HelpfulCount:     r.HelpfulCount,
			AdminReply:       r.AdminReply,
			AdminReplyAuthor: r.AdminReplyAuthor,
			AdminRepliedAt:   r.AdminRepliedAt,
			RejectionReason:  r.RejectionReason,
			CreatedAt:        r.CreatedAt,
		})
	}

	return &types.AdminReviewListResponse{
		Items: dtos,
		Total: filteredTotal,
		Stats: stats,
	}, nil
}

// UpdateReviewStatus approves or rejects a review
func (l *ReviewAdminLogic) UpdateReviewStatus(reviewID string, status string, rejectionReason string, adminUsername string) error {
	revUUID, err := uuid.Parse(reviewID)
	if err != nil {
		return xerr.NewErrMsg("شناسه نظر نامعتبر است")
	}

	validStatus := model.ReviewStatus(status)
	if validStatus != model.ReviewStatusApproved && validStatus != model.ReviewStatusRejected && validStatus != model.ReviewStatusPending {
		return xerr.NewErrMsg("وضعیت نامعتبر است (تنها approved یا rejected مجاز است)")
	}

	var rev model.Review
	if err := l.svcCtx.DB.WithContext(l.ctx).Where("id = ?", revUUID).First(&rev).Error; err != nil {
		return xerr.NewErrMsg("نظر مورد نظر یافت نشد")
	}

	rev.Status = validStatus
	if validStatus == model.ReviewStatusRejected {
		rev.RejectionReason = rejectionReason
	} else if validStatus == model.ReviewStatusApproved {
		rev.RejectionReason = ""
	}

	if err := l.svcCtx.DB.WithContext(l.ctx).Save(&rev).Error; err != nil {
		return xerr.NewErrCode(xerr.DatabaseError)
	}

	return nil
}

// ReplyReview submits or updates official admin response
func (l *ReviewAdminLogic) ReplyReview(reviewID string, reply string, author string, adminUsername string) error {
	revUUID, err := uuid.Parse(reviewID)
	if err != nil {
		return xerr.NewErrMsg("شناسه نظر نامعتبر است")
	}

	cleanedReply := strings.TrimSpace(reply)
	if cleanedReply == "" {
		return xerr.NewErrMsg("متن پاسخ نمی‌تواند خالی باشد")
	}

	if author == "" {
		author = "پشتیبانی رسمی آناهیکس"
	}

	var rev model.Review
	if err := l.svcCtx.DB.WithContext(l.ctx).Where("id = ?", revUUID).First(&rev).Error; err != nil {
		return xerr.NewErrMsg("نظر مورد نظر یافت نشد")
	}

	now := time.Now()
	rev.AdminReply = cleanedReply
	rev.AdminReplyAuthor = author
	rev.AdminRepliedAt = &now

	if err := l.svcCtx.DB.WithContext(l.ctx).Save(&rev).Error; err != nil {
		return xerr.NewErrCode(xerr.DatabaseError)
	}

	return nil
}

// FeatureReview toggles the pinned/featured status of a review
func (l *ReviewAdminLogic) FeatureReview(reviewID string, isFeatured bool) error {
	revUUID, err := uuid.Parse(reviewID)
	if err != nil {
		return xerr.NewErrMsg("شناسه نظر نامعتبر است")
	}

	return l.svcCtx.DB.WithContext(l.ctx).Model(&model.Review{}).
		Where("id = ?", revUUID).
		Update("is_featured", isFeatured).Error
}

// DeleteReview deletes a review
func (l *ReviewAdminLogic) DeleteReview(reviewID string) error {
	revUUID, err := uuid.Parse(reviewID)
	if err != nil {
		return xerr.NewErrMsg("شناسه نظر نامعتبر است")
	}

	return l.svcCtx.DB.WithContext(l.ctx).Where("id = ?", revUUID).Delete(&model.Review{}).Error
}
