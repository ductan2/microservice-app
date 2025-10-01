package repositories

import (
	"context"
	"user-services/internal/models"

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
	return r.db.WithContext(ctx).Create(event).Error
}

func (r *outboxRepository) GetUnpublished(ctx context.Context, limit int) ([]models.Outbox, error) {
	var events []models.Outbox
	err := r.db.WithContext(ctx).
		Where("published_at IS NULL").
		Order("created_at ASC").
		Limit(limit).
		Find(&events).Error
	return events, err
}

func (r *outboxRepository) MarkAsPublished(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).
		Model(&models.Outbox{}).
		Where("id = ?", id).
		Update("published_at", "NOW()").Error
}

func (r *outboxRepository) DeletePublished(ctx context.Context, olderThanID int64) error {
	return r.db.WithContext(ctx).
		Where("id <= ? AND published_at IS NOT NULL", olderThanID).
		Delete(&models.Outbox{}).Error
}
