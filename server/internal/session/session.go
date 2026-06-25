package session

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Manager manages authentication sessions in Redis.
//
// Key design:
//
//	auth:session:{userID}:{sessionID}     = "1"               (active session)
//	auth:user_current:{userID}            = "{sessionID}"      (current session pointer)
//
// On login: replace old session → create new session → update current pointer.
// On logout: delete current session → conditionally clear current pointer.
// On auth: validate session AND check current pointer matches token sessionID.
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

// CreateSession stores a session in Redis.
func (m *Manager) CreateSession(ctx context.Context, userID uint64, sessionID string) error {
	key := sessionKey(userID, sessionID)
	return m.client.Set(ctx, key, "1", m.ttl).Err()
}

// ValidateSession checks if a session exists in Redis.
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

// DestroyAllUserSessions removes all sessions for a user (does not touch current pointer).
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

// ─── Current session pointer ────────────────────────────────────────────────────

func currentSessionKey(userID uint64) string {
	return fmt.Sprintf("auth:user_current:%d", userID)
}

// GetCurrentSessionID returns the current session ID for a user.
// Returns empty string if no current session pointer exists.
func (m *Manager) GetCurrentSessionID(ctx context.Context, userID uint64) (string, error) {
	val, err := m.client.Get(ctx, currentSessionKey(userID)).Result()
	if err == redis.Nil {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("redis error reading current session: %w", err)
	}
	return val, nil
}

// SetCurrentSessionID sets the current session pointer with session TTL.
func (m *Manager) SetCurrentSessionID(ctx context.Context, userID uint64, sessionID string) error {
	return m.client.Set(ctx, currentSessionKey(userID), sessionID, m.ttl).Err()
}

// ClearCurrentSession deletes the current session pointer for a user.
func (m *Manager) ClearCurrentSession(ctx context.Context, userID uint64) error {
	return m.client.Del(ctx, currentSessionKey(userID)).Err()
}

// ReplaceUserSession performs an atomic session replacement:
//  1. Read old current sessionID
//  2. Delete old session (if exists)
//  3. Create new session
//  4. Update current pointer to new sessionID
func (m *Manager) ReplaceUserSession(ctx context.Context, userID uint64, newSessionID string) error {
	// Delete old session if current pointer exists
	oldSessionID, err := m.GetCurrentSessionID(ctx, userID)
	if err != nil {
		return err
	}
	if oldSessionID != "" {
		m.client.Del(ctx, sessionKey(userID, oldSessionID))
	}

	// Create new session
	if err := m.CreateSession(ctx, userID, newSessionID); err != nil {
		return err
	}

	// Update current pointer
	if err := m.SetCurrentSessionID(ctx, userID, newSessionID); err != nil {
		return err
	}

	return nil
}

// ValidateSessionWithCurrent validates that the token's session is the current active session.
// Returns ErrSessionReplaced if current session pointer exists but doesn't match.
// Returns the regular session validation error if session doesn't exist.
// Returns nil if session is valid and is the current session.
func (m *Manager) ValidateSessionWithCurrent(ctx context.Context, userID uint64, sessionID string) error {
	// Check current session pointer
	current, err := m.GetCurrentSessionID(ctx, userID)
	if err != nil {
		return err
	}
	if current != "" && current != sessionID {
		return ErrSessionReplaced
	}

	// Check session existence
	return m.ValidateSession(ctx, userID, sessionID)
}

// ErrSessionReplaced is returned when a token's session has been replaced by a newer login.
var ErrSessionReplaced = fmt.Errorf("session replaced by newer login")

// Ping checks Redis connectivity.
func (m *Manager) Ping(ctx context.Context) error {
	return m.client.Ping(ctx).Err()
}

func sessionKey(userID uint64, sessionID string) string {
	return fmt.Sprintf("auth:session:%d:%s", userID, sessionID)
}
