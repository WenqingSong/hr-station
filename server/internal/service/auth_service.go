package service

import (
	"context"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	apperror "offer-catcher/server/internal/errors"
	"offer-catcher/server/internal/model"
	"offer-catcher/server/internal/repository"
)

// RegisterRequest represents a user registration request.
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

// LoginRequest represents a user login request.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthUser represents user info returned in auth responses.
type AuthUser struct {
	ID    uint64 `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

// AuthResponse represents a successful registration or login response.
type AuthResponse struct {
	Token string   `json:"token"`
	User  AuthUser `json:"user"`
}

// Claims represents JWT claims for the application.
type Claims struct {
	UserID uint64 `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// AuthService handles authentication business logic.
type AuthService struct {
	userRepo       *repository.UserRepository
	jwtSecret      string
	jwtExpireHours int
}

// NewAuthService creates a new AuthService.
func NewAuthService(userRepo *repository.UserRepository, jwtSecret string, jwtExpireHours int) *AuthService {
	return &AuthService{
		userRepo:       userRepo,
		jwtSecret:      jwtSecret,
		jwtExpireHours: jwtExpireHours,
	}
}

func validateEmail(email string) error {
	email = strings.TrimSpace(email)
	if email == "" {
		return apperror.New(apperror.CodeInvalidParam, "email is required", 400)
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return apperror.New(apperror.CodeInvalidParam, "invalid email format", 400)
	}
	return nil
}

func validatePassword(password string) error {
	password = strings.TrimSpace(password)
	if password == "" {
		return apperror.New(apperror.CodeInvalidParam, "password is required", 400)
	}
	if len(password) < 6 {
		return apperror.New(apperror.CodeInvalidParam, "password must be at least 6 characters", 400)
	}
	return nil
}

// Register registers a new user and returns an auth response with JWT.
func (s *AuthService) Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error) {
	req.Email = strings.TrimSpace(req.Email)
	if err := validateEmail(req.Email); err != nil {
		return nil, err
	}
	if err := validatePassword(req.Password); err != nil {
		return nil, err
	}

	// Check if email already exists
	existing, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check email: %w", err)
	}
	if existing != nil {
		return nil, apperror.ErrEmailRegistered
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &model.User{
		Email:        req.Email,
		PasswordHash: string(hash),
		Name:         req.Name,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	token, err := s.generateToken(user.ID, user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &AuthResponse{
		Token: token,
		User: AuthUser{
			ID:    user.ID,
			Email: user.Email,
			Name:  user.Name,
		},
	}, nil
}

// Login authenticates a user and returns an auth response with JWT.
func (s *AuthService) Login(ctx context.Context, req LoginRequest) (*AuthResponse, error) {
	req.Email = strings.TrimSpace(req.Email)
	if err := validateEmail(req.Email); err != nil {
		return nil, err
	}
	if strings.TrimSpace(req.Password) == "" {
		return nil, apperror.New(apperror.CodeInvalidParam, "password is required", 400)
	}

	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return nil, apperror.ErrInvalidCreds
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, apperror.ErrInvalidCreds
	}

	token, err := s.generateToken(user.ID, user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &AuthResponse{
		Token: token,
		User: AuthUser{
			ID:    user.ID,
			Email: user.Email,
			Name:  user.Name,
		},
	}, nil
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
		Email: user.Email,
		Name:  user.Name,
	}, nil
}

// ParseToken parses a JWT token string and returns the user ID.
func (s *AuthService) ParseToken(tokenString string) (uint64, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})
	if err != nil {
		return 0, apperror.ErrUnauthorized
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return 0, apperror.ErrUnauthorized
	}

	return claims.UserID, nil
}

// generateToken generates a JWT token for the given user.
func (s *AuthService) generateToken(userID uint64, email string) (string, error) {
	now := time.Now()
	claims := &Claims{
		UserID: userID,
		Email:  email,
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
