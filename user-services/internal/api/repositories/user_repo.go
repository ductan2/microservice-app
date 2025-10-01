package repositories

import (
	"context"
	"errors"

	"user-services/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepository struct {
	DB *gorm.DB
}

// NewUserRepository creates a new user repository instance
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{DB: db}
}

// CreateUser inserts a new user with unique email
func (r *UserRepository) CreateUser(ctx context.Context, email, passwordHash string) (models.User, error) {
	user := models.User{
		Email:           email,
		EmailNormalized: email,
		EmailVerified:   false,
		PasswordHash:    passwordHash,
	}

	if err := r.DB.WithContext(ctx).Create(&user).Error; err != nil {
		return models.User{}, err
	}

	return user, nil
}

// CheckEmailExists checks if email already exists
func (r *UserRepository) CheckEmailExists(ctx context.Context, email string) (bool, error) {
	var count int64
	err := r.DB.WithContext(ctx).Model(&models.User{}).Where("email = ?", email).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetUserByEmail retrieves a user by their email address
func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (models.User, error) {
	var user models.User
	err := r.DB.WithContext(ctx).Preload("Profile").First(&user, "email = ?", email).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.User{}, errors.New("user not found")
		}
		return models.User{}, err
	}
	return user, nil
}

// GetByEmail retrieves a user by email (alias for auth service)
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.DB.WithContext(ctx).Preload("Profile").Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByID retrieves a user by UUID
func (r *UserRepository) GetByID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	var user models.User
	err := r.DB.WithContext(ctx).Preload("Profile").Where("id = ?", userID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByID retrieves a user by their ID (string version)
func (r *UserRepository) GetUserByID(ctx context.Context, userID string) (models.User, error) {
	var user models.User
	err := r.DB.WithContext(ctx).Where("id = ?", userID).First(&user).Preload("Profile").Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.User{}, errors.New("user not found")
		}
		return models.User{}, err
	}
	return user, nil
}

// UpdateUser updates user information
func (r *UserRepository) UpdateUser(ctx context.Context, user *models.User) error {
	return r.DB.WithContext(ctx).Save(user).Error
}

// UpdatePassword updates user's password hash
func (r *UserRepository) UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	return r.DB.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", userID).
		Update("password_hash", passwordHash).Error
}

// GetByVerificationToken retrieves a user by their email verification token
func (r *UserRepository) GetByVerificationToken(ctx context.Context, tokenHash string) (*models.User, error) {
	var user models.User
	err := r.DB.WithContext(ctx).
		Preload("Profile").
		Where("email_verification_token = ?", tokenHash).
		First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// DeleteUser soft deletes a user
func (r *UserRepository) DeleteUser(ctx context.Context, userID string) error {
	return r.DB.WithContext(ctx).Where("id = ?", userID).Delete(&models.User{}).Error
}
