package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"
	"user-services/internal/api/repositories"
	"user-services/internal/models"
	"user-services/internal/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PasswordService interface {
	InitiatePasswordReset(ctx context.Context, email string) error
	VerifyResetToken(ctx context.Context, token string) (*uuid.UUID, error)
	CompletePasswordReset(ctx context.Context, token, newPassword string) error
	ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error
	CleanupExpiredResets(ctx context.Context) error
}

type passwordService struct {
	userRepo          *repositories.UserRepository
	passwordResetRepo repositories.PasswordResetRepository
	auditLogRepo      repositories.AuditLogRepository
	outboxRepo        repositories.OutboxRepository
	userProfileRepo   repositories.UserProfileRepository
}

func NewPasswordService(
	userRepo *repositories.UserRepository,
	passwordResetRepo repositories.PasswordResetRepository,
	auditLogRepo repositories.AuditLogRepository,
	outboxRepo repositories.OutboxRepository,
	userProfileRepo repositories.UserProfileRepository,
) PasswordService {
	return &passwordService{
		userRepo:          userRepo,
		passwordResetRepo: passwordResetRepo,
		auditLogRepo:      auditLogRepo,
		outboxRepo:        outboxRepo,
		userProfileRepo:   userProfileRepo,
	}
}

func (s *passwordService) InitiatePasswordReset(ctx context.Context, email string) error {
	// 1. Find user by email
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Don't reveal if email exists - just return success
			return nil
		}
		return fmt.Errorf("failed to get user: %w", err)
	}

	// 2. Delete any existing password reset tokens for this user
	_ = s.passwordResetRepo.DeleteByUserID(ctx, user.ID)

	// 3. Generate secure random token (32 bytes = 64 hex chars)
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return fmt.Errorf("failed to generate token: %w", err)
	}
	token := hex.EncodeToString(tokenBytes)

	// 4. Hash the token for storage (use simple SHA256, not bcrypt)
	tokenHash := utils.HashToken(token)

	// 5. Create password reset record (expires in 1 hour)
	passwordReset := &models.PasswordReset{
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	if err := s.passwordResetRepo.Create(ctx, passwordReset); err != nil {
		return fmt.Errorf("failed to create password reset: %w", err)
	}

	// 6. Get user profile for name
	profile, err := s.userProfileRepo.GetByUserID(ctx, user.ID)
	var displayName string
	if err == nil && profile != nil {
		displayName = profile.DisplayName
	}

	// 7. Build reset link
	frontendURL := utils.GetEnv("FRONTEND_URL", "http://localhost:3000")
	resetLink := fmt.Sprintf("%s/reset-password?token=%s", frontendURL, token)

	// 8. Create outbox event for password reset email
	payloadData := map[string]any{
		"email":              user.Email,
		"name":               displayName,
		"reset_link":         resetLink,
		"resetLink":          resetLink, // alternative key
		"expires_in_minutes": 60,
		"expiresInMinutes":   60, // alternative key
	}
	payloadBytes, err := json.Marshal(payloadData)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	outboxEvent := &models.Outbox{
		AggregateID: user.ID,
		Topic:       "user.password_reset",
		Type:        "PasswordResetRequested",
		Payload:     payloadBytes,
		CreatedAt:   time.Now(),
	}

	if err := s.outboxRepo.Create(ctx, outboxEvent); err != nil {
		return fmt.Errorf("failed to create outbox event: %w", err)
	}

	// 9. Log audit event
	auditLog := &models.AuditLog{
		UserID: &user.ID,
		Action: "password.reset_requested",
		Metadata: map[string]any{
			"email": user.Email,
		},
		CreatedAt: time.Now(),
	}
	_ = s.auditLogRepo.Create(ctx, auditLog)

	return nil
}

func (s *passwordService) VerifyResetToken(ctx context.Context, token string) (*uuid.UUID, error) {
	// Hash the provided token
	tokenHash := utils.HashToken(token)

	// Find valid reset record
	reset, err := s.passwordResetRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		return nil, errors.New("invalid or expired token")
	}

	return &reset.UserID, nil
}

func (s *passwordService) CompletePasswordReset(ctx context.Context, token, newPassword string) error {
	// 1. Validate new password
	if err := utils.ValidatePassword(newPassword); err != nil {
		return err
	}

	// 2. Hash the token to find the reset record
	tokenHash := utils.HashToken(token)

	reset, err := s.passwordResetRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		return errors.New("invalid or expired token")
	}

	// 3. Hash new password
	newPasswordHash, err := utils.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// 4. Update user's password
	if err := s.userRepo.UpdatePassword(ctx, reset.UserID, newPasswordHash); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// 5. Consume the reset token
	if err := s.passwordResetRepo.Consume(ctx, reset.ID); err != nil {
		return fmt.Errorf("failed to consume token: %w", err)
	}

	// 6. Log audit event
	auditLog := &models.AuditLog{
		UserID: &reset.UserID,
		Action: "password.reset_completed",
		Metadata: map[string]any{
			"reset_id": reset.ID,
		},
		CreatedAt: time.Now(),
	}
	_ = s.auditLogRepo.Create(ctx, auditLog)

	return nil
}

func (s *passwordService) ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error {
	// 1. Get user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// 2. Verify old password
	if err := utils.ComparePassword(user.PasswordHash, oldPassword); err != nil {
		return errors.New("invalid old password")
	}

	// 3. Validate new password
	if err := utils.ValidatePassword(newPassword); err != nil {
		return err
	}

	// 4. Hash new password
	newPasswordHash, err := utils.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// 5. Update password
	if err := s.userRepo.UpdatePassword(ctx, userID, newPasswordHash); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// 6. Log audit event
	auditLog := &models.AuditLog{
		UserID: &userID,
		Action: "password.changed",
		Metadata: map[string]any{
			"method": "authenticated_change",
		},
		CreatedAt: time.Now(),
	}
	_ = s.auditLogRepo.Create(ctx, auditLog)

	return nil
}

func (s *passwordService) CleanupExpiredResets(ctx context.Context) error {
	return s.passwordResetRepo.DeleteExpired(ctx)
}
