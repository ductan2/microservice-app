package controllers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"order-services/internal/dto"
	"order-services/internal/services"
	"order-services/pkg/utils"
)

// OrderController handles order-related HTTP requests
type OrderController struct {
	orderService services.OrderService
}

// NewOrderController creates a new order controller instance
func NewOrderController(orderService services.OrderService) *OrderController {
	return &OrderController{
		orderService: orderService,
	}
}

// CreateOrder creates a new order
// @Summary Create a new order
// @Description Creates a new order with the provided items and optional coupon
// @Tags orders
// @Accept json
// @Produce json
// @Param request body dto.CreateOrderRequest true "Order creation request"
// @Param Authorization header string true "Bearer JWT token"
// @Success 201 {object} dto.APIResponse{data=dto.OrderResponse}
// @Failure 400 {object} dto.APIResponse
// @Failure 401 {object} dto.APIResponse
// @Failure 404 {object} dto.APIResponse
// @Failure 500 {object} dto.APIResponse
// @Router /api/v1/orders [post]
func (c *OrderController) CreateOrder(ctx *gin.Context) {
	var req dto.CreateOrderRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(ctx, err)
		return
	}

	// Get user ID from JWT token (set by auth middleware)
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

	// Create service request
	createReq := &services.CreateOrderRequest{
		UserID:        userUUID,
		Items:         make([]services.OrderItemRequest, len(req.Items)),
		CouponCode:    req.CouponCode,
		CustomerEmail: req.CustomerEmail,
		CustomerName:  req.CustomerName,
		Metadata:      req.Metadata,
	}

	// Convert items
	for i, item := range req.Items {
		createReq.Items[i] = services.OrderItemRequest{
			CourseID:      item.CourseID,
			Quantity:      item.Quantity,
			PriceSnapshot: *item.PriceSnapshot,
		}
	}

	// Create order
	order, err := c.orderService.CreateOrder(ctx, createReq)
	if err != nil {
		if utils.IsValidationError(err) {
			utils.ErrorResponse(ctx, http.StatusBadRequest, dto.ErrCodeValidationFailed, err.Error())
		} else if utils.IsNotFoundError(err) {
			utils.ErrorResponse(ctx, http.StatusNotFound, dto.ErrCodeNotFound, err.Error())
		} else {
			utils.ErrorResponse(ctx, http.StatusInternalServerError, dto.ErrCodeInternalError, "Failed to create order")
		}
		return
	}

	// Convert to response
	var response dto.OrderResponse
	response.FromModel(order)

	utils.SuccessResponse(ctx, http.StatusCreated, response)
}

// GetOrder retrieves an order by ID
// @Summary Get an order by ID
// @Description Retrieves an order by its ID for the authenticated user
// @Tags orders
// @Produce json
// @Param id path string true "Order ID"
// @Param Authorization header string true "Bearer JWT token"
// @Success 200 {object} dto.APIResponse{data=dto.OrderResponse}
// @Failure 400 {object} dto.APIResponse
// @Failure 401 {object} dto.APIResponse
// @Failure 404 {object} dto.APIResponse
// @Failure 500 {object} dto.APIResponse
// @Router /api/v1/orders/{id} [get]
func (c *OrderController) GetOrder(ctx *gin.Context) {
	// Parse order ID
	orderID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		utils.ErrorResponse(ctx, http.StatusBadRequest, dto.ErrCodeBadRequest, "Invalid order ID")
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

	// Get order
	order, err := c.orderService.GetOrder(ctx, orderID, userUUID)
	if err != nil {
		if utils.IsNotFoundError(err) {
			utils.ErrorResponse(ctx, http.StatusNotFound, dto.ErrCodeOrderNotFound, "Order not found")
		} else if utils.IsUnauthorizedError(err) {
			utils.ErrorResponse(ctx, http.StatusForbidden, dto.ErrCodeForbidden, "Access denied")
		} else {
			utils.ErrorResponse(ctx, http.StatusInternalServerError, dto.ErrCodeInternalError, "Failed to retrieve order")
		}
		return
	}

	// Convert to response
	var response dto.OrderResponse
	response.FromModel(order)

	utils.SuccessResponse(ctx, http.StatusOK, response)
}

