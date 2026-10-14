package admin

import (
	"context"
	"strings"
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

type AdminQueueOrderItem struct {
	ID                   uuid.UUID             `json:"id"`
	OrderID              string                `json:"order_id"`
	OrderNumber          string                `json:"order_number"`
	CustomerPhone        string                `json:"customer_phone"`
	CustomerEmail        string                `json:"customer_email"`
	Status               model.OrderStatus     `json:"status"`
	DeliveryType         model.DeliveryType    `json:"delivery_type"`
	TotalAmountToman     int64                 `json:"total_amount_toman"`
	PaymentMethod        model.PaymentMethod   `json:"payment_method"`
	ProductTitle         string                `json:"product_title"`
	CustomerCredentials  string                `json:"customer_credentials,omitempty"`
	AccountEmail         string                `json:"account_email,omitempty"`
	AccountPassword      string                `json:"account_password,omitempty"`
	TwoFactorPrompt      string                `json:"two_factor_prompt,omitempty"`
	TwoFactorCode        string                `json:"two_factor_code,omitempty"`
	TwoFactorStatus      model.TwoFactorStatus `json:"two_factor_status"`
	TwoFactorRequestedAt *time.Time            `json:"two_factor_requested_at,omitempty"`
	CreatedAt            time.Time             `json:"created_at"`
	UpdatedAt            time.Time             `json:"updated_at"`
}

type AdminOrderDetailResponse struct {
	Order               model.Order `json:"order"`
	CustomerCredentials string      `json:"customer_credentials,omitempty"`
}

func (l *OrderDeskLogic) ListQueueOrders(status string, deliveryType string) ([]AdminQueueOrderItem, error) {
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

	resultList := make([]AdminQueueOrderItem, 0, len(orders))
	for _, o := range orders {
		item := AdminQueueOrderItem{
			ID:                   o.ID,
			OrderID:              o.ID.String(),
			OrderNumber:          o.OrderNumber,
			CustomerPhone:        o.CustomerPhone,
			CustomerEmail:        o.CustomerEmail,
			Status:               o.Status,
			DeliveryType:         o.DeliveryType,
			TotalAmountToman:     o.TotalAmountToman,
			PaymentMethod:        o.PaymentMethod,
			TwoFactorPrompt:      o.TwoFactorPrompt,
			TwoFactorCode:        o.TwoFactorCode,
			TwoFactorStatus:      o.TwoFactorStatus,
			TwoFactorRequestedAt: o.TwoFactorRequestedAt,
			CreatedAt:            o.CreatedAt,
			UpdatedAt:            o.UpdatedAt,
		}

		if len(o.Items) > 0 {
			item.ProductTitle = o.Items[0].ProductTitle
		}

		if o.CustomerCredentialsEncrypted != "" && l.svcCtx.Vault != nil {
			dec, decErr := l.svcCtx.Vault.Decrypt(o.CustomerCredentialsEncrypted)
			if decErr == nil && dec != "" {
				item.CustomerCredentials = dec
				// Parse "Email: ... | Pass: ... | Notes: ..."
				parts := strings.Split(dec, " | ")
				for _, part := range parts {
					if strings.HasPrefix(part, "Email: ") {
						item.AccountEmail = strings.TrimPrefix(part, "Email: ")
					} else if strings.HasPrefix(part, "Pass: ") {
						item.AccountPassword = strings.TrimPrefix(part, "Pass: ")
					}
				}
			}
		}

		resultList = append(resultList, item)
	}

	return resultList, nil
}

func (l *OrderDeskLogic) GetOrderDetail(orderIDStr string) (*AdminOrderDetailResponse, error) {
	var order model.Order
	orderUUID, err := uuid.Parse(orderIDStr)
	if err == nil {
		if err := l.svcCtx.DB.WithContext(l.ctx).Preload("Items").Where("id = ?", orderUUID).First(&order).Error; err != nil {
			return nil, xerr.NewErrCode(xerr.OrderNotFound)
		}
	} else {
		if err := l.svcCtx.DB.WithContext(l.ctx).Preload("Items").Where("order_number = ?", orderIDStr).First(&order).Error; err != nil {
			return nil, xerr.NewErrCode(xerr.OrderNotFound)
		}
	}

	plainCreds := ""
	if order.CustomerCredentialsEncrypted != "" && l.svcCtx.Vault != nil {
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
	var order model.Order
	orderUUID, err := uuid.Parse(orderIDStr)
	if err == nil {
		if err := l.svcCtx.DB.WithContext(l.ctx).Where("id = ?", orderUUID).First(&order).Error; err != nil {
			return xerr.NewErrCode(xerr.OrderNotFound)
		}
	} else {
		if err := l.svcCtx.DB.WithContext(l.ctx).Where("order_number = ?", orderIDStr).First(&order).Error; err != nil {
			return xerr.NewErrCode(xerr.OrderNotFound)
		}
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
	var order model.Order
	orderUUID, err := uuid.Parse(orderIDStr)
	if err == nil {
		if err := l.svcCtx.DB.WithContext(l.ctx).Where("id = ?", orderUUID).First(&order).Error; err != nil {
			return xerr.NewErrCode(xerr.OrderNotFound)
		}
	} else {
		if err := l.svcCtx.DB.WithContext(l.ctx).Where("order_number = ?", orderIDStr).First(&order).Error; err != nil {
			return xerr.NewErrCode(xerr.OrderNotFound)
		}
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
