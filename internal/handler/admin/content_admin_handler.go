package admin

import (
	"strconv"

	"github.com/EbadiDev/anahix-server/internal/logic/admin"
	"github.com/EbadiDev/anahix-server/internal/svc"
	"github.com/EbadiDev/anahix-server/internal/types"
	"github.com/EbadiDev/anahix-server/pkg/result"
	"github.com/gin-gonic/gin"
)

type ContentAdminHandler struct {
	svcCtx *svc.ServiceContext
}

func NewContentAdminHandler(svcCtx *svc.ServiceContext) *ContentAdminHandler {
	return &ContentAdminHandler{svcCtx: svcCtx}
}

// ---------------- BLOG ARTICLES ----------------

func (h *ContentAdminHandler) ListArticles(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	search := c.Query("search")
	category := c.Query("category")
	publishedStr := c.Query("is_published")

	var isPublished *bool
	if publishedStr != "" && publishedStr != "all" {
		val := publishedStr == "true" || publishedStr == "1"
		isPublished = &val
	}

	l := admin.NewContentAdminLogic(c.Request.Context(), h.svcCtx)
	resp, err := l.ListArticlesAdmin(page, pageSize, search, category, isPublished)
	result.HttpResult(c, resp, err)
}

func (h *ContentAdminHandler) GetArticle(c *gin.Context) {
	id := c.Param("id")
	l := admin.NewContentAdminLogic(c.Request.Context(), h.svcCtx)
	resp, err := l.GetArticleAdmin(id)
	result.HttpResult(c, resp, err)
}

func (h *ContentAdminHandler) CreateArticle(c *gin.Context) {
	var req types.CreateArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		result.ParamErrorResult(c, err)
		return
	}

	l := admin.NewContentAdminLogic(c.Request.Context(), h.svcCtx)
	resp, err := l.CreateArticle(&req)
	result.HttpResult(c, resp, err)
}

func (h *ContentAdminHandler) UpdateArticle(c *gin.Context) {
	id := c.Param("id")
	var req types.UpdateArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		result.ParamErrorResult(c, err)
		return
	}

	l := admin.NewContentAdminLogic(c.Request.Context(), h.svcCtx)
	resp, err := l.UpdateArticle(id, &req)
	result.HttpResult(c, resp, err)
}

func (h *ContentAdminHandler) DeleteArticle(c *gin.Context) {
	id := c.Param("id")
	l := admin.NewContentAdminLogic(c.Request.Context(), h.svcCtx)
	err := l.DeleteArticle(id)
	result.HttpResult(c, gin.H{"deleted": true}, err)
}

// ---------------- HELP GUIDES ----------------

func (h *ContentAdminHandler) ListHelpGuides(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	search := c.Query("search")
	category := c.Query("category")
	serviceTarget := c.Query("service_target")

	l := admin.NewContentAdminLogic(c.Request.Context(), h.svcCtx)
	resp, err := l.ListHelpGuidesAdmin(page, pageSize, search, category, serviceTarget)
	result.HttpResult(c, resp, err)
}

func (h *ContentAdminHandler) GetHelpGuide(c *gin.Context) {
	id := c.Param("id")
	l := admin.NewContentAdminLogic(c.Request.Context(), h.svcCtx)
	resp, err := l.GetHelpGuideAdmin(id)
	result.HttpResult(c, resp, err)
}

func (h *ContentAdminHandler) CreateHelpGuide(c *gin.Context) {
	var req types.CreateHelpGuideRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		result.ParamErrorResult(c, err)
		return
	}

	l := admin.NewContentAdminLogic(c.Request.Context(), h.svcCtx)
	resp, err := l.CreateHelpGuide(&req)
	result.HttpResult(c, resp, err)
}

func (h *ContentAdminHandler) UpdateHelpGuide(c *gin.Context) {
	id := c.Param("id")
	var req types.UpdateHelpGuideRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		result.ParamErrorResult(c, err)
		return
	}

	l := admin.NewContentAdminLogic(c.Request.Context(), h.svcCtx)
	resp, err := l.UpdateHelpGuide(id, &req)
	result.HttpResult(c, resp, err)
}

func (h *ContentAdminHandler) DeleteHelpGuide(c *gin.Context) {
	id := c.Param("id")
	l := admin.NewContentAdminLogic(c.Request.Context(), h.svcCtx)
	err := l.DeleteHelpGuide(id)
	result.HttpResult(c, gin.H{"deleted": true}, err)
}

// ---------------- CATEGORIES ----------------

func (h *ContentAdminHandler) ListCategories(c *gin.Context) {
	catType := c.Query("type")
	l := admin.NewContentAdminLogic(c.Request.Context(), h.svcCtx)
	resp, err := l.ListCategories(catType)
	result.HttpResult(c, resp, err)
}

func (h *ContentAdminHandler) CreateCategory(c *gin.Context) {
	var req types.CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		result.ParamErrorResult(c, err)
		return
	}

	l := admin.NewContentAdminLogic(c.Request.Context(), h.svcCtx)
	resp, err := l.CreateCategory(&req)
	result.HttpResult(c, resp, err)
}

