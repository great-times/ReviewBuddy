package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"changebuddy/internal/model"
	"changebuddy/internal/service/settings"
)

type SettingsHandler struct{ svc *settings.Service }

func NewSettingsHandler(s *settings.Service) *SettingsHandler { return &SettingsHandler{svc: s} }

func (h *SettingsHandler) Register(r *gin.RouterGroup) {
	r.GET("/settings/agent", h.getAgent)
	r.PUT("/settings/agent", ReadWriteRequired(), h.updateAgent)
	r.GET("/agent/types", h.agentTypes)
	r.POST("/agent/health", h.agentHealth)
}

func (h *SettingsHandler) getAgent(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"data": h.svc.AgentSettings()})
}

func (h *SettingsHandler) updateAgent(c *gin.Context) {
	var in model.AgentSettings
	if err := c.ShouldBindJSON(&in); err != nil {
		badRequest(c, err)
		return
	}
	out, err := h.svc.UpdateAgentSettings(c.Request.Context(), in)
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, out)
}

func (h *SettingsHandler) agentTypes(c *gin.Context) {
	ok(c, settings.AgentTypes())
}

func (h *SettingsHandler) agentHealth(c *gin.Context) {
	cfg := h.svc.AgentConfig()
	if cfg.Provider == "mock" {
		ok(c, gin.H{"healthy": true, "message": "Mock Agent 可用"})
		return
	}
	if cfg.BaseURL == "" || cfg.Model == "" {
		ok(c, gin.H{"healthy": false, "error": "请先配置 API Base URL 和模型"})
		return
	}
	ok(c, gin.H{"healthy": true, "message": "配置已完整，保存后生成与预审请求会使用该 Agent"})
}
