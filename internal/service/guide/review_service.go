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

type ReviewService struct {
	repo      *repo.ReviewRepo
	guideRepo *repo.GuideRepo
	tplRepo   *repo.TemplateRepo
	userRepo  *repo.UserRepo
	agent     agent.Adapter
	knowledge *knowledge.Service
}

func NewReviewService(r *repo.ReviewRepo, g *repo.GuideRepo, tpl *repo.TemplateRepo, users *repo.UserRepo, ag agent.Adapter, kn *knowledge.Service) *ReviewService {
	return &ReviewService{repo: r, guideRepo: g, tplRepo: tpl, userRepo: users, agent: ag, knowledge: kn}
}

func (s *ReviewService) ListByGuide(guideID string) ([]model.Review, error) {
	return s.repo.ListByGuide(guideID)
}

func (s *ReviewService) Create(guideID, reviewerUserID, reviewer string) (*model.Review, error) {
	g, err := s.guideRepo.Get(guideID)
	if err != nil {
		return nil, err
	}
	if g == nil {
		return nil, errors.New("guide not found")
	}
	if reviewerUserID != "" && s.userRepo != nil {
		u, err := s.userRepo.Get(reviewerUserID)
		if err != nil {
			return nil, err
		}
		reviewer = u.Username
	}
	if strings.TrimSpace(reviewer) == "" {
		return nil, errors.New("reviewer is required")
	}
	v := &model.Review{
		ID: uuid.NewString(), GuideID: guideID, GuideVersion: g.CurrentVersion,
		Reviewer: reviewer, ReviewerUserID: reviewerUserID, Status: "pending", CreatedAt: time.Now().Format(time.RFC3339),
	}
	if err := s.repo.Create(v); err != nil {
		return nil, err
	}
	// 评审材料进入评审中
	g.Status = "reviewing"
	g.UpdatedAt = time.Now().Format(time.RFC3339)
	_ = s.guideRepo.Update(g)
	return v, nil
}

func (s *ReviewService) Decide(reviewID, status, note string) (*model.Review, error) {
	if status != "approved" && status != "rejected" {
		return nil, errors.New("status must be approved or rejected")
	}
	rv, err := s.repo.Get(reviewID)
	if err != nil || rv == nil {
		return rv, err
	}
	now := time.Now().Format(time.RFC3339)
	if err := s.repo.Decide(reviewID, status, note, now); err != nil {
		return nil, err
	}
	rv.Status = status
	rv.DecisionNote = note
	rv.FinishedAt = now
	// 同步评审材料状态，并把人工评审意见沉淀为 AI 规则候选与模板演进建议。
	if g, _ := s.guideRepo.Get(rv.GuideID); g != nil {
		if status == "approved" {
			g.Status = "approved"
		} else {
			g.Status = "draft"
		}
		g.UpdatedAt = now
		_ = s.guideRepo.Update(g)
		s.createLearningSuggestion(context.Background(), rv, g, note)
	}
	return rv, nil
}

func (s *ReviewService) createLearningSuggestion(ctx context.Context, rv *model.Review, g *model.Guide, note string) {
	note = strings.TrimSpace(note)
	if note == "" || s.knowledge == nil {
		return
	}
	suggestion := s.fallbackLearningSuggestion(rv, g, note)
	if s.agent != nil {
		if ai, err := s.analyzeLearningSuggestion(ctx, rv, g, note); err == nil {
			suggestion = ai
		}
	}
	_, _ = s.knowledge.AddLearningSuggestion(suggestion)
}

