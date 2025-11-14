package controllers

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v78/paymentintent"

	"order-services/internal/config"
	"order-services/internal/dto"
	"order-services/internal/services"
	"order-services/pkg/utils"
)

// PaymentController handles payment-related HTTP requests
type PaymentController struct {
	paymentService services.PaymentService
	config         *config.Config
}

// NewPaymentController creates a new payment controller instance
func NewPaymentController(paymentService services.PaymentService, cfg *config.Config) *PaymentController {
	return &PaymentController{
		paymentService: paymentService,
		config:         cfg,
	}
}

// CreatePaymentIntent creates a new payment intent for an order
// @Summary Create payment intent
// @Description Creates a Stripe payment intent for the specified order
// @Tags payments
// @Accept json
// @Produce json
// @Param id path string true "Order ID"
// @Param request body dto.CreatePaymentIntentRequest false "Payment intent options"
// @Param Authorization header string true "Bearer JWT token"
// @Success 201 {object} dto.APIResponse{data=dto.CreatePaymentIntentResponse}
// @Failure 400 {object} dto.APIResponse
// @Failure 401 {object} dto.APIResponse
// @Failure 404 {object} dto.APIResponse
// @Failure 500 {object} dto.APIResponse
// @Router /api/v1/orders/{id}/pay [post]
func (c *PaymentController) CreatePaymentIntent(ctx *gin.Context) {
	// Parse order ID
	orderID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		utils.ErrorResponse(ctx, http.StatusBadRequest, dto.ErrCodeBadRequest, "Invalid order ID")
		return
	}

	var req dto.CreatePaymentIntentRequest
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

	// Create payment intent
	stripePI, err := c.paymentService.CreatePaymentIntent(ctx, orderID, userUUID)
	if err != nil {
		if utils.IsNotFoundError(err) {
			utils.ErrorResponse(ctx, http.StatusNotFound, dto.ErrCodeOrderNotFound, "Order not found")
		} else if utils.IsValidationError(err) {
			utils.ErrorResponse(ctx, http.StatusBadRequest, dto.ErrCodeValidationFailed, err.Error())
		} else if utils.IsConflictError(err) {
			utils.ErrorResponse(ctx, http.StatusConflict, dto.ErrCodeConflict, err.Error())
		} else {
			utils.ErrorResponse(ctx, http.StatusInternalServerError, dto.ErrCodeInternalError, "Failed to create payment intent")
		}
		return
	}

	// Convert to response
	response := dto.CreatePaymentIntentResponse{
		ClientSecret:    stripePI.ClientSecret,
		PaymentIntentID: stripePI.ID,
		Status:          string(stripePI.Status),
		Amount:          stripePI.Amount,
		Currency:        string(stripePI.Currency),
	}

	// Add next action information
	if stripePI.NextAction != nil {
		response.NextAction = &dto.NextAction{
			Type: string(stripePI.NextAction.Type),
		}
		if stripePI.NextAction.RedirectToURL != nil {
			response.NextAction.RedirectURL = stripePI.NextAction.RedirectToURL.URL
		}
		if stripePI.NextAction.UseStripeSDK != nil {
			useStripeSDK := true
			response.NextAction.UseStripeSDK = &useStripeSDK
		}
	}

	utils.SuccessResponse(ctx, http.StatusCreated, response)
}

