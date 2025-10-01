package repository

import (
	"content-services/internal/models"
	"context"

	"gorm.io/gorm"
)

type OutboxRepository interface {
	Create(ctx context.Context, event *models.Outbox) error
	GetUnpublished(ctx context.Context, limit int) ([]models.Outbox, error)
	MarkAsPublished(ctx context.Context, id int64) error
	DeletePublished(ctx context.Context, olderThanID int64) error
}

type outboxRepository struct {
	db *gorm.DB
}

func NewOutboxRepository(db *gorm.DB) OutboxRepository {
	return &outboxRepository{db: db}
}

func (r *outboxRepository) Create(ctx context.Context, event *models.Outbox) error {
	// TODO: implement
	return nil
}

func (r *outboxRepository) GetUnpublished(ctx context.Context, limit int) ([]models.Outbox, error) {
	// TODO: implement - WHERE published_at IS NULL ORDER BY id LIMIT
	return nil, nil
}

func (r *outboxRepository) MarkAsPublished(ctx context.Context, id int64) error {
	// TODO: implement - UPDATE published_at = now()
	return nil
}

func (r *outboxRepository) DeletePublished(ctx context.Context, olderThanID int64) error {
	// TODO: implement - cleanup old published events
	return nil
}
