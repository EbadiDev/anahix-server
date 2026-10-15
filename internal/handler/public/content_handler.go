package public

import (
	"strconv"

	"github.com/EbadiDev/anahix-server/internal/logic/public"
	"github.com/EbadiDev/anahix-server/internal/svc"
	"github.com/EbadiDev/anahix-server/internal/types"
	"github.com/EbadiDev/anahix-server/pkg/result"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ContentHandler struct {
	svcCtx *svc.ServiceContext
}

func NewContentHandler(svcCtx *svc.ServiceContext) *ContentHandler {
	return &ContentHandler{svcCtx: svcCtx}
}

// ListCategories returns categories for blog or help
func (h *ContentHandler) ListCategories(c *gin.Context) {
	catType := c.Query("type")
	l := public.NewContentLogic(c.Request.Context(), h.svcCtx)
	resp, err := l.ListCategories(catType)
	result.HttpResult(c, resp, err)
}

// ListArticles returns paginated blog articles
func (h *ContentHandler) ListArticles(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "12"))
	category := c.Query("category")
	tag := c.Query("tag")
	search := c.Query("search")

	l := public.NewContentLogic(c.Request.Context(), h.svcCtx)
	resp, err := l.ListArticles(page, pageSize, category, tag, search)
	result.HttpResult(c, resp, err)
}

// GetArticleDetail returns article by slug
func (h *ContentHandler) GetArticleDetail(c *gin.Context) {
	slug := c.Param("slug")
	l := public.NewContentLogic(c.Request.Context(), h.svcCtx)
	resp, err := l.GetArticleBySlug(slug)
	result.HttpResult(c, resp, err)
}

// SubmitComment creates a comment for an article
func (h *ContentHandler) SubmitComment(c *gin.Context) {
	slug := c.Param("slug")
	var req types.SubmitArticleCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		result.ParamErrorResult(c, err)
		return
	}

	var customerUUID *uuid.UUID
	if cID := c.GetString("customer_id"); cID != "" {
		if parsed, err := uuid.Parse(cID); err == nil {
			customerUUID = &parsed
		}
	}

	l := public.NewContentLogic(c.Request.Context(), h.svcCtx)
	err := l.SubmitArticleComment(slug, &req, customerUUID)
	result.HttpResult(c, gin.H{"status": "ok"}, err)
}

// ListHelpGuides returns paginated knowledge base guides
func (h *ContentHandler) ListHelpGuides(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "15"))
	category := c.Query("category")
	serviceTarget := c.Query("service_target")
	search := c.Query("search")

	l := public.NewContentLogic(c.Request.Context(), h.svcCtx)
	resp, err := l.ListHelpGuides(page, pageSize, category, serviceTarget, search)
	result.HttpResult(c, resp, err)
}

// GetHelpGuideDetail returns help guide by slug
func (h *ContentHandler) GetHelpGuideDetail(c *gin.Context) {
	slug := c.Param("slug")
	l := public.NewContentLogic(c.Request.Context(), h.svcCtx)
	resp, err := l.GetHelpGuideBySlug(slug)
	result.HttpResult(c, resp, err)
}

// VoteHelpGuide handles helpful / unhelpful voting
func (h *ContentHandler) VoteHelpGuide(c *gin.Context) {
	slug := c.Param("slug")
	var req types.GuideFeedbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		result.ParamErrorResult(c, err)
		return
	}

	voterKey := c.ClientIP()
	if cPhone := c.GetString("customer_phone"); cPhone != "" {
		voterKey = cPhone
	}

	l := public.NewContentLogic(c.Request.Context(), h.svcCtx)
	err := l.VoteHelpGuide(slug, req.IsHelpful, voterKey)
	result.HttpResult(c, gin.H{"status": "recorded"}, err)
}

// ListTemplates returns pre-built drafting templates from DB
func (h *ContentHandler) ListTemplates(c *gin.Context) {
	tmplType := c.Query("type")
	l := public.NewContentLogic(c.Request.Context(), h.svcCtx)
	resp, err := l.ListContentTemplates(tmplType)
	result.HttpResult(c, resp, err)
}

// ListServices returns dynamic AI service targets from DB
func (h *ContentHandler) ListServices(c *gin.Context) {
	l := public.NewContentLogic(c.Request.Context(), h.svcCtx)
	resp, err := l.ListServiceTargets()
	result.HttpResult(c, resp, err)
}
