package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"reviewbuddy/internal/model"
	"reviewbuddy/internal/service/reviewconfig"
)

type ReviewConfigHandler struct{ svc *reviewconfig.Service }

func NewReviewConfigHandler(s *reviewconfig.Service) *ReviewConfigHandler {
	return &ReviewConfigHandler{svc: s}
}

func (h *ReviewConfigHandler) RegisterReadOnly(r *gin.RouterGroup) {
	r.GET("/review-roles", h.listRoles)
	r.GET("/review-domains", h.listDomains)
	r.GET("/review-domains/:id/role-users", h.listDomainRoleUsers)
	r.GET("/review-scenarios", h.listScenarios)
	r.GET("/me/domains", h.myDomains)
}

func (h *ReviewConfigHandler) RegisterAdmin(r *gin.RouterGroup) {
	r.POST("/review-roles", h.createRole)
	r.PUT("/review-roles/:key", h.updateRole)
	r.DELETE("/review-roles/:key", h.deleteRole)
	r.POST("/review-domains", h.createDomain)
	r.PUT("/review-domains/:id", h.updateDomain)
	r.DELETE("/review-domains/:id", h.deleteDomain)
	r.PUT("/review-domains/:id/role-users/:roleKey", h.saveDomainRoleUsers)
	r.GET("/users-with-domains", h.usersWithDomains)
	r.GET("/users/:id/domains", h.userDomains)
	r.PUT("/users/:id/domains", h.saveUserDomains)
	r.POST("/review-scenarios", h.createScenario)
	r.PUT("/review-scenarios/:id", h.updateScenario)
	r.DELETE("/review-scenarios/:id", h.deleteScenario)
}

func (h *ReviewConfigHandler) listRoles(c *gin.Context) {
	items, err := h.svc.ListRoles()
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, items)
}

func (h *ReviewConfigHandler) createRole(c *gin.Context) {
	var in model.ReviewRole
	if err := c.ShouldBindJSON(&in); err != nil {
		badRequest(c, err)
		return
	}
	item, err := h.svc.CreateRole(&in)
	if err != nil {
		badRequest(c, err)
		return
	}
	ok(c, item)
}

func (h *ReviewConfigHandler) updateRole(c *gin.Context) {
	var in model.ReviewRole
	if err := c.ShouldBindJSON(&in); err != nil {
		badRequest(c, err)
		return
	}
	item, err := h.svc.UpdateRole(c.Param("key"), &in)
	if err != nil {
		badRequest(c, err)
		return
	}
	ok(c, item)
}

func (h *ReviewConfigHandler) deleteRole(c *gin.Context) {
	if err := h.svc.DeleteRole(c.Param("key")); err != nil {
		badRequest(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *ReviewConfigHandler) listDomains(c *gin.Context) {
	items, err := h.svc.ListDomains()
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, items)
}

func (h *ReviewConfigHandler) createDomain(c *gin.Context) {
	var in model.ReviewDomain
	if err := c.ShouldBindJSON(&in); err != nil {
		badRequest(c, err)
		return
	}
	item, err := h.svc.SaveDomain("", &in)
	if err != nil {
		badRequest(c, err)
		return
	}
	ok(c, item)
}

func (h *ReviewConfigHandler) updateDomain(c *gin.Context) {
	var in model.ReviewDomain
	if err := c.ShouldBindJSON(&in); err != nil {
		badRequest(c, err)
		return
	}
	item, err := h.svc.SaveDomain(c.Param("id"), &in)
	if err != nil {
		badRequest(c, err)
		return
	}
	ok(c, item)
}

func (h *ReviewConfigHandler) deleteDomain(c *gin.Context) {
	if err := h.svc.DeleteDomain(c.Param("id")); err != nil {
		badRequest(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *ReviewConfigHandler) listDomainRoleUsers(c *gin.Context) {
	items, err := h.svc.ListDomainRoleUsers(c.Param("id"))
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, items)
}

func (h *ReviewConfigHandler) saveDomainRoleUsers(c *gin.Context) {
	var in model.DomainRoleUsers
	if err := c.ShouldBindJSON(&in); err != nil {
		badRequest(c, err)
		return
	}
	in.DomainID = c.Param("id")
	in.RoleKey = c.Param("roleKey")
	item, err := h.svc.SaveDomainRoleUsers(&in)
	if err != nil {
		badRequest(c, err)
		return
	}
	ok(c, item)
}

func (h *ReviewConfigHandler) myDomains(c *gin.Context) {
	u := CurrentUser(c)
	if u == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "请先登录"})
		return
	}
	item, err := h.svc.ListUserDomains(u.ID)
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, item)
}

func (h *ReviewConfigHandler) usersWithDomains(c *gin.Context) {
	items, err := h.svc.ListUsersWithDomains()
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, items)
}

func (h *ReviewConfigHandler) userDomains(c *gin.Context) {
	item, err := h.svc.ListUserDomains(c.Param("id"))
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, item)
}

func (h *ReviewConfigHandler) saveUserDomains(c *gin.Context) {
	var in model.UserDomains
	if err := c.ShouldBindJSON(&in); err != nil {
		badRequest(c, err)
		return
	}
	in.UserID = c.Param("id")
	item, err := h.svc.SaveUserDomains(&in)
	if err != nil {
		badRequest(c, err)
		return
	}
	ok(c, item)
}

func (h *ReviewConfigHandler) listScenarios(c *gin.Context) {
	items, err := h.svc.ListScenarios()
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, items)
}

func (h *ReviewConfigHandler) createScenario(c *gin.Context) {
	var in model.ReviewScenario
	if err := c.ShouldBindJSON(&in); err != nil {
		badRequest(c, err)
		return
	}
	item, err := h.svc.SaveScenario("", &in)
	if err != nil {
		badRequest(c, err)
		return
	}
	ok(c, item)
}

func (h *ReviewConfigHandler) updateScenario(c *gin.Context) {
	var in model.ReviewScenario
	if err := c.ShouldBindJSON(&in); err != nil {
		badRequest(c, err)
		return
	}
	item, err := h.svc.SaveScenario(c.Param("id"), &in)
	if err != nil {
		badRequest(c, err)
		return
	}
	ok(c, item)
}

func (h *ReviewConfigHandler) deleteScenario(c *gin.Context) {
	if err := h.svc.DeleteScenario(c.Param("id")); err != nil {
		badRequest(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
