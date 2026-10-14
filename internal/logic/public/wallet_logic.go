package public

import (
	"context"

	"github.com/EbadiDev/anahix-server/internal/model"
	"github.com/EbadiDev/anahix-server/internal/svc"
	"github.com/EbadiDev/anahix-server/internal/types"
	"github.com/EbadiDev/anahix-server/pkg/xerr"
)

type WalletLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewWalletLogic(ctx context.Context, svcCtx *svc.ServiceContext) *WalletLogic {
	return &WalletLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *WalletLogic) GetCustomerLedger(customerPhone string) ([]types.WalletLedgerItemDTO, error) {
	if customerPhone == "" {
		return nil, xerr.NewErrCode(xerr.InvalidParams)
	}

	var txs []model.Transaction
	err := l.svcCtx.DB.WithContext(l.ctx).
		Where("customer_phone = ?", customerPhone).
		Order("created_at DESC").
		Limit(50).
		Find(&txs).Error
	if err != nil {
		return nil, xerr.NewErrCode(xerr.DatabaseError)
	}

	result := make([]types.WalletLedgerItemDTO, 0, len(txs))
	for _, t := range txs {
		title := t.Description
		if title == "" {
			switch t.OrderType {
			case model.TxTypeStoreOrder:
				title = "خرید اشتراک هوش مصنوعی"
			case model.TxTypeVirtualNumber:
				title = "خرید شماره مجازی اختصاصی"
			case model.TxTypeDeposit:
				title = "افزایش موجودی کیف پول"
			case model.TxTypeRefund:
				title = "استرداد وجه به کیف پول"
			default:
				title = "تراکنش مالی"
			}
		}

		result = append(result, types.WalletLedgerItemDTO{
			ID:                t.ID.String(),
			TransactionNumber: t.TransactionNumber,
			Title:             title,
			Type:              string(t.OrderType),
			AmountToman:       t.AmountToman,
			Status:            string(t.Status),
			PaymentMethod:     string(t.PaymentPlatform),
			PaymentRefID:      t.PaymentRefID,
			CreatedAt:         t.CreatedAt,
		})
	}

	return result, nil
}
