package db

import (
	"database/sql"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schemaSQL string

// Open 打开 SQLite 并执行幂等建表
func Open(path string) (*sql.DB, error) {
	if dir := filepath.Dir(path); dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("create db dir: %w", err)
		}
	}

	dsn := fmt.Sprintf("file:%s?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=foreign_keys(on)", path)
	conn, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	if err := conn.Ping(); err != nil {
		return nil, err
	}
	if _, err := conn.Exec(schemaSQL); err != nil {
		return nil, fmt.Errorf("init schema: %w", err)
	}
	if err := migrate(conn); err != nil {
		return nil, fmt.Errorf("migrate schema: %w", err)
	}
	return conn, nil
}

func migrate(conn *sql.DB) error {
	_, err := conn.Exec(`ALTER TABLE users ADD COLUMN password_hash TEXT NOT NULL DEFAULT ''`)
	if err != nil && !strings.Contains(strings.ToLower(err.Error()), "duplicate column") {
		return err
	}
	_, err = conn.Exec(`ALTER TABLE templates ADD COLUMN library_id TEXT NOT NULL DEFAULT ''`)
	if err != nil && !strings.Contains(strings.ToLower(err.Error()), "duplicate column") {
		return err
	}
	now := time.Now().Format(time.RFC3339)
	_, err = conn.Exec(`
		INSERT INTO template_libraries (id,name,description,created_at,updated_at)
		SELECT 'default', '标准评审模板库', '内置模板库，用于承载默认评审模板。', ?, ?
		WHERE NOT EXISTS (SELECT 1 FROM template_libraries WHERE id='default')
	`, now, now)
	if err != nil {
		return err
	}
	_, err = conn.Exec(`UPDATE templates SET library_id='default' WHERE library_id=''`)
	if err != nil {
		return err
	}
	_, err = conn.Exec(`CREATE INDEX IF NOT EXISTS idx_templates_library_id ON templates(library_id)`)
	if err != nil {
		return err
	}
	return nil
}
