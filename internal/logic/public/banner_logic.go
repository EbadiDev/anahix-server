package public

import (
	"context"

	"github.com/EbadiDev/anahix-server/internal/model"
	"github.com/EbadiDev/anahix-server/internal/svc"
	"github.com/EbadiDev/anahix-server/internal/types"
)

type BannerLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewBannerLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BannerLogic {
	return &BannerLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BannerLogic) ListPublicBanners(position string) ([]types.BannerItemResponse, error) {
	query := l.svcCtx.DB.WithContext(l.ctx).Model(&model.Banner{}).Where("is_active = ?", true)
	if position != "" {
		query = query.Where("position = ?", position)
	}

	var banners []model.Banner
	if err := query.Order("sort_order asc, created_at desc").Find(&banners).Error; err != nil {
		return nil, err
	}

	resp := make([]types.BannerItemResponse, len(banners))
	for i, b := range banners {
		resp[i] = types.BannerItemResponse{
			ID:         b.ID.String(),
			Title:      b.Title,
			Subtitle:   b.Subtitle,
			BadgeText:  b.BadgeText,
			ImageURL:   b.ImageURL,
			LinkURL:    b.LinkURL,
			Position:   string(b.Position),
			BgGradient: b.BgGradient,
			ButtonText: b.ButtonText,
			SortOrder:  b.SortOrder,
			IsActive:   b.IsActive,
		}
	}

	return resp, nil
}
