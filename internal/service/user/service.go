package user

import (
	"errors"
	"time"

	"github.com/google/uuid"

	"reviewbuddy/internal/model"
	"reviewbuddy/internal/repo"
)

type RoleChecker interface {
	RoleExists(key string) (bool, error)
}

type Service struct {
	repo  *repo.UserRepo
	roles RoleChecker
}

func NewService(r *repo.UserRepo, roles RoleChecker) *Service { return &Service{repo: r, roles: roles} }

func (s *Service) List() ([]model.User, error) { return s.repo.List() }

func (s *Service) Create(u *model.User) (*model.User, error) {
	if u.Username == "" {
		return nil, errors.New("username is required")
	}
	if len(u.Roles) == 0 && u.Role == "" {
		u.Roles = []string{"readonly"}
	}
	if err := s.validateRoles(u); err != nil {
		return nil, err
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
	if err := s.validateRoles(u); err != nil {
		return nil, err
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
		{Username: "平台管理员", Roles: []string{"admin"}},
		{Username: "只读观察员", Roles: []string{"readonly"}},
		{Username: "开发评审人", Roles: []string{"developer"}},
		{Username: "运维评审人", Roles: []string{"ops"}},
		{Username: "测试评审人", Roles: []string{"tester"}},
		{Username: "架构评审人", Roles: []string{"architect"}},
		{Username: "设计评审人", Roles: []string{"designer"}},
	}
	for i := range defaults {
		_, _ = s.Create(&defaults[i])
	}
}

func (s *Service) validateRoles(u *model.User) error {
	roles := u.Roles
	if len(roles) == 0 && u.Role != "" {
		roles = []string{u.Role}
	}
	if len(roles) == 0 {
		return errors.New("role is required")
	}
	seen := map[string]bool{}
	normalized := []string{}
	for _, role := range roles {
		if role == "" || seen[role] {
			continue
		}
		if ok, err := s.roleExists(role); err != nil {
			return err
		} else if !ok {
			return errors.New("invalid role")
		}
		seen[role] = true
		normalized = append(normalized, role)
	}
	if len(normalized) == 0 {
		return errors.New("role is required")
	}
	u.Roles = normalized
	u.Role = normalized[0]
	for _, role := range normalized {
		if role == "admin" {
			u.Role = "admin"
			return nil
		}
	}
	for _, role := range normalized {
		if role != "readonly" {
			u.Role = role
			return nil
		}
	}
	return nil
}

func (s *Service) roleExists(role string) (bool, error) {
	if s.roles == nil {
		return role == "admin" || role == "readonly", nil
	}
	return s.roles.RoleExists(role)
}
