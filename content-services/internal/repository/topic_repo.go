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
	// TODO: implement
	return nil
}

func (r *topicRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Topic, error) {
	// TODO: implement
	return nil, nil
}

func (r *topicRepository) GetBySlug(ctx context.Context, slug string) (*models.Topic, error) {
	// TODO: implement
	return nil, nil
}

func (r *topicRepository) GetAll(ctx context.Context) ([]models.Topic, error) {
	// TODO: implement
	return nil, nil
}

func (r *topicRepository) Update(ctx context.Context, topic *models.Topic) error {
	// TODO: implement
	return nil
}

func (r *topicRepository) Delete(ctx context.Context, id uuid.UUID) error {
	// TODO: implement
	return nil
}
