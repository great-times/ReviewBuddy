package knowledge

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	"reviewbuddy/internal/model"
	"reviewbuddy/internal/repo"
)

type Service struct {
	repo    *repo.KnowledgeRepo
	tplRepo *repo.TemplateRepo
}

func NewService(r *repo.KnowledgeRepo, tpl *repo.TemplateRepo) *Service {
	return &Service{repo: r, tplRepo: tpl}
}

func (s *Service) ListIssues() ([]model.ReviewIssue, error) { return s.repo.ListIssues() }

func (s *Service) AddIssue(i *model.ReviewIssue) (*model.ReviewIssue, error) {
	i.ID = uuid.NewString()
	if i.Frequency <= 0 {
		i.Frequency = 1
	}
	i.CreatedAt = time.Now().Format(time.RFC3339)
	if err := s.repo.AddIssue(i); err != nil {
		return nil, err
	}
	return i, nil
}

func (s *Service) ListRules(onlyEnabled bool) ([]model.KnowledgeRule, error) {
	return s.repo.ListRules(onlyEnabled)
}

func (s *Service) AddRule(k *model.KnowledgeRule) (*model.KnowledgeRule, error) {
	now := time.Now().Format(time.RFC3339)
	k.ID = uuid.NewString()
	k.Enabled = true
	k.CreatedAt = now
	k.UpdatedAt = now
	if err := s.repo.AddRule(k); err != nil {
		return nil, err
	}
	return k, nil
}

func (s *Service) AddLearningSuggestion(item *model.ReviewLearningSuggestion) (*model.ReviewLearningSuggestion, error) {
	item.ID = uuid.NewString()
	if item.Status == "" {
		item.Status = "pending"
	}
	item.CreatedAt = time.Now().Format(time.RFC3339)
	if item.Issues == nil {
		item.Issues = []model.ReviewIssue{}
	}
	if item.Rules == nil {
		item.Rules = []model.KnowledgeRule{}
	}
	if err := s.repo.AddLearningSuggestion(item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *Service) ListLearningSuggestions(status string) ([]model.ReviewLearningSuggestion, error) {
	return s.repo.ListLearningSuggestions(status)
}

func (s *Service) ApplyLearningSuggestion(id string) (*model.ReviewLearningSuggestion, error) {
	item, err := s.repo.GetLearningSuggestion(id)
	if err != nil || item == nil {
		return item, err
	}
	if item.Status == "applied" {
		return item, nil
	}
	now := time.Now().Format(time.RFC3339)
	for i := range item.Issues {
		issue := item.Issues[i]
		issue.SourceReviewID = item.ReviewID
		issue.ID = ""
		issue.CreatedAt = ""
		if _, err := s.AddIssue(&issue); err != nil {
			return nil, err
		}
	}
	for i := range item.Rules {
		rule := item.Rules[i]
		rule.ID = ""
		rule.CreatedAt = ""
		rule.UpdatedAt = ""
		if _, err := s.AddRule(&rule); err != nil {
			return nil, err
		}
	}
	if s.tplRepo != nil && item.TemplateID != "" && strings.TrimSpace(item.TemplateSuggestion) != "" {
		tpl, err := s.tplRepo.Get(item.TemplateID)
		if err != nil {
			return nil, err
		}
		if tpl == nil {
			return nil, errors.New("template not found")
		}
		addition := "\n\n## 评审规则沉淀\n" + strings.TrimSpace(item.TemplateSuggestion) + "\n"
		if !strings.Contains(tpl.Content, strings.TrimSpace(item.TemplateSuggestion)) {
			tpl.Content = strings.TrimRight(tpl.Content, "\n") + addition
			tpl.CurrentVersion++
			tpl.UpdatedAt = now
			if err := s.tplRepo.Update(tpl); err != nil {
				return nil, err
			}
			_ = s.tplRepo.AddVersion(&model.TemplateVersion{
				ID: uuid.NewString(), TemplateID: tpl.ID, Version: tpl.CurrentVersion, Content: tpl.Content,
				ChangeNote: "应用 AI 评审沉淀建议", CreatedBy: "AI", CreatedAt: now,
			})
		}
	}
	item.Status = "applied"
	item.AppliedAt = now
	if err := s.repo.MarkLearningSuggestionApplied(id, now); err != nil {
		return nil, err
	}
	return item, nil
}

// Recall 召回与上下文相关的历史问题与规则，组装成可注入 Prompt 的知识块。
// 当前为关键词降级方案；接入 embedding 后可替换为向量相似度。
func (s *Service) Recall(context string, limit int) string {
	if limit <= 0 {
		limit = 5
	}
	keyword := firstKeyword(context)
	issues, _ := s.repo.SearchIssues(keyword, limit)
	rules, _ := s.repo.ListRules(true)

	var b strings.Builder
	if len(rules) > 0 {
		b.WriteString("## 已沉淀的审查规则\n")
		for i, r := range rules {
			if i >= limit {
				break
			}
			b.WriteString("- [" + r.RuleType + "] " + r.Title + "：" + r.Suggestion + "\n")
		}
	}
	if len(issues) > 0 {
		b.WriteString("\n## 历史评审中出现过的同类问题\n")
		for _, it := range issues {
			b.WriteString("- 问题：" + it.ProblemDesc + "；正确做法：" + it.CorrectPractice + "\n")
		}
	}
	return b.String()
}

func (s *Service) Counts() (issues, rules int, err error) { return s.repo.Counts() }

func firstKeyword(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	fields := strings.Fields(s)
	if len(fields) == 0 {
		return ""
	}
	return fields[0]
}
