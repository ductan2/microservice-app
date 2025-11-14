package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v78"
	"github.com/stripe/stripe-go/v78/refund"

	"order-services/internal/config"
	"order-services/internal/models"
	"order-services/internal/repositories"
	"order-services/internal/types"
)

var (
	ErrRefundNotFound       = errors.New("refund not found")
	ErrRefundAlreadyProcessed = errors.New("refund already processed")
	ErrRefundAmountExceedsPayment = errors.New("refund amount exceeds payment amount")
	ErrRefundWindowExpired  = errors.New("refund window has expired")
	ErrInvalidRefundReason  = errors.New("invalid refund reason")
	ErrRefundNotAllowed      = errors.New("refund not allowed for this order status")
)

// RefundService defines the business logic interface for refund processing
type RefundService interface {
	CreateRefundRequest(ctx context.Context, orderID, userID uuid.UUID, reason string, amount *int64) (*models.RefundRequest, error)
	ProcessRefundRequest(ctx context.Context, refundID uuid.UUID, approve bool, adminReason string) error
	ProcessStripeRefund(ctx context.Context, paymentIntentID string, amount int64, reason string) (*stripe.Refund, error)
	GetRefundRequests(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.RefundRequest, int64, error)
	GetRefundRequest(ctx context.Context, refundID uuid.UUID) (*models.RefundRequest, error)
	GetAllRefundRequests(ctx context.Context, limit, offset int) ([]models.RefundRequest, int64, error)
	HandleStripeRefundWebhook(ctx context.Context, stripeRefund *stripe.Refund) error
	GetRefundStats(ctx context.Context, timeRange *types.TimeRange) (*repositories.RefundStats, error)
}


// refundService implements RefundService
type refundService struct {
	orderRepo      repositories.OrderRepository
	paymentRepo    repositories.PaymentRepository
	refundRepo     repositories.RefundRepository
	outboxRepo     repositories.OutboxRepository
	notificationService NotificationService
	config         *config.Config
}

// NewRefundService creates a new refund service instance
func NewRefundService(
	orderRepo repositories.OrderRepository,
	paymentRepo repositories.PaymentRepository,
	refundRepo repositories.RefundRepository,
	outboxRepo repositories.OutboxRepository,
	notificationService NotificationService,
	config *config.Config,
) RefundService {
	// Set Stripe key
	stripe.Key = config.StripeSecretKey

	return &refundService{
		orderRepo:           orderRepo,
		paymentRepo:         paymentRepo,
		refundRepo:          refundRepo,
		outboxRepo:          outboxRepo,
		notificationService: notificationService,
		config:              config,
	}
}

// CreateRefundRequest creates a new refund request
func (s *refundService) CreateRefundRequest(ctx context.Context, orderID, userID uuid.UUID, reason string, amount *int64) (*models.RefundRequest, error) {
	// Validate reason
	if !isValidRefundReason(reason) {
		return nil, ErrInvalidRefundReason
	}

	// Get order
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	if order == nil {
		return nil, ErrOrderNotFound
	}

	// Verify user owns the order
	if order.UserID != userID {
		return nil, ErrUnauthorizedOrder
	}

	// Check if order allows refunds
	if !s.canRefundOrder(order) {
		return nil, ErrRefundNotAllowed
	}

	// Get payment
	payment, err := s.paymentRepo.GetByOrderID(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment: %w", err)
	}

	if payment == nil {
		return nil, ErrPaymentNotFound
	}

	// Only paid orders can be refunded
	if payment.Status != models.PaymentStatusSucceeded {
		return nil, ErrOrderNotPaid
	}

	// Validate refund amount
	refundAmount := payment.Amount
	if amount != nil {
		refundAmount = *amount
	}

	if refundAmount <= 0 {
		return nil, fmt.Errorf("refund amount must be positive")
	}

	if refundAmount > payment.Amount {
		return nil, ErrRefundAmountExceedsPayment
	}

	// Check if refund already exists
	existingRefund, err := s.refundRepo.GetByOrderID(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing refund: %w", err)
	}

	if existingRefund != nil && existingRefund.Status != models.RefundStatusRejected {
		return nil, ErrRefundAlreadyProcessed
	}

	// Create refund request
	refundRequest := &models.RefundRequest{
		OrderID: orderID,
		UserID:  userID,
		Amount:  refundAmount,
		Reason:  reason,
		Status:  models.RefundStatusPending,
	}

	// Save refund request
	err = s.refundRepo.Create(ctx, refundRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to create refund request: %w", err)
	}

	// Create notification for admin
	err = s.notificationService.SendNewRefundRequest(ctx, refundRequest)
	if err != nil {
		// Log error but don't fail the request
		log.Printf("Warning: Failed to send refund request notification: %v", err)
	}

	return refundRequest, nil
}

