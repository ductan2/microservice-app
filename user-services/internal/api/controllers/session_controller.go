package controllers

import (
	"net/http"

	"user-services/internal/api/middleware"
	"user-services/internal/api/services"
	"user-services/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
	userIDValue, exists := ctx.Get(middleware.ContextUserIDKey())
	if !exists {
		utils.Fail(ctx, "Unauthorized", http.StatusUnauthorized, nil)
		return
	}

	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		utils.Fail(ctx, "Unauthorized", http.StatusUnauthorized, "invalid user context")
		return
	}

	sessions, err := c.sessionService.GetUserSessions(ctx.Request.Context(), userID)
	if err != nil {
		utils.Fail(ctx, "Failed to retrieve sessions", http.StatusInternalServerError, err.Error())
		return
	}

	userAgent := ctx.GetHeader("User-Agent")
	ipAddr := ctx.ClientIP()
	for i := range sessions {
		ipMatches := false
		if sessions[i].IPAddr != nil && ipAddr != "" {
			ipMatches = *sessions[i].IPAddr == ipAddr
		} else if sessions[i].IPAddr == nil && ipAddr == "" {
			ipMatches = true
		}
		if sessions[i].UserAgent == userAgent && ipMatches {
			sessions[i].IsCurrent = true
		}
	}

	utils.Success(ctx, sessions)
}

// RevokeSession godoc
// @Summary Revoke a session
// @Tags sessions
// @Param id path string true "Session ID"
// @Success 204
// @Router /sessions/{id} [delete]
func (c *SessionController) RevokeSession(ctx *gin.Context) {
	sessionID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		utils.Fail(ctx, "Invalid session ID", http.StatusBadRequest, err.Error())
		return
	}

	if err := c.sessionService.RevokeSession(ctx.Request.Context(), sessionID); err != nil {
		if err == services.ErrSessionNotFound {
			utils.Fail(ctx, "Session not found", http.StatusNotFound, nil)
			return
		}

		utils.Fail(ctx, "Failed to revoke session", http.StatusInternalServerError, err.Error())
		return
	}

	ctx.Status(http.StatusNoContent)
}

// RevokeAllSessions godoc
// @Summary Revoke all sessions
// @Tags sessions
// @Success 204
// @Router /sessions/revoke-all [post]
func (c *SessionController) RevokeAllSessions(ctx *gin.Context) {
	userIDValue, exists := ctx.Get(middleware.ContextUserIDKey())
	if !exists {
		utils.Fail(ctx, "Unauthorized", http.StatusUnauthorized, nil)
		return
	}

	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		utils.Fail(ctx, "Unauthorized", http.StatusUnauthorized, "invalid user context")
		return
	}

	if err := c.sessionService.RevokeAllUserSessions(ctx.Request.Context(), userID); err != nil {
		utils.Fail(ctx, "Failed to revoke sessions", http.StatusInternalServerError, err.Error())
		return
	}

	ctx.Status(http.StatusNoContent)
}

func (c *SessionController) ListSessionsByUserID(ctx *gin.Context) {
	userID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		utils.Fail(ctx, "Invalid user ID", http.StatusBadRequest, err.Error())
		return
	}

	sessions, err := c.sessionService.GetUserSessions(ctx.Request.Context(), userID)
	if err != nil {
		utils.Fail(ctx, "Failed to retrieve sessions", http.StatusInternalServerError, err.Error())
		return
	}

	utils.Success(ctx, sessions)
}
