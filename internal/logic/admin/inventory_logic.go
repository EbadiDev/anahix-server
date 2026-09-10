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

type InventoryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewInventoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *InventoryLogic {
	return &InventoryLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *InventoryLogic) BulkImport(req *types.BulkImportInventoryRequest) (*types.BulkImportInventoryResponse, error) {
	prodUUID, err := uuid.Parse(req.ProductID)
	if err != nil {
		return nil, xerr.NewErrCode(xerr.InvalidParams)
	}

	var product model.Product
	if err := l.svcCtx.DB.WithContext(l.ctx).Where("id = ?", prodUUID).First(&product).Error; err != nil {
		return nil, xerr.NewErrCode(xerr.ProductNotFound)
	}

	var itemsToInsert []model.InventoryItem
	for _, rawItem := range req.Items {
		trimmed := strings.TrimSpace(rawItem)
		if trimmed == "" {
			continue
		}

		encrypted, err := l.svcCtx.Vault.Encrypt(trimmed)
		if err != nil {
			return nil, xerr.NewErrCode(xerr.InvalidItemPayload)
		}

		itemsToInsert = append(itemsToInsert, model.InventoryItem{
			ID:               uuid.New(),
			ProductID:        product.ID,
			ItemType:         model.ItemType(req.ItemType),
			PayloadEncrypted: encrypted,
			Status:           model.InventoryAvailable,
			BatchTag:         req.BatchTag,
		})
	}

	if len(itemsToInsert) == 0 {
		return &types.BulkImportInventoryResponse{ImportedCount: 0}, nil
	}

	// Batch insert (GORM CreateInBatches)
	if err := l.svcCtx.DB.WithContext(l.ctx).CreateInBatches(itemsToInsert, 100).Error; err != nil {
		return nil, xerr.NewErrCode(xerr.DatabaseError)
	}

	return &types.BulkImportInventoryResponse{
		ImportedCount: len(itemsToInsert),
	}, nil
}
