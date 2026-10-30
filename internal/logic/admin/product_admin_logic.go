package admin

import (
	"context"
	"fmt"
	"time"

	"github.com/EbadiDev/anahix-server/internal/model"
	"github.com/EbadiDev/anahix-server/internal/logic/notification"
	"github.com/EbadiDev/anahix-server/internal/svc"
	"github.com/EbadiDev/anahix-server/internal/types"
	"github.com/EbadiDev/anahix-server/pkg/xerr"
	"github.com/google/uuid"
)

type ProductAdminLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewProductAdminLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ProductAdminLogic {
	return &ProductAdminLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ProductAdminLogic) ListProducts() ([]model.Product, error) {
	var products []model.Product
	err := l.svcCtx.DB.WithContext(l.ctx).
		Order("created_at desc").
		Find(&products).Error
	if err != nil {
		return nil, xerr.NewErrCode(xerr.DatabaseError)
	}

	for i := range products {
		var stock int64
		if products[i].StockMode == "always_in_stock" {
			stock = 999
		} else if products[i].StockMode == "manual" {
			stock = int64(products[i].ManualStock)
		} else if products[i].StockMode == "out_of_stock" {
			stock = 0
		} else if products[i].DeliveryType == model.DeliveryInstant {
			if products[i].ExternalProvider != "" || products[i].DeliveryStrategy == "third_party_instant" {
				stock = int64(products[i].ManualStock)
			} else {
				l.svcCtx.DB.WithContext(l.ctx).
					Model(&model.InventoryItem{}).
					Where("product_id = ? AND status = ?", products[i].ID, model.InventoryAvailable).
					Count(&stock)
			}
		} else {
			stock = 99
		}
		if products[i].OutOfStockText == "" || products[i].OutOfStockText == "در حال تهیه و تامین مجدد" {
			products[i].OutOfStockText = "ناموجود"
		}
		products[i].AvailableStock = stock
	}

	return products, nil
}

func (l *ProductAdminLogic) CreateProduct(p *model.Product) (*model.Product, error) {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}

	if p.Slug == "" {
		p.Slug = fmt.Sprintf("prod-%s", p.ID.String()[:8])
	}

	if err := l.svcCtx.DB.WithContext(l.ctx).Create(p).Error; err != nil {
		return nil, xerr.NewErrCode(xerr.DatabaseError)
	}

	return p, nil
}

func (l *ProductAdminLogic) GetProduct(prodID uuid.UUID) (*model.Product, error) {
	var p model.Product
	if err := l.svcCtx.DB.WithContext(l.ctx).First(&p, "id = ?", prodID).Error; err != nil {
		return nil, xerr.NewErrCode(xerr.ProductNotFound)
	}

	var stock int64
	if p.StockMode == "always_in_stock" {
		stock = 999
	} else if p.StockMode == "manual" {
		stock = int64(p.ManualStock)
	} else if p.StockMode == "out_of_stock" {
		stock = 0
	} else if p.DeliveryType == model.DeliveryInstant {
		if p.ExternalProvider != "" || p.DeliveryStrategy == "third_party_instant" {
			stock = int64(p.ManualStock)
		} else {
			l.svcCtx.DB.WithContext(l.ctx).
				Model(&model.InventoryItem{}).
				Where("product_id = ? AND status = ?", p.ID, model.InventoryAvailable).
				Count(&stock)
		}
	} else {
		stock = 99
	}
	if p.OutOfStockText == "" || p.OutOfStockText == "در حال تهیه و تامین مجدد" {
		p.OutOfStockText = "ناموجود"
	}
	p.AvailableStock = stock

	return &p, nil
}

