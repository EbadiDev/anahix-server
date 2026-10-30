package admin

import (
	"context"
	"strings"
	"time"

	"github.com/EbadiDev/anahix-server/internal/model"
	"github.com/EbadiDev/anahix-server/internal/logic/notification"
	"github.com/EbadiDev/anahix-server/internal/svc"
	"github.com/EbadiDev/anahix-server/internal/types"
	"github.com/EbadiDev/anahix-server/pkg/xerr"
	"github.com/google/uuid"
	"gorm.io/gorm"
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

	// Trigger Stock Alert SMS for waiting customers
	notifier := notification.NewSMSPatternNotifier(l.svcCtx)
	notifier.NotifyRestock(l.ctx, product.ID.String(), product.TitleFA, product.Slug)

	return &types.BulkImportInventoryResponse{
		ImportedCount: len(itemsToInsert),
	}, nil
}

func (l *InventoryLogic) GetStats() (*types.InventoryStatsResponse, error) {
	var total, available, reserved, sold, void int64

	db := l.svcCtx.DB.WithContext(l.ctx).Model(&model.InventoryItem{})
	_ = db.Count(&total).Error
	_ = l.svcCtx.DB.WithContext(l.ctx).Model(&model.InventoryItem{}).Where("status = ?", model.InventoryAvailable).Count(&available).Error
	_ = l.svcCtx.DB.WithContext(l.ctx).Model(&model.InventoryItem{}).Where("status = ?", model.InventoryReserved).Count(&reserved).Error
	_ = l.svcCtx.DB.WithContext(l.ctx).Model(&model.InventoryItem{}).Where("status IN (?, ?)", model.InventorySold, model.InventoryDelivered).Count(&sold).Error
	_ = l.svcCtx.DB.WithContext(l.ctx).Model(&model.InventoryItem{}).Where("status = ?", model.InventoryVoid).Count(&void).Error

	return &types.InventoryStatsResponse{
		TotalCount:         total,
		AvailableCount:     available,
		ReservedCount:      reserved,
		SoldDeliveredCount: sold,
		VoidCount:          void,
	}, nil
}

func (l *InventoryLogic) ListInventory(productID string, status string, page int, pageSize int) (*types.ListInventoryResponse, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	query := l.svcCtx.DB.WithContext(l.ctx).Model(&model.InventoryItem{})
	if productID != "" {
		if pUUID, err := uuid.Parse(productID); err == nil {
			query = query.Where("product_id = ?", pUUID)
		}
	}
	if status != "" && status != "all" {
		query = query.Where("status = ?", status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, xerr.NewErrCode(xerr.DatabaseError)
	}

	var items []model.InventoryItem
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&items).Error; err != nil {
		return nil, xerr.NewErrCode(xerr.DatabaseError)
	}

	// Fetch product titles for display
	productMap := make(map[uuid.UUID]string)
	var productIDs []uuid.UUID
	for _, it := range items {
		productIDs = append(productIDs, it.ProductID)
	}
	if len(productIDs) > 0 {
		var prods []model.Product
		if err := l.svcCtx.DB.WithContext(l.ctx).Where("id IN ?", productIDs).Find(&prods).Error; err == nil {
			for _, p := range prods {
				productMap[p.ID] = p.TitleFA
			}
		}
	}

	dtos := make([]types.InventoryItemDTO, 0, len(items))
	for _, it := range items {
		prodTitle := productMap[it.ProductID]
		if prodTitle == "" {
			prodTitle = "محصول نامشخص"
		}

		preview := "••••••••"
		if dec, err := l.svcCtx.Vault.Decrypt(it.PayloadEncrypted); err == nil {
			if len(dec) > 12 {
				preview = dec[:8] + "..."
			} else {
				preview = dec
			}
		}

		var resOrderStr *string
		if it.ReservedOrderID != nil {
			s := it.ReservedOrderID.String()
			resOrderStr = &s
		}

		dtos = append(dtos, types.InventoryItemDTO{
			ID:              it.ID.String(),
			ProductID:       it.ProductID.String(),
			ProductTitle:    prodTitle,
			ItemType:        string(it.ItemType),
			Status:          string(it.Status),
			CostUSD:         it.CostUSD,
			CostToman:       it.CostToman,
			ExpirationDate:  it.ExpirationDate,
			BatchTag:        it.BatchTag,
			ReservedOrderID: resOrderStr,
			DeliveredAt:     it.DeliveredAt,
			CreatedAt:       it.CreatedAt,
			PayloadPreview:  preview,
		})
	}

	return &types.ListInventoryResponse{
		Total: total,
		Items: dtos,
	}, nil
}

