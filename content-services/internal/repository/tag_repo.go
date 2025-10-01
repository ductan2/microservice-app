package repository

import (
	"content-services/internal/models"
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TagRepository interface {
	Create(ctx context.Context, tag *models.Tag) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Tag, error)
	GetBySlug(ctx context.Context, slug string) (*models.Tag, error)
	GetAll(ctx context.Context) ([]models.Tag, error)
	Delete(ctx context.Context, id uuid.UUID) error

	// Content tagging
	AddTagToContent(ctx context.Context, tagID uuid.UUID, kind string, objectID uuid.UUID) error
	RemoveTagFromContent(ctx context.Context, tagID uuid.UUID, kind string, objectID uuid.UUID) error
	GetContentTags(ctx context.Context, kind string, objectID uuid.UUID) ([]models.Tag, error)
}

type tagRepository struct {
	db *gorm.DB
}

func NewTagRepository(db *gorm.DB) TagRepository {
	return &tagRepository{db: db}
}

func (r *tagRepository) Create(ctx context.Context, tag *models.Tag) error {
	// TODO: implement
	return nil
}

func (r *tagRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Tag, error) {
	// TODO: implement
	return nil, nil
}

func (r *tagRepository) GetBySlug(ctx context.Context, slug string) (*models.Tag, error) {
	// TODO: implement
	return nil, nil
}

func (r *tagRepository) GetAll(ctx context.Context) ([]models.Tag, error) {
	// TODO: implement
	return nil, nil
}

func (r *tagRepository) Delete(ctx context.Context, id uuid.UUID) error {
	// TODO: implement
	return nil
}

func (r *tagRepository) AddTagToContent(ctx context.Context, tagID uuid.UUID, kind string, objectID uuid.UUID) error {
	// TODO: implement
	return nil
}

func (r *tagRepository) RemoveTagFromContent(ctx context.Context, tagID uuid.UUID, kind string, objectID uuid.UUID) error {
	// TODO: implement
	return nil
}

func (r *tagRepository) GetContentTags(ctx context.Context, kind string, objectID uuid.UUID) ([]models.Tag, error) {
	// TODO: implement
	return nil, nil
}
