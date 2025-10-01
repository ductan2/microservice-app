package controllers

import (
	"user-services/internal/api/services"

	"github.com/gin-gonic/gin"
)

type SessionController struct {
	sessionService services.SessionService
}

func NewSessionController(sessionService services.SessionService) *SessionController {
	return &SessionController{
		sessionService: sessionService,
	}
}

// GetActiveSessions godoc
// @Summary Get active sessions
// @Tags sessions
// @Produce json
// @Success 200 {array} dto.SessionResponse
// @Router /sessions [get]
func (c *SessionController) GetActiveSessions(ctx *gin.Context) {
	// TODO: implement - get user sessions
}

// RevokeSession godoc
// @Summary Revoke a session
// @Tags sessions
// @Param id path string true "Session ID"
// @Success 204
// @Router /sessions/{id} [delete]
func (c *SessionController) RevokeSession(ctx *gin.Context) {
	// TODO: implement
}

// RevokeAllSessions godoc
// @Summary Revoke all sessions
// @Tags sessions
// @Success 204
// @Router /sessions/revoke-all [post]
func (c *SessionController) RevokeAllSessions(ctx *gin.Context) {
	// TODO: implement
}
