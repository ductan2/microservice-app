package routes

import (
	"bff-services/internal/api/controllers"
	"bff-services/internal/cache"
	middleware "bff-services/internal/middlewares"

	"github.com/gin-gonic/gin"
)

// SetupOrderRoutes wires up order endpoints behind authentication.
func SetupOrderRoutes(api *gin.RouterGroup, controllers *controllers.Controllers, sessionCache *cache.SessionCache) {
	if controllers == nil || controllers.Order == nil || sessionCache == nil {
		return
	}

	orders := api.Group("/orders")
	orders.Use(middleware.AuthRequired(sessionCache))
	{
		orders.POST("", controllers.Order.CreateOrder)
		orders.GET("", controllers.Order.ListOrders)
		orders.GET("/:id", controllers.Order.GetOrder)
		orders.POST("/:id/cancel", controllers.Order.CancelOrder)
	}
}
