package logic

import (
	"context"
	"testing"

	"github.com/EbadiDev/anahix-server/internal/config"
	"github.com/EbadiDev/anahix-server/internal/logic/admin"
	"github.com/EbadiDev/anahix-server/internal/logic/public"
	"github.com/EbadiDev/anahix-server/internal/model"
	"github.com/EbadiDev/anahix-server/internal/svc"
	"github.com/EbadiDev/anahix-server/internal/types"
	"github.com/EbadiDev/anahix-server/pkg/crypto"
	"github.com/EbadiDev/anahix-server/pkg/orm"
	"github.com/stretchr/testify/assert"
)

func setupContentTestDB(t *testing.T) *svc.ServiceContext {
	db, err := orm.InitTestDB()
	assert.NoError(t, err)

	vault, err := crypto.NewVault("test-master-key-32bytes-secret!")
	assert.NoError(t, err)

	cfg := &config.Config{
		JWT: config.JWTConfig{
			AccessSecret:      "test-secret-content-1234567890",
			AccessExpireHours: 24,
		},
		Vault: config.VaultConfig{
			MasterKey: "test-master-key-32bytes-secret!",
		},
	}

	return &svc.ServiceContext{
		Config: cfg,
		DB:     db,
		Vault:  vault,
	}
}

func TestBlogArticleLifecycle(t *testing.T) {
	svcCtx := setupContentTestDB(t)
	ctx := context.Background()

	adminLogic := admin.NewContentAdminLogic(ctx, svcCtx)
	pubLogic := public.NewContentLogic(ctx, svcCtx)

	// 1. Create Article via Admin
	isPub := true
	createReq := &types.CreateArticleRequest{
		Title:        "راهنمای خرید جمینای ادونسد در ایران",
		Slug:         "how-to-buy-gemini-test",
		Excerpt:      "آموزش کامل خرید و فعال‌سازی اکانت جمینای ۱۸ ماهه",
		Content:      "## معرفی مدل Gemini 1.5 Pro\nمدل جمینای گوگل با پنجره کانتکست ۲ میلیون توکن ارائه شده است.\n\n### ویژگی‌های کلیدی\n- فضای ۲ ترابایت ابری\n- ادغام در داکس و جیمیل\n\n```bash\ncurl -X GET https://gemini.google.com\n```",
		CategorySlug: "tutorials",
		CategoryName: "آموزش و ترفندها",
		Tags:         []string{"جمینای", "گوگل", "خرید هوش مصنوعی"},
		AuthorName:   "تحریریه آناهیکس",
		IsFeatured:   true,
		IsPublished:  &isPub,
	}

	created, err := adminLogic.CreateArticle(createReq)
	assert.NoError(t, err)
	assert.NotEmpty(t, created.ID)
	assert.Equal(t, "how-to-buy-gemini-test", created.Slug)
	assert.GreaterOrEqual(t, created.ReadingTimeMinutes, 1)
	assert.True(t, len(created.TableOfContents) >= 2, "TOC should contain extracted headings")

	// 2. Fetch Public Article by Slug
	detail, err := pubLogic.GetArticleBySlug("how-to-buy-gemini-test")
	assert.NoError(t, err)
	assert.Equal(t, created.Title, detail.Title)
	assert.Equal(t, "آموزش و ترفندها", detail.CategoryName)

	// 3. Submit Reader Comment
	commentReq := &types.SubmitArticleCommentRequest{
		AuthorName: "علی رضایی",
		Email:      "ali@example.com",
		Content:    "بسیار مقاله عالی و کاربردی بود، تشکر از آناهیکس",
	}
	err = pubLogic.SubmitArticleComment("how-to-buy-gemini-test", commentReq, nil)
	assert.NoError(t, err)

	// Verify comment exists
	comments, err := adminLogic.ListComments(created.ID, "all")
	assert.NoError(t, err)
	assert.Len(t, comments, 1)
	assert.Equal(t, "علی رضایی", comments[0].AuthorName)

	// Admin Replies to Comment
	err = adminLogic.UpdateCommentStatus(comments[0].ID, "approved", "سلام علی عزیز، خوشحالیم که براتون مفید بود.")
	assert.NoError(t, err)

	// 4. Update Article
	newTitle := "راهنمای جامع خرید جمینای ادونسد در ایران (آپدیت جدید)"
	updated, err := adminLogic.UpdateArticle(created.ID, &types.UpdateArticleRequest{
		Title: newTitle,
	})
	assert.NoError(t, err)
	assert.Equal(t, newTitle, updated.Title)

	// 5. List Public Articles
	listResp, err := pubLogic.ListArticles(1, 10, "tutorials", "", "")
	assert.NoError(t, err)
	assert.Equal(t, int64(1), listResp.Total)
	assert.Equal(t, newTitle, listResp.Items[0].Title)

	// 6. Delete Article
	err = adminLogic.DeleteArticle(created.ID)
	assert.NoError(t, err)

	_, err = pubLogic.GetArticleBySlug("how-to-buy-gemini-test")
	assert.Error(t, err)
}

