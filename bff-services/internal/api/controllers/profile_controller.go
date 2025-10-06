package controllers

import (
	"net/http"

	"bff-services/internal/api/dto"
	middleware "bff-services/internal/middlewares"
	"bff-services/internal/services"
	"bff-services/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ProfileController struct {
	userService services.UserService
}

func NewProfileController(userService services.UserService) *ProfileController {
	return &ProfileController{userService: userService}
}

func (p *ProfileController) GetProfile(c *gin.Context) {
	userID, email, sessionID, ok := getUserContextFromMiddleware(c)
	if !ok {
		return
	}

	resp, err := p.userService.GetProfileWithContext(c.Request.Context(), userID, email, sessionID)
	if err != nil {
		utils.Fail(c, "Unable to fetch profile", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}

func (p *ProfileController) UpdateProfile(c *gin.Context) {
	userID, email, sessionID, ok := getUserContextFromMiddleware(c)
	if !ok {
		return
	}

	var req dto.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, "Invalid request data", http.StatusBadRequest, err.Error())
		return
	}

	resp, err := p.userService.UpdateProfileWithContext(c.Request.Context(), userID, email, sessionID, req)
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

// getUserContextFromMiddleware extracts user context set by auth middleware
func getUserContextFromMiddleware(c *gin.Context) (userID, email, sessionID string, ok bool) {
	userIDValue, exists := c.Get(middleware.ContextUserIDKey())
	if !exists {
		utils.Fail(c, "Unauthorized", http.StatusUnauthorized, "user context not found")
		return "", "", "", false
	}

	emailValue, exists := c.Get(middleware.ContextUserEmailKey())
	if !exists {
		utils.Fail(c, "Unauthorized", http.StatusUnauthorized, "email context not found")
		return "", "", "", false
	}

	sessionIDValue, exists := c.Get(middleware.ContextSessionIDKey())
	if !exists {
		utils.Fail(c, "Unauthorized", http.StatusUnauthorized, "session context not found")
		return "", "", "", false
	}

	// Convert UUID to string for internal communication
	userIDUUID, ok := userIDValue.(uuid.UUID)
	if !ok {
		utils.Fail(c, "Unauthorized", http.StatusUnauthorized, "invalid user ID type")
		return "", "", "", false
	}

	emailStr, ok := emailValue.(string)
	if !ok {
		utils.Fail(c, "Unauthorized", http.StatusUnauthorized, "invalid email type")
		return "", "", "", false
	}

	sessionIDUUID, ok := sessionIDValue.(uuid.UUID)
	if !ok {
		utils.Fail(c, "Unauthorized", http.StatusUnauthorized, "invalid session ID type")
		return "", "", "", false
	}

	return userIDUUID.String(), emailStr, sessionIDUUID.String(), true
}
