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

	respondWithServiceResponse(c, resp)
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

	respondWithServiceResponse(c, resp)
}

func (a *AuthController) Logout(c *gin.Context) {
	token, ok := requireBearerToken(c)
	if !ok {
		return
	}

	resp, err := a.userService.Logout(c.Request.Context(), token)
	if err != nil {
		utils.Fail(c, "Unable to logout", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}

func (a *AuthController) VerifyEmail(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		utils.Fail(c, "Verification token is required", http.StatusBadRequest, "missing token")
		return
	}

	resp, err := a.userService.VerifyEmail(c.Request.Context(), token)
	if err != nil {
		utils.Fail(c, "Unable to verify email", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}
