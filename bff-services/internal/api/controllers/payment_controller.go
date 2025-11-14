package controllers

import (
	"net/http"

	"bff-services/internal/api/dto"
	middleware "bff-services/internal/middlewares"
	"bff-services/internal/services"
	"bff-services/internal/utils"

	"github.com/gin-gonic/gin"
)

// PaymentController proxies payment operations to the order service.
type PaymentController struct {
	paymentService services.PaymentService
}

// NewPaymentController constructs a new PaymentController.
func NewPaymentController(paymentService services.PaymentService) *PaymentController {
	return &PaymentController{
		paymentService: paymentService,
	}
}

func (p *PaymentController) CreatePaymentIntent(c *gin.Context) {
	if p.paymentService == nil {
		utils.Fail(c, "Payment service unavailable", http.StatusServiceUnavailable, "payment service not configured")
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

	var req dto.CreatePaymentIntentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, "Invalid request data", http.StatusBadRequest, err.Error())
		return
	}

	resp, err := p.paymentService.CreatePaymentIntent(c.Request.Context(), token, userID, email, sessionID, orderID, req)
	if err != nil {
		utils.Fail(c, "Unable to create payment intent", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}

func (p *PaymentController) ConfirmPayment(c *gin.Context) {
	if p.paymentService == nil {
		utils.Fail(c, "Payment service unavailable", http.StatusServiceUnavailable, "payment service not configured")
		return
	}

	_, _, _, ok := middleware.GetUserContextFromMiddleware(c)
	if !ok {
		return
	}

	token := getOptionalBearerToken(c)
	if token == "" {
		utils.Fail(c, "Unauthorized", http.StatusUnauthorized, "missing bearer token")
		return
	}

	paymentIntentID := c.Param("payment_intent_id")
	if paymentIntentID == "" {
		utils.Fail(c, "Payment intent ID is required", http.StatusBadRequest, "missing payment intent id")
		return
	}

	var req dto.ConfirmPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, "Invalid request data", http.StatusBadRequest, err.Error())
		return
	}

	resp, err := p.paymentService.ConfirmPayment(c.Request.Context(), token, paymentIntentID, req)
	if err != nil {
		utils.Fail(c, "Unable to confirm payment", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}

func (p *PaymentController) GetPaymentByOrderID(c *gin.Context) {
	if p.paymentService == nil {
		utils.Fail(c, "Payment service unavailable", http.StatusServiceUnavailable, "payment service not configured")
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

	resp, err := p.paymentService.GetPaymentByOrderID(c.Request.Context(), token, userID, email, sessionID, orderID)
	if err != nil {
		utils.Fail(c, "Unable to fetch payment", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}

func (p *PaymentController) GetPaymentMethods(c *gin.Context) {
	if p.paymentService == nil {
		utils.Fail(c, "Payment service unavailable", http.StatusServiceUnavailable, "payment service not configured")
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

	resp, err := p.paymentService.GetPaymentMethods(c.Request.Context(), token, userID, email, sessionID)
	if err != nil {
		utils.Fail(c, "Unable to fetch payment methods", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}

func (p *PaymentController) GetPaymentHistory(c *gin.Context) {
	if p.paymentService == nil {
		utils.Fail(c, "Payment service unavailable", http.StatusServiceUnavailable, "payment service not configured")
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

	var query dto.PaymentHistoryQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		utils.Fail(c, "Invalid query parameters", http.StatusBadRequest, err.Error())
		return
	}

	resp, err := p.paymentService.GetPaymentHistory(c.Request.Context(), token, userID, email, sessionID, query)
	if err != nil {
		utils.Fail(c, "Unable to fetch payment history", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}

func (p *PaymentController) GetStripeConfig(c *gin.Context) {
	if p.paymentService == nil {
		utils.Fail(c, "Payment service unavailable", http.StatusServiceUnavailable, "payment service not configured")
		return
	}

	resp, err := p.paymentService.GetStripeConfig(c.Request.Context())
	if err != nil {
		utils.Fail(c, "Unable to fetch Stripe config", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}
