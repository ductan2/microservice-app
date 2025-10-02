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
	Update(ctx context.Context, level *models.Level) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type levelRepository struct {
	db *gorm.DB
}

func NewLevelRepository(db *gorm.DB) LevelRepository {
	return &levelRepository{db: db}
}

func (r *levelRepository) Create(ctx context.Context, level *models.Level) error {
	return r.db.WithContext(ctx).Create(level).Error
}

func (r *levelRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Level, error) {
	var level models.Level
	if err := r.db.WithContext(ctx).First(&level, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &level, nil
}

func (r *levelRepository) GetByCode(ctx context.Context, code string) (*models.Level, error) {
	var level models.Level
	if err := r.db.WithContext(ctx).First(&level, "code = ?", code).Error; err != nil {
		return nil, err
	}
	return &level, nil
}

func (r *levelRepository) GetAll(ctx context.Context) ([]models.Level, error) {
	var levels []models.Level
	if err := r.db.WithContext(ctx).Order("code ASC").Find(&levels).Error; err != nil {
		return nil, err
	}
	return levels, nil
}

func (r *levelRepository) Update(ctx context.Context, level *models.Level) error {
	return r.db.WithContext(ctx).Save(level).Error
}

func (r *levelRepository) Delete(ctx context.Context, id uuid.UUID) error {
	res := r.db.WithContext(ctx).Delete(&models.Level{}, "id = ?", id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
