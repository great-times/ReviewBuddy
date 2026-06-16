package repo

import (
	"database/sql"
	"encoding/json"
	"time"
)

type SettingsRepo struct{ db *sql.DB }

func NewSettingsRepo(db *sql.DB) *SettingsRepo { return &SettingsRepo{db: db} }

func (r *SettingsRepo) Get(key string, out any) (bool, error) {
	var raw string
	err := r.db.QueryRow(`SELECT value FROM settings WHERE key=?`, key).Scan(&raw)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, json.Unmarshal([]byte(raw), out)
}

func (r *SettingsRepo) Put(key string, value any) error {
	raw, err := json.Marshal(value)
	if err != nil {
		return err
	}
	_, err = r.db.Exec(`
		INSERT INTO settings (key,value,updated_at) VALUES (?,?,?)
		ON CONFLICT(key) DO UPDATE SET value=excluded.value, updated_at=excluded.updated_at
	`, key, string(raw), time.Now().Format(time.RFC3339))
	return err
}
