package controllers

import (
	"net/http"
	"strings"

	"user-services/internal/api/dto"
	"user-services/internal/api/services"
	"user-services/internal/models"
	"user-services/internal/utils"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	Service *services.AuthService
}

func NewAuthController(authService *services.AuthService) *AuthController {
	return &AuthController{Service: authService}
}

func (a AuthController) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, "Invalid request data", http.StatusBadRequest, err.Error())
		return
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))

	result, err := a.Service.Register(c.Request.Context(), email, req.Password, req.Name)
	if err != nil {
		utils.Fail(c, "Failed to register", http.StatusInternalServerError, err.Error())
		return
	}

	// Return success message without token - user needs to verify email
	utils.Created(c, gin.H{
		"message": "Registration successful! Please check your email to verify your account.",
		"email":   result.User.Email,
	})
}

func (a AuthController) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, "Invalid request data", http.StatusBadRequest, err.Error())
		return
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))
	userAgent := c.GetHeader("User-Agent")
	ipAddr := c.ClientIP()

	result, err := a.Service.Login(c.Request.Context(), email, req.Password, req.MFACode, userAgent, ipAddr)
	if err != nil {
		// Handle specific error types
		switch err {
		case utils.ErrInvalidMFACode:
			utils.Fail(c, "Invalid MFA code", http.StatusUnauthorized, err.Error())
		case utils.ErrEmailNotVerified:
			utils.Fail(c, "Email not verified", http.StatusUnauthorized, err.Error())
		case utils.ErrInvalidCredentials:
			utils.Fail(c, "Invalid email or password", http.StatusUnauthorized, err.Error())
		default:
			utils.Fail(c, "Internal server error", http.StatusInternalServerError, err.Error())
		}
		return
	}

	response := dto.AuthResponse{
		AccessToken:  result.Token,
		RefreshToken: result.RefreshToken,
		ExpiresAt:    result.ExpiresAt,
		User:         modelToPublicUser(result.User),
	}
	utils.Success(c, response)
}

func (a AuthController) Logout(c *gin.Context) {
	utils.Success(c, gin.H{"message": "Logged out successfully"})
}

func (a AuthController) VerifyEmail(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		utils.Fail(c, "Verification token is required", http.StatusBadRequest, "missing token")
		return
	}

	if err := a.Service.VerifyEmail(c.Request.Context(), token); err != nil {
		utils.Fail(c, "Email verification failed", http.StatusBadRequest, err.Error())
		return
	}

	utils.Success(c, gin.H{
		"message": "Email verified successfully! You can now login.",
	})
}

func modelToPublicUser(user models.User) dto.PublicUser {
	return dto.PublicUser{
		ID:            user.ID,
		Email:         user.Email,
		EmailVerified: user.EmailVerified,
		Status:        user.Status,
		CreatedAt:     user.CreatedAt,
		Profile: &dto.UserProfile{
			DisplayName: user.Profile.DisplayName,
			AvatarURL:   user.Profile.AvatarURL,
			Locale:      user.Profile.Locale,
			TimeZone:    user.Profile.TimeZone,
		},
		UpdatedAt: user.UpdatedAt,
	}
}
