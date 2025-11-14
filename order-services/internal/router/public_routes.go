package router

import (
	"net/http"

	"order-services/internal/controllers"
	"order-services/pkg/utils"

	"github.com/gin-gonic/gin"
)

func registerPublicRoutes(group *gin.RouterGroup, paymentCtrl *controllers.PaymentController) {
	if paymentCtrl != nil {
		group.GET("/stripe/config", paymentCtrl.GetStripeConfig)
		group.POST("/stripe/webhook", paymentCtrl.ProcessStripeWebhook)
		return
	}

	group.GET("/stripe/config", func(c *gin.Context) {
		utils.SuccessResponse(c, http.StatusOK, gin.H{
			"publishable_key": "",
		})
	})
}
