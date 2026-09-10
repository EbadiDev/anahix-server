package admin

import (
	"context"
	"time"

	"github.com/EbadiDev/anahix-server/internal/model"
	"github.com/EbadiDev/anahix-server/internal/svc"
	"github.com/EbadiDev/anahix-server/internal/types"
	"github.com/EbadiDev/anahix-server/pkg/xerr"
	"github.com/google/uuid"
)

type OrderDeskLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewOrderDeskLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OrderDeskLogic {
	return &OrderDeskLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

type AdminOrderDetailResponse struct {
	Order               model.Order `json:"order"`
	CustomerCredentials string      `json:"customer_credentials,omitempty"`
}

func (l *OrderDeskLogic) ListQueueOrders(status string, deliveryType string) ([]model.Order, error) {
	query := l.svcCtx.DB.WithContext(l.ctx).Preload("Items").Order("created_at asc")

	if status != "" {
		query = query.Where("status = ?", status)
	} else {
		// Default to orders needing operator attention
		query = query.Where("status IN ?", []model.OrderStatus{
			model.OrderPaid,
			model.OrderInQueue,
			model.OrderProcessing,
			model.OrderAwaitingCustomerAction,
		})
	}

	if deliveryType != "" {
		query = query.Where("delivery_type = ?", deliveryType)
	}

	var orders []model.Order
	if err := query.Find(&orders).Error; err != nil {
		return nil, xerr.NewErrCode(xerr.DatabaseError)
	}

	return orders, nil
}

func (l *OrderDeskLogic) GetOrderDetail(orderIDStr string) (*AdminOrderDetailResponse, error) {
	orderUUID, err := uuid.Parse(orderIDStr)
	if err != nil {
		return nil, xerr.NewErrCode(xerr.InvalidParams)
	}

	var order model.Order
	if err := l.svcCtx.DB.WithContext(l.ctx).Preload("Items").Where("id = ?", orderUUID).First(&order).Error; err != nil {
		return nil, xerr.NewErrCode(xerr.OrderNotFound)
	}

	plainCreds := ""
	if order.CustomerCredentialsEncrypted != "" {
		dec, decErr := l.svcCtx.Vault.Decrypt(order.CustomerCredentialsEncrypted)
		if decErr == nil {
			plainCreds = dec
		}
	}

	return &AdminOrderDetailResponse{
		Order:               order,
		CustomerCredentials: plainCreds,
	}, nil
}

func (l *OrderDeskLogic) RequestTwoFactor(orderIDStr string, req *types.RequestTwoFactorRequest) error {
	orderUUID, err := uuid.Parse(orderIDStr)
	if err != nil {
		return xerr.NewErrCode(xerr.InvalidParams)
	}

	var order model.Order
	if err := l.svcCtx.DB.WithContext(l.ctx).Where("id = ?", orderUUID).First(&order).Error; err != nil {
		return xerr.NewErrCode(xerr.OrderNotFound)
	}

	now := time.Now()
	order.Status = model.OrderAwaitingCustomerAction
	order.TwoFactorStatus = model.TwoFactorRequested
	order.TwoFactorPrompt = req.PromptMessage
	order.TwoFactorRequestedAt = &now

	if err := l.svcCtx.DB.WithContext(l.ctx).Save(&order).Error; err != nil {
		return xerr.NewErrCode(xerr.DatabaseError)
	}

	return nil
}

func (l *OrderDeskLogic) CompleteOrder(orderIDStr string, req *types.CompleteOrderRequest) error {
	orderUUID, err := uuid.Parse(orderIDStr)
	if err != nil {
		return xerr.NewErrCode(xerr.InvalidParams)
	}

	var order model.Order
	if err := l.svcCtx.DB.WithContext(l.ctx).Where("id = ?", orderUUID).First(&order).Error; err != nil {
		return xerr.NewErrCode(xerr.OrderNotFound)
	}

	now := time.Now()
	order.Status = model.OrderCompleted
	order.CompletedAt = &now
	order.TwoFactorStatus = model.TwoFactorVerified
	if req.OperatorNotes != "" {
		order.OperatorNotes = req.OperatorNotes
	}

	// ZEROIZE / PURGE customer credentials for privacy guardrail
	order.CustomerCredentialsEncrypted = ""

	if err := l.svcCtx.DB.WithContext(l.ctx).Save(&order).Error; err != nil {
		return xerr.NewErrCode(xerr.DatabaseError)
	}

	return nil
}
