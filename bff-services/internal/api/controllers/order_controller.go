package controllers

import (
	"net/http"

	"bff-services/internal/api/dto"
	middleware "bff-services/internal/middlewares"
	"bff-services/internal/services"
	"bff-services/internal/utils"

	"github.com/gin-gonic/gin"
)

// OrderController proxies order-related requests to the order service.
type OrderController struct {
	orderService services.OrderService
}

// NewOrderController constructs a new OrderController instance.
func NewOrderController(orderService services.OrderService) *OrderController {
	return &OrderController{
		orderService: orderService,
	}
}

func (o *OrderController) CreateOrder(c *gin.Context) {
	if o.orderService == nil {
		utils.Fail(c, "Order service unavailable", http.StatusServiceUnavailable, "order service not configured")
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

	var req dto.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, "Invalid request data", http.StatusBadRequest, err.Error())
		return
	}

	resp, err := o.orderService.CreateOrder(c.Request.Context(), token, userID, email, sessionID, req)
	if err != nil {
		utils.Fail(c, "Unable to create order", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}

func (o *OrderController) ListOrders(c *gin.Context) {
	if o.orderService == nil {
		utils.Fail(c, "Order service unavailable", http.StatusServiceUnavailable, "order service not configured")
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

	var query dto.OrderListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		utils.Fail(c, "Invalid query parameters", http.StatusBadRequest, err.Error())
		return
	}

	resp, err := o.orderService.ListOrders(c.Request.Context(), token, userID, email, sessionID, query)
	if err != nil {
		utils.Fail(c, "Unable to list orders", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}

func (o *OrderController) GetOrder(c *gin.Context) {
	if o.orderService == nil {
		utils.Fail(c, "Order service unavailable", http.StatusServiceUnavailable, "order service not configured")
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

	orderID := c.Param("id")
	if orderID == "" {
		utils.Fail(c, "Order ID is required", http.StatusBadRequest, "missing order id")
		return
	}

	resp, err := o.orderService.GetOrder(c.Request.Context(), token, userID, email, sessionID, orderID)
	if err != nil {
		utils.Fail(c, "Unable to fetch order", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}

func (o *OrderController) CancelOrder(c *gin.Context) {
	if o.orderService == nil {
		utils.Fail(c, "Order service unavailable", http.StatusServiceUnavailable, "order service not configured")
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

	orderID := c.Param("id")
	if orderID == "" {
		utils.Fail(c, "Order ID is required", http.StatusBadRequest, "missing order id")
		return
	}

	var req dto.CancelOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, "Invalid request data", http.StatusBadRequest, err.Error())
		return
	}

	resp, err := o.orderService.CancelOrder(c.Request.Context(), token, userID, email, sessionID, orderID, req)
	if err != nil {
		utils.Fail(c, "Unable to cancel order", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}
