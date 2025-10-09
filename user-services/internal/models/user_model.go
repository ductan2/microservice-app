package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// User represents the canonical identity
type User struct {
	ID                      uuid.UUID    `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	Email                   string       `gorm:"type:citext;uniqueIndex;not null" json:"email"`
	EmailNormalized         string       `gorm:"type:citext;index:users_email_norm_idx" json:"-"`
	PasswordHash            string       `gorm:"type:text;not null" json:"-"`
	EmailVerified           bool         `gorm:"default:false;not null" json:"email_verified"`
	EmailVerificationToken  string       `gorm:"type:text" json:"-"`
	EmailVerificationExpiry sql.NullTime `gorm:"type:timestamptz" json:"-"`
	Status                  string       `gorm:"type:text;default:'active';not null;check:status IN ('active','locked','disabled','deleted')" json:"status"`
	CreatedAt               time.Time    `json:"created_at"`
	UpdatedAt               time.Time    `json:"updated_at"`
	Profile                 UserProfile  `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;foreignKey:UserID;references:ID" json:"profile"`
	DeletedAt               sql.NullTime `gorm:"type:timestamptz" json:"deleted_at,omitempty"`
	LastLoginAt             sql.NullTime `gorm:"type:timestamptz" json:"last_login_at,omitempty"`
	LastLoginIP             *string      `gorm:"type:inet" json:"last_login_ip,omitempty"`
	LockoutUntil            sql.NullTime `gorm:"type:timestamptz" json:"lockout_until,omitempty"`
}

// UserProfile stores non-auth PII
type UserProfile struct {
	UserID      uuid.UUID `gorm:"type:uuid;primaryKey" json:"user_id"`
	DisplayName string    `gorm:"type:text" json:"display_name,omitempty"`
	AvatarURL   string    `gorm:"type:text" json:"avatar_url,omitempty"`
	Locale      string    `gorm:"type:text;default:'en'" json:"locale"`
	TimeZone    string    `gorm:"type:text;default:'UTC'" json:"time_zone"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Session represents server-side JWT tracking
type Session struct {
	ID        uuid.UUID    `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	UserID    uuid.UUID    `gorm:"type:uuid;not null;index:sessions_user_expires_idx;constraint:OnDelete:CASCADE" json:"user_id"`
	UserAgent string       `gorm:"type:text" json:"user_agent,omitempty"`
	IPAddr    string       `gorm:"type:inet" json:"ip_addr,omitempty"`
	CreatedAt time.Time    `json:"created_at"`
	ExpiresAt time.Time    `gorm:"not null;index:sessions_user_expires_idx" json:"expires_at"`
	RevokedAt sql.NullTime `json:"revoked_at,omitempty"`
}

// RefreshToken implements rotating refresh tokens
type RefreshToken struct {
	ID         uuid.UUID    `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	SessionID  uuid.UUID    `gorm:"type:uuid;not null;index:refresh_tokens_session_idx;constraint:OnDelete:CASCADE" json:"session_id"`
	TokenHash  string       `gorm:"type:text;uniqueIndex;not null" json:"-"`
	IssuedAt   time.Time    `gorm:"default:now();not null" json:"issued_at"`
	ExpiresAt  time.Time    `gorm:"not null;index:refresh_tokens_session_idx" json:"expires_at"`
	ConsumedAt sql.NullTime `json:"consumed_at,omitempty"`
	RevokedAt  sql.NullTime `json:"revoked_at,omitempty"`
}

// MFAMethod supports TOTP and WebAuthn
type MFAMethod struct {
	ID          uuid.UUID    `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	UserID      uuid.UUID    `gorm:"type:uuid;not null;constraint:OnDelete:CASCADE" json:"user_id"`
	Type        string       `gorm:"type:text;not null;check:type IN ('totp','webauthn')" json:"type"`
	Label       string       `gorm:"type:text" json:"label,omitempty"`
	Secret      string       `gorm:"type:text" json:"-"` // encrypted at rest
	WebAuthnPub string       `gorm:"type:text" json:"-"`
	AddedAt     time.Time    `gorm:"default:now();not null" json:"added_at"`
	LastUsedAt  sql.NullTime `json:"last_used_at,omitempty"`
}

// LoginAttempt tracks login attempts for throttling
type LoginAttempt struct {
	ID        int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    *uuid.UUID `gorm:"type:uuid;index:login_attempts_user_time_idx" json:"user_id,omitempty"`
	Email     string     `gorm:"type:citext" json:"email,omitempty"`
	IPAddr    string     `gorm:"type:inet" json:"ip_addr,omitempty"`
	Success   bool       `gorm:"not null" json:"success"`
	Reason    string     `gorm:"type:text" json:"reason,omitempty"`
	CreatedAt time.Time  `gorm:"default:now();not null;index:login_attempts_user_time_idx" json:"created_at"`
}

// PasswordReset manages password reset flow
type PasswordReset struct {
	ID         uuid.UUID    `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	UserID     uuid.UUID    `gorm:"type:uuid;not null;constraint:OnDelete:CASCADE" json:"user_id"`
	TokenHash  string       `gorm:"type:text;uniqueIndex;not null" json:"-"`
	ExpiresAt  time.Time    `gorm:"not null" json:"expires_at"`
	ConsumedAt sql.NullTime `json:"consumed_at,omitempty"`
}

// Role for RBAC
type Role struct {
	ID   uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	Name string    `gorm:"type:text;uniqueIndex;not null" json:"name"`
}

// UserRole junction table
type UserRole struct {
	UserID uuid.UUID `gorm:"type:uuid;primaryKey;constraint:OnDelete:CASCADE" json:"user_id"`
	RoleID uuid.UUID `gorm:"type:uuid;primaryKey;constraint:OnDelete:CASCADE" json:"role_id"`
}

// AuditLog append-only audit trail
type AuditLog struct {
	ID        int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    *uuid.UUID     `gorm:"type:uuid;index:audit_logs_user_time_idx" json:"user_id,omitempty"`
	ActorID   *uuid.UUID     `gorm:"type:uuid" json:"actor_id,omitempty"`
	Action    string         `gorm:"type:text;not null" json:"action"`
	IPAddr    string         `gorm:"type:inet" json:"ip_addr,omitempty"`
	Metadata  map[string]any `gorm:"type:jsonb;default:'{}';not null" json:"metadata"`
	CreatedAt time.Time      `gorm:"default:now();not null;index:audit_logs_user_time_idx" json:"created_at"`
}

// Outbox for cross-service events (transactional outbox pattern)
type Outbox struct {
	ID          int64        `gorm:"primaryKey;autoIncrement" json:"id"`
	AggregateID uuid.UUID    `gorm:"type:uuid;not null" json:"aggregate_id"`
	Topic       string       `gorm:"type:text;not null" json:"topic"`
	Type        string       `gorm:"type:text;not null" json:"type"`
	Payload     []byte       `gorm:"type:jsonb"`
	CreatedAt   time.Time    `gorm:"default:now();not null" json:"created_at"`
	PublishedAt sql.NullTime `gorm:"index:outbox_unpublished_idx,where:published_at IS NULL" json:"published_at,omitempty"`
}
