package controllers

import (
	"net/http"

	"user-services/internal/api/dto"
	"user-services/internal/api/services"
	"user-services/internal/utils"

	"github.com/gin-gonic/gin"
)

type UserController struct {
	userService services.UserService
}

func NewUserController(userService services.UserService) *UserController {
	return &UserController{
		userService: userService,
	}
}

// ListUsers godoc
// @Summary List all users with pagination
// @Tags users
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Param status query string false "Filter by status" Enums(active, locked, disabled, deleted)
// @Param search query string false "Search by email"
// @Success 200 {object} dto.PaginatedResponse
// @Router /users [get]
func (c *UserController) ListUsers(ctx *gin.Context) {
	var req dto.ListUsersRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		utils.Fail(ctx, "Invalid request parameters", http.StatusBadRequest, err.Error())
		return
	}

	result, err := c.userService.ListUsers(ctx.Request.Context(), req)
	if err != nil {
		utils.Fail(ctx, "Failed to retrieve users", http.StatusInternalServerError, err.Error())
		return
	}

	utils.Success(ctx, result)
}
