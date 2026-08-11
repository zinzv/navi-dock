package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/navi-dock/navi-dock/internal/model"
	"github.com/navi-dock/navi-dock/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUsernameEmpty      = errors.New("username is required")
	ErrUsernameTooShort   = errors.New("username must be at least 2 characters")
	ErrUsernameTooLong    = errors.New("username must be at most 32 characters")
	ErrPasswordEmpty      = errors.New("password is required")
	ErrPasswordTooShort   = errors.New("password must be at least 6 characters")
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrUserExists         = errors.New("username already exists")
	ErrUserNotFound       = errors.New("user not found")
	ErrWrongPassword      = errors.New("current password is incorrect")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrForbidden          = errors.New("forbidden")
	ErrLastAdmin          = errors.New("cannot delete the last admin")
	ErrInvalidToken       = errors.New("invalid or expired token")
)

type AuthService struct {
	users  *repository.UserRepo
	secret []byte
}

func NewAuthService(users *repository.UserRepo, secret string) *AuthService {
	if secret == "" {
		secret = "navi-dock-dev-secret-change-me"
	}
	return &AuthService{users: users, secret: []byte(secret)}
}

type tokenClaims struct {
	UID  string `json:"uid"`
	Name string `json:"name"`
	Role string `json:"role"`
	Exp  int64  `json:"exp"`
}

type AuthStatus struct {
	SetupRequired bool             `json:"setup_required"`
	User          *model.UserPublic `json:"user,omitempty"`
}

type LoginResult struct {
	Token string           `json:"token"`
	User  model.UserPublic `json:"user"`
}

func (s *AuthService) Status(token string) (AuthStatus, error) {
	count, err := s.users.Count()
	if err != nil {
		return AuthStatus{}, err
	}
	status := AuthStatus{SetupRequired: count == 0}
	if token == "" {
		return status, nil
	}
	user, err := s.UserFromToken(token)
	if err != nil {
		return status, nil
	}
	pub := user.Public()
	status.User = &pub
	return status, nil
}

func (s *AuthService) Login(input model.LoginInput) (*LoginResult, error) {
	username := strings.TrimSpace(input.Username)
	if username == "" || input.Password == "" {
		return nil, ErrInvalidCredentials
	}
	user, err := s.users.FindByUsername(username)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}
	token, err := s.issueToken(user)
	if err != nil {
		return nil, err
	}
	return &LoginResult{Token: token, User: user.Public()}, nil
}

func (s *AuthService) CreateUser(input model.CreateUserInput, actor *model.User) (*model.UserPublic, error) {
	username, password, err := validateCredentials(input.Username, input.Password)
	if err != nil {
		return nil, err
	}

	count, err := s.users.Count()
	if err != nil {
		return nil, err
	}

	role := strings.TrimSpace(input.Role)
	if role == "" {
		role = "admin"
	}
	if role != "admin" && role != "user" {
		role = "user"
	}

	if count == 0 {
		role = "admin"
	} else {
		if actor == nil || actor.Role != "admin" {
			return nil, ErrForbidden
		}
	}

	existing, err := s.users.FindByUsername(username)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrUserExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		ID:           repository.NewID(),
		Username:     username,
		PasswordHash: string(hash),
		Role:         role,
	}
	if err := s.users.Create(user); err != nil {
		return nil, err
	}
	pub := user.Public()
	return &pub, nil
}

func (s *AuthService) ChangePassword(userID string, input model.ChangePasswordInput) error {
	if input.NewPassword == "" {
		return ErrPasswordEmpty
	}
	if utf8.RuneCountInString(input.NewPassword) < 6 {
		return ErrPasswordTooShort
	}
	user, err := s.users.FindByID(userID)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.OldPassword)); err != nil {
		return ErrWrongPassword
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.users.UpdatePassword(userID, string(hash))
}

func (s *AuthService) ListUsers(actor *model.User) ([]model.UserPublic, error) {
	if actor == nil {
		return nil, ErrUnauthorized
	}
	users, err := s.users.List()
	if err != nil {
		return nil, err
	}
	out := make([]model.UserPublic, 0, len(users))
	for _, u := range users {
		out = append(out, u.Public())
	}
	return out, nil
}

func (s *AuthService) DeleteUser(actor *model.User, id string) error {
	if actor == nil || actor.Role != "admin" {
		return ErrForbidden
	}
	target, err := s.users.FindByID(id)
	if err != nil {
		return err
	}
	if target == nil {
		return ErrUserNotFound
	}
	if target.Role == "admin" {
		n, err := s.users.CountAdmins()
		if err != nil {
			return err
		}
		if n <= 1 {
			return ErrLastAdmin
		}
	}
	return s.users.Delete(id)
}

func (s *AuthService) UserFromToken(token string) (*model.User, error) {
	claims, err := s.parseToken(token)
	if err != nil {
		return nil, err
	}
	user, err := s.users.FindByID(claims.UID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrInvalidToken
	}
	return user, nil
}

func (s *AuthService) issueToken(user *model.User) (string, error) {
	claims := tokenClaims{
		UID:  user.ID,
		Name: user.Username,
		Role: user.Role,
		Exp:  0, // 0 = never expires
	}
	raw, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	payload := base64.RawURLEncoding.EncodeToString(raw)
	mac := hmac.New(sha256.New, s.secret)
	_, _ = mac.Write([]byte(payload))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return payload + "." + sig, nil
}

func (s *AuthService) parseToken(token string) (*tokenClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return nil, ErrInvalidToken
	}
	payload, sig := parts[0], parts[1]
	mac := hmac.New(sha256.New, s.secret)
	_, _ = mac.Write([]byte(payload))
	expected := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(expected), []byte(sig)) {
		return nil, ErrInvalidToken
	}
	raw, err := base64.RawURLEncoding.DecodeString(payload)
	if err != nil {
		return nil, ErrInvalidToken
	}
	var claims tokenClaims
	if err := json.Unmarshal(raw, &claims); err != nil {
		return nil, ErrInvalidToken
	}
	if claims.UID == "" {
		return nil, ErrInvalidToken
	}
	// Exp == 0 means never expires; otherwise honor legacy expiry.
	if claims.Exp > 0 && claims.Exp < time.Now().Unix() {
		return nil, ErrInvalidToken
	}
	return &claims, nil
}

func validateCredentials(username, password string) (string, string, error) {
	username = strings.TrimSpace(username)
	if username == "" {
		return "", "", ErrUsernameEmpty
	}
	n := utf8.RuneCountInString(username)
	if n < 2 {
		return "", "", ErrUsernameTooShort
	}
	if n > 32 {
		return "", "", ErrUsernameTooLong
	}
	if password == "" {
		return "", "", ErrPasswordEmpty
	}
	if utf8.RuneCountInString(password) < 6 {
		return "", "", ErrPasswordTooShort
	}
	return username, password, nil
}
