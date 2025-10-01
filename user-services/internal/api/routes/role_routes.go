package routes

import (
	"user-services/internal/api/controllers"

	"github.com/gin-gonic/gin"
)

func RegisterRoleRoutes(router *gin.RouterGroup, controller *controllers.RoleController) {
	// Role management (admin only)
	roles := router.Group("/roles")
	{
		roles.POST("", controller.CreateRole)       // POST /roles
		roles.GET("", controller.GetAllRoles)       // GET /roles
		roles.DELETE("/:id", controller.DeleteRole) // DELETE /roles/:id
	}

	// User role assignment (admin only)
	users := router.Group("/users")
	{
		users.POST("/:id/roles", controller.AssignRoleToUser)     // POST /users/:id/roles
		users.DELETE("/:id/roles", controller.RemoveRoleFromUser) // DELETE /users/:id/roles
	}
}
