package admin

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"strings"
	"time"

	"github.com/EbadiDev/anahix-server/internal/model"
	"github.com/EbadiDev/anahix-server/internal/svc"
	"github.com/EbadiDev/anahix-server/internal/types"
	"github.com/EbadiDev/anahix-server/pkg/xerr"
	"github.com/google/uuid"
)

type ContentAdminLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewContentAdminLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ContentAdminLogic {
	return &ContentAdminLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// ---------------- BLOG ARTICLES MANAGEMENT ----------------

// ListArticlesAdmin returns all articles (published & draft) with search and filters
func (l *ContentAdminLogic) ListArticlesAdmin(page, pageSize int, search, categorySlug string, isPublished *bool) (*types.ContentListResponse[types.ArticleListItemDTO], error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	query := l.svcCtx.DB.WithContext(l.ctx).Model(&model.Article{})

	if isPublished != nil {
		query = query.Where("is_published = ?", *isPublished)
	}

	if categorySlug != "" && categorySlug != "all" {
		query = query.Where("category_slug = ?", categorySlug)
	}

	if strings.TrimSpace(search) != "" {
		s := "%" + strings.TrimSpace(search) + "%"
		query = query.Where("title LIKE ? OR slug LIKE ? OR excerpt LIKE ?", s, s, s)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, xerr.NewErrCode(xerr.DatabaseError)
	}

	offset := (page - 1) * pageSize
	var articles []model.Article
	if err := query.Order("created_at desc").Offset(offset).Limit(pageSize).Find(&articles).Error; err != nil {
		return nil, xerr.NewErrCode(xerr.DatabaseError)
	}

	items := make([]types.ArticleListItemDTO, len(articles))
	for i, a := range articles {
		items[i] = types.ArticleListItemDTO{
			ID:                  a.ID.String(),
			Title:               a.Title,
			Slug:                a.Slug,
			Excerpt:             a.Excerpt,
			CoverImage:          a.CoverImage,
			CategorySlug:        a.CategorySlug,
			CategoryName:        a.CategoryName,
			Tags:                a.GetTagsSlice(),
			AuthorName:          a.AuthorName,
			AuthorRole:          a.AuthorRole,
			AuthorAvatar:        a.AuthorAvatar,
			ReadingTimeMinutes:  a.ReadingTimeMinutes,
			ViewCount:           a.ViewCount,
			LikeCount:           a.LikeCount,
			IsFeatured:          a.IsFeatured,
			IsPublished:         a.IsPublished,
			PublishedAt:         a.PublishedAt,
			RelatedProductTitle: a.RelatedProductTitle,
			RelatedProductSlug:  a.RelatedProductSlug,
			CreatedAt:           a.CreatedAt,
		}
	}

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))

	return &types.ContentListResponse[types.ArticleListItemDTO]{
		Items:      items,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// GetArticleAdmin returns complete article data for admin editing
func (l *ContentAdminLogic) GetArticleAdmin(id string) (*types.ArticleDetailDTO, error) {
	uID, err := uuid.Parse(id)
	if err != nil {
		return nil, xerr.NewErrMsg("شناسه مقاله نامعتبر است")
	}

	var a model.Article
	if err := l.svcCtx.DB.WithContext(l.ctx).Where("id = ?", uID).First(&a).Error; err != nil {
		return nil, xerr.NewErrMsg("مقاله مورد نظر یافت نشد")
	}

	var toc []types.TOCItemDTO
	if a.TableOfContentsJSON != "" {
		_ = json.Unmarshal([]byte(a.TableOfContentsJSON), &toc)
	}

	var comments []model.ArticleComment
	l.svcCtx.DB.WithContext(l.ctx).Where("article_id = ?", a.ID).Order("created_at desc").Find(&comments)

	commentDTOs := make([]types.ArticleCommentDTO, len(comments))
	for i, c := range comments {
		commentDTOs[i] = types.ArticleCommentDTO{
			ID:         c.ID.String(),
			ArticleID:  c.ArticleID.String(),
			AuthorName: c.AuthorName,
			Content:    c.Content,
			Status:     c.Status,
			AdminReply: c.AdminReply,
			CreatedAt:  c.CreatedAt,
		}
	}

	var relPID string
	if a.RelatedProductID != nil {
		relPID = a.RelatedProductID.String()
	}

	return &types.ArticleDetailDTO{
		ID:                  a.ID.String(),
		Title:               a.Title,
		Slug:                a.Slug,
		Excerpt:             a.Excerpt,
		Content:             a.Content,
		CoverImage:          a.CoverImage,
		CategorySlug:        a.CategorySlug,
		CategoryName:        a.CategoryName,
		Tags:                a.GetTagsSlice(),
		AuthorName:          a.AuthorName,
		AuthorRole:          a.AuthorRole,
		AuthorAvatar:        a.AuthorAvatar,
		ReadingTimeMinutes:  a.ReadingTimeMinutes,
		ViewCount:           a.ViewCount,
		LikeCount:           a.LikeCount,
		IsFeatured:          a.IsFeatured,
		IsPublished:         a.IsPublished,
		PublishedAt:         a.PublishedAt,
		RelatedProductID:    relPID,
		RelatedProductTitle: a.RelatedProductTitle,
		RelatedProductSlug:  a.RelatedProductSlug,
		RelatedProductPrice: a.RelatedProductPrice,
		MetaTitle:           a.MetaTitle,
		MetaDescription:     a.MetaDescription,
		MetaKeywords:        a.MetaKeywords,
		CanonicalURL:        a.CanonicalURL,
		TableOfContents:     toc,
		Comments:            commentDTOs,
		CreatedAt:           a.CreatedAt,
		UpdatedAt:           a.UpdatedAt,
	}, nil
}

// CreateArticle inserts a new article with auto calculated reading time and TOC
func (l *ContentAdminLogic) CreateArticle(req *types.CreateArticleRequest) (*types.ArticleDetailDTO, error) {
	slug := strings.TrimSpace(req.Slug)
	if slug == "" {
		slug = generateSlug(req.Title)
	}

	// Check slug uniqueness
	var count int64
	l.svcCtx.DB.WithContext(l.ctx).Model(&model.Article{}).Where("slug = ?", slug).Count(&count)
	if count > 0 {
		slug = fmt.Sprintf("%s-%d", slug, time.Now().Unix()%10000)
	}

	readingTime := req.ReadingTimeMinutes
	if readingTime <= 0 {
		readingTime = model.CalculateReadingTime(req.Content)
	}

	tocJSON := req.TableOfContentsJSON
	if strings.TrimSpace(tocJSON) == "" || tocJSON == "[]" {
		autoTOC := extractTOCFromAdminContent(req.Content)
		if b, err := json.Marshal(autoTOC); err == nil {
			tocJSON = string(b)
		}
	}

	tagsJSON := "[]"
	if len(req.Tags) > 0 {
		if b, err := json.Marshal(req.Tags); err == nil {
			tagsJSON = string(b)
		}
	}

	isPub := true
	if req.IsPublished != nil {
		isPub = *req.IsPublished
	}

	var publishedAt *time.Time
	if isPub {
		now := time.Now()
		publishedAt = &now
	}

	var relPID *uuid.UUID
	if req.RelatedProductID != "" {
		if pID, err := uuid.Parse(req.RelatedProductID); err == nil {
			relPID = &pID
		}
	}

	authorName := strings.TrimSpace(req.AuthorName)
	if authorName == "" {
		authorName = "تحریریه آناهیکس"
	}
	authorRole := strings.TrimSpace(req.AuthorRole)
	if authorRole == "" {
		authorRole = "کارشناس هوش مصنوعی و امنیت"
	}

	metaTitle := req.MetaTitle
	if metaTitle == "" {
		metaTitle = req.Title
	}
	metaDesc := req.MetaDescription
	if metaDesc == "" {
		metaDesc = req.Excerpt
	}

	article := model.Article{
		Title:               strings.TrimSpace(req.Title),
		Slug:                slug,
		Excerpt:             strings.TrimSpace(req.Excerpt),
		Content:             req.Content,
		CoverImage:          strings.TrimSpace(req.CoverImage),
		CategorySlug:        strings.TrimSpace(req.CategorySlug),
		CategoryName:        strings.TrimSpace(req.CategoryName),
		Tags:                tagsJSON,
		AuthorName:          authorName,
		AuthorRole:          authorRole,
		AuthorAvatar:        strings.TrimSpace(req.AuthorAvatar),
		ReadingTimeMinutes:  readingTime,
		IsFeatured:          req.IsFeatured,
		IsPublished:         isPub,
		PublishedAt:         publishedAt,
		RelatedProductID:    relPID,
		RelatedProductTitle: strings.TrimSpace(req.RelatedProductTitle),
		RelatedProductSlug:  strings.TrimSpace(req.RelatedProductSlug),
		RelatedProductPrice: req.RelatedProductPrice,
		MetaTitle:           metaTitle,
		MetaDescription:     metaDesc,
		MetaKeywords:        strings.TrimSpace(req.MetaKeywords),
		CanonicalURL:        strings.TrimSpace(req.CanonicalURL),
		TableOfContentsJSON: tocJSON,
	}

	if err := l.svcCtx.DB.WithContext(l.ctx).Create(&article).Error; err != nil {
		return nil, xerr.NewErrCode(xerr.DatabaseError)
	}

	return l.GetArticleAdmin(article.ID.String())
}

// UpdateArticle updates an existing article
func (l *ContentAdminLogic) UpdateArticle(id string, req *types.UpdateArticleRequest) (*types.ArticleDetailDTO, error) {
	uID, err := uuid.Parse(id)
	if err != nil {
		return nil, xerr.NewErrMsg("شناسه مقاله نامعتبر است")
	}

	var a model.Article
	if err := l.svcCtx.DB.WithContext(l.ctx).Where("id = ?", uID).First(&a).Error; err != nil {
		return nil, xerr.NewErrMsg("مقاله مورد نظر یافت نشد")
	}

	updates := map[string]interface{}{}

	if req.Title != "" {
		updates["title"] = strings.TrimSpace(req.Title)
	}
	if req.Slug != "" && req.Slug != a.Slug {
		updates["slug"] = strings.TrimSpace(req.Slug)
	}
	if req.Excerpt != "" {
		updates["excerpt"] = strings.TrimSpace(req.Excerpt)
	}
	if req.Content != "" {
		updates["content"] = req.Content
		if req.ReadingTimeMinutes <= 0 {
			updates["reading_time_minutes"] = model.CalculateReadingTime(req.Content)
		}
		// update TOC
		autoTOC := extractTOCFromAdminContent(req.Content)
		if b, err := json.Marshal(autoTOC); err == nil {
			updates["table_of_contents_json"] = string(b)
		}
	}
	if req.CoverImage != "" {
		updates["cover_image"] = strings.TrimSpace(req.CoverImage)
	}
	if req.CategorySlug != "" {
		updates["category_slug"] = strings.TrimSpace(req.CategorySlug)
	}
	if req.CategoryName != "" {
		updates["category_name"] = strings.TrimSpace(req.CategoryName)
	}
	if len(req.Tags) > 0 {
		if b, err := json.Marshal(req.Tags); err == nil {
			updates["tags"] = string(b)
		}
	}
	if req.AuthorName != "" {
		updates["author_name"] = strings.TrimSpace(req.AuthorName)
	}
	if req.AuthorRole != "" {
		updates["author_role"] = strings.TrimSpace(req.AuthorRole)
	}
	if req.AuthorAvatar != "" {
		updates["author_avatar"] = strings.TrimSpace(req.AuthorAvatar)
	}
	if req.ReadingTimeMinutes > 0 {
		updates["reading_time_minutes"] = req.ReadingTimeMinutes
	}
	if req.IsFeatured != nil {
		updates["is_featured"] = *req.IsFeatured
	}
	if req.IsPublished != nil {
		updates["is_published"] = *req.IsPublished
		if *req.IsPublished && a.PublishedAt == nil {
			now := time.Now()
			updates["published_at"] = &now
		}
	}
	if req.RelatedProductID != "" {
		if pID, err := uuid.Parse(req.RelatedProductID); err == nil {
			updates["related_product_id"] = pID
		}
	}
	if req.RelatedProductTitle != "" {
		updates["related_product_title"] = strings.TrimSpace(req.RelatedProductTitle)
	}
	if req.RelatedProductSlug != "" {
		updates["related_product_slug"] = strings.TrimSpace(req.RelatedProductSlug)
	}
	if req.RelatedProductPrice > 0 {
		updates["related_product_price"] = req.RelatedProductPrice
	}
	if req.MetaTitle != "" {
		updates["meta_title"] = strings.TrimSpace(req.MetaTitle)
	}
	if req.MetaDescription != "" {
		updates["meta_description"] = strings.TrimSpace(req.MetaDescription)
	}
	if req.MetaKeywords != "" {
		updates["meta_keywords"] = strings.TrimSpace(req.MetaKeywords)
	}
	if req.CanonicalURL != "" {
		updates["canonical_url"] = strings.TrimSpace(req.CanonicalURL)
	}
	if req.TableOfContentsJSON != "" {
		updates["table_of_contents_json"] = req.TableOfContentsJSON
	}

	updates["updated_at"] = time.Now()

	if err := l.svcCtx.DB.WithContext(l.ctx).Model(&a).Updates(updates).Error; err != nil {
		return nil, xerr.NewErrCode(xerr.DatabaseError)
	}

	return l.GetArticleAdmin(id)
}

// DeleteArticle removes an article
func (l *ContentAdminLogic) DeleteArticle(id string) error {
	uID, err := uuid.Parse(id)
	if err != nil {
		return xerr.NewErrMsg("شناسه مقاله نامعتبر است")
	}

	if err := l.svcCtx.DB.WithContext(l.ctx).Where("id = ?", uID).Delete(&model.Article{}).Error; err != nil {
		return xerr.NewErrCode(xerr.DatabaseError)
	}
	return nil
}

// ---------------- HELP GUIDES / KNOWLEDGE BASE MANAGEMENT ----------------

// ListHelpGuidesAdmin returns all knowledge base guides for admin
func (l *ContentAdminLogic) ListHelpGuidesAdmin(page, pageSize int, search, categorySlug, serviceTarget string) (*types.ContentListResponse[types.HelpGuideListItemDTO], error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	query := l.svcCtx.DB.WithContext(l.ctx).Model(&model.HelpGuide{})

	if categorySlug != "" && categorySlug != "all" {
		query = query.Where("category_slug = ?", categorySlug)
	}

	if serviceTarget != "" && serviceTarget != "all" {
		query = query.Where("service_target = ?", serviceTarget)
	}

	if strings.TrimSpace(search) != "" {
		s := "%" + strings.TrimSpace(search) + "%"
		query = query.Where("title LIKE ? OR slug LIKE ? OR excerpt LIKE ?", s, s, s)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, xerr.NewErrCode(xerr.DatabaseError)
	}

	offset := (page - 1) * pageSize
	var guides []model.HelpGuide
	if err := query.Order("sort_order asc, created_at desc").Offset(offset).Limit(pageSize).Find(&guides).Error; err != nil {
		return nil, xerr.NewErrCode(xerr.DatabaseError)
	}

	items := make([]types.HelpGuideListItemDTO, len(guides))
	for i, g := range guides {
		items[i] = types.HelpGuideListItemDTO{
			ID:                  g.ID.String(),
			Title:               g.Title,
			Slug:                g.Slug,
			Excerpt:             g.Excerpt,
			CategorySlug:        g.CategorySlug,
			CategoryName:        g.CategoryName,
			Icon:                g.Icon,
			Difficulty:          g.Difficulty,
			ServiceTarget:       g.ServiceTarget,
			ReadingTimeMinutes:  g.ReadingTimeMinutes,
			ViewCount:           g.ViewCount,
			HelpfulYes:          g.HelpfulYes,
			HelpfulNo:           g.HelpfulNo,
			IsFeatured:          g.IsFeatured,
			IsPublished:         g.IsPublished,
			PublishedAt:         g.PublishedAt,
			RelatedProductTitle: g.RelatedProductTitle,
			RelatedProductSlug:  g.RelatedProductSlug,
			CreatedAt:           g.CreatedAt,
		}
	}

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))

	return &types.ContentListResponse[types.HelpGuideListItemDTO]{
		Items:      items,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// GetHelpGuideAdmin returns single guide for editing
func (l *ContentAdminLogic) GetHelpGuideAdmin(id string) (*types.HelpGuideDetailDTO, error) {
	uID, err := uuid.Parse(id)
	if err != nil {
		return nil, xerr.NewErrMsg("شناسه راهنما نامعتبر است")
	}

	var g model.HelpGuide
	if err := l.svcCtx.DB.WithContext(l.ctx).Where("id = ?", uID).First(&g).Error; err != nil {
		return nil, xerr.NewErrMsg("راهنمای مورد نظر یافت نشد")
	}

	var toc []types.TOCItemDTO
	if g.TableOfContentsJSON != "" {
		_ = json.Unmarshal([]byte(g.TableOfContentsJSON), &toc)
	}

	var relPID string
	if g.RelatedProductID != nil {
		relPID = g.RelatedProductID.String()
	}

	return &types.HelpGuideDetailDTO{
		ID:                  g.ID.String(),
		Title:               g.Title,
		Slug:                g.Slug,
		Excerpt:             g.Excerpt,
		Content:             g.Content,
		CategorySlug:        g.CategorySlug,
		CategoryName:        g.CategoryName,
		Icon:                g.Icon,
		Difficulty:          g.Difficulty,
		ServiceTarget:       g.ServiceTarget,
		ReadingTimeMinutes:  g.ReadingTimeMinutes,
		ViewCount:           g.ViewCount,
		HelpfulYes:          g.HelpfulYes,
		HelpfulNo:           g.HelpfulNo,
		IsFeatured:          g.IsFeatured,
		IsPublished:         g.IsPublished,
		PublishedAt:         g.PublishedAt,
		RelatedProductID:    relPID,
		RelatedProductTitle: g.RelatedProductTitle,
		RelatedProductSlug:  g.RelatedProductSlug,
		RelatedProductPrice: g.RelatedProductPrice,
		MetaTitle:           g.MetaTitle,
		MetaDescription:     g.MetaDescription,
		MetaKeywords:        g.MetaKeywords,
		TableOfContents:     toc,
		CreatedAt:           g.CreatedAt,
		UpdatedAt:           g.UpdatedAt,
	}, nil
}

// CreateHelpGuide adds a new guide
func (l *ContentAdminLogic) CreateHelpGuide(req *types.CreateHelpGuideRequest) (*types.HelpGuideDetailDTO, error) {
	slug := strings.TrimSpace(req.Slug)
	if slug == "" {
		slug = generateSlug(req.Title)
	}

	var count int64
	l.svcCtx.DB.WithContext(l.ctx).Model(&model.HelpGuide{}).Where("slug = ?", slug).Count(&count)
	if count > 0 {
		slug = fmt.Sprintf("%s-%d", slug, time.Now().Unix()%10000)
	}

	readingTime := req.ReadingTimeMinutes
	if readingTime <= 0 {
		readingTime = model.CalculateReadingTime(req.Content)
	}

	tocJSON := req.TableOfContentsJSON
	if strings.TrimSpace(tocJSON) == "" || tocJSON == "[]" {
		autoTOC := extractTOCFromAdminContent(req.Content)
		if b, err := json.Marshal(autoTOC); err == nil {
			tocJSON = string(b)
		}
	}

	isPub := true
	if req.IsPublished != nil {
		isPub = *req.IsPublished
	}

	var publishedAt *time.Time
	if isPub {
		now := time.Now()
		publishedAt = &now
	}

	var relPID *uuid.UUID
	if req.RelatedProductID != "" {
		if pID, err := uuid.Parse(req.RelatedProductID); err == nil {
			relPID = &pID
		}
	}

	difficulty := req.Difficulty
	if difficulty == "" {
		difficulty = "ساده"
	}

	serviceTarget := req.ServiceTarget
	if serviceTarget == "" {
		serviceTarget = "general"
	}

	guide := model.HelpGuide{
		Title:               strings.TrimSpace(req.Title),
		Slug:                slug,
		Excerpt:             strings.TrimSpace(req.Excerpt),
		Content:             req.Content,
		CategorySlug:        strings.TrimSpace(req.CategorySlug),
		CategoryName:        strings.TrimSpace(req.CategoryName),
		Icon:                strings.TrimSpace(req.Icon),
		Difficulty:          difficulty,
		ServiceTarget:       serviceTarget,
		ReadingTimeMinutes:  readingTime,
		SortOrder:           req.SortOrder,
		IsFeatured:          req.IsFeatured,
		IsPublished:         isPub,
		PublishedAt:         publishedAt,
		RelatedProductID:    relPID,
		RelatedProductTitle: strings.TrimSpace(req.RelatedProductTitle),
		RelatedProductSlug:  strings.TrimSpace(req.RelatedProductSlug),
		RelatedProductPrice: req.RelatedProductPrice,
		MetaTitle:           req.MetaTitle,
		MetaDescription:     req.MetaDescription,
		MetaKeywords:        strings.TrimSpace(req.MetaKeywords),
		TableOfContentsJSON: tocJSON,
	}

	if err := l.svcCtx.DB.WithContext(l.ctx).Create(&guide).Error; err != nil {
		return nil, xerr.NewErrCode(xerr.DatabaseError)
	}

	return l.GetHelpGuideAdmin(guide.ID.String())
}

// UpdateHelpGuide modifies an existing guide
func (l *ContentAdminLogic) UpdateHelpGuide(id string, req *types.UpdateHelpGuideRequest) (*types.HelpGuideDetailDTO, error) {
	uID, err := uuid.Parse(id)
	if err != nil {
		return nil, xerr.NewErrMsg("شناسه راهنما نامعتبر است")
	}

	var g model.HelpGuide
	if err := l.svcCtx.DB.WithContext(l.ctx).Where("id = ?", uID).First(&g).Error; err != nil {
		return nil, xerr.NewErrMsg("راهنمای مورد نظر یافت نشد")
	}

	updates := map[string]interface{}{}

	if req.Title != "" {
		updates["title"] = strings.TrimSpace(req.Title)
	}
	if req.Slug != "" && req.Slug != g.Slug {
		updates["slug"] = strings.TrimSpace(req.Slug)
	}
	if req.Excerpt != "" {
		updates["excerpt"] = strings.TrimSpace(req.Excerpt)
	}
	if req.Content != "" {
		updates["content"] = req.Content
		if req.ReadingTimeMinutes <= 0 {
			updates["reading_time_minutes"] = model.CalculateReadingTime(req.Content)
		}
		autoTOC := extractTOCFromAdminContent(req.Content)
		if b, err := json.Marshal(autoTOC); err == nil {
			updates["table_of_contents_json"] = string(b)
		}
	}
	if req.CategorySlug != "" {
		updates["category_slug"] = strings.TrimSpace(req.CategorySlug)
	}
	if req.CategoryName != "" {
		updates["category_name"] = strings.TrimSpace(req.CategoryName)
	}
	if req.Icon != "" {
		updates["icon"] = strings.TrimSpace(req.Icon)
	}
	if req.Difficulty != "" {
		updates["difficulty"] = req.Difficulty
	}
	if req.ServiceTarget != "" {
		updates["service_target"] = req.ServiceTarget
	}
	if req.ReadingTimeMinutes > 0 {
		updates["reading_time_minutes"] = req.ReadingTimeMinutes
	}
	if req.SortOrder != 0 {
		updates["sort_order"] = req.SortOrder
	}
	if req.IsFeatured != nil {
		updates["is_featured"] = *req.IsFeatured
	}
	if req.IsPublished != nil {
		updates["is_published"] = *req.IsPublished
		if *req.IsPublished && g.PublishedAt == nil {
			now := time.Now()
			updates["published_at"] = &now
		}
	}
	if req.RelatedProductID != "" {
		if pID, err := uuid.Parse(req.RelatedProductID); err == nil {
			updates["related_product_id"] = pID
		}
	}
	if req.RelatedProductTitle != "" {
		updates["related_product_title"] = strings.TrimSpace(req.RelatedProductTitle)
	}
	if req.RelatedProductSlug != "" {
		updates["related_product_slug"] = strings.TrimSpace(req.RelatedProductSlug)
	}
	if req.RelatedProductPrice > 0 {
		updates["related_product_price"] = req.RelatedProductPrice
	}
	if req.MetaTitle != "" {
		updates["meta_title"] = strings.TrimSpace(req.MetaTitle)
	}
	if req.MetaDescription != "" {
		updates["meta_description"] = strings.TrimSpace(req.MetaDescription)
	}
	if req.MetaKeywords != "" {
		updates["meta_keywords"] = strings.TrimSpace(req.MetaKeywords)
	}
	if req.TableOfContentsJSON != "" {
		updates["table_of_contents_json"] = req.TableOfContentsJSON
	}

	updates["updated_at"] = time.Now()

	if err := l.svcCtx.DB.WithContext(l.ctx).Model(&g).Updates(updates).Error; err != nil {
		return nil, xerr.NewErrCode(xerr.DatabaseError)
	}

	return l.GetHelpGuideAdmin(id)
}

// DeleteHelpGuide removes a guide
func (l *ContentAdminLogic) DeleteHelpGuide(id string) error {
	uID, err := uuid.Parse(id)
	if err != nil {
		return xerr.NewErrMsg("شناسه راهنما نامعتبر است")
	}

	if err := l.svcCtx.DB.WithContext(l.ctx).Where("id = ?", uID).Delete(&model.HelpGuide{}).Error; err != nil {
		return xerr.NewErrCode(xerr.DatabaseError)
	}
	return nil
}

// ---------------- CATEGORY MANAGEMENT ----------------

func (l *ContentAdminLogic) ListCategories(catType string) ([]types.CategoryDTO, error) {
	var cats []model.ContentCategory
	query := l.svcCtx.DB.WithContext(l.ctx)
	if catType != "" && catType != "all" {
		query = query.Where("type = ?", catType)
	}
	if err := query.Order("type asc, sort_order asc, name asc").Find(&cats).Error; err != nil {
		return nil, xerr.NewErrCode(xerr.DatabaseError)
	}

	res := make([]types.CategoryDTO, len(cats))
	for i, c := range cats {
		var count int64 = 0
		if c.Type == "blog" {
			l.svcCtx.DB.WithContext(l.ctx).Model(&model.Article{}).Where("category_slug = ?", c.Slug).Count(&count)
		} else {
			l.svcCtx.DB.WithContext(l.ctx).Model(&model.HelpGuide{}).Where("category_slug = ?", c.Slug).Count(&count)
		}

		res[i] = types.CategoryDTO{
			ID:           c.ID.String(),
			Name:         c.Name,
			Slug:         c.Slug,
			Type:         c.Type,
			Icon:         c.Icon,
			Description:  c.Description,
			SortOrder:    c.SortOrder,
			IsActive:     c.IsActive,
			ArticleCount: count,
		}
	}
	return res, nil
}

func (l *ContentAdminLogic) CreateCategory(req *types.CreateCategoryRequest) (*types.CategoryDTO, error) {
	name := strings.TrimSpace(req.Name)
	slug := strings.TrimSpace(req.Slug)
	if slug == "" {
		slug = generateSlug(name)
	}

	isAct := true
	if req.IsActive != nil {
		isAct = *req.IsActive
	}

	icon := strings.TrimSpace(req.Icon)
	if icon == "" {
		icon = "BookOpen"
	}

	category := model.ContentCategory{
		Name:        name,
		Slug:        slug,
		Type:        req.Type,
		Icon:        icon,
		Description: strings.TrimSpace(req.Description),
		SortOrder:   req.SortOrder,
		IsActive:    isAct,
	}

	if err := l.svcCtx.DB.WithContext(l.ctx).Create(&category).Error; err != nil {
		return nil, xerr.NewErrMsg("دسته با این نام یا شناسه تکراری است")
	}

	return &types.CategoryDTO{
		ID:          category.ID.String(),
		Name:        category.Name,
		Slug:        category.Slug,
		Type:        category.Type,
		Icon:        category.Icon,
		Description: category.Description,
		SortOrder:   category.SortOrder,
		IsActive:    category.IsActive,
	}, nil
}

func (l *ContentAdminLogic) UpdateCategory(id string, req *types.UpdateCategoryRequest) error {
	uID, err := uuid.Parse(id)
	if err != nil {
		return xerr.NewErrMsg("شناسه دسته نامعتبر است")
	}

	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = strings.TrimSpace(req.Name)
	}
	if req.Slug != "" {
		updates["slug"] = strings.TrimSpace(req.Slug)
	}
	if req.Type != "" {
		updates["type"] = req.Type
	}
	if req.Icon != "" {
		updates["icon"] = strings.TrimSpace(req.Icon)
	}
	if req.Description != "" {
		updates["description"] = strings.TrimSpace(req.Description)
	}
	if req.SortOrder != 0 {
		updates["sort_order"] = req.SortOrder
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	updates["updated_at"] = time.Now()

	if err := l.svcCtx.DB.WithContext(l.ctx).Model(&model.ContentCategory{}).Where("id = ?", uID).Updates(updates).Error; err != nil {
		return xerr.NewErrCode(xerr.DatabaseError)
	}
	return nil
}

func (l *ContentAdminLogic) DeleteCategory(id string) error {
	uID, err := uuid.Parse(id)
	if err != nil {
		return xerr.NewErrMsg("شناسه دسته نامعتبر است")
	}

	if err := l.svcCtx.DB.WithContext(l.ctx).Where("id = ?", uID).Delete(&model.ContentCategory{}).Error; err != nil {
		return xerr.NewErrCode(xerr.DatabaseError)
	}
	return nil
}

// ---------------- COMMENTS MANAGEMENT ----------------

func (l *ContentAdminLogic) ListComments(articleID string, status string) ([]types.ArticleCommentDTO, error) {
	query := l.svcCtx.DB.WithContext(l.ctx).Model(&model.ArticleComment{})

	if articleID != "" && articleID != "all" {
		if aID, err := uuid.Parse(articleID); err == nil {
			query = query.Where("article_id = ?", aID)
		}
	}

	if status != "" && status != "all" {
		query = query.Where("status = ?", status)
	}

	var comments []model.ArticleComment
	if err := query.Order("created_at desc").Find(&comments).Error; err != nil {
		return nil, xerr.NewErrCode(xerr.DatabaseError)
	}

	res := make([]types.ArticleCommentDTO, len(comments))
	for i, c := range comments {
		res[i] = types.ArticleCommentDTO{
			ID:         c.ID.String(),
			ArticleID:  c.ArticleID.String(),
			AuthorName: c.AuthorName,
			Content:    c.Content,
			Status:     c.Status,
			AdminReply: c.AdminReply,
			CreatedAt:  c.CreatedAt,
		}
	}
	return res, nil
}

func (l *ContentAdminLogic) UpdateCommentStatus(id string, status string, reply string) error {
	uID, err := uuid.Parse(id)
	if err != nil {
		return xerr.NewErrMsg("شناسه نظر نامعتبر است")
	}

	updates := map[string]interface{}{}
	if status != "" {
		updates["status"] = status
	}
	if reply != "" {
		updates["admin_reply"] = strings.TrimSpace(reply)
	}
	updates["updated_at"] = time.Now()

	return l.svcCtx.DB.WithContext(l.ctx).Model(&model.ArticleComment{}).Where("id = ?", uID).Updates(updates).Error
}

func (l *ContentAdminLogic) DeleteComment(id string) error {
	uID, err := uuid.Parse(id)
	if err != nil {
		return xerr.NewErrMsg("شناسه نظر نامعتبر است")
	}
	return l.svcCtx.DB.WithContext(l.ctx).Where("id = ?", uID).Delete(&model.ArticleComment{}).Error
}

// Helpers
func generateSlug(title string) string {
	s := strings.ToLower(title)
	s = strings.ReplaceAll(s, " ", "-")
	reg := regexp.MustCompile(`[^a-z0-9\x{0600}-\x{06FF}\-]`)
	s = reg.ReplaceAllString(s, "")
	s = strings.Trim(s, "-")
	if s == "" {
		s = fmt.Sprintf("post-%d", time.Now().Unix())
	}
	return s
}

func extractTOCFromAdminContent(content string) []types.TOCItemDTO {
	var toc []types.TOCItemDTO
	lines := strings.Split(content, "\n")

	reH2 := regexp.MustCompile(`^##\s+(.+)$`)
	reH3 := regexp.MustCompile(`^###\s+(.+)$`)
	reHTMLH2 := regexp.MustCompile(`(?i)<h2[^>]*>(.*?)</h2>`)
	reHTMLH3 := regexp.MustCompile(`(?i)<h3[^>]*>(.*?)</h3>`)

	idx := 1
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if m := reH2.FindStringSubmatch(trimmed); len(m) > 1 {
			toc = append(toc, types.TOCItemDTO{ID: fmt.Sprintf("heading-%d", idx), Text: cleanAdminHeading(m[1]), Level: 2})
			idx++
		} else if m := reH3.FindStringSubmatch(trimmed); len(m) > 1 {
			toc = append(toc, types.TOCItemDTO{ID: fmt.Sprintf("heading-%d", idx), Text: cleanAdminHeading(m[1]), Level: 3})
			idx++
		} else if m := reHTMLH2.FindStringSubmatch(trimmed); len(m) > 1 {
			toc = append(toc, types.TOCItemDTO{ID: fmt.Sprintf("heading-%d", idx), Text: cleanAdminHeading(m[1]), Level: 2})
			idx++
		} else if m := reHTMLH3.FindStringSubmatch(trimmed); len(m) > 1 {
			toc = append(toc, types.TOCItemDTO{ID: fmt.Sprintf("heading-%d", idx), Text: cleanAdminHeading(m[1]), Level: 3})
			idx++
		}
	}
	return toc
}

