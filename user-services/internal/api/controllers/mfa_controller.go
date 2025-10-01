package controllers

import (
	"user-services/internal/api/services"

	"github.com/gin-gonic/gin"
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
	// TODO: implement
}

// VerifyMFASetup godoc
// @Summary Verify MFA setup
// @Tags mfa
// @Accept json
// @Param request body dto.MFAVerifyRequest true "MFA Verify Request"
// @Success 200
// @Router /mfa/verify [post]
func (c *MFAController) VerifyMFASetup(ctx *gin.Context) {
	// TODO: implement
}

// DisableMFA godoc
// @Summary Disable MFA
// @Tags mfa
// @Accept json
// @Param request body dto.MFADisableRequest true "MFA Disable Request"
// @Success 200
// @Router /mfa/disable [post]
func (c *MFAController) DisableMFA(ctx *gin.Context) {
	// TODO: implement
}

// GetMFAMethods godoc
// @Summary Get user's MFA methods
// @Tags mfa
// @Produce json
// @Success 200 {array} dto.MFASetupResponse
// @Router /mfa/methods [get]
func (c *MFAController) GetMFAMethods(ctx *gin.Context) {
	// TODO: implement
}
