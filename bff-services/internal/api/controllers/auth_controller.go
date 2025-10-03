package controllers

import (
	"net/http"

	"bff-services/internal/api/dto"
	"bff-services/internal/services"
	"bff-services/internal/utils"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	userService services.UserService
}

func NewAuthController(userService services.UserService) *AuthController {
	return &AuthController{userService: userService}
}

func (a *AuthController) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, "Invalid request data", http.StatusBadRequest, err.Error())
		return
	}

	resp, err := a.userService.Register(c.Request.Context(), req)
	if err != nil {
		utils.Fail(c, "Unable to register user", http.StatusBadGateway, err.Error())
		return
	}

	respondWithEnvelope(c, resp)
}

func (a *AuthController) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, "Invalid request data", http.StatusBadRequest, err.Error())
		return
	}

	resp, err := a.userService.Login(c.Request.Context(), req, c.GetHeader("User-Agent"), c.ClientIP())
	if err != nil {
		utils.Fail(c, "Unable to login", http.StatusBadGateway, err.Error())
		return
	}

	respondWithEnvelope(c, resp)
}

func respondWithEnvelope(c *gin.Context, resp *services.HTTPResponse) {
	if resp == nil {
		c.Status(http.StatusNoContent)
		return
	}

	if resp.IsBodyEmpty() {
		c.Status(resp.StatusCode)
		return
	}

	c.JSON(resp.StatusCode, resp.Body)
}
