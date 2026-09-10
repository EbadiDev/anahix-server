package public

import (
	"context"

	"github.com/EbadiDev/anahix-server/internal/model"
	"github.com/EbadiDev/anahix-server/internal/svc"
	"github.com/EbadiDev/anahix-server/internal/types"
	"github.com/EbadiDev/anahix-server/pkg/xerr"
)

type CatalogLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCatalogLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CatalogLogic {
	return &CatalogLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CatalogLogic) ListProducts() ([]types.ProductListItem, error) {
	var products []model.Product
	err := l.svcCtx.DB.WithContext(l.ctx).
		Where("is_active = ?", true).
		Order("created_at desc").
		Find(&products).Error
	if err != nil {
		return nil, xerr.NewErrCode(xerr.DatabaseError)
	}

	result := make([]types.ProductListItem, len(products))
	for i, p := range products {
		var availableStock int64 = 0
		if p.DeliveryType == model.DeliveryInstant {
			l.svcCtx.DB.WithContext(l.ctx).
				Model(&model.InventoryItem{}).
				Where("product_id = ? AND status = ?", p.ID, model.InventoryAvailable).
				Count(&availableStock)
		}

		result[i] = types.ProductListItem{
			ID:             p.ID.String(),
			Title:          p.Title,
			TitleFA:        p.TitleFA,
			Slug:           p.Slug,
			Category:       string(p.Category),
			DeliveryType:   string(p.DeliveryType),
			PriceToman:     p.PriceToman,
			BaseCostUSD:    p.BaseCostUSD,
			SLAMinMinutes:  p.SLAMinMinutes,
			SLAMaxMinutes:  p.SLAMaxMinutes,
			SLADisplayText: p.SLADisplayText,
			OperatingHours: p.OperatingHours,
			IconURL:        p.IconURL,
			BannerURL:      p.BannerURL,
			AvailableStock: availableStock,
			IsActive:       p.IsActive,
		}
	}

	return result, nil
}

func (l *CatalogLogic) GetProductDetail(slug string) (*types.ProductDetail, error) {
	var p model.Product
	err := l.svcCtx.DB.WithContext(l.ctx).
		Where("slug = ? AND is_active = ?", slug, true).
		First(&p).Error
	if err != nil {
		return nil, xerr.NewErrCode(xerr.ProductNotFound)
	}

	var availableStock int64 = 0
	if p.DeliveryType == model.DeliveryInstant {
		l.svcCtx.DB.WithContext(l.ctx).
			Model(&model.InventoryItem{}).
			Where("product_id = ? AND status = ?", p.ID, model.InventoryAvailable).
			Count(&availableStock)
	}

	detail := &types.ProductDetail{
		ProductListItem: types.ProductListItem{
			ID:             p.ID.String(),
			Title:          p.Title,
			TitleFA:        p.TitleFA,
			Slug:           p.Slug,
			Category:       string(p.Category),
			DeliveryType:   string(p.DeliveryType),
			PriceToman:     p.PriceToman,
			BaseCostUSD:    p.BaseCostUSD,
			SLAMinMinutes:  p.SLAMinMinutes,
			SLAMaxMinutes:  p.SLAMaxMinutes,
			SLADisplayText: p.SLADisplayText,
			OperatingHours: p.OperatingHours,
			IconURL:        p.IconURL,
			BannerURL:      p.BannerURL,
			AvailableStock: availableStock,
			IsActive:       p.IsActive,
		},
		Description:  p.Description,
		Instructions: p.Instructions,
	}

	return detail, nil
}
