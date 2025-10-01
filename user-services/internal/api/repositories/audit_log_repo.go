package repositories

import (
	"context"
	"user-services/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuditLogRepository interface {
	Create(ctx context.Context, log *models.AuditLog) error
	GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.AuditLog, error)
	GetByAction(ctx context.Context, action string, limit, offset int) ([]models.AuditLog, error)
	CountByUserID(ctx context.Context, userID uuid.UUID) (int64, error)
}

type auditLogRepository struct {
	db *gorm.DB
}

func NewAuditLogRepository(db *gorm.DB) AuditLogRepository {
	return &auditLogRepository{db: db}
}

func (r *auditLogRepository) Create(ctx context.Context, log *models.AuditLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *auditLogRepository) GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.AuditLog, error) {
	// TODO: implement
	return nil, nil
}

func (r *auditLogRepository) GetByAction(ctx context.Context, action string, limit, offset int) ([]models.AuditLog, error) {
	// TODO: implement
	return nil, nil
}

func (r *auditLogRepository) CountByUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
	// TODO: implement
	return 0, nil
}
