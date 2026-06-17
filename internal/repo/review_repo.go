package repo

import (
	"database/sql"

	"reviewbuddy/internal/model"
)

type ReviewRepo struct{ db *sql.DB }

func NewReviewRepo(db *sql.DB) *ReviewRepo { return &ReviewRepo{db: db} }

func (r *ReviewRepo) ListByGuide(guideID string) ([]model.Review, error) {
	rows, err := r.db.Query(`SELECT id,guide_id,guide_version,reviewer,COALESCE(reviewer_user_id,''),status,decision_note,created_at,COALESCE(finished_at,'')
		FROM reviews WHERE guide_id=? ORDER BY created_at DESC`, guideID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.Review{}
	for rows.Next() {
		var v model.Review
		if err := rows.Scan(&v.ID, &v.GuideID, &v.GuideVersion, &v.Reviewer, &v.ReviewerUserID, &v.Status, &v.DecisionNote, &v.CreatedAt, &v.FinishedAt); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func (r *ReviewRepo) Get(id string) (*model.Review, error) {
	row := r.db.QueryRow(`SELECT id,guide_id,guide_version,reviewer,COALESCE(reviewer_user_id,''),status,decision_note,created_at,COALESCE(finished_at,'') FROM reviews WHERE id=?`, id)
	var v model.Review
	err := row.Scan(&v.ID, &v.GuideID, &v.GuideVersion, &v.Reviewer, &v.ReviewerUserID, &v.Status, &v.DecisionNote, &v.CreatedAt, &v.FinishedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *ReviewRepo) Create(v *model.Review) error {
	_, err := r.db.Exec(`INSERT INTO reviews (id,guide_id,guide_version,reviewer,reviewer_user_id,status,decision_note,created_at)
		VALUES (?,?,?,?,?,?,?,?)`, v.ID, v.GuideID, v.GuideVersion, v.Reviewer, v.ReviewerUserID, v.Status, v.DecisionNote, v.CreatedAt)
	return err
}

func (r *ReviewRepo) ListAll() ([]model.Review, error) {
	rows, err := r.db.Query(`SELECT id,guide_id,guide_version,reviewer,COALESCE(reviewer_user_id,''),status,decision_note,created_at,COALESCE(finished_at,'')
		FROM reviews ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.Review{}
	for rows.Next() {
		var v model.Review
		if err := rows.Scan(&v.ID, &v.GuideID, &v.GuideVersion, &v.Reviewer, &v.ReviewerUserID, &v.Status, &v.DecisionNote, &v.CreatedAt, &v.FinishedAt); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func (r *ReviewRepo) Decide(id, status, note, finishedAt string) error {
	_, err := r.db.Exec(`UPDATE reviews SET status=?,decision_note=?,finished_at=? WHERE id=?`, status, note, finishedAt, id)
	return err
}

func (r *ReviewRepo) AddComment(c *model.ReviewComment) error {
	_, err := r.db.Exec(`INSERT INTO review_comments (id,review_id,anchor,severity,category,content,resolved,created_at)
		VALUES (?,?,?,?,?,?,?,?)`, c.ID, c.ReviewID, c.Anchor, c.Severity, c.Category, c.Content, boolToInt(c.Resolved), c.CreatedAt)
	return err
}

func (r *ReviewRepo) ListComments(reviewID string) ([]model.ReviewComment, error) {
	rows, err := r.db.Query(`SELECT id,review_id,anchor,severity,category,content,resolved,created_at
		FROM review_comments WHERE review_id=? ORDER BY created_at`, reviewID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.ReviewComment{}
	for rows.Next() {
		var c model.ReviewComment
		var resolved int
		if err := rows.Scan(&c.ID, &c.ReviewID, &c.Anchor, &c.Severity, &c.Category, &c.Content, &resolved, &c.CreatedAt); err != nil {
			return nil, err
		}
		c.Resolved = resolved != 0
		out = append(out, c)
	}
	return out, rows.Err()
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
