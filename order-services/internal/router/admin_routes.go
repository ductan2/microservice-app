package router

import (
	"net/http"

	"order-services/internal/controllers"
	"order-services/pkg/utils"

	"github.com/gin-gonic/gin"
)

func registerAdminRoutes(group *gin.RouterGroup, orderCtrl *controllers.OrderController, couponCtrl *controllers.CouponController) {
	registerAdminOrderRoutes(group, orderCtrl)
	registerAdminCouponRoutes(group, couponCtrl)
}

func registerAdminOrderRoutes(group *gin.RouterGroup, ctrl *controllers.OrderController) {
	if ctrl != nil {
		group.GET("/orders", ctrl.ListAllOrders)
		group.PUT("/orders/:id", ctrl.UpdateOrder)
		group.GET("/orders/stats", ctrl.GetOrderStats)
		return
	}

	group.GET("/orders", func(c *gin.Context) {
		utils.SuccessResponse(c, http.StatusOK, gin.H{
			"orders": []interface{}{},
			"total":  0,
		})
	})
	group.PUT("/orders/:id", func(c *gin.Context) {
		utils.ErrorResponse(c, http.StatusNotImplemented, "NOT_IMPLEMENTED", "Update order endpoint - implementation pending")
	})
	group.GET("/orders/stats", func(c *gin.Context) {
		utils.SuccessResponse(c, http.StatusOK, gin.H{
			"total_orders":        0,
			"total_revenue":       0,
			"pending_orders":      0,
			"completed_orders":    0,
			"cancelled_orders":    0,
			"failed_orders":       0,
			"refunded_orders":     0,
			"average_order_value": 0,
		})
	})
}

func registerAdminCouponRoutes(group *gin.RouterGroup, ctrl *controllers.CouponController) {
	if ctrl != nil {
		group.POST("/coupons", ctrl.CreateCoupon)
		group.PUT("/coupons/:id", ctrl.UpdateCoupon)
		group.DELETE("/coupons/:id", ctrl.DeleteCoupon)
		group.GET("/coupons/stats", ctrl.GetCouponStats)
		group.POST("/coupons/bulk", ctrl.CreateBulkCoupons)
		return
	}

	group.POST("/coupons", func(c *gin.Context) {
		utils.ErrorResponse(c, http.StatusNotImplemented, "NOT_IMPLEMENTED", "Create coupon endpoint - implementation pending")
	})
	group.PUT("/coupons/:id", func(c *gin.Context) {
		utils.ErrorResponse(c, http.StatusNotImplemented, "NOT_IMPLEMENTED", "Update coupon endpoint - implementation pending")
	})
	group.DELETE("/coupons/:id", func(c *gin.Context) {
		utils.ErrorResponse(c, http.StatusNotImplemented, "NOT_IMPLEMENTED", "Delete coupon endpoint - implementation pending")
	})
	group.GET("/coupons/stats", func(c *gin.Context) {
		utils.SuccessResponse(c, http.StatusOK, gin.H{
			"total_coupons":         0,
			"active_coupons":        0,
			"total_redemptions":     0,
			"total_discount_amount": 0,
		})
	})
	group.POST("/coupons/bulk", func(c *gin.Context) {
		utils.ErrorResponse(c, http.StatusNotImplemented, "NOT_IMPLEMENTED", "Create bulk coupons endpoint - implementation pending")
	})
}
