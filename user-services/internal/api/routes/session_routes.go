package routes

import (
	"user-services/internal/api/controllers"
	"user-services/internal/api/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterSessionRoutes(router *gin.RouterGroup, controller *controllers.SessionController) {
	sessions := router.Group("/sessions")
	sessions.Use(middleware.AuthRequired())
	{
		sessions.GET("", controller.GetActiveSessions)             // GET /sessions
		sessions.DELETE("/:id", controller.RevokeSession)          // DELETE /sessions/:id
		sessions.POST("/revoke-all", controller.RevokeAllSessions) // POST /sessions/revoke-all
	}
}
