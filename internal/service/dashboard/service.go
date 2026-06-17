package dashboard

import (
	"reviewbuddy/internal/model"
	"reviewbuddy/internal/repo"
)

type Service struct {
	templates    *repo.TemplateRepo
	guides       *repo.GuideRepo
	reviews      *repo.ReviewRepo
	knowledge    *repo.KnowledgeRepo
	reviewConfig *repo.ReviewConfigRepo
}

func NewService(t *repo.TemplateRepo, g *repo.GuideRepo, r *repo.ReviewRepo, k *repo.KnowledgeRepo, rc *repo.ReviewConfigRepo) *Service {
	return &Service{templates: t, guides: g, reviews: r, knowledge: k, reviewConfig: rc}
}

func (s *Service) Summary(userID string) (*model.DashboardSummary, error) {
	templates, err := s.templates.List("", "")
	if err != nil {
		return nil, err
	}
	guides, err := s.guides.List("")
	if err != nil {
		return nil, err
	}
	reviews, err := s.reviews.ListAll()
	if err != nil {
		return nil, err
	}
	domains, err := s.reviewConfig.ListDomains()
	if err != nil {
		return nil, err
	}
	myDomains, err := s.reviewConfig.ListUserDomains(userID)
	if err != nil {
		return nil, err
	}
	issues, err := s.knowledge.ListIssues()
	if err != nil {
		return nil, err
	}
	rules, err := s.knowledge.ListRules(false)
	if err != nil {
		return nil, err
	}
	return &model.DashboardSummary{
		Templates:   templates,
		Guides:      guides,
		Reviews:     reviews,
		Domains:     domains,
		MyDomainIDs: myDomains.DomainIDs,
		IssueCount:  len(issues),
		RuleCount:   len(rules),
	}, nil
}
