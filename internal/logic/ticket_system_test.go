package logic

import (
	"context"
	"testing"

	"github.com/EbadiDev/anahix-server/internal/logic/admin"
	"github.com/EbadiDev/anahix-server/internal/logic/provider/smsir"
	"github.com/EbadiDev/anahix-server/internal/logic/public"
	"github.com/EbadiDev/anahix-server/internal/model"
	"github.com/EbadiDev/anahix-server/internal/svc"
	"github.com/EbadiDev/anahix-server/internal/types"
	"github.com/EbadiDev/anahix-server/pkg/crypto"
	"github.com/EbadiDev/anahix-server/pkg/orm"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func setupTestTicketSystem(t *testing.T) (*svc.ServiceContext, model.Customer) {
	db, err := orm.InitTestDB()
	assert.NoError(t, err)

	vault, err := crypto.NewVault("test-master-key-32bytes-secret!")
	assert.NoError(t, err)

	svcCtx := &svc.ServiceContext{
		DB:          db,
		Vault:       vault,
		SMSIRClient: smsir.NewClient(),
	}

	cust := model.Customer{
		ID:        uuid.New(),
		Phone:     "09129998877",
		FullName:  "کاربر تستی تیکت",
		Email:     "tester@anahix.ir",
		IsActive:  true,
	}
	assert.NoError(t, db.Create(&cust).Error)

	return svcCtx, cust
}

func TestCompleteTicketFlow(t *testing.T) {
	svcCtx, cust := setupTestTicketSystem(t)
	ctx := context.Background()

	pubLogic := public.NewTicketLogic(ctx, svcCtx)
	admLogic := admin.NewTicketAdminLogic(ctx, svcCtx)

	// 1. Customer creates a ticket with attachment
	createReq := &types.CreateTicketRequest{
		Subject:        "خطای فعال‌سازی لایسنس پلاس",
		Department:     string(model.DepartmentLicenseSupport),
		Priority:       string(model.TicketPriorityHigh),
		RelatedOrderNo: "ORD-2026-TEST",
		Message:        "لایسنس کار نمی‌کند. اسکرین‌شات خطا پیوست شد.",
		Attachments: []types.TicketAttachmentDTO{
			{
				ID:       "att-1",
				FileName: "error-screenshot.png",
				FileURL:  "/uploads/tickets/error-screenshot.png",
				FileSize: 54000,
				MimeType: "image/png",
				IsImage:  true,
			},
		},
	}

	created, err := pubLogic.CreateTicket(createReq, cust.ID.String())
	assert.NoError(t, err)
	assert.NotEmpty(t, created.Ticket.ID)
	assert.Equal(t, "خطای فعال‌سازی لایسنس پلاس", created.Ticket.Subject)
	assert.Equal(t, string(model.TicketStatusAwaitingAdmin), created.Ticket.Status)
	assert.Equal(t, 1, len(created.Messages))
	assert.Equal(t, 1, len(created.Messages[0].Attachments))
	assert.Equal(t, "error-screenshot.png", created.Messages[0].Attachments[0].FileName)

	ticketID := created.Ticket.ID

	// 2. Customer lists my tickets
	myTickets, err := pubLogic.ListMyTickets(cust.ID.String())
	assert.NoError(t, err)
	assert.Len(t, myTickets, 1)
	assert.Equal(t, ticketID, myTickets[0].ID)

	// 3. Admin lists tickets with search/filter
	adminList, err := admLogic.ListTickets(&types.AdminTicketListRequest{
		Status: "awaiting_admin",
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(1), adminList.Total)
	assert.Equal(t, int64(1), adminList.Stats.AwaitingAdmin)

	// 4. Admin leaves an internal note (is_internal = true)
	internalNoteReq := &types.AdminReplyTicketRequest{
		Message:    "اپراتور ۱: لایسنس مشتری در پرتال کلود ریستارت شد.",
		IsInternal: true,
	}
	noteMsg, err := admLogic.ReplyTicket(ticketID, internalNoteReq, "admin_operator")
	assert.NoError(t, err)
	assert.True(t, noteMsg.IsInternal)

	// 5. Verify customer CANNOT see the internal note
	customerDetail, err := pubLogic.GetTicketDetail(ticketID, cust.ID.String())
	assert.NoError(t, err)
	assert.Len(t, customerDetail.Messages, 1, "Customer should only see 1 public message, internal note must be hidden")

	// 6. Admin can see both the public message and internal note
	adminDetail, err := admLogic.GetTicketDetail(ticketID)
	assert.NoError(t, err)
	assert.Len(t, adminDetail.Messages, 2, "Admin must see both public messages and internal staff notes")

	// 7. Admin sends a public reply with SMS notification trigger
	adminReplyReq := &types.AdminReplyTicketRequest{
		Message:    "کاربر گرامی، مشکل رفع گردید. لطفاً مجددا بررسی فرمایید.",
		IsInternal: false,
		Status:     string(model.TicketStatusAwaitingUser),
		NotifySMS:  true,
	}
	replyMsg, err := admLogic.ReplyTicket(ticketID, adminReplyReq, "admin_operator")
	assert.NoError(t, err)
	assert.False(t, replyMsg.IsInternal)

	// Verify customer now sees 2 messages (ticket opening + admin reply)
	customerDetailAfterReply, err := pubLogic.GetTicketDetail(ticketID, cust.ID.String())
	assert.NoError(t, err)
	assert.Len(t, customerDetailAfterReply.Messages, 2)
	assert.Equal(t, string(model.TicketStatusAwaitingUser), customerDetailAfterReply.Ticket.Status)

	// 8. Customer rates the ticket
	rateReq := &types.RateTicketRequest{
		Rating:  5,
		Comment: "عالی بود، خیلی سریع حل شد",
	}
	err = pubLogic.RateTicket(ticketID, rateReq, cust.ID.String())
	assert.NoError(t, err)

	// 9. Customer closes the ticket
	err = pubLogic.CloseTicket(ticketID, cust.ID.String())
	assert.NoError(t, err)

	finalDetail, err := pubLogic.GetTicketDetail(ticketID, cust.ID.String())
	assert.NoError(t, err)
	assert.Equal(t, string(model.TicketStatusClosed), finalDetail.Ticket.Status)
	assert.Equal(t, 5, finalDetail.Ticket.SatisfactionRating)
	assert.Equal(t, "عالی بود، خیلی سریع حل شد", finalDetail.Ticket.SatisfactionComment)
}
