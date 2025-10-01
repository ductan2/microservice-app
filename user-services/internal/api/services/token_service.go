package services

import (
	"context"
	"user-services/internal/api/dto"
	"user-services/internal/api/repositories"

	"github.com/google/uuid"
)

type TokenService interface {
	GenerateTokenPair(ctx context.Context, userID, sessionID uuid.UUID) (*dto.AuthResponse, error)
	RefreshAccessToken(ctx context.Context, refreshToken string) (*dto.RefreshTokenResponse, error)
	ValidateAccessToken(ctx context.Context, token string) (*uuid.UUID, error)
	RevokeRefreshToken(ctx context.Context, tokenHash string) error
	CleanupExpiredTokens(ctx context.Context) error
}

type tokenService struct {
	refreshTokenRepo repositories.RefreshTokenRepository
	sessionRepo      repositories.SessionRepository
}

func NewTokenService(
	refreshTokenRepo repositories.RefreshTokenRepository,
	sessionRepo repositories.SessionRepository,
) TokenService {
	return &tokenService{
		refreshTokenRepo: refreshTokenRepo,
		sessionRepo:      sessionRepo,
	}
}

func (s *tokenService) GenerateTokenPair(ctx context.Context, userID, sessionID uuid.UUID) (*dto.AuthResponse, error) {
	// TODO: implement JWT generation and refresh token creation
	return nil, nil
}

func (s *tokenService) RefreshAccessToken(ctx context.Context, refreshToken string) (*dto.RefreshTokenResponse, error) {
	// TODO: implement token rotation
	return nil, nil
}

func (s *tokenService) ValidateAccessToken(ctx context.Context, token string) (*uuid.UUID, error) {
	// TODO: implement JWT validation
	return nil, nil
}

func (s *tokenService) RevokeRefreshToken(ctx context.Context, tokenHash string) error {
	// TODO: implement
	return nil
}

func (s *tokenService) CleanupExpiredTokens(ctx context.Context) error {
	// TODO: implement
	return nil
}