// ProcessRefundRequest processes a refund request (admin action)
func (s *refundService) ProcessRefundRequest(ctx context.Context, refundID uuid.UUID, approve bool, adminReason string) error {
	// Get refund request
	refundRequest, err := s.refundRepo.GetByID(ctx, refundID)
	if err != nil {
		return fmt.Errorf("failed to get refund request: %w", err)
	}

	if refundRequest == nil {
		return ErrRefundNotFound
	}

	// Check if already processed
	if refundRequest.Status != models.RefundStatusPending {
		return ErrRefundAlreadyProcessed
	}

	// Get order and payment
	order, err := s.orderRepo.GetByID(ctx, refundRequest.OrderID)
	if err != nil {
		return fmt.Errorf("failed to get order: %w", err)
	}

	if order == nil {
		return ErrOrderNotFound
	}

	payment, err := s.paymentRepo.GetByOrderID(ctx, refundRequest.OrderID)
	if err != nil {
		return fmt.Errorf("failed to get payment: %w", err)
	}

	if payment == nil {
		return ErrPaymentNotFound
	}

	// Update refund request status
	newStatus := models.RefundStatusRejected
	if approve {
		newStatus = models.RefundStatusApproved
	}

	err = s.refundRepo.UpdateStatus(ctx, refundID, newStatus, adminReason)
	if err != nil {
		return fmt.Errorf("failed to update refund status: %w", err)
	}

	// Create refund event payload
	eventData := map[string]interface{}{
		"refund_id":     refundID,
		"order_id":      refundRequest.OrderID,
		"user_id":       refundRequest.UserID,
		"amount":        refundRequest.Amount,
		"status":        newStatus,
		"admin_reason":  adminReason,
		"processed_at":  time.Now(),
	}

	// Process refund if approved
	if approve {
		// Process actual Stripe refund
		stripeRefund, err := s.ProcessStripeRefund(ctx, payment.StripePaymentIntentID, refundRequest.Amount, refundRequest.Reason)
		if err != nil {
			// Mark as failed but don't roll back
			s.refundRepo.UpdateStatus(ctx, refundID, models.RefundStatusFailed, fmt.Sprintf("Stripe processing failed: %v", err))

			// Create failed refund event
			eventData["status"] = models.RefundStatusFailed
			eventData["error"] = err.Error()

			outboxEvent := &models.Outbox{
				AggregateID: refundRequest.OrderID,
				Topic:       "order.events",
				Type:        "order.refund_failed",
				Payload:     s.createEventPayload(eventData),
			}
			s.outboxRepo.Create(ctx, outboxEvent)

			return fmt.Errorf("failed to process Stripe refund: %w", err)
		}

		// Update refund request with Stripe refund ID
		err = s.refundRepo.UpdateStripeRefundID(ctx, refundID, stripeRefund.ID)
		if err != nil {
			log.Printf("Warning: Failed to update refund request with Stripe refund ID: %v", err)
		}

		// Mark order as refunded
		err = s.orderRepo.UpdateStatus(ctx, refundRequest.OrderID, models.OrderStatusRefunded, nil, "Refund processed")
		if err != nil {
			log.Printf("Warning: Failed to update order status to refunded: %v", err)
		}

		// Create refund events
		outboxEvent := &models.Outbox{
			AggregateID: refundRequest.OrderID,
			Topic:       "order.events",
			Type:        "order.refunded",
			Payload:     s.createEventPayload(eventData),
		}
		s.outboxRepo.Create(ctx, outboxEvent)

		// Send notifications
		err = s.notificationService.SendRefundProcessed(ctx, refundRequest, stripeRefund)
		if err != nil {
			log.Printf("Warning: Failed to send refund processed notification: %v", err)
		}

	} else {
		// Rejected refund
		eventData["status"] = models.RefundStatusRejected

		outboxEvent := &models.Outbox{
			AggregateID: refundRequest.OrderID,
			Topic:       "order.events",
			Type:        "order.refund_rejected",
			Payload:     s.createEventPayload(eventData),
		}
		s.outboxRepo.Create(ctx, outboxEvent)

		// Send rejection notification
		err = s.notificationService.SendRefundRejected(ctx, refundRequest, adminReason)
		if err != nil {
			log.Printf("Warning: Failed to send refund rejected notification: %v", err)
		}
	}

	return nil
}

