package cmd

import (
	"encoding/json"
	"log"
	"time"

	"github.com/EbadiDev/anahix-server/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func seedReviewsData(db *gorm.DB) {
	var count int64
	db.Model(&model.Review{}).Count(&count)
	if count > 0 {
		return
	}

	var products []model.Product
	db.Find(&products)
	if len(products) == 0 {
		return
	}

	makeTags := func(tags []string) string {
		b, _ := json.Marshal(tags)
		return string(b)
	}

	now := time.Now()
	twoDaysAgo := now.Add(-48 * time.Hour)
	fiveDaysAgo := now.Add(-120 * time.Hour)
	tenDaysAgo := now.Add(-240 * time.Hour)
	oneHourAgo := now.Add(-1 * time.Hour)

	var reviewsToInsert []model.Review

	for _, prod := range products {
		if prod.Slug == "gemini-advanced-18m" || prod.Slug == "gemini-pro" || prod.GroupTag == "gemini" {
			reviewsToInsert = append(reviewsToInsert,
				model.Review{
					ID:              uuid.New(),
					ProductID:       prod.ID,
					ProductSlug:     prod.Slug,
					ProductTitle:    prod.TitleFA,
					OrderNumber:     "ORD-2410-8912",
					CustomerName:    "مهندس علیرضا کمالی",
					CustomerPhone:   "0912***4821",
					Rating:          5,
					Comment:         "من اشتراک جمینای ۱۸ ماهه رو تهیه کردم. تحویل دقیقاً در کمتر از ۳۰ ثانیه انجام شد و روی اکانت گوگل خودم فعال شد. سرعت ادونس فوق‌العاده بالاست و بدون هیچ فیلتر و دردسری کار می‌کنه.",
					ExperienceTags:  makeTags([]string{"تحویل زیر ۱ دقیقه", "کیفیت بدون قطعی", "فعالسازی سریع و آسان"}),
					Status:          model.ReviewStatusApproved,
					IsVerifiedBuyer: true,
					IsFeatured:      true,
					HelpfulCount:    18,
					AdminReply:      "درود بر شما علیرضا عزیز؛ از اعتماد چندین‌باره شما به آناهیکس سپاسگزاریم. رضایت شما و تجربه لذت‌بخش از هوش مصنوعی، مهم‌ترین هدف تیم ماست.",
					AdminReplyAuthor: "پشتیبانی رسمی آناهیکس",
					AdminRepliedAt:  &twoDaysAgo,
					CreatedAt:       fiveDaysAgo,
				},
				model.Review{
					ID:              uuid.New(),
					ProductID:       prod.ID,
					ProductSlug:     prod.Slug,
					ProductTitle:    prod.TitleFA,
					OrderNumber:     "ORD-2410-6721",
					CustomerName:    "مریم حسینی (طراح UI/UX)",
					CustomerPhone:   "0935***7712",
					Rating:          5,
					Comment:         "پشتیبانی تلگرام بسیار سریع و مودبانه راهنمایی کردن. برای تولید محتوا و کدنویسی واقعا عصای دستمه. قیمت نسبت به سایت‌های دیگه خیلی منصفانه‌تر بود.",
					ExperienceTags:  makeTags([]string{"پشتیبانی عالی و صبور", "قیمت منصفانه و رقابتی", "تضمین اصالت قانونی"}),
					Status:          model.ReviewStatusApproved,
					IsVerifiedBuyer: true,
					IsFeatured:      false,
					HelpfulCount:    9,
					CreatedAt:       tenDaysAgo,
				},
				model.Review{
					ID:              uuid.New(),
					ProductID:       prod.ID,
					ProductSlug:     prod.Slug,
					ProductTitle:    prod.TitleFA,
					OrderNumber:     "ORD-2410-9941",
					CustomerName:    "پویا نعمتی",
					CustomerPhone:   "0919***3345",
					Rating:          5,
					Comment:         "عالی و بدون قطعی. از سرعت فعال‌سازی شگفت‌زده شدم.",
					ExperienceTags:  makeTags([]string{"تحویل زیر ۱ دقیقه", "کیفیت بدون قطعی"}),
					Status:          model.ReviewStatusApproved,
					IsVerifiedBuyer: true,
					HelpfulCount:    4,
					CreatedAt:       twoDaysAgo,
				},
				// A pending review waiting for admin moderation
				model.Review{
					ID:              uuid.New(),
					ProductID:       prod.ID,
					ProductSlug:     prod.Slug,
					ProductTitle:    prod.TitleFA,
					OrderNumber:     "ORD-2410-1044",
					CustomerName:    "سهراب رضوانی",
					CustomerPhone:   "0936***9012",
					Rating:          5,
					Comment:         "سلام، آیا برای تمدید بعد از ۱۸ ماه هم همین آفر برقرار هست؟ در کل فعال‌سازی عالی بود ممنونم.",
					ExperienceTags:  makeTags([]string{"فعالسازی سریع و آسان", "پشتیبانی عالی و صبور"}),
					Status:          model.ReviewStatusPending,
					IsVerifiedBuyer: true,
					CreatedAt:       oneHourAgo,
				},
			)
		}

		if prod.Slug == "chatgpt-plus-personal" || prod.GroupTag == "chatgpt" {
			reviewsToInsert = append(reviewsToInsert,
				model.Review{
					ID:              uuid.New(),
					ProductID:       prod.ID,
					ProductSlug:     prod.Slug,
					ProductTitle:    prod.TitleFA,
					OrderNumber:     "ORD-2410-3321",
					CustomerName:    "دکتر حامد صادقی",
					CustomerPhone:   "0912***9911",
					Rating:          5,
					Comment:         "اکانت شخصی من رو بدون نیاز به پسورد طولانی و با تبادل لایو کد ۲FA فعال کردن. این شیوه امنیتی که به کار بردن واقعا خیالم رو راحت کرد. ممنون از اپراتور صبورتون.",
					ExperienceTags:  makeTags([]string{"پاسخگویی سریع ۲FA", "پشتیبانی عالی و صبور", "تضمین اصالت قانونی"}),
					Status:          model.ReviewStatusApproved,
					IsVerifiedBuyer: true,
					IsFeatured:      true,
					HelpfulCount:    24,
					AdminReply:      "سلام و درود حامد گرامی؛ حفظ حریم خصوصی و امنیت حساب کاربری شما اولویت درجه یک آناهیکس است. خوشحالیم که پروتکل تبادل زنده ۲FA رضایت خاطر شما را جلب نموده است.",
					AdminReplyAuthor: "پشتیبانی رسمی آناهیکس",
					AdminRepliedAt:  &fiveDaysAgo,
					CreatedAt:       tenDaysAgo,
				},
				model.Review{
					ID:              uuid.New(),
					ProductID:       prod.ID,
					ProductSlug:     prod.Slug,
					ProductTitle:    prod.TitleFA,
					OrderNumber:     "ORD-2410-5412",
					CustomerName:    "نیلوفر تهرانی",
					CustomerPhone:   "0918***1123",
					Rating:          5,
					Comment:         "سریع فعال شد و همه هیستوری چت‌هام حفظ شد. دستمریزاد.",
					ExperienceTags:  makeTags([]string{"فعالسازی سریع و آسان", "کیفیت بدون قطعی"}),
					Status:          model.ReviewStatusApproved,
					IsVerifiedBuyer: true,
					HelpfulCount:    11,
					CreatedAt:       fiveDaysAgo,
				},
				// Pending review
				model.Review{
					ID:              uuid.New(),
					ProductID:       prod.ID,
					ProductSlug:     prod.Slug,
					ProductTitle:    prod.TitleFA,
					OrderNumber:     "ORD-2410-7711",
					CustomerName:    "فرهاد افشار",
					CustomerPhone:   "0915***6644",
					Rating:          4,
					Comment:         "فعال‌سازی حدود ۲۰ دقیقه طول کشید که برای حالت دستی معقوله، کیفیت اکانت تا الان بدون نقص بوده.",
					ExperienceTags:  makeTags([]string{"کیفیت بدون قطعی", "تضمین اصالت قانونی"}),
					Status:          model.ReviewStatusPending,
					IsVerifiedBuyer: true,
					CreatedAt:       twoDaysAgo,
				},
			)
		}

		if prod.Slug == "claude-pro-personal" || prod.GroupTag == "claude" {
			reviewsToInsert = append(reviewsToInsert,
				model.Review{
					ID:              uuid.New(),
					ProductID:       prod.ID,
					ProductSlug:     prod.Slug,
					ProductTitle:    prod.TitleFA,
					OrderNumber:     "ORD-2410-4490",
					CustomerName:    "سعید ابراهیمی (توسعه‌دهنده فول‌استک)",
					CustomerPhone:   "0933***7788",
					Rating:          5,
					Comment:         "برای پروژه ری‌اکت و پایتون مدام با کلود Sonnet 3.5 کد می‌زنم. فعال‌سازی روی ایمیل شخصی خودم در عرض چند دقیقه انجام شد. پینگ و استبیلیتی عالیه.",
					ExperienceTags:  makeTags([]string{"تحویل زیر ۱ دقیقه", "کیفیت بدون قطعی", "فعالسازی سریع و آسان"}),
					Status:          model.ReviewStatusApproved,
					IsVerifiedBuyer: true,
					IsFeatured:      true,
					HelpfulCount:    15,
					AdminReply:      "سلام آقا سعید عزیز؛ امیدواریم همراهی کلود ۳.۵ سرعت توسعه پروژه‌های ارزشمندتان را چند برابر کند.",
					AdminReplyAuthor: "پشتیبانی رسمی آناهیکس",
					AdminRepliedAt:  &twoDaysAgo,
					CreatedAt:       fiveDaysAgo,
				},
			)
		}
	}

	for _, r := range reviewsToInsert {
		db.Create(&r)
	}

	if len(reviewsToInsert) > 0 {
		log.Printf("🌱 Seeded %d verified buyer reviews, social proof ratings, and admin replies\n", len(reviewsToInsert))
	}
}
