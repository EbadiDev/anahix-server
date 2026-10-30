package public

import (
	"context"

	"github.com/EbadiDev/anahix-server/internal/model"
	"github.com/EbadiDev/anahix-server/internal/svc"
	"github.com/EbadiDev/anahix-server/internal/types"
	"github.com/EbadiDev/anahix-server/pkg/xerr"
)

type PaymentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPaymentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PaymentLogic {
	return &PaymentLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PaymentLogic) ListPaymentMethods() ([]types.PaymentMethodItem, error) {
	var payments []model.Payment
	err := l.svcCtx.DB.WithContext(l.ctx).
		Where("display_mode != ? AND (enable = ? OR display_mode = ?)", "hidden", true, "disabled_visible").
		Order("sort asc, id asc").
		Find(&payments).Error
	if err != nil {
		return nil, xerr.NewErrCode(xerr.DatabaseError)
	}

	result := make([]types.PaymentMethodItem, len(payments))
	for i, p := range payments {
		mode := p.DisplayMode
		if mode == "" {
			if p.Enable {
				mode = "active"
			} else {
				mode = "hidden"
			}
		}
		result[i] = types.PaymentMethodItem{
			ID:          p.Id,
			Name:        p.Name,
			NameFA:      p.NameFA,
			Platform:    string(p.Platform),
			Icon:        p.Icon,
			Description: p.Description,
			Enable:      mode == "active",
			DisplayMode: mode,
		}
	}

	return result, nil
}
