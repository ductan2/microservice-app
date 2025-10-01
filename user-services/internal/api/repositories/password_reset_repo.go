package repositories

import (
	"context"
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
	// TODO: implement
	return nil
}

func (r *passwordResetRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*models.PasswordReset, error) {
	// TODO: implement
	return nil, nil
}

func (r *passwordResetRepository) Consume(ctx context.Context, id uuid.UUID) error {
	// TODO: implement
	return nil
}

func (r *passwordResetRepository) DeleteExpired(ctx context.Context) error {
	// TODO: implement
	return nil
}

func (r *passwordResetRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	// TODO: implement
	return nil
}
