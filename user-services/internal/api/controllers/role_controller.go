package controllers

import (
	"net/http"

	"user-services/internal/api/dto"
	"user-services/internal/api/services"
	"user-services/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
	var req dto.CreateRoleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.Fail(ctx, "Invalid role payload", http.StatusBadRequest, err.Error())
		return
	}

	role, err := c.roleService.CreateRole(ctx.Request.Context(), req.Name)
	if err != nil {
		switch err {
		case services.ErrInvalidRoleName:
			utils.Fail(ctx, "Role name is required", http.StatusBadRequest, err.Error())
			return
		case services.ErrRoleAlreadyExists:
			utils.Fail(ctx, "Role already exists", http.StatusConflict, err.Error())
			return
		default:
			utils.Fail(ctx, "Failed to create role", http.StatusInternalServerError, err.Error())
			return
		}
	}

	utils.Created(ctx, role)
}

// GetAllRoles godoc
// @Summary Get all roles
// @Tags roles
// @Produce json
// @Success 200 {array} dto.RoleResponse
// @Router /roles [get]
func (c *RoleController) GetAllRoles(ctx *gin.Context) {
	roles, err := c.roleService.GetAllRoles(ctx.Request.Context())
	if err != nil {
		utils.Fail(ctx, "Failed to fetch roles", http.StatusInternalServerError, err.Error())
		return
	}

	utils.Success(ctx, roles)
}

// DeleteRole godoc
// @Summary Delete a role (admin only)
// @Tags roles
// @Param id path string true "Role ID"
// @Success 204
// @Router /roles/{id} [delete]
func (c *RoleController) DeleteRole(ctx *gin.Context) {
	roleIDParam := ctx.Param("id")
	roleID, err := uuid.Parse(roleIDParam)
	if err != nil {
		utils.Fail(ctx, "Invalid role ID", http.StatusBadRequest, err.Error())
		return
	}

	if err := c.roleService.DeleteRole(ctx.Request.Context(), roleID); err != nil {
		if err == services.ErrRoleNotFound {
			utils.Fail(ctx, "Role not found", http.StatusNotFound, err.Error())
			return
		}
		utils.Fail(ctx, "Failed to delete role", http.StatusInternalServerError, err.Error())
		return
	}

	ctx.Status(http.StatusNoContent)
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
	userIDParam := ctx.Param("id")
	userID, err := uuid.Parse(userIDParam)
	if err != nil {
		utils.Fail(ctx, "Invalid user ID", http.StatusBadRequest, err.Error())
		return
	}

	var req dto.AssignRoleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.Fail(ctx, "Invalid request payload", http.StatusBadRequest, err.Error())
		return
	}

	if err := c.roleService.AssignRoleToUser(ctx.Request.Context(), userID, req.RoleName); err != nil {
		switch err {
		case services.ErrInvalidRoleName:
			utils.Fail(ctx, "Role name is required", http.StatusBadRequest, err.Error())
			return
		case services.ErrRoleNotFound:
			utils.Fail(ctx, "Role not found", http.StatusNotFound, err.Error())
			return
		default:
			utils.Fail(ctx, "Failed to assign role", http.StatusInternalServerError, err.Error())
			return
		}
	}

	utils.Success(ctx, gin.H{"message": "Role assigned successfully"})
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
	userIDParam := ctx.Param("id")
	userID, err := uuid.Parse(userIDParam)
	if err != nil {
		utils.Fail(ctx, "Invalid user ID", http.StatusBadRequest, err.Error())
		return
	}

	var req dto.RemoveRoleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.Fail(ctx, "Invalid request payload", http.StatusBadRequest, err.Error())
		return
	}

	if err := c.roleService.RemoveRoleFromUser(ctx.Request.Context(), userID, req.RoleName); err != nil {
		switch err {
		case services.ErrInvalidRoleName:
			utils.Fail(ctx, "Role name is required", http.StatusBadRequest, err.Error())
			return
		case services.ErrRoleNotFound:
			utils.Fail(ctx, "Role not found", http.StatusNotFound, err.Error())
			return
		case services.ErrUserRoleNotFound:
			utils.Fail(ctx, "User does not have this role", http.StatusNotFound, err.Error())
			return
		default:
			utils.Fail(ctx, "Failed to remove role", http.StatusInternalServerError, err.Error())
			return
		}
	}

	utils.Success(ctx, gin.H{"message": "Role removed successfully"})
}
