package repository

import (
	"context"
	"database/sql"
	"fmt"

	"offer-catcher/server/internal/model"
)

// UserRepository handles database operations for users.
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new UserRepository.
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// CreateByPhone inserts a new user with phone.
func (r *UserRepository) CreateByPhone(ctx context.Context, user *model.User) error {
	result, err := r.db.ExecContext(ctx,
		"INSERT INTO users (phone, password_hash, name) VALUES (?, ?, ?)",
		user.Phone, user.PasswordHash, user.Name,
	)
	if err != nil {
		return fmt.Errorf("failed to create user by phone: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}
	user.ID = uint64(id)
	return nil
}

// FindByPhone finds a user by phone. Returns nil if not found.
func (r *UserRepository) FindByPhone(ctx context.Context, phone string) (*model.User, error) {
	user := &model.User{}
	err := r.db.QueryRowContext(ctx,
		"SELECT id, COALESCE(email,''), phone, password_hash, name, created_at, updated_at FROM users WHERE phone = ?",
		phone,
	).Scan(&user.ID, &user.Email, &user.Phone, &user.PasswordHash, &user.Name, &user.CreatedAt, &user.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find user by phone: %w", err)
	}
	return user, nil
}

// FindByID finds a user by ID. Returns nil if not found.
func (r *UserRepository) FindByID(ctx context.Context, id uint64) (*model.User, error) {
	user := &model.User{}
	err := r.db.QueryRowContext(ctx,
		"SELECT id, COALESCE(email,''), COALESCE(phone,''), password_hash, name, created_at, updated_at FROM users WHERE id = ?",
		id,
	).Scan(&user.ID, &user.Email, &user.Phone, &user.PasswordHash, &user.Name, &user.CreatedAt, &user.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find user by id: %w", err)
	}
	return user, nil
}
