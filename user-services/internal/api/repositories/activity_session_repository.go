package repositories

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/google/uuid"
	"user-services/internal/models"
)

// ActivitySessionRepository interface for user activity session data access
type ActivitySessionRepository interface {
	Create(ctx context.Context, session *models.UserActivitySession) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.UserActivitySession, error)
	GetBySessionID(ctx context.Context, sessionID uuid.UUID) (*models.UserActivitySession, error)
	GetActiveSession(ctx context.Context, sessionID uuid.UUID) (*models.UserActivitySession, error)
	GetByUserID(ctx context.Context, userID uuid.UUID, limit int, offset int) ([]models.UserActivitySession, int64, error)
	GetSessionStats(ctx context.Context, userID uuid.UUID) (*SessionStats, error)
	UpdateEndTime(ctx context.Context, id uuid.UUID, endTime time.Time, durationMs int64) error
	UpdateSessionInfo(ctx context.Context, sessionID uuid.UUID, userAgent *string, ipAddr *string) error
	DeleteSession(ctx context.Context, id uuid.UUID) error
	DeleteSessionBySessionID(ctx context.Context, sessionID uuid.UUID) error
}

// SessionStats represents aggregated session statistics
type SessionStats struct {
	TotalSessions      int64 `json:"total_sessions"`
	TotalDurationMs    int64 `json:"total_duration_ms"`
	AverageDurationMs int64 `json:"average_duration_ms"`
	LongestDurationMs  int64 `json:"longest_duration_ms"`
	ShortestDurationMs int64 `json:"shortest_duration_ms"`
}

// activitySessionRepository implements ActivitySessionRepository
type activitySessionRepository struct {
	db *gorm.DB
}

// NewActivitySessionRepository creates a new activity session repository
func NewActivitySessionRepository(db *gorm.DB) ActivitySessionRepository {
	return &activitySessionRepository{db: db}
}

// Create creates a new user activity session
func (r *activitySessionRepository) Create(ctx context.Context, session *models.UserActivitySession) error {
	return r.db.WithContext(ctx).Create(session).Error
}

// GetByID retrieves an activity session by ID
func (r *activitySessionRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.UserActivitySession, error) {
	var session models.UserActivitySession
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&session).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &session, nil
}

// GetBySessionID retrieves an activity session by session ID
func (r *activitySessionRepository) GetBySessionID(ctx context.Context, sessionID uuid.UUID) (*models.UserActivitySession, error) {
	var session models.UserActivitySession
	err := r.db.WithContext(ctx).Where("session_id = ?", sessionID).First(&session).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &session, nil
}

// GetActiveSession retrieves an active (not ended) session by session ID
func (r *activitySessionRepository) GetActiveSession(ctx context.Context, sessionID uuid.UUID) (*models.UserActivitySession, error) {
	var session models.UserActivitySession
	err := r.db.WithContext(ctx).Where("session_id = ? AND ended_at IS NULL", sessionID).First(&session).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &session, nil
}

// GetByUserID retrieves activity sessions for a user with pagination
func (r *activitySessionRepository) GetByUserID(ctx context.Context, userID uuid.UUID, limit int, offset int) ([]models.UserActivitySession, int64, error) {
	var sessions []models.UserActivitySession
	var total int64

	err := r.db.WithContext(ctx).Model(&models.UserActivitySession{}).
		Where("user_id = ?", userID).
		Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&sessions).Error

	return sessions, total, err
}

// GetSessionStats retrieves aggregated session statistics for a user
func (r *activitySessionRepository) GetSessionStats(ctx context.Context, userID uuid.UUID) (*SessionStats, error) {
	var result struct {
		TotalSessions      sql.NullInt64 `json:"total_sessions"`
		TotalDurationMs    sql.NullInt64 `json:"total_duration_ms"`
		LongestDurationMs  sql.NullInt64 `json:"longest_duration_ms"`
		ShortestDurationMs sql.NullInt64 `json:"shortest_duration_ms"`
	}

	// Query for completed sessions only
	err := r.db.WithContext(ctx).Raw(`
		SELECT
			COUNT(*) as total_sessions,
			COALESCE(SUM(duration_ms), 0) as total_duration_ms,
			MAX(duration_ms) as longest_duration_ms,
			MIN(duration_ms) as shortest_duration_ms
		FROM user_activity_sessions
		WHERE user_id = ? AND duration_ms > 0
	`, userID).Scan(&result).Error

	if err != nil {
		return nil, err
	}

	totalSessions := result.TotalSessions.Int64
	totalDuration := result.TotalDurationMs.Int64

	stats := &SessionStats{
		TotalSessions:   totalSessions,
		TotalDurationMs: totalDuration,
	}

	// Calculate average if there are completed sessions
	if totalSessions > 0 {
		stats.AverageDurationMs = totalDuration / totalSessions
		stats.LongestDurationMs = result.LongestDurationMs.Int64
		stats.ShortestDurationMs = result.ShortestDurationMs.Int64
	}

	return stats, nil
}

// UpdateEndTime updates the end time and duration for an activity session
func (r *activitySessionRepository) UpdateEndTime(ctx context.Context, id uuid.UUID, endTime time.Time, durationMs int64) error {
	session := models.UserActivitySession{
		EndedAt:    sql.NullTime{Time: endTime, Valid: true},
		DurationMs: durationMs,
	}

	return r.db.WithContext(ctx).Model(&models.UserActivitySession{}).
		Where("id = ?", id).
		Updates(session).Error
}

// UpdateSessionInfo updates user agent and IP address for an active session
func (r *activitySessionRepository) UpdateSessionInfo(ctx context.Context, sessionID uuid.UUID, userAgent *string, ipAddr *string) error {
	updates := map[string]interface{}{}
	if userAgent != nil {
		updates["user_agent"] = *userAgent
	}
	if ipAddr != nil {
		updates["ip_addr"] = *ipAddr
	}

	return r.db.WithContext(ctx).
		Model(&models.UserActivitySession{}).
		Where("session_id = ? AND ended_at IS NULL", sessionID).
		Updates(updates).Error
}

// DeleteSession deletes an activity session by ID
func (r *activitySessionRepository) DeleteSession(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.UserActivitySession{}, id).Error
}

// DeleteSessionBySessionID deletes an activity session by session ID
func (r *activitySessionRepository) DeleteSessionBySessionID(ctx context.Context, sessionID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("session_id = ?", sessionID).Delete(&models.UserActivitySession{}).Error
}

// GetActiveSessionsByUser retrieves all active sessions for a user
func (r *activitySessionRepository) GetActiveSessionsByUser(ctx context.Context, userID uuid.UUID) ([]models.UserActivitySession, error) {
	var sessions []models.UserActivitySession
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND ended_at IS NULL", userID).
		Find(&sessions).Error
	return sessions, err
}

// CleanupOldSessions removes activity sessions older than the specified duration
func (r *activitySessionRepository) CleanupOldSessions(ctx context.Context, olderThan time.Time) (int64, error) {
	result := r.db.WithContext(ctx).
		Where("created_at < ?", olderThan).
		Delete(&models.UserActivitySession{})
	return result.RowsAffected, result.Error
}