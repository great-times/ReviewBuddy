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
	if err := migrateReviewConfig(conn, now); err != nil {
		return err
	}
	return nil
}

func migrateReviewConfig(conn *sql.DB, now string) error {
	_, err := conn.Exec(`
		CREATE TABLE IF NOT EXISTS review_roles (
			id TEXT PRIMARY KEY,
			role_key TEXT NOT NULL UNIQUE,
			name TEXT NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			system INTEGER NOT NULL DEFAULT 0,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);
		CREATE TABLE IF NOT EXISTS review_domains (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			description TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);
		CREATE TABLE IF NOT EXISTS domain_role_users (
			domain_id TEXT NOT NULL,
			role_key TEXT NOT NULL,
			user_id TEXT NOT NULL,
			PRIMARY KEY (domain_id, role_key, user_id)
		);
		CREATE INDEX IF NOT EXISTS idx_domain_role_users_domain ON domain_role_users(domain_id);
		CREATE TABLE IF NOT EXISTS user_domains (
			user_id TEXT NOT NULL,
			domain_id TEXT NOT NULL,
			PRIMARY KEY (user_id, domain_id)
		);
		CREATE INDEX IF NOT EXISTS idx_user_domains_domain ON user_domains(domain_id);
		CREATE TABLE IF NOT EXISTS review_scenarios (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			description TEXT NOT NULL DEFAULT '',
			role_keys TEXT NOT NULL DEFAULT '[]',
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);
	`)
	if err != nil {
		return err
	}
	defaultRoles := []struct {
		key, name, desc string
		system          int
	}{
		{"admin", "管理员", "平台管理与配置维护", 1},
		{"readonly", "只读", "新注册用户默认角色，仅可查看", 1},
		{"developer", "开发", "研发实现与代码评审", 0},
		{"ops", "运维", "发布、运行与稳定性评审", 0},
		{"tester", "测试", "测试策略与质量验证评审", 0},
		{"architect", "架构", "技术方案与架构风险评审", 0},
		{"designer", "设计", "产品体验与交互设计评审", 0},
	}
	for _, r := range defaultRoles {
		if _, err := conn.Exec(`
			INSERT INTO review_roles (id,role_key,name,description,system,created_at,updated_at)
			SELECT ?,?,?,?,?,?,?
			WHERE NOT EXISTS (SELECT 1 FROM review_roles WHERE role_key=?)
		`, r.key, r.key, r.name, r.desc, r.system, now, now, r.key); err != nil {
			return err
		}
	}
	if _, err := conn.Exec(`
		INSERT INTO review_domains (id,name,description,created_at,updated_at)
		SELECT 'default', '默认领域', '评审领域，可按业务线或系统继续拆分。', ?, ?
		WHERE NOT EXISTS (SELECT 1 FROM review_domains WHERE id='default')
	`, now, now); err != nil {
		return err
	}
	_, err = conn.Exec(`
		INSERT INTO review_scenarios (id,name,description,role_keys,created_at,updated_at)
		SELECT 'standard', '标准评审', '默认评审场景，覆盖常见研发协同角色。', '["developer","tester","architect"]', ?, ?
		WHERE NOT EXISTS (SELECT 1 FROM review_scenarios WHERE id='standard')
	`, now, now)
	return err
}