// ListOrders retrieves a paginated list of user's orders
// @Summary List user orders
// @Description Retrieves a paginated list of orders for the authenticated user
// @Tags orders
// @Produce json
// @Param limit query int false "Number of items per page (default: 20, max: 100)"
// @Param offset query int false "Number of items to skip (default: 0)"
// @Param page query int false "Page number (alternative to offset)"
// @Param status query string false "Filter by status"
// @Param sort_by query string false "Sort field (default: created_at)"
// @Param sort_order query string false "Sort order (asc or desc, default: desc)"
// @Param Authorization header string true "Bearer JWT token"
// @Success 200 {object} dto.APIResponse{data=dto.OrderListResponse}
// @Failure 400 {object} dto.APIResponse
// @Failure 401 {object} dto.APIResponse
// @Failure 500 {object} dto.APIResponse
// @Router /api/v1/orders [get]
func (c *OrderController) ListOrders(ctx *gin.Context) {
	// Parse pagination parameters
	var pagination dto.PaginationParams
	if err := ctx.ShouldBindQuery(&pagination); err != nil {
		utils.ValidationError(ctx, err)
		return
	}

	// Set defaults
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

	// Get orders
	orders, total, err := c.orderService.ListUserOrders(ctx, userUUID, limit, offset)
	if err != nil {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, dto.ErrCodeInternalError, "Failed to retrieve orders")
		return
	}

	// Convert to response
	orderResponses := make([]dto.OrderResponse, len(orders))
	for i, order := range orders {
		orderResponses[i].FromModel(&order)
	}

	// Create paginated response
	meta := dto.CalculatePagination(int(total), limit, offset)
	response := dto.OrderListResponse{
		Orders: orderResponses,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}

	utils.SuccessResponseWithMeta(ctx, http.StatusOK, response, meta)
}

// CancelOrder cancels an order
// @Summary Cancel an order
// @Description Cancels an order if it's in a cancellable state
// @Tags orders
// @Accept json
// @Produce json
// @Param id path string true "Order ID"
// @Param request body dto.CancelOrderRequest true "Cancellation request"
// @Param Authorization header string true "Bearer JWT token"
// @Success 200 {object} dto.APIResponse
// @Failure 400 {object} dto.APIResponse
// @Failure 401 {object} dto.APIResponse
// @Failure 404 {object} dto.APIResponse
// @Failure 409 {object} dto.APIResponse
// @Failure 500 {object} dto.APIResponse
// @Router /api/v1/orders/{id}/cancel [post]
func (c *OrderController) CancelOrder(ctx *gin.Context) {
	// Parse order ID
	orderID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		utils.ErrorResponse(ctx, http.StatusBadRequest, dto.ErrCodeBadRequest, "Invalid order ID")
		return
	}

	var req dto.CancelOrderRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
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

	// Cancel order
	err = c.orderService.CancelOrder(ctx, orderID, userUUID, req.Reason)
	if err != nil {
		if utils.IsNotFoundError(err) {
			utils.ErrorResponse(ctx, http.StatusNotFound, dto.ErrCodeOrderNotFound, "Order not found")
		} else if utils.IsUnauthorizedError(err) {
			utils.ErrorResponse(ctx, http.StatusForbidden, dto.ErrCodeForbidden, "Access denied")
		} else if utils.IsConflictError(err) {
			utils.ErrorResponse(ctx, http.StatusConflict, dto.ErrCodeOrderCannotCancel, err.Error())
		} else {
			utils.ErrorResponse(ctx, http.StatusInternalServerError, dto.ErrCodeInternalError, "Failed to cancel order")
		}
		return
	}

	utils.SuccessResponse(ctx, http.StatusOK, nil, "Order cancelled successfully")
}

