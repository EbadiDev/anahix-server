package notification

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/EbadiDev/anahix-server/internal/model"
	"github.com/EbadiDev/anahix-server/internal/svc"
)

type TicketNotifier struct {
	svcCtx *svc.ServiceContext
}

func NewTicketNotifier(svcCtx *svc.ServiceContext) *TicketNotifier {
	return &TicketNotifier{svcCtx: svcCtx}
}

// NotifyCustomerOnReply dispatches SMS and/or Email alerts to the customer when a ticket receives a reply
func (n *TicketNotifier) NotifyCustomerOnReply(ctx context.Context, ticket *model.Ticket, replySnippet string, adminName string, forceSMS bool, forceEmail bool) {
	if ticket == nil {
		return
	}

	shouldSendSMS := ticket.NotifyViaSMS || forceSMS
	shouldSendEmail := ticket.NotifyViaEmail || forceEmail

	// 1. Dispatch SMS Notification
	if shouldSendSMS && strings.TrimSpace(ticket.CustomerPhone) != "" {
		go func(phone string, tNum string, subj string) {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("⚠️ Panic recovered in ticket SMS notification: %v", r)
				}
			}()

			var settings []model.SystemSetting
			_ = n.svcCtx.DB.Where("category = ?", "sms").Find(&settings).Error

			configMap := make(map[string]string)
			for _, s := range settings {
				configMap[s.Key] = s.Value
			}

			apiKey := strings.TrimSpace(configMap["smsir_api_key"])
			lineNumStr := strings.TrimSpace(configMap["smsir_line_number"])
			lineNum, _ := strconv.ParseInt(lineNumStr, 10, 64)
			if lineNum <= 0 {
				lineNum = 30007732
			}

			smsText := fmt.Sprintf("آناهیکس: کاربر گرامی، پاسخ جدیدی برای تیکت پشتیبانی #%s (%s) ثبت شد.\nمشاهده پاسخ: anahix.ir/profile?tab=support", tNum, subj)

			_, err := n.svcCtx.SMSIRClient.SendBulk(context.Background(), apiKey, lineNum, smsText, []string{phone})
			if err != nil {
				log.Printf("❌ Failed to send ticket reply SMS to %s: %v", phone, err)
			} else {
				log.Printf("📱 Successfully dispatched ticket reply SMS notification to %s for ticket #%s", phone, tNum)
			}
		}(ticket.CustomerPhone, ticket.TicketNumber, ticket.Subject)
	}

	// 2. Dispatch Email Notification
	if shouldSendEmail && strings.TrimSpace(ticket.CustomerEmail) != "" {
		go func(email string, tNum string, subj string, snippet string) {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("⚠️ Panic recovered in ticket email notification: %v", r)
				}
			}()

			// Structured email alert simulation & logging
			log.Printf("📧 [TICKET EMAIL NOTIFICATION] To: %s | Subject: پاسخ جدید به تیکت #%s | Ticket: %s | Snippet: %s",
				email, tNum, subj, snippet)
		}(ticket.CustomerEmail, ticket.TicketNumber, ticket.Subject, replySnippet)
	}
}