func cleanAdminHeading(raw string) string {
	r := regexp.MustCompile(`<[^>]*>`)
	cleaned := r.ReplaceAllString(raw, "")
	cleaned = strings.Trim(cleaned, "*_`# ")
	return strings.TrimSpace(cleaned)
}

// ---------------- CONTENT TEMPLATES MANAGEMENT ----------------

func (l *ContentAdminLogic) ListTemplatesAdmin(contentType string) ([]types.ContentTemplateDTO, error) {
	var list []model.ContentTemplate
	query := l.svcCtx.DB.WithContext(l.ctx)
	if contentType != "" && contentType != "all" {
		query = query.Where("type = ? OR type = 'all'", contentType)
	}

	if err := query.Order("sort_order asc, title asc").Find(&list).Error; err != nil {
		return nil, xerr.NewErrCode(xerr.DatabaseError)
	}

	res := make([]types.ContentTemplateDTO, len(list))
	for i, t := range list {
		res[i] = types.ContentTemplateDTO{
			ID:              t.ID.String(),
			Title:           t.Title,
			Slug:            t.Slug,
			Type:            t.Type,
			Icon:            t.Icon,
			Description:     t.Description,
			TemplateContent: t.TemplateContent,
			SortOrder:       t.SortOrder,
			IsActive:        t.IsActive,
		}
	}
	return res, nil
}

