package routes

import (
	"bff-services/internal/api/controllers"
	"bff-services/internal/cache"
	middleware "bff-services/internal/middlewares"

	"github.com/gin-gonic/gin"
)

// SetupMFARoutes configures MFA-related routes
func SetupMFARoutes(api *gin.RouterGroup, controllers *controllers.Controllers, sessionCache *cache.SessionCache) {
	if controllers == nil || controllers.MFA == nil || sessionCache == nil {
		return
	}

	mfa := api.Group("/mfa")
	{
		mfa.GET("/methods", controllers.MFA.Methods)
	}

	// Protected MFA routes
	protectedMFA := mfa.Group("")
	protectedMFA.Use(middleware.AuthRequired(sessionCache))
	{
		protectedMFA.POST("/setup", controllers.MFA.Setup)
		protectedMFA.POST("/verify", controllers.MFA.Verify)
		protectedMFA.POST("/disable", controllers.MFA.Disable)
	}
}
