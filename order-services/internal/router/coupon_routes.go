package router

import (
	"net/http"

	"order-services/internal/controllers"
	"order-services/pkg/utils"

	"github.com/gin-gonic/gin"
)

func registerCouponRoutes(group *gin.RouterGroup, ctrl *controllers.CouponController) {
	if ctrl != nil {
		group.GET("/coupons", ctrl.ListAvailableCoupons)
		group.GET("/coupons/:id", ctrl.GetCoupon)
		group.POST("/coupons/validate", ctrl.ValidateCoupon)
		group.GET("/coupons/usage", ctrl.GetUserCouponUsage)
		return
	}

	group.GET("/coupons", func(c *gin.Context) {
		utils.SuccessResponse(c, http.StatusOK, gin.H{
			"coupons": []interface{}{},
			"total":   0,
		})
	})
	group.GET("/coupons/:id", func(c *gin.Context) {
		utils.ErrorResponse(c, http.StatusNotImplemented, "NOT_IMPLEMENTED", "Get coupon endpoint - implementation pending")
	})
	group.POST("/coupons/validate", func(c *gin.Context) {
		utils.ErrorResponse(c, http.StatusNotImplemented, "NOT_IMPLEMENTED", "Validate coupon endpoint - implementation pending")
	})
	group.GET("/coupons/usage", func(c *gin.Context) {
		utils.SuccessResponse(c, http.StatusOK, gin.H{
			"coupon_usages": []interface{}{},
		})
	})
}
