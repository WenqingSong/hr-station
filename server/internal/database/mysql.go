package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"offer-catcher/server/internal/config"
)

// MySQL wraps a *sql.DB connection with lifecycle management.
type MySQL struct {
	DB *sql.DB
}

// NewMySQL creates a new MySQL connection using the provided DatabaseConfig.
func NewMySQL(cfg config.DatabaseConfig) (*MySQL, error) {
	if cfg.DSN == "" {
		return nil, fmt.Errorf("database DSN is empty")
	}

	db, err := sql.Open("mysql", cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)
	db.SetConnMaxIdleTime(10 * time.Minute)

	return &MySQL{DB: db}, nil
}

// Ping verifies the database connection is alive within the given context.
func (m *MySQL) Ping(ctx context.Context) error {
	return m.DB.PingContext(ctx)
}

// Close closes the database connection and releases resources.
func (m *MySQL) Close() error {
	return m.DB.Close()
}
