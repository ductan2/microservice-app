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
		users.POST("/register", controller.RegisterUser)       // POST /users/register
		users.POST("/login", controller.LoginUser)             // POST /users/login
		users.POST("/logout", controller.LogoutUser)           // POST /users/logout
		users.GET("/verify-email", controller.VerifyUserEmail) // GET /users/verify-email

		// Profile routes (authenticated)
		profile := users.Group("/profile")
		// Use InternalAuthRequired for internal communication from BFF
		profile.Use(middleware.InternalAuthRequired())
		{
			profile.GET("", controller.GetUserProfile)    // GET /users/profile
			profile.PUT("", controller.UpdateUserProfile) // PUT /users/profile
		}

		// User management routes (authenticated)
		users.Use(middleware.InternalAuthRequired())
		{
			users.GET("", controller.ListAllUsers)            // GET /users (list all users)
			users.GET("/:id", controller.GetUserByID)         // GET /users/:id (get specific user)
			users.PUT("/:id/role", controller.UpdateUserRole) // PUT /users/:id/role (update role)
		}
	}
}
