package dto

import (
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

// StartSessionRequest represents the request to start a user activity session
type StartSessionRequest struct {
	SessionID  uuid.UUID `json:"session_id" binding:"required"`
	IPAddr     *string   `json:"ip_addr,omitempty"`
	UserAgent  *string   `json:"user_agent,omitempty"`
}

// EndSessionRequest represents the request to end a user activity session
type EndSessionRequest struct {
	SessionID   uuid.UUID `json:"session_id" binding:"required"`
	Reason      *string   `json:"reason,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

// ActivitySessionResponse represents a user activity session response
type ActivitySessionResponse struct {
	ID         uuid.UUID     `json:"id"`
	UserID     uuid.UUID     `json:"user_id"`
	SessionID  uuid.UUID     `json:"session_id"`
	StartedAt  time.Time     `json:"started_at"`
	EndedAt    *time.Time    `json:"ended_at,omitempty"`
	DurationMs int64         `json:"duration_ms"`
	IPAddr     *string       `json:"ip_addr,omitempty"`
	UserAgent  *string       `json:"user_agent,omitempty"`
	CreatedAt  time.Time     `json:"created_at"`
	UpdatedAt  time.Time     `json:"updated_at"`
}

// SessionStatsResponse represents aggregated session statistics
type SessionStatsResponse struct {
	TotalSessions      int64 `json:"total_sessions"`
	TotalDurationMs    int64 `json:"total_duration_ms"`
	AverageDurationMs int64 `json:"average_duration_ms"`
	LongestDurationMs  int64 `json:"longest_duration_ms"`
	ShortestDurationMs int64 `json:"shortest_duration_ms"`
}

// UpdateSessionRequest represents the request to update an active session
type UpdateSessionRequest struct {
	SessionID  uuid.UUID `json:"session_id" binding:"required"`
	UserAgent  *string   `json:"user_agent,omitempty"`
	IPAddr     *string   `json:"ip_addr,omitempty"`
}

// SessionHistoryRequest represents pagination for session history
type SessionHistoryRequest struct {
	Page      int      `form:"page" binding:"min=1"`
	Limit     int      `form:"limit" binding:"min=1,max=100"`
	StartDate *time.Time `form:"start_date"`
	EndDate   *time.Time `form:"end_date"`
}

// Validate validates the StartSessionRequest
func (r *StartSessionRequest) Validate() error {
	if r.SessionID == uuid.Nil {
		return &ValidationError{
			Field:   "session_id",
			Message: "session_id is required",
		}
	}
	return nil
}

// Validate validates the EndSessionRequest
func (r *EndSessionRequest) Validate() error {
	if r.SessionID == uuid.Nil {
		return &ValidationError{
			Field:   "session_id",
			Message: "session_id is required",
		}
	}
	return nil
}

// Validate validates the UpdateSessionRequest
func (r *UpdateSessionRequest) Validate() error {
	if r.SessionID == uuid.Nil {
		return &ValidationError{
			Field:   "session_id",
			Message: "session_id is required",
		}
	}
	if r.IPAddr != nil {
		if len(*r.IPAddr) > 45 { // Max length for IP addresses
			return &ValidationError{
				Field:   "ip_addr",
				Message: "IP address too long",
			}
		}
	}
	if r.UserAgent != nil && utf8.RuneCountInString(*r.UserAgent) > 500 {
		return &ValidationError{
			Field:   "user_agent",
			Message: "user agent too long",
		}
	}
	return nil
}

// Validate validates the SessionHistoryRequest
func (r *SessionHistoryRequest) Validate() error {
	if r.Page <= 0 {
		r.Page = 1
	}
	if r.Limit <= 0 || r.Limit > 100 {
		r.Limit = 20
	}
	return nil
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Message
}