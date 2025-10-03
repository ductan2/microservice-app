package controllers

import (
	"net/http"

	"bff-services/internal/api/dto"
	"bff-services/internal/services"
	"bff-services/internal/utils"

	"github.com/gin-gonic/gin"
)

type MFAController struct {
	userService services.UserService
}

func NewMFAController(userService services.UserService) *MFAController {
	return &MFAController{userService: userService}
}

func (m *MFAController) Setup(c *gin.Context) {
	token, ok := requireBearerToken(c)
	if !ok {
		return
	}

	var req dto.MFASetupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, "Invalid request data", http.StatusBadRequest, err.Error())
		return
	}

	resp, err := m.userService.SetupMFA(c.Request.Context(), token, req)
	if err != nil {
		utils.Fail(c, "Unable to setup MFA", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}

func (m *MFAController) Verify(c *gin.Context) {
	token, ok := requireBearerToken(c)
	if !ok {
		return
	}

	var req dto.MFAVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, "Invalid request data", http.StatusBadRequest, err.Error())
		return
	}

	resp, err := m.userService.VerifyMFA(c.Request.Context(), token, req)
	if err != nil {
		utils.Fail(c, "Unable to verify MFA", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}

func (m *MFAController) Disable(c *gin.Context) {
	token, ok := requireBearerToken(c)
	if !ok {
		return
	}

	var req dto.MFADisableRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, "Invalid request data", http.StatusBadRequest, err.Error())
		return
	}

	resp, err := m.userService.DisableMFA(c.Request.Context(), token, req)
	if err != nil {
		utils.Fail(c, "Unable to disable MFA", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}

func (m *MFAController) Methods(c *gin.Context) {
	token, ok := requireBearerToken(c)
	if !ok {
		return
	}

	resp, err := m.userService.GetMFAMethods(c.Request.Context(), token)
	if err != nil {
		utils.Fail(c, "Unable to fetch MFA methods", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}
