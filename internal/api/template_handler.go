package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"changebuddy/internal/model"
	"changebuddy/internal/service/template"
)

type TemplateHandler struct{ svc *template.Service }

func NewTemplateHandler(s *template.Service) *TemplateHandler { return &TemplateHandler{svc: s} }

func (h *TemplateHandler) Register(r *gin.RouterGroup) {
	r.GET("/template-libraries", h.listLibraries)
	r.POST("/template-libraries", ReadWriteRequired(), h.createLibrary)

	g := r.Group("/templates")
	g.GET("", h.list)
	g.POST("", ReadWriteRequired(), h.create)
	g.GET("/:id", h.get)
	g.PUT("/:id", ReadWriteRequired(), h.update)
	g.GET("/:id/versions", h.versions)
}

func (h *TemplateHandler) list(c *gin.Context) {
	items, err := h.svc.List(c.Query("libraryId"), c.Query("category"))
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, items)
}

func (h *TemplateHandler) listLibraries(c *gin.Context) {
	items, err := h.svc.ListLibraries()
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, items)
}

func (h *TemplateHandler) createLibrary(c *gin.Context) {
	var in model.TemplateLibrary
	if err := c.ShouldBindJSON(&in); err != nil {
		badRequest(c, err)
		return
	}
	l, err := h.svc.CreateLibrary(&in)
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, l)
}

func (h *TemplateHandler) get(c *gin.Context) {
	t, err := h.svc.Get(c.Param("id"))
	if err != nil {
		fail(c, err)
		return
	}
	if t == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	ok(c, t)
}

func (h *TemplateHandler) create(c *gin.Context) {
	var in model.Template
	if err := c.ShouldBindJSON(&in); err != nil {
		badRequest(c, err)
		return
	}
	t, err := h.svc.Create(&in)
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, t)
}

func (h *TemplateHandler) update(c *gin.Context) {
	var in model.Template
	if err := c.ShouldBindJSON(&in); err != nil {
		badRequest(c, err)
		return
	}
	t, err := h.svc.Update(c.Param("id"), &in)
	if err != nil {
		fail(c, err)
		return
	}
	if t == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	ok(c, t)
}

func (h *TemplateHandler) versions(c *gin.Context) {
	items, err := h.svc.ListVersions(c.Param("id"))
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, items)
}
