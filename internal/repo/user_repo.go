package repo

import (
	"database/sql"
	"time"

	"reviewbuddy/internal/model"
)

type UserRepo struct{ db *sql.DB }

func NewUserRepo(db *sql.DB) *UserRepo { return &UserRepo{db: db} }

func (r *UserRepo) List() ([]model.User, error) {
	rows, err := r.db.Query(`SELECT id,username,password_hash,role,enabled,created_at,updated_at FROM users ORDER BY role, username`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.User{}
	for rows.Next() {
		var u model.User
		var enabled int
		if err := rows.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Role, &enabled, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		u.Enabled = enabled != 0
		if err := r.fillRoles(&u); err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

func (r *UserRepo) Create(u *model.User) error {
	normalizeUserRoles(u)
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.Exec(`INSERT INTO users (id,username,password_hash,role,enabled,created_at,updated_at) VALUES (?,?,?,?,?,?,?)`,
		u.ID, u.Username, u.PasswordHash, u.Role, boolToInt(u.Enabled), u.CreatedAt, u.UpdatedAt); err != nil {
		return err
	}
	if err := saveUserRolesTx(tx, u.ID, u.Roles); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *UserRepo) Update(u *model.User) error {
	normalizeUserRoles(u)
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.Exec(`UPDATE users SET username=?,role=?,enabled=?,updated_at=? WHERE id=?`,
		u.Username, u.Role, boolToInt(u.Enabled), u.UpdatedAt, u.ID); err != nil {
		return err
	}
	if err := saveUserRolesTx(tx, u.ID, u.Roles); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *UserRepo) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM auth_tokens WHERE user_id=?`, id)
	if err != nil {
		return err
	}
	_, err = r.db.Exec(`DELETE FROM user_domains WHERE user_id=?`, id)
	if err != nil {
		return err
	}
	_, err = r.db.Exec(`DELETE FROM domain_role_users WHERE user_id=?`, id)
	if err != nil {
		return err
	}
	_, err = r.db.Exec(`DELETE FROM user_roles WHERE user_id=?`, id)
	if err != nil {
		return err
	}
	_, err = r.db.Exec(`DELETE FROM users WHERE id=?`, id)
	return err
}

func (r *UserRepo) Get(id string) (*model.User, error) {
	return r.scanOne(`SELECT id,username,password_hash,role,enabled,created_at,updated_at FROM users WHERE id=?`, id)
}

func (r *UserRepo) GetByUsername(username string) (*model.User, error) {
	return r.scanOne(`SELECT id,username,password_hash,role,enabled,created_at,updated_at FROM users WHERE username=?`, username)
}

func (r *UserRepo) CountPasswordUsers() (int, error) {
	var n int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM users WHERE password_hash != ''`).Scan(&n)
	return n, err
}

func (r *UserRepo) CountRole(role string) (int, error) {
	var n int
	err := r.db.QueryRow(`SELECT COUNT(DISTINCT user_id) FROM user_roles WHERE role_key=?`, role).Scan(&n)
	return n, err
}

func (r *UserRepo) SaveToken(t *model.AuthToken) error {
	_, err := r.db.Exec(`INSERT INTO auth_tokens (token,user_id,expires_at,created_at) VALUES (?,?,?,?)`,
		t.Token, t.UserID, t.ExpiresAt, t.CreatedAt)
	return err
}

func (r *UserRepo) GetToken(token string) (*model.AuthToken, error) {
	var t model.AuthToken
	err := r.db.QueryRow(`SELECT token,user_id,expires_at,created_at FROM auth_tokens WHERE token=?`, token).
		Scan(&t.Token, &t.UserID, &t.ExpiresAt, &t.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *UserRepo) DeleteToken(token string) error {
	_, err := r.db.Exec(`DELETE FROM auth_tokens WHERE token=?`, token)
	return err
}

func (r *UserRepo) DeleteExpiredTokens(now time.Time) error {
	_, err := r.db.Exec(`DELETE FROM auth_tokens WHERE expires_at <= ?`, now.Format(time.RFC3339))
	return err
}

func (r *UserRepo) scanOne(query string, args ...any) (*model.User, error) {
	var u model.User
	var enabled int
	err := r.db.QueryRow(query, args...).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Role, &enabled, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}
	u.Enabled = enabled != 0
	if err := r.fillRoles(&u); err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) fillRoles(u *model.User) error {
	rows, err := r.db.Query(`SELECT role_key FROM user_roles WHERE user_id=? ORDER BY role_key`, u.ID)
	if err != nil {
		return err
	}
	defer rows.Close()
	roles := []string{}
	for rows.Next() {
		var role string
		if err := rows.Scan(&role); err != nil {
			return err
		}
		roles = append(roles, role)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if len(roles) == 0 && u.Role != "" {
		roles = append(roles, u.Role)
	}
	u.Roles = roles
	u.Role = primaryRole(roles, u.Role)
	return nil
}

func saveUserRolesTx(tx *sql.Tx, userID string, roles []string) error {
	if _, err := tx.Exec(`DELETE FROM user_roles WHERE user_id=?`, userID); err != nil {
		return err
	}
	for _, role := range roles {
		if _, err := tx.Exec(`INSERT INTO user_roles (user_id,role_key) VALUES (?,?)`, userID, role); err != nil {
			return err
		}
	}
	return nil
}

func normalizeUserRoles(u *model.User) {
	roles := uniqueNonEmpty(u.Roles)
	if len(roles) == 0 && u.Role != "" {
		roles = []string{u.Role}
	}
	if len(roles) == 0 {
		roles = []string{"readonly"}
	}
	u.Roles = roles
	u.Role = primaryRole(roles, u.Role)
}

func uniqueNonEmpty(in []string) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, v := range in {
		if v == "" || seen[v] {
			continue
		}
		seen[v] = true
		out = append(out, v)
	}
	return out
}

func primaryRole(roles []string, fallback string) string {
	for _, role := range roles {
		if role == "admin" {
			return role
		}
	}
	for _, role := range roles {
		if role != "readonly" {
			return role
		}
	}
	if len(roles) > 0 {
		return roles[0]
	}
	return fallback
}
