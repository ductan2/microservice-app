package services

import (
	"context"
	"user-services/internal/api/repositories"

	"github.com/google/uuid"
)

type PasswordService interface {
	InitiatePasswordReset(ctx context.Context, email string) error
	VerifyResetToken(ctx context.Context, token string) (*uuid.UUID, error)
	CompletePasswordReset(ctx context.Context, token, newPassword string) error
	ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error
	CleanupExpiredResets(ctx context.Context) error
}

type passwordService struct {
	userRepo          repositories.UserRepository
	passwordResetRepo repositories.PasswordResetRepository
	auditLogRepo      repositories.AuditLogRepository
}

func NewPasswordService(
	userRepo repositories.UserRepository,
	passwordResetRepo repositories.PasswordResetRepository,
	auditLogRepo repositories.AuditLogRepository,
) PasswordService {
	return &passwordService{
		userRepo:          userRepo,
		passwordResetRepo: passwordResetRepo,
		auditLogRepo:      auditLogRepo,
	}
}

func (s *passwordService) InitiatePasswordReset(ctx context.Context, email string) error {
	// TODO: implement - generate token, save to DB, send email
	return nil
}

func (s *passwordService) VerifyResetToken(ctx context.Context, token string) (*uuid.UUID, error) {
	// TODO: implement - verify token and return user ID
	return nil, nil
}

func (s *passwordService) CompletePasswordReset(ctx context.Context, token, newPassword string) error {
	// TODO: implement - verify token, update password, consume token
	return nil
}

func (s *passwordService) ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error {
	// TODO: implement - verify old password, update to new
	return nil
}

func (s *passwordService) CleanupExpiredResets(ctx context.Context) error {
	// TODO: implement
	return nil
}
