package template

import (
	"errors"
	"time"

	"github.com/google/uuid"

	"reviewbuddy/internal/model"
	"reviewbuddy/internal/repo"
)

type Service struct{ repo *repo.TemplateRepo }

func NewService(r *repo.TemplateRepo) *Service { return &Service{repo: r} }

func (s *Service) List(libraryID, category string) ([]model.Template, error) {
	return s.repo.List(libraryID, category)
}

func (s *Service) Get(id string) (*model.Template, error) { return s.repo.Get(id) }

func (s *Service) ListLibraries() ([]model.TemplateLibrary, error) {
	return s.repo.ListLibraries()
}

func (s *Service) CreateLibrary(l *model.TemplateLibrary) (*model.TemplateLibrary, error) {
	if l.Name == "" {
		return nil, errors.New("name is required")
	}
	now := time.Now().Format(time.RFC3339)
	l.ID = uuid.NewString()
	l.CreatedAt = now
	l.UpdatedAt = now
	if err := s.repo.CreateLibrary(l); err != nil {
		return nil, err
	}
	return l, nil
}

func (s *Service) Create(t *model.Template) (*model.Template, error) {
	if t.Name == "" {
		return nil, errors.New("name is required")
	}
	if t.LibraryID == "" {
		t.LibraryID = "default"
	}
	now := time.Now().Format(time.RFC3339)
	t.ID = uuid.NewString()
	t.CurrentVersion = 1
	if t.Status == "" {
		t.Status = "active"
	}
	t.CreatedAt = now
	t.UpdatedAt = now
	if err := s.repo.Create(t); err != nil {
		return nil, err
	}
	_ = s.repo.AddVersion(&model.TemplateVersion{
		ID: uuid.NewString(), TemplateID: t.ID, Version: 1, Content: t.Content,
		ChangeNote: "初始版本", CreatedBy: t.CreatedBy, CreatedAt: now,
	})
	return t, nil
}

func (s *Service) Update(id string, in *model.Template) (*model.Template, error) {
	cur, err := s.repo.Get(id)
	if err != nil {
		return nil, err
	}
	if cur == nil {
		return nil, nil
	}
	now := time.Now().Format(time.RFC3339)
	contentChanged := in.Content != cur.Content
	cur.Name = in.Name
	if in.LibraryID != "" {
		cur.LibraryID = in.LibraryID
	}
	cur.Category = in.Category
	cur.Description = in.Description
	cur.Variables = in.Variables
	cur.Content = in.Content
	if in.Status != "" {
		cur.Status = in.Status
	}
	cur.UpdatedAt = now
	if contentChanged {
		cur.CurrentVersion++
		_ = s.repo.AddVersion(&model.TemplateVersion{
			ID: uuid.NewString(), TemplateID: cur.ID, Version: cur.CurrentVersion, Content: cur.Content,
			ChangeNote: in.Description, CreatedBy: cur.CreatedBy, CreatedAt: now,
		})
	}
	if err := s.repo.Update(cur); err != nil {
		return nil, err
	}
	return cur, nil
}

func (s *Service) ListVersions(id string) ([]model.TemplateVersion, error) {
	return s.repo.ListVersions(id)
}