// ConfirmPayment confirms a payment after client-side authentication
// @Summary Confirm payment
// @Description Confirms a payment after client-side authentication is complete
// @Tags payments
// @Accept json
// @Produce json
// @Param payment_intent_id path string true "Payment Intent ID"
// @Param request body dto.ConfirmPaymentRequest true "Payment confirmation request"
// @Param Authorization header string true "Bearer JWT token"
// @Success 200 {object} dto.APIResponse{data=dto.ConfirmPaymentResponse}
// @Failure 400 {object} dto.APIResponse
// @Failure 401 {object} dto.APIResponse
// @Failure 404 {object} dto.APIResponse
// @Failure 500 {object} dto.APIResponse
// @Router /api/v1/payments/{payment_intent_id}/confirm [post]
func (c *PaymentController) ConfirmPayment(ctx *gin.Context) {
	paymentIntentID := ctx.Param("payment_intent_id")
	if paymentIntentID == "" {
		utils.ErrorResponse(ctx, http.StatusBadRequest, dto.ErrCodeBadRequest, "Payment intent ID is required")
		return
	}

	var req dto.ConfirmPaymentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(ctx, err)
		return
	}

	// Confirm payment
	payment, err := c.paymentService.ConfirmPayment(ctx, paymentIntentID)
	if err != nil {
		if utils.IsNotFoundError(err) {
			utils.ErrorResponse(ctx, http.StatusNotFound, dto.ErrCodePaymentNotFound, "Payment not found")
		} else if utils.IsValidationError(err) {
			utils.ErrorResponse(ctx, http.StatusBadRequest, dto.ErrCodeValidationFailed, err.Error())
		} else {
			utils.ErrorResponse(ctx, http.StatusInternalServerError, dto.ErrCodeInternalError, "Failed to confirm payment")
		}
		return
	}

	// Get payment intent from Stripe for additional information
	stripePI, err := paymentintent.Get(paymentIntentID, nil)
	if err != nil {
		// Log error but continue with what we have
		stripePI = nil
	}

	// Convert to response
	paymentResponse := dto.PaymentResponse{}.FromModel(payment)
	response := dto.ConfirmPaymentResponse{
		PaymentIntentID: payment.StripePaymentIntentID,
		Status:          payment.Status,
		Amount:          payment.Amount,
		Currency:        payment.Currency,
		Payment:         paymentResponse,
	}

	if stripePI != nil {
		response.Status = string(stripePI.Status)

		// Add next action information
		if stripePI.NextAction != nil {
			response.NextAction = &dto.NextAction{
				Type: string(stripePI.NextAction.Type),
			}
			if stripePI.NextAction.RedirectToURL != nil {
				response.NextAction.RedirectURL = stripePI.NextAction.RedirectToURL.URL
			}
			if stripePI.NextAction.UseStripeSDK != nil {
				useStripeSDK := true
				response.NextAction.UseStripeSDK = &useStripeSDK
			}
		}

		// Add failure information
		if stripePI.LastPaymentError != nil {
			response.FailureReason = &stripePI.LastPaymentError.Msg
		}
	}

	// Use charge information from payment model (already populated by ConfirmPayment)
	if payment.StripeChargeID != "" {
		response.ChargeID = &payment.StripeChargeID
	}
	if payment.StripeReceiptURL != "" {
		response.ReceiptURL = &payment.StripeReceiptURL
	}

	utils.SuccessResponse(ctx, http.StatusOK, response)
}

// GetPayment retrieves payment information by payment intent ID
// @Summary Get payment information
// @Description Retrieves payment information for a payment intent
// @Tags payments
// @Produce json
// @Param payment_intent_id path string true "Payment Intent ID"
// @Param Authorization header string true "Bearer JWT token"
// @Success 200 {object} dto.APIResponse{data=dto.PaymentResponse}
// @Failure 400 {object} dto.APIResponse
// @Failure 401 {object} dto.APIResponse
// @Failure 404 {object} dto.APIResponse
// @Failure 500 {object} dto.APIResponse
// @Router /api/v1/payments/{payment_intent_id} [get]
func (c *PaymentController) GetPayment(ctx *gin.Context) {
	paymentIntentID := ctx.Param("payment_intent_id")
	if paymentIntentID == "" {
		utils.ErrorResponse(ctx, http.StatusBadRequest, dto.ErrCodeBadRequest, "Payment intent ID is required")
		return
	}

	// Get payment by order ID (this would need to be implemented in PaymentService)
	// For now, we'll get it by payment intent ID
	// payment, err := c.paymentService.GetPaymentByPaymentIntentID(ctx, paymentIntentID)
	// This method doesn't exist yet, so we'll need to implement it or get by order ID

	utils.ErrorResponse(ctx, http.StatusNotImplemented, dto.ErrCodeInternalError, "GetPaymentByPaymentIntentID not implemented")
}

