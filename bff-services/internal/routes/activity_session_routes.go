package routes

import (
	"github.com/gin-gonic/gin"

	"bff-services/internal/api/controllers"
	middleware "bff-services/internal/middlewares"
	"bff-services/internal/cache"
)

// SetupActivitySessionRoutes sets up activity session routes
func SetupActivitySessionRoutes(api *gin.RouterGroup, controllers *controllers.Controllers, sessionCache *cache.SessionCache) {
	if controllers.ActivitySession == nil || sessionCache == nil {
		return
	}

	// Create activity session routes group
	activitySessionGroup := api.Group("/activity-sessions")

	// Apply AuthRequired middleware to all session routes
	activitySessionGroup.Use(middleware.AuthRequired(sessionCache))

	// Session management endpoints
	activitySessionGroup.POST("/start", controllers.ActivitySession.StartSession)
	activitySessionGroup.POST("/end", controllers.ActivitySession.EndSession)
	activitySessionGroup.POST("/update", controllers.ActivitySession.UpdateSession)
	activitySessionGroup.GET("", controllers.ActivitySession.GetSessions)
	activitySessionGroup.GET("/stats", controllers.ActivitySession.GetSessionStats)
}