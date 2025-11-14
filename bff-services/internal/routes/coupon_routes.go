package routes

import (
	"bff-services/internal/api/controllers"
	"bff-services/internal/cache"
	middleware "bff-services/internal/middlewares"

	"github.com/gin-gonic/gin"
)

// SetupCouponRoutes wires coupon-related endpoints.
func SetupCouponRoutes(api *gin.RouterGroup, controllers *controllers.Controllers, sessionCache *cache.SessionCache) {
	if controllers == nil || controllers.Coupon == nil || sessionCache == nil {
		return
	}

	coupons := api.Group("/coupons")
	coupons.Use(middleware.AuthRequired(sessionCache))
	{
		coupons.GET("", controllers.Coupon.ListAvailableCoupons)
		coupons.GET("/:id", controllers.Coupon.GetCoupon)
		coupons.POST("/validate", controllers.Coupon.ValidateCoupon)
		coupons.GET("/usage", controllers.Coupon.GetUserCouponUsage)
	}
}
