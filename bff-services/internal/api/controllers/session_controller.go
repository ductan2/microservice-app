package controllers

import (
	"net/http"

	"bff-services/internal/services"
	"bff-services/internal/utils"

	"github.com/gin-gonic/gin"
)

type SessionController struct {
	userService services.UserService
}

func NewSessionController(userService services.UserService) *SessionController {
	return &SessionController{userService: userService}
}

func (s *SessionController) List(c *gin.Context) {
	token, ok := requireBearerToken(c)
	if !ok {
		return
	}

	resp, err := s.userService.GetSessions(c.Request.Context(), token)
	if err != nil {
		utils.Fail(c, "Unable to fetch sessions", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}

func (s *SessionController) Delete(c *gin.Context) {
	token, ok := requireBearerToken(c)
	if !ok {
		return
	}

	sessionID := c.Param("id")
	if sessionID == "" {
		utils.Fail(c, "Session ID is required", http.StatusBadRequest, "missing session id")
		return
	}

	resp, err := s.userService.DeleteSession(c.Request.Context(), token, sessionID)
	if err != nil {
		utils.Fail(c, "Unable to revoke session", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}

func (s *SessionController) RevokeAll(c *gin.Context) {
	token, ok := requireBearerToken(c)
	if !ok {
		return
	}

	resp, err := s.userService.RevokeAllSessions(c.Request.Context(), token)
	if err != nil {
		utils.Fail(c, "Unable to revoke sessions", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}
