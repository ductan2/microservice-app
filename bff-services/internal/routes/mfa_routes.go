package routes

import (
	"bff-services/internal/api/controllers"

	"github.com/gin-gonic/gin"
)

// SetupMFARoutes configures MFA-related routes
func SetupMFARoutes(api *gin.RouterGroup, controllers *controllers.Controllers) {
	if controllers.MFA == nil {
		return
	}

	mfa := api.Group("/mfa")
	{
		mfa.POST("/setup", controllers.MFA.Setup)
		mfa.POST("/verify", controllers.MFA.Verify)
		mfa.POST("/disable", controllers.MFA.Disable)
		mfa.GET("/methods", controllers.MFA.Methods)
	}
}
