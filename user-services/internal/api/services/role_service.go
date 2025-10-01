package services

import (
	"context"
	"user-services/internal/api/dto"
	"user-services/internal/api/repositories"

	"github.com/google/uuid"
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
	// TODO: implement
	return nil, nil
}

func (s *roleService) GetRoleByName(ctx context.Context, name string) (*dto.RoleResponse, error) {
	// TODO: implement
	return nil, nil
}

func (s *roleService) GetAllRoles(ctx context.Context) ([]dto.RoleResponse, error) {
	// TODO: implement
	return nil, nil
}

func (s *roleService) DeleteRole(ctx context.Context, roleID uuid.UUID) error {
	// TODO: implement
	return nil
}

func (s *roleService) AssignRoleToUser(ctx context.Context, userID uuid.UUID, roleName string) error {
	// TODO: implement
	return nil
}

func (s *roleService) RemoveRoleFromUser(ctx context.Context, userID uuid.UUID, roleName string) error {
	// TODO: implement
	return nil
}

func (s *roleService) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]string, error) {
	// TODO: implement
	return nil, nil
}

func (s *roleService) CheckUserHasRole(ctx context.Context, userID uuid.UUID, roleName string) (bool, error) {
	// TODO: implement
	return false, nil
}
