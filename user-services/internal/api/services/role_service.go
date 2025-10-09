package services

import (
	"context"
	"errors"
	"strings"
	"user-services/internal/api/dto"
	"user-services/internal/api/repositories"
	"user-services/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrInvalidRoleName   = errors.New("invalid role name")
	ErrRoleAlreadyExists = errors.New("role already exists")
	ErrRoleNotFound      = errors.New("role not found")
	ErrUserRoleNotFound  = errors.New("user role assignment not found")
)

type RoleService interface {
	CreateRole(ctx context.Context, name string) (*dto.RoleResponse, error)
	GetRoleByName(ctx context.Context, name string) (*dto.RoleResponse, error)
	GetAllRoles(ctx context.Context) ([]dto.RoleResponse, error)
	DeleteRole(ctx context.Context, roleID uuid.UUID) error

	AssignRoleToUser(ctx context.Context, userID uuid.UUID, roleName string) error
	RemoveRoleFromUser(ctx context.Context, userID uuid.UUID, roleName string) error
	GetUserRoles(ctx context.Context, userID uuid.UUID) ([]string, error)
	CheckUserHasRole(ctx context.Context, userID uuid.UUID, roleName string) (bool, error)
}

type roleService struct {
	roleRepo repositories.RoleRepository
}

func NewRoleService(roleRepo repositories.RoleRepository) RoleService {
	return &roleService{
		roleRepo: roleRepo,
	}
}

func (s *roleService) CreateRole(ctx context.Context, name string) (*dto.RoleResponse, error) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return nil, ErrInvalidRoleName
	}

	if _, err := s.roleRepo.GetByName(ctx, trimmed); err == nil {
		return nil, ErrRoleAlreadyExists
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	role := &models.Role{Name: trimmed}
	if err := s.roleRepo.Create(ctx, role); err != nil {
		return nil, err
	}

	return &dto.RoleResponse{ID: role.ID, Name: role.Name}, nil
}

func (s *roleService) GetRoleByName(ctx context.Context, name string) (*dto.RoleResponse, error) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return nil, ErrInvalidRoleName
	}

	role, err := s.roleRepo.GetByName(ctx, trimmed)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRoleNotFound
		}
		return nil, err
	}

	return &dto.RoleResponse{ID: role.ID, Name: role.Name}, nil
}

func (s *roleService) GetAllRoles(ctx context.Context) ([]dto.RoleResponse, error) {
	roles, err := s.roleRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.RoleResponse, len(roles))
	for i, role := range roles {
		responses[i] = dto.RoleResponse{ID: role.ID, Name: role.Name}
	}
	return responses, nil
}

func (s *roleService) DeleteRole(ctx context.Context, roleID uuid.UUID) error {
	if err := s.roleRepo.Delete(ctx, roleID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRoleNotFound
		}
		return err
	}
	return nil
}

func (s *roleService) AssignRoleToUser(ctx context.Context, userID uuid.UUID, roleName string) error {
	trimmed := strings.TrimSpace(roleName)
	if trimmed == "" {
		return ErrInvalidRoleName
	}

	role, err := s.roleRepo.GetByName(ctx, trimmed)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRoleNotFound
		}
		return err
	}

	return s.roleRepo.AssignRoleToUser(ctx, userID, role.ID)
}

func (s *roleService) RemoveRoleFromUser(ctx context.Context, userID uuid.UUID, roleName string) error {
	trimmed := strings.TrimSpace(roleName)
	if trimmed == "" {
		return ErrInvalidRoleName
	}

	role, err := s.roleRepo.GetByName(ctx, trimmed)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRoleNotFound
		}
		return err
	}

	if err := s.roleRepo.RemoveRoleFromUser(ctx, userID, role.ID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserRoleNotFound
		}
		return err
	}
	return nil
}

func (s *roleService) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]string, error) {
	roles, err := s.roleRepo.GetUserRoles(ctx, userID)
	if err != nil {
		return nil, err
	}

	roleNames := make([]string, len(roles))
	for i, role := range roles {
		roleNames[i] = role.Name
	}
	return roleNames, nil
}

func (s *roleService) CheckUserHasRole(ctx context.Context, userID uuid.UUID, roleName string) (bool, error) {
	trimmed := strings.TrimSpace(roleName)
	if trimmed == "" {
		return false, ErrInvalidRoleName
	}

	roles, err := s.roleRepo.GetUserRoles(ctx, userID)
	if err != nil {
		return false, err
	}

	for _, role := range roles {
		if strings.EqualFold(role.Name, trimmed) {
			return true, nil
		}
	}
	return false, nil
}
