package admin

import (
	"context"
	"fmt"

	"github.com/EbadiDev/anahix-server/internal/model"
	"github.com/EbadiDev/anahix-server/internal/svc"
	"github.com/EbadiDev/anahix-server/internal/types"
	"github.com/google/uuid"
)

type BannerAdminLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewBannerAdminLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BannerAdminLogic {
	return &BannerAdminLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BannerAdminLogic) ListBanners() ([]types.BannerItemResponse, error) {
	var banners []model.Banner
	if err := l.svcCtx.DB.WithContext(l.ctx).Order("sort_order asc, created_at desc").Find(&banners).Error; err != nil {
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

func (l *BannerAdminLogic) CreateBanner(req *types.CreateBannerRequest) (*types.BannerItemResponse, error) {
	pos := model.BannerPosition(req.Position)
	if pos == "" {
		pos = model.BannerPositionMainSlider
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	bgGradient := req.BgGradient
	if bgGradient == "" {
		bgGradient = "from-blue-600 via-indigo-600 to-purple-700"
	}

	btnText := req.ButtonText
	if btnText == "" {
		btnText = "مشاهده و خرید"
	}

	banner := model.Banner{
		ID:         uuid.New(),
		Title:      req.Title,
		Subtitle:   req.Subtitle,
		BadgeText:  req.BadgeText,
		ImageURL:   req.ImageURL,
		LinkURL:    req.LinkURL,
		Position:   pos,
		BgGradient: bgGradient,
		ButtonText: btnText,
		SortOrder:  req.SortOrder,
		IsActive:   isActive,
	}

	if err := l.svcCtx.DB.WithContext(l.ctx).Create(&banner).Error; err != nil {
		return nil, err
	}

	return &types.BannerItemResponse{
		ID:         banner.ID.String(),
		Title:      banner.Title,
		Subtitle:   banner.Subtitle,
		BadgeText:  banner.BadgeText,
		ImageURL:   banner.ImageURL,
		LinkURL:    banner.LinkURL,
		Position:   string(banner.Position),
		BgGradient: banner.BgGradient,
		ButtonText: banner.ButtonText,
		SortOrder:  banner.SortOrder,
		IsActive:   banner.IsActive,
	}, nil
}

func (l *BannerAdminLogic) UpdateBanner(idStr string, req *types.UpdateBannerRequest) (*types.BannerItemResponse, error) {
	id, err := uuid.Parse(idStr)
	if err != nil {
		return nil, fmt.Errorf("شناسه بنر نامعتبر است")
	}

	var banner model.Banner
	if err := l.svcCtx.DB.WithContext(l.ctx).First(&banner, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("بنر یافت نشد")
	}

	if req.Title != nil {
		banner.Title = *req.Title
	}
	if req.Subtitle != nil {
		banner.Subtitle = *req.Subtitle
	}
	if req.BadgeText != nil {
		banner.BadgeText = *req.BadgeText
	}
	if req.ImageURL != nil {
		banner.ImageURL = *req.ImageURL
	}
	if req.LinkURL != nil {
		banner.LinkURL = *req.LinkURL
	}
	if req.Position != nil {
		banner.Position = model.BannerPosition(*req.Position)
	}
	if req.BgGradient != nil {
		banner.BgGradient = *req.BgGradient
	}
	if req.ButtonText != nil {
		banner.ButtonText = *req.ButtonText
	}
	if req.SortOrder != nil {
		banner.SortOrder = *req.SortOrder
	}
	if req.IsActive != nil {
		banner.IsActive = *req.IsActive
	}

	if err := l.svcCtx.DB.WithContext(l.ctx).Save(&banner).Error; err != nil {
		return nil, err
	}

	return &types.BannerItemResponse{
		ID:         banner.ID.String(),
		Title:      banner.Title,
		Subtitle:   banner.Subtitle,
		BadgeText:  banner.BadgeText,
		ImageURL:   banner.ImageURL,
		LinkURL:    banner.LinkURL,
		Position:   string(banner.Position),
		BgGradient: banner.BgGradient,
		ButtonText: banner.ButtonText,
		SortOrder:  banner.SortOrder,
		IsActive:   banner.IsActive,
	}, nil
}

func (l *BannerAdminLogic) DeleteBanner(idStr string) error {
	id, err := uuid.Parse(idStr)
	if err != nil {
		return fmt.Errorf("شناسه بنر نامعتبر است")
	}
	return l.svcCtx.DB.WithContext(l.ctx).Delete(&model.Banner{}, "id = ?", id).Error
}
