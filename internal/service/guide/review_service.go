package guide

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	"reviewbuddy/internal/model"
	"reviewbuddy/internal/repo"
	"reviewbuddy/internal/service/knowledge"
)

type ReviewService struct {
	repo      *repo.ReviewRepo
	guideRepo *repo.GuideRepo
	tplRepo   *repo.TemplateRepo
	userRepo  *repo.UserRepo
	knowledge *knowledge.Service
}

func NewReviewService(r *repo.ReviewRepo, g *repo.GuideRepo, tpl *repo.TemplateRepo, users *repo.UserRepo, kn *knowledge.Service) *ReviewService {
	return &ReviewService{repo: r, guideRepo: g, tplRepo: tpl, userRepo: users, knowledge: kn}
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
		s.learnFromDecision(rv, g, note, now)
	}
	return rv, nil
}

func (s *ReviewService) learnFromDecision(rv *model.Review, g *model.Guide, note, now string) {
	note = strings.TrimSpace(note)
	if note == "" || s.knowledge == nil {
		return
	}
	category := "人工评审意见"
	if rv.Status == "rejected" {
		category = "驳回意见"
	}
	_, _ = s.knowledge.AddIssue(&model.ReviewIssue{
		SourceReviewID:   rv.ID,
		Category:         category,
		TriggerCondition: "评审人提交评审意见",
		ProblemDesc:      note,
		CorrectPractice:  "后续 AI 预审需优先检查并提醒：" + note,
		ChangeType:       g.RiskLevel,
		Frequency:        1,
	})
	_, _ = s.knowledge.AddRule(&model.KnowledgeRule{
		Title:      "来自人工评审：" + shortText(note, 24),
		RuleType:   "rule",
		Pattern:    note,
		Suggestion: "评审评审材料时检查：" + note,
		Enabled:    true,
	})
	if s.tplRepo == nil || g.TemplateID == "" {
		return
	}
	tpl, err := s.tplRepo.Get(g.TemplateID)
	if err != nil || tpl == nil || strings.Contains(tpl.Content, note) {
		return
	}
	tpl.Content = strings.TrimRight(tpl.Content, "\n") + "\n\n## AI 评审规则沉淀\n- " + note + "\n"
	tpl.CurrentVersion++
	tpl.UpdatedAt = now
	_ = s.tplRepo.Update(tpl)
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
