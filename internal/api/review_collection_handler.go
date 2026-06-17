package api

import (
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"

	"reviewbuddy/internal/model"
	"reviewbuddy/internal/service/guide"
)

type ReviewCollectionHandler struct{ svc *guide.CollectionService }

func NewReviewCollectionHandler(s *guide.CollectionService) *ReviewCollectionHandler {
	return &ReviewCollectionHandler{svc: s}
}

func (h *ReviewCollectionHandler) Register(r *gin.RouterGroup) {
	g := r.Group("/review-collections")
	g.GET("", h.list)
	g.POST("", ReadWriteRequired(), h.create)
	g.PUT("/:id", ReadWriteRequired(), h.update)
	g.GET("/:id/export-eml", h.exportEML)
}

func (h *ReviewCollectionHandler) list(c *gin.Context) {
	items, err := h.svc.List()
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, items)
}

func (h *ReviewCollectionHandler) create(c *gin.Context) {
	var in model.ReviewCollection
	if err := c.ShouldBindJSON(&in); err != nil {
		badRequest(c, err)
		return
	}
	if u := CurrentUser(c); u != nil {
		in.CreatedBy = u.Username
	}
	item, err := h.svc.Create(&in)
	if err != nil {
		badRequest(c, err)
		return
	}
	ok(c, item)
}

func (h *ReviewCollectionHandler) update(c *gin.Context) {
	var in model.ReviewCollection
	if err := c.ShouldBindJSON(&in); err != nil {
		badRequest(c, err)
		return
	}
	item, err := h.svc.Update(c.Param("id"), &in)
	if err != nil {
		badRequest(c, err)
		return
	}
	if item == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	ok(c, item)
}

func (h *ReviewCollectionHandler) exportEML(c *gin.Context) {
	filename, data, err := h.svc.ExportEML(c.Param("id"))
	if err != nil {
		fail(c, err)
		return
	}
	c.Header("Content-Type", "message/rfc822")
	c.Header("Content-Disposition", "attachment; filename*=UTF-8''"+url.PathEscape(filename))
	c.Data(http.StatusOK, "message/rfc822", data)
}
