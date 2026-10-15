package cmd

import (
	"log"
	"time"

	"github.com/EbadiDev/anahix-server/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func seedTicketsData(db *gorm.DB) {
	var count int64
	db.Model(&model.Ticket{}).Count(&count)
	if count > 0 {
		return
	}

	var customer model.Customer
	if err := db.Where("phone = ?", "09928376377").First(&customer).Error; err != nil {
		// Fallback to any customer
		if err := db.First(&customer).Error; err != nil {
			customer = model.Customer{
				ID:        uuid.New(),
				Phone:     "09928376377",
				FullName:  "کاربر ویژه آناهیکس",
				Email:     "user@anahix.ir",
				IsActive:  true,
				CreatedAt: time.Now(),
			}
			db.Create(&customer)
		}
	}

	now := time.Now()
	twoHoursAgo := now.Add(-2 * time.Hour)
	oneHourAgo := now.Add(-1 * time.Hour)
	twentyMinsAgo := now.Add(-20 * time.Minute)

	ticket1ID := uuid.New()
	t1 := model.Ticket{
		ID:                  ticket1ID,
		TicketNumber:        "TK-78421",
		CustomerID:          customer.ID,
		CustomerPhone:       customer.Phone,
		CustomerName:        customer.FullName,
		CustomerEmail:       customer.Email,
		Subject:             "راهنمایی در مورد فعال‌سازی لایسنس ChatGPT Plus",
		Department:          string(model.DepartmentLicenseSupport),
		Priority:            model.TicketPriorityHigh,
		Status:              model.TicketStatusResolved,
		RelatedOrderNo:      "ORD-2410-8912",
		NotifyViaSMS:        true,
		NotifyViaEmail:      true,
		SatisfactionRating:  5,
		SatisfactionComment: "پاسخگویی فوق‌العاده سریع و دقیق کمتر از ۱۰ دقیقه",
		LastReplyAt:         &twentyMinsAgo,
		LastReplyBy:         "customer",
		CreatedAt:           twoHoursAgo,
		UpdatedAt:           twentyMinsAgo,
	}

	m1 := model.TicketMessage{
		ID:         uuid.New(),
		TicketID:   ticket1ID,
		SenderType: "customer",
		SenderID:   customer.ID.String(),
		SenderName: customer.FullName,
		Message:    "سلام، وقت بخیر. من سفارش اشتراک چت جی‌پی‌تی پلاس را ثبت کردم، اما هنگام لاگین در سایت OpenAI با خطای اعتبارسنجی مواجه می‌شوم. اسکرین‌شات پیام خطا را پیوست می‌کنم. ممنون می‌شم راهنمایی بفرمایید.",
		CreatedAt:  twoHoursAgo,
		UpdatedAt:  twoHoursAgo,
	}
	m1.SetAttachments([]model.TicketAttachment{
		{
			ID:       "att-sample-1",
			FileName: "openai-login-error.png",
			FileURL:  "/uploads/tickets/openai-login-error.png",
			FileSize: 148200,
			MimeType: "image/png",
			IsImage:  true,
		},
	})

	m2 := model.TicketMessage{
		ID:         uuid.New(),
		TicketID:   ticket1ID,
		SenderType: "admin",
		SenderID:   "admin",
		SenderName: "پشتیبانی فنی آناهیکس",
		Message:    "درود بر شما کاربر گرامی؛ پیام خطای شما بررسی شد. این خطا مربوط به باقی‌ماندن کش سشن قبلی مرورگر است. لطفاً یک‌بار کش مرورگر خود را پاک نموده یا در پنجره ناشناس (Incognito Window) مجدداً لاگین فرمایید. لایسنس شما فعال و بدون محدودیت است.",
		CreatedAt:  oneHourAgo,
		UpdatedAt:  oneHourAgo,
	}

	m3 := model.TicketMessage{
		ID:         uuid.New(),
		TicketID:   ticket1ID,
		SenderType: "customer",
		SenderID:   customer.ID.String(),
		SenderName: customer.FullName,
		Message:    "خیلی ممنون، در پنجره ناشناس تست کردم و بلافاصله وارد شد و اشتراک پلاس فعال بود. سپاس از پیگیری و پشتیبانی عالی شما.",
		CreatedAt:  twentyMinsAgo,
		UpdatedAt:  twentyMinsAgo,
	}

	db.Create(&t1)
	db.Create(&m1)
	db.Create(&m2)
	db.Create(&m3)

	// Second sample ticket: Active awaiting admin response
	ticket2ID := uuid.New()
	t2 := model.Ticket{
		ID:             ticket2ID,
		TicketNumber:   "TK-91204",
		CustomerID:     customer.ID,
		CustomerPhone:  customer.Phone,
		CustomerName:   customer.FullName,
		CustomerEmail:  customer.Email,
		Subject:        "درخواست استعلام وضعیت شماره مجازی تلگرام",
		Department:     string(model.DepartmentVirtualNumbers),
		Priority:       model.TicketPriorityMedium,
		Status:         model.TicketStatusAwaitingAdmin,
		RelatedOrderNo: "VN-89214",
		NotifyViaSMS:   true,
		NotifyViaEmail: true,
		LastReplyAt:    &now,
		LastReplyBy:    "customer",
		CreatedAt:      now.Add(-15 * time.Minute),
		UpdatedAt:      now,
	}

	m2_1 := model.TicketMessage{
		ID:         uuid.New(),
		TicketID:   ticket2ID,
		SenderType: "customer",
		SenderID:   customer.ID.String(),
		SenderName: customer.FullName,
		Message:    "سلام، کد پیامک برای شماره مجازی تلگرام ارسال نشد و زمان آن تمام شد. آیا هزینه به صورت خودکار به کیف پول برگشت داده می‌شود؟",
		CreatedAt:  now.Add(-15 * time.Minute),
		UpdatedAt:  now.Add(-15 * time.Minute),
	}

	db.Create(&t2)
	db.Create(&m2_1)

	log.Println("🌱 Seeded initial support tickets and conversation threads into database")
}
