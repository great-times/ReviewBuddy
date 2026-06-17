package repo

import (
	"database/sql"
	"encoding/json"
	"strings"

	"reviewbuddy/internal/model"
)

type TemplateRepo struct{ db *sql.DB }

func NewTemplateRepo(db *sql.DB) *TemplateRepo { return &TemplateRepo{db: db} }

func (r *TemplateRepo) List(libraryID, category string) ([]model.Template, error) {
	q := `SELECT id,library_id,name,category,description,content,variables,quality_score,usage_count,
		current_version,status,created_by,created_at,updated_at FROM templates`
	args := []any{}
	conds := []string{}
	if libraryID != "" {
		conds = append(conds, "library_id = ?")
		args = append(args, libraryID)
	}
	if category != "" {
		conds = append(conds, "category = ?")
		args = append(args, category)
	}
	if len(conds) > 0 {
		q += " WHERE " + strings.Join(conds, " AND ")
	}
	q += " ORDER BY updated_at DESC"
	rows, err := r.db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.Template{}
	for rows.Next() {
		t, err := scanTemplate(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (r *TemplateRepo) Get(id string) (*model.Template, error) {
	row := r.db.QueryRow(`SELECT id,library_id,name,category,description,content,variables,quality_score,usage_count,
		current_version,status,created_by,created_at,updated_at FROM templates WHERE id = ?`, id)
	t, err := scanTemplate(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *TemplateRepo) Create(t *model.Template) error {
	vars, _ := json.Marshal(t.Variables)
	_, err := r.db.Exec(`INSERT INTO templates
		(id,library_id,name,category,description,content,variables,quality_score,usage_count,current_version,status,created_by,created_at,updated_at)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		t.ID, t.LibraryID, t.Name, t.Category, t.Description, t.Content, string(vars), t.QualityScore, t.UsageCount,
		t.CurrentVersion, t.Status, t.CreatedBy, t.CreatedAt, t.UpdatedAt)
	return err
}

func (r *TemplateRepo) Update(t *model.Template) error {
	vars, _ := json.Marshal(t.Variables)
	_, err := r.db.Exec(`UPDATE templates SET library_id=?,name=?,category=?,description=?,content=?,variables=?,
		quality_score=?,current_version=?,status=?,updated_at=? WHERE id=?`,
		t.LibraryID, t.Name, t.Category, t.Description, t.Content, string(vars), t.QualityScore,
		t.CurrentVersion, t.Status, t.UpdatedAt, t.ID)
	return err
}

func (r *TemplateRepo) IncrUsage(id string) error {
	_, err := r.db.Exec(`UPDATE templates SET usage_count = usage_count + 1 WHERE id = ?`, id)
	return err
}

func (r *TemplateRepo) AddVersion(v *model.TemplateVersion) error {
	_, err := r.db.Exec(`INSERT INTO template_versions (id,template_id,version,content,change_note,created_by,created_at)
		VALUES (?,?,?,?,?,?,?)`, v.ID, v.TemplateID, v.Version, v.Content, v.ChangeNote, v.CreatedBy, v.CreatedAt)
	return err
}

func (r *TemplateRepo) ListVersions(templateID string) ([]model.TemplateVersion, error) {
	rows, err := r.db.Query(`SELECT id,template_id,version,content,change_note,created_by,created_at
		FROM template_versions WHERE template_id=? ORDER BY version DESC`, templateID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.TemplateVersion{}
	for rows.Next() {
		var v model.TemplateVersion
		if err := rows.Scan(&v.ID, &v.TemplateID, &v.Version, &v.Content, &v.ChangeNote, &v.CreatedBy, &v.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func (r *TemplateRepo) ListLibraries() ([]model.TemplateLibrary, error) {
	rows, err := r.db.Query(`SELECT id,name,description,created_at,updated_at FROM template_libraries ORDER BY created_at ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.TemplateLibrary{}
	for rows.Next() {
		var l model.TemplateLibrary
		if err := rows.Scan(&l.ID, &l.Name, &l.Description, &l.CreatedAt, &l.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

func (r *TemplateRepo) CreateLibrary(l *model.TemplateLibrary) error {
	_, err := r.db.Exec(`INSERT INTO template_libraries (id,name,description,created_at,updated_at) VALUES (?,?,?,?,?)`,
		l.ID, l.Name, l.Description, l.CreatedAt, l.UpdatedAt)
	return err
}

type scanner interface {
	Scan(dest ...any) error
}

func scanTemplate(s scanner) (model.Template, error) {
	var t model.Template
	var vars string
	err := s.Scan(&t.ID, &t.LibraryID, &t.Name, &t.Category, &t.Description, &t.Content, &vars, &t.QualityScore,
		&t.UsageCount, &t.CurrentVersion, &t.Status, &t.CreatedBy, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return t, err
	}
	if vars != "" {
		_ = json.Unmarshal([]byte(vars), &t.Variables)
	}
	if t.Variables == nil {
		t.Variables = []string{}
	}
	return t, nil
}
