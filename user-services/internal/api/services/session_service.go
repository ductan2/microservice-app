package services

import (
	"context"
	"errors"
	"time"

	"user-services/internal/api/dto"
	"user-services/internal/api/repositories"
	"user-services/internal/cache"
	"user-services/internal/models"
	"user-services/internal/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrSessionNotFound   = errors.New("session not found")
	defaultSessionExpiry = 30 * 24 * time.Hour
)

type SessionService interface {
	CreateSession(ctx context.Context, userID uuid.UUID, userAgent, ipAddr string) (*dto.SessionResponse, error)
	GetUserSessions(ctx context.Context, userID uuid.UUID) ([]dto.SessionResponse, error)
	RevokeSession(ctx context.Context, sessionID uuid.UUID) error
	RevokeAllUserSessions(ctx context.Context, userID uuid.UUID) error
	CleanupExpiredSessions(ctx context.Context) error
}

type sessionService struct {
	sessionRepo  repositories.SessionRepository
	sessionCache *cache.SessionCache
}

func NewSessionService(sessionRepo repositories.SessionRepository, sessionCache *cache.SessionCache) SessionService {
	return &sessionService{
		sessionRepo:  sessionRepo,
		sessionCache: sessionCache,
	}
}

func (s *sessionService) CreateSession(ctx context.Context, userID uuid.UUID, userAgent, ipAddr string) (*dto.SessionResponse, error) {
	now := time.Now()
	session := &models.Session{
		UserID:    userID,
		UserAgent: userAgent,
		IPAddr:    utils.SanitizeIPAddress(ipAddr),
		CreatedAt: now,
		ExpiresAt: now.Add(defaultSessionExpiry),
	}

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, err
	}

	response := modelToSessionDTO(*session)
	response.IsCurrent = true

	return &response, nil
}

func (s *sessionService) GetUserSessions(ctx context.Context, userID uuid.UUID) ([]dto.SessionResponse, error) {
	sessions, err := s.sessionRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	responses := make([]dto.SessionResponse, 0, len(sessions))
	for _, session := range sessions {
		if session.RevokedAt.Valid {
			continue
		}
		if session.ExpiresAt.Before(now) {
			continue
		}

		responses = append(responses, modelToSessionDTO(session))
	}

	return responses, nil
}

func (s *sessionService) RevokeSession(ctx context.Context, sessionID uuid.UUID) error {
	// First, revoke in database
	if err := s.sessionRepo.Revoke(ctx, sessionID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrSessionNotFound
		}
		return err
	}

	// Then, delete from Redis
	if err := s.sessionCache.DeleteSession(ctx, sessionID); err != nil {
		// Log error but don't fail the operation - DB is source of truth
		// In production, you might want to use a proper logger
	}

	return nil
}

func (s *sessionService) RevokeAllUserSessions(ctx context.Context, userID uuid.UUID) error {
	// First, get all active sessions for the user
	sessions, err := s.sessionRepo.GetByUserID(ctx, userID)
	if err != nil {
		return err
	}

	// Collect session IDs
	sessionIDs := make([]uuid.UUID, 0, len(sessions))
	for _, session := range sessions {
		if !session.RevokedAt.Valid {
			sessionIDs = append(sessionIDs, session.ID)
		}
	}

	// Revoke in database
	if err := s.sessionRepo.RevokeAllByUserID(ctx, userID); err != nil {
		return err
	}

	// Delete from Redis
	if len(sessionIDs) > 0 {
		if err := s.sessionCache.DeleteAllUserSessions(ctx, sessionIDs); err != nil {
			// Log error but don't fail the operation - DB is source of truth
			// In production, you might want to use a proper logger
		}
	}

	return nil
}

func (s *sessionService) CleanupExpiredSessions(ctx context.Context) error {
	return s.sessionRepo.DeleteExpired(ctx)
}

func modelToSessionDTO(session models.Session) dto.SessionResponse {
	return dto.SessionResponse{
		ID:        session.ID,
		UserAgent: session.UserAgent,
		IPAddr:    session.IPAddr,
		CreatedAt: session.CreatedAt,
		ExpiresAt: session.ExpiresAt,
		IsCurrent: false,
	}
}
