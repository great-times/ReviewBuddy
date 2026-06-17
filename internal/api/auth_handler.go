package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"reviewbuddy/internal/model"
	"reviewbuddy/internal/service/auth"
)

type AuthHandler struct{ svc *auth.Service }

func NewAuthHandler(s *auth.Service) *AuthHandler { return &AuthHandler{svc: s} }

func (h *AuthHandler) Register(r *gin.RouterGroup) {
	g := r.Group("/auth")
	g.POST("/register", h.register)
	g.POST("/login", h.login)
	g.POST("/logout", h.AuthRequired(), h.logout)
	g.GET("/me", h.AuthRequired(), h.me)
}

func (h *AuthHandler) AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		u, err := h.svc.UserByToken(bearerToken(c))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "请先登录"})
			return
		}
		c.Set("user", u)
		c.Next()
	}
}

func (h *AuthHandler) AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		u := CurrentUser(c)
		if u == nil || u.Role != "admin" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "仅管理员可访问"})
			return
		}
		c.Next()
	}
}

func ReadWriteRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		u := CurrentUser(c)
		if u != nil && u.Role == "readonly" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "只读角色不能执行增删改操作"})
			return
		}
		c.Next()
	}
}

func CurrentUser(c *gin.Context) *model.User {
	v, _ := c.Get("user")
	u, _ := v.(*model.User)
	return u
}

func bearerToken(c *gin.Context) string {
	authz := c.GetHeader("Authorization")
	if strings.HasPrefix(authz, "Bearer ") {
		return strings.TrimSpace(strings.TrimPrefix(authz, "Bearer "))
	}
	return c.Query("token")
}

func (h *AuthHandler) register(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, err)
		return
	}
	result, err := h.svc.Register(req.Username, req.Password)
	if err != nil {
		status := http.StatusBadRequest
		if err == auth.ErrUsernameTaken {
			status = http.StatusConflict
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *AuthHandler) login(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, err)
		return
	}
	result, err := h.svc.Login(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *AuthHandler) logout(c *gin.Context) {
	_ = h.svc.Logout(bearerToken(c))
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *AuthHandler) me(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"user": CurrentUser(c)})
}
