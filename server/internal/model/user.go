package model

import "time"

// User represents a user in the system.
// Maps to the `users` table in MySQL.
type User struct {
	ID           uint64
	Email        string
	PasswordHash string
	Name         string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
