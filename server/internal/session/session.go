package session

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Manager manages authentication sessions in Redis.
// Design: JWT token contains a session_id claim. Redis stores active sessions.
// On login: create session (key: auth:session:{user_id}:{session_id}) with TTL.
// On logout: delete session key. Middleware checks session existence.
type Manager struct {
	client *redis.Client
	ttl    time.Duration
}

// NewManager creates a new session manager.
func NewManager(client *redis.Client, ttlSeconds int) *Manager {
	return &Manager{
		client: client,
		ttl:    time.Duration(ttlSeconds) * time.Second,
	}
}

// CreateSession stores a session in Redis and returns the session ID.
func (m *Manager) CreateSession(ctx context.Context, userID uint64, sessionID string) error {
	key := sessionKey(userID, sessionID)
	return m.client.Set(ctx, key, "1", m.ttl).Err()
}

// ValidateSession checks if a session exists and refreshes its TTL.
func (m *Manager) ValidateSession(ctx context.Context, userID uint64, sessionID string) error {
	key := sessionKey(userID, sessionID)
	val, err := m.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return fmt.Errorf("session not found")
	}
	if err != nil {
		return fmt.Errorf("redis error: %w", err)
	}
	if val != "1" {
		return fmt.Errorf("invalid session data")
	}
	return nil
}

// DestroySession removes a session from Redis.
func (m *Manager) DestroySession(ctx context.Context, userID uint64, sessionID string) error {
	key := sessionKey(userID, sessionID)
	return m.client.Del(ctx, key).Err()
}

// DestroyAllUserSessions removes all sessions for a user.
func (m *Manager) DestroyAllUserSessions(ctx context.Context, userID uint64) error {
	pattern := fmt.Sprintf("auth:session:%d:*", userID)
	iter := m.client.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		if err := m.client.Del(ctx, iter.Val()).Err(); err != nil {
			return err
		}
	}
	return iter.Err()
}

// Ping checks Redis connectivity.
func (m *Manager) Ping(ctx context.Context) error {
	return m.client.Ping(ctx).Err()
}

func sessionKey(userID uint64, sessionID string) string {
	return fmt.Sprintf("auth:session:%d:%s", userID, sessionID)
}