// ProcessStripeRefund processes a refund through Stripe
func (s *refundService) ProcessStripeRefund(ctx context.Context, paymentIntentID string, amount int64, reason string) (*stripe.Refund, error) {
	params := &stripe.RefundParams{
		PaymentIntent: stripe.String(paymentIntentID),
		Amount:        stripe.Int64(amount),
		Reason:        stripe.String("requested_by_customer"),
		Metadata: map[string]string{
			"reason": reason,
			"source":  "order_service",
		},
	}

	refund, err := refund.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create Stripe refund: %w", err)
	}

	return refund, nil
}

// HandleStripeRefundWebhook handles Stripe refund webhook events
func (s *refundService) HandleStripeRefundWebhook(ctx context.Context, stripeRefund *stripe.Refund) error {
	if stripeRefund.Status == "succeeded" {
		// Find refund request by Stripe refund ID
		refundRequest, err := s.refundRepo.GetByStripeRefundID(ctx, stripeRefund.ID)
		if err != nil {
			return fmt.Errorf("failed to find refund request for Stripe refund %s: %w", stripeRefund.ID, err)
		}

		if refundRequest != nil {
			// Update refund request status
			err = s.refundRepo.UpdateStatus(ctx, refundRequest.ID, models.RefundStatusProcessed, "Refund processed successfully")
			if err != nil {
				return fmt.Errorf("failed to update refund request status: %w", err)
			}

			// Create success event
			eventData := map[string]interface{}{
				"refund_id":       refundRequest.ID,
				"stripe_refund_id": stripeRefund.ID,
				"status":         models.RefundStatusProcessed,
				"processed_at":   time.Now(),
			}

			outboxEvent := &models.Outbox{
				AggregateID: refundRequest.OrderID,
				Topic:       "order.events",
				Type:        "order.refund_completed",
				Payload:     s.createEventPayload(eventData),
			}
			s.outboxRepo.Create(ctx, outboxEvent)

			// Send completion notification
			err = s.notificationService.SendRefundCompleted(ctx, refundRequest, stripeRefund)
			if err != nil {
				log.Printf("Warning: Failed to send refund completed notification: %v", err)
			}
		}
	}

	return nil
}

// GetRefundRequests retrieves refund requests for a user
func (s *refundService) GetRefundRequests(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.RefundRequest, int64, error) {
	return s.refundRepo.GetByUserID(ctx, userID, limit, offset)
}

// GetRefundRequest retrieves a specific refund request
func (s *refundService) GetRefundRequest(ctx context.Context, refundID uuid.UUID) (*models.RefundRequest, error) {
	return s.refundRepo.GetByID(ctx, refundID)
}

// GetAllRefundRequests retrieves all refund requests (admin)
func (s *refundService) GetAllRefundRequests(ctx context.Context, limit, offset int) ([]models.RefundRequest, int64, error) {
	return s.refundRepo.GetAll(ctx, limit, offset)
}

// GetRefundStats retrieves refund statistics
func (s *refundService) GetRefundStats(ctx context.Context, timeRange *types.TimeRange) (*repositories.RefundStats, error) {
	return s.refundRepo.GetStats(ctx, timeRange)
}

// Helper methods

func (s *refundService) canRefundOrder(order *models.Order) bool {
	// Only paid orders can be refunded
	if order.Status != models.OrderStatusPaid {
		return false
	}

	// Check refund window (30 days from paid date)
	if order.PaidAt.Valid {
		refundWindow := time.Duration(s.config.RefundWindowDays) * 24 * time.Hour
		if time.Since(order.PaidAt.Time) > refundWindow {
			return false
		}
	}

	return true
}

func isValidRefundReason(reason string) bool {
	validReasons := []string{
		"dissatisfaction",
		"technical_issues",
		"duplicate_purchase",
		"accidental_purchase",
		"course_not_as_described",
		"other",
	}

	for _, validReason := range validReasons {
		if reason == validReason {
			return true
		}
	}

	return false
}

func (s *refundService) createEventPayload(data map[string]interface{}) []byte {
	// In production, use proper JSON marshaling
	return []byte(fmt.Sprintf(`{"event": "order.refund_updated", "data": %v}`, data))
}