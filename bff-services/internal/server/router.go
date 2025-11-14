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
	QuizAttemptService  services.QuizAttemptService
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
	setupAPIRoutes(r, ctrl, deps)

	// Test route
	r.POST("/test-login", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "test works"})
	})

	return r
}

// setupAPIRoutes configures all API routes using the separated route files
func setupAPIRoutes(r *gin.Engine, controllers *controllers.Controllers, deps Deps) {
	api := r.Group("/api/v1")

	// Test route inside group
	api.POST("/test-inside", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "inside group works"})
	})

	// Setup routes from separate files
	routes.SetupAuthRoutes(api, controllers, deps.SessionCache)
	routes.SetupPasswordRoutes(api, controllers, deps.SessionCache)
	routes.SetupMFARoutes(api, controllers, deps.SessionCache)
	routes.SetupSessionRoutes(api, controllers, deps.SessionCache)
	routes.SetupContentRoutes(api, controllers, deps.SessionCache)
	routes.SetupLessonRoutes(api, controllers, deps.SessionCache)
	routes.SetupQuizAttemptRoutes(api, controllers, deps.SessionCache)
	routes.SetupUserRoutes(api, controllers, deps.SessionCache, deps.UserService)
	routes.SetupNotificationRoutes(api, controllers, deps.SessionCache)
	routes.SetupActivitySessionRoutes(api, controllers, deps.SessionCache)
	routes.SetupDashboardRoutes(api, controllers, deps.SessionCache)
}
