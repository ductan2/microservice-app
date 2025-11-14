package routes

import (
	"bff-services/internal/api/controllers"
	"bff-services/internal/cache"
	middleware "bff-services/internal/middlewares"

	"github.com/gin-gonic/gin"
)

// SetupContentRoutes configures content-related routes
func SetupContentRoutes(api *gin.RouterGroup, controllers *controllers.Controllers, sessionCache *cache.SessionCache) {
	if controllers == nil || controllers.Content == nil || sessionCache == nil {
		return
	}

	content := api.Group("/content")
	{
		content.POST("/graphql", controllers.Content.ProxyGraphQL)
	}

	// Protected media upload routes
	media := content.Group("/media")
	media.Use(middleware.AuthRequired(sessionCache))
	{
		media.POST("/images", controllers.Content.UploadImages)
	}
}