func (l *ProductAdminLogic) UpdateProduct(prodID uuid.UUID, req *types.AdminUpdateProductRequest) (*model.Product, error) {
	var existing model.Product
	if err := l.svcCtx.DB.WithContext(l.ctx).First(&existing, "id = ?", prodID).Error; err != nil {
		return nil, xerr.NewErrCode(xerr.ProductNotFound)
	}

	wasOutOfStock := existing.StockMode == "out_of_stock" || (existing.StockMode == "manual" && existing.ManualStock <= 0)

	if req.Title != nil {
		existing.Title = *req.Title
	}
	if req.TitleFA != nil {
		existing.TitleFA = *req.TitleFA
	}
	if req.Slug != nil && *req.Slug != "" {
		existing.Slug = *req.Slug
	}
	if req.Description != nil {
		existing.Description = *req.Description
	}
	if req.Category != nil {
		existing.Category = *req.Category
	}
	if req.DeliveryType != nil {
		existing.DeliveryType = *req.DeliveryType
	}
	if req.SLAMinMinutes != nil {
		existing.SLAMinMinutes = *req.SLAMinMinutes
	}
	if req.SLAMaxMinutes != nil {
		existing.SLAMaxMinutes = *req.SLAMaxMinutes
	}
	if req.SLADisplayText != nil {
		existing.SLADisplayText = *req.SLADisplayText
	}
	if req.OperatingHours != nil {
		existing.OperatingHours = *req.OperatingHours
	}
	if req.BaseCostUSD != nil {
		existing.BaseCostUSD = *req.BaseCostUSD
	}
	if req.PriceToman != nil {
		existing.PriceToman = *req.PriceToman
	}
	if req.OriginalPriceToman != nil {
		existing.OriginalPriceToman = *req.OriginalPriceToman
	}
	if req.DiscountPercent != nil {
		existing.DiscountPercent = *req.DiscountPercent
	}
	if req.IsActive != nil {
		existing.IsActive = *req.IsActive
	}
	if req.Instructions != nil {
		existing.Instructions = *req.Instructions
	}
	if req.GroupTag != nil {
		existing.GroupTag = *req.GroupTag
	}
	if req.GroupTitleFA != nil {
		existing.GroupTitleFA = *req.GroupTitleFA
	}
	if req.DeliveryStrategy != nil {
		existing.DeliveryStrategy = *req.DeliveryStrategy
	}
	if req.StockMode != nil {
		existing.StockMode = *req.StockMode
	}
	if req.ManualStock != nil {
		existing.ManualStock = *req.ManualStock
	}
	if req.StockDisplayMode != nil {
		existing.StockDisplayMode = *req.StockDisplayMode
	}
	if req.OutOfStockText != nil {
		existing.OutOfStockText = *req.OutOfStockText
	}
	if req.MetaTitle != nil {
		existing.MetaTitle = *req.MetaTitle
	}
	if req.MetaDescription != nil {
		existing.MetaDescription = *req.MetaDescription
	}
	if req.OgImageURL != nil {
		existing.OgImageURL = *req.OgImageURL
	}
	if req.ImageURL != nil {
		existing.ImageURL = *req.ImageURL
	}
	if req.IconURL != nil {
		existing.IconURL = *req.IconURL
	}
	if req.IsFlashSale != nil {
		existing.IsFlashSale = *req.IsFlashSale
	}
	if req.FlashSalePriceToman != nil {
		existing.FlashSalePriceToman = *req.FlashSalePriceToman
	}
	if req.FlashSaleStock != nil {
		existing.FlashSaleStock = *req.FlashSaleStock
	}
	if req.FlashSaleSold != nil {
		existing.FlashSaleSold = *req.FlashSaleSold
	}
	if req.FlashSaleEndAt != nil {
		if *req.FlashSaleEndAt != "" {
			if t, err := time.Parse(time.RFC3339, *req.FlashSaleEndAt); err == nil {
				existing.FlashSaleEndAt = &t
			} else if t, err := time.Parse("2006-01-02T15:04", *req.FlashSaleEndAt); err == nil {
				existing.FlashSaleEndAt = &t
			} else if t, err := time.Parse("2006-01-02", *req.FlashSaleEndAt); err == nil {
				existing.FlashSaleEndAt = &t
			}
		} else {
			existing.FlashSaleEndAt = nil
		}
	}
	if req.BadgesJSON != nil {
		existing.BadgesJSON = *req.BadgesJSON
	}
	if req.RatingAverage != nil {
		existing.RatingAverage = *req.RatingAverage
	}
	if req.IsBestSeller != nil {
		existing.IsBestSeller = *req.IsBestSeller
	}
	if req.IsFeatured != nil {
		existing.IsFeatured = *req.IsFeatured
	}
	if req.PlansJSON != nil {
		existing.PlansJSON = *req.PlansJSON
	}
	if req.ComparisonJSON != nil {
		existing.ComparisonJSON = *req.ComparisonJSON
	}
	if req.GuideJSON != nil {
		existing.GuideJSON = *req.GuideJSON
	}
	if req.FaqJSON != nil {
		existing.FaqJSON = *req.FaqJSON
	}
	if req.FeaturesJSON != nil {
		existing.FeaturesJSON = *req.FeaturesJSON
	}

	if err := l.svcCtx.DB.WithContext(l.ctx).Save(&existing).Error; err != nil {
		return nil, xerr.NewErrCode(xerr.DatabaseError)
	}

	// Trigger Stock Alert SMS if restocked
	isNowInStock := (existing.StockMode != "out_of_stock") && (existing.StockMode != "manual" || existing.ManualStock > 0)
	if wasOutOfStock && isNowInStock {
		notifier := notification.NewSMSPatternNotifier(l.svcCtx)
		notifier.NotifyRestock(l.ctx, existing.ID.String(), existing.TitleFA, existing.Slug)
	}

	return &existing, nil
}

func (l *ProductAdminLogic) DeleteProduct(prodID uuid.UUID) (*model.Product, error) {
	var existing model.Product
	if err := l.svcCtx.DB.WithContext(l.ctx).First(&existing, "id = ?", prodID).Error; err != nil {
		return nil, xerr.NewErrCode(xerr.ProductNotFound)
	}

	if err := l.svcCtx.DB.WithContext(l.ctx).Delete(&existing).Error; err != nil {
		return nil, xerr.NewErrCode(xerr.DatabaseError)
	}

	return &existing, nil
}
