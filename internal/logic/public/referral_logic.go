package public

import (
	"context"
	"fmt"
	"time"

	"github.com/EbadiDev/anahix-server/internal/model"
	"github.com/EbadiDev/anahix-server/internal/svc"
	"github.com/EbadiDev/anahix-server/internal/types"
	"github.com/EbadiDev/anahix-server/pkg/xerr"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ReferralLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewReferralLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReferralLogic {
	return &ReferralLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ReferralLogic) GetReferralStats(customerPhone string) (*types.ReferralStatsResponse, error) {
	var customer model.Customer
	if err := l.svcCtx.DB.WithContext(l.ctx).Where("phone = ?", customerPhone).First(&customer).Error; err != nil {
		return nil, xerr.NewErrMsg("کاربر یافت نشد")
	}

	// Count total referred users
	var totalReferrals int64
	l.svcCtx.DB.WithContext(l.ctx).Model(&model.Referral{}).
		Where("referrer_customer_id = ?", customer.ID).
		Count(&totalReferrals)

	// Count successful orders made by referred customers
	var successfulOrders int64
	l.svcCtx.DB.WithContext(l.ctx).Model(&model.ReferralCommission{}).
		Where("referrer_customer_id = ?", customer.ID).
		Count(&successfulOrders)

	// Sum total lifetime commission
	var totalCommissionToman int64
	l.svcCtx.DB.WithContext(l.ctx).Model(&model.ReferralCommission{}).
		Where("referrer_customer_id = ?", customer.ID).
		Select("COALESCE(SUM(commission_toman), 0)").
		Scan(&totalCommissionToman)

	referralCode := customer.ReferralCode
	if referralCode == "" {
		referralCode = fmt.Sprintf("ANX-%s", customer.Phone[len(customer.Phone)-4:])
	}

	return &types.ReferralStatsResponse{
		ReferralCode:                referralCode,
		ReferralURL:                 fmt.Sprintf("https://anahix.com/?ref=%s", referralCode),
		CommissionRatePercent:       10.0,
		TotalReferrals:              int(totalReferrals),
		SuccessfulOrders:            int(successfulOrders),
		TotalCommissionToman:        totalCommissionToman,
		WithdrawableCommissionToman: customer.ReferralBalanceToman,
	}, nil
}

func (l *ReferralLogic) TransferToWallet(customerPhone string) (*types.TransferCommissionResponse, error) {
	var customer model.Customer
	if err := l.svcCtx.DB.WithContext(l.ctx).Where("phone = ?", customerPhone).First(&customer).Error; err != nil {
		return nil, xerr.NewErrMsg("کاربر یافت نشد")
	}

	if customer.ReferralBalanceToman <= 0 {
		return nil, xerr.NewErrMsg("موجودی کمیسیون قابل انتقال صفر است")
	}

	transferAmount := customer.ReferralBalanceToman

	err := l.svcCtx.DB.WithContext(l.ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Deduct referral balance and add to main balance
		if err := tx.Model(&customer).Updates(map[string]interface{}{
			"balance_toman":          customer.BalanceToman + transferAmount,
			"referral_balance_toman": 0,
		}).Error; err != nil {
			return err
		}

		// 2. Mark confirmed commissions as TRANSFERRED_TO_WALLET
		if err := tx.Model(&model.ReferralCommission{}).
			Where("referrer_customer_id = ? AND status = ?", customer.ID, model.CommissionConfirmed).
			Update("status", model.CommissionTransferredToWallet).Error; err != nil {
			return err
		}

		// 3. Record in financial transaction ledger
		txNum := fmt.Sprintf("TX-REF-%d", time.Now().UnixNano()/1e6)
		ledgerTx := model.Transaction{
			ID:                uuid.New(),
			TransactionNumber: txNum,
			OrderType:         model.TxTypeDeposit,
			CustomerPhone:     customer.Phone,
			CustomerEmail:     customer.Email,
			AmountToman:       transferAmount,
			PaymentPlatform:   model.PlatformBalance,
			PaymentRefID:      "REFERRAL_TRANSFER",
			Status:            model.TxStatusSuccess,
			Description:       fmt.Sprintf("انتقال کمیسیون معرف به موجودی کیف پول (%d تومان)", transferAmount),
			CreatedAt:         time.Now(),
		}
		if err := tx.Create(&ledgerTx).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, xerr.NewErrCode(xerr.DatabaseError)
	}

	return &types.TransferCommissionResponse{
		TransferredToman: transferAmount,
		NewBalanceToman:  customer.BalanceToman,
		Message:          fmt.Sprintf("مبلغ %d تومان کمیسیون با موفقیت به کیف پول منتقل گردید", transferAmount),
	}, nil
}
