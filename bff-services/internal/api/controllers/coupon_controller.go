package controllers

import (
	"net/http"

	"bff-services/internal/api/dto"
	middleware "bff-services/internal/middlewares"
	"bff-services/internal/services"
	"bff-services/internal/utils"

	"github.com/gin-gonic/gin"
)

// CouponController proxies coupon requests to the order service.
type CouponController struct {
	couponService services.CouponService
}

// NewCouponController constructs a new CouponController.
func NewCouponController(couponService services.CouponService) *CouponController {
	return &CouponController{
		couponService: couponService,
	}
}

func (cpc *CouponController) ListAvailableCoupons(c *gin.Context) {
	if cpc.couponService == nil {
		utils.Fail(c, "Coupon service unavailable", http.StatusServiceUnavailable, "coupon service not configured")
		return
	}

	userID, email, sessionID, ok := middleware.GetUserContextFromMiddleware(c)
	if !ok {
		return
	}

	token := getOptionalBearerToken(c)
	if token == "" {
		utils.Fail(c, "Unauthorized", http.StatusUnauthorized, "missing bearer token")
		return
	}

	var query dto.CouponListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		utils.Fail(c, "Invalid query parameters", http.StatusBadRequest, err.Error())
		return
	}

	resp, err := cpc.couponService.ListAvailableCoupons(c.Request.Context(), token, userID, email, sessionID, query)
	if err != nil {
		utils.Fail(c, "Unable to fetch coupons", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}

func (cpc *CouponController) GetCoupon(c *gin.Context) {
	if cpc.couponService == nil {
		utils.Fail(c, "Coupon service unavailable", http.StatusServiceUnavailable, "coupon service not configured")
		return
	}

	userID, email, sessionID, ok := middleware.GetUserContextFromMiddleware(c)
	if !ok {
		return
	}

	token := getOptionalBearerToken(c)
	if token == "" {
		utils.Fail(c, "Unauthorized", http.StatusUnauthorized, "missing bearer token")
		return
	}

	couponID := c.Param("id")
	if couponID == "" {
		utils.Fail(c, "Coupon ID is required", http.StatusBadRequest, "missing coupon id")
		return
	}

	resp, err := cpc.couponService.GetCoupon(c.Request.Context(), token, userID, email, sessionID, couponID)
	if err != nil {
		utils.Fail(c, "Unable to fetch coupon", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}

func (cpc *CouponController) ValidateCoupon(c *gin.Context) {
	if cpc.couponService == nil {
		utils.Fail(c, "Coupon service unavailable", http.StatusServiceUnavailable, "coupon service not configured")
		return
	}

	userID, email, sessionID, ok := middleware.GetUserContextFromMiddleware(c)
	if !ok {
		return
	}

	token := getOptionalBearerToken(c)
	if token == "" {
		utils.Fail(c, "Unauthorized", http.StatusUnauthorized, "missing bearer token")
		return
	}

	var req dto.ValidateCouponRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, "Invalid request data", http.StatusBadRequest, err.Error())
		return
	}
	req.UserID = userID

	resp, err := cpc.couponService.ValidateCoupon(c.Request.Context(), token, userID, email, sessionID, req)
	if err != nil {
		utils.Fail(c, "Unable to validate coupon", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}

func (cpc *CouponController) GetUserCouponUsage(c *gin.Context) {
	if cpc.couponService == nil {
		utils.Fail(c, "Coupon service unavailable", http.StatusServiceUnavailable, "coupon service not configured")
		return
	}

	userID, email, sessionID, ok := middleware.GetUserContextFromMiddleware(c)
	if !ok {
		return
	}

	token := getOptionalBearerToken(c)
	if token == "" {
		utils.Fail(c, "Unauthorized", http.StatusUnauthorized, "missing bearer token")
		return
	}

	resp, err := cpc.couponService.GetUserCouponUsage(c.Request.Context(), token, userID, email, sessionID)
	if err != nil {
		utils.Fail(c, "Unable to fetch coupon usage", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}
