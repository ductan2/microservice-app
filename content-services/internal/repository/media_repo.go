package repository

import (
	"content-services/internal/models"
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MediaRepository interface {
	Create(ctx context.Context, media *models.MediaAsset) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.MediaAsset, error)
	GetByIDs(ctx context.Context, ids []uuid.UUID) ([]models.MediaAsset, error)
	GetBySHA256(ctx context.Context, sha256 string) (*models.MediaAsset, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type mediaRepository struct {
	db *gorm.DB
}

func NewMediaRepository(db *gorm.DB) MediaRepository {
	return &mediaRepository{db: db}
}

func (r *mediaRepository) Create(ctx context.Context, media *models.MediaAsset) error {
	// TODO: implement
	return nil
}

func (r *mediaRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.MediaAsset, error) {
	// TODO: implement
	return nil, nil
}

func (r *mediaRepository) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]models.MediaAsset, error) {
	// TODO: implement
	return nil, nil
}

func (r *mediaRepository) GetBySHA256(ctx context.Context, sha256 string) (*models.MediaAsset, error) {
	// TODO: implement - check for duplicate uploads
	return nil, nil
}

func (r *mediaRepository) Delete(ctx context.Context, id uuid.UUID) error {
	// TODO: implement
	return nil
}
