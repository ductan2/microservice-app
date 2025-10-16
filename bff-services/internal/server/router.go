package server

import (
	"bff-services/internal/api/controllers"
	"bff-services/internal/cache"
	"bff-services/internal/routes"
	"bff-services/internal/services"

	"github.com/gin-gonic/gin"
)

type Deps struct {
	UserService         services.UserService
	ContentService      services.ContentService
	LessonService       services.LessonService
	NotificationService services.NotificationService
	SessionCache        *cache.SessionCache
}

func NewRouter(deps Deps) *gin.Engine {
	r := gin.New()

	// Setup global middlewares
	setupGlobalMiddlewares(r)

	// Setup health check
	r.GET("/health", controllers.Health)

	// Initialize controllers
	ctrl := initControllers(deps)

	// Setup API routes
	setupAPIRoutes(r, ctrl, deps.SessionCache)

	// Test route
	r.POST("/test-login", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "test works"})
	})

	return r
}

// setupAPIRoutes configures all API routes using the separated route files
func setupAPIRoutes(r *gin.Engine, controllers *controllers.Controllers, sessionCache *cache.SessionCache) {
	api := r.Group("/api/v1")

	// Test route inside group
	api.POST("/test-inside", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "inside group works"})
	})

	// Health check

	// Setup routes from separate files
	routes.SetupAuthRoutes(api, controllers, sessionCache)
	routes.SetupPasswordRoutes(api, controllers)
	routes.SetupMFARoutes(api, controllers)
	routes.SetupSessionRoutes(api, controllers)
	routes.SetupContentRoutes(api, controllers)
	routes.SetupLessonRoutes(api, controllers, sessionCache)
	routes.SetupUserRoutes(api, controllers, sessionCache)
	routes.SetupNotificationRoutes(api, controllers)
}
