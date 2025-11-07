package services

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"user-services/internal/api/repositories"
	"user-services/internal/api/dto"
)

// ActivitySessionService handles user activity session business logic
type ActivitySessionService interface {
	StartActivitySession(ctx context.Context, req *dto.StartSessionRequest, userID uuid.UUID) (*dto.ActivitySessionResponse, error)
	EndActivitySession(ctx context.Context, req *dto.EndSessionRequest, userID uuid.UUID) (*dto.ActivitySessionResponse, error)
	GetActivitySessions(ctx context.Context, userID uuid.UUID, page int, limit int, startDate *time.Time, endDate *time.Time) ([]dto.ActivitySessionResponse, int64, error)
	GetSessionStats(ctx context.Context, userID uuid.UUID) (*dto.SessionStatsResponse, error)
	UpdateActiveSession(ctx context.Context, sessionID uuid.UUID, userAgent *string, ipAddr *string, userID uuid.UUID) error
}

// activitySessionService implements ActivitySessionService
type activitySessionService struct {
	repo repositories.ActivitySessionRepository
	db   *gorm.DB
}

// NewActivitySessionService creates a new activity session service
func NewActivitySessionService(repo repositories.ActivitySessionRepository, db *gorm.DB) ActivitySessionService {
	return &activitySessionService{
		repo: repo,
		db:   db,
	}
}

// StartActivitySession creates a new activity session
func (s *activitySessionService) StartActivitySession(ctx context.Context, req *StartSessionRequest, userID uuid.UUID) (*UserActivitySessionResponse, error) {
	// Check if session already exists
	existing, err := s.repo.GetBySessionID(ctx, req.SessionID)
	if err != nil {
		return nil, err
	}

	if existing != nil {
		// Session already exists, return existing session
		return mapToActivitySessionResponse(existing), nil
	}

	// Validate that the session exists in the system
	// This is a simple validation - in production you might want to check against the sessions table
	if req.SessionID == uuid.Nil {
		return nil, &ValidationError{
			Field:   "session_id",
			Message: "invalid session_id",
		}
	}

	// Create new activity session
	session := &models.UserActivitySession{
		UserID:    userID,
		SessionID: req.SessionID,
		StartedAt: time.Now().UTC(),
		IPAddr:    dereferenceString(req.IPAddr),
		UserAgent: dereferenceString(req.UserAgent),
	}

	err = s.repo.Create(ctx, session)
	if err != nil {
		return nil, err
	}

	return mapToActivitySessionResponse(session), nil
}

// EndActivitySession ends an active session and calculates duration
func (s *activitySessionService) EndActivitySession(ctx context.Context, req *EndSessionRequest, userID uuid.UUID) (*UserActivitySessionResponse, error) {
	// Get the active session
	session, err := s.repo.GetActiveSession(ctx, req.SessionID)
	if err != nil {
		return nil, err
	}

	if session == nil {
		return nil, &SessionError{
			Code:    "SESSION_NOT_FOUND",
			Message: "Active session not found",
		}
	}

	// Verify ownership
	if session.UserID != userID {
		return nil, &SessionError{
			Code:    "UNAUTHORIZED",
			Message: "Unauthorized to access this session",
		}
	}

	// Calculate end time
	endTime := time.Now().UTC()
	if req.CompletedAt != nil {
		endTime = *req.CompletedAt
	}

	// Calculate duration
	duration := endTime.Sub(session.StartedAt).Milliseconds()

	// Update session end time and duration
	err = s.repo.UpdateEndTime(ctx, session.ID, endTime, duration)
	if err != nil {
		return nil, err
	}

	// Refresh session from database
	updatedSession, err := s.repo.GetByID(ctx, session.ID)
	if err != nil {
		return nil, err
	}

	return mapToActivitySessionResponse(updatedSession), nil
}

// GetActivitySessions retrieves user's activity sessions with pagination
func (s *activitySessionService) GetActivitySessions(ctx context.Context, userID uuid.UUID, page int, limit int, startDate *time.Time, endDate *time.Time) ([]UserActivitySessionResponse, int64, error) {
	offset := (page - 1) * limit

	sessions, total, err := s.repo.GetByUserID(ctx, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]dto.ActivitySessionResponse, len(sessions))
	for i, session := range sessions {
		responses[i] = *mapToActivitySessionResponse(&session)
	}

	return responses, total, nil
}

// GetSessionStats retrieves aggregated session statistics for a user
func (s *activitySessionService) GetSessionStats(ctx context.Context, userID uuid.UUID) (*SessionStatsResponse, error) {
	stats, err := s.repo.GetSessionStats(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &dto.SessionStatsResponse{
		TotalSessions:      stats.TotalSessions,
		TotalDurationMs:    stats.TotalDurationMs,
		AverageDurationMs: stats.AverageDurationMs,
		LongestDurationMs:  stats.LongestDurationMs,
		ShortestDurationMs: stats.ShortestDurationMs,
	}, nil
}

// UpdateActiveSession updates user agent and IP address for an active session
func (s *activitySessionService) UpdateActiveSession(ctx context.Context, sessionID uuid.UUID, userAgent *string, ipAddr *string, userID uuid.UUID) error {
	// Check if active session exists and belongs to user
	session, err := s.repo.GetActiveSession(ctx, sessionID)
	if err != nil {
		return err
	}

	if session == nil {
		return &SessionError{
			Code:    "SESSION_NOT_FOUND",
			Message: "Active session not found",
		}
	}

	if session.UserID != userID {
		return &SessionError{
			Code:    "UNAUTHORIZED",
			Message: "Unauthorized to access this session",
		}
	}

	return s.repo.UpdateSessionInfo(ctx, sessionID, userAgent, ipAddr)
}

// Helper functions

// mapToActivitySessionResponse converts a UserActivitySession to ActivitySessionResponse
func mapToActivitySessionResponse(session *models.UserActivitySession) *dto.ActivitySessionResponse {
	var endTime *time.Time
	if session.EndedAt.Valid {
		endTime = &session.EndedAt.Time
	}

	return &dto.ActivitySessionResponse{
		ID:         session.ID,
		UserID:     session.UserID,
		SessionID:  session.SessionID,
		StartedAt:  session.StartedAt,
		EndedAt:    endTime,
		DurationMs: session.DurationMs,
		IPAddr:     dereferenceString(session.IPAddr),
		UserAgent:  dereferenceString(session.UserAgent),
		CreatedAt:  session.CreatedAt,
		UpdatedAt:  session.UpdatedAt,
	}
}

// dereferenceString returns a string pointer if the string is not empty, otherwise nil
func dereferenceString(s *string) *string {
	if s != nil && *s != "" {
		return s
	}
	return nil
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Message
}

// SessionError represents a session-related error
type SessionError struct {
	Code    string
	Message string
}

func (e *SessionError) Error() string {
	return e.Message
}