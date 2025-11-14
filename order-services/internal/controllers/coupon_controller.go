package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"order-services/internal/dto"
	"order-services/internal/services"
	"order-services/pkg/utils"
)

// CouponController handles coupon-related HTTP requests
type CouponController struct {
	couponService services.CouponService
}

// NewCouponController creates a new coupon controller instance
func NewCouponController(couponService services.CouponService) *CouponController {
	return &CouponController{
		couponService: couponService,
	}
}

// ValidateCoupon validates a coupon code
// @Summary Validate coupon
// @Description Validates a coupon code and returns discount information
// @Tags coupons
// @Accept json
// @Produce json
// @Param request body dto.ValidateCouponRequest true "Coupon validation request"
// @Success 200 {object} dto.APIResponse{data=dto.ValidateCouponResponse}
// @Failure 400 {object} dto.APIResponse
// @Failure 404 {object} dto.APIResponse
// @Failure 500 {object} dto.APIResponse
// @Router /api/v1/coupons/validate [post]
func (c *CouponController) ValidateCoupon(ctx *gin.Context) {
	var req dto.ValidateCouponRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(ctx, err)
		return
	}

	// Validate coupon
	coupon, discountAmount, err := c.couponService.ValidateCoupon(ctx, req.Code, req.UserID, req.OrderAmount, req.CourseIDs)
	if err != nil {
		response := dto.ValidateCouponResponse{
			Valid:         false,
			DiscountAmount: 0,
			FinalAmount:    req.OrderAmount,
			Message:        err.Error(),
		}

		// Determine specific error type for better error handling
		if utils.IsNotFoundError(err) {
			utils.ErrorResponse(ctx, http.StatusNotFound, dto.ErrCodeCouponNotFound, "Coupon not found")
		} else if utils.IsExpiredError(err) {
			utils.SuccessResponse(ctx, http.StatusOK, response)
		} else if utils.IsInactiveError(err) {
			utils.SuccessResponse(ctx, http.StatusOK, response)
		} else if utils.IsValidationError(err) {
			utils.SuccessResponse(ctx, http.StatusOK, response)
		} else {
			utils.ErrorResponse(ctx, http.StatusInternalServerError, dto.ErrCodeInternalError, "Failed to validate coupon")
		}
		return
	}

	// Convert coupon to response
	var couponResponse dto.CouponResponse
	couponResponse.FromModel(coupon)

	// Build success response
	response := dto.ValidateCouponResponse{
		Valid:          true,
		Coupon:         &couponResponse,
		DiscountAmount: discountAmount,
		FinalAmount:    req.OrderAmount - discountAmount,
		Message:        "Coupon is valid",
	}

	utils.SuccessResponse(ctx, http.StatusOK, response)
}

// ListAvailableCoupons retrieves a list of available coupons for a user
// @Summary List available coupons
// @Description Retrieves a list of coupons currently available for the authenticated user
// @Tags coupons
// @Produce json
// @Param limit query int false "Number of items per page (default: 20, max: 100)"
// @Param offset query int false "Number of items to skip (default: 0)"
// @Param Authorization header string true "Bearer JWT token"
// @Success 200 {object} dto.APIResponse{data=dto.CouponListResponse}
// @Failure 400 {object} dto.APIResponse
// @Failure 401 {object} dto.APIResponse
// @Failure 500 {object} dto.APIResponse
// @Router /api/v1/coupons [get]
func (c *CouponController) ListAvailableCoupons(ctx *gin.Context) {
	// Parse pagination parameters
	var pagination dto.PaginationParams
	if err := ctx.ShouldBindQuery(&pagination); err != nil {
		utils.ValidationError(ctx, err)
		return
	}

	// Get user ID from JWT token
	userID, exists := ctx.Get("user_id")
	if !exists {
		utils.ErrorResponse(ctx, http.StatusUnauthorized, dto.ErrCodeUnauthorized, "User not authenticated")
		return
	}

	userUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		utils.ErrorResponse(ctx, http.StatusBadRequest, dto.ErrCodeBadRequest, "Invalid user ID")
		return
	}

	// Get available coupons
	coupons, err := c.couponService.ListAvailableCoupons(ctx, userUUID)
	if err != nil {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, dto.ErrCodeInternalError, "Failed to retrieve coupons")
		return
	}

	// Convert to response
	couponResponses := make([]dto.CouponResponse, len(coupons))
	for i, coupon := range coupons {
		couponResponses[i].FromModel(&coupon)

		// Add user-specific information if needed
		userUsage, err := c.couponService.GetUserCouponUsage(ctx, coupon.ID, userUUID)
		if err == nil {
			canRedeem := true
			message := "Available for use"

			// Check if user can redeem
			if coupon.PerUserLimit != nil && userUsage >= *coupon.PerUserLimit {
				canRedeem = false
				message = "Usage limit exceeded"
			}

			couponResponses[i].WithUserUsage(userUsage, canRedeem, message)
		}
	}

	// Apply pagination
	limit := pagination.Limit
	if limit <= 0 {
		limit = 20
	} else if limit > 100 {
		limit = 100
	}

	offset := pagination.Offset
	if pagination.Page > 0 {
		offset = (pagination.Page - 1) * limit
	}

	// Simple pagination for now (in production, you'd implement this in the service)
	total := int64(len(couponResponses))
	if offset >= len(couponResponses) {
		couponResponses = []dto.CouponResponse{}
	} else {
		end := offset + limit
		if end > len(couponResponses) {
			end = len(couponResponses)
		}
		couponResponses = couponResponses[offset:end]
	}

	response := dto.CouponListResponse{
		Coupons: couponResponses,
		Total:   total,
		Limit:   limit,
		Offset:  offset,
	}

	meta := dto.CalculatePagination(int(total), limit, offset)
	utils.SuccessResponseWithMeta(ctx, http.StatusOK, response, meta)
}

// GetCoupon retrieves a coupon by ID
// @Summary Get coupon by ID
// @Description Retrieves coupon information by ID
// @Tags coupons
// @Produce json
// @Param id path string true "Coupon ID"
// @Param user_id query string false "User ID for usage information (admin only)"
// @Param Authorization header string true "Bearer JWT token"
// @Success 200 {object} dto.APIResponse{data=dto.CouponResponse}
// @Failure 400 {object} dto.APIResponse
// @Failure 401 {object} dto.APIResponse
// @Failure 404 {object} dto.APIResponse
// @Failure 500 {object} dto.APIResponse
// @Router /api/v1/coupons/{id} [get]
func (c *CouponController) GetCoupon(ctx *gin.Context) {
	// Parse coupon ID
	couponID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		utils.ErrorResponse(ctx, http.StatusBadRequest, dto.ErrCodeBadRequest, "Invalid coupon ID")
		return
	}

	// Get coupon
	coupon, err := c.couponService.GetCoupon(ctx, couponID)
	if err != nil {
		if utils.IsNotFoundError(err) {
			utils.ErrorResponse(ctx, http.StatusNotFound, dto.ErrCodeCouponNotFound, "Coupon not found")
		} else {
			utils.ErrorResponse(ctx, http.StatusInternalServerError, dto.ErrCodeInternalError, "Failed to retrieve coupon")
		}
		return
	}

	// Convert to response
	var response dto.CouponResponse
	response.FromModel(coupon)

	// Add user-specific usage information if requested
	if userIDStr := ctx.Query("user_id"); userIDStr != "" {
		userUUID, err := uuid.Parse(userIDStr)
		if err != nil {
			utils.ErrorResponse(ctx, http.StatusBadRequest, dto.ErrCodeBadRequest, "Invalid user ID")
			return
		}

		userUsage, err := c.couponService.GetUserCouponUsage(ctx, couponID, userUUID)
		if err == nil {
			canRedeem := true
			message := "Available for use"

			// Check if user can redeem
			if coupon.PerUserLimit != nil && userUsage >= *coupon.PerUserLimit {
				canRedeem = false
				message = "Usage limit exceeded"
			}

			response.WithUserUsage(userUsage, canRedeem, message)
		}
	}

	utils.SuccessResponse(ctx, http.StatusOK, response)
}

