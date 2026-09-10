package public

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/EbadiDev/anahix-server/internal/model"
	"github.com/EbadiDev/anahix-server/internal/svc"
	"github.com/EbadiDev/anahix-server/internal/types"
	"github.com/EbadiDev/anahix-server/pkg/xerr"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type OrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OrderLogic {
	return &OrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// generateSecureToken creates a 32-byte hex random string for order access
func generateSecureToken() string {
	bytes := make([]byte, 32)
	_, _ = rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// generateOrderNumber creates a human-readable order code
func generateOrderNumber() string {
	now := time.Now()
	randomSuffix := make([]byte, 2)
	_, _ = rand.Read(randomSuffix)
	return fmt.Sprintf("ANX-%s-%X", now.Format("0601021504"), randomSuffix)
}

func (l *OrderLogic) CreateOrder(req *types.CreateOrderRequest) (*types.CreateOrderResponse, error) {
	prodUUID, err := uuid.Parse(req.ProductID)
	if err != nil {
		return nil, xerr.NewErrCode(xerr.InvalidParams)
	}

	var product model.Product
	if err := l.svcCtx.DB.WithContext(l.ctx).Where("id = ? AND is_active = ?", prodUUID, true).First(&product).Error; err != nil {
		return nil, xerr.NewErrCode(xerr.ProductNotFound)
	}

	accessToken := generateSecureToken()
	orderNumber := generateOrderNumber()
	expiresAt := time.Now().Add(20 * time.Minute)

	var createdOrder model.Order

	// Transaction to atomically reserve stock (if instant) and create order
	txErr := l.svcCtx.DB.WithContext(l.ctx).Transaction(func(tx *gorm.DB) error {
		var reservedItemID *uuid.UUID

		if product.DeliveryType == model.DeliveryInstant {
			// Find and lock 1 available inventory item using SELECT ... FOR UPDATE SKIP LOCKED
			var item model.InventoryItem
			err := tx.Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
				Where("product_id = ? AND status = ?", product.ID, model.InventoryAvailable).
				First(&item).Error
			if err != nil {
				return xerr.NewErrCode(xerr.OutOfStock)
			}

			reservedUntil := expiresAt
			item.Status = model.InventoryReserved
			item.ReservedUntil = &reservedUntil

			if err := tx.Save(&item).Error; err != nil {
				return err
			}
			reservedItemID = &item.ID
		}

		// Encrypt customer account credentials if this is a timed on-account upgrade
		var encryptedCreds string
		if product.DeliveryType == model.DeliveryManualTimed && req.CustomerAccountEmail != "" {
			plainCreds := fmt.Sprintf("Email: %s | Pass: %s | Notes: %s",
				req.CustomerAccountEmail, req.CustomerAccountPassword, req.CustomerNotes)
			enc, encErr := l.svcCtx.Vault.Encrypt(plainCreds)
			if encErr != nil {
				return encErr
			}
			encryptedCreds = enc
		}

		orderID := uuid.New()
		order := model.Order{
			ID:                           orderID,
			OrderNumber:                  orderNumber,
			CustomerPhone:                req.CustomerPhone,
			CustomerEmail:                req.CustomerEmail,
			AccessToken:                  accessToken,
			Status:                       model.OrderPendingPayment,
			DeliveryType:                 product.DeliveryType,
			TotalAmountToman:             product.PriceToman,
			PaymentMethod:                model.PaymentMethod(req.PaymentMethod),
			ExpiresAt:                    expiresAt,
			CustomerCredentialsEncrypted: encryptedCreds,
			TwoFactorStatus:              model.TwoFactorIdle,
			Items: []model.OrderItem{
				{
					OrderID:         orderID,
					ProductID:       product.ID,
					InventoryItemID: reservedItemID,
					ProductTitle:    product.TitleFA,
					ProductSlug:     product.Slug,
					DeliveryType:    product.DeliveryType,
					UnitPriceToman:  product.PriceToman,
					Instructions:    product.Instructions,
				},
			},
		}

		if err := tx.Create(&order).Error; err != nil {
			return err
		}

		// If stock was reserved, link the order ID to the inventory item
		if reservedItemID != nil {
			if err := tx.Model(&model.InventoryItem{}).
				Where("id = ?", *reservedItemID).
				Update("reserved_order_id", orderID).Error; err != nil {
				return err
			}
		}

		createdOrder = order
		return nil
	})

	if txErr != nil {
		return nil, txErr
	}

	paymentURL := fmt.Sprintf("/checkout/pay/%s?token=%s", createdOrder.ID.String(), accessToken)

	return &types.CreateOrderResponse{
		OrderID:          createdOrder.ID.String(),
		OrderNumber:      createdOrder.OrderNumber,
		AccessToken:      createdOrder.AccessToken,
		TotalAmountToman: createdOrder.TotalAmountToman,
		PaymentMethod:    string(createdOrder.PaymentMethod),
		PaymentURL:       paymentURL,
		ExpiresAt:        createdOrder.ExpiresAt,
	}, nil
}

func (l *OrderLogic) GetOrderDetail(orderIDStr string, token string) (*types.OrderDetailResponse, error) {
	orderUUID, err := uuid.Parse(orderIDStr)
	if err != nil {
		return nil, xerr.NewErrCode(xerr.InvalidParams)
	}

	var order model.Order
	err = l.svcCtx.DB.WithContext(l.ctx).
		Preload("Items").
		Where("id = ? AND access_token = ?", orderUUID, token).
		First(&order).Error
	if err != nil {
		return nil, xerr.NewErrCode(xerr.OrderNotFound)
	}

	resp := &types.OrderDetailResponse{
		OrderID:              order.ID.String(),
		OrderNumber:          order.OrderNumber,
		Status:               string(order.Status),
		DeliveryType:         string(order.DeliveryType),
		TotalAmountToman:     order.TotalAmountToman,
		CustomerPhone:        order.CustomerPhone,
		PaymentMethod:        string(order.PaymentMethod),
		CreatedAt:            order.CreatedAt,
		PaidAt:               order.PaidAt,
		CompletedAt:          order.CompletedAt,
		TwoFactorPrompt:      order.TwoFactorPrompt,
		TwoFactorStatus:      string(order.TwoFactorStatus),
		TwoFactorRequestedAt: order.TwoFactorRequestedAt,
	}

	// If the order is paid or completed, decrypt any fulfilled items
	isDelivered := order.Status == model.OrderPaid || order.Status == model.OrderCompleted
	if isDelivered {
		var deliveredDTOs []types.DeliveredItemDTO
		for _, item := range order.Items {
			payloadText := ""
			if item.InventoryItemID != nil {
				var invItem model.InventoryItem
				if err := l.svcCtx.DB.WithContext(l.ctx).Where("id = ?", *item.InventoryItemID).First(&invItem).Error; err == nil {
					decrypted, decErr := l.svcCtx.Vault.Decrypt(invItem.PayloadEncrypted)
					if decErr == nil {
						payloadText = decrypted
					}
				}
			}

			deliveredDTOs = append(deliveredDTOs, types.DeliveredItemDTO{
				ProductTitle: item.ProductTitle,
				ProductSlug:  item.ProductSlug,
				DeliveryType: string(item.DeliveryType),
				Payload:      payloadText,
				Instructions: item.Instructions,
			})
		}
		resp.DeliveredItems = deliveredDTOs
	}

	return resp, nil
}

func (l *OrderLogic) SubmitTwoFactor(orderIDStr string, token string, code string) error {
	orderUUID, err := uuid.Parse(orderIDStr)
	if err != nil {
		return xerr.NewErrCode(xerr.InvalidParams)
	}

	var order model.Order
	err = l.svcCtx.DB.WithContext(l.ctx).
		Where("id = ? AND access_token = ?", orderUUID, token).
		First(&order).Error
	if err != nil {
		return xerr.NewErrCode(xerr.OrderNotFound)
	}

	if order.TwoFactorStatus != model.TwoFactorRequested {
		return xerr.NewErrCodeMsg(xerr.OrderInvalidStatus, "در حال حاضر نیازی به کد تایید دومرحله‌ای نیست")
	}

	order.TwoFactorCode = code
	order.TwoFactorStatus = model.TwoFactorSubmitted

	if err := l.svcCtx.DB.WithContext(l.ctx).Save(&order).Error; err != nil {
		return xerr.NewErrCode(xerr.DatabaseError)
	}

	return nil
}