func (s *ReviewService) analyzeLearningSuggestion(ctx context.Context, rv *model.Review, g *model.Guide, note string) (*model.ReviewLearningSuggestion, error) {
	prompt := `请把以下人工评审意见提炼为可复用的评审知识候选，不要直接修改模板。
只输出 JSON，格式：
{"summary":"一句话总结","issues":[{"category":"...","triggerCondition":"...","problemDesc":"...","correctPractice":"...","changeType":"..."}],"rules":[{"title":"...","ruleType":"rule","pattern":"...","suggestion":"..."}],"templateSuggestion":"建议追加到模板中的 Markdown 条目，若无则为空"}

评审材料标题：` + g.Title + `
风险等级：` + g.RiskLevel + `
评审结论：` + rv.Status + `
人工评审意见：
` + note
	raw, err := s.agent.Complete(ctx, &agent.CompletionRequest{
		Messages: []agent.Message{
			{Role: "system", Content: "你是评审知识工程师，擅长把评审意见提炼为可复用规则和模板更新建议。"},
			{Role: "user", Content: prompt},
		},
		Temperature: 0,
	})
	if err != nil {
		return nil, err
	}
	var parsed struct {
		Summary            string                `json:"summary"`
		Issues             []model.ReviewIssue   `json:"issues"`
		Rules              []model.KnowledgeRule `json:"rules"`
		TemplateSuggestion string                `json:"templateSuggestion"`
	}
	if err := json.Unmarshal([]byte(extractJSONObject(raw)), &parsed); err != nil {
		return nil, err
	}
	item := s.fallbackLearningSuggestion(rv, g, note)
	if strings.TrimSpace(parsed.Summary) != "" {
		item.Summary = parsed.Summary
	}
	if len(parsed.Issues) > 0 {
		item.Issues = parsed.Issues
	}
	if len(parsed.Rules) > 0 {
		item.Rules = parsed.Rules
	}
	item.TemplateSuggestion = parsed.TemplateSuggestion
	s.normalizeLearningSuggestion(item, rv, g)
	return item, nil
}

func (s *ReviewService) fallbackLearningSuggestion(rv *model.Review, g *model.Guide, note string) *model.ReviewLearningSuggestion {
	category := "人工评审意见"
	if rv.Status == "rejected" {
		category = "驳回意见"
	}
	item := &model.ReviewLearningSuggestion{
		ReviewID: rv.ID, GuideID: g.ID, TemplateID: g.TemplateID, Status: "pending", RawNote: note,
		Summary: "AI 已根据评审意见生成待确认的知识沉淀候选。",
		Issues: []model.ReviewIssue{{
			Category: category, TriggerCondition: "评审人提交评审意见", ProblemDesc: note,
			CorrectPractice: "后续 AI 预审需优先检查并提醒：" + note, ChangeType: g.RiskLevel, Frequency: 1,
		}},
		Rules: []model.KnowledgeRule{{
			Title: "来自人工评审：" + shortText(note, 24), RuleType: "rule", Pattern: note,
			Suggestion: "评审材料时检查：" + note, Enabled: true,
		}},
		TemplateSuggestion: "- " + note,
	}
	s.normalizeLearningSuggestion(item, rv, g)
	return item
}

func (s *ReviewService) normalizeLearningSuggestion(item *model.ReviewLearningSuggestion, rv *model.Review, g *model.Guide) {
	item.ReviewID = rv.ID
	item.GuideID = g.ID
	item.TemplateID = g.TemplateID
	item.Status = "pending"
	item.RawNote = strings.TrimSpace(item.RawNote)
	for i := range item.Issues {
		item.Issues[i].SourceReviewID = rv.ID
		if item.Issues[i].ChangeType == "" {
			item.Issues[i].ChangeType = g.RiskLevel
		}
		if item.Issues[i].Frequency <= 0 {
			item.Issues[i].Frequency = 1
		}
	}
	for i := range item.Rules {
		if item.Rules[i].RuleType == "" {
			item.Rules[i].RuleType = "rule"
		}
		item.Rules[i].Enabled = true
	}
}

func extractJSONObject(s string) string {
	start := strings.Index(s, "{")
	end := strings.LastIndex(s, "}")
	if start >= 0 && end > start {
		return s[start : end+1]
	}
	return s
}

func shortText(s string, limit int) string {
	r := []rune(strings.TrimSpace(s))
	if len(r) <= limit {
		return string(r)
	}
	return string(r[:limit]) + "..."
}

func (s *ReviewService) AddComment(c *model.ReviewComment) (*model.ReviewComment, error) {
	c.ID = uuid.NewString()
	if c.Severity == "" {
		c.Severity = "info"
	}
	c.CreatedAt = time.Now().Format(time.RFC3339)
	if err := s.repo.AddComment(c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *ReviewService) ListComments(reviewID string) ([]model.ReviewComment, error) {
	return s.repo.ListComments(reviewID)
}
