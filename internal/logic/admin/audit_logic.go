package admin

import (
	"context"
	"strings"

	"github.com/EbadiDev/anahix-server/internal/model"
	"github.com/EbadiDev/anahix-server/internal/svc"
	"github.com/EbadiDev/anahix-server/internal/types"
	"github.com/EbadiDev/anahix-server/pkg/xerr"
)

type AuditLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAuditLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AuditLogic {
	return &AuditLogic{ctx: ctx, svcCtx: svcCtx}
}

func (l *AuditLogic) ListTransactions(page, pageSize int, search, status, orderType, paymentPlatform string) (*types.TransactionListResponse, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 25
	}
	offset := (page - 1) * pageSize

	db := l.svcCtx.DB.WithContext(l.ctx).Model(&model.Transaction{})

	search = strings.TrimSpace(search)
	if search != "" {
		like := "%" + search + "%"
		db = db.Where("transaction_number LIKE ? OR customer_phone LIKE ? OR customer_email LIKE ? OR payment_ref_id LIKE ?",
			like, like, like, like)
	}

	if status != "" {
		db = db.Where("status = ?", status)
	}
	if orderType != "" {
		db = db.Where("order_type = ?", orderType)
	}
	if paymentPlatform != "" {
		db = db.Where("payment_platform = ?", paymentPlatform)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, xerr.NewErrCodeMsg(xerr.DatabaseError, "خطا در شمارش لاگ تراکنش‌ها")
	}

	var items []model.Transaction
	if err := db.Order("created_at desc").Offset(offset).Limit(pageSize).Find(&items).Error; err != nil {
		return nil, xerr.NewErrCodeMsg(xerr.DatabaseError, "خطا در دریافت لیست تراکنش‌ها")
	}

	if items == nil {
		items = []model.Transaction{}
	}

	return &types.TransactionListResponse{
		Total: total,
		Items: items,
	}, nil
}

func (l *AuditLogic) ListAdminActions(page, pageSize int, search, action, resource, adminUsername string) (*types.AdminActionLogListResponse, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 25
	}
	offset := (page - 1) * pageSize

	db := l.svcCtx.DB.WithContext(l.ctx).Model(&model.AdminActionLog{})

	search = strings.TrimSpace(search)
	if search != "" {
		like := "%" + search + "%"
		db = db.Where("description LIKE ? OR admin_username LIKE ? OR resource_id LIKE ? OR action LIKE ?",
			like, like, like, like)
	}

	if action != "" {
		db = db.Where("action = ?", action)
	}
	if resource != "" {
		db = db.Where("resource = ?", resource)
	}
	if adminUsername != "" {
		db = db.Where("admin_username = ?", adminUsername)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, xerr.NewErrCodeMsg(xerr.DatabaseError, "خطا در شمارش لاگ رویدادهای مدیران")
	}

	var items []model.AdminActionLog
	if err := db.Order("created_at desc").Offset(offset).Limit(pageSize).Find(&items).Error; err != nil {
		return nil, xerr.NewErrCodeMsg(xerr.DatabaseError, "خطا در استعلام لاگ فعالیت مدیران")
	}

	if items == nil {
		items = []model.AdminActionLog{}
	}

	return &types.AdminActionLogListResponse{
		Total: total,
		Items: items,
	}, nil
}