// UpdateOrder updates an order (admin only)
// @Summary Update an order
// @Description Updates an order's status or metadata (admin only)
// @Tags orders
// @Accept json
// @Produce json
// @Param id path string true "Order ID"
// @Param request body dto.UpdateOrderRequest true "Update request"
// @Param Authorization header string true "Bearer JWT token"
// @Success 200 {object} dto.APIResponse{data=dto.OrderResponse}
// @Failure 400 {object} dto.APIResponse
// @Failure 401 {object} dto.APIResponse
// @Failure 403 {object} dto.APIResponse
// @Failure 404 {object} dto.APIResponse
// @Failure 500 {object} dto.APIResponse
// @Router /api/v1/orders/{id} [put]
func (c *OrderController) UpdateOrder(ctx *gin.Context) {
	// Check if user is admin
	if !utils.IsAdmin(ctx) {
		utils.ErrorResponse(ctx, http.StatusForbidden, dto.ErrCodeForbidden, "Admin access required")
		return
	}

	// Parse order ID
	orderID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		utils.ErrorResponse(ctx, http.StatusBadRequest, dto.ErrCodeBadRequest, "Invalid order ID")
		return
	}

	var req dto.UpdateOrderRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(ctx, err)
		return
	}

	// Update order status if provided
	if req.Status != nil {
		reason := ""
		if req.FailureReason != nil {
			reason = *req.FailureReason
		}

		err = c.orderService.UpdateOrderStatus(ctx, orderID, *req.Status, reason)
		if err != nil {
			if utils.IsNotFoundError(err) {
				utils.ErrorResponse(ctx, http.StatusNotFound, dto.ErrCodeOrderNotFound, "Order not found")
			} else if utils.IsValidationError(err) {
				utils.ErrorResponse(ctx, http.StatusBadRequest, dto.ErrCodeInvalidOrderStatus, err.Error())
			} else {
				utils.ErrorResponse(ctx, http.StatusInternalServerError, dto.ErrCodeInternalError, "Failed to update order")
			}
			return
		}
	}

	// Get updated order
	order, err := c.orderService.GetOrder(ctx, orderID, uuid.Nil) // Admin can access any order
	if err != nil {
		if utils.IsNotFoundError(err) {
			utils.ErrorResponse(ctx, http.StatusNotFound, dto.ErrCodeOrderNotFound, "Order not found")
		} else {
			utils.ErrorResponse(ctx, http.StatusInternalServerError, dto.ErrCodeInternalError, "Failed to retrieve updated order")
		}
		return
	}

	// Convert to response
	var response dto.OrderResponse
	response.FromModel(order)

	utils.SuccessResponse(ctx, http.StatusOK, response)
}

