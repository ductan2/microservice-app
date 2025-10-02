package routes

import (
	"user-services/internal/api/controllers"
	"user-services/internal/api/middleware"
	"user-services/internal/cache"

	"github.com/gin-gonic/gin"
)

func RegisterProfileRoutes(router *gin.RouterGroup, controller *controllers.ProfileController, sessionCache *cache.SessionCache) {
	profile := router.Group("/profile")
	profile.Use(middleware.AuthRequired(sessionCache))
	{
		profile.GET("", controller.GetProfile)    // GET /profile
		profile.PUT("", controller.UpdateProfile) // PUT /profile
	}
}
