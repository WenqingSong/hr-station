package model

import "time"

// User represents a user in the system.
// Maps to the `users` table in MySQL.
type User struct {
	ID           uint64
	Email        string
	Phone        string
	PasswordHash string
	Name         string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// LoginRequest represents a login request.
type LoginRequest struct {
	Phone    string `json:"phone"`
	Password string `json:"password"`
}

// RegisterRequest represents a registration request.
type RegisterRequest struct {
	Phone    string `json:"phone"`
	Password string `json:"password"`
}

// AuthTokenResponse is returned on successful login/registration.
type AuthTokenResponse struct {
	Token string `json:"token"`
	User  *User  `json:"user"`
}

// Claims represents JWT claims.
type Claims struct {
	UserID    uint64 `json:"user_id"`
	Phone     string `json:"phone"`
	SessionID string `json:"session_id"`
	IssuedAt  int64  `json:"iat"`
	ExpiresAt int64  `json:"exp"`
}