func (h *ContentAdminHandler) UpdateCategory(c *gin.Context) {
	id := c.Param("id")
	var req types.UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		result.ParamErrorResult(c, err)
		return
	}

	l := admin.NewContentAdminLogic(c.Request.Context(), h.svcCtx)
	err := l.UpdateCategory(id, &req)
	result.HttpResult(c, gin.H{"status": "ok"}, err)
}

func (h *ContentAdminHandler) DeleteCategory(c *gin.Context) {
	id := c.Param("id")
	l := admin.NewContentAdminLogic(c.Request.Context(), h.svcCtx)
	err := l.DeleteCategory(id)
	result.HttpResult(c, gin.H{"deleted": true}, err)
}

// ---------------- COMMENTS ----------------

func (h *ContentAdminHandler) ListComments(c *gin.Context) {
	articleID := c.Query("article_id")
	status := c.Query("status")

	l := admin.NewContentAdminLogic(c.Request.Context(), h.svcCtx)
	resp, err := l.ListComments(articleID, status)
	result.HttpResult(c, resp, err)
}

type UpdateCommentStatusReq struct {
	Status string `json:"status"`
	Reply  string `json:"reply"`
}

func (h *ContentAdminHandler) UpdateCommentStatus(c *gin.Context) {
	id := c.Param("id")
	var req UpdateCommentStatusReq
	_ = c.ShouldBindJSON(&req)

	l := admin.NewContentAdminLogic(c.Request.Context(), h.svcCtx)
	err := l.UpdateCommentStatus(id, req.Status, req.Reply)
	result.HttpResult(c, gin.H{"status": "ok"}, err)
}

func (h *ContentAdminHandler) DeleteComment(c *gin.Context) {
	id := c.Param("id")
	l := admin.NewContentAdminLogic(c.Request.Context(), h.svcCtx)
	err := l.DeleteComment(id)
	result.HttpResult(c, gin.H{"deleted": true}, err)
}

// ---------------- TEMPLATES ----------------

func (h *ContentAdminHandler) ListTemplates(c *gin.Context) {
	tmplType := c.Query("type")
	l := admin.NewContentAdminLogic(c.Request.Context(), h.svcCtx)
	resp, err := l.ListTemplatesAdmin(tmplType)
	result.HttpResult(c, resp, err)
}

func (h *ContentAdminHandler) CreateTemplate(c *gin.Context) {
	var req types.CreateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		result.ParamErrorResult(c, err)
		return
	}

	l := admin.NewContentAdminLogic(c.Request.Context(), h.svcCtx)
	resp, err := l.CreateTemplate(&req)
	result.HttpResult(c, resp, err)
}

func (h *ContentAdminHandler) UpdateTemplate(c *gin.Context) {
	id := c.Param("id")
	var req types.UpdateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		result.ParamErrorResult(c, err)
		return
	}

	l := admin.NewContentAdminLogic(c.Request.Context(), h.svcCtx)
	err := l.UpdateTemplate(id, &req)
	result.HttpResult(c, gin.H{"status": "ok"}, err)
}

func (h *ContentAdminHandler) DeleteTemplate(c *gin.Context) {
	id := c.Param("id")
	l := admin.NewContentAdminLogic(c.Request.Context(), h.svcCtx)
	err := l.DeleteTemplate(id)
	result.HttpResult(c, gin.H{"deleted": true}, err)
}

// ---------------- SERVICE TARGETS ----------------

func (h *ContentAdminHandler) ListServices(c *gin.Context) {
	l := admin.NewContentAdminLogic(c.Request.Context(), h.svcCtx)
	resp, err := l.ListServiceTargetsAdmin()
	result.HttpResult(c, resp, err)
}

func (h *ContentAdminHandler) CreateServiceTarget(c *gin.Context) {
	var req types.CreateServiceTargetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		result.ParamErrorResult(c, err)
		return
	}

	l := admin.NewContentAdminLogic(c.Request.Context(), h.svcCtx)
	resp, err := l.CreateServiceTarget(&req)
	result.HttpResult(c, resp, err)
}

func (h *ContentAdminHandler) UpdateServiceTarget(c *gin.Context) {
	id := c.Param("id")
	var req types.CreateServiceTargetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		result.ParamErrorResult(c, err)
		return
	}

	l := admin.NewContentAdminLogic(c.Request.Context(), h.svcCtx)
	err := l.UpdateServiceTarget(id, &req)
	result.HttpResult(c, gin.H{"status": "ok"}, err)
}

func (h *ContentAdminHandler) DeleteServiceTarget(c *gin.Context) {
	id := c.Param("id")
	l := admin.NewContentAdminLogic(c.Request.Context(), h.svcCtx)
	err := l.DeleteServiceTarget(id)
	result.HttpResult(c, gin.H{"deleted": true}, err)
}
