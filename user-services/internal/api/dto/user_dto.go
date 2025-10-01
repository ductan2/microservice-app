package dto

import (
	"time"

	"github.com/google/uuid"
)

// PublicUser exposed via API (no sensitive data)
type PublicUser struct {
	ID            uuid.UUID    `json:"id"`
	Email         string       `json:"email"`
	EmailVerified bool         `json:"email_verified"`
	Status        string       `json:"status"`
	Profile       *UserProfile `json:"profile,omitempty"`
	Roles         []string     `json:"roles,omitempty"`
	CreatedAt     time.Time    `json:"created_at"`
	UpdatedAt     time.Time    `json:"updated_at"`
}

// UserProfile for non-auth data
type UserProfile struct {
	DisplayName string `json:"display_name,omitempty"`
	AvatarURL   string `json:"avatar_url,omitempty"`
	Locale      string `json:"locale"`
	TimeZone    string `json:"time_zone"`
}

// UpdateProfileRequest updates user profile
type UpdateProfileRequest struct {
	DisplayName string `json:"display_name,omitempty" binding:"omitempty,min=2"`
	AvatarURL   string `json:"avatar_url,omitempty" binding:"omitempty,url"`
	Locale      string `json:"locale,omitempty" binding:"omitempty,len=2"`
	TimeZone    string `json:"time_zone,omitempty"`
}

// UpdateUserStatusRequest for admin actions
type UpdateUserStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=active locked disabled deleted"`
	Reason string `json:"reason,omitempty"`
}

// AssignRoleRequest assigns role to user
type AssignRoleRequest struct {
	RoleName string `json:"role_name" binding:"required"`
}

// RemoveRoleRequest removes role from user
type RemoveRoleRequest struct {
	RoleName string `json:"role_name" binding:"required"`
}

// CreateRoleRequest creates new role
type CreateRoleRequest struct {
	Name string `json:"name" binding:"required"`
}

// RoleResponse represents a role
type RoleResponse struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

// AuditLogResponse for audit trail
type AuditLogResponse struct {
	ID        int64          `json:"id"`
	UserID    *uuid.UUID     `json:"user_id,omitempty"`
	ActorID   *uuid.UUID     `json:"actor_id,omitempty"`
	Action    string         `json:"action"`
	IPAddr    string         `json:"ip_addr,omitempty"`
	Metadata  map[string]any `json:"metadata"`
	CreatedAt time.Time      `json:"created_at"`
}

// ListUsersRequest for pagination and filtering
type ListUsersRequest struct {
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	Status   string `form:"status" binding:"omitempty,oneof=active locked disabled deleted"`
	Search   string `form:"search" binding:"omitempty"`
}

// PaginatedResponse generic pagination wrapper
type PaginatedResponse struct {
	Data       any `json:"data"`
	Page       int `json:"page"`
	PageSize   int `json:"page_size"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}
