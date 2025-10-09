package routes

import (
	"bff-services/internal/api/controllers"

	"github.com/gin-gonic/gin"
)

// SetupPasswordRoutes configures password management routes
func SetupPasswordRoutes(api *gin.RouterGroup, controllers *controllers.Controllers) {
	if controllers.Password == nil {
		return
	}

	api.POST("/password/reset/request", controllers.Password.RequestReset)
	api.POST("/password/reset/confirm", controllers.Password.ConfirmReset)
	api.POST("/password/change", controllers.Password.ChangePassword)
}
