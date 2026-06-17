package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"reviewbuddy/internal/model"
	"reviewbuddy/internal/service/auth"
	"reviewbuddy/internal/service/user"
)

type UserHandler struct {
	svc  *user.Service
	auth *auth.Service
}

func NewUserHandler(s *user.Service, a *auth.Service) *UserHandler {
	return &UserHandler{svc: s, auth: a}
}

func (h *UserHandler) RegisterReadOnly(r *gin.RouterGroup) {
	g := r.Group("/users")
	g.GET("", h.list)
}

func (h *UserHandler) RegisterAdmin(r *gin.RouterGroup) {
	g := r.Group("/users")
	g.POST("", h.create)
	g.PUT("/:id", h.update)
	g.DELETE("/:id", h.delete)
}

func (h *UserHandler) list(c *gin.Context) {
	items, err := h.svc.List()
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, items)
}

func (h *UserHandler) create(c *gin.Context) {
	var in model.User
	if err := c.ShouldBindJSON(&in); err != nil {
		badRequest(c, err)
		return
	}
	v, err := h.svc.Create(&in)
	if err != nil {
		badRequest(c, err)
		return
	}
	ok(c, v)
}

func (h *UserHandler) update(c *gin.Context) {
	var in model.User
	if err := c.ShouldBindJSON(&in); err != nil {
		badRequest(c, err)
		return
	}
	v, err := h.svc.Update(c.Param("id"), &in)
	if err != nil {
		badRequest(c, err)
		return
	}
	ok(c, v)
}

func (h *UserHandler) delete(c *gin.Context) {
	actor := CurrentUser(c)
	actorID := ""
	if actor != nil {
		actorID = actor.ID
	}
	if err := h.auth.DeleteUser(actorID, c.Param("id")); err != nil {
		if err == auth.ErrCannotDeleteSelf || err == auth.ErrLastAdmin {
			badRequest(c, err)
			return
		}
		fail(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
