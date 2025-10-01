package repositories

import (
	"context"
	"user-services/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SessionRepository interface {
	Create(ctx context.Context, session *models.Session) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Session, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]models.Session, error)
	Revoke(ctx context.Context, id uuid.UUID) error
	RevokeAllByUserID(ctx context.Context, userID uuid.UUID) error
	DeleteExpired(ctx context.Context) error
}

type sessionRepository struct {
	db *gorm.DB
}

func NewSessionRepository(db *gorm.DB) SessionRepository {
	return &sessionRepository{db: db}
}

func (r *sessionRepository) Create(ctx context.Context, session *models.Session) error {
	return r.db.WithContext(ctx).Create(session).Error
}

func (r *sessionRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Session, error) {
	// TODO: implement
	return nil, nil
}

func (r *sessionRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]models.Session, error) {
	// TODO: implement
	return nil, nil
}

func (r *sessionRepository) Revoke(ctx context.Context, id uuid.UUID) error {
	// TODO: implement
	return nil
}

func (r *sessionRepository) RevokeAllByUserID(ctx context.Context, userID uuid.UUID) error {
	// TODO: implement
	return nil
}

func (r *sessionRepository) DeleteExpired(ctx context.Context) error {
	// TODO: implement
	return nil
}
