package routes

import (
	"bff-services/internal/api/controllers"
	"bff-services/internal/cache"
	middleware "bff-services/internal/middlewares"

	"github.com/gin-gonic/gin"
)

// SetupUserRoutes configures user management routes
func SetupUserRoutes(api *gin.RouterGroup, controllers *controllers.Controllers, sessionCache *cache.SessionCache) {
	if controllers.User == nil {
		return
	}

	// Public user routes

	// Protected user routes
	if sessionCache != nil {
		users := api.Group("/users")
		users.Use(middleware.AuthRequired(sessionCache))
		{
			users.GET("", controllers.User.ListUsersWithProgress)
			// Register specific routes before parameterized routes
			users.GET("/:id", controllers.User.GetUserById)
		}
	}
}
