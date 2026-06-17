package guide

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	"reviewbuddy/internal/model"
	"reviewbuddy/internal/repo"
	"reviewbuddy/internal/service/agent"
	"reviewbuddy/internal/service/knowledge"
)

type Service struct {
	repo      *repo.GuideRepo
	tplRepo   *repo.TemplateRepo
	agent     agent.Adapter
	knowledge *knowledge.Service
}

func NewService(r *repo.GuideRepo, tpl *repo.TemplateRepo, ag agent.Adapter, kn *knowledge.Service) *Service {
	return &Service{repo: r, tplRepo: tpl, agent: ag, knowledge: kn}
}

func (s *Service) List(status string) ([]model.Guide, error) { return s.repo.List(status) }
func (s *Service) Get(id string) (*model.Guide, error)       { return s.repo.Get(id) }

func (s *Service) Update(id string, in *model.Guide) (*model.Guide, error) {
	cur, err := s.repo.Get(id)
	if err != nil || cur == nil {
		return cur, err
	}
	now := time.Now().Format(time.RFC3339)
	if in.Title != "" {
		cur.Title = in.Title
	}
	cur.Content = in.Content
	if in.Status != "" {
		cur.Status = in.Status
	}
	if in.RiskLevel != "" {
		cur.RiskLevel = in.RiskLevel
	}
	cur.UpdatedAt = now
	if err := s.repo.Update(cur); err != nil {
		return nil, err
	}
	return cur, nil
}

// GenerateRequest AI 生成请求
type GenerateRequest struct {
	Title      string             `json:"title"`
	TemplateID string             `json:"templateId"`
	ChangeType string             `json:"changeType"`
	RiskLevel  string             `json:"riskLevel"`
	Variables  map[string]string  `json:"variables"`
	Context    string             `json:"context"`
	Images     []agent.ImageInput `json:"images"`
}

// buildPrompt 组装生成 Prompt：模板 + 变量 + 上下文 + RAG 召回的历史经验
func (s *Service) buildPrompt(req *GenerateRequest) ([]agent.Message, *model.Template, error) {
	var tplContent string
	var tpl *model.Template
	if req.TemplateID != "" {
		t, err := s.tplRepo.Get(req.TemplateID)
		if err != nil {
			return nil, nil, err
		}
		if t != nil {
			tpl = t
			tplContent = t.Content
		}
	}

	recall := ""
	if s.knowledge != nil {
		recall = s.knowledge.Recall(req.ChangeType+" "+req.Context, 5)
	}

	var sb strings.Builder
	sb.WriteString("请基于以下模板与上下文，生成一份结构清晰、可评审的「评审材料」（Markdown 格式）。\n")
	sb.WriteString("务必包含：背景与目标、核心内容、风险与约束、评审关注点、结论与后续动作。\n\n")
	if tplContent != "" {
		sb.WriteString("### 模板\n" + tplContent + "\n\n")
	}
	if req.ChangeType != "" {
		sb.WriteString("### 材料类型\n" + req.ChangeType + "\n\n")
	}
	if len(req.Variables) > 0 {
		vars, _ := json.Marshal(req.Variables)
		sb.WriteString("### 变量\n" + string(vars) + "\n\n")
	}
	if req.Context != "" {
		sb.WriteString("### 上下文\n" + req.Context + "\n\n")
	}
	if recall != "" {
		sb.WriteString("### 请特别参考以下历史经验，避免重复同类问题\n" + recall + "\n")
	}

	msgs := []agent.Message{
		{Role: "system", Content: "你是资深的评审专家，输出严谨、可落地的评审材料。"},
		{Role: "user", Content: sb.String()},
	}
	return msgs, tpl, nil
}

