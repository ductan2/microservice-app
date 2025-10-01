package repositories

import (
	"context"
	"user-services/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserProfileRepository interface {
	Create(ctx context.Context, profile *models.UserProfile) error
	GetByUserID(ctx context.Context, userID uuid.UUID) (*models.UserProfile, error)
	Update(ctx context.Context, profile *models.UserProfile) error
	Delete(ctx context.Context, userID uuid.UUID) error
}

type userProfileRepository struct {
	db *gorm.DB
}

func NewUserProfileRepository(db *gorm.DB) UserProfileRepository {
	return &userProfileRepository{db: db}
}

func (r *userProfileRepository) Create(ctx context.Context, profile *models.UserProfile) error {
	return r.db.WithContext(ctx).Create(profile).Error
}

func (r *userProfileRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*models.UserProfile, error) {
	// TODO: implement
	return nil, nil
}

func (r *userProfileRepository) Update(ctx context.Context, profile *models.UserProfile) error {
	// TODO: implement
	return nil
}

func (r *userProfileRepository) Delete(ctx context.Context, userID uuid.UUID) error {
	// TODO: implement
	return nil
}