// CreateCoupon creates a new coupon (admin only)
// @Summary Create coupon
// @Description Creates a new coupon (admin only)
// @Tags coupons
// @Accept json
// @Produce json
// @Param request body dto.CreateCouponRequest true "Coupon creation request"
// @Param Authorization header string true "Bearer JWT token"
// @Success 201 {object} dto.APIResponse{data=dto.CouponResponse}
// @Failure 400 {object} dto.APIResponse
// @Failure 401 {object} dto.APIResponse
// @Failure 403 {object} dto.APIResponse
// @Failure 500 {object} dto.APIResponse
// @Router /api/v1/admin/coupons [post]
func (c *CouponController) CreateCoupon(ctx *gin.Context) {
	// Check if user is admin
	if !utils.IsAdmin(ctx) {
		utils.ErrorResponse(ctx, http.StatusForbidden, dto.ErrCodeForbidden, "Admin access required")
		return
	}

	var req dto.CreateCouponRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(ctx, err)
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		utils.ErrorResponse(ctx, http.StatusBadRequest, dto.ErrCodeValidationFailed, err.Error())
		return
	}

	// This would need to be implemented in CouponService
	// For now, return not implemented
	utils.ErrorResponse(ctx, http.StatusNotImplemented, dto.ErrCodeInternalError, "CreateCoupon not implemented")
}

// UpdateCoupon updates an existing coupon (admin only)
// @Summary Update coupon
// @Description Updates an existing coupon (admin only)
// @Tags coupons
// @Accept json
// @Produce json
// @Param id path string true "Coupon ID"
// @Param request body dto.UpdateCouponRequest true "Coupon update request"
// @Param Authorization header string true "Bearer JWT token"
// @Success 200 {object} dto.APIResponse{data=dto.CouponResponse}
// @Failure 400 {object} dto.APIResponse
// @Failure 401 {object} dto.APIResponse
// @Failure 403 {object} dto.APIResponse
// @Failure 404 {object} dto.APIResponse
// @Failure 500 {object} dto.APIResponse
// @Router /api/v1/admin/coupons/{id} [put]
func (c *CouponController) UpdateCoupon(ctx *gin.Context) {
	// Check if user is admin
	if !utils.IsAdmin(ctx) {
		utils.ErrorResponse(ctx, http.StatusForbidden, dto.ErrCodeForbidden, "Admin access required")
		return
	}

	// Parse coupon ID
	_, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		utils.ErrorResponse(ctx, http.StatusBadRequest, dto.ErrCodeBadRequest, "Invalid coupon ID")
		return
	}

	var req dto.UpdateCouponRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(ctx, err)
		return
	}

	// This would need to be implemented in CouponService
	// For now, return not implemented
	utils.ErrorResponse(ctx, http.StatusNotImplemented, dto.ErrCodeInternalError, "UpdateCoupon not implemented")
}

// DeleteCoupon deletes a coupon (admin only)
// @Summary Delete coupon
// @Description Deletes a coupon (admin only)
// @Tags coupons
// @Produce json
// @Param id path string true "Coupon ID"
// @Param Authorization header string true "Bearer JWT token"
// @Success 200 {object} dto.APIResponse
// @Failure 400 {object} dto.APIResponse
// @Failure 401 {object} dto.APIResponse
// @Failure 403 {object} dto.APIResponse
// @Failure 404 {object} dto.APIResponse
// @Failure 500 {object} dto.APIResponse
// @Router /api/v1/admin/coupons/{id} [delete]
func (c *CouponController) DeleteCoupon(ctx *gin.Context) {
	// Check if user is admin
	if !utils.IsAdmin(ctx) {
		utils.ErrorResponse(ctx, http.StatusForbidden, dto.ErrCodeForbidden, "Admin access required")
		return
	}

	// Parse coupon ID
	_, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		utils.ErrorResponse(ctx, http.StatusBadRequest, dto.ErrCodeBadRequest, "Invalid coupon ID")
		return
	}

	// This would need to be implemented in CouponService
	// For now, return not implemented
	utils.ErrorResponse(ctx, http.StatusNotImplemented, dto.ErrCodeInternalError, "DeleteCoupon not implemented")
}

