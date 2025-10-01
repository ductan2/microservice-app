package routes

import (
	"user-services/internal/api/controllers"

	"github.com/gin-gonic/gin"
)

func RegisterAuditRoutes(router *gin.RouterGroup, controller *controllers.AuditController) {
	audit := router.Group("/audit")
	{
		audit.GET("/users/:id", controller.GetUserAuditLogs) // GET /audit/users/:id
		audit.GET("/actions", controller.GetActionAuditLogs) // GET /audit/actions
	}
}
