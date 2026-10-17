package public

import (
	"context"
	"errors"
	"time"

	"github.com/EbadiDev/anahix-server/internal/model"
	"github.com/EbadiDev/anahix-server/internal/svc"
	"github.com/EbadiDev/anahix-server/internal/types"
	"github.com/EbadiDev/anahix-server/pkg/xerr"
	"github.com/google/uuid"
)

type NotificationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewNotificationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *NotificationLogic {
	return &NotificationLogic{ctx: ctx, svcCtx: svcCtx}
}

// ListNotifications returns customer notifications and unread count, auto-seeding starter notifications if empty
func (l *NotificationLogic) ListNotifications(customerIDStr string) (*types.NotificationListResponse, error) {
	custUUID, err := uuid.Parse(customerIDStr)
	if err != nil {
		return nil, xerr.NewErrMsg("شناسه مشتری نامعتبر است")
	}

	var count int64
	l.svcCtx.DB.WithContext(l.ctx).Model(&model.CustomerNotification{}).
		Where("customer_id = ?", custUUID).Count(&count)

	// Auto-seed welcome notifications for customer if first time viewing
	if count == 0 {
		starterNotifs := []model.CustomerNotification{
			{
				ID:         uuid.New(),
				CustomerID: custUUID,
				Title:      "خوش‌آمدید به خانواده آناهیکس!",
				Message:    "حساب کاربری شما با موفقیت فعال گردید. از این پس می‌توانید کلیه اشتراک‌های قانونی هوش مصنوعی و شماره‌های مجازی را با تحویل آنی تهیه فرمایید.",
				Type:       "system",
				Link:       "/profile?tab=overview",
				IsRead:     false,
				CreatedAt:  time.Now().Add(-10 * time.Minute),
			},
			{
				ID:         uuid.New(),
				CustomerID: custUUID,
				Title:      "ارتقای سطح به باشگاه مشتریان VIP",
				Message:    "سطح کاربری شما در باشگاه مشتریان آناهیکس ثبت شد. با هر خرید، از تخفیف‌های خودکار سیستمی بهره‌مند خواهید شد.",
				Type:       "loyalty",
				Link:       "/profile?tab=loyalty",
				IsRead:     false,
				CreatedAt:  time.Now().Add(-5 * time.Minute),
			},
			{
				ID:         uuid.New(),
				CustomerID: custUUID,
				Title:      "پروتکل امنیتی Zeroize فعال است",
				Message:    "تمامی اطلاعات هویتی و سشن‌های شما تحت استاندارد رمزنگاری AES-256 محافظت شده و هیچ‌گونه رمز عبوری در سرور باقی نخواهد ماند.",
				Type:       "security",
				Link:       "/profile?tab=security",
				IsRead:     false,
				CreatedAt:  time.Now().Add(-2 * time.Minute),
			},
		}

		for _, n := range starterNotifs {
			l.svcCtx.DB.WithContext(l.ctx).Create(&n)
		}
	}

	var notifs []model.CustomerNotification
	if err := l.svcCtx.DB.WithContext(l.ctx).
		Where("customer_id = ?", custUUID).
		Order("created_at desc").
		Find(&notifs).Error; err != nil {
		return nil, xerr.NewErrCode(xerr.DatabaseError)
	}

	var unreadCount int
	items := make([]types.CustomerNotificationDTO, len(notifs))
	for i, n := range notifs {
		if !n.IsRead {
			unreadCount++
		}
		items[i] = types.CustomerNotificationDTO{
			ID:        n.ID.String(),
			Title:     n.Title,
			Message:   n.Message,
			Type:      n.Type,
			Link:      n.Link,
			IsRead:    n.IsRead,
			CreatedAt: n.CreatedAt,
		}
	}

	return &types.NotificationListResponse{
		Total:       len(items),
		UnreadCount: unreadCount,
		Items:       items,
	}, nil
}

// MarkAsRead marks a single notification as read in database
func (l *NotificationLogic) MarkAsRead(customerIDStr string, notifIDStr string) error {
	custUUID, err := uuid.Parse(customerIDStr)
	if err != nil {
		return xerr.NewErrMsg("شناسه مشتری نامعتبر است")
	}
	notifUUID, err := uuid.Parse(notifIDStr)
	if err != nil {
		return xerr.NewErrMsg("شناسه اعلان نامعتبر است")
	}

	res := l.svcCtx.DB.WithContext(l.ctx).Model(&model.CustomerNotification{}).
		Where("id = ? AND customer_id = ?", notifUUID, custUUID).
		Update("is_read", true)

	if res.Error != nil {
		return xerr.NewErrCode(xerr.DatabaseError)
	}
	if res.RowsAffected == 0 {
		return errors.New("اعلان یافت نشد")
	}
	return nil
}

// MarkAllAsRead marks all notifications of customer as read
func (l *NotificationLogic) MarkAllAsRead(customerIDStr string) error {
	custUUID, err := uuid.Parse(customerIDStr)
	if err != nil {
		return xerr.NewErrMsg("شناسه مشتری نامعتبر است")
	}

	return l.svcCtx.DB.WithContext(l.ctx).Model(&model.CustomerNotification{}).
		Where("customer_id = ? AND is_read = ?", custUUID, false).
		Update("is_read", true).Error
}

// GetAdminNotificationStats returns platform notification metrics
func (l *NotificationLogic) GetAdminNotificationStats() (*types.NotificationStatsResponse, error) {
	var total, read, unread, recipients int64
	db := l.svcCtx.DB.WithContext(l.ctx).Model(&model.CustomerNotification{})

	db.Count(&total)
	l.svcCtx.DB.WithContext(l.ctx).Model(&model.CustomerNotification{}).Where("is_read = ?", true).Count(&read)
	l.svcCtx.DB.WithContext(l.ctx).Model(&model.CustomerNotification{}).Where("is_read = ?", false).Count(&unread)
	l.svcCtx.DB.WithContext(l.ctx).Model(&model.CustomerNotification{}).Distinct("customer_id").Count(&recipients)

	return &types.NotificationStatsResponse{
		TotalNotifications: total,
		TotalRead:          read,
		TotalUnread:        unread,
		ActiveRecipients:   recipients,
	}, nil
}
