package reviewconfig

import (
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"

	"changebuddy/internal/model"
	"changebuddy/internal/repo"
)

var roleKeyPattern = regexp.MustCompile(`^[a-z][a-z0-9_]{1,31}$`)

type Service struct {
	repo  *repo.ReviewConfigRepo
	users *repo.UserRepo
}

func NewService(r *repo.ReviewConfigRepo, users *repo.UserRepo) *Service {
	return &Service{repo: r, users: users}
}

func (s *Service) ListRoles() ([]model.ReviewRole, error) { return s.repo.ListRoles() }

func (s *Service) RoleExists(key string) (bool, error) {
	if key == "" {
		return false, nil
	}
	return s.repo.RoleExists(key)
}

func (s *Service) CreateRole(item *model.ReviewRole) (*model.ReviewRole, error) {
	item.Key = strings.TrimSpace(item.Key)
	item.Name = strings.TrimSpace(item.Name)
	if item.Name == "" {
		return nil, errors.New("role name is required")
	}
	if !roleKeyPattern.MatchString(item.Key) {
		return nil, errors.New("role key must use lowercase letters, numbers or underscore")
	}
	now := time.Now().Format(time.RFC3339)
	item.ID = uuid.NewString()
	item.System = false
	item.CreatedAt = now
	item.UpdatedAt = now
	if err := s.repo.CreateRole(item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *Service) UpdateRole(key string, item *model.ReviewRole) (*model.ReviewRole, error) {
	item.Key = key
	item.Name = strings.TrimSpace(item.Name)
	if item.Name == "" {
		return nil, errors.New("role name is required")
	}
	item.UpdatedAt = time.Now().Format(time.RFC3339)
	if err := s.repo.UpdateRole(item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *Service) DeleteRole(key string) error {
	if key == "admin" || key == "readonly" {
		return errors.New("system role cannot be deleted")
	}
	n, err := s.users.CountRole(key)
	if err != nil {
		return err
	}
	if n > 0 {
		return errors.New("role is assigned to users")
	}
	return s.repo.DeleteRole(key)
}

func (s *Service) ListDomains() ([]model.ReviewDomain, error) { return s.repo.ListDomains() }

func (s *Service) SaveDomain(id string, item *model.ReviewDomain) (*model.ReviewDomain, error) {
	item.Name = strings.TrimSpace(item.Name)
	if item.Name == "" {
		return nil, errors.New("domain name is required")
	}
	now := time.Now().Format(time.RFC3339)
	if id == "" {
		item.ID = uuid.NewString()
		item.CreatedAt = now
	} else {
		item.ID = id
	}
	item.UpdatedAt = now
	if item.CreatedAt == "" {
		item.CreatedAt = now
	}
	if err := s.repo.SaveDomain(item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *Service) DeleteDomain(id string) error {
	if id == "default" {
		return errors.New("default domain cannot be deleted")
	}
	return s.repo.DeleteDomain(id)
}

func (s *Service) ListDomainRoleUsers(domainID string) ([]model.DomainRoleUsers, error) {
	return s.repo.ListDomainRoleUsers(domainID)
}

func (s *Service) SaveDomainRoleUsers(item *model.DomainRoleUsers) (*model.DomainRoleUsers, error) {
	if item.DomainID == "" || item.RoleKey == "" {
		return nil, errors.New("domain and role are required")
	}
	return item, s.repo.SaveDomainRoleUsers(item)
}

func (s *Service) ListScenarios() ([]model.ReviewScenario, error) { return s.repo.ListScenarios() }

func (s *Service) SaveScenario(id string, item *model.ReviewScenario) (*model.ReviewScenario, error) {
	item.Name = strings.TrimSpace(item.Name)
	if item.Name == "" {
		return nil, errors.New("scenario name is required")
	}
	now := time.Now().Format(time.RFC3339)
	if id == "" {
		item.ID = uuid.NewString()
		item.CreatedAt = now
	} else {
		item.ID = id
	}
	item.UpdatedAt = now
	if item.CreatedAt == "" {
		item.CreatedAt = now
	}
	if item.RoleKeys == nil {
		item.RoleKeys = []string{}
	}
	if err := s.repo.SaveScenario(item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *Service) DeleteScenario(id string) error {
	if id == "standard" {
		return errors.New("default scenario cannot be deleted")
	}
	return s.repo.DeleteScenario(id)
}
