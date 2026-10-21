package admin

import (
	"context"
	"fmt"
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
		label := c.Label
		if label == "" {
			if c.DiscountType == model.DiscountTypePercentage {
				label = fmt.Sprintf("%d%% تخفیف", c.DiscountValue)
			} else {
				if c.DiscountValue >= 1000 {
					label = fmt.Sprintf("%d هزار تومان", c.DiscountValue/1000)
				} else {
					label = fmt.Sprintf("%d تومان", c.DiscountValue)
				}
			}
		}

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
			Label:               label,
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

	label := strings.TrimSpace(req.Label)
	if label == "" {
		if req.DiscountType == "percentage" {
			label = fmt.Sprintf("%d%% تخفیف", req.DiscountValue)
		} else {
			if req.DiscountValue >= 1000 {
				label = fmt.Sprintf("%d هزار تومان", req.DiscountValue/1000)
			} else {
				label = fmt.Sprintf("%d تومان", req.DiscountValue)
			}
		}
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
		Label:               label,
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
		Label:               coupon.Label,
		ExpiresAt:           coupon.ExpiresAt,
		IsActive:            coupon.IsActive,
		CreatedAt:           coupon.CreatedAt,
	}, nil
}

func (l *CouponAdminLogic) DeleteCoupon(id string) error {
	couponID, err := uuid.Parse(id)
	if err != nil {
		return xerr.NewErrMsg("شناسه کوپن نامعتبر است")
	}

	if err := l.svcCtx.DB.WithContext(l.ctx).Where("id = ?", couponID).Delete(&model.Coupon{}).Error; err != nil {
		return xerr.NewErrCode(xerr.DatabaseError)
	}
	return nil
}

func (l *CouponAdminLogic) ToggleCoupon(id string) (*types.CouponListItem, error) {
	couponID, err := uuid.Parse(id)
	if err != nil {
		return nil, xerr.NewErrMsg("شناسه کوپن نامعتبر است")
	}

	var coupon model.Coupon
	if err := l.svcCtx.DB.WithContext(l.ctx).First(&coupon, "id = ?", couponID).Error; err != nil {
		return nil, xerr.NewErrMsg("کوپن یافت نشد")
	}

	coupon.IsActive = !coupon.IsActive
	if err := l.svcCtx.DB.WithContext(l.ctx).Save(&coupon).Error; err != nil {
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
		CurrentUsage:        coupon.CurrentUsage,
		Category:            coupon.Category,
		Label:               coupon.Label,
		ExpiresAt:           coupon.ExpiresAt,
		IsActive:            coupon.IsActive,
		CreatedAt:           coupon.CreatedAt,
	}, nil
}
