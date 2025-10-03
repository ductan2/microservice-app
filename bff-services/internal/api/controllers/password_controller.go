package controllers

import (
	"net/http"

	"bff-services/internal/api/dto"
	"bff-services/internal/services"
	"bff-services/internal/utils"

	"github.com/gin-gonic/gin"
)

type PasswordController struct {
	userService services.UserService
}

func NewPasswordController(userService services.UserService) *PasswordController {
	return &PasswordController{userService: userService}
}

func (p *PasswordController) RequestReset(c *gin.Context) {
	var req dto.PasswordResetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, "Invalid request data", http.StatusBadRequest, err.Error())
		return
	}

	resp, err := p.userService.RequestPasswordReset(c.Request.Context(), req)
	if err != nil {
		utils.Fail(c, "Unable to initiate password reset", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}

func (p *PasswordController) ConfirmReset(c *gin.Context) {
	var req dto.PasswordResetConfirmRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, "Invalid request data", http.StatusBadRequest, err.Error())
		return
	}

	resp, err := p.userService.ConfirmPasswordReset(c.Request.Context(), req)
	if err != nil {
		utils.Fail(c, "Unable to confirm password reset", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}

func (p *PasswordController) ChangePassword(c *gin.Context) {
	token, ok := requireBearerToken(c)
	if !ok {
		return
	}

	var req dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, "Invalid request data", http.StatusBadRequest, err.Error())
		return
	}

	resp, err := p.userService.ChangePassword(c.Request.Context(), token, req)
	if err != nil {
		utils.Fail(c, "Unable to change password", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}
