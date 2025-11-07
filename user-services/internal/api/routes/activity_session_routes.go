package routes

import (
	"github.com/gin-gonic/gin"

	"user-services/internal/api/controllers"
	"user-services/internal/api/middleware"
	"user-services/internal/cache"
)

// SetupActivitySessionRoutes sets up activity session routes
func RegisterActivitySessionRoutes(router *gin.RouterGroup, activitySessionController *controllers.ActivitySessionController, sessionCache *cache.SessionCache) {
	if activitySessionController == nil || sessionCache == nil {
		return
	}

	activitySessionGroup := router.Group("/activity-sessions")
	activitySessionGroup.Use(middleware.InternalAuthRequired())
	{
		activitySessionGroup.POST("/start", activitySessionController.StartSession)
		activitySessionGroup.POST("/end", activitySessionController.EndSession)
		activitySessionGroup.GET("", activitySessionController.GetSessions)
		activitySessionGroup.GET("/stats", activitySessionController.GetSessionStats)
		activitySessionGroup.POST("/update", activitySessionController.UpdateSession)
	}
}