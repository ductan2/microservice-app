package repositories

import (
	"context"
	"time"
	"user-services/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PasswordResetRepository interface {
	Create(ctx context.Context, reset *models.PasswordReset) error
	GetByTokenHash(ctx context.Context, tokenHash string) (*models.PasswordReset, error)
	Consume(ctx context.Context, id uuid.UUID) error
	DeleteExpired(ctx context.Context) error
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error
}

type passwordResetRepository struct {
	db *gorm.DB
}

func NewPasswordResetRepository(db *gorm.DB) PasswordResetRepository {
	return &passwordResetRepository{db: db}
}

func (r *passwordResetRepository) Create(ctx context.Context, reset *models.PasswordReset) error {
	return r.db.WithContext(ctx).Create(reset).Error
}

func (r *passwordResetRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*models.PasswordReset, error) {
	var reset models.PasswordReset
	err := r.db.WithContext(ctx).
		Where("token_hash = ?", tokenHash).
		Where("expires_at > ?", time.Now()).
		Where("consumed_at IS NULL").
		First(&reset).Error

	if err != nil {
		return nil, err
	}
	return &reset, nil
}

func (r *passwordResetRepository) Consume(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&models.PasswordReset{}).
		Where("id = ?", id).
		Update("consumed_at", time.Now()).Error
}

func (r *passwordResetRepository) DeleteExpired(ctx context.Context) error {
	return r.db.WithContext(ctx).
		Where("expires_at < ?", time.Now()).
		Or("consumed_at IS NOT NULL").
		Delete(&models.PasswordReset{}).Error
}

func (r *passwordResetRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Delete(&models.PasswordReset{}).Error
}
