package routes

import (
	"bff-services/internal/api/controllers"
	"bff-services/internal/cache"
	middleware "bff-services/internal/middlewares"

	"github.com/gin-gonic/gin"
)

// SetupPaymentRoutes wires payment-related endpoints.
func SetupPaymentRoutes(api *gin.RouterGroup, controllers *controllers.Controllers, sessionCache *cache.SessionCache) {
	if controllers == nil || controllers.Payment == nil {
		return
	}

	// Public Stripe config endpoint
	api.GET("/stripe/config", controllers.Payment.GetStripeConfig)

	if sessionCache == nil {
		return
	}

	protected := api.Group("/")
	protected.Use(middleware.AuthRequired(sessionCache))
	{
		protected.POST("/orders/:id/pay", controllers.Payment.CreatePaymentIntent)
		protected.POST("/payments/:payment_intent_id/confirm", controllers.Payment.ConfirmPayment)
		protected.GET("/orders/:id/payment", controllers.Payment.GetPaymentByOrderID)
		protected.GET("/payment-methods", controllers.Payment.GetPaymentMethods)
		protected.GET("/payments", controllers.Payment.GetPaymentHistory)
	}
}
