package repo

import (
	"database/sql"
	"encoding/json"

	"reviewbuddy/internal/model"
)

type ReviewCollectionRepo struct{ db *sql.DB }

func NewReviewCollectionRepo(db *sql.DB) *ReviewCollectionRepo { return &ReviewCollectionRepo{db: db} }

func (r *ReviewCollectionRepo) List() ([]model.ReviewCollection, error) {
	rows, err := r.db.Query(`SELECT id,title,domain_id,guide_ids,status,decision_note,created_by,created_at,updated_at FROM review_collections ORDER BY updated_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.ReviewCollection{}
	for rows.Next() {
		item, err := scanReviewCollection(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *ReviewCollectionRepo) Get(id string) (*model.ReviewCollection, error) {
	row := r.db.QueryRow(`SELECT id,title,domain_id,guide_ids,status,decision_note,created_by,created_at,updated_at FROM review_collections WHERE id=?`, id)
	item, err := scanReviewCollection(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *ReviewCollectionRepo) Create(item *model.ReviewCollection) error {
	guideIDs, _ := json.Marshal(item.GuideIDs)
	_, err := r.db.Exec(`INSERT INTO review_collections (id,title,domain_id,guide_ids,status,decision_note,created_by,created_at,updated_at)
		VALUES (?,?,?,?,?,?,?,?,?)`, item.ID, item.Title, item.DomainID, string(guideIDs), item.Status, item.DecisionNote, item.CreatedBy, item.CreatedAt, item.UpdatedAt)
	return err
}

func (r *ReviewCollectionRepo) Update(item *model.ReviewCollection) error {
	guideIDs, _ := json.Marshal(item.GuideIDs)
	_, err := r.db.Exec(`UPDATE review_collections SET title=?,domain_id=?,guide_ids=?,status=?,decision_note=?,updated_at=? WHERE id=?`,
		item.Title, item.DomainID, string(guideIDs), item.Status, item.DecisionNote, item.UpdatedAt, item.ID)
	return err
}

func scanReviewCollection(s scanner) (model.ReviewCollection, error) {
	var item model.ReviewCollection
	var guideIDs string
	err := s.Scan(&item.ID, &item.Title, &item.DomainID, &guideIDs, &item.Status, &item.DecisionNote, &item.CreatedBy, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		return item, err
	}
	_ = json.Unmarshal([]byte(guideIDs), &item.GuideIDs)
	if item.GuideIDs == nil {
		item.GuideIDs = []string{}
	}
	return item, nil
}
