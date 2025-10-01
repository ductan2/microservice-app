package repository

import (
	"content-services/internal/models"
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LevelRepository interface {
	Create(ctx context.Context, level *models.Level) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Level, error)
	GetByCode(ctx context.Context, code string) (*models.Level, error)
	GetAll(ctx context.Context) ([]models.Level, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type levelRepository struct {
	db *gorm.DB
}

func NewLevelRepository(db *gorm.DB) LevelRepository {
	return &levelRepository{db: db}
}

func (r *levelRepository) Create(ctx context.Context, level *models.Level) error {
	// TODO: implement
	return nil
}

func (r *levelRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Level, error) {
	// TODO: implement
	return nil, nil
}

func (r *levelRepository) GetByCode(ctx context.Context, code string) (*models.Level, error) {
	// TODO: implement
	return nil, nil
}

func (r *levelRepository) GetAll(ctx context.Context) ([]models.Level, error) {
	// TODO: implement
	return nil, nil
}

func (r *levelRepository) Delete(ctx context.Context, id uuid.UUID) error {
	// TODO: implement
	return nil
}
