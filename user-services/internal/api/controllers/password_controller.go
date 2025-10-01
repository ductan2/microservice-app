package controllers

import (
	"user-services/internal/api/services"

	"github.com/gin-gonic/gin"
)

type PasswordController struct {
	passwordService services.PasswordService
}

func NewPasswordController(passwordService services.PasswordService) *PasswordController {
	return &PasswordController{
		passwordService: passwordService,
	}
}

// RequestPasswordReset godoc
// @Summary Request password reset
// @Tags password
// @Accept json
// @Param request body dto.PasswordResetRequestDTO true "Password Reset Request"
// @Success 200
// @Router /password/reset/request [post]
func (c *PasswordController) RequestPasswordReset(ctx *gin.Context) {
	// TODO: implement
}

// ConfirmPasswordReset godoc
// @Summary Confirm password reset
// @Tags password
// @Accept json
// @Param request body dto.PasswordResetConfirmDTO true "Password Reset Confirm"
// @Success 200
// @Router /password/reset/confirm [post]
func (c *PasswordController) ConfirmPasswordReset(ctx *gin.Context) {
	// TODO: implement
}

// ChangePassword godoc
// @Summary Change password (authenticated)
// @Tags password
// @Accept json
// @Param request body dto.ChangePasswordRequest true "Change Password Request"
// @Success 200
// @Router /password/change [post]
func (c *PasswordController) ChangePassword(ctx *gin.Context) {
	// TODO: implement
}
