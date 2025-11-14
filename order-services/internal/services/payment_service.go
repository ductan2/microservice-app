package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v78"
	"github.com/stripe/stripe-go/v78/paymentintent"
	"github.com/stripe/stripe-go/v78/webhook"

	"order-services/internal/config"
	"order-services/internal/models"
	"order-services/internal/repositories"
)

var (
	ErrPaymentNotFound      = errors.New("payment not found")
	ErrInvalidPaymentIntent = errors.New("invalid payment intent")
	ErrWebhookSignature     = errors.New("invalid webhook signature")
	ErrDuplicateWebhook     = errors.New("duplicate webhook event")
	ErrPaymentFailed        = errors.New("payment failed")
	ErrOrderNotPaid         = errors.New("order is not paid")
)

// PaymentService defines the business logic interface for payment processing
type PaymentService interface {
	CreatePaymentIntent(ctx context.Context, orderID, userID uuid.UUID) (*stripe.PaymentIntent, error)
	ProcessWebhook(ctx context.Context, webhookBody []byte, stripeSignature string) error
	GetPaymentByOrderID(ctx context.Context, orderID uuid.UUID) (*models.Payment, error)
	HandlePaymentSuccess(ctx context.Context, paymentIntentID string) error
	HandlePaymentFailure(ctx context.Context, paymentIntentID, failureReason string) error
	ConfirmPayment(ctx context.Context, paymentIntentID string) (*models.Payment, error)
}

// paymentService implements the payment business logic
type paymentService struct {
	orderRepo     repositories.OrderRepository
	paymentRepo   repositories.PaymentRepository
	outboxRepo    repositories.OutboxRepository
	webhookRepo   repositories.WebhookEventRepository
	stripeKey     string
	webhookSecret string
}

// NewPaymentService creates a new payment service instance
func NewPaymentService(
	orderRepo repositories.OrderRepository,
	paymentRepo repositories.PaymentRepository,
	outboxRepo repositories.OutboxRepository,
	webhookRepo repositories.WebhookEventRepository,
	config *config.Config,
) PaymentService {
	// Set Stripe key
	stripe.Key = config.StripeSecretKey

	return &paymentService{
		orderRepo:     orderRepo,
		paymentRepo:   paymentRepo,
		outboxRepo:    outboxRepo,
		webhookRepo:   webhookRepo,
		stripeKey:     config.StripeSecretKey,
		webhookSecret: config.StripeWebhookSecret,
	}
}

