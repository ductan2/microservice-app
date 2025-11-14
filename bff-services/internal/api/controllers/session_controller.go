package controllers

import (
	"net/http"

	middleware "bff-services/internal/middlewares"
	"bff-services/internal/services"
	"bff-services/internal/utils"

	"github.com/gin-gonic/gin"
)

// SessionController handles user session management operations.
type SessionController struct {
	userService services.UserService
}

// NewSessionController constructs a new SessionController.
func NewSessionController(userService services.UserService) *SessionController {
	return &SessionController{userService: userService}
}

func (s *SessionController) List(c *gin.Context) {
	userID, email, sessionID, ok := middleware.GetUserContextFromMiddleware(c)
	if !ok {
		return
	}

	resp, err := s.userService.GetSessions(c.Request.Context(), userID, email, sessionID)
	if err != nil {
		utils.Fail(c, "Unable to fetch sessions", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}

func (s *SessionController) Delete(c *gin.Context) {
	userID, email, sessionID, ok := middleware.GetUserContextFromMiddleware(c)
	if !ok {
		return
	}

	deleteSessionID := c.Param("id")
	if deleteSessionID == "" {
		utils.Fail(c, "Session ID is required", http.StatusBadRequest, "missing session id")
		return
	}

	resp, err := s.userService.DeleteSession(c.Request.Context(), userID, email, sessionID, deleteSessionID)
	if err != nil {
		utils.Fail(c, "Unable to revoke session", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}

func (s *SessionController) RevokeAll(c *gin.Context) {
	userID, email, sessionID, ok := middleware.GetUserContextFromMiddleware(c)
	if !ok {
		return
	}

	resp, err := s.userService.RevokeAllSessions(c.Request.Context(), userID, email, sessionID)
	if err != nil {
		utils.Fail(c, "Unable to revoke sessions", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}

func (s *SessionController) ListByUserID(c *gin.Context) {
	userID, email, sessionID, ok := middleware.GetUserContextFromMiddleware(c)
	if !ok {
		return
	}

	targetUserID := c.Param("id")
	if targetUserID == "" {
		utils.Fail(c, "User ID is required", http.StatusBadRequest, "missing user ID")
		return
	}

	resp, err := s.userService.ListSessionsByUserID(c.Request.Context(), targetUserID, userID, email, sessionID)
	if err != nil {
		utils.Fail(c, "Unable to fetch sessions", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}
