package api

import (
	"github.com/gin-gonic/gin"

	"reviewbuddy/internal/service/dashboard"
)

type DashboardHandler struct{ svc *dashboard.Service }

func NewDashboardHandler(s *dashboard.Service) *DashboardHandler { return &DashboardHandler{svc: s} }

func (h *DashboardHandler) Register(r *gin.RouterGroup) {
	r.GET("/dashboard", h.summary)
}

func (h *DashboardHandler) summary(c *gin.Context) {
	u := CurrentUser(c)
	userID := ""
	if u != nil {
		userID = u.ID
	}
	v, err := h.svc.Summary(userID)
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, v)
}
