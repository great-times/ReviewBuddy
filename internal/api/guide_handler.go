package api

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"changebuddy/internal/model"
	"changebuddy/internal/service/agent"
	"changebuddy/internal/service/guide"
)

type GuideHandler struct{ svc *guide.Service }

func NewGuideHandler(s *guide.Service) *GuideHandler { return &GuideHandler{svc: s} }

func (h *GuideHandler) Register(r *gin.RouterGroup) {
	g := r.Group("/guides")
	g.GET("", h.list)
	g.POST("", ReadWriteRequired(), h.create)
	g.GET("/:id", h.get)
	g.PUT("/:id", ReadWriteRequired(), h.update)
	g.POST("/generate", h.generate)
	g.POST("/precheck", h.precheck)
	g.POST("/precheck/stream", h.precheckStream)
}

func (h *GuideHandler) list(c *gin.Context) {
	items, err := h.svc.List(c.Query("status"))
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, items)
}

func (h *GuideHandler) get(c *gin.Context) {
	g, err := h.svc.Get(c.Param("id"))
	if err != nil {
		fail(c, err)
		return
	}
	if g == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	ok(c, g)
}

func (h *GuideHandler) create(c *gin.Context) {
	var in model.Guide
	if err := c.ShouldBindJSON(&in); err != nil {
		badRequest(c, err)
		return
	}
	if u := CurrentUser(c); u != nil {
		in.CreatedBy = u.Username
	}
	g, err := h.svc.Create(&in)
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, g)
}

func (h *GuideHandler) update(c *gin.Context) {
	var in model.Guide
	if err := c.ShouldBindJSON(&in); err != nil {
		badRequest(c, err)
		return
	}
	g, err := h.svc.Update(c.Param("id"), &in)
	if err != nil {
		fail(c, err)
		return
	}
	if g == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	ok(c, g)
}

// setupSSE 配置 SSE 响应头并返回 flusher
func setupSSE(c *gin.Context) (http.Flusher, bool) {
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")
	flusher, ok := c.Writer.(http.Flusher)
	return flusher, ok
}

// generate 通过 SSE 流式返回生成内容
func (h *GuideHandler) generate(c *gin.Context) {
	var req guide.GenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, err)
		return
	}
	flusher, ok := setupSSE(c)
	if !ok {
		fail(c, io.ErrUnexpectedEOF)
		return
	}

	_, err := h.svc.GenerateStream(c.Request.Context(), &req, func(ch agent.Chunk) {
		if ch.Done {
			c.SSEvent("done", "")
		} else {
			c.SSEvent("chunk", ch.Delta)
		}
		flusher.Flush()
	})
	if err != nil {
		c.SSEvent("error", err.Error())
		flusher.Flush()
	}
}

// precheckStream 通过 SSE 流式返回预审：chunk 为模型原始片段，结束时以 result 事件回传结构化结果
func (h *GuideHandler) precheckStream(c *gin.Context) {
	var body struct {
		Content string             `json:"content"`
		Images  []agent.ImageInput `json:"images"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		badRequest(c, err)
		return
	}
	flusher, ok := setupSSE(c)
	if !ok {
		fail(c, io.ErrUnexpectedEOF)
		return
	}

	res, err := h.svc.AIPrecheckStream(c.Request.Context(), body.Content, body.Images, func(ch agent.Chunk) {
		if !ch.Done && ch.Delta != "" {
			c.SSEvent("chunk", ch.Delta)
			flusher.Flush()
		}
	})
	if err != nil {
		c.SSEvent("error", err.Error())
		flusher.Flush()
		return
	}
	if payload, mErr := json.Marshal(res); mErr == nil {
		c.SSEvent("result", string(payload))
	}
	c.SSEvent("done", "")
	flusher.Flush()
}

func (h *GuideHandler) precheck(c *gin.Context) {
	var body struct {
		Content string             `json:"content"`
		Images  []agent.ImageInput `json:"images"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		badRequest(c, err)
		return
	}
	res, err := h.svc.AIPrecheck(c.Request.Context(), body.Content, body.Images)
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, res)
}
