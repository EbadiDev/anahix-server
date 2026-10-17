package types

import "time"

type CustomerNotificationDTO struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Message   string    `json:"message"`
	Type      string    `json:"type"`
	Link      string    `json:"link,omitempty"`
	IsRead    bool      `json:"is_read"`
	CreatedAt time.Time `json:"created_at"`
}

type NotificationListResponse struct {
	Total       int                       `json:"total"`
	UnreadCount int                       `json:"unread_count"`
	Items       []CustomerNotificationDTO `json:"items"`
}

type NotificationStatsResponse struct {
	TotalNotifications int64 `json:"total_notifications"`
	TotalRead          int64 `json:"total_read"`
	TotalUnread        int64 `json:"total_unread"`
	ActiveRecipients   int64 `json:"active_recipients"`
}