// GetPaymentByOrderID retrieves payment information for an order
// @Summary Get payment by order ID
// @Description Retrieves payment information for the specified order
// @Tags payments
// @Produce json
// @Param id path string true "Order ID"
// @Param Authorization header string true "Bearer JWT token"
// @Success 200 {object} dto.APIResponse{data=dto.PaymentResponse}
// @Failure 400 {object} dto.APIResponse
// @Failure 401 {object} dto.APIResponse
// @Failure 404 {object} dto.APIResponse
// @Failure 500 {object} dto.APIResponse
// @Router /api/v1/orders/{id}/payment [get]
func (c *PaymentController) GetPaymentByOrderID(ctx *gin.Context) {
	// Parse order ID
	orderID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		utils.ErrorResponse(ctx, http.StatusBadRequest, dto.ErrCodeBadRequest, "Invalid order ID")
		return
	}

	// Get payment
	payment, err := c.paymentService.GetPaymentByOrderID(ctx, orderID)
	if err != nil {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, dto.ErrCodeInternalError, "Failed to retrieve payment")
		return
	}

	if payment == nil {
		utils.ErrorResponse(ctx, http.StatusNotFound, dto.ErrCodePaymentNotFound, "Payment not found")
		return
	}

	// Convert to response
	response := dto.PaymentResponse{}.FromModel(payment)

	utils.SuccessResponse(ctx, http.StatusOK, response)
}

// GetStripeConfig retrieves Stripe configuration for the frontend
// @Summary Get Stripe configuration
// @Description Retrieves Stripe publishable key and configuration for the frontend
// @Tags payments
// @Produce json
// @Success 200 {object} dto.APIResponse{data=dto.StripeConfigResponse}
// @Router /api/v1/stripe/config [get]
func (c *PaymentController) GetStripeConfig(ctx *gin.Context) {
	configResp := dto.DefaultStripeConfig()

	if c.config != nil && c.config.StripePublishableKey != "" {
		configResp.PublishableKey = c.config.StripePublishableKey
	} else {
		configResp.PublishableKey = os.Getenv("STRIPE_PUBLISHABLE_KEY")
	}

	utils.SuccessResponse(ctx, http.StatusOK, configResp)
}

// ProcessStripeWebhook processes incoming Stripe webhook events
// @Summary Process Stripe webhook
// @Description Processes incoming Stripe webhook events
// @Tags webhooks
// @Accept json
// @Produce json
// @Param stripe_signature header string true "Stripe signature"
// @Param webhook_body body string true "Raw webhook payload"
// @Success 200 {object} dto.APIResponse{data=dto.WebhookEventResponse}
// @Failure 400 {object} dto.APIResponse
// @Failure 500 {object} dto.APIResponse
// @Router /api/v1/stripe/webhook [post]
func (c *PaymentController) ProcessStripeWebhook(ctx *gin.Context) {
	// Get Stripe signature from header
	stripeSignature := ctx.GetHeader("stripe-signature")
	if stripeSignature == "" {
		utils.ErrorResponse(ctx, http.StatusBadRequest, dto.ErrCodeWebhookSignature, "Stripe signature is required")
		return
	}

	// Read raw body
	body, err := ctx.GetRawData()
	if err != nil {
		utils.ErrorResponse(ctx, http.StatusBadRequest, dto.ErrCodeBadRequest, "Failed to read request body")
		return
	}

	// Process webhook
	err = c.paymentService.ProcessWebhook(ctx, body, stripeSignature)
	if err != nil {
		if utils.IsValidationError(err) {
			utils.ErrorResponse(ctx, http.StatusBadRequest, dto.ErrCodeWebhookSignature, err.Error())
		} else {
			utils.ErrorResponse(ctx, http.StatusInternalServerError, dto.ErrCodeInternalError, "Failed to process webhook")
		}
		return
	}

	response := dto.WebhookEventResponse{
		Processed: true,
		Message:   "Webhook processed successfully",
	}

	utils.SuccessResponse(ctx, http.StatusOK, response)
}

