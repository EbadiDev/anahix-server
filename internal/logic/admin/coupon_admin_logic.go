package admin

import (
	"context"
	"strings"

	"github.com/EbadiDev/anahix-server/internal/model"
	"github.com/EbadiDev/anahix-server/internal/svc"
	"github.com/EbadiDev/anahix-server/internal/types"
	"github.com/EbadiDev/anahix-server/pkg/xerr"
	"github.com/google/uuid"
)

type CouponAdminLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCouponAdminLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CouponAdminLogic {
	return &CouponAdminLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CouponAdminLogic) ListCoupons() ([]types.CouponListItem, error) {
	var coupons []model.Coupon
	if err := l.svcCtx.DB.WithContext(l.ctx).Order("created_at DESC").Find(&coupons).Error; err != nil {
		return nil, xerr.NewErrCode(xerr.DatabaseError)
	}

	result := make([]types.CouponListItem, 0, len(coupons))
	for _, c := range coupons {
		result = append(result, types.CouponListItem{
			ID:                  c.ID.String(),
			Code:                c.Code,
			DiscountType:        string(c.DiscountType),
			DiscountValue:       c.DiscountValue,
			MinOrderAmountToman: c.MinOrderAmountToman,
			MaxDiscountToman:    c.MaxDiscountToman,
			PerUserLimit:        c.PerUserLimit,
			MaxUsage:            c.MaxUsage,
			CurrentUsage:        c.CurrentUsage,
			Category:            c.Category,
			ExpiresAt:           c.ExpiresAt,
			IsActive:            c.IsActive,
			CreatedAt:           c.CreatedAt,
		})
	}

	return result, nil
}

func (l *CouponAdminLogic) CreateCoupon(req *types.CreateCouponRequest) (*types.CouponListItem, error) {
	code := strings.ToUpper(strings.TrimSpace(req.Code))
	if code == "" {
		return nil, xerr.NewErrMsg("کد تخفیف نمی‌تواند خالی باشد")
	}

	// Check if already exists
	var count int64
	l.svcCtx.DB.WithContext(l.ctx).Model(&model.Coupon{}).Where("code = ?", code).Count(&count)
	if count > 0 {
		return nil, xerr.NewErrMsg("کد تخفیف با این نام از قبل وجود دارد")
	}

	category := req.Category
	if category == "" {
		category = "all"
	}

	coupon := model.Coupon{
		ID:                  uuid.New(),
		Code:                code,
		DiscountType:        model.DiscountType(req.DiscountType),
		DiscountValue:       req.DiscountValue,
		MinOrderAmountToman: req.MinOrderAmountToman,
		MaxDiscountToman:    req.MaxDiscountToman,
		PerUserLimit:        req.PerUserLimit,
		MaxUsage:            req.MaxUsage,
		Category:            category,
		ExpiresAt:           req.ExpiresAt,
		IsActive:            req.IsActive,
	}

	if err := l.svcCtx.DB.WithContext(l.ctx).Create(&coupon).Error; err != nil {
		return nil, xerr.NewErrCode(xerr.DatabaseError)
	}

	return &types.CouponListItem{
		ID:                  coupon.ID.String(),
		Code:                coupon.Code,
		DiscountType:        string(coupon.DiscountType),
		DiscountValue:       coupon.DiscountValue,
		MinOrderAmountToman: coupon.MinOrderAmountToman,
		MaxDiscountToman:    coupon.MaxDiscountToman,
		PerUserLimit:        coupon.PerUserLimit,
		MaxUsage:            coupon.MaxUsage,
		CurrentUsage:        0,
		Category:            coupon.Category,
		ExpiresAt:           coupon.ExpiresAt,
		IsActive:            coupon.IsActive,
		CreatedAt:           coupon.CreatedAt,
	}, nil
}
