package controllers

import (
	"net/http"
	"user-services/internal/api/dto"
	"user-services/internal/api/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
	var req dto.PasswordResetRequestDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.passwordService.InitiatePasswordReset(ctx.Request.Context(), req.Email); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initiate password reset"})
		return
	}

	// Always return success even if email doesn't exist (security best practice)
	ctx.JSON(http.StatusOK, gin.H{
		"message": "If the email exists, a password reset link has been sent",
	})
}

// ConfirmPasswordReset godoc
// @Summary Confirm password reset
// @Tags password
// @Accept json
// @Param request body dto.PasswordResetConfirmDTO true "Password Reset Confirm"
// @Success 200
// @Router /password/reset/confirm [post]
func (c *PasswordController) ConfirmPasswordReset(ctx *gin.Context) {
	var req dto.PasswordResetConfirmDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.passwordService.CompletePasswordReset(ctx.Request.Context(), req.Token, req.NewPassword); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Password has been reset successfully",
	})
}

// ChangePassword godoc
// @Summary Change password (authenticated)
// @Tags password
// @Accept json
// @Param request body dto.ChangePasswordRequest true "Change Password Request"
// @Success 200
// @Router /password/change [post]
func (c *PasswordController) ChangePassword(ctx *gin.Context) {
	// Extract user ID from JWT context
	userIDStr, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req dto.ChangePasswordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.passwordService.ChangePassword(ctx.Request.Context(), userID, req.OldPassword, req.NewPassword); err != nil {
		if err.Error() == "invalid old password" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to change password"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Password changed successfully",
	})
}
