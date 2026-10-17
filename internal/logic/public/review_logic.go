package public

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/EbadiDev/anahix-server/internal/model"
	"github.com/EbadiDev/anahix-server/internal/svc"
	"github.com/EbadiDev/anahix-server/internal/types"
	"github.com/EbadiDev/anahix-server/pkg/xerr"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ReviewLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewReviewLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReviewLogic {
	return &ReviewLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// maskPhone masks mobile number for privacy e.g. 09123456789 -> 0912***6789
func maskPhone(phone string) string {
	cleaned := strings.TrimSpace(phone)
	if len(cleaned) < 8 {
		return "کاربر آناهیکس"
	}
	return cleaned[:4] + "***" + cleaned[len(cleaned)-4:]
}

// formatCustomerName produces a friendly display name for social proof
func formatCustomerName(name, phone string) string {
	if strings.TrimSpace(name) != "" {
		return strings.TrimSpace(name)
	}
	if strings.TrimSpace(phone) != "" {
		return fmt.Sprintf("خریدار (%s)", maskPhone(phone))
	}
	return "خریدار تاییدشده آناهیکس"
}

// GetProductReviews returns approved reviews and comprehensive summary stats for a product
func (l *ReviewLogic) GetProductReviews(productSlug string, ratingFilter int, sortOrder string, onlyWithComments bool, page, pageSize int) (*types.ProductReviewsResponse, error) {
	if pageSize <= 0 {
		pageSize = 20
	}
	if page <= 0 {
		page = 1
	}

	// 1. Resolve product by slug or ID
	var product model.Product
	pQuery := l.svcCtx.DB.WithContext(l.ctx).Where("slug = ?", productSlug)
	if pUUID, err := uuid.Parse(productSlug); err == nil {
		pQuery = l.svcCtx.DB.WithContext(l.ctx).Where("slug = ? OR id = ?", productSlug, pUUID)
	}
	pErr := pQuery.First(&product).Error
	if pErr != nil {
		return nil, xerr.NewErrMsg("محصول مورد نظر یافت نشد")
	}

	// 2. Fetch all approved reviews for this product to compute summary
	var allApproved []model.Review
	if err := l.svcCtx.DB.WithContext(l.ctx).
		Where("product_id = ? AND status = ?", product.ID, model.ReviewStatusApproved).
		Find(&allApproved).Error; err != nil {
		return nil, xerr.NewErrCode(xerr.DatabaseError)
	}

	totalApproved := int64(len(allApproved))
	summary := types.ProductReviewsSummary{
		TotalReviews:    totalApproved,
		TopTags:         []types.ExperienceTagSummary{},
		RatingBreakdown: types.RatingBreakdown{},
	}

	if totalApproved > 0 {
		var sumRating int64
		var positiveCount int64
		tagCounts := make(map[string]int64)

		for _, rev := range allApproved {
			sumRating += int64(rev.Rating)
			if rev.Rating >= 4 {
				positiveCount++
			}

			// Breakdown counts
			switch rev.Rating {
			case 5:
				summary.RatingBreakdown.Star5++
			case 4:
				summary.RatingBreakdown.Star4++
			case 3:
				summary.RatingBreakdown.Star3++
			case 2:
				summary.RatingBreakdown.Star2++
			case 1:
				summary.RatingBreakdown.Star1++
			}

			if rev.IsVerifiedBuyer {
				summary.VerifiedBuyersCount++
			}

			// Tag frequency
			tags := rev.GetTagsSlice()
			for _, t := range tags {
				trimmed := strings.TrimSpace(t)
				if trimmed != "" {
					tagCounts[trimmed]++
				}
			}
		}

		// Calculate percentages
		summary.RatingBreakdown.Star5Percent = int(math.Round(float64(summary.RatingBreakdown.Star5) / float64(totalApproved) * 100))
		summary.RatingBreakdown.Star4Percent = int(math.Round(float64(summary.RatingBreakdown.Star4) / float64(totalApproved) * 100))
		summary.RatingBreakdown.Star3Percent = int(math.Round(float64(summary.RatingBreakdown.Star3) / float64(totalApproved) * 100))
		summary.RatingBreakdown.Star2Percent = int(math.Round(float64(summary.RatingBreakdown.Star2) / float64(totalApproved) * 100))
		summary.RatingBreakdown.Star1Percent = int(math.Round(float64(summary.RatingBreakdown.Star1) / float64(totalApproved) * 100))

		avg := float64(sumRating) / float64(totalApproved)
		summary.AverageRating = math.Round(avg*10) / 10
		summary.SatisfactionRate = int(math.Round(float64(positiveCount) / float64(totalApproved) * 100))

		// Sort top tags by frequency
		type tagKV struct {
			Tag   string
			Count int64
		}
		var tagList []tagKV
		for k, v := range tagCounts {
			tagList = append(tagList, tagKV{Tag: k, Count: v})
		}
		sort.Slice(tagList, func(i, j int) bool {
			return tagList[i].Count > tagList[j].Count
		})

		for _, item := range tagList {
			percent := int(math.Round(float64(item.Count) / float64(totalApproved) * 100))
			summary.TopTags = append(summary.TopTags, types.ExperienceTagSummary{
				Tag:     item.Tag,
				Count:   item.Count,
				Percent: percent,
			})
		}
	} else {
		summary.AverageRating = 5.0
		summary.SatisfactionRate = 100
	}

	// 3. Query filtered & paginated reviews
	query := l.svcCtx.DB.WithContext(l.ctx).
		Where("product_id = ? AND status = ?", product.ID, model.ReviewStatusApproved)

	if ratingFilter >= 1 && ratingFilter <= 5 {
		query = query.Where("rating = ?", ratingFilter)
	}
	if onlyWithComments {
		query = query.Where("comment IS NOT NULL AND comment != ''")
	}

	var filteredCount int64
	query.Model(&model.Review{}).Count(&filteredCount)

	// Order by featured first, then chosen sort
	switch sortOrder {
	case "highest":
		query = query.Order("is_featured desc, rating desc, created_at desc")
	case "lowest":
		query = query.Order("is_featured desc, rating asc, created_at desc")
	case "helpful":
		query = query.Order("is_featured desc, helpful_count desc, created_at desc")
	default:
		// newest
		query = query.Order("is_featured desc, created_at desc")
	}

	offset := (page - 1) * pageSize
	var pagedReviews []model.Review
	if err := query.Offset(offset).Limit(pageSize).Find(&pagedReviews).Error; err != nil {
		return nil, xerr.NewErrCode(xerr.DatabaseError)
	}

	var dtos []types.ReviewItemDTO
	for _, r := range pagedReviews {
		dtos = append(dtos, types.ReviewItemDTO{
			ID:               r.ID.String(),
			ProductID:        r.ProductID.String(),
			ProductSlug:      r.ProductSlug,
			ProductTitle:     r.ProductTitle,
			OrderID:          safeUUIDString(r.OrderID),
			OrderNumber:      r.OrderNumber,
			CustomerName:     formatCustomerName(r.CustomerName, r.CustomerPhone),
			CustomerPhone:    maskPhone(r.CustomerPhone),
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
			CreatedAt:        r.CreatedAt,
		})
	}

	return &types.ProductReviewsResponse{
		Summary: summary,
		Reviews: dtos,
		Total:   filteredCount,
	}, nil
}

// SubmitReview handles submission by verified buyers only, with moderation for text comments
func (l *ReviewLogic) SubmitReview(req *types.SubmitReviewRequest, customerIDStr, customerMobile, clientIP string) (*types.SubmitReviewResponse, error) {
	if req.Rating < 1 || req.Rating > 5 {
		return nil, xerr.NewErrMsg("امتیاز باید بین ۱ تا ۵ ستاره باشد")
	}

	// 1. Resolve Product
	var product model.Product
	var pErr error
	if req.ProductID != "" {
		pID, err := uuid.Parse(req.ProductID)
		if err == nil {
			pErr = l.svcCtx.DB.WithContext(l.ctx).Where("id = ?", pID).First(&product).Error
		}
	}
	if product.ID == uuid.Nil && req.ProductSlug != "" {
		pErr = l.svcCtx.DB.WithContext(l.ctx).Where("slug = ?", req.ProductSlug).First(&product).Error
	}
	if pErr != nil || product.ID == uuid.Nil {
		return nil, xerr.NewErrMsg("محصول مورد نظر یافت نشد")
	}

	// 2. Strict Verified Buyer Check (خریدار قطعی)
	// Check if this customer (or provided order credentials) has a completed/paid order containing this product
	var validOrder *model.Order
	var orderItem *model.OrderItem

	// Strategy A: If OrderID / OrderNumber + AccessToken provided
	if req.OrderAccessToken != "" && (req.OrderID != "" || req.OrderNumber != "") {
		var o model.Order
		q := l.svcCtx.DB.WithContext(l.ctx).Where("access_token = ?", req.OrderAccessToken)
		if req.OrderID != "" {
			if parsedOID, err := uuid.Parse(req.OrderID); err == nil {
				q = q.Where("id = ?", parsedOID)
			}
		} else {
			q = q.Where("order_number = ?", req.OrderNumber)
		}
		if err := q.First(&o).Error; err == nil {
			if o.Status == model.OrderCompleted || o.Status == model.OrderPaid {
				var oi model.OrderItem
				if err := l.svcCtx.DB.WithContext(l.ctx).
					Where("order_id = ? AND product_id = ?", o.ID, product.ID).
					First(&oi).Error; err == nil {
					validOrder = &o
					orderItem = &oi
				}
			}
		}
	}

	// Strategy B: If customer is logged in (CustomerID or CustomerMobile)
	if validOrder == nil && (customerIDStr != "" || customerMobile != "") {
		var orders []model.Order
		q := l.svcCtx.DB.WithContext(l.ctx).Where("status IN (?, ?)", model.OrderCompleted, model.OrderPaid)
		if customerIDStr != "" {
			cUUID, err := uuid.Parse(customerIDStr)
			if err == nil {
				q = q.Where("customer_id = ? OR customer_phone = ?", cUUID, customerMobile)
			} else {
				q = q.Where("customer_phone = ?", customerMobile)
			}
		} else {
			q = q.Where("customer_phone = ?", customerMobile)
		}

		if err := q.Order("created_at desc").Find(&orders).Error; err == nil {
			for _, ord := range orders {
				var oi model.OrderItem
				if err := l.svcCtx.DB.WithContext(l.ctx).
					Where("order_id = ? AND product_id = ?", ord.ID, product.ID).
					First(&oi).Error; err == nil {
					validOrder = &ord
					orderItem = &oi
					break
				}
			}
		}
	}

	// Strategy C: Check by CustomerPhone if provided in request
	if validOrder == nil && req.CustomerPhone != "" {
		var ord model.Order
		if err := l.svcCtx.DB.WithContext(l.ctx).
			Joins("JOIN order_items ON order_items.order_id = orders.id").
			Where("orders.customer_phone = ? AND order_items.product_id = ? AND orders.status IN (?, ?)",
				req.CustomerPhone, product.ID, model.OrderCompleted, model.OrderPaid).
			Order("orders.created_at desc").
			First(&ord).Error; err == nil {
			var oi model.OrderItem
			_ = l.svcCtx.DB.WithContext(l.ctx).Where("order_id = ? AND product_id = ?", ord.ID, product.ID).First(&oi).Error
			validOrder = &ord
			orderItem = &oi
		}
	}

	// If still not found, deny submission!
	if validOrder == nil {
		return nil, xerr.NewErrMsg("ثبت نظر و امتیاز تنها برای خریداران قطعی این محصول مجاز است. شما هنوز خریدی از این محصول انجام نداده‌اید یا وضعیت سفارش تکمیل نشده است.")
	}

	// 3. Resolve Customer Name and Phone
	finalPhone := validOrder.CustomerPhone
	if finalPhone == "" {
		finalPhone = customerMobile
	}
	if finalPhone == "" {
		finalPhone = req.CustomerPhone
	}

	finalName := req.CustomerName
	if finalName == "" {
		// Look up in customer table
		var c model.Customer
		if l.svcCtx.DB.WithContext(l.ctx).Where("phone = ?", finalPhone).First(&c).Error == nil {
			finalName = c.FullName
		}
	}
	if finalName == "" {
		finalName = formatCustomerName("", finalPhone)
	}

	// 4. Serialize Experience Tags
	cleanTags := []string{}
	for _, t := range req.ExperienceTags {
		tr := strings.TrimSpace(t)
		if tr != "" {
			cleanTags = append(cleanTags, tr)
		}
	}
	tagsJSON, _ := json.Marshal(cleanTags)

	// 5. Moderation Rule:
	// "طوری باشه که اگر کامنت متن میخواد بنویسه اول از سمت ادمین تایید بشه بعد کامنت نمایش داده بشه در کنارش هم کاربر بتونه یکی از تگ های تحویل سریع و.... از اینا انتخاب کنه"
	hasTextComment := len(strings.TrimSpace(req.Comment)) > 0
	var reviewStatus model.ReviewStatus
	var requiresApproval bool
	var responseMsg string

	if hasTextComment {
		reviewStatus = model.ReviewStatusPending
		requiresApproval = true
		responseMsg = "نظر و امتیاز ارزشمند شما با موفقیت ثبت شد و پس از بررسی تیم پشتیبانی در سایت نمایش داده خواهد شد."
	} else {
		// Rating & Tags without textual comments are automatically approved into social proof
		reviewStatus = model.ReviewStatusApproved
		requiresApproval = false
		responseMsg = "امتیاز و برچسب‌های تجربه خرید شما با موفقیت ثبت و اعمال شد. از بازخورد شما متشکریم!"
	}

	var customerUUID *uuid.UUID
	if validOrder.CustomerID != nil {
		customerUUID = validOrder.CustomerID
	} else if customerIDStr != "" {
		if cid, err := uuid.Parse(customerIDStr); err == nil {
			customerUUID = &cid
		}
	}

	// 6. Check if user already reviewed this order or product
	var existing model.Review
	checkQuery := l.svcCtx.DB.WithContext(l.ctx).
		Where("product_id = ? AND (order_id = ? OR (customer_phone != '' AND customer_phone = ?))",
			product.ID, validOrder.ID, finalPhone)

	if err := checkQuery.First(&existing).Error; err == nil {
		// Update existing review
		existing.Rating = req.Rating
		existing.Comment = strings.TrimSpace(req.Comment)
		existing.ExperienceTags = string(tagsJSON)
		existing.Status = reviewStatus
		existing.CustomerName = finalName
		existing.CustomerPhone = finalPhone
		existing.OrderID = &validOrder.ID
		existing.OrderNumber = validOrder.OrderNumber
		existing.IsVerifiedBuyer = true

		if err := l.svcCtx.DB.WithContext(l.ctx).Save(&existing).Error; err != nil {
			return nil, xerr.NewErrCode(xerr.DatabaseError)
		}

		return &types.SubmitReviewResponse{
			ReviewID:         existing.ID.String(),
			Status:           string(existing.Status),
			Message:          "نظر و امتیاز شما برای این محصول با موفقیت به‌روزرسانی شد.",
			RequiresApproval: requiresApproval,
		}, nil
	}

	// 7. Create New Review
	newReview := model.Review{
		ID:              uuid.New(),
		ProductID:       product.ID,
		ProductSlug:     product.Slug,
		ProductTitle:    product.TitleFA,
		OrderID:         &validOrder.ID,
		OrderNumber:     validOrder.OrderNumber,
		CustomerID:      customerUUID,
		CustomerPhone:   finalPhone,
		CustomerName:    finalName,
		Rating:          req.Rating,
		Comment:         strings.TrimSpace(req.Comment),
		ExperienceTags:  string(tagsJSON),
		Status:          reviewStatus,
		IsVerifiedBuyer: true,
		IsFeatured:      false,
		HelpfulCount:    0,
	}

	if err := l.svcCtx.DB.WithContext(l.ctx).Create(&newReview).Error; err != nil {
		return nil, xerr.NewErrCode(xerr.DatabaseError)
	}

	_ = orderItem // verified

	return &types.SubmitReviewResponse{
		ReviewID:         newReview.ID.String(),
		Status:           string(newReview.Status),
		Message:          responseMsg,
		RequiresApproval: requiresApproval,
	}, nil
}

// VoteHelpful records a vote for a review and increments counter
func (l *ReviewLogic) VoteHelpful(reviewIDStr string, voterKey string) (int, error) {
	revUUID, err := uuid.Parse(reviewIDStr)
	if err != nil {
		return 0, xerr.NewErrMsg("شناسه نظر نامعتبر است")
	}

	var rev model.Review
	if err := l.svcCtx.DB.WithContext(l.ctx).Where("id = ? AND status = ?", revUUID, model.ReviewStatusApproved).First(&rev).Error; err != nil {
		return 0, xerr.NewErrMsg("نظر مورد نظر یافت نشد")
	}

	if voterKey == "" {
		voterKey = "anon"
	}

	// Check if already voted -> Toggle (Cancel if already voted)
	var existing model.ReviewHelpfulVote
	if err := l.svcCtx.DB.WithContext(l.ctx).
		Where("review_id = ? AND voter_key = ?", revUUID, voterKey).
		First(&existing).Error; err == nil {
		l.svcCtx.DB.WithContext(l.ctx).Delete(&existing)
		newCount := rev.HelpfulCount - 1
		if newCount < 0 {
			newCount = 0
		}
		l.svcCtx.DB.WithContext(l.ctx).Model(&rev).Update("helpful_count", newCount)
		return newCount, nil
	}

	vote := model.ReviewHelpfulVote{
		ID:        uuid.New(),
		ReviewID:  revUUID,
		VoterKey:  voterKey,
		CreatedAt: time.Now(),
	}

	if err := l.svcCtx.DB.WithContext(l.ctx).Create(&vote).Error; err != nil {
		return rev.HelpfulCount, err
	}

	l.svcCtx.DB.WithContext(l.ctx).Model(&rev).Update("helpful_count", gorm.Expr("helpful_count + 1"))
	return rev.HelpfulCount + 1, nil
}

// CheckEligibility checks if customer is a verified buyer and whether they already reviewed
func (l *ReviewLogic) CheckEligibility(productSlug string, customerIDStr, customerMobile string) (*types.ReviewEligibilityResponse, error) {
	var product model.Product
	pQuery := l.svcCtx.DB.WithContext(l.ctx).Where("slug = ?", productSlug)
	if pUUID, err := uuid.Parse(productSlug); err == nil {
		pQuery = l.svcCtx.DB.WithContext(l.ctx).Where("slug = ? OR id = ?", productSlug, pUUID)
	}
	if err := pQuery.First(&product).Error; err != nil {
		return &types.ReviewEligibilityResponse{Eligible: false, Reason: "محصول یافت نشد"}, nil
	}

	if customerIDStr == "" && customerMobile == "" {
		return &types.ReviewEligibilityResponse{
			Eligible:     false,
			Reason:       "جهت ثبت نظر ابتدا وارد حساب کاربری خود شوید یا اطلاعات سفارش خود را وارد کنید.",
			ProductTitle: product.TitleFA,
		}, nil
	}

	// Find completed/paid order for this product
	var orders []model.Order
	q := l.svcCtx.DB.WithContext(l.ctx).Where("status IN (?, ?)", model.OrderCompleted, model.OrderPaid)
	if customerIDStr != "" {
		if cUUID, err := uuid.Parse(customerIDStr); err == nil {
			q = q.Where("customer_id = ? OR customer_phone = ?", cUUID, customerMobile)
		} else {
			q = q.Where("customer_phone = ?", customerMobile)
		}
	} else {
		q = q.Where("customer_phone = ?", customerMobile)
	}

	_ = q.Order("created_at desc").Find(&orders).Error

	var matchedOrder *model.Order
	for _, o := range orders {
		var oi model.OrderItem
		if l.svcCtx.DB.WithContext(l.ctx).Where("order_id = ? AND product_id = ?", o.ID, product.ID).First(&oi).Error == nil {
			matchedOrder = &o
			break
		}
	}

	if matchedOrder == nil {
		return &types.ReviewEligibilityResponse{
			Eligible:     false,
			Reason:       "تنها خریداران قطعی این محصول پس از تکمیل سفارش امکان ثبت نظر و امتیاز دارند.",
			ProductTitle: product.TitleFA,
		}, nil
	}

	// Check if existing review
	var existingReview model.Review
	var existingDTO *types.ReviewItemDTO
	if err := l.svcCtx.DB.WithContext(l.ctx).
		Where("product_id = ? AND (order_id = ? OR customer_phone = ?)", product.ID, matchedOrder.ID, matchedOrder.CustomerPhone).
		First(&existingReview).Error; err == nil {
		existingDTO = &types.ReviewItemDTO{
			ID:               existingReview.ID.String(),
			ProductID:        existingReview.ProductID.String(),
			ProductSlug:      existingReview.ProductSlug,
			ProductTitle:     existingReview.ProductTitle,
			OrderID:          safeUUIDString(existingReview.OrderID),
			OrderNumber:      existingReview.OrderNumber,
			CustomerName:     existingReview.CustomerName,
			Rating:           existingReview.Rating,
			Comment:          existingReview.Comment,
			ExperienceTags:   existingReview.GetTagsSlice(),
			Status:           string(existingReview.Status),
			IsVerifiedBuyer:  existingReview.IsVerifiedBuyer,
			IsFeatured:       existingReview.IsFeatured,
			HelpfulCount:     existingReview.HelpfulCount,
			AdminReply:       existingReview.AdminReply,
			AdminReplyAuthor: existingReview.AdminReplyAuthor,
			AdminRepliedAt:   existingReview.AdminRepliedAt,
			CreatedAt:        existingReview.CreatedAt,
		}
	}

	return &types.ReviewEligibilityResponse{
		Eligible:       true,
		OrderID:        matchedOrder.ID.String(),
		OrderNumber:    matchedOrder.OrderNumber,
		ProductTitle:   product.TitleFA,
		ExistingReview: existingDTO,
	}, nil
}

// GetCustomerReviews returns all reviews submitted by the authenticated customer
func (l *ReviewLogic) GetCustomerReviews(customerIDStr, customerMobile string) ([]types.ReviewItemDTO, error) {
	query := l.svcCtx.DB.WithContext(l.ctx).Model(&model.Review{})
	if customerIDStr != "" {
		if cUUID, err := uuid.Parse(customerIDStr); err == nil {
			query = query.Where("customer_id = ? OR customer_phone = ?", cUUID, customerMobile)
		} else {
			query = query.Where("customer_phone = ?", customerMobile)
		}
	} else if customerMobile != "" {
		query = query.Where("customer_phone = ?", customerMobile)
	} else {
		return []types.ReviewItemDTO{}, nil
	}

	var reviews []model.Review
	if err := query.Order("created_at desc").Find(&reviews).Error; err != nil {
		return nil, xerr.NewErrCode(xerr.DatabaseError)
	}

	var dtos []types.ReviewItemDTO
	for _, r := range reviews {
		dtos = append(dtos, types.ReviewItemDTO{
			ID:               r.ID.String(),
			ProductID:        r.ProductID.String(),
			ProductSlug:      r.ProductSlug,
			ProductTitle:     r.ProductTitle,
			OrderID:          safeUUIDString(r.OrderID),
			OrderNumber:      r.OrderNumber,
			CustomerName:     r.CustomerName,
			CustomerPhone:    maskPhone(r.CustomerPhone),
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
	return dtos, nil
}

func safeUUIDString(u *uuid.UUID) string {
	if u == nil {
		return ""
	}
	return u.String()
}