func (l *ContentAdminLogic) CreateTemplate(req *types.CreateTemplateRequest) (*types.ContentTemplateDTO, error) {
	title := strings.TrimSpace(req.Title)
	slug := strings.TrimSpace(req.Slug)
	if slug == "" {
		slug = generateSlug(title)
	}

	var count int64
	l.svcCtx.DB.WithContext(l.ctx).Model(&model.ContentTemplate{}).Where("slug = ?", slug).Count(&count)
	if count > 0 {
		slug = fmt.Sprintf("%s-%d", slug, time.Now().Unix()%10000)
	}

	isAct := true
	if req.IsActive != nil {
		isAct = *req.IsActive
	}

	tmplType := req.Type
	if tmplType == "" {
		tmplType = "all"
	}

	icon := strings.TrimSpace(req.Icon)
	if icon == "" {
		icon = "FileText"
	}

	tmpl := model.ContentTemplate{
		Title:           title,
		Slug:            slug,
		Type:            tmplType,
		Icon:            icon,
		Description:     strings.TrimSpace(req.Description),
		TemplateContent: req.TemplateContent,
		SortOrder:       req.SortOrder,
		IsActive:        isAct,
	}

	if err := l.svcCtx.DB.WithContext(l.ctx).Create(&tmpl).Error; err != nil {
		return nil, xerr.NewErrCode(xerr.DatabaseError)
	}

	return &types.ContentTemplateDTO{
		ID:              tmpl.ID.String(),
		Title:           tmpl.Title,
		Slug:            tmpl.Slug,
		Type:            tmpl.Type,
		Icon:            tmpl.Icon,
		Description:     tmpl.Description,
		TemplateContent: tmpl.TemplateContent,
		SortOrder:       tmpl.SortOrder,
		IsActive:        tmpl.IsActive,
	}, nil
}

