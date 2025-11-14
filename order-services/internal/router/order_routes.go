package router

import (
	"net/http"

	"order-services/internal/controllers"
	"order-services/pkg/utils"

	"github.com/gin-gonic/gin"
)

func registerOrderRoutes(group *gin.RouterGroup, ctrl *controllers.OrderController) {
	if ctrl != nil {
		group.POST("/orders", ctrl.CreateOrder)
		group.GET("/orders", ctrl.ListOrders)
		group.GET("/orders/:id", ctrl.GetOrder)
		group.POST("/orders/:id/cancel", ctrl.CancelOrder)
		return
	}

	group.POST("/orders", func(c *gin.Context) {
		utils.ErrorResponse(c, http.StatusNotImplemented, "NOT_IMPLEMENTED", "Order creation endpoint - implementation pending")
	})
	group.GET("/orders", func(c *gin.Context) {
		utils.SuccessResponse(c, http.StatusOK, gin.H{
			"orders": []interface{}{},
			"total":  0,
		})
	})
	group.GET("/orders/:id", func(c *gin.Context) {
		utils.ErrorResponse(c, http.StatusNotImplemented, "NOT_IMPLEMENTED", "Get order endpoint - implementation pending")
	})
	group.POST("/orders/:id/cancel", func(c *gin.Context) {
		utils.ErrorResponse(c, http.StatusNotImplemented, "NOT_IMPLEMENTED", "Cancel order endpoint - implementation pending")
	})
}
