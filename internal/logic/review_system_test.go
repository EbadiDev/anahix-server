package logic

import (
	"context"
	"testing"
	"time"

	"github.com/EbadiDev/anahix-server/internal/logic/admin"
	"github.com/EbadiDev/anahix-server/internal/logic/public"
	"github.com/EbadiDev/anahix-server/internal/model"
	"github.com/EbadiDev/anahix-server/internal/svc"
	"github.com/EbadiDev/anahix-server/internal/types"
	"github.com/EbadiDev/anahix-server/pkg/crypto"
	"github.com/EbadiDev/anahix-server/pkg/orm"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func setupTestReviewSystem(t *testing.T) (*svc.ServiceContext, model.Customer, model.Product) {
	db, err := orm.InitTestDB()
	assert.NoError(t, err)

	vault, err := crypto.NewVault("test-master-key-32bytes-secret!")
	assert.NoError(t, err)

	svcCtx := &svc.ServiceContext{
		DB:    db,
		Vault: vault,
	}

	uid := uuid.New().String()[:6]
	cust := model.Customer{
		ID:           uuid.New(),
		Phone:        "0912" + uid,
		FullName:     "خریدار تستی آناهیکس",
		BalanceToman: 5000000,
		IsActive:     true,
	}
	assert.NoError(t, db.Create(&cust).Error)

	prod := model.Product{
		ID:               uuid.New(),
		Title:            "Gemini Pro Review Test " + uid,
		TitleFA:          "گوگل جمینای تست " + uid,
		Slug:             "gemini-review-test-" + uid,
		Category:         model.CategoryAILink,
		DeliveryType:     model.DeliveryInstant,
		DeliveryStrategy: "manual_operator",
		PriceToman:       1480000,
		IsActive:         true,
	}
	assert.NoError(t, db.Create(&prod).Error)

	return svcCtx, cust, prod
}

func TestUnverifiedBuyerCannotReview(t *testing.T) {
	svcCtx, cust, prod := setupTestReviewSystem(t)
	ctx := context.Background()

	reviewLogic := public.NewReviewLogic(ctx, svcCtx)

	// User attempts to review without having bought the product
	submitReq := &types.SubmitReviewRequest{
		ProductSlug:    prod.Slug,
		Rating:         5,
		Comment:        "تلاش برای ثبت نظر بدون خرید!",
		ExperienceTags: []string{"تحویل زیر ۱ دقیقه"},
	}

	_, err := reviewLogic.SubmitReview(submitReq, cust.ID.String(), cust.Phone, "127.0.0.1")
	assert.Error(t, err, "User without verified purchase MUST NOT be allowed to review")
	assert.Contains(t, err.Error(), "خریداران قطعی")
}

func TestVerifiedBuyerReviewAndModerationWorkflow(t *testing.T) {
	svcCtx, cust, prod := setupTestReviewSystem(t)
	ctx := context.Background()

	// 1. Create a verified completed order for this customer & product
	orderID := uuid.New()
	completedOrder := model.Order{
		ID:                  orderID,
		OrderNumber:         "ORD-TEST-9988",
		CustomerID:          &cust.ID,
		CustomerPhone:       cust.Phone,
		CustomerEmail:       "test@anahix.com",
		AccessToken:         "tok-test-secure-123",
		Status:              model.OrderCompleted,
		DeliveryType:        model.DeliveryInstant,
		TotalAmountToman:    prod.PriceToman,
		SubtotalAmountToman: prod.PriceToman,
		PaymentMethod:       model.PaymentMethodBalance,
		ExpiresAt:           time.Now().Add(1 * time.Hour),
		Items: []model.OrderItem{
			{
				ID:             uuid.New(),
				OrderID:        orderID,
				ProductID:      prod.ID,
				ProductTitle:   prod.TitleFA,
				ProductSlug:    prod.Slug,
				DeliveryType:   prod.DeliveryType,
				UnitPriceToman: prod.PriceToman,
				Quantity:       1,
				SubtotalToman:  prod.PriceToman,
			},
		},
	}
	assert.NoError(t, svcCtx.DB.Create(&completedOrder).Error)

	reviewLogic := public.NewReviewLogic(ctx, svcCtx)
	adminLogic := admin.NewReviewAdminLogic(ctx, svcCtx)

	// 2. Check Eligibility -> MUST be eligible
	eligibility, err := reviewLogic.CheckEligibility(prod.Slug, cust.ID.String(), cust.Phone)
	assert.NoError(t, err)
	assert.True(t, eligibility.Eligible, "Verified buyer must be marked as eligible")
	assert.Equal(t, "ORD-TEST-9988", eligibility.OrderNumber)

	// 3. Customer submits review WITH text comment
	// Requirement: Text comments MUST be put in "pending" status waiting for admin approval
	submitReq := &types.SubmitReviewRequest{
		ProductSlug:    prod.Slug,
		OrderID:        orderID.String(),
		Rating:         5,
		Comment:        "تحویل بسیار عالی و قانونی، سرعت اکانت بی‌نظیره!",
		ExperienceTags: []string{"تحویل زیر ۱ دقیقه", "کیفیت بدون قطعی", "پشتیبانی عالی و صبور"},
	}

	submitResp, err := reviewLogic.SubmitReview(submitReq, cust.ID.String(), cust.Phone, "127.0.0.1")
	assert.NoError(t, err)
	assert.True(t, submitResp.RequiresApproval, "Reviews with text comment require admin approval")
	assert.Equal(t, "pending", submitResp.Status)

	// 4. Public storefront should NOT display the pending review yet
	pubResBefore, err := reviewLogic.GetProductReviews(prod.Slug, 0, "newest", false, 1, 20)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), pubResBefore.Summary.TotalReviews, "Pending review must NOT be publicly listed")
	assert.Len(t, pubResBefore.Reviews, 0)

	// 5. Admin lists reviews -> Finds the pending review
	adminList, err := adminLogic.ListReviews(&types.AdminReviewListRequest{
		Status: "pending",
	})
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, adminList.Total, int64(1))
	assert.Equal(t, submitResp.ReviewID, adminList.Items[0].ID)

	// 6. Admin approves the review
	err = adminLogic.UpdateReviewStatus(submitResp.ReviewID, "approved", "", "admin")
	assert.NoError(t, err)

	// 7. Public storefront NOW displays the approved review with Social Proof stats!
	pubResAfter, err := reviewLogic.GetProductReviews(prod.Slug, 0, "newest", false, 1, 20)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), pubResAfter.Summary.TotalReviews)
	assert.Equal(t, 5.0, pubResAfter.Summary.AverageRating)
	assert.Equal(t, 100, pubResAfter.Summary.SatisfactionRate)
	assert.Len(t, pubResAfter.Summary.TopTags, 3)
	assert.Len(t, pubResAfter.Reviews, 1)

	revItem := pubResAfter.Reviews[0]
	assert.True(t, revItem.IsVerifiedBuyer)
	assert.Equal(t, "ORD-TEST-9988", revItem.OrderNumber)
	assert.Equal(t, 5, revItem.Rating)
	assert.Contains(t, revItem.Comment, "تحویل بسیار عالی")

	// 8. Admin writes an official reply to the review
	err = adminLogic.ReplyReview(submitResp.ReviewID, "سلام و درود؛ از نظر پرمهر شما متشکریم.", "پشتیبانی رسمی آناهیکس", "admin")
	assert.NoError(t, err)

	// 9. Public storefront verifies admin reply is displayed
	pubResWithReply, err := reviewLogic.GetProductReviews(prod.Slug, 0, "newest", false, 1, 20)
	assert.NoError(t, err)
	assert.Equal(t, "سلام و درود؛ از نظر پرمهر شما متشکریم.", pubResWithReply.Reviews[0].AdminReply)
	assert.Equal(t, "پشتیبانی رسمی آناهیکس", pubResWithReply.Reviews[0].AdminReplyAuthor)

	// 10. Helpful upvoting
	newHelpful, err := reviewLogic.VoteHelpful(submitResp.ReviewID, "user-ip-1")
	assert.NoError(t, err)
	assert.Equal(t, 1, newHelpful)

	// Duplicate vote by same key does not increment
	dupHelpful, err := reviewLogic.VoteHelpful(submitResp.ReviewID, "user-ip-1")
	assert.NoError(t, err)
	assert.Equal(t, 1, dupHelpful)
}
