package router

import (
	"net/http"
	"time"

	"order-services/internal/controllers"
	"order-services/pkg/utils"

	"github.com/gin-gonic/gin"
)

func registerHealthRoutes(r *gin.Engine, orderCtrl *controllers.OrderController) {
	if orderCtrl != nil {
		r.GET("/health", orderCtrl.Health)
		return
	}

	r.GET("/health", func(ctx *gin.Context) {
		utils.SuccessResponse(ctx, http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now(),
			"version":   "1.0.0",
			"service":   "order-services",
			"checks": gin.H{
				"database": "unknown",
				"redis":    "unknown",
				"rabbitmq": "unknown",
			},
		})
	})
}
