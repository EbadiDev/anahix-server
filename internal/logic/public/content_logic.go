package public

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"strings"

	"github.com/EbadiDev/anahix-server/internal/model"
	"github.com/EbadiDev/anahix-server/internal/svc"
	"github.com/EbadiDev/anahix-server/internal/types"
	"github.com/EbadiDev/anahix-server/pkg/xerr"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ContentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewContentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ContentLogic {
	return &ContentLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// ListCategories returns active taxonomy categories
func (l *ContentLogic) ListCategories(catType string) ([]types.CategoryDTO, error) {
	var cats []model.ContentCategory
	query := l.svcCtx.DB.WithContext(l.ctx).Where("is_active = ?", true)
	if catType != "" && catType != "all" {
		query = query.Where("type = ?", catType)
	}
	if err := query.Order("sort_order asc, name asc").Find(&cats).Error; err != nil {
		return nil, xerr.NewErrCode(xerr.DatabaseError)
	}

	allDTOs := make([]types.CategoryDTO, len(cats))
	dtoMap := make(map[string]*types.CategoryDTO)

	for i, c := range cats {
		var count int64 = 0
		if c.Type == "blog" {
			l.svcCtx.DB.WithContext(l.ctx).Model(&model.Article{}).Where("category_slug = ? AND is_published = ?", c.Slug, true).Count(&count)
		} else if c.Type == "product" {
			l.svcCtx.DB.WithContext(l.ctx).Model(&model.Product{}).Where("(category = ? OR group_tag = ?) AND is_active = ?", c.Slug, c.Slug, true).Count(&count)
		} else {
			l.svcCtx.DB.WithContext(l.ctx).Model(&model.HelpGuide{}).Where("category_slug = ? AND is_published = ?", c.Slug, true).Count(&count)
		}

		var pid *string
		if c.ParentID != nil {
			s := c.ParentID.String()
			pid = &s
		}

		allDTOs[i] = types.CategoryDTO{
			ID:           c.ID.String(),
			ParentID:     pid,
			Name:         c.Name,
			Slug:         c.Slug,
			Type:         c.Type,
			Icon:         c.Icon,
			Description:  c.Description,
			SortOrder:    c.SortOrder,
			IsActive:     c.IsActive,
			ArticleCount: count,
			Children:     []types.CategoryDTO{},
		}
		dtoMap[allDTOs[i].ID] = &allDTOs[i]
	}

	// Build nested hierarchy
	var rootCats []types.CategoryDTO
	for _, dto := range allDTOs {
		if dto.ParentID != nil && *dto.ParentID != "" {
			if parent, exists := dtoMap[*dto.ParentID]; exists {
				parent.Children = append(parent.Children, dto)
				continue
			}
		}
		rootCats = append(rootCats, dto)
	}

	// If hierarchy built top-level items, return rootCats with nested children; otherwise return all
	if len(rootCats) > 0 {
		// Reflect the mutated children
		finalRes := make([]types.CategoryDTO, len(rootCats))
		for i, r := range rootCats {
			if ptr, ok := dtoMap[r.ID]; ok {
				finalRes[i] = *ptr
			} else {
				finalRes[i] = r
			}
		}
		return finalRes, nil
	}

	return allDTOs, nil
}

// ListArticles returns paginated published blog articles
func (l *ContentLogic) ListArticles(page, pageSize int, categorySlug, tag, search string) (*types.ContentListResponse[types.ArticleListItemDTO], error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 12
	}

	query := l.svcCtx.DB.WithContext(l.ctx).Model(&model.Article{}).Where("is_published = ?", true)

	if categorySlug != "" && categorySlug != "all" {
		query = query.Where("category_slug = ?", categorySlug)
	}

	if tag != "" {
		tagPattern := "%" + tag + "%"
		query = query.Where("tags LIKE ?", tagPattern)
	}

	if strings.TrimSpace(search) != "" {
		s := "%" + strings.TrimSpace(search) + "%"
		query = query.Where("title LIKE ? OR excerpt LIKE ? OR content LIKE ? OR tags LIKE ?", s, s, s, s)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, xerr.NewErrCode(xerr.DatabaseError)
	}

	offset := (page - 1) * pageSize
	var articles []model.Article
	if err := query.Order("is_featured desc, published_at desc, created_at desc").
		Offset(offset).Limit(pageSize).Find(&articles).Error; err != nil {
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

// GetArticleBySlug returns single article with persistent unique IP view tracking in database
func (l *ContentLogic) GetArticleBySlug(slug string, clientIP string) (*types.ArticleDetailDTO, error) {
	var a model.Article
	if err := l.svcCtx.DB.WithContext(l.ctx).
		Where("slug = ? AND is_published = ?", slug, true).
		First(&a).Error; err != nil {
		return nil, xerr.NewErrMsg("مقاله مورد نظر یافت نشد")
	}

	if clientIP == "" {
		clientIP = "127.0.0.1"
	}

	// Unique IP view tracking in database
	var existingView model.ArticleView
	viewErr := l.svcCtx.DB.WithContext(l.ctx).
		Where("article_id = ? AND ip = ?", a.ID, clientIP).
		First(&existingView).Error

	if viewErr != nil {
		newView := model.ArticleView{
			ArticleID: a.ID,
			IP:        clientIP,
		}
		_ = l.svcCtx.DB.WithContext(l.ctx).Create(&newView).Error
		_ = l.svcCtx.DB.WithContext(l.ctx).Model(&model.Article{}).Where("id = ?", a.ID).
			UpdateColumn("view_count", gorm.Expr("view_count + ?", 1)).Error
		a.ViewCount += 1
	}

	// Parse Table of Contents
	var toc []types.TOCItemDTO
	if a.TableOfContentsJSON != "" {
		_ = json.Unmarshal([]byte(a.TableOfContentsJSON), &toc)
	}
	if len(toc) == 0 {
		toc = extractTOCFromContent(a.Content)
	}

	// Load Approved Comments
	var comments []model.ArticleComment
	l.svcCtx.DB.WithContext(l.ctx).
		Where("article_id = ? AND status = ?", a.ID, "approved").
		Order("created_at desc").
		Find(&comments)

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

	// Load Related Articles (same category, excluding current)
	var related []model.Article
	l.svcCtx.DB.WithContext(l.ctx).
		Where("category_slug = ? AND id != ? AND is_published = ?", a.CategorySlug, a.ID, true).
		Order("published_at desc").
		Limit(3).
		Find(&related)

	if len(related) < 3 {
		var extra []model.Article
		excludeIDs := []uuid.UUID{a.ID}
		for _, r := range related {
			excludeIDs = append(excludeIDs, r.ID)
		}
		l.svcCtx.DB.WithContext(l.ctx).
			Where("id NOT IN ? AND is_published = ?", excludeIDs, true).
			Order("is_featured desc, published_at desc").
			Limit(3 - len(related)).
			Find(&extra)
		related = append(related, extra...)
	}

	relatedDTOs := make([]types.ArticleListItemDTO, len(related))
	for i, r := range related {
		relatedDTOs[i] = types.ArticleListItemDTO{
			ID:                  r.ID.String(),
			Title:               r.Title,
			Slug:                r.Slug,
			Excerpt:             r.Excerpt,
			CoverImage:          r.CoverImage,
			CategorySlug:        r.CategorySlug,
			CategoryName:        r.CategoryName,
			Tags:                r.GetTagsSlice(),
			AuthorName:          r.AuthorName,
			AuthorRole:          r.AuthorRole,
			AuthorAvatar:        r.AuthorAvatar,
			ReadingTimeMinutes:  r.ReadingTimeMinutes,
			ViewCount:           r.ViewCount,
			LikeCount:           r.LikeCount,
			IsFeatured:          r.IsFeatured,
			IsPublished:         r.IsPublished,
			PublishedAt:         r.PublishedAt,
			RelatedProductTitle: r.RelatedProductTitle,
			RelatedProductSlug:  r.RelatedProductSlug,
			CreatedAt:           r.CreatedAt,
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
		RelatedArticles:     relatedDTOs,
		CreatedAt:           a.CreatedAt,
		UpdatedAt:           a.UpdatedAt,
	}, nil
}

// SubmitArticleComment creates a reader comment awaiting admin approval
func (l *ContentLogic) SubmitArticleComment(slug string, req *types.SubmitArticleCommentRequest, customerID *uuid.UUID) error {
	var a model.Article
	if err := l.svcCtx.DB.WithContext(l.ctx).Where("slug = ? AND is_published = ?", slug, true).First(&a).Error; err != nil {
		return xerr.NewErrMsg("مقاله مورد نظر یافت نشد")
	}

	content := strings.TrimSpace(req.Content)
	authorName := strings.TrimSpace(req.AuthorName)
	if content == "" || authorName == "" {
		return xerr.NewErrMsg("نام و متن دیدگاه الزامی است")
	}

	comment := model.ArticleComment{
		ArticleID:  a.ID,
		CustomerID: customerID,
		AuthorName: authorName,
		Email:      strings.TrimSpace(req.Email),
		Content:    content,
		Status:     "pending", // Must be approved by admin before public display
	}

	if err := l.svcCtx.DB.WithContext(l.ctx).Create(&comment).Error; err != nil {
		return xerr.NewErrCode(xerr.DatabaseError)
	}
	return nil
}

// ListHelpGuides returns paginated knowledge base guides
func (l *ContentLogic) ListHelpGuides(page, pageSize int, categorySlug, serviceTarget, search string) (*types.ContentListResponse[types.HelpGuideListItemDTO], error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 15
	}

	query := l.svcCtx.DB.WithContext(l.ctx).Model(&model.HelpGuide{}).Where("is_published = ?", true)

	if categorySlug != "" && categorySlug != "all" {
		query = query.Where("category_slug = ?", categorySlug)
	}

	if serviceTarget != "" && serviceTarget != "all" {
		query = query.Where("service_target = ?", serviceTarget)
	}

	if strings.TrimSpace(search) != "" {
		s := "%" + strings.TrimSpace(search) + "%"
		query = query.Where("title LIKE ? OR excerpt LIKE ? OR content LIKE ? OR difficulty LIKE ?", s, s, s, s)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, xerr.NewErrCode(xerr.DatabaseError)
	}

	offset := (page - 1) * pageSize
	var guides []model.HelpGuide
	if err := query.Order("sort_order asc, is_featured desc, helpful_yes desc, created_at desc").
		Offset(offset).Limit(pageSize).Find(&guides).Error; err != nil {
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
			AuthorName:          g.AuthorName,
			AuthorRole:          g.AuthorRole,
			AuthorAvatar:        g.AuthorAvatar,
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

// GetHelpGuideBySlug returns single knowledge base guide with persistent unique IP view tracking in database
func (l *ContentLogic) GetHelpGuideBySlug(slug string, clientIP string) (*types.HelpGuideDetailDTO, error) {
	var g model.HelpGuide
	if err := l.svcCtx.DB.WithContext(l.ctx).
		Where("slug = ? AND is_published = ?", slug, true).
		First(&g).Error; err != nil {
		return nil, xerr.NewErrMsg("راهنمای مورد نظر یافت نشد")
	}

	if clientIP == "" {
		clientIP = "127.0.0.1"
	}

	// Unique IP view tracking in database
	var existingView model.HelpGuideView
	viewErr := l.svcCtx.DB.WithContext(l.ctx).
		Where("guide_id = ? AND ip = ?", g.ID, clientIP).
		First(&existingView).Error

	if viewErr != nil {
		newView := model.HelpGuideView{
			GuideID: g.ID,
			IP:      clientIP,
		}
		_ = l.svcCtx.DB.WithContext(l.ctx).Create(&newView).Error
		_ = l.svcCtx.DB.WithContext(l.ctx).Model(&model.HelpGuide{}).Where("id = ?", g.ID).
			UpdateColumn("view_count", gorm.Expr("view_count + ?", 1)).Error
		g.ViewCount += 1
	}

	// Parse Table of Contents
	var toc []types.TOCItemDTO
	if g.TableOfContentsJSON != "" {
		_ = json.Unmarshal([]byte(g.TableOfContentsJSON), &toc)
	}
	if len(toc) == 0 {
		toc = extractTOCFromContent(g.Content)
	}

	// Related guides
	var related []model.HelpGuide
	l.svcCtx.DB.WithContext(l.ctx).
		Where("(category_slug = ? OR service_target = ?) AND id != ? AND is_published = ?", g.CategorySlug, g.ServiceTarget, g.ID, true).
		Order("is_featured desc, helpful_yes desc").
		Limit(4).
		Find(&related)

	relatedDTOs := make([]types.HelpGuideListItemDTO, len(related))
	for i, r := range related {
		relatedDTOs[i] = types.HelpGuideListItemDTO{
			ID:                  r.ID.String(),
			Title:               r.Title,
			Slug:                r.Slug,
			Excerpt:             r.Excerpt,
			CategorySlug:        r.CategorySlug,
			CategoryName:        r.CategoryName,
			Icon:                r.Icon,
			Difficulty:          r.Difficulty,
			ServiceTarget:       r.ServiceTarget,
			AuthorName:          r.AuthorName,
			AuthorRole:          r.AuthorRole,
			AuthorAvatar:        r.AuthorAvatar,
			ReadingTimeMinutes:  r.ReadingTimeMinutes,
			ViewCount:           r.ViewCount,
			HelpfulYes:          r.HelpfulYes,
			HelpfulNo:           r.HelpfulNo,
			IsFeatured:          r.IsFeatured,
			IsPublished:         r.IsPublished,
			PublishedAt:         r.PublishedAt,
			RelatedProductTitle: r.RelatedProductTitle,
			RelatedProductSlug:  r.RelatedProductSlug,
			CreatedAt:           r.CreatedAt,
		}
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
		AuthorName:          g.AuthorName,
		AuthorRole:          g.AuthorRole,
		AuthorAvatar:        g.AuthorAvatar,
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
		RelatedGuides:       relatedDTOs,
		CreatedAt:           g.CreatedAt,
		UpdatedAt:           g.UpdatedAt,
	}, nil
}

// VoteHelpGuide records or toggles helpful / unhelpful feedback for the guide
func (l *ContentLogic) VoteHelpGuide(slug string, isHelpful bool, voterKey string) (int64, int64, string, error) {
	var g model.HelpGuide
	if err := l.svcCtx.DB.WithContext(l.ctx).Where("slug = ?", slug).First(&g).Error; err != nil {
		return 0, 0, "", xerr.NewErrMsg("راهنمای مورد نظر یافت نشد")
	}

	if voterKey == "" {
		voterKey = "anon"
	}

	// Check if already voted
	var existing model.GuideFeedback
	err := l.svcCtx.DB.WithContext(l.ctx).
		Where("guide_id = ? AND voter_key = ?", g.ID, voterKey).
		First(&existing).Error

	if err == nil {
		// Existing vote found
		if existing.IsHelpful == isHelpful {
			// User clicked same button again -> Toggle OFF (Cancel vote)
			if err := l.svcCtx.DB.WithContext(l.ctx).Delete(&existing).Error; err != nil {
				return g.HelpfulYes, g.HelpfulNo, "", xerr.NewErrCode(xerr.DatabaseError)
			}
			if isHelpful {
				if g.HelpfulYes > 0 {
					g.HelpfulYes--
				}
				l.svcCtx.DB.Model(&model.HelpGuide{}).Where("id = ?", g.ID).Update("helpful_yes", g.HelpfulYes)
			} else {
				if g.HelpfulNo > 0 {
					g.HelpfulNo--
				}
				l.svcCtx.DB.Model(&model.HelpGuide{}).Where("id = ?", g.ID).Update("helpful_no", g.HelpfulNo)
			}
			return g.HelpfulYes, g.HelpfulNo, "cancelled", nil
		}

		// User switched vote (from helpful to unhelpful or vice versa)
		existing.IsHelpful = isHelpful
		if err := l.svcCtx.DB.WithContext(l.ctx).Save(&existing).Error; err != nil {
			return g.HelpfulYes, g.HelpfulNo, "", xerr.NewErrCode(xerr.DatabaseError)
		}
		if isHelpful {
			g.HelpfulYes++
			if g.HelpfulNo > 0 {
				g.HelpfulNo--
			}
		} else {
			g.HelpfulNo++
			if g.HelpfulYes > 0 {
				g.HelpfulYes--
			}
		}
		l.svcCtx.DB.Model(&model.HelpGuide{}).Where("id = ?", g.ID).Updates(map[string]interface{}{
			"helpful_yes": g.HelpfulYes,
			"helpful_no":  g.HelpfulNo,
		})
		return g.HelpfulYes, g.HelpfulNo, "switched", nil
	}

	// Record vote
	feedback := model.GuideFeedback{
		GuideID:   g.ID,
		VoterKey:  voterKey,
		IsHelpful: isHelpful,
	}
	if err := l.svcCtx.DB.WithContext(l.ctx).Create(&feedback).Error; err != nil {
		return g.HelpfulYes, g.HelpfulNo, "", xerr.NewErrCode(xerr.DatabaseError)
	}

	// Increment counter
	if isHelpful {
		g.HelpfulYes++
		l.svcCtx.DB.Model(&model.HelpGuide{}).Where("id = ?", g.ID).Update("helpful_yes", g.HelpfulYes)
	} else {
		g.HelpfulNo++
		l.svcCtx.DB.Model(&model.HelpGuide{}).Where("id = ?", g.ID).Update("helpful_no", g.HelpfulNo)
	}

	return g.HelpfulYes, g.HelpfulNo, "voted", nil
}

// extractTOCFromContent scans Markdown or HTML headings (#, ##, <h2>, <h3>) to build Table of Contents
func extractTOCFromContent(content string) []types.TOCItemDTO {
	var toc []types.TOCItemDTO
	lines := strings.Split(content, "\n")

	// Regex for markdown headings
	reH2 := regexp.MustCompile(`^##\s+(.+)$`)
	reH3 := regexp.MustCompile(`^###\s+(.+)$`)

	// Regex for HTML headings
	reHTMLH2 := regexp.MustCompile(`(?i)<h2[^>]*>(.*?)</h2>`)
	reHTMLH3 := regexp.MustCompile(`(?i)<h3[^>]*>(.*?)</h3>`)

	idx := 1
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if m := reH2.FindStringSubmatch(trimmed); len(m) > 1 {
			title := cleanHeadingText(m[1])
			id := fmt.Sprintf("heading-%d", idx)
			toc = append(toc, types.TOCItemDTO{ID: id, Text: title, Level: 2})
			idx++
		} else if m := reH3.FindStringSubmatch(trimmed); len(m) > 1 {
			title := cleanHeadingText(m[1])
			id := fmt.Sprintf("heading-%d", idx)
			toc = append(toc, types.TOCItemDTO{ID: id, Text: title, Level: 3})
			idx++
		} else if m := reHTMLH2.FindStringSubmatch(trimmed); len(m) > 1 {
			title := cleanHeadingText(m[1])
			id := fmt.Sprintf("heading-%d", idx)
			toc = append(toc, types.TOCItemDTO{ID: id, Text: title, Level: 2})
			idx++
		} else if m := reHTMLH3.FindStringSubmatch(trimmed); len(m) > 1 {
			title := cleanHeadingText(m[1])
			id := fmt.Sprintf("heading-%d", idx)
			toc = append(toc, types.TOCItemDTO{ID: id, Text: title, Level: 3})
			idx++
		}
	}
	return toc
}

func cleanHeadingText(raw string) string {
	// Strip HTML tags and markdown formatting
	r := regexp.MustCompile(`<[^>]*>`)
	cleaned := r.ReplaceAllString(raw, "")
	cleaned = strings.Trim(cleaned, "*_`# ")
	return strings.TrimSpace(cleaned)
}

// ListContentTemplates returns pre-built drafting templates from DB
func (l *ContentLogic) ListContentTemplates(contentType string) ([]types.ContentTemplateDTO, error) {
	var list []model.ContentTemplate
	query := l.svcCtx.DB.WithContext(l.ctx).Where("is_active = ?", true)
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

// ListServiceTargets returns dynamic AI service targets from DB
func (l *ContentLogic) ListServiceTargets() ([]types.ServiceTargetDTO, error) {
	var list []model.ServiceTarget
	if err := l.svcCtx.DB.WithContext(l.ctx).Where("is_active = ?", true).
		Order("sort_order asc, key asc").Find(&list).Error; err != nil {
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

// LikeArticle toggles like_count for an article and persists in DB per voter
func (l *ContentLogic) LikeArticle(slug string, voterKey string) (int64, bool, error) {
	var article model.Article
	err := l.svcCtx.DB.WithContext(l.ctx).
		Where("slug = ? AND is_published = ?", slug, true).
		First(&article).Error
	if err != nil {
		return 0, false, xerr.NewErrMsg("مقاله مورد نظر یافت نشد")
	}

	if voterKey == "" {
		voterKey = "anon"
	}

	var existing model.ArticleLike
	likeErr := l.svcCtx.DB.WithContext(l.ctx).
		Where("article_id = ? AND voter_key = ?", article.ID, voterKey).
		First(&existing).Error

	if likeErr == nil {
		// Already liked -> Toggle OFF (Unlike)
		if err := l.svcCtx.DB.WithContext(l.ctx).Delete(&existing).Error; err != nil {
			return article.LikeCount, true, xerr.NewErrCode(xerr.DatabaseError)
		}
		if article.LikeCount > 0 {
			article.LikeCount--
		}
		l.svcCtx.DB.WithContext(l.ctx).Model(&article).Update("like_count", article.LikeCount)
		return article.LikeCount, false, nil
	}

	// Not liked yet -> Toggle ON (Like)
	newLike := model.ArticleLike{
		ArticleID: article.ID,
		VoterKey:  voterKey,
	}
	if err := l.svcCtx.DB.WithContext(l.ctx).Create(&newLike).Error; err != nil {
		return article.LikeCount, false, xerr.NewErrCode(xerr.DatabaseError)
	}

	article.LikeCount++
	l.svcCtx.DB.WithContext(l.ctx).Model(&article).Update("like_count", article.LikeCount)
	return article.LikeCount, true, nil
}
