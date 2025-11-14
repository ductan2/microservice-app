package routes

import (
	"bff-services/internal/api/controllers"
	"bff-services/internal/cache"
	middleware "bff-services/internal/middlewares"

	"github.com/gin-gonic/gin"
)

// SetupSessionRoutes configures session management routes
func SetupSessionRoutes(api *gin.RouterGroup, controllers *controllers.Controllers, sessionCache *cache.SessionCache) {
	if controllers == nil || controllers.Session == nil {
		return
	}

	sessions := api.Group("/sessions")
	{
		// Public session routes
		sessions.POST("/user/:id", controllers.Session.ListByUserID)
	}

	// Protected session management routes
	if sessionCache != nil {
		protectedSessions := sessions.Group("")
		protectedSessions.Use(middleware.AuthRequired(sessionCache))
		{
			protectedSessions.GET("", controllers.Session.List)
			protectedSessions.DELETE("/:id", controllers.Session.Delete)
			protectedSessions.POST("/revoke-all", controllers.Session.RevokeAll)
		}
	}
}