// GetOrderStats retrieves order statistics (admin only)
// @Summary Get order statistics
// @Description Retrieves order statistics for the given criteria (admin only)
// @Tags orders
// @Produce json
// @Param user_id query string false "Filter by user ID"
// @Param start_date query string false "Filter by start date (YYYY-MM-DD)"
// @Param end_date query string false "Filter by end date (YYYY-MM-DD)"
// @Param Authorization header string true "Bearer JWT token"
// @Success 200 {object} dto.APIResponse{data=dto.OrderStatsResponse}
// @Failure 400 {object} dto.APIResponse
// @Failure 401 {object} dto.APIResponse
// @Failure 403 {object} dto.APIResponse
// @Failure 500 {object} dto.APIResponse
// @Router /api/v1/admin/orders/stats [get]
func (c *OrderController) GetOrderStats(ctx *gin.Context) {
	// Check if user is admin
	if !utils.IsAdmin(ctx) {
		utils.ErrorResponse(ctx, http.StatusForbidden, dto.ErrCodeForbidden, "Admin access required")
		return
	}

	// Parse query parameters
	var userID *uuid.UUID
	if userIDStr := ctx.Query("user_id"); userIDStr != "" {
		parsedID, err := uuid.Parse(userIDStr)
		if err != nil {
			utils.ErrorResponse(ctx, http.StatusBadRequest, dto.ErrCodeBadRequest, "Invalid user ID")
			return
		}
		userID = &parsedID
	}

	// Parse date range (implementation depends on order service)
	// This is a simplified version - in production, you'd parse date strings properly
	// var timeRange *services.TimeRange
	if startDate := ctx.Query("start_date"); startDate != "" {
		// Parse start date
		if endDate := ctx.Query("end_date"); endDate != "" {
			// Parse end date
			// timeRange = &services.TimeRange{...}
		}
	}

	// Placeholder assignment until stats filtering implementation uses userID.
	_ = userID

	// Get statistics (this method needs to be implemented in OrderService)
	// For now, return empty stats
	stats := &dto.OrderStatsResponse{
		TotalOrders:       0,
		TotalRevenue:      0,
		PendingOrders:     0,
		CompletedOrders:   0,
		CancelledOrders:   0,
		FailedOrders:      0,
		RefundedOrders:    0,
		AverageOrderValue: 0,
	}

	utils.SuccessResponse(ctx, http.StatusOK, stats)
}

// ListAllOrders retrieves all orders with pagination (admin only)
// @Summary List all orders
// @Description Retrieves all orders with pagination (admin only)
// @Tags orders
// @Produce json
// @Param limit query int false "Number of items per page (default: 20, max: 100)"
// @Param offset query int false "Number of items to skip (default: 0)"
// @Param status query string false "Filter by status"
// @Param user_id query string false "Filter by user ID"
// @Param Authorization header string true "Bearer JWT token"
// @Success 200 {object} dto.APIResponse{data=dto.OrderListResponse}
// @Failure 400 {object} dto.APIResponse
// @Failure 401 {object} dto.APIResponse
// @Failure 403 {object} dto.APIResponse
// @Failure 500 {object} dto.APIResponse
// @Router /api/v1/admin/orders [get]
func (c *OrderController) ListAllOrders(ctx *gin.Context) {
	// Check if user is admin
	if !utils.IsAdmin(ctx) {
		utils.ErrorResponse(ctx, http.StatusForbidden, dto.ErrCodeForbidden, "Admin access required")
		return
	}

	// Parse pagination parameters
	var pagination dto.PaginationParams
	if err := ctx.ShouldBindQuery(&pagination); err != nil {
		utils.ValidationError(ctx, err)
		return
	}

	// Set defaults
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

	// Parse user ID filter
	var userID *uuid.UUID
	if userIDStr := ctx.Query("user_id"); userIDStr != "" {
		parsedID, err := uuid.Parse(userIDStr)
		if err != nil {
			utils.ErrorResponse(ctx, http.StatusBadRequest, dto.ErrCodeBadRequest, "Invalid user ID")
			return
		}
		userID = &parsedID
	}

	// Placeholder assignment until list filtering uses userID.
	_ = userID

	// This method needs to be implemented in OrderService
	// For now, return empty response
	response := dto.OrderListResponse{
		Orders: []dto.OrderResponse{},
		Total:  0,
		Limit:  limit,
		Offset: offset,
	}

	meta := dto.CalculatePagination(int(response.Total), limit, offset)
	utils.SuccessResponseWithMeta(ctx, http.StatusOK, response, meta)
}

// Health check for order service
func (c *OrderController) Health(ctx *gin.Context) {
	health := dto.HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now(),
		Version:   "1.0.0",
		Services: map[string]string{
			"order_service": "healthy",
		},
		Checks: map[string]bool{
			"database": false, // Would check database connectivity
			"redis":    false, // Would check Redis connectivity
			"rabbitmq": false, // Would check RabbitMQ connectivity
		},
	}

	utils.SuccessResponse(ctx, http.StatusOK, health)
}
