package services

import (
	"context"
	"user-services/internal/api/dto"
	"user-services/internal/api/repositories"

	"github.com/google/uuid"
)

type SessionService interface {
	CreateSession(ctx context.Context, userID uuid.UUID, userAgent, ipAddr string) (*dto.SessionResponse, error)
	GetUserSessions(ctx context.Context, userID uuid.UUID) ([]dto.SessionResponse, error)
	RevokeSession(ctx context.Context, sessionID uuid.UUID) error
	RevokeAllUserSessions(ctx context.Context, userID uuid.UUID) error
	CleanupExpiredSessions(ctx context.Context) error
}

type sessionService struct {
	sessionRepo repositories.SessionRepository
}

func NewSessionService(sessionRepo repositories.SessionRepository) SessionService {
	return &sessionService{
		sessionRepo: sessionRepo,
	}
}

func (s *sessionService) CreateSession(ctx context.Context, userID uuid.UUID, userAgent, ipAddr string) (*dto.SessionResponse, error) {
	// TODO: implement
	return nil, nil
}

func (s *sessionService) GetUserSessions(ctx context.Context, userID uuid.UUID) ([]dto.SessionResponse, error) {
	// TODO: implement
	return nil, nil
}

func (s *sessionService) RevokeSession(ctx context.Context, sessionID uuid.UUID) error {
	// TODO: implement
	return nil
}

func (s *sessionService) RevokeAllUserSessions(ctx context.Context, userID uuid.UUID) error {
	// TODO: implement
	return nil
}

func (s *sessionService) CleanupExpiredSessions(ctx context.Context) error {
	// TODO: implement
	return nil
}
