package knowledge

import (
	"strings"
	"time"

	"github.com/google/uuid"

	"reviewbuddy/internal/model"
	"reviewbuddy/internal/repo"
)

type Service struct{ repo *repo.KnowledgeRepo }

func NewService(r *repo.KnowledgeRepo) *Service { return &Service{repo: r} }

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