// CreatePaymentIntent creates a Stripe PaymentIntent for an order
func (s *paymentService) CreatePaymentIntent(ctx context.Context, orderID, userID uuid.UUID) (*stripe.PaymentIntent, error) {
	// Get and validate order
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	if order == nil {
		return nil, ErrOrderNotFound
	}

	// Verify user authorization
	if order.UserID != userID {
		return nil, ErrUnauthorizedOrder
	}

	// Check if order can be paid
	if order.Status != models.OrderStatusCreated && order.Status != models.OrderStatusPendingPayment {
		return nil, fmt.Errorf("%w: order status is %s", ErrInvalidOrderStatus, order.Status)
	}

	// Check if order has expired
	if order.ExpiresAt.Valid && order.ExpiresAt.Time.Before(time.Now()) {
		return nil, ErrOrderExpired
	}

	// Create PaymentIntent parameters
	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(order.TotalAmount),
		Currency: stripe.String(string(order.Currency)),
		Metadata: map[string]string{
			"order_id": orderID.String(),
			"user_id":  userID.String(),
		},
		AutomaticPaymentMethods: &stripe.PaymentIntentAutomaticPaymentMethodsParams{
			Enabled: stripe.Bool(true),
		},
	}

	// Add customer email if available
	if order.CustomerEmail != "" {
		params.ReceiptEmail = stripe.String(order.CustomerEmail)
	}

	params.PaymentMethodOptions = &stripe.PaymentIntentPaymentMethodOptionsParams{
		Card: &stripe.PaymentIntentPaymentMethodOptionsCardParams{
			RequestThreeDSecure: stripe.String("automatic"),
		},
	}

	// Create PaymentIntent
	pi, err := paymentintent.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment intent: %w", err)
	}

	// Start transaction to save payment record and update order
	err = s.orderRepo.WithTx(ctx, func(ctx context.Context) error {
		// Create payment record
		payment := &models.Payment{
			OrderID:               orderID,
			StripePaymentIntentID: pi.ID,
			Amount:                order.TotalAmount,
			Currency:              order.Currency,
			Status:                string(pi.Status),
		}

		if err := s.paymentRepo.Create(ctx, payment); err != nil {
			return fmt.Errorf("failed to create payment record: %w", err)
		}

		// Update order with payment intent ID
		if err := s.orderRepo.UpdatePaymentIntentID(ctx, orderID, pi.ID); err != nil {
			return fmt.Errorf("failed to update order payment intent: %w", err)
		}

		// Create payment.created event
		event := &models.Outbox{
			AggregateID: orderID,
			Topic:       "payment.events",
			Type:        "payment.created",
			Payload:     s.createPaymentEventPayload(payment, "created"),
		}
		if err := s.outboxRepo.Create(ctx, event); err != nil {
			return fmt.Errorf("failed to create payment event: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return pi, nil
}

// ProcessWebhook processes incoming Stripe webhook events
func (s *paymentService) ProcessWebhook(ctx context.Context, webhookBody []byte, stripeSignature string) error {
	// Verify webhook signature
	event, err := webhook.ConstructEvent(webhookBody, stripeSignature, s.webhookSecret)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrWebhookSignature, err)
	}

	// Check if webhook event has already been processed
	if err := s.webhookRepo.IsProcessed(ctx, event.ID); err != nil {
		if err == ErrDuplicateWebhook {
			// Already processed, return success
			return nil
		}
		return fmt.Errorf("failed to check webhook processing status: %w", err)
	}

	// Start transaction for webhook processing
	err = s.orderRepo.WithTx(ctx, func(ctx context.Context) error {
		// Mark webhook as processed
		webhookEvent := &models.WebhookEvent{
			StripeEventID: event.ID,
			Type:          string(event.Type),
			Payload:       webhookBody,
		}

		if err := s.webhookRepo.Create(ctx, webhookEvent); err != nil {
			return fmt.Errorf("failed to create webhook event record: %w", err)
		}

		// Process the event
		if err := s.processStripeEvent(ctx, event); err != nil {
			return fmt.Errorf("failed to process stripe event: %w", err)
		}

		// Mark webhook as processed
		return s.webhookRepo.MarkAsProcessed(ctx, event.ID)
	})

	if err != nil {
		// Log error but don't fail the webhook response to avoid retry storms
		log.Printf("Error processing webhook %s: %v", event.ID, err)
		return nil // Return success to Stripe
	}

	return nil
}

// processStripeEvent handles specific Stripe event types
func (s *paymentService) processStripeEvent(ctx context.Context, event stripe.Event) error {
	switch event.Type {
	case "payment_intent.succeeded":
		return s.handlePaymentIntentSucceeded(ctx, event)
	case "payment_intent.payment_failed":
		return s.handlePaymentIntentFailed(ctx, event)
	case "payment_intent.canceled":
		return s.handlePaymentIntentCanceled(ctx, event)
	case "payment_intent.requires_action":
		return s.handlePaymentIntentRequiresAction(ctx, event)
	default:
		// Log unhandled event type but don't return error
		log.Printf("Unhandled webhook event type: %s", event.Type)
		return nil
	}
}

// handlePaymentIntentSucceeded handles successful payment events
func (s *paymentService) handlePaymentIntentSucceeded(ctx context.Context, event stripe.Event) error {
	var paymentIntent stripe.PaymentIntent
	if err := json.Unmarshal(event.Data.Raw, &paymentIntent); err != nil {
		return fmt.Errorf("failed to parse payment intent: %w", err)
	}

	return s.HandlePaymentSuccess(ctx, paymentIntent.ID)
}

// handlePaymentIntentFailed handles failed payment events
func (s *paymentService) handlePaymentIntentFailed(ctx context.Context, event stripe.Event) error {
	var paymentIntent stripe.PaymentIntent
	if err := json.Unmarshal(event.Data.Raw, &paymentIntent); err != nil {
		return fmt.Errorf("failed to parse payment intent: %w", err)
	}

	failureReason := "Payment failed"
	if paymentIntent.LastPaymentError != nil {
		failureReason = paymentIntent.LastPaymentError.Msg
	}

	return s.HandlePaymentFailure(ctx, paymentIntent.ID, failureReason)
}

// handlePaymentIntentCanceled handles canceled payment events
func (s *paymentService) handlePaymentIntentCanceled(ctx context.Context, event stripe.Event) error {
	var paymentIntent stripe.PaymentIntent
	if err := json.Unmarshal(event.Data.Raw, &paymentIntent); err != nil {
		return fmt.Errorf("failed to parse payment intent: %w", err)
	}

	return s.UpdateOrderStatus(ctx, paymentIntent.ID, models.OrderStatusCancelled, "Payment canceled")
}

// handlePaymentIntentRequiresAction handles payment action required events
func (s *paymentService) handlePaymentIntentRequiresAction(ctx context.Context, event stripe.Event) error {
	var paymentIntent stripe.PaymentIntent
	if err := json.Unmarshal(event.Data.Raw, &paymentIntent); err != nil {
		return fmt.Errorf("failed to parse payment intent: %w", err)
	}

	// Update payment status
	return s.UpdatePaymentStatus(ctx, paymentIntent.ID, string(paymentIntent.Status))
}

// HandlePaymentSuccess processes a successful payment
func (s *paymentService) HandlePaymentSuccess(ctx context.Context, paymentIntentID string) error {
	// Get payment record
	payment, err := s.paymentRepo.GetByStripePaymentIntentID(ctx, paymentIntentID)
	if err != nil {
		return fmt.Errorf("failed to get payment: %w", err)
	}

	if payment == nil {
		return ErrPaymentNotFound
	}

	// Get order
	order, err := s.orderRepo.GetByID(ctx, payment.OrderID)
	if err != nil {
		return fmt.Errorf("failed to get order: %w", err)
	}

	if order == nil {
		return ErrOrderNotFound
	}

	// Start transaction for payment success processing
	return s.orderRepo.WithTx(ctx, func(ctx context.Context) error {
		// Update payment status
		if err := s.paymentRepo.UpdateStatus(ctx, payment.ID, models.PaymentStatusSucceeded); err != nil {
			return fmt.Errorf("failed to update payment status: %w", err)
		}

		// Mark order as paid
		if err := s.orderRepo.MarkAsPaid(ctx, order.ID, paymentIntentID); err != nil {
			return fmt.Errorf("failed to mark order as paid: %w", err)
		}

		// Create payment.succeeded event
		paymentEvent := &models.Outbox{
			AggregateID: order.ID,
			Topic:       "payment.events",
			Type:        "payment.succeeded",
			Payload:     s.createPaymentEventPayload(payment, "succeeded"),
		}
		if err := s.outboxRepo.Create(ctx, paymentEvent); err != nil {
			return fmt.Errorf("failed to create payment success event: %w", err)
		}

		// Create order.paid event (for enrollment service)
		orderEvent := &models.Outbox{
			AggregateID: order.ID,
			Topic:       "order.events",
			Type:        "order.paid",
			Payload:     s.createOrderPaidEventPayload(order, payment),
		}
		if err := s.outboxRepo.Create(ctx, orderEvent); err != nil {
			return fmt.Errorf("failed to create order paid event: %w", err)
		}

		return nil
	})
}

// HandlePaymentFailure processes a failed payment
func (s *paymentService) HandlePaymentFailure(ctx context.Context, paymentIntentID, failureReason string) error {
	// Get payment record
	payment, err := s.paymentRepo.GetByStripePaymentIntentID(ctx, paymentIntentID)
	if err != nil {
		return fmt.Errorf("failed to get payment: %w", err)
	}

	if payment == nil {
		return ErrPaymentNotFound
	}

	// Get order
	order, err := s.orderRepo.GetByID(ctx, payment.OrderID)
	if err != nil {
		return fmt.Errorf("failed to get order: %w", err)
	}

	if order == nil {
		return ErrOrderNotFound
	}

	// Start transaction for payment failure processing
	return s.orderRepo.WithTx(ctx, func(ctx context.Context) error {
		// Update payment status and failure details
		if err := s.paymentRepo.MarkAsFailed(ctx, payment.ID, failureReason, ""); err != nil {
			return fmt.Errorf("failed to mark payment as failed: %w", err)
		}

		// Mark order as failed
		if err := s.orderRepo.MarkAsFailed(ctx, order.ID, failureReason); err != nil {
			return fmt.Errorf("failed to mark order as failed: %w", err)
		}

		// Create payment.failed event
		paymentEvent := &models.Outbox{
			AggregateID: order.ID,
			Topic:       "payment.events",
			Type:        "payment.failed",
			Payload:     s.createPaymentFailedEventPayload(payment, failureReason),
		}
		if err := s.outboxRepo.Create(ctx, paymentEvent); err != nil {
			return fmt.Errorf("failed to create payment failed event: %w", err)
		}

		// Create order.failed event
		orderEvent := &models.Outbox{
			AggregateID: order.ID,
			Topic:       "order.events",
			Type:        "order.failed",
			Payload:     s.createOrderFailedEventPayload(order, failureReason),
		}
		if err := s.outboxRepo.Create(ctx, orderEvent); err != nil {
			return fmt.Errorf("failed to create order failed event: %w", err)
		}

		return nil
	})
}

// GetPaymentByOrderID retrieves payment information for an order
func (s *paymentService) GetPaymentByOrderID(ctx context.Context, orderID uuid.UUID) (*models.Payment, error) {
	payment, err := s.paymentRepo.GetByOrderID(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment for order: %w", err)
	}

	return payment, nil
}

// ConfirmPayment confirms a payment after client-side authentication
func (s *paymentService) ConfirmPayment(ctx context.Context, paymentIntentID string) (*models.Payment, error) {
	// Retrieve payment intent from Stripe to get current status
	params := &stripe.PaymentIntentParams{}
	params.AddExpand("latest_charge")
	pi, err := paymentintent.Get(paymentIntentID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve payment intent: %w", err)
	}

	// Get payment record
	payment, err := s.paymentRepo.GetByStripePaymentIntentID(ctx, paymentIntentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment: %w", err)
	}

	if payment == nil {
		return nil, ErrPaymentNotFound
	}

	// Update payment status
	payment.Status = string(pi.Status)
	if pi.LatestCharge != nil {
		payment.StripeChargeID = pi.LatestCharge.ID
		payment.StripeReceiptURL = pi.LatestCharge.ReceiptURL
		if pi.LatestCharge.PaymentMethodDetails != nil {
			payment.PaymentMethodType = string(pi.LatestCharge.PaymentMethodDetails.Type)
		}
	}

	if err := s.paymentRepo.Update(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to update payment: %w", err)
	}

	// Handle successful payment
	if pi.Status == "succeeded" {
		if err := s.HandlePaymentSuccess(ctx, paymentIntentID); err != nil {
			return nil, fmt.Errorf("failed to handle payment success: %w", err)
		}
	}

	return payment, nil
}

// UpdatePaymentStatus updates payment status (helper method)
func (s *paymentService) UpdatePaymentStatus(ctx context.Context, paymentIntentID, status string) error {
	return s.paymentRepo.UpdateStatusByStripePaymentIntentID(ctx, paymentIntentID, status)
}

// UpdateOrderStatus updates order status (helper method)
func (s *paymentService) UpdateOrderStatus(ctx context.Context, paymentIntentID, status, reason string) error {
	payment, err := s.paymentRepo.GetByStripePaymentIntentID(ctx, paymentIntentID)
	if err != nil {
		return err
	}

	if payment == nil {
		return ErrPaymentNotFound
	}

	timestamp := &sql.NullTime{Time: time.Now(), Valid: true}
	return s.orderRepo.UpdateStatus(ctx, payment.OrderID, status, timestamp, reason)
}

// Event payload creation methods
func (s *paymentService) createPaymentEventPayload(payment *models.Payment, eventType string) []byte {
	payload := map[string]interface{}{
		"payment_id":        payment.ID,
		"order_id":          payment.OrderID,
		"stripe_payment_id": payment.StripePaymentIntentID,
		"amount":            payment.Amount,
		"currency":          payment.Currency,
		"status":            payment.Status,
		"event_type":        eventType,
		"created_at":        payment.CreatedAt,
	}

	// Convert to JSON (simplified for example)
	return s.marshalPayload(payload)
}

func (s *paymentService) createPaymentFailedEventPayload(payment *models.Payment, reason string) []byte {
	payload := map[string]interface{}{
		"payment_id":        payment.ID,
		"order_id":          payment.OrderID,
		"stripe_payment_id": payment.StripePaymentIntentID,
		"amount":            payment.Amount,
		"currency":          payment.Currency,
		"status":            payment.Status,
		"failure_reason":    reason,
		"failed_at":         time.Now(),
	}

	// Convert to JSON (simplified for example)
	return s.marshalPayload(payload)
}

func (s *paymentService) createOrderPaidEventPayload(order *models.Order, payment *models.Payment) []byte {
	payload := map[string]interface{}{
		"order_id":          order.ID,
		"user_id":           order.UserID,
		"payment_id":        payment.ID,
		"total_amount":      order.TotalAmount,
		"currency":          order.Currency,
		"payment_intent_id": payment.StripePaymentIntentID,
		"paid_at":           time.Now(),
		"items":             order.OrderItems,
	}

	// Convert to JSON (simplified for example)
	return s.marshalPayload(payload)
}

func (s *paymentService) createOrderFailedEventPayload(order *models.Order, reason string) []byte {
	payload := map[string]interface{}{
		"order_id":       order.ID,
		"user_id":        order.UserID,
		"total_amount":   order.TotalAmount,
		"currency":       order.Currency,
		"failure_reason": reason,
		"failed_at":      time.Now(),
	}

	// Convert to JSON (simplified for example)
	return s.marshalPayload(payload)
}

func (s *paymentService) marshalPayload(payload map[string]interface{}) []byte {
	data, err := json.Marshal(payload)
	if err != nil {
		return []byte(`{"error": "failed to marshal payload"}`)
	}
	return data
}
