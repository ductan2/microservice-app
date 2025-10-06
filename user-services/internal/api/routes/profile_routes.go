package routes

import (
	"user-services/internal/api/controllers"
	"user-services/internal/api/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterProfileRoutes(router *gin.RouterGroup, controller *controllers.ProfileController) {
	profile := router.Group("/profile")
	// Use InternalAuthRequired for internal communication from BFF
	profile.Use(middleware.InternalAuthRequired())
	{
		profile.GET("", controller.GetProfile)    // GET /profile
		profile.PUT("", controller.UpdateProfile) // PUT /profile
	}
}
