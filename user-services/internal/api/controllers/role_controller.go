package controllers

import (
	"user-services/internal/api/services"

	"github.com/gin-gonic/gin"
)

type RoleController struct {
	roleService services.RoleService
}

func NewRoleController(roleService services.RoleService) *RoleController {
	return &RoleController{
		roleService: roleService,
	}
}

// CreateRole godoc
// @Summary Create a new role (admin only)
// @Tags roles
// @Accept json
// @Produce json
// @Param request body dto.CreateRoleRequest true "Create Role Request"
// @Success 201 {object} dto.RoleResponse
// @Router /roles [post]
func (c *RoleController) CreateRole(ctx *gin.Context) {
	// TODO: implement
}

// GetAllRoles godoc
// @Summary Get all roles
// @Tags roles
// @Produce json
// @Success 200 {array} dto.RoleResponse
// @Router /roles [get]
func (c *RoleController) GetAllRoles(ctx *gin.Context) {
	// TODO: implement
}

// DeleteRole godoc
// @Summary Delete a role (admin only)
// @Tags roles
// @Param id path string true "Role ID"
// @Success 204
// @Router /roles/{id} [delete]
func (c *RoleController) DeleteRole(ctx *gin.Context) {
	// TODO: implement
}

// AssignRoleToUser godoc
// @Summary Assign role to user (admin only)
// @Tags roles
// @Accept json
// @Param id path string true "User ID"
// @Param request body dto.AssignRoleRequest true "Assign Role Request"
// @Success 200
// @Router /users/{id}/roles [post]
func (c *RoleController) AssignRoleToUser(ctx *gin.Context) {
	// TODO: implement
}

// RemoveRoleFromUser godoc
// @Summary Remove role from user (admin only)
// @Tags roles
// @Accept json
// @Param id path string true "User ID"
// @Param request body dto.RemoveRoleRequest true "Remove Role Request"
// @Success 200
// @Router /users/{id}/roles [delete]
func (c *RoleController) RemoveRoleFromUser(ctx *gin.Context) {
	// TODO: implement
}
