package repositories

import (
	"context"
	"time"
	"user-services/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LoginAttemptRepository interface {
	Create(ctx context.Context, attempt *models.LoginAttempt) error
	CountRecentByEmail(ctx context.Context, email string, since time.Time) (int64, error)
	CountRecentByIP(ctx context.Context, ipAddr string, since time.Time) (int64, error)
	GetRecentByUserID(ctx context.Context, userID uuid.UUID, limit int) ([]models.LoginAttempt, error)
	DeleteOlderThan(ctx context.Context, before time.Time) error
}

type loginAttemptRepository struct {
	db *gorm.DB
}

func NewLoginAttemptRepository(db *gorm.DB) LoginAttemptRepository {
	return &loginAttemptRepository{db: db}
}

func (r *loginAttemptRepository) Create(ctx context.Context, attempt *models.LoginAttempt) error {
	return r.db.WithContext(ctx).Create(attempt).Error
}

func (r *loginAttemptRepository) CountRecentByEmail(ctx context.Context, email string, since time.Time) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.LoginAttempt{}).Where("email = ? AND created_at >= ?", email, since).Count(&count).Error
	return count, err
}

func (r *loginAttemptRepository) CountRecentByIP(ctx context.Context, ipAddr string, since time.Time) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.LoginAttempt{}).Where("ip_addr = ? AND created_at >= ?", ipAddr, since).Count(&count).Error
	return count, err
}

func (r *loginAttemptRepository) GetRecentByUserID(ctx context.Context, userID uuid.UUID, limit int) ([]models.LoginAttempt, error) {
	var attempts []models.LoginAttempt
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC").Limit(limit).Find(&attempts).Error
	return attempts, err
}

func (r *loginAttemptRepository) DeleteOlderThan(ctx context.Context, before time.Time) error {
	return r.db.WithContext(ctx).Where("created_at < ?", before).Delete(&models.LoginAttempt{}).Error
}
