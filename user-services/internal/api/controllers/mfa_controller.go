package controllers

import (
	"net/http"
	"user-services/internal/api/dto"
	"user-services/internal/api/middleware"
	"user-services/internal/api/services"
	"user-services/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type MFAController struct {
	mfaService services.MFAService
}

func NewMFAController(mfaService services.MFAService) *MFAController {
	return &MFAController{
		mfaService: mfaService,
	}
}

// SetupMFA godoc
// @Summary Setup MFA (TOTP or WebAuthn)
// @Tags mfa
// @Accept json
// @Produce json
// @Param request body dto.MFASetupRequest true "MFA Setup Request"
// @Success 200 {object} dto.MFASetupResponse
// @Router /mfa/setup [post]
func (c *MFAController) SetupMFA(ctx *gin.Context) {
	userID, ok := ctx.Get(middleware.ContextUserIDKey())
	if !ok {
		utils.Fail(ctx, "User not found", http.StatusUnauthorized, "user_not_found")
		return
	}

	var req dto.MFASetupRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.Fail(ctx, "Invalid request data", http.StatusBadRequest, err.Error())
		return
	}

	result, err := c.mfaService.SetupMFA(ctx.Request.Context(), userID.(uuid.UUID), req.Type, req.Label)
	if err != nil {
		utils.Fail(ctx, "Failed to setup MFA", http.StatusInternalServerError, err.Error())
		return
	}

	utils.Success(ctx, result)
}

// VerifyMFASetup godoc
// @Summary Verify MFA setup (activate method after user enters code)
// @Tags mfa
// @Accept json
// @Produce json
// @Param request body dto.VerifyMFARequest true "MFA Verify Request"
// @Success 200
// @Router /mfa/verify [post]
func (c *MFAController) VerifyMFASetup(ctx *gin.Context) {
	userID, ok := ctx.Get(middleware.ContextUserIDKey())
	if !ok {
		utils.Fail(ctx, "User not found", http.StatusUnauthorized, "user_not_found")
		return
	}

	var req dto.MFAVerifyRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.Fail(ctx, "Invalid request data", http.StatusBadRequest, err.Error())
		return
	}

	if err := c.mfaService.VerifyMFASetup(ctx.Request.Context(), userID.(uuid.UUID), req.MethodID, req.Code); err != nil {
		utils.Fail(ctx, "Failed to verify MFA", http.StatusUnauthorized, err.Error())
		return
	}

	utils.Success(ctx, gin.H{"message": "MFA verified successfully"})
}

// DisableMFA godoc
// @Summary Disable MFA (requires password)
// @Tags mfa
// @Accept json
// @Produce json
// @Param request body dto.DisableMFARequest true "MFA Disable Request"
// @Success 200
// @Router /mfa/disable [post]
func (c *MFAController) DisableMFA(ctx *gin.Context) {
	userID, ok := ctx.Get(middleware.ContextUserIDKey())
	if !ok {
		utils.Fail(ctx, "User not found", http.StatusUnauthorized, "user_not_found")
		return
	}

	var req dto.MFADisableRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.Fail(ctx, "Invalid request data", http.StatusBadRequest, err.Error())
		return
	}

	if err := c.mfaService.DisableMFA(ctx.Request.Context(), userID.(uuid.UUID), req.MethodID, req.Password); err != nil {
		utils.Fail(ctx, "Failed to disable MFA", http.StatusInternalServerError, err.Error())
		return
	}

	utils.Success(ctx, gin.H{"message": "MFA disabled"})
}

// GetMFAMethods godoc
// @Summary Get user's MFA methods
// @Tags mfa
// @Produce json
// @Success 200 {array} dto.MFASetupResponse
// @Router /mfa/methods [get]
func (c *MFAController) GetMFAMethods(ctx *gin.Context) {
	userID, ok := ctx.Get(middleware.ContextUserIDKey())
	if !ok {
		utils.Fail(ctx, "User not found", http.StatusUnauthorized, "user_not_found")
		return
	}

	result, err := c.mfaService.GetUserMFAMethods(ctx.Request.Context(), userID.(uuid.UUID))
	if err != nil {
		utils.Fail(ctx, "Failed to get MFA methods", http.StatusInternalServerError, err.Error())
		return
	}

	utils.Success(ctx, result)
}
