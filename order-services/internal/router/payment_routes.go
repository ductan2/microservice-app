package router

import (
	"net/http"

	"order-services/internal/controllers"
	"order-services/pkg/utils"

	"github.com/gin-gonic/gin"
)

func registerPaymentRoutes(group *gin.RouterGroup, ctrl *controllers.PaymentController) {
	if ctrl != nil {
		group.POST("/orders/:id/pay", ctrl.CreatePaymentIntent)
		group.POST("/payments/:payment_intent_id/confirm", ctrl.ConfirmPayment)
		group.GET("/payments/:payment_intent_id", ctrl.GetPayment)
		group.GET("/orders/:id/payment", ctrl.GetPaymentByOrderID)
		group.GET("/payment-methods", ctrl.GetPaymentMethods)
		group.GET("/payments", ctrl.GetPaymentHistory)
		return
	}

	group.POST("/orders/:id/pay", func(c *gin.Context) {
		utils.ErrorResponse(c, http.StatusNotImplemented, "NOT_IMPLEMENTED", "Create payment intent endpoint - implementation pending")
	})
	group.POST("/payments/:payment_intent_id/confirm", func(c *gin.Context) {
		utils.ErrorResponse(c, http.StatusNotImplemented, "NOT_IMPLEMENTED", "Confirm payment endpoint - implementation pending")
	})
	group.GET("/payments/:payment_intent_id", func(c *gin.Context) {
		utils.ErrorResponse(c, http.StatusNotImplemented, "NOT_IMPLEMENTED", "Get payment endpoint - implementation pending")
	})
	group.GET("/orders/:id/payment", func(c *gin.Context) {
		utils.ErrorResponse(c, http.StatusNotImplemented, "NOT_IMPLEMENTED", "Get payment by order endpoint - implementation pending")
	})
	group.GET("/payment-methods", func(c *gin.Context) {
		utils.SuccessResponse(c, http.StatusOK, gin.H{"payment_methods": []interface{}{}})
	})
	group.GET("/payments", func(c *gin.Context) {
		utils.SuccessResponse(c, http.StatusOK, gin.H{"payments": []interface{}{}, "total": 0})
	})
}
