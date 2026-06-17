-- ReviewBuddy 初始 schema (SQLite)
-- 所有表使用 IF NOT EXISTS，启动时幂等执行。

CREATE TABLE IF NOT EXISTS users (
    id         TEXT PRIMARY KEY,
    username   TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL DEFAULT '',
    role       TEXT NOT NULL DEFAULT 'developer',
    enabled    INTEGER NOT NULL DEFAULT 1,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS review_roles (
    id          TEXT PRIMARY KEY,
    role_key    TEXT NOT NULL UNIQUE,
    name        TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    system      INTEGER NOT NULL DEFAULT 0,
    created_at  TEXT NOT NULL,
    updated_at  TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS review_domains (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL UNIQUE,
    description TEXT NOT NULL DEFAULT '',
    created_at  TEXT NOT NULL,
    updated_at  TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS domain_role_users (
    domain_id TEXT NOT NULL,
    role_key  TEXT NOT NULL,
    user_id   TEXT NOT NULL,
    PRIMARY KEY (domain_id, role_key, user_id)
);
CREATE INDEX IF NOT EXISTS idx_domain_role_users_domain ON domain_role_users(domain_id);

CREATE TABLE IF NOT EXISTS user_domains (
    user_id   TEXT NOT NULL,
    domain_id TEXT NOT NULL,
    PRIMARY KEY (user_id, domain_id)
);
CREATE INDEX IF NOT EXISTS idx_user_domains_domain ON user_domains(domain_id);

CREATE TABLE IF NOT EXISTS review_scenarios (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL UNIQUE,
    description TEXT NOT NULL DEFAULT '',
    role_keys   TEXT NOT NULL DEFAULT '[]',
    created_at  TEXT NOT NULL,
    updated_at  TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS auth_tokens (
    token      TEXT PRIMARY KEY,
    user_id    TEXT NOT NULL,
    expires_at TEXT NOT NULL,
    created_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_auth_tokens_user_id ON auth_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_auth_tokens_expires_at ON auth_tokens(expires_at);

CREATE TABLE IF NOT EXISTS template_libraries (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    created_at  TEXT NOT NULL,
    updated_at  TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS templates (
    id              TEXT PRIMARY KEY,
    library_id      TEXT NOT NULL DEFAULT '',
    name            TEXT NOT NULL,
    category        TEXT NOT NULL DEFAULT '',
    description     TEXT NOT NULL DEFAULT '',
    content         TEXT NOT NULL DEFAULT '',
    variables       TEXT NOT NULL DEFAULT '[]',  -- JSON 数组
    quality_score   REAL NOT NULL DEFAULT 0,
    usage_count     INTEGER NOT NULL DEFAULT 0,
    current_version INTEGER NOT NULL DEFAULT 1,
    status          TEXT NOT NULL DEFAULT 'active',
    created_by      TEXT NOT NULL DEFAULT '',
    created_at      TEXT NOT NULL,
    updated_at      TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS template_versions (
    id          TEXT PRIMARY KEY,
    template_id TEXT NOT NULL,
    version     INTEGER NOT NULL,
    content     TEXT NOT NULL DEFAULT '',
    change_note TEXT NOT NULL DEFAULT '',
    created_by  TEXT NOT NULL DEFAULT '',
    created_at  TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_template_versions_tid ON template_versions(template_id);

CREATE TABLE IF NOT EXISTS guides (
    id              TEXT PRIMARY KEY,
    title           TEXT NOT NULL,
    template_id     TEXT NOT NULL DEFAULT '',
    content         TEXT NOT NULL DEFAULT '',
    variables       TEXT NOT NULL DEFAULT '{}',  -- JSON 对象
    status          TEXT NOT NULL DEFAULT 'draft', -- draft/reviewing/approved/archived
    risk_level      TEXT NOT NULL DEFAULT 'medium',
    current_version INTEGER NOT NULL DEFAULT 1,
    created_by      TEXT NOT NULL DEFAULT '',
    created_at      TEXT NOT NULL,
    updated_at      TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_guides_status ON guides(status);

CREATE TABLE IF NOT EXISTS guide_versions (
    id         TEXT PRIMARY KEY,
    guide_id   TEXT NOT NULL,
    version    INTEGER NOT NULL,
    content    TEXT NOT NULL DEFAULT '',
    created_by TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_guide_versions_gid ON guide_versions(guide_id);

CREATE TABLE IF NOT EXISTS reviews (
    id            TEXT PRIMARY KEY,
    guide_id      TEXT NOT NULL,
    guide_version INTEGER NOT NULL DEFAULT 1,
    reviewer      TEXT NOT NULL DEFAULT '',
    reviewer_user_id TEXT NOT NULL DEFAULT '',
    status        TEXT NOT NULL DEFAULT 'pending', -- pending/approved/rejected
    decision_note TEXT NOT NULL DEFAULT '',
    created_at    TEXT NOT NULL,
    finished_at   TEXT
);
CREATE INDEX IF NOT EXISTS idx_reviews_gid ON reviews(guide_id);

CREATE TABLE IF NOT EXISTS review_comments (
    id         TEXT PRIMARY KEY,
    review_id  TEXT NOT NULL,
    anchor     TEXT NOT NULL DEFAULT '',
    severity   TEXT NOT NULL DEFAULT 'info', -- info/warning/critical
    category   TEXT NOT NULL DEFAULT '',
    content    TEXT NOT NULL DEFAULT '',
    resolved   INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_review_comments_rid ON review_comments(review_id);

-- 自学习原料：评审/复盘中确认的问题
CREATE TABLE IF NOT EXISTS review_issues (
    id                TEXT PRIMARY KEY,
    source_review_id  TEXT NOT NULL DEFAULT '',
    category          TEXT NOT NULL DEFAULT '',
    trigger_condition TEXT NOT NULL DEFAULT '',
    problem_desc      TEXT NOT NULL DEFAULT '',
    correct_practice  TEXT NOT NULL DEFAULT '',
    change_type       TEXT NOT NULL DEFAULT '',
    frequency         INTEGER NOT NULL DEFAULT 1,
    embedding         BLOB,
    created_at        TEXT NOT NULL
);

-- 提炼出的审查规则
CREATE TABLE IF NOT EXISTS knowledge_rules (
    id          TEXT PRIMARY KEY,
    title       TEXT NOT NULL,
    rule_type   TEXT NOT NULL DEFAULT '',
    pattern     TEXT NOT NULL DEFAULT '',
    suggestion  TEXT NOT NULL DEFAULT '',
    derived_from TEXT NOT NULL DEFAULT '[]', -- JSON ids
    enabled     INTEGER NOT NULL DEFAULT 1,
    hit_count   INTEGER NOT NULL DEFAULT 0,
    created_at  TEXT NOT NULL,
    updated_at  TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS review_learning_suggestions (
    id                  TEXT PRIMARY KEY,
    review_id           TEXT NOT NULL DEFAULT '',
    guide_id            TEXT NOT NULL DEFAULT '',
    template_id         TEXT NOT NULL DEFAULT '',
    status              TEXT NOT NULL DEFAULT 'pending',
    raw_note            TEXT NOT NULL DEFAULT '',
    summary             TEXT NOT NULL DEFAULT '',
    issues              TEXT NOT NULL DEFAULT '[]',
    rules               TEXT NOT NULL DEFAULT '[]',
    template_suggestion TEXT NOT NULL DEFAULT '',
    created_at          TEXT NOT NULL,
    applied_at          TEXT
);
CREATE INDEX IF NOT EXISTS idx_learning_suggestions_status ON review_learning_suggestions(status);

-- RAG 知识片段（规则/最佳实践/历史问题统一向量化）
CREATE TABLE IF NOT EXISTS knowledge_chunks (
    id          TEXT PRIMARY KEY,
    source_type TEXT NOT NULL DEFAULT '',
    source_id   TEXT NOT NULL DEFAULT '',
    text        TEXT NOT NULL DEFAULT '',
    embedding   BLOB,
    created_at  TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS audit_logs (
    id          TEXT PRIMARY KEY,
    actor       TEXT NOT NULL DEFAULT '',
    action      TEXT NOT NULL DEFAULT '',
    target_type TEXT NOT NULL DEFAULT '',
    target_id   TEXT NOT NULL DEFAULT '',
    detail      TEXT NOT NULL DEFAULT '{}',
    created_at  TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS settings (
    key        TEXT PRIMARY KEY,
    value      TEXT NOT NULL DEFAULT '{}',
    updated_at TEXT NOT NULL
);
