package routes

import (
	"user-services/internal/api/controllers"
	"user-services/internal/api/middleware"
	"user-services/internal/cache"

	"github.com/gin-gonic/gin"
)

func RegisterMFARoutes(router *gin.RouterGroup, controller *controllers.MFAController, sessionCache *cache.SessionCache) {

	mfa := router.Group("/mfa")
	mfa.Use(middleware.AuthRequired(sessionCache))
	{
		mfa.POST("/setup", controller.SetupMFA)        // POST /mfa/setup
		mfa.POST("/verify", controller.VerifyMFASetup) // POST /mfa/verify
		mfa.POST("/disable", controller.DisableMFA)    // POST /mfa/disable
		mfa.GET("/methods", controller.GetMFAMethods)  // GET /mfa/methods
	}
}
