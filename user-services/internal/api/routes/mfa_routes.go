package routes

import (
	"user-services/internal/api/controllers"

	"github.com/gin-gonic/gin"
)

func RegisterMFARoutes(router *gin.RouterGroup, controller *controllers.MFAController) {
	mfa := router.Group("/mfa")
	{
		mfa.POST("/setup", controller.SetupMFA)        // POST /mfa/setup
		mfa.POST("/verify", controller.VerifyMFASetup) // POST /mfa/verify
		mfa.POST("/disable", controller.DisableMFA)    // POST /mfa/disable
		mfa.GET("/methods", controller.GetMFAMethods)  // GET /mfa/methods
	}
}