func (l *InventoryLogic) ReplaceKey(req *types.ReplaceKeyRequest) (*types.ReplaceKeyResponse, error) {
	orderUUID, err := uuid.Parse(req.OrderID)
	if err != nil {
		return nil, xerr.NewErrCode(xerr.InvalidParams)
	}
	oldItemUUID, err := uuid.Parse(req.OldInventoryItemID)
	if err != nil {
		return nil, xerr.NewErrCode(xerr.InvalidParams)
	}

	var newPayload string
	var newItemID uuid.UUID

	err = l.svcCtx.DB.WithContext(l.ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Fetch old inventory item
		var oldItem model.InventoryItem
		if err := tx.Where("id = ?", oldItemUUID).First(&oldItem).Error; err != nil {
			return xerr.NewErrMsg("آیتم انبار قبلی یافت نشد")
		}

		// 2. Mark old item as VOID
		oldItem.Status = model.InventoryVoid
		oldItem.BatchTag = oldItem.BatchTag + " [REPLACED: " + req.Reason + "]"
		if err := tx.Save(&oldItem).Error; err != nil {
			return err
		}

		// 3. Find replacement available item from same product
		var replacement model.InventoryItem
		if err := tx.Where("product_id = ? AND status = ?", oldItem.ProductID, model.InventoryAvailable).
			First(&replacement).Error; err != nil {
			return xerr.NewErrMsg("موجودی جایگزین برای این محصول موجود نیست")
		}

		now := time.Now()
		replacement.Status = model.InventorySold
		replacement.ReservedOrderID = &orderUUID
		replacement.SoldAt = &now
		replacement.DeliveredAt = &now
		if err := tx.Save(&replacement).Error; err != nil {
			return err
		}

		// 4. Update order item with new inventory item and payload
		dec, decErr := l.svcCtx.Vault.Decrypt(replacement.PayloadEncrypted)
		if decErr != nil {
			return decErr
		}
		newPayload = dec
		newItemID = replacement.ID

		if err := tx.Model(&model.OrderItem{}).
			Where("order_id = ? AND inventory_item_id = ?", orderUUID, oldItemUUID).
			Updates(map[string]interface{}{
				"inventory_item_id": replacement.ID,
				"delivered_payload": dec,
				"instructions":      "کلید لایسنس جدید جایگزین‌شده: " + dec,
			}).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &types.ReplaceKeyResponse{
		Success:            true,
		NewInventoryItemID: newItemID.String(),
		DeliveredPayload:   newPayload,
		Message:            "کلید با موفقیت با آیتم جدید از انبار جایگزین گردید",
	}, nil
}

func (l *InventoryLogic) ExportKeys(productID string, status string) ([]string, error) {
	query := l.svcCtx.DB.WithContext(l.ctx).Model(&model.InventoryItem{})
	if productID != "" {
		if pUUID, err := uuid.Parse(productID); err == nil {
			query = query.Where("product_id = ?", pUUID)
		}
	}
	if status != "" && status != "all" {
		query = query.Where("status = ?", status)
	}

	var items []model.InventoryItem
	if err := query.Find(&items).Error; err != nil {
		return nil, xerr.NewErrCode(xerr.DatabaseError)
	}

	result := make([]string, 0, len(items))
	for _, it := range items {
		dec, err := l.svcCtx.Vault.Decrypt(it.PayloadEncrypted)
		if err == nil {
			result = append(result, dec)
		}
	}

	return result, nil
}

