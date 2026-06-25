package service

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	apperror "offer-catcher/server/internal/errors"
	"offer-catcher/server/internal/model"
	"offer-catcher/server/internal/repository"
	"offer-catcher/server/internal/session"
)

// AuthUser represents user info returned in auth responses.
type AuthUser struct {
	ID    uint64 `json:"id"`
	Phone string `json:"phone"`
	Name  string `json:"name"`
}

// AuthResponse represents a successful registration or login response.
type AuthResponse struct {
	Token string   `json:"token"`
	User  AuthUser `json:"user"`
}

// Claims represents JWT claims for the application.
type Claims struct {
	UserID    uint64 `json:"user_id"`
	Phone     string `json:"phone"`
	SessionID string `json:"session_id"`
	jwt.RegisteredClaims
}

var phoneRegex = regexp.MustCompile(`^1[3-9]\d{9}$`)

// AuthService handles authentication business logic.
type AuthService struct {
	userRepo       *repository.UserRepository
	sessionMgr     *session.Manager
	jwtSecret      string
	jwtExpireHours int
}

// NewAuthService creates a new AuthService.
func NewAuthService(userRepo *repository.UserRepository, sessionMgr *session.Manager, jwtSecret string, jwtExpireHours int) *AuthService {
	return &AuthService{
		userRepo:       userRepo,
		sessionMgr:     sessionMgr,
		jwtSecret:      jwtSecret,
		jwtExpireHours: jwtExpireHours,
	}
}

func validatePhone(phone string) error {
	phone = strings.TrimSpace(phone)
	if phone == "" {
		return apperror.New(apperror.CodeInvalidParam, "手机号不能为空", 400)
	}
	if !phoneRegex.MatchString(phone) {
		return apperror.New(apperror.CodeInvalidParam, "手机号格式不正确", 400)
	}
	return nil
}

func validatePassword(password string) error {
	password = strings.TrimSpace(password)
	if password == "" {
		return apperror.New(apperror.CodeInvalidParam, "密码不能为空", 400)
	}
	if len(password) < 6 {
		return apperror.New(apperror.CodeInvalidParam, "密码长度不能少于 6 位", 400)
	}
	return nil
}

// Register registers a new user and returns an auth response with JWT.
func (s *AuthService) Register(ctx context.Context, req model.RegisterRequest) (*AuthResponse, error) {
	req.Phone = strings.TrimSpace(req.Phone)
	if err := validatePhone(req.Phone); err != nil {
		return nil, err
	}
	if err := validatePassword(req.Password); err != nil {
		return nil, err
	}

	// Check if phone already exists
	existing, err := s.userRepo.FindByPhone(ctx, req.Phone)
	if err != nil {
		return nil, fmt.Errorf("failed to check phone: %w", err)
	}
	if existing != nil {
		return nil, apperror.New(apperror.CodeDuplicatePhone, "该手机号已被注册", 409)
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &model.User{
		Phone:        req.Phone,
		PasswordHash: string(hash),
		Name:         req.Phone,
	}

	if err := s.userRepo.CreateByPhone(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Create session and generate token
	return s.createSessionAndToken(ctx, user)
}

// Login authenticates a user and returns an auth response with JWT.
func (s *AuthService) Login(ctx context.Context, req model.LoginRequest) (*AuthResponse, error) {
	req.Phone = strings.TrimSpace(req.Phone)
	if err := validatePhone(req.Phone); err != nil {
		return nil, err
	}
	if strings.TrimSpace(req.Password) == "" {
		return nil, apperror.New(apperror.CodeInvalidParam, "密码不能为空", 400)
	}

	user, err := s.userRepo.FindByPhone(ctx, req.Phone)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return nil, apperror.New(apperror.CodeInvalidCreds, "手机号或密码错误", 401)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, apperror.New(apperror.CodeInvalidCreds, "手机号或密码错误", 401)
	}

	return s.createSessionAndToken(ctx, user)
}

// Me returns the current user info.
func (s *AuthService) Me(ctx context.Context, userID uint64) (*AuthUser, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return nil, apperror.ErrNotFound
	}

	return &AuthUser{
		ID:    user.ID,
		Phone: user.Phone,
		Name:  user.Name,
	}, nil
}

// Logout removes the session from Redis and conditionally clears the current pointer.
func (s *AuthService) Logout(ctx context.Context, userID uint64, sessionID string) error {
	// Delete current session
	if err := s.sessionMgr.DestroySession(ctx, userID, sessionID); err != nil {
		return err
	}

	// Only clear current pointer if it matches this session (prevent old device
	// logout from deleting the new device's current pointer)
	current, err := s.sessionMgr.GetCurrentSessionID(ctx, userID)
	if err != nil {
		return err
	}
	if current == sessionID {
		return s.sessionMgr.ClearCurrentSession(ctx, userID)
	}
	return nil
}

// ParseToken parses a JWT token and returns user_id + session_id.
func (s *AuthService) ParseToken(tokenString string) (uint64, string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})
	if err != nil {
		return 0, "", apperror.ErrUnauthorized
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return 0, "", apperror.ErrUnauthorized
	}

	return claims.UserID, claims.SessionID, nil
}

// ValidateSession validates the session with current-session check.
// Returns apperror.ErrSessionReplaced if the token's session was replaced by a newer login.
func (s *AuthService) ValidateSession(ctx context.Context, userID uint64, sessionID string) error {
	return s.sessionMgr.ValidateSessionWithCurrent(ctx, userID, sessionID)
}

// createSessionAndToken replaces old session, creates a new one, and generates a JWT.
func (s *AuthService) createSessionAndToken(ctx context.Context, user *model.User) (*AuthResponse, error) {
	sessionID := uuid.NewString()
	// Replace old session — deletes old session, creates new one, updates current pointer
	if err := s.sessionMgr.ReplaceUserSession(ctx, user.ID, sessionID); err != nil {
		return nil, fmt.Errorf("failed to replace session: %w", err)
	}

	token, err := s.generateToken(user.ID, user.Phone, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &AuthResponse{
		Token: token,
		User: AuthUser{
			ID:    user.ID,
			Phone: user.Phone,
			Name:  user.Name,
		},
	}, nil
}

// generateToken generates a JWT token with session_id claim.
func (s *AuthService) generateToken(userID uint64, phone, sessionID string) (string, error) {
	now := time.Now()
	claims := &Claims{
		UserID:    userID,
		Phone:     phone,
		SessionID: sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(s.jwtExpireHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "offer-catcher",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}
