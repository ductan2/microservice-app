package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"user-services/internal/api/repositories"
	"user-services/internal/config"
	"user-services/internal/models"
	"user-services/internal/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const defaultResetTokenTTL = time.Hour

var (
	// ErrResetTokenInvalid indicates the reset token is missing, consumed, or expired.
	ErrResetTokenInvalid = errors.New("invalid or expired password reset token")
	// ErrPasswordMismatch indicates the provided current password is incorrect.
	ErrPasswordMismatch = errors.New("current password is incorrect")
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
	resetTTL          time.Duration
}

func NewPasswordService(
	userRepo *repositories.UserRepository,
	passwordResetRepo repositories.PasswordResetRepository,
	auditLogRepo repositories.AuditLogRepository,
	outboxRepo repositories.OutboxRepository,
) PasswordService {
	return &passwordService{
		userRepo:          userRepo,
		passwordResetRepo: passwordResetRepo,
		auditLogRepo:      auditLogRepo,
		outboxRepo:        outboxRepo,
		resetTTL:          defaultResetTokenTTL,
	}
}

func (s *passwordService) InitiatePasswordReset(ctx context.Context, email string) error {
	normalizedEmail := strings.TrimSpace(strings.ToLower(email))
	if err := utils.ValidateEmail(normalizedEmail); err != nil {
		return err
	}

	user, err := s.userRepo.GetUserByEmail(ctx, normalizedEmail)
	if err != nil {
		if errors.Is(err, repositories.ErrUserNotFound) || errors.Is(err, gorm.ErrRecordNotFound) {
			// Avoid user enumeration by returning success even if the email is unknown.
			return nil
		}
		return err
	}

	if err := s.passwordResetRepo.DeleteByUserID(ctx, user.ID); err != nil {
		return err
	}

	rawToken, tokenHash, err := generateResetToken()
	if err != nil {
		return err
	}

	reset := &models.PasswordReset{
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(s.resetTTL),
	}
	if err := s.passwordResetRepo.Create(ctx, reset); err != nil {
		return err
	}

	s.recordAudit(ctx, user.ID, "user.password_reset.requested", map[string]any{
		"email": user.Email,
	})

	if s.outboxRepo != nil {
		if err := s.enqueuePasswordResetEmail(ctx, user, rawToken); err != nil {
			return err
		}
	}

	return nil
}

func (s *passwordService) VerifyResetToken(ctx context.Context, token string) (*uuid.UUID, error) {
	reset, err := s.getActiveReset(ctx, token)
	if err != nil {
		return nil, err
	}
	return &reset.UserID, nil
}

func (s *passwordService) CompletePasswordReset(ctx context.Context, token, newPassword string) error {
	if err := utils.ValidatePassword(newPassword); err != nil {
		return err
	}

	reset, err := s.getActiveReset(ctx, token)
	if err != nil {
		return err
	}

	user, err := s.userRepo.GetUserByID(ctx, reset.UserID.String())
	if err != nil {
		if errors.Is(err, repositories.ErrUserNotFound) || errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrResetTokenInvalid
		}
		return err
	}

	hash, err := utils.HashPassword(newPassword)
	if err != nil {
		return err
	}

	user.PasswordHash = hash
	if err := s.userRepo.UpdateUser(ctx, &user); err != nil {
		return err
	}

	if err := s.passwordResetRepo.Consume(ctx, reset.ID); err != nil {
		return err
	}

	if err := s.passwordResetRepo.DeleteByUserID(ctx, reset.UserID); err != nil {
		return err
	}

	s.recordAudit(ctx, user.ID, "user.password_reset.completed", map[string]any{
		"reset_id": reset.ID.String(),
	})

	return nil
}

func (s *passwordService) ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error {
	if err := utils.ValidatePassword(newPassword); err != nil {
		return err
	}

	user, err := s.userRepo.GetUserByID(ctx, userID.String())
	if err != nil {
		return err
	}

	if err := utils.CheckPassword(user.PasswordHash, oldPassword); err != nil {
		return ErrPasswordMismatch
	}

	hash, err := utils.HashPassword(newPassword)
	if err != nil {
		return err
	}

	user.PasswordHash = hash
	if err := s.userRepo.UpdateUser(ctx, &user); err != nil {
		return err
	}

	if err := s.passwordResetRepo.DeleteByUserID(ctx, userID); err != nil {
		return err
	}

	s.recordAudit(ctx, user.ID, "user.password.changed", map[string]any{
		"initiator": "self",
	})

	return nil
}

func (s *passwordService) CleanupExpiredResets(ctx context.Context) error {
	return s.passwordResetRepo.DeleteExpired(ctx)
}

func (s *passwordService) getActiveReset(ctx context.Context, token string) (*models.PasswordReset, error) {
	trimmed := strings.TrimSpace(token)
	if trimmed == "" {
		return nil, ErrResetTokenInvalid
	}

	tokenHash := hashToken(trimmed)
	reset, err := s.passwordResetRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrResetTokenInvalid
		}
		return nil, err
	}

	if reset.ConsumedAt.Valid {
		return nil, ErrResetTokenInvalid
	}

	if time.Now().After(reset.ExpiresAt) {
		return nil, ErrResetTokenInvalid
	}

	return reset, nil
}

func (s *passwordService) recordAudit(ctx context.Context, userID uuid.UUID, action string, metadata map[string]any) {
	if s.auditLogRepo == nil {
		return
	}

	audit := &models.AuditLog{
		UserID:    &userID,
		ActorID:   &userID,
		Action:    action,
		Metadata:  metadata,
		CreatedAt: time.Now(),
	}

	_ = s.auditLogRepo.Create(ctx, audit)
}

func (s *passwordService) enqueuePasswordResetEmail(ctx context.Context, user models.User, token string) error {
	resetURL := buildPasswordResetURL(token)

	event := &models.Outbox{
		AggregateID: user.ID,
		Topic:       "user.events",
		Type:        "user.password_reset",
		Payload: map[string]any{
			"user_id":            user.ID.String(),
			"email":              user.Email,
			"reset_link":         resetURL,
			"reset_token":        token,
			"expires_in_minutes": int(s.resetTTL / time.Minute),
			"appName":            config.GetAppName(),
			"supportEmail":       config.GetSupportEmail(),
		},
		CreatedAt: time.Now(),
	}

	return s.outboxRepo.Create(ctx, event)
}

func generateResetToken() (string, string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", "", err
	}

	raw := base64.RawURLEncoding.EncodeToString(buf)
	return raw, hashToken(raw), nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func buildPasswordResetURL(token string) string {
	base := strings.TrimRight(config.GetPublicAppURL(), "/")
	path := strings.TrimLeft(config.GetPasswordResetPath(), "/")
	if path == "" {
		return fmt.Sprintf("%s?token=%s", base, token)
	}
	return fmt.Sprintf("%s/%s?token=%s", base, path, token)
}
