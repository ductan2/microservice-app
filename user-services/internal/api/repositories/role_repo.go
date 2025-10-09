package repositories

import (
	"context"
	"strings"
	"user-services/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type RoleRepository interface {
	Create(ctx context.Context, role *models.Role) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Role, error)
	GetByName(ctx context.Context, name string) (*models.Role, error)
	GetAll(ctx context.Context) ([]models.Role, error)
	Delete(ctx context.Context, id uuid.UUID) error

	// User role assignments
	AssignRoleToUser(ctx context.Context, userID, roleID uuid.UUID) error
	RemoveRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error
	GetUserRoles(ctx context.Context, userID uuid.UUID) ([]models.Role, error)
	GetRoleUsers(ctx context.Context, roleID uuid.UUID) ([]uuid.UUID, error)
}

type roleRepository struct {
	db *gorm.DB
}

func NewRoleRepository(db *gorm.DB) RoleRepository {
	return &roleRepository{db: db}
}

func (r *roleRepository) Create(ctx context.Context, role *models.Role) error {
	return r.db.WithContext(ctx).Create(role).Error
}

func (r *roleRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Role, error) {
	var role models.Role
	if err := r.db.WithContext(ctx).First(&role, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) GetByName(ctx context.Context, name string) (*models.Role, error) {
	var role models.Role
	if err := r.db.WithContext(ctx).
		Where("LOWER(name) = ?", strings.ToLower(name)).
		First(&role).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) GetAll(ctx context.Context) ([]models.Role, error) {
	var roles []models.Role
	if err := r.db.WithContext(ctx).
		Order("name ASC").
		Find(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}

func (r *roleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&models.Role{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *roleRepository) AssignRoleToUser(ctx context.Context, userID, roleID uuid.UUID) error {
	userRole := models.UserRole{
		UserID: userID,
		RoleID: roleID,
	}
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&userRole).Error
}

func (r *roleRepository) RemoveRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Where("user_id = ? AND role_id = ?", userID, roleID).
		Delete(&models.UserRole{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *roleRepository) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]models.Role, error) {
	var roles []models.Role
	if err := r.db.WithContext(ctx).
		Model(&models.Role{}).
		Joins("JOIN user_roles ON user_roles.role_id = roles.id").
		Where("user_roles.user_id = ?", userID).
		Order("roles.name ASC").
		Find(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}

func (r *roleRepository) GetRoleUsers(ctx context.Context, roleID uuid.UUID) ([]uuid.UUID, error) {
	var userIDs []uuid.UUID
	if err := r.db.WithContext(ctx).
		Model(&models.UserRole{}).
		Where("role_id = ?", roleID).
		Pluck("user_id", &userIDs).Error; err != nil {
		return nil, err
	}
	return userIDs, nil
}
