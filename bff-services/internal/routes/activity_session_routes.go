package routes

import (
	"bff-services/internal/api/controllers"
	"bff-services/internal/cache"
	middleware "bff-services/internal/middlewares"

	"github.com/gin-gonic/gin"
)

// SetupActivitySessionRoutes configures activity session management routes
func SetupActivitySessionRoutes(api *gin.RouterGroup, controllers *controllers.Controllers, sessionCache *cache.SessionCache) {
	if controllers == nil || controllers.ActivitySession == nil || sessionCache == nil {
		return
	}

	// Protected activity session routes
	activitySessions := api.Group("/activity-sessions")
	activitySessions.Use(middleware.AuthRequired(sessionCache))
	{
		activitySessions.POST("/start", controllers.ActivitySession.StartSession)
		activitySessions.POST("/end", controllers.ActivitySession.EndSession)
		activitySessions.POST("/update", controllers.ActivitySession.UpdateSession)
		activitySessions.GET("", controllers.ActivitySession.GetSessions)
		activitySessions.GET("/stats", controllers.ActivitySession.GetSessionStats)
	}
}