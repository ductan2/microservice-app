package routes

import (
	"bff-services/internal/api/controllers"

	"github.com/gin-gonic/gin"
)

// SetupSessionRoutes configures session management routes
func SetupSessionRoutes(api *gin.RouterGroup, controllers *controllers.Controllers) {
	if controllers.Session == nil {
		return
	}

	sessions := api.Group("/sessions")
	{
		sessions.GET("", controllers.Session.List)
		sessions.DELETE("/:id", controllers.Session.Delete)
		sessions.POST("/revoke-all", controllers.Session.RevokeAll)
		sessions.POST("/user/:id", controllers.Session.ListByUserID)
	}
}