// GetPaymentMethods retrieves available payment methods for a user
// @Summary Get payment methods
// @Description Retrieves available payment methods for the authenticated user
// @Tags payments
// @Produce json
// @Param Authorization header string true "Bearer JWT token"
// @Success 200 {object} dto.APIResponse{data=dto.PaymentMethodsResponse}
// @Failure 401 {object} dto.APIResponse
// @Failure 500 {object} dto.APIResponse
// @Router /api/v1/payment-methods [get]
func (c *PaymentController) GetPaymentMethods(ctx *gin.Context) {
	// Get user ID from JWT token
	_, exists := ctx.Get("user_id")
	if !exists {
		utils.ErrorResponse(ctx, http.StatusUnauthorized, dto.ErrCodeUnauthorized, "User not authenticated")
		return
	}

	// This would need to be implemented in PaymentService
	// For now, return empty response
	response := dto.PaymentMethodsResponse{
		PaymentMethods: []dto.PaymentMethod{},
	}

	utils.SuccessResponse(ctx, http.StatusOK, response)
}

// GetPaymentHistory retrieves payment history for a user
// @Summary Get payment history
// @Description Retrieves payment history for the authenticated user with pagination
// @Tags payments
// @Produce json
// @Param limit query int false "Number of items per page (default: 20, max: 100)"
// @Param offset query int false "Number of items to skip (default: 0)"
// @Param status query string false "Filter by payment status"
// @Param Authorization header string true "Bearer JWT token"
// @Success 200 {object} dto.APIResponse{data=dto.PaymentHistoryResponse}
// @Failure 400 {object} dto.APIResponse
// @Failure 401 {object} dto.APIResponse
// @Failure 500 {object} dto.APIResponse
// @Router /api/v1/payments [get]
func (c *PaymentController) GetPaymentHistory(ctx *gin.Context) {
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
	_, exists := ctx.Get("user_id")
	if !exists {
		utils.ErrorResponse(ctx, http.StatusUnauthorized, dto.ErrCodeUnauthorized, "User not authenticated")
		return
	}

	// This would need to be implemented in PaymentService
	// For now, return empty response
	response := dto.PaymentHistoryResponse{
		Payments: []dto.PaymentHistoryItem{},
		Total:    0,
		Limit:    limit,
		Offset:   offset,
	}

	meta := dto.CalculatePagination(int(response.Total), limit, offset)
	utils.SuccessResponseWithMeta(ctx, http.StatusOK, response, meta)
}

// GetPaymentStats retrieves payment statistics (admin only)
// @Summary Get payment statistics
// @Description Retrieves payment statistics for the given criteria (admin only)
// @Tags payments
// @Produce json
// @Param user_id query string false "Filter by user ID"
// @Param start_date query string false "Filter by start date (YYYY-MM-DD)"
// @Param end_date query string false "Filter by end date (YYYY-MM-DD)"
// @Param Authorization header string true "Bearer JWT token"
// @Success 200 {object} dto.APIResponse{data=dto.PaymentStatsResponse}
// @Failure 400 {object} dto.APIResponse
// @Failure 401 {object} dto.APIResponse
// @Failure 403 {object} dto.APIResponse
// @Failure 500 {object} dto.APIResponse
// @Router /api/v1/admin/payments/stats [get]
func (c *PaymentController) GetPaymentStats(ctx *gin.Context) {
	// Check if user is admin
	if !utils.IsAdmin(ctx) {
		utils.ErrorResponse(ctx, http.StatusForbidden, dto.ErrCodeForbidden, "Admin access required")
		return
	}

	// This would need to be implemented in PaymentService
	// For now, return empty stats
	stats := &dto.PaymentStatsResponse{
		TotalPayments:    0,
		SuccessfulAmount: 0,
		FailedAmount:     0,
		PendingAmount:    0,
		SuccessRate:      0,
		AverageAmount:    0,
	}

	utils.SuccessResponse(ctx, http.StatusOK, stats)
}
