package repo

import (
	"database/sql"
	"encoding/json"

	"reviewbuddy/internal/model"
)

type ReviewConfigRepo struct{ db *sql.DB }

func NewReviewConfigRepo(db *sql.DB) *ReviewConfigRepo { return &ReviewConfigRepo{db: db} }

func (r *ReviewConfigRepo) ListRoles() ([]model.ReviewRole, error) {
	rows, err := r.db.Query(`SELECT id,role_key,name,description,system,created_at,updated_at FROM review_roles ORDER BY system DESC, name ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.ReviewRole{}
	for rows.Next() {
		var item model.ReviewRole
		var system int
		if err := rows.Scan(&item.ID, &item.Key, &item.Name, &item.Description, &system, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		item.System = system == 1
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *ReviewConfigRepo) RoleExists(key string) (bool, error) {
	var n int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM review_roles WHERE role_key=?`, key).Scan(&n)
	return n > 0, err
}

func (r *ReviewConfigRepo) CreateRole(item *model.ReviewRole) error {
	_, err := r.db.Exec(`INSERT INTO review_roles (id,role_key,name,description,system,created_at,updated_at) VALUES (?,?,?,?,?,?,?)`,
		item.ID, item.Key, item.Name, item.Description, boolToInt(item.System), item.CreatedAt, item.UpdatedAt)
	return err
}

func (r *ReviewConfigRepo) UpdateRole(item *model.ReviewRole) error {
	_, err := r.db.Exec(`UPDATE review_roles SET name=?,description=?,updated_at=? WHERE role_key=? AND system=0`,
		item.Name, item.Description, item.UpdatedAt, item.Key)
	return err
}

func (r *ReviewConfigRepo) DeleteRole(key string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.Exec(`DELETE FROM domain_role_users WHERE role_key=?`, key); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM review_roles WHERE role_key=? AND system=0`, key); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *ReviewConfigRepo) ListDomains() ([]model.ReviewDomain, error) {
	rows, err := r.db.Query(`SELECT id,name,description,created_at,updated_at FROM review_domains ORDER BY created_at ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.ReviewDomain{}
	for rows.Next() {
		var item model.ReviewDomain
		if err := rows.Scan(&item.ID, &item.Name, &item.Description, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *ReviewConfigRepo) SaveDomain(item *model.ReviewDomain) error {
	_, err := r.db.Exec(`INSERT INTO review_domains (id,name,description,created_at,updated_at) VALUES (?,?,?,?,?)
		ON CONFLICT(id) DO UPDATE SET name=excluded.name,description=excluded.description,updated_at=excluded.updated_at`,
		item.ID, item.Name, item.Description, item.CreatedAt, item.UpdatedAt)
	return err
}

func (r *ReviewConfigRepo) DeleteDomain(id string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.Exec(`DELETE FROM domain_role_users WHERE domain_id=?`, id); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM user_domains WHERE domain_id=?`, id); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM review_domains WHERE id=?`, id); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *ReviewConfigRepo) ListDomainRoleUsers(domainID string) ([]model.DomainRoleUsers, error) {
	rows, err := r.db.Query(`SELECT domain_id,role_key,user_id FROM domain_role_users WHERE domain_id=? ORDER BY role_key`, domainID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	byRole := map[string]*model.DomainRoleUsers{}
	order := []string{}
	for rows.Next() {
		var domain, role, userID string
		if err := rows.Scan(&domain, &role, &userID); err != nil {
			return nil, err
		}
		if byRole[role] == nil {
			byRole[role] = &model.DomainRoleUsers{DomainID: domain, RoleKey: role, UserIDs: []string{}}
			order = append(order, role)
		}
		byRole[role].UserIDs = append(byRole[role].UserIDs, userID)
	}
	out := []model.DomainRoleUsers{}
	for _, role := range order {
		out = append(out, *byRole[role])
	}
	return out, rows.Err()
}

func (r *ReviewConfigRepo) SaveDomainRoleUsers(item *model.DomainRoleUsers) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.Exec(`DELETE FROM domain_role_users WHERE domain_id=? AND role_key=?`, item.DomainID, item.RoleKey); err != nil {
		return err
	}
	for _, userID := range item.UserIDs {
		if _, err := tx.Exec(`INSERT INTO domain_role_users (domain_id,role_key,user_id) VALUES (?,?,?)`, item.DomainID, item.RoleKey, userID); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *ReviewConfigRepo) ListUserDomains(userID string) (*model.UserDomains, error) {
	rows, err := r.db.Query(`SELECT domain_id FROM user_domains WHERE user_id=? ORDER BY domain_id`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := &model.UserDomains{UserID: userID, DomainIDs: []string{}}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out.DomainIDs = append(out.DomainIDs, id)
	}
	return out, rows.Err()
}

func (r *ReviewConfigRepo) ListAllUserDomains() (map[string][]string, error) {
	rows, err := r.db.Query(`SELECT user_id,domain_id FROM user_domains ORDER BY user_id, domain_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string][]string{}
	for rows.Next() {
		var userID, domainID string
		if err := rows.Scan(&userID, &domainID); err != nil {
			return nil, err
		}
		out[userID] = append(out[userID], domainID)
	}
	return out, rows.Err()
}

func (r *ReviewConfigRepo) SaveUserDomains(item *model.UserDomains) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.Exec(`DELETE FROM user_domains WHERE user_id=?`, item.UserID); err != nil {
		return err
	}
	for _, domainID := range item.DomainIDs {
		if _, err := tx.Exec(`INSERT INTO user_domains (user_id,domain_id) VALUES (?,?)`, item.UserID, domainID); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *ReviewConfigRepo) ListScenarios() ([]model.ReviewScenario, error) {
	rows, err := r.db.Query(`SELECT id,name,description,role_keys,created_at,updated_at FROM review_scenarios ORDER BY created_at ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.ReviewScenario{}
	for rows.Next() {
		var item model.ReviewScenario
		var roleKeys string
		if err := rows.Scan(&item.ID, &item.Name, &item.Description, &roleKeys, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		_ = json.Unmarshal([]byte(roleKeys), &item.RoleKeys)
		if item.RoleKeys == nil {
			item.RoleKeys = []string{}
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *ReviewConfigRepo) SaveScenario(item *model.ReviewScenario) error {
	roleKeys, _ := json.Marshal(item.RoleKeys)
	_, err := r.db.Exec(`INSERT INTO review_scenarios (id,name,description,role_keys,created_at,updated_at) VALUES (?,?,?,?,?,?)
		ON CONFLICT(id) DO UPDATE SET name=excluded.name,description=excluded.description,role_keys=excluded.role_keys,updated_at=excluded.updated_at`,
		item.ID, item.Name, item.Description, string(roleKeys), item.CreatedAt, item.UpdatedAt)
	return err
}

func (r *ReviewConfigRepo) DeleteScenario(id string) error {
	_, err := r.db.Exec(`DELETE FROM review_scenarios WHERE id=?`, id)
	return err
}
