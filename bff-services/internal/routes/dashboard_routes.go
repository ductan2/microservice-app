package routes

import (
	"bff-services/internal/api/controllers"
	"bff-services/internal/cache"
	middleware "bff-services/internal/middlewares"

	"github.com/gin-gonic/gin"
)

// SetupDashboardRoutes configures dashboard-related routes
func SetupDashboardRoutes(api *gin.RouterGroup, controllers *controllers.Controllers, sessionCache *cache.SessionCache) {
	if controllers == nil || controllers.Dashboard == nil || sessionCache == nil {
		return
	}

	// Protected dashboard routes
	dashboard := api.Group("/dashboard")
	dashboard.Use(middleware.AuthRequired(sessionCache))
	{
		dashboard.GET("/summary", controllers.Dashboard.GetSummary)
	}
}
