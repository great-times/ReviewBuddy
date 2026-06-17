package api

import (
	"github.com/gin-gonic/gin"

	"reviewbuddy/internal/model"
	"reviewbuddy/internal/service/knowledge"
)

type KnowledgeHandler struct{ svc *knowledge.Service }

func NewKnowledgeHandler(s *knowledge.Service) *KnowledgeHandler { return &KnowledgeHandler{svc: s} }

func (h *KnowledgeHandler) Register(r *gin.RouterGroup) {
	r.GET("/knowledge/issues", h.listIssues)
	r.POST("/knowledge/issues", ReadWriteRequired(), h.addIssue)
	r.GET("/knowledge/rules", h.listRules)
	r.POST("/knowledge/rules", ReadWriteRequired(), h.addRule)
	r.GET("/metrics/quality", h.metrics)
}

func (h *KnowledgeHandler) listIssues(c *gin.Context) {
	items, err := h.svc.ListIssues()
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, items)
}

func (h *KnowledgeHandler) addIssue(c *gin.Context) {
	var in model.ReviewIssue
	if err := c.ShouldBindJSON(&in); err != nil {
		badRequest(c, err)
		return
	}
	v, err := h.svc.AddIssue(&in)
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, v)
}

func (h *KnowledgeHandler) listRules(c *gin.Context) {
	items, err := h.svc.ListRules(false)
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, items)
}

func (h *KnowledgeHandler) addRule(c *gin.Context) {
	var in model.KnowledgeRule
	if err := c.ShouldBindJSON(&in); err != nil {
		badRequest(c, err)
		return
	}
	v, err := h.svc.AddRule(&in)
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, v)
}

func (h *KnowledgeHandler) metrics(c *gin.Context) {
	issues, rules, err := h.svc.Counts()
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, gin.H{"issueCount": issues, "ruleCount": rules})
}
