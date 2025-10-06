package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// SessionData holds the information stored in Redis for each session
type SessionData struct {
	UserID    uuid.UUID `json:"user_id"`
	Email     string    `json:"email"`
	UserAgent string    `json:"user_agent"`
	IPAddr    string    `json:"ip_addr"`
	CreatedAt time.Time `json:"created_at"`
}

// SessionCache provides Redis operations for session management
type SessionCache struct {
	client *redis.Client
}

// NewSessionCache creates a new session cache instance
func NewSessionCache(client *redis.Client) *SessionCache {
	return &SessionCache{
		client: client,
	}
}

// StoreSession saves a session to Redis with TTL matching JWT expiration
func (sc *SessionCache) StoreSession(ctx context.Context, sessionID uuid.UUID, data SessionData, ttl time.Duration) error {
	key := fmt.Sprintf("session:%s", sessionID.String())

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal session data: %w", err)
	}

	err = sc.client.Set(ctx, key, jsonData, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to store session in Redis: %w", err)
	}

	return nil
}

// GetSession retrieves a session from Redis
func (sc *SessionCache) GetSession(ctx context.Context, sessionID uuid.UUID) (*SessionData, error) {
	key := fmt.Sprintf("session:%s", sessionID.String())

	val, err := sc.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("session not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get session from Redis: %w", err)
	}

	var data SessionData
	if err := json.Unmarshal([]byte(val), &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session data: %w", err)
	}

	return &data, nil
}

// DeleteSession removes a session from Redis
func (sc *SessionCache) DeleteSession(ctx context.Context, sessionID uuid.UUID) error {
	key := fmt.Sprintf("session:%s", sessionID.String())

	err := sc.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete session from Redis: %w", err)
	}

	return nil
}

// DeleteAllUserSessions removes all sessions for a specific user
// This requires maintaining a user-to-sessions mapping
func (sc *SessionCache) DeleteAllUserSessions(ctx context.Context, sessionIDs []uuid.UUID) error {
	if len(sessionIDs) == 0 {
		return nil
	}

	keys := make([]string, len(sessionIDs))
	for i, id := range sessionIDs {
		keys[i] = fmt.Sprintf("session:%s", id.String())
	}

	err := sc.client.Del(ctx, keys...).Err()
	if err != nil {
		return fmt.Errorf("failed to delete user sessions from Redis: %w", err)
	}

	return nil
}

// SessionExists checks if a session exists in Redis (returns true/false)
func (sc *SessionCache) SessionExists(ctx context.Context, sessionID uuid.UUID) (bool, error) {
	key := fmt.Sprintf("session:%s", sessionID.String())

	exists, err := sc.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check session existence: %w", err)
	}

	return exists > 0, nil
}