// GetUserCouponUsage retrieves a user's coupon usage history
// @Summary Get user coupon usage
// @Description Retrieves coupon usage history for the authenticated user
// @Tags coupons
// @Produce json
// @Param Authorization header string true "Bearer JWT token"
// @Success 200 {object} dto.APIResponse{data=dto.UserCouponUsageResponse}
// @Failure 401 {object} dto.APIResponse
// @Failure 500 {object} dto.APIResponse
// @Router /api/v1/coupons/usage [get]
func (c *CouponController) GetUserCouponUsage(ctx *gin.Context) {
	// Get user ID from JWT token
	userID, exists := ctx.Get("user_id")
	if !exists {
		utils.ErrorResponse(ctx, http.StatusUnauthorized, dto.ErrCodeUnauthorized, "User not authenticated")
		return
	}

	userUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		utils.ErrorResponse(ctx, http.StatusBadRequest, dto.ErrCodeBadRequest, "Invalid user ID")
		return
	}

	// This would need to be implemented in CouponService
	// For now, return empty response
	response := dto.UserCouponUsageResponse{
		UserID:       userUUID,
		CouponUsages: []dto.CouponUsageResponse{},
		TotalSavings: 0,
	}

	utils.SuccessResponse(ctx, http.StatusOK, response)
}

// GetCouponStats retrieves coupon statistics (admin only)
// @Summary Get coupon statistics
// @Description Retrieves coupon statistics for the given criteria (admin only)
// @Tags coupons
// @Produce json
// @Param start_date query string false "Filter by start date (YYYY-MM-DD)"
// @Param end_date query string false "Filter by end date (YYYY-MM-DD)"
// @Param Authorization header string true "Bearer JWT token"
// @Success 200 {object} dto.APIResponse{data=dto.CouponStatsResponse}
// @Failure 400 {object} dto.APIResponse
// @Failure 401 {object} dto.APIResponse
// @Failure 403 {object} dto.APIResponse
// @Failure 500 {object} dto.APIResponse
// @Router /api/v1/admin/coupons/stats [get]
func (c *CouponController) GetCouponStats(ctx *gin.Context) {
	// Check if user is admin
	if !utils.IsAdmin(ctx) {
		utils.ErrorResponse(ctx, http.StatusForbidden, dto.ErrCodeForbidden, "Admin access required")
		return
	}

	// This would need to be implemented in CouponService
	// For now, return empty stats
	stats := &dto.CouponStatsResponse{
		TotalCoupons:       0,
		ActiveCoupons:      0,
		TotalRedemptions:   0,
		TotalDiscountGiven: 0,
		AverageDiscount:    0,
		TopUsedCoupons:     []dto.CouponUsageStats{},
	}

	utils.SuccessResponse(ctx, http.StatusOK, stats)
}

// CreateBulkCoupons creates multiple coupons at once (admin only)
// @Summary Create bulk coupons
// @Description Creates multiple coupons at once (admin only)
// @Tags coupons
// @Accept json
// @Produce json
// @Param request body dto.BulkCouponRequest true "Bulk coupon creation request"
// @Param Authorization header string true "Bearer JWT token"
// @Success 201 {object} dto.APIResponse{data=dto.BulkCouponResponse}
// @Failure 400 {object} dto.APIResponse
// @Failure 401 {object} dto.APIResponse
// @Failure 403 {object} dto.APIResponse
// @Failure 500 {object} dto.APIResponse
// @Router /api/v1/admin/coupons/bulk [post]
func (c *CouponController) CreateBulkCoupons(ctx *gin.Context) {
	// Check if user is admin
	if !utils.IsAdmin(ctx) {
		utils.ErrorResponse(ctx, http.StatusForbidden, dto.ErrCodeForbidden, "Admin access required")
		return
	}

	var req dto.BulkCouponRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(ctx, err)
		return
	}

	// This would need to be implemented in CouponService
	// For now, return not implemented
	utils.ErrorResponse(ctx, http.StatusNotImplemented, dto.ErrCodeInternalError, "CreateBulkCoupons not implemented")
}