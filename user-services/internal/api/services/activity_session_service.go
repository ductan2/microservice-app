package services

import (
	"context"
	"errors"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"user-services/internal/api/dto"
	"user-services/internal/api/repositories"
	"user-services/internal/models"
)

// ActivitySessionService handles user activity session business logic
type ActivitySessionService interface {
	StartActivitySession(ctx context.Context, req *dto.StartSessionRequest, userID uuid.UUID, ginCtx *gin.Context) (*dto.ActivitySessionResponse, error)
	EndActivitySession(ctx context.Context, req *dto.EndSessionRequest, userID uuid.UUID) (*dto.ActivitySessionResponse, error)
	GetActivitySessions(ctx context.Context, userID uuid.UUID, page int, limit int, startDate *time.Time, endDate *time.Time) ([]dto.ActivitySessionResponse, int64, error)
	GetSessionStats(ctx context.Context, userID uuid.UUID) (*dto.SessionStatsResponse, error)
	UpdateActiveSession(ctx context.Context, sessionID uuid.UUID, userAgent *string, ipAddr *string, userID uuid.UUID, ginCtx *gin.Context) error
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

// getClientIP extracts the real client IP address from the request
func getClientIP(ginCtx *gin.Context) string {
	// Check X-Forwarded-For header first (for reverse proxies)
	if xff := ginCtx.GetHeader("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		for i, char := range xff {
			if char == ',' {
				return xff[:i]
			}
		}
		return xff
	}

	// Check X-Real-IP header
	if xri := ginCtx.GetHeader("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	return ginCtx.ClientIP()
}

// StartActivitySession creates a new activity session
func (s *activitySessionService) StartActivitySession(ctx context.Context, req *dto.StartSessionRequest, userID uuid.UUID, ginCtx *gin.Context) (*dto.ActivitySessionResponse, error) {
	// Check if session already exists
	existing, err := s.repo.GetBySessionID(ctx, req.SessionID)
	if err != nil {
		return nil, err
	}

	if existing != nil {
		return mapToActivitySessionResponse(existing), nil
	}

	if req.SessionID == uuid.Nil {
		return nil, errors.New("invalid session_id")
	}

	// Get IP address from backend request
	clientIP := getClientIP(ginCtx)

	// Get user agent from request header, fallback to request body if provided
	userAgent := ginCtx.GetHeader("User-Agent")
	if req.UserAgent != nil && *req.UserAgent != "" {
		userAgent = *req.UserAgent
	}

	// Create new activity session
	session := &models.UserActivitySession{
		UserID:    userID,
		SessionID: req.SessionID,
		StartedAt: time.Now().UTC(),
		IPAddr:    clientIP,
		UserAgent: userAgent,
	}

	err = s.repo.Create(ctx, session)
	if err != nil {
		return nil, err
	}

	return mapToActivitySessionResponse(session), nil
}

// EndActivitySession ends an active session and calculates duration
func (s *activitySessionService) EndActivitySession(ctx context.Context, req *dto.EndSessionRequest, userID uuid.UUID) (*dto.ActivitySessionResponse, error) {
	// Get the active session
	session, err := s.repo.GetActiveSession(ctx, req.SessionID)
	if err != nil {
		return nil, err
	}

	// Verify ownership
	if session.UserID != userID {
		return nil, errors.New("unauthorized to access this session")
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
func (s *activitySessionService) GetActivitySessions(ctx context.Context, userID uuid.UUID, page int, limit int, startDate *time.Time, endDate *time.Time) ([]dto.ActivitySessionResponse, int64, error) {
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
func (s *activitySessionService) GetSessionStats(ctx context.Context, userID uuid.UUID) (*dto.SessionStatsResponse, error) {
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
func (s *activitySessionService) UpdateActiveSession(ctx context.Context, sessionID uuid.UUID, userAgent *string, ipAddr *string, userID uuid.UUID, ginCtx *gin.Context) error {
	// Check if active session exists and belongs to user
	session, err := s.repo.GetActiveSession(ctx, sessionID)
	if err != nil {
		return err
	}

	if session == nil {
		return errors.New("active session not found")
	}

	if session.UserID != userID {
		return errors.New("unauthorized to access this session")
	}

	// Get IP address from backend if not provided
	var finalIPAddr *string
	if ipAddr != nil && *ipAddr != "" {
		finalIPAddr = ipAddr
	} else {
		clientIP := getClientIP(ginCtx)
		finalIPAddr = &clientIP
	}

	// Get user agent from request header if not provided
	var finalUserAgent *string
	if userAgent != nil && *userAgent != "" {
		finalUserAgent = userAgent
	} else {
		headerUA := ginCtx.GetHeader("User-Agent")
		finalUserAgent = &headerUA
	}

	return s.repo.UpdateSessionInfo(ctx, sessionID, finalUserAgent, finalIPAddr)
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
		IPAddr:     &session.IPAddr,
		UserAgent:  &session.UserAgent,
		CreatedAt:  session.CreatedAt,
		UpdatedAt:  session.UpdatedAt,
	}
}