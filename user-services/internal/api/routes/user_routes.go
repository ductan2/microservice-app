package routes

import (
	"user-services/internal/api/controllers"
	"user-services/internal/api/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterUserRoutes registers all user-related routes (auth, profile, user management)
func RegisterUserRoutes(router *gin.RouterGroup, controller *controllers.UserController) {
	users := router.Group("/users")
	{
		// Authentication routes (public)
		users.POST("/register", controller.RegisterUser)
		users.POST("/login", controller.LoginUser)
		users.POST("/logout", controller.LogoutUser)
		users.GET("/verify-email", controller.VerifyUserEmail)

		// Profile routes (authenticated)
		profile := users.Group("/profile")
		// Use InternalAuthRequired for internal communication from BFF
		profile.Use(middleware.InternalAuthRequired())
		{
			profile.GET("", controller.GetUserProfile)
			profile.PUT("", controller.UpdateUserProfile)
		}

		// User management routes (authenticated)
		users.Use(middleware.InternalAuthRequired())
		{
			users.GET("", controller.ListAllUsers)
			users.GET("/:id", controller.GetUserByID)
			users.PUT("/:id/role", controller.UpdateUserRole)
			users.POST("/:id/lock", controller.LockAccount)
			users.POST("/:id/unlock", controller.UnlockAccount)
			users.DELETE("/:id/delete", controller.SoftDeleteAccount)
			users.POST("/:id/restore", controller.RestoreAccount)
		}
	}
}
