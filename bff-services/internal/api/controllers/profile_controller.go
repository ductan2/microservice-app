package controllers

import (
	"net/http"

	"bff-services/internal/api/dto"
	"bff-services/internal/services"
	"bff-services/internal/utils"

	"github.com/gin-gonic/gin"
)

type ProfileController struct {
	userService services.UserService
}

func NewProfileController(userService services.UserService) *ProfileController {
	return &ProfileController{userService: userService}
}

func (p *ProfileController) GetProfile(c *gin.Context) {
	token, ok := requireBearerToken(c)
	if !ok {
		return
	}

	resp, err := p.userService.GetProfile(c.Request.Context(), token)
	if err != nil {
		utils.Fail(c, "Unable to fetch profile", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}

func (p *ProfileController) UpdateProfile(c *gin.Context) {
	token, ok := requireBearerToken(c)
	if !ok {
		return
	}

	var req dto.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, "Invalid request data", http.StatusBadRequest, err.Error())
		return
	}

	resp, err := p.userService.UpdateProfile(c.Request.Context(), token, req)
	if err != nil {
		utils.Fail(c, "Unable to update profile", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}

func (p *ProfileController) CheckAuth(c *gin.Context) {
	token, ok := requireBearerToken(c)
	if !ok {
		return
	}

	resp, err := p.userService.CheckAuth(c.Request.Context(), token)
	if err != nil {
		utils.Fail(c, "Unable to verify authentication", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}
