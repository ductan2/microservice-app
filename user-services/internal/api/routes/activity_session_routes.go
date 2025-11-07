package routes

import (
	"github.com/gin-gonic/gin"

	"user-services/internal/api/controllers"
	"user-services/internal/middleware"
)

// SetupActivitySessionRoutes sets up activity session routes
func SetupActivitySessionRoutes(router *gin.Engine, activitySessionController *controllers.ActivitySessionController) {
	// Create activity session routes group
	activitySessionGroup := router.Group("/api/v1/sessions")

	// Apply InternalAuthRequired middleware to all session routes
	activitySessionGroup.Use(middleware.InternalAuthRequired())

	// Session start and end endpoints
	activitySessionGroup.POST("/start", activitySessionController.StartSession)
	activitySessionGroup.POST("/end", activitySessionController.EndSession)

	// Session management endpoints
	activitySessionGroup.GET("", activitySessionController.GetSessions)
	activitySessionGroup.GET("/stats", activitySessionController.GetSessionStats)
	activitySessionGroup.POST("/update", activitySessionController.UpdateSession)
}