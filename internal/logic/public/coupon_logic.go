package public

import (
	"context"
	"strings"
	"time"

	"github.com/EbadiDev/anahix-server/internal/model"
	"github.com/EbadiDev/anahix-server/internal/svc"
	"github.com/EbadiDev/anahix-server/internal/types"
	"github.com/EbadiDev/anahix-server/pkg/xerr"
)

type CouponLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCouponLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CouponLogic {
	return &CouponLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CouponLogic) ValidateCoupon(req *types.ValidateCouponRequest) (*types.ValidateCouponResponse, error) {
	code := strings.ToUpper(strings.TrimSpace(req.Code))
	if code == "" {
		return nil, xerr.NewErrMsg("کد تخفیف معتبر نیست")
	}

	var coupon model.Coupon
	err := l.svcCtx.DB.WithContext(l.ctx).Where("code = ? AND is_active = ?", code, true).First(&coupon).Error
	if err != nil {
		return &types.ValidateCouponResponse{
			Valid:   false,
			Code:    code,
			Message: "کد تخفیف وارد شده نامعتبر یا منقضی شده است",
		}, nil
	}

	// 1. Expiration check
	now := time.Now()
	if coupon.ExpiresAt != nil && now.After(*coupon.ExpiresAt) {
		return &types.ValidateCouponResponse{
			Valid:   false,
			Code:    code,
			Message: "مهلت استفاده از این کد تخفیف به پایان رسیده است",
		}, nil
	}

	// 2. Global Max Usage check
	if coupon.MaxUsage > 0 && coupon.CurrentUsage >= coupon.MaxUsage {
		return &types.ValidateCouponResponse{
			Valid:   false,
			Code:    code,
			Message: "سقف استفاده سراسری از این کد تخفیف تکمیل شده است",
		}, nil
	}

	// 3. Minimum Order Amount check
	if coupon.MinOrderAmountToman > 0 && req.SubtotalToman < coupon.MinOrderAmountToman {
		return &types.ValidateCouponResponse{
			Valid:   false,
			Code:    code,
			Message: "حداقل مبلغ سفارش برای استفاده از این کوپن رعایت نشده است",
		}, nil
	}

	// 4. Category check (if coupon is restricted to a category)
	if coupon.Category != "" && coupon.Category != "all" && req.Category != "" && req.Category != coupon.Category {
		return &types.ValidateCouponResponse{
			Valid:   false,
			Code:    code,
			Message: "این کد تخفیف برای دسته‌بندی محصول انتخابی قابل استفاده نیست",
		}, nil
	}

	// 5. Per-User Usage check
	if req.CustomerPhone != "" && coupon.PerUserLimit > 0 {
		var userUsageCount int64
		l.svcCtx.DB.WithContext(l.ctx).Model(&model.CouponUsage{}).
			Where("coupon_id = ? AND customer_phone = ?", coupon.ID, req.CustomerPhone).
			Count(&userUsageCount)
		if int(userUsageCount) >= coupon.PerUserLimit {
			return &types.ValidateCouponResponse{
				Valid:   false,
				Code:    code,
				Message: "شما قبلاً از این کد تخفیف استفاده کرده‌اید",
			}, nil
		}
	}

	// Calculate discount
	var discountToman int64
	if coupon.DiscountType == model.DiscountTypePercentage {
		discountToman = (req.SubtotalToman * coupon.DiscountValue) / 100
		if coupon.MaxDiscountToman > 0 && discountToman > coupon.MaxDiscountToman {
			discountToman = coupon.MaxDiscountToman
		}
	} else {
		discountToman = coupon.DiscountValue
	}

	if discountToman > req.SubtotalToman {
		discountToman = req.SubtotalToman
	}

	finalAmount := req.SubtotalToman - discountToman

	return &types.ValidateCouponResponse{
		Valid:            true,
		Code:             coupon.Code,
		DiscountType:     string(coupon.DiscountType),
		DiscountValue:    coupon.DiscountValue,
		DiscountToman:    discountToman,
		FinalAmountToman: finalAmount,
		Message:          "کد تخفیف با موفقیت اعمال شد",
	}, nil
}
