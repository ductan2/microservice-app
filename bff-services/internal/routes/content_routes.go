package routes

import (
	"bff-services/internal/api/controllers"

	"github.com/gin-gonic/gin"
)

// SetupContentRoutes configures content-related routes
func SetupContentRoutes(api *gin.RouterGroup, controllers *controllers.Controllers) {
	if controllers.Content == nil {
		return
	}

	content := api.Group("/content")
	{
		content.POST("/graphql", controllers.Content.ProxyGraphQL)
	}
}
