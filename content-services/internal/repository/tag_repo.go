package repository

import (
	"content-services/internal/models"
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type TagRepository interface {
	Create(ctx context.Context, tag *models.Tag) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Tag, error)
	GetBySlug(ctx context.Context, slug string) (*models.Tag, error)
	GetAll(ctx context.Context) ([]models.Tag, error)
	Update(ctx context.Context, tag *models.Tag) error
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
	return r.db.WithContext(ctx).Create(tag).Error
}

func (r *tagRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Tag, error) {
	var tag models.Tag
	if err := r.db.WithContext(ctx).First(&tag, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &tag, nil
}

func (r *tagRepository) GetBySlug(ctx context.Context, slug string) (*models.Tag, error) {
	var tag models.Tag
	if err := r.db.WithContext(ctx).First(&tag, "slug = ?", slug).Error; err != nil {
		return nil, err
	}
	return &tag, nil
}

func (r *tagRepository) GetAll(ctx context.Context) ([]models.Tag, error) {
	var tags []models.Tag
	if err := r.db.WithContext(ctx).Order("slug ASC").Find(&tags).Error; err != nil {
		return nil, err
	}
	return tags, nil
}

func (r *tagRepository) Update(ctx context.Context, tag *models.Tag) error {
	return r.db.WithContext(ctx).Save(tag).Error
}

func (r *tagRepository) Delete(ctx context.Context, id uuid.UUID) error {
	res := r.db.WithContext(ctx).Delete(&models.Tag{}, "id = ?", id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *tagRepository) AddTagToContent(ctx context.Context, tagID uuid.UUID, kind string, objectID uuid.UUID) error {
	ct := models.ContentTag{TagID: tagID, Kind: kind, ObjectID: objectID}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&ct).Error
}

func (r *tagRepository) RemoveTagFromContent(ctx context.Context, tagID uuid.UUID, kind string, objectID uuid.UUID) error {
	res := r.db.WithContext(ctx).Delete(&models.ContentTag{}, "tag_id = ? AND kind = ? AND object_id = ?", tagID, kind, objectID)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *tagRepository) GetContentTags(ctx context.Context, kind string, objectID uuid.UUID) ([]models.Tag, error) {
	var tags []models.Tag
	err := r.db.WithContext(ctx).
		Joins("JOIN content_tags ON content_tags.tag_id = tags.id").
		Where("content_tags.kind = ? AND content_tags.object_id = ?", kind, objectID).
		Order("tags.slug ASC").
		Find(&tags).Error
	if err != nil {
		return nil, err
	}
	return tags, nil
}