// GenerateStream 流式生成，边生成边回调
func (s *Service) GenerateStream(ctx context.Context, req *GenerateRequest, onChunk func(agent.Chunk)) (string, error) {
	msgs, tpl, err := s.buildPrompt(req)
	if err != nil {
		return "", err
	}
	var buf strings.Builder
	err = s.agent.CompleteStream(ctx, &agent.CompletionRequest{Messages: msgs, Images: req.Images, Temperature: 0.3}, func(c agent.Chunk) {
		buf.WriteString(c.Delta)
		onChunk(c)
	})
	if err != nil {
		return "", err
	}
	if tpl != nil {
		_ = s.tplRepo.IncrUsage(tpl.ID)
	}
	return buf.String(), nil
}

// Create 落库一份评审材料（生成完成后由前端回写，或非流式生成后调用）
func (s *Service) Create(g *model.Guide) (*model.Guide, error) {
	if g.Title == "" {
		return nil, errors.New("title is required")
	}
	now := time.Now().Format(time.RFC3339)
	g.ID = uuid.NewString()
	g.Status = "draft"
	g.CurrentVersion = 1
	if g.RiskLevel == "" {
		g.RiskLevel = "medium"
	}
	g.CreatedAt = now
	g.UpdatedAt = now
	if err := s.repo.Create(g); err != nil {
		return nil, err
	}
	return g, nil
}

// buildPrecheckMessages 组装预审 Prompt（结合 RAG 召回的历史经验）
func (s *Service) buildPrecheckMessages(content string) []agent.Message {
	recall := ""
	if s.knowledge != nil {
		recall = s.knowledge.Recall(content, 8)
	}
	prompt := "请审查以下评审材料，找出潜在问题（尤其是目标不清、依据不足、风险点缺失、结论不可执行）。\n" +
		"以 JSON 返回，格式：{\"summary\":\"...\",\"findings\":[{\"severity\":\"info|warning|critical\",\"category\":\"...\",\"excerpt\":\"...\",\"problem\":\"...\",\"suggestion\":\"...\"}]}。\n"
	if recall != "" {
		prompt += "参考已沉淀经验：\n" + recall + "\n"
	}
	prompt += "\n### 待审查内容\n" + content
	return []agent.Message{
		{Role: "system", Content: "你是严格的评审专家，只输出 JSON。"},
		{Role: "user", Content: prompt},
	}
}

// parsePrecheck 把模型原始输出解析为结构化结果，解析失败时降级为摘要文本，保证不阻塞
func parsePrecheck(raw string) *model.AIPrecheckResult {
	var res model.AIPrecheckResult
	if jsonStr := extractJSON(raw); jsonStr != "" {
		if err := json.Unmarshal([]byte(jsonStr), &res); err == nil {
			return &res
		}
	}
	res.Summary = strings.TrimSpace(raw)
	return &res
}

// AIPrecheck 用 Agent 结合规则做预审
func (s *Service) AIPrecheck(ctx context.Context, content string, images []agent.ImageInput) (*model.AIPrecheckResult, error) {
	msgs := s.buildPrecheckMessages(content)
	raw, err := s.agent.Complete(ctx, &agent.CompletionRequest{Messages: msgs, Images: images, Temperature: 0})
	if err != nil {
		return nil, err
	}
	return parsePrecheck(raw), nil
}

// AIPrecheckStream 流式预审：边生成边回调原始片段，结束后解析为结构化结果返回
func (s *Service) AIPrecheckStream(ctx context.Context, content string, images []agent.ImageInput, onChunk func(agent.Chunk)) (*model.AIPrecheckResult, error) {
	msgs := s.buildPrecheckMessages(content)
	var buf strings.Builder
	err := s.agent.CompleteStream(ctx, &agent.CompletionRequest{Messages: msgs, Images: images, Temperature: 0}, func(c agent.Chunk) {
		buf.WriteString(c.Delta)
		onChunk(c)
	})
	if err != nil {
		return nil, err
	}
	return parsePrecheck(buf.String()), nil
}

func extractJSON(s string) string {
	start := strings.Index(s, "{")
	end := strings.LastIndex(s, "}")
	if start >= 0 && end > start {
		return s[start : end+1]
	}
	return ""
}