func (l *ContentAdminLogic) UpdateTemplate(id string, req *types.UpdateTemplateRequest) error {
	uID, err := uuid.Parse(id)
	if err != nil {
		return xerr.NewErrMsg("شناسه قالب نامعتبر است")
	}

	updates := map[string]interface{}{}
	if req.Title != "" {
		updates["title"] = strings.TrimSpace(req.Title)
	}
	if req.Slug != "" {
		updates["slug"] = strings.TrimSpace(req.Slug)
	}
	if req.Type != "" {
		updates["type"] = req.Type
	}
	if req.Icon != "" {
		updates["icon"] = strings.TrimSpace(req.Icon)
	}
	if req.Description != "" {
		updates["description"] = strings.TrimSpace(req.Description)
	}
	if req.TemplateContent != "" {
		updates["template_content"] = req.TemplateContent
	}
	if req.SortOrder != 0 {
		updates["sort_order"] = req.SortOrder
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	updates["updated_at"] = time.Now()

	return l.svcCtx.DB.WithContext(l.ctx).Model(&model.ContentTemplate{}).Where("id = ?", uID).Updates(updates).Error
}

func (l *ContentAdminLogic) DeleteTemplate(id string) error {
	uID, err := uuid.Parse(id)
	if err != nil {
		return xerr.NewErrMsg("شناسه قالب نامعتبر است")
	}
	return l.svcCtx.DB.WithContext(l.ctx).Where("id = ?", uID).Delete(&model.ContentTemplate{}).Error
}

// ---------------- SERVICE TARGETS MANAGEMENT ----------------

func (l *ContentAdminLogic) ListServiceTargetsAdmin() ([]types.ServiceTargetDTO, error) {
	var list []model.ServiceTarget
	if err := l.svcCtx.DB.WithContext(l.ctx).Order("sort_order asc, key asc").Find(&list).Error; err != nil {
		return nil, xerr.NewErrCode(xerr.DatabaseError)
	}

	res := make([]types.ServiceTargetDTO, len(list))
	for i, s := range list {
		res[i] = types.ServiceTargetDTO{
			ID:        s.ID.String(),
			Key:       s.Key,
			Label:     s.Label,
			Icon:      s.Icon,
			SortOrder: s.SortOrder,
			IsActive:  s.IsActive,
		}
	}
	return res, nil
}

func (l *ContentAdminLogic) CreateServiceTarget(req *types.CreateServiceTargetRequest) (*types.ServiceTargetDTO, error) {
	key := strings.ToLower(strings.TrimSpace(req.Key))
	label := strings.TrimSpace(req.Label)
	icon := strings.TrimSpace(req.Icon)
	if icon == "" {
		icon = "Sparkles"
	}

	isAct := true
	if req.IsActive != nil {
		isAct = *req.IsActive
	}

	target := model.ServiceTarget{
		Key:       key,
		Label:     label,
		Icon:      icon,
		SortOrder: req.SortOrder,
		IsActive:  isAct,
	}

	if err := l.svcCtx.DB.WithContext(l.ctx).Create(&target).Error; err != nil {
		return nil, xerr.NewErrMsg("سرویس با این کلید تکراری است")
	}

	return &types.ServiceTargetDTO{
		ID:        target.ID.String(),
		Key:       target.Key,
		Label:     target.Label,
		Icon:      target.Icon,
		SortOrder: target.SortOrder,
		IsActive:  target.IsActive,
	}, nil
}

func (l *ContentAdminLogic) UpdateServiceTarget(id string, req *types.CreateServiceTargetRequest) error {
	uID, err := uuid.Parse(id)
	if err != nil {
		return xerr.NewErrMsg("شناسه سرویس نامعتبر است")
	}

	updates := map[string]interface{}{}
	if req.Key != "" {
		updates["key"] = strings.ToLower(strings.TrimSpace(req.Key))
	}
	if req.Label != "" {
		updates["label"] = strings.TrimSpace(req.Label)
	}
	if req.Icon != "" {
		updates["icon"] = strings.TrimSpace(req.Icon)
	}
	if req.SortOrder != 0 {
		updates["sort_order"] = req.SortOrder
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	updates["updated_at"] = time.Now()

	return l.svcCtx.DB.WithContext(l.ctx).Model(&model.ServiceTarget{}).Where("id = ?", uID).Updates(updates).Error
}

func (l *ContentAdminLogic) DeleteServiceTarget(id string) error {
	uID, err := uuid.Parse(id)
	if err != nil {
		return xerr.NewErrMsg("شناسه سرویس نامعتبر است")
	}
	return l.svcCtx.DB.WithContext(l.ctx).Where("id = ?", uID).Delete(&model.ServiceTarget{}).Error
}
