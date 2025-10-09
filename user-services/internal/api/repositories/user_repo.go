package repositories

import (
	"context"
	"errors"
	"time"

	"user-services/internal/models"
	"user-services/internal/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepository interface {
	CreateUser(ctx context.Context, email, passwordHash string) (models.User, error)
	CheckEmailExists(ctx context.Context, email string) (bool, error)
	GetUserByEmail(ctx context.Context, email string) (models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetByID(ctx context.Context, userID uuid.UUID) (*models.User, error)
	GetUserByID(ctx context.Context, userID string) (models.User, error)
	UpdateUser(ctx context.Context, user *models.User) error
	UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error
	UpdateLastLogin(ctx context.Context, userID uuid.UUID, at time.Time, ip string) error
	GetByVerificationToken(ctx context.Context, tokenHash string) (*models.User, error)
	DeleteUser(ctx context.Context, userID string) error
	ListUsers(ctx context.Context, page, pageSize int, status, search string) ([]models.User, int64, error)
}

type userRepository struct {
	DB *gorm.DB
}

// NewUserRepository creates a new user repository instance
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{DB: db}
}

// CreateUser inserts a new user with unique email
func (r *userRepository) CreateUser(ctx context.Context, email, passwordHash string) (models.User, error) {
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
func (r *userRepository) CheckEmailExists(ctx context.Context, email string) (bool, error) {
	var count int64
	err := r.DB.WithContext(ctx).Model(&models.User{}).Where("email = ?", email).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetUserByEmail retrieves a user by their email address
func (r *userRepository) GetUserByEmail(ctx context.Context, email string) (models.User, error) {
	var user models.User
	err := r.DB.WithContext(ctx).Preload("Profile").Preload("Roles").First(&user, "email = ?", email).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.User{}, errors.New("user not found")
		}
		return models.User{}, err
	}
	return user, nil
}

// GetByEmail retrieves a user by email (alias for auth service)
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.DB.WithContext(ctx).Preload("Profile").Preload("Roles").Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByID retrieves a user by UUID
func (r *userRepository) GetByID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	var user models.User
	err := r.DB.WithContext(ctx).Preload("Profile").Preload("Roles").Where("id = ?", userID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByID retrieves a user by their ID (string version)
func (r *userRepository) GetUserByID(ctx context.Context, userID string) (models.User, error) {
	var user models.User
	err := r.DB.WithContext(ctx).Preload("Profile").Preload("Roles").Where("id = ?", userID).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.User{}, errors.New("user not found")
		}
		return models.User{}, err
	}
	return user, nil
}

// UpdateUser updates user information
func (r *userRepository) UpdateUser(ctx context.Context, user *models.User) error {
	return r.DB.WithContext(ctx).Save(user).Error
}

// UpdatePassword updates user's password hash
func (r *userRepository) UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	return r.DB.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", userID).
		Update("password_hash", passwordHash).Error
}

// UpdateLastLogin updates last login timestamp and ip
func (r *userRepository) UpdateLastLogin(ctx context.Context, userID uuid.UUID, at time.Time, ip string) error {
	updates := map[string]any{
		"last_login_at": at,
	}

	// Sanitize IP address - returns empty string if invalid
	sanitizedIP := utils.SanitizeIPAddress(ip)
	if sanitizedIP != "" {
		updates["last_login_ip"] = &sanitizedIP
	} else {
		updates["last_login_ip"] = nil
	}

	return r.DB.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", userID).
		Updates(updates).Error
}

// GetByVerificationToken retrieves a user by their email verification token
func (r *userRepository) GetByVerificationToken(ctx context.Context, tokenHash string) (*models.User, error) {
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
func (r *userRepository) DeleteUser(ctx context.Context, userID string) error {
	return r.DB.WithContext(ctx).Where("id = ?", userID).Delete(&models.User{}).Error
}

// ListUsers retrieves a paginated list of users with optional filtering
func (r *userRepository) ListUsers(ctx context.Context, page, pageSize int, status, search string) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	query := r.DB.WithContext(ctx).Model(&models.User{})

	// Apply filters
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if search != "" {
		query = query.Where("email ILIKE ?", "%"+search+"%")
	}

	// Count total records
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	offset := (page - 1) * pageSize
	if err := query.
		Preload("Profile").
		Preload("Roles").
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}
