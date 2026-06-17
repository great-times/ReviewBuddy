package repo

import (
	"database/sql"
	"encoding/json"

	"reviewbuddy/internal/model"
)

type KnowledgeRepo struct{ db *sql.DB }

func NewKnowledgeRepo(db *sql.DB) *KnowledgeRepo { return &KnowledgeRepo{db: db} }

func (r *KnowledgeRepo) AddIssue(i *model.ReviewIssue) error {
	_, err := r.db.Exec(`INSERT INTO review_issues
		(id,source_review_id,category,trigger_condition,problem_desc,correct_practice,change_type,frequency,created_at)
		VALUES (?,?,?,?,?,?,?,?,?)`,
		i.ID, i.SourceReviewID, i.Category, i.TriggerCondition, i.ProblemDesc, i.CorrectPractice, i.ChangeType, i.Frequency, i.CreatedAt)
	return err
}

func (r *KnowledgeRepo) ListIssues() ([]model.ReviewIssue, error) {
	rows, err := r.db.Query(`SELECT id,source_review_id,category,trigger_condition,problem_desc,correct_practice,change_type,frequency,created_at
		FROM review_issues ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.ReviewIssue{}
	for rows.Next() {
		var i model.ReviewIssue
		if err := rows.Scan(&i.ID, &i.SourceReviewID, &i.Category, &i.TriggerCondition, &i.ProblemDesc, &i.CorrectPractice, &i.ChangeType, &i.Frequency, &i.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, i)
	}
	return out, rows.Err()
}

// SearchIssues 简易关键词召回（embedding 缺省时的降级方案）
func (r *KnowledgeRepo) SearchIssues(keyword string, limit int) ([]model.ReviewIssue, error) {
	like := "%" + keyword + "%"
	rows, err := r.db.Query(`SELECT id,source_review_id,category,trigger_condition,problem_desc,correct_practice,change_type,frequency,created_at
		FROM review_issues
		WHERE problem_desc LIKE ? OR category LIKE ? OR change_type LIKE ? OR trigger_condition LIKE ?
		ORDER BY frequency DESC, created_at DESC LIMIT ?`, like, like, like, like, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.ReviewIssue{}
	for rows.Next() {
		var i model.ReviewIssue
		if err := rows.Scan(&i.ID, &i.SourceReviewID, &i.Category, &i.TriggerCondition, &i.ProblemDesc, &i.CorrectPractice, &i.ChangeType, &i.Frequency, &i.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, i)
	}
	return out, rows.Err()
}

func (r *KnowledgeRepo) ListRules(onlyEnabled bool) ([]model.KnowledgeRule, error) {
	q := `SELECT id,title,rule_type,pattern,suggestion,enabled,hit_count,created_at,updated_at FROM knowledge_rules`
	if onlyEnabled {
		q += " WHERE enabled = 1"
	}
	q += " ORDER BY hit_count DESC, updated_at DESC"
	rows, err := r.db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.KnowledgeRule{}
	for rows.Next() {
		var k model.KnowledgeRule
		var enabled int
		if err := rows.Scan(&k.ID, &k.Title, &k.RuleType, &k.Pattern, &k.Suggestion, &enabled, &k.HitCount, &k.CreatedAt, &k.UpdatedAt); err != nil {
			return nil, err
		}
		k.Enabled = enabled != 0
		out = append(out, k)
	}
	return out, rows.Err()
}

func (r *KnowledgeRepo) AddRule(k *model.KnowledgeRule) error {
	_, err := r.db.Exec(`INSERT INTO knowledge_rules (id,title,rule_type,pattern,suggestion,derived_from,enabled,hit_count,created_at,updated_at)
		VALUES (?,?,?,?,?,?,?,?,?,?)`, k.ID, k.Title, k.RuleType, k.Pattern, k.Suggestion, "[]", boolToInt(k.Enabled), k.HitCount, k.CreatedAt, k.UpdatedAt)
	return err
}

func (r *KnowledgeRepo) AddLearningSuggestion(s *model.ReviewLearningSuggestion) error {
	issues, _ := json.Marshal(s.Issues)
	rules, _ := json.Marshal(s.Rules)
	_, err := r.db.Exec(`INSERT INTO review_learning_suggestions
		(id,review_id,guide_id,template_id,status,raw_note,summary,issues,rules,template_suggestion,created_at,applied_at)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?)`,
		s.ID, s.ReviewID, s.GuideID, s.TemplateID, s.Status, s.RawNote, s.Summary, string(issues), string(rules), s.TemplateSuggestion, s.CreatedAt, nullableString(s.AppliedAt))
	return err
}

func (r *KnowledgeRepo) ListLearningSuggestions(status string) ([]model.ReviewLearningSuggestion, error) {
	q := `SELECT id,review_id,guide_id,template_id,status,raw_note,summary,issues,rules,template_suggestion,created_at,COALESCE(applied_at,'')
		FROM review_learning_suggestions`
	args := []any{}
	if status != "" {
		q += " WHERE status=?"
		args = append(args, status)
	}
	q += " ORDER BY created_at DESC"
	rows, err := r.db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.ReviewLearningSuggestion{}
	for rows.Next() {
		item, err := scanLearningSuggestion(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *KnowledgeRepo) GetLearningSuggestion(id string) (*model.ReviewLearningSuggestion, error) {
	row := r.db.QueryRow(`SELECT id,review_id,guide_id,template_id,status,raw_note,summary,issues,rules,template_suggestion,created_at,COALESCE(applied_at,'')
		FROM review_learning_suggestions WHERE id=?`, id)
	item, err := scanLearningSuggestion(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *KnowledgeRepo) MarkLearningSuggestionApplied(id, appliedAt string) error {
	_, err := r.db.Exec(`UPDATE review_learning_suggestions SET status='applied', applied_at=? WHERE id=?`, appliedAt, id)
	return err
}

// Stats 用于度量看板
func (r *KnowledgeRepo) Counts() (issues, rules int, err error) {
	_ = r.db.QueryRow(`SELECT COUNT(*) FROM review_issues`).Scan(&issues)
	err = r.db.QueryRow(`SELECT COUNT(*) FROM knowledge_rules`).Scan(&rules)
	return issues, rules, err
}

var _ = sql.ErrNoRows

func scanLearningSuggestion(s scanner) (model.ReviewLearningSuggestion, error) {
	var item model.ReviewLearningSuggestion
	var issues, rules string
	err := s.Scan(&item.ID, &item.ReviewID, &item.GuideID, &item.TemplateID, &item.Status, &item.RawNote, &item.Summary, &issues, &rules, &item.TemplateSuggestion, &item.CreatedAt, &item.AppliedAt)
	if err != nil {
		return item, err
	}
	_ = json.Unmarshal([]byte(issues), &item.Issues)
	_ = json.Unmarshal([]byte(rules), &item.Rules)
	if item.Issues == nil {
		item.Issues = []model.ReviewIssue{}
	}
	if item.Rules == nil {
		item.Rules = []model.KnowledgeRule{}
	}
	return item, nil
}

func nullableString(s string) any {
	if s == "" {
		return nil
	}
	return s
}
