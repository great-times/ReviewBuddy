package user

import (
	"errors"
	"time"

	"github.com/google/uuid"

	"changebuddy/internal/model"
	"changebuddy/internal/repo"
)

var roles = map[string]bool{
	"admin":     true,
	"readonly":  true,
	"developer": true,
	"ops":       true,
	"tester":    true,
	"architect": true,
	"designer":  true,
}

type Service struct{ repo *repo.UserRepo }

func NewService(r *repo.UserRepo) *Service { return &Service{repo: r} }

func (s *Service) List() ([]model.User, error) { return s.repo.List() }

func (s *Service) Create(u *model.User) (*model.User, error) {
	if u.Username == "" {
		return nil, errors.New("username is required")
	}
	if u.Role == "" {
		u.Role = "readonly"
	}
	if !roles[u.Role] {
		return nil, errors.New("invalid role")
	}
	now := time.Now().Format(time.RFC3339)
	u.ID = uuid.NewString()
	u.Enabled = true
	u.CreatedAt = now
	u.UpdatedAt = now
	if err := s.repo.Create(u); err != nil {
		return nil, err
	}
	return u, nil
}

func (s *Service) Update(id string, u *model.User) (*model.User, error) {
	if u.Username == "" {
		return nil, errors.New("username is required")
	}
	if !roles[u.Role] {
		return nil, errors.New("invalid role")
	}
	u.ID = id
	u.UpdatedAt = time.Now().Format(time.RFC3339)
	if err := s.repo.Update(u); err != nil {
		return nil, err
	}
	return u, nil
}

func (s *Service) Delete(id string) error { return s.repo.Delete(id) }

func (s *Service) SeedDefaults() {
	existing, err := s.repo.List()
	if err != nil || len(existing) > 0 {
		return
	}
	defaults := []model.User{
		{Username: "平台管理员", Role: "admin"},
		{Username: "只读观察员", Role: "readonly"},
		{Username: "开发评审人", Role: "developer"},
		{Username: "运维评审人", Role: "ops"},
		{Username: "测试评审人", Role: "tester"},
		{Username: "架构评审人", Role: "architect"},
		{Username: "设计评审人", Role: "designer"},
	}
	for i := range defaults {
		_, _ = s.Create(&defaults[i])
	}
}
