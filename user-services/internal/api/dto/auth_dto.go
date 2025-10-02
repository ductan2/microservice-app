package dto

import (
	"time"

	"github.com/google/uuid"
)

// RegisterRequest represents user registration
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Name     string `json:"name" binding:"required"`
	Password string `json:"password" binding:"required,min=8"`
}

// LoginRequest represents login attempt
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
	MFACode  string `json:"mfa_code,omitempty"`
}

// AuthResponse after successful authentication
type AuthResponse struct {
	AccessToken  string     `json:"access_token"`
	RefreshToken string     `json:"refresh_token"`
	ExpiresAt    time.Time  `json:"expires_at"`
	User         PublicUser `json:"user"`
	MFARequired  bool       `json:"mfa_required,omitempty"`
}

// RefreshTokenRequest to rotate tokens
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// RefreshTokenResponse with new tokens
type RefreshTokenResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// LogoutRequest to revoke session
type LogoutRequest struct {
	SessionID string `json:"session_id,omitempty"`
}

// PasswordResetRequestDTO initiates password reset
type PasswordResetRequestDTO struct {
	Email string `json:"email" binding:"required,email"`
}

// PasswordResetConfirmDTO completes password reset
type PasswordResetConfirmDTO struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

// ChangePasswordRequest for authenticated users
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

// VerifyEmailRequest to verify email address
type VerifyEmailRequest struct {
	Token string `json:"token" binding:"required"`
}

// SessionResponse represents active session
type SessionResponse struct {
	ID        uuid.UUID `json:"id"`
	UserAgent string    `json:"user_agent,omitempty"`
	IPAddr    string    `json:"ip_addr,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
	IsCurrent bool      `json:"is_current"`
}
