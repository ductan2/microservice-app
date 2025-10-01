package services

import (
	"context"
	"user-services/internal/api/repositories"

	"github.com/google/uuid"
)

type ThrottleService interface {
	RecordLoginAttempt(ctx context.Context, userID *uuid.UUID, email, ipAddr string, success bool, reason string) error
	IsThrottled(ctx context.Context, email, ipAddr string) (bool, error)
	GetRecentAttempts(ctx context.Context, userID uuid.UUID) (int, error)
	CleanupOldAttempts(ctx context.Context) error
}

type throttleService struct {
	loginAttemptRepo repositories.LoginAttemptRepository
}

func NewThrottleService(loginAttemptRepo repositories.LoginAttemptRepository) ThrottleService {
	return &throttleService{
		loginAttemptRepo: loginAttemptRepo,
	}
}

func (s *throttleService) RecordLoginAttempt(ctx context.Context, userID *uuid.UUID, email, ipAddr string, success bool, reason string) error {
	// TODO: implement
	return nil
}

func (s *throttleService) IsThrottled(ctx context.Context, email, ipAddr string) (bool, error) {
	// TODO: implement - check failed attempts in last N minutes
	return false, nil
}

func (s *throttleService) GetRecentAttempts(ctx context.Context, userID uuid.UUID) (int, error) {
	// TODO: implement
	return 0, nil
}

func (s *throttleService) CleanupOldAttempts(ctx context.Context) error {
	// TODO: implement
	return nil
}
