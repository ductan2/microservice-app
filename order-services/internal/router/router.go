package router

import (
	"order-services/internal/controllers"
	"order-services/internal/middleware"

	"github.com/gin-gonic/gin"
)

// Dependencies groups the controller dependencies required by the router.
type Dependencies struct {
	OrderController   *controllers.OrderController
	PaymentController *controllers.PaymentController
	CouponController  *controllers.CouponController
	JWTSecret         string
}

// NewRouter initializes the Gin router with all routes and middleware.
func NewRouter(deps Dependencies) *gin.Engine {
	r := gin.New()

	// Global middleware
	r.Use(middleware.Logging())
	r.Use(gin.Recovery())
	r.Use(middleware.CORS())
	r.Use(middleware.RequestID())

	// Health route
	registerHealthRoutes(r, deps.OrderController)

	// API v1 routes
	v1 := r.Group("/api/v1")

	// Public routes
	registerPublicRoutes(v1, deps.PaymentController)

	// Protected routes requiring authentication
	protected := v1.Group("/")
	protected.Use(middleware.JWTAuth(deps.JWTSecret))

	registerOrderRoutes(protected, deps.OrderController)
	registerPaymentRoutes(protected, deps.PaymentController)
	registerCouponRoutes(protected, deps.CouponController)

	// Admin routes
	admin := protected.Group("/admin")
	admin.Use(middleware.AdminOnly())
	registerAdminRoutes(admin, deps.OrderController, deps.CouponController)

	return r
}
