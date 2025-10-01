package controllers

import (
	"errors"
	"net/http"
	"strings"

	"user-services/internal/api/dto"
	"user-services/internal/api/services"
	"user-services/internal/utils"

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
		utils.Fail(ctx, "Invalid request data", http.StatusBadRequest, err.Error())
		return
	}

	email := strings.TrimSpace(strings.ToLower(req.Email))
	if email == "" {
		utils.Fail(ctx, "Email is required", http.StatusBadRequest, nil)
		return
	}

	if err := c.passwordService.InitiatePasswordReset(ctx.Request.Context(), email); err != nil {
		switch {
		case errors.Is(err, utils.ErrEmailRequired), errors.Is(err, utils.ErrInvalidEmail):
			utils.Fail(ctx, "Invalid email address", http.StatusBadRequest, err.Error())
		default:
			utils.Fail(ctx, "Failed to initiate password reset", http.StatusInternalServerError, err.Error())
		}
		return
	}

	utils.Success(ctx, gin.H{"message": "Password reset request received"})
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
		utils.Fail(ctx, "Invalid request data", http.StatusBadRequest, err.Error())
		return
	}

	if strings.TrimSpace(req.Token) == "" || strings.TrimSpace(req.NewPassword) == "" {
		utils.Fail(ctx, "Token and new password are required", http.StatusBadRequest, nil)
		return
	}

	if err := c.passwordService.CompletePasswordReset(ctx.Request.Context(), strings.TrimSpace(req.Token), req.NewPassword); err != nil {
		switch {
		case errors.Is(err, services.ErrResetTokenInvalid):
			utils.Fail(ctx, "Invalid or expired reset token", http.StatusBadRequest, err.Error())
		case errors.Is(err, utils.ErrWeakPassword), errors.Is(err, utils.ErrPasswordRequired):
			utils.Fail(ctx, "New password does not meet requirements", http.StatusBadRequest, err.Error())
		default:
			utils.Fail(ctx, "Failed to reset password", http.StatusInternalServerError, err.Error())
		}
		return
	}

	utils.Success(ctx, gin.H{"message": "Password has been reset"})
}

// ChangePassword godoc
// @Summary Change password (authenticated)
// @Tags password
// @Accept json
// @Param request body dto.ChangePasswordRequest true "Change Password Request"
// @Success 200
// @Router /password/change [post]
func (c *PasswordController) ChangePassword(ctx *gin.Context) {
	var req dto.ChangePasswordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.Fail(ctx, "Invalid request data", http.StatusBadRequest, err.Error())
		return
	}

	if strings.TrimSpace(req.OldPassword) == "" || strings.TrimSpace(req.NewPassword) == "" {
		utils.Fail(ctx, "Current and new password are required", http.StatusBadRequest, nil)
		return
	}

	userIDValue, exists := ctx.Get("userID")
	if !exists {
		utils.Fail(ctx, "User not authenticated", http.StatusUnauthorized, nil)
		return
	}

	var userID uuid.UUID
	switch v := userIDValue.(type) {
	case uuid.UUID:
		userID = v
	case string:
		parsed, err := uuid.Parse(v)
		if err != nil {
			utils.Fail(ctx, "Invalid user identifier", http.StatusUnauthorized, err.Error())
			return
		}
		userID = parsed
	default:
		utils.Fail(ctx, "Invalid user identifier", http.StatusUnauthorized, nil)
		return
	}

	if err := c.passwordService.ChangePassword(ctx.Request.Context(), userID, req.OldPassword, req.NewPassword); err != nil {
		switch {
		case errors.Is(err, services.ErrPasswordMismatch):
			utils.Fail(ctx, "Current password is incorrect", http.StatusBadRequest, err.Error())
		case errors.Is(err, utils.ErrWeakPassword), errors.Is(err, utils.ErrPasswordRequired):
			utils.Fail(ctx, "New password does not meet requirements", http.StatusBadRequest, err.Error())
		default:
			utils.Fail(ctx, "Failed to change password", http.StatusInternalServerError, err.Error())
		}
		return
	}

	utils.Success(ctx, gin.H{"message": "Password changed successfully"})
}
