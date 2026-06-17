package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"reviewbuddy/internal/model"
	"reviewbuddy/internal/service/guide"
)

type ReviewHandler struct{ svc *guide.ReviewService }

func NewReviewHandler(s *guide.ReviewService) *ReviewHandler { return &ReviewHandler{svc: s} }

func (h *ReviewHandler) Register(r *gin.RouterGroup) {
	r.GET("/guides/:id/reviews", h.listByGuide)
	r.POST("/guides/:id/reviews", ReadWriteRequired(), h.create)
	r.POST("/reviews/:rid/decision", ReadWriteRequired(), h.decide)
	r.GET("/reviews/:rid/comments", h.listComments)
	r.POST("/reviews/:rid/comments", ReadWriteRequired(), h.addComment)
}

func (h *ReviewHandler) listByGuide(c *gin.Context) {
	items, err := h.svc.ListByGuide(c.Param("id"))
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, items)
}

func (h *ReviewHandler) create(c *gin.Context) {
	var body struct {
		Reviewer       string `json:"reviewer"`
		ReviewerUserID string `json:"reviewerUserId"`
	}
	_ = c.ShouldBindJSON(&body)
	v, err := h.svc.Create(c.Param("id"), body.ReviewerUserID, body.Reviewer)
	if err != nil {
		badRequest(c, err)
		return
	}
	ok(c, v)
}

func (h *ReviewHandler) decide(c *gin.Context) {
	var body struct {
		Status string `json:"status"`
		Note   string `json:"note"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		badRequest(c, err)
		return
	}
	v, err := h.svc.Decide(c.Param("rid"), body.Status, body.Note)
	if err != nil {
		badRequest(c, err)
		return
	}
	if v == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	ok(c, v)
}

func (h *ReviewHandler) listComments(c *gin.Context) {
	items, err := h.svc.ListComments(c.Param("rid"))
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, items)
}

func (h *ReviewHandler) addComment(c *gin.Context) {
	var in model.ReviewComment
	if err := c.ShouldBindJSON(&in); err != nil {
		badRequest(c, err)
		return
	}
	in.ReviewID = c.Param("rid")
	v, err := h.svc.AddComment(&in)
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, v)
}
