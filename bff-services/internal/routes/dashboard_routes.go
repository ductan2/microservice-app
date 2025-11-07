package routes

import (
	"bff-services/internal/api/controllers"
	"bff-services/internal/cache"
	middleware "bff-services/internal/middlewares"

	"github.com/gin-gonic/gin"
)

func SetupDashboardRoutes(api *gin.RouterGroup, controllers *controllers.Controllers, sessionCache *cache.SessionCache) {
	if controllers.Dashboard == nil || sessionCache == nil {
		return
	}

	dashboard := api.Group("/dashboard")
	dashboard.Use(middleware.AuthRequired(sessionCache))
	{
		dashboard.GET("/summary", controllers.Dashboard.GetSummary)
	}
}
