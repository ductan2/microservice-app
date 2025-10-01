package repositories

import (
	"context"
	"user-services/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MFARepository interface {
	Create(ctx context.Context, mfa *models.MFAMethod) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.MFAMethod, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]models.MFAMethod, error)
	GetTOTPByUserID(ctx context.Context, userID uuid.UUID) (*models.MFAMethod, error)
	UpdateLastUsed(ctx context.Context, id uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type mfaRepository struct {
	db *gorm.DB
}

func NewMFARepository(db *gorm.DB) MFARepository {
	return &mfaRepository{db: db}
}

func (r *mfaRepository) Create(ctx context.Context, mfa *models.MFAMethod) error {
	return r.db.WithContext(ctx).Create(mfa).Error
}

func (r *mfaRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.MFAMethod, error) {
	var m models.MFAMethod
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&m).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *mfaRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]models.MFAMethod, error) {
	var methods []models.MFAMethod
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&methods).Error
	return methods, err
}

func (r *mfaRepository) GetTOTPByUserID(ctx context.Context, userID uuid.UUID) (*models.MFAMethod, error) {
	var m models.MFAMethod
	if err := r.db.WithContext(ctx).Where("user_id = ? AND type = ?", userID, "totp").First(&m).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *mfaRepository) UpdateLastUsed(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&models.MFAMethod{}).Where("id = ?", id).Update("last_used_at", gorm.Expr("now()")).Error
}

func (r *mfaRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&models.MFAMethod{}).Error
}
