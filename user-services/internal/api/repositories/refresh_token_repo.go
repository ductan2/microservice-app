package repositories

import (
	"context"
	"user-services/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RefreshTokenRepository interface {
	Create(ctx context.Context, token *models.RefreshToken) error
	GetByTokenHash(ctx context.Context, tokenHash string) (*models.RefreshToken, error)
	Consume(ctx context.Context, id uuid.UUID) error
	Revoke(ctx context.Context, id uuid.UUID) error
	RevokeBySessionID(ctx context.Context, sessionID uuid.UUID) error
	DeleteExpired(ctx context.Context) error
}

type refreshTokenRepository struct {
	db *gorm.DB
}

func NewRefreshTokenRepository(db *gorm.DB) RefreshTokenRepository {
	return &refreshTokenRepository{db: db}
}

func (r *refreshTokenRepository) Create(ctx context.Context, token *models.RefreshToken) error {
	return r.db.WithContext(ctx).Create(token).Error
}

func (r *refreshTokenRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*models.RefreshToken, error) {
	var t models.RefreshToken
	if err := r.db.WithContext(ctx).Where("token_hash = ?", tokenHash).First(&t).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *refreshTokenRepository) Consume(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&models.RefreshToken{}).Where("id = ?", id).Update("consumed_at", gorm.Expr("now()")).Error
}

func (r *refreshTokenRepository) Revoke(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&models.RefreshToken{}).Where("id = ?", id).Update("revoked_at", gorm.Expr("now()")).Error
}

func (r *refreshTokenRepository) RevokeBySessionID(ctx context.Context, sessionID uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&models.RefreshToken{}).Where("session_id = ? AND revoked_at IS NULL", sessionID).Update("revoked_at", gorm.Expr("now()")).Error
}

func (r *refreshTokenRepository) DeleteExpired(ctx context.Context) error {
	return r.db.WithContext(ctx).Where("expires_at < now()").Delete(&models.RefreshToken{}).Error
}
