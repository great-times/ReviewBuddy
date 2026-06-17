package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"reviewbuddy/internal/model"
	"reviewbuddy/internal/repo"
)

const TokenTTL = 24 * time.Hour * 14

var (
	ErrInvalidCredentials = errors.New("用户名或密码错误")
	ErrUsernameTaken      = errors.New("用户名已存在")
	ErrUnauthorized       = errors.New("请先登录")
	ErrCannotDeleteSelf   = errors.New("不能删除当前登录用户")
	ErrLastAdmin          = errors.New("不能删除最后一个管理员")
)

type Service struct {
	users *repo.UserRepo
}

func NewService(users *repo.UserRepo) *Service {
	return &Service{users: users}
}

type LoginResult struct {
	Token     string      `json:"token"`
	ExpiresAt string      `json:"expiresAt"`
	User      *model.User `json:"user"`
}

func (s *Service) Register(username, password string) (*LoginResult, error) {
	username = strings.TrimSpace(username)
	if username == "" || len(password) < 4 {
		return nil, errors.New("用户名不能为空，密码至少 4 位")
	}
	if _, err := s.users.GetByUsername(username); err == nil {
		return nil, ErrUsernameTaken
	} else if err != sql.ErrNoRows {
		return nil, err
	}

	authUsers, err := s.users.CountPasswordUsers()
	if err != nil {
		return nil, err
	}
	role := "readonly"
	if authUsers == 0 {
		role = "admin"
	}

	now := time.Now().Format(time.RFC3339)
	u := &model.User{
		ID:           uuid.NewString(),
		Username:     username,
		PasswordHash: hashPassword(password),
		Role:         role,
		Enabled:      true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := s.users.Create(u); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unique") {
			return nil, ErrUsernameTaken
		}
		return nil, err
	}
	return s.issue(u)
}

func (s *Service) Login(username, password string) (*LoginResult, error) {
	u, err := s.users.GetByUsername(strings.TrimSpace(username))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}
	if !verifyPassword(password, u.PasswordHash) {
		return nil, ErrInvalidCredentials
	}
	return s.issue(u)
}

func (s *Service) Logout(token string) error {
	if token == "" {
		return nil
	}
	return s.users.DeleteToken(token)
}

func (s *Service) UserByToken(token string) (*model.User, error) {
	if token == "" {
		return nil, ErrUnauthorized
	}
	tok, err := s.users.GetToken(token)
	if err != nil {
		return nil, ErrUnauthorized
	}
	expiresAt, err := time.Parse(time.RFC3339, tok.ExpiresAt)
	if err != nil || time.Now().After(expiresAt) {
		_ = s.users.DeleteToken(token)
		return nil, ErrUnauthorized
	}
	u, err := s.users.Get(tok.UserID)
	if err != nil {
		return nil, ErrUnauthorized
	}
	return u, nil
}

func (s *Service) DeleteUser(actorID, id string) error {
	if actorID == id {
		return ErrCannotDeleteSelf
	}
	u, err := s.users.Get(id)
	if err != nil {
		return err
	}
	if u.Role == "admin" {
		n, err := s.users.CountRole("admin")
		if err != nil {
			return err
		}
		if n <= 1 {
			return ErrLastAdmin
		}
	}
	return s.users.Delete(id)
}

func (s *Service) issue(u *model.User) (*LoginResult, error) {
	token, err := randomHex(32)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	expiresAt := now.Add(TokenTTL)
	if err := s.users.SaveToken(&model.AuthToken{
		Token:     token,
		UserID:    u.ID,
		ExpiresAt: expiresAt.Format(time.RFC3339),
		CreatedAt: now.Format(time.RFC3339),
	}); err != nil {
		return nil, err
	}
	_ = s.users.DeleteExpiredTokens(now)
	return &LoginResult{Token: token, ExpiresAt: expiresAt.Format(time.RFC3339), User: u}, nil
}

func hashPassword(password string) string {
	salt, err := randomHex(16)
	if err != nil {
		panic(err)
	}
	sum := sha256.Sum256([]byte(salt + ":" + password))
	return fmt.Sprintf("sha256$%s$%s", salt, hex.EncodeToString(sum[:]))
}

func verifyPassword(password, stored string) bool {
	parts := strings.Split(stored, "$")
	if len(parts) != 3 || parts[0] != "sha256" {
		return false
	}
	sum := sha256.Sum256([]byte(parts[1] + ":" + password))
	expected := []byte(parts[2])
	actual := []byte(hex.EncodeToString(sum[:]))
	return subtle.ConstantTimeCompare(expected, actual) == 1
}

func randomHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
