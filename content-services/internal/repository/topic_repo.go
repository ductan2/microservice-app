package repository

import (
	"content-services/internal/models"
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TopicRepository interface {
	Create(ctx context.Context, topic *models.Topic) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Topic, error)
	GetBySlug(ctx context.Context, slug string) (*models.Topic, error)
	GetAll(ctx context.Context) ([]models.Topic, error)
	Update(ctx context.Context, topic *models.Topic) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type topicRepository struct {
	db *gorm.DB
}

func NewTopicRepository(db *gorm.DB) TopicRepository {
	return &topicRepository{db: db}
}

func (r *topicRepository) Create(ctx context.Context, topic *models.Topic) error {
	return r.db.WithContext(ctx).Create(topic).Error
}

func (r *topicRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Topic, error) {
	var topic models.Topic
	if err := r.db.WithContext(ctx).First(&topic, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &topic, nil
}

func (r *topicRepository) GetBySlug(ctx context.Context, slug string) (*models.Topic, error) {
	var topic models.Topic
	if err := r.db.WithContext(ctx).First(&topic, "slug = ?", slug).Error; err != nil {
		return nil, err
	}
	return &topic, nil
}

func (r *topicRepository) GetAll(ctx context.Context) ([]models.Topic, error) {
	var topics []models.Topic
	if err := r.db.WithContext(ctx).Order("created_at DESC").Find(&topics).Error; err != nil {
		return nil, err
	}
	return topics, nil
}

func (r *topicRepository) Update(ctx context.Context, topic *models.Topic) error {
	return r.db.WithContext(ctx).Save(topic).Error
}

func (r *topicRepository) Delete(ctx context.Context, id uuid.UUID) error {
	res := r.db.WithContext(ctx).Delete(&models.Topic{}, "id = ?", id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
