package repo

import (
	"database/sql"
	"encoding/json"

	"reviewbuddy/internal/model"
)

type GuideRepo struct{ db *sql.DB }

func NewGuideRepo(db *sql.DB) *GuideRepo { return &GuideRepo{db: db} }

func (r *GuideRepo) List(status string) ([]model.Guide, error) {
	q := `SELECT id,title,template_id,content,variables,status,risk_level,current_version,created_by,created_at,updated_at FROM guides`
	args := []any{}
	if status != "" {
		q += " WHERE status = ?"
		args = append(args, status)
	}
	q += " ORDER BY updated_at DESC"
	rows, err := r.db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.Guide{}
	for rows.Next() {
		g, err := scanGuide(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, g)
	}
	return out, rows.Err()
}

func (r *GuideRepo) Get(id string) (*model.Guide, error) {
	row := r.db.QueryRow(`SELECT id,title,template_id,content,variables,status,risk_level,current_version,created_by,created_at,updated_at FROM guides WHERE id=?`, id)
	g, err := scanGuide(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &g, nil
}

func (r *GuideRepo) Create(g *model.Guide) error {
	vars, _ := json.Marshal(g.Variables)
	_, err := r.db.Exec(`INSERT INTO guides
		(id,title,template_id,content,variables,status,risk_level,current_version,created_by,created_at,updated_at)
		VALUES (?,?,?,?,?,?,?,?,?,?,?)`,
		g.ID, g.Title, g.TemplateID, g.Content, string(vars), g.Status, g.RiskLevel,
		g.CurrentVersion, g.CreatedBy, g.CreatedAt, g.UpdatedAt)
	return err
}

func (r *GuideRepo) Update(g *model.Guide) error {
	vars, _ := json.Marshal(g.Variables)
	_, err := r.db.Exec(`UPDATE guides SET title=?,content=?,variables=?,status=?,risk_level=?,current_version=?,updated_at=? WHERE id=?`,
		g.Title, g.Content, string(vars), g.Status, g.RiskLevel, g.CurrentVersion, g.UpdatedAt, g.ID)
	return err
}

func scanGuide(s scanner) (model.Guide, error) {
	var g model.Guide
	var vars string
	err := s.Scan(&g.ID, &g.Title, &g.TemplateID, &g.Content, &vars, &g.Status, &g.RiskLevel,
		&g.CurrentVersion, &g.CreatedBy, &g.CreatedAt, &g.UpdatedAt)
	if err != nil {
		return g, err
	}
	if vars != "" {
		_ = json.Unmarshal([]byte(vars), &g.Variables)
	}
	if g.Variables == nil {
		g.Variables = map[string]string{}
	}
	return g, nil
}
