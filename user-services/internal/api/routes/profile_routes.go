package routes

import (
	"user-services/internal/api/controllers"

	"github.com/gin-gonic/gin"
)

func RegisterProfileRoutes(router *gin.RouterGroup, controller *controllers.ProfileController) {
	profile := router.Group("/profile")
	{
		profile.GET("", controller.GetProfile)    // GET /profile
		profile.PUT("", controller.UpdateProfile) // PUT /profile
	}
}