func TestHelpGuideLifecycleAndVoting(t *testing.T) {
	svcCtx := setupContentTestDB(t)
	ctx := context.Background()

	adminLogic := admin.NewContentAdminLogic(ctx, svcCtx)
	pubLogic := public.NewContentLogic(ctx, svcCtx)

	// 1. Create Help Guide via Admin
	isPub := true
	guideReq := &types.CreateHelpGuideRequest{
		Title:         "رفع خطای Access Denied در کلود",
		Slug:          "fix-claude-403-test",
		Excerpt:       "حل خطای ارور ۴۰۳ و تحریم در ورود به کلاود",
		Content:       "## علت ارور تحریم در کلود\nآنتروپیک دسترسی آی‌پی‌های ایران را مسدود می‌کند.\n\n### راه‌حل اول: بستن WebRTC\nاکستنشن WebRTC Control را فعال نمایید.",
		CategorySlug:  "troubleshooting",
		CategoryName:  "رفع ارورها و خطاهای تحریم",
		Icon:          "AlertTriangle",
		Difficulty:    "ساده",
		ServiceTarget: "claude",
		IsPublished:   &isPub,
	}

	created, err := adminLogic.CreateHelpGuide(guideReq)
	assert.NoError(t, err)
	assert.NotEmpty(t, created.ID)
	assert.Equal(t, "fix-claude-403-test", created.Slug)

	// 2. Public Detail
	guideDetail, err := pubLogic.GetHelpGuideBySlug("fix-claude-403-test")
	assert.NoError(t, err)
	assert.Equal(t, int64(0), guideDetail.HelpfulYes)

	// 3. User Votes Helpful (👍)
	err = pubLogic.VoteHelpGuide("fix-claude-403-test", true, "192.168.1.50")
	assert.NoError(t, err)

	// Prevent duplicate vote from same voterKey
	err = pubLogic.VoteHelpGuide("fix-claude-403-test", true, "192.168.1.50")
	assert.Error(t, err, "Duplicate vote should be rejected")

	// Another user votes Unhelpful (👎)
	err = pubLogic.VoteHelpGuide("fix-claude-403-test", false, "192.168.1.51")
	assert.NoError(t, err)

	// Verify vote counts in DB
	var g model.HelpGuide
	svcCtx.DB.Where("slug = ?", "fix-claude-403-test").First(&g)
	assert.Equal(t, int64(1), g.HelpfulYes)
	assert.Equal(t, int64(1), g.HelpfulNo)
}

func TestContentTaxonomyAndCategories(t *testing.T) {
	svcCtx := setupContentTestDB(t)
	ctx := context.Background()

	adminLogic := admin.NewContentAdminLogic(ctx, svcCtx)
	pubLogic := public.NewContentLogic(ctx, svcCtx)

	// 1. Create Category
	catReq := &types.CreateCategoryRequest{
		Name:        "اخبار هوش مصنوعی",
		Slug:        "ai-news-test",
		Type:        "blog",
		Icon:        "Sparkles",
		Description: "جدیدترین آپدیت‌های دنیای AI",
		SortOrder:   1,
	}
	cat, err := adminLogic.CreateCategory(catReq)
	assert.NoError(t, err)
	assert.Equal(t, "ai-news-test", cat.Slug)

	// 2. List Public Categories
	cats, err := pubLogic.ListCategories("blog")
	assert.NoError(t, err)
	assert.Len(t, cats, 1)
	assert.Equal(t, "اخبار هوش مصنوعی", cats[0].Name)

	// 3. Delete Category
	err = adminLogic.DeleteCategory(cat.ID)
	assert.NoError(t, err)

	catsAfter, err := pubLogic.ListCategories("blog")
	assert.NoError(t, err)
	assert.Len(t, catsAfter, 0)
}
