package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v78"

	"order-services/internal/config"
	"order-services/internal/models"
)

// NotificationService handles integration with the notification service
type NotificationService interface {
	SendOrderConfirmation(ctx context.Context, order *models.Order) error
	SendPaymentConfirmation(ctx context.Context, order *models.Order, payment *models.Payment) error
	SendOrderCancelled(ctx context.Context, order *models.Order, reason string) error
	SendPaymentFailed(ctx context.Context, order *models.Order, payment *models.Payment, reason string) error
	SendCouponRedeemed(ctx context.Context, userID uuid.UUID, coupon *models.Coupon, orderID uuid.UUID) error
	SendLowBalanceWarning(ctx context.Context, userID uuid.UUID) error
	SendNewRefundRequest(ctx context.Context, refundRequest *models.RefundRequest) error
	SendRefundProcessed(ctx context.Context, refundRequest *models.RefundRequest, stripeRefund *stripe.Refund) error
	SendRefundRejected(ctx context.Context, refundRequest *models.RefundRequest, adminReason string) error
	SendRefundCompleted(ctx context.Context, refundRequest *models.RefundRequest, stripeRefund *stripe.Refund) error
}

// notificationService implements NotificationService
type notificationService struct {
	baseURL    string
	httpClient *http.Client
	config     *config.Config
}

// NewNotificationService creates a new notification service instance
func NewNotificationService(config *config.Config) NotificationService {
	return &notificationService{
		baseURL: getNotificationServiceURL(config),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		config: config,
	}
}

// SendOrderConfirmation sends an order confirmation notification
func (s *notificationService) SendOrderConfirmation(ctx context.Context, order *models.Order) error {
	notification := NotificationRequest{
		UserID:  order.ID, // Use order ID as notification user identifier
		Type:    "order_confirmation",
		Title:   "Order Confirmation",
		Message: fmt.Sprintf("Your order #%s has been created successfully. Total: $%.2f", order.ID.String(), float64(order.TotalAmount)/100),
		Data: map[string]interface{}{
			"order_id":     order.ID,
			"total_amount": order.TotalAmount,
			"currency":     order.Currency,
			"status":       order.Status,
			"items":        order.OrderItems,
		},
		Channels: []string{"email", "push"},
		Priority: "normal",
	}

	return s.sendNotification(ctx, notification)
}

// SendPaymentConfirmation sends a payment confirmation notification
func (s *notificationService) SendPaymentConfirmation(ctx context.Context, order *models.Order, payment *models.Payment) error {
	notification := NotificationRequest{
		UserID:  order.UserID,
		Type:    "payment_confirmation",
		Title:   "Payment Successful",
		Message: fmt.Sprintf("Payment of $%.2f for order #%s has been processed successfully", float64(payment.Amount)/100, order.ID.String()),
		Data: map[string]interface{}{
			"order_id":          order.ID,
			"payment_id":        payment.ID,
			"payment_intent_id": payment.StripePaymentIntentID,
			"amount":            payment.Amount,
			"currency":          payment.Currency,
			"payment_status":    payment.Status,
			"receipt_url":       payment.StripeReceiptURL,
		},
		Channels: []string{"email", "push"},
		Priority: "high",
	}

	return s.sendNotification(ctx, notification)
}

// SendOrderCancelled sends an order cancellation notification
func (s *notificationService) SendOrderCancelled(ctx context.Context, order *models.Order, reason string) error {
	notification := NotificationRequest{
		UserID:  order.UserID,
		Type:    "order_cancelled",
		Title:   "Order Cancelled",
		Message: fmt.Sprintf("Your order #%s has been cancelled. Reason: %s", order.ID.String(), reason),
		Data: map[string]interface{}{
			"order_id":     order.ID,
			"reason":       reason,
			"total_amount": order.TotalAmount,
			"currency":     order.Currency,
			"cancelled_at": order.CancelledAt,
		},
		Channels: []string{"email", "push"},
		Priority: "normal",
	}

	return s.sendNotification(ctx, notification)
}

// SendPaymentFailed sends a payment failure notification
func (s *notificationService) SendPaymentFailed(ctx context.Context, order *models.Order, payment *models.Payment, reason string) error {
	notification := NotificationRequest{
		UserID:  order.UserID,
		Type:    "payment_failed",
		Title:   "Payment Failed",
		Message: fmt.Sprintf("Payment for order #%s has failed. Please try again or update your payment method. Reason: %s", order.ID.String(), reason),
		Data: map[string]interface{}{
			"order_id":          order.ID,
			"payment_id":        payment.ID,
			"payment_intent_id": payment.StripePaymentIntentID,
			"amount":            payment.Amount,
			"currency":          payment.Currency,
			"failure_reason":    reason,
			"failure_message":   payment.FailureMessage,
		},
		Channels: []string{"email", "push"},
		Priority: "high",
	}

	return s.sendNotification(ctx, notification)
}

// SendCouponRedeemed sends a coupon redemption notification
func (s *notificationService) SendCouponRedeemed(ctx context.Context, userID uuid.UUID, coupon *models.Coupon, orderID uuid.UUID) error {
	discountText := ""
	if coupon.Type == "percentage" && coupon.PercentOff != nil {
		discountText = fmt.Sprintf("%d%% off", *coupon.PercentOff)
	} else if coupon.Type == "fixed_amount" && coupon.AmountOff != nil {
		discountText = fmt.Sprintf("$%.2f off", float64(*coupon.AmountOff)/100)
	}

	notification := NotificationRequest{
		UserID:  userID,
		Type:    "coupon_redeemed",
		Title:   "Coupon Applied",
		Message: fmt.Sprintf("Coupon '%s' (%s) has been applied to your order #%s", coupon.Code, discountText, orderID.String()),
		Data: map[string]interface{}{
			"coupon_id":   coupon.ID,
			"coupon_code": coupon.Code,
			"coupon_name": coupon.Name,
			"discount":    discountText,
			"order_id":    orderID,
		},
		Channels: []string{"email", "push"},
		Priority: "normal",
	}

	return s.sendNotification(ctx, notification)
}

// SendLowBalanceWarning sends a low balance warning notification
func (s *notificationService) SendLowBalanceWarning(ctx context.Context, userID uuid.UUID) error {
	notification := NotificationRequest{
		UserID:  userID,
		Type:    "low_balance_warning",
		Title:   "Low Balance Warning",
		Message: "Your account balance is low. Please add funds to continue using our services.",
		Data: map[string]interface{}{
			"user_id": userID,
		},
		Channels: []string{"email", "push"},
		Priority: "medium",
	}

	return s.sendNotification(ctx, notification)
}

// SendNewRefundRequest notifies admins about a new refund request
func (s *notificationService) SendNewRefundRequest(ctx context.Context, refundRequest *models.RefundRequest) error {
	notification := NotificationRequest{
		UserID:  refundRequest.UserID,
		Type:    "refund_request_submitted",
		Title:   "New Refund Request",
		Message: fmt.Sprintf("Refund request submitted for order %s", refundRequest.OrderID.String()),
		Data: map[string]interface{}{
			"refund_id": refundRequest.ID,
			"order_id":  refundRequest.OrderID,
			"user_id":   refundRequest.UserID,
			"amount":    refundRequest.Amount,
			"reason":    refundRequest.Reason,
			"status":    refundRequest.Status,
		},
		Channels: []string{"email"},
		Priority: "medium",
	}

	return s.sendNotification(ctx, notification)
}

// SendRefundProcessed notifies the user when a refund has been processed
func (s *notificationService) SendRefundProcessed(ctx context.Context, refundRequest *models.RefundRequest, stripeRefund *stripe.Refund) error {
	refundAmount := float64(refundRequest.Amount) / 100
	notification := NotificationRequest{
		UserID:  refundRequest.UserID,
		Type:    "refund_processed",
		Title:   "Refund Processed",
		Message: fmt.Sprintf("Your refund for order %s has been processed. Amount: $%.2f", refundRequest.OrderID.String(), refundAmount),
		Data: map[string]interface{}{
			"refund_id":        refundRequest.ID,
			"order_id":         refundRequest.OrderID,
			"amount":           refundRequest.Amount,
			"stripe_refund_id": stripeRefund.ID,
			"status":           refundRequest.Status,
		},
		Channels: []string{"email", "push"},
		Priority: "high",
	}

	return s.sendNotification(ctx, notification)
}

// SendRefundRejected notifies the user when a refund has been rejected
func (s *notificationService) SendRefundRejected(ctx context.Context, refundRequest *models.RefundRequest, adminReason string) error {
	notification := NotificationRequest{
		UserID:  refundRequest.UserID,
		Type:    "refund_rejected",
		Title:   "Refund Request Rejected",
		Message: fmt.Sprintf("Your refund request for order %s has been rejected. Reason: %s", refundRequest.OrderID.String(), adminReason),
		Data: map[string]interface{}{
			"refund_id":    refundRequest.ID,
			"order_id":     refundRequest.OrderID,
			"amount":       refundRequest.Amount,
			"admin_reason": adminReason,
			"status":       models.RefundStatusRejected,
		},
		Channels: []string{"email"},
		Priority: "normal",
	}

	return s.sendNotification(ctx, notification)
}

// SendRefundCompleted notifies the user when Stripe confirms the refund completion
func (s *notificationService) SendRefundCompleted(ctx context.Context, refundRequest *models.RefundRequest, stripeRefund *stripe.Refund) error {
	notification := NotificationRequest{
		UserID:  refundRequest.UserID,
		Type:    "refund_completed",
		Title:   "Refund Completed",
		Message: fmt.Sprintf("Refund for order %s has been completed by the payment processor.", refundRequest.OrderID.String()),
		Data: map[string]interface{}{
			"refund_id":        refundRequest.ID,
			"order_id":         refundRequest.OrderID,
			"amount":           refundRequest.Amount,
			"stripe_refund_id": stripeRefund.ID,
			"status":           models.RefundStatusProcessed,
		},
		Channels: []string{"email", "push"},
		Priority: "normal",
	}

	return s.sendNotification(ctx, notification)
}

// Internal structs

type NotificationRequest struct {
	UserID   uuid.UUID              `json:"user_id"`
	Type     string                 `json:"type"`
	Title    string                 `json:"title"`
	Message  string                 `json:"message"`
	Data     map[string]interface{} `json:"data"`
	Channels []string               `json:"channels"`
	Priority string                 `json:"priority"` // low, normal, medium, high, urgent
}

// Helper methods

func (s *notificationService) sendNotification(ctx context.Context, notification NotificationRequest) error {
	url := fmt.Sprintf("%s/api/v1/notifications", s.baseURL)

	// Marshal request body
	body, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("failed to marshal notification request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create notification request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Internal-Service", "order-service")
	req.Header.Set("Authorization", "Bearer "+s.getInternalAuthToken())

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("notification service returned status: %d", resp.StatusCode)
	}

	log.Printf("Notification sent: type=%s, user_id=%s, title=%s",
		notification.Type, notification.UserID, notification.Title)

	return nil
}

func (s *notificationService) getInternalAuthToken() string {
	// In production, this would use proper internal service authentication
	// For now, return a mock token
	return "internal-service-token"
}

func getNotificationServiceURL(config *config.Config) string {
	// In production, this would come from environment variables or service discovery
	// For now, return a default URL
	return "http://notification-services:8007"
}

// MockNotificationService implements a mock notification service for testing
type MockNotificationService struct {
	notifications []NotificationRequest
}

// NewMockNotificationService creates a new mock notification service
func NewMockNotificationService() *MockNotificationService {
	return &MockNotificationService{
		notifications: make([]NotificationRequest, 0),
	}
}

// GetNotifications returns all sent notifications (for testing)
func (s *MockNotificationService) GetNotifications() []NotificationRequest {
	return s.notifications
}

// ClearNotifications clears all sent notifications (for testing)
func (s *MockNotificationService) ClearNotifications() {
	s.notifications = make([]NotificationRequest, 0)
}

// SendOrderConfirmation sends order confirmation (mock implementation)
func (s *MockNotificationService) SendOrderConfirmation(ctx context.Context, order *models.Order) error {
	notification := NotificationRequest{
		UserID:  order.UserID,
		Type:    "order_confirmation",
		Title:   "Order Confirmation",
		Message: fmt.Sprintf("Order #%s created", order.ID.String()),
	}
	s.notifications = append(s.notifications, notification)
	log.Printf("Mock: Order confirmation sent for order %s", order.ID)
	return nil
}

// SendPaymentConfirmation sends payment confirmation (mock implementation)
func (s *MockNotificationService) SendPaymentConfirmation(ctx context.Context, order *models.Order, payment *models.Payment) error {
	notification := NotificationRequest{
		UserID:  order.UserID,
		Type:    "payment_confirmation",
		Title:   "Payment Successful",
		Message: fmt.Sprintf("Payment of $%.2f processed", float64(payment.Amount)/100),
	}
	s.notifications = append(s.notifications, notification)
	log.Printf("Mock: Payment confirmation sent for order %s", order.ID)
	return nil
}

// SendOrderCancelled sends order cancellation (mock implementation)
func (s *MockNotificationService) SendOrderCancelled(ctx context.Context, order *models.Order, reason string) error {
	notification := NotificationRequest{
		UserID:  order.UserID,
		Type:    "order_cancelled",
		Title:   "Order Cancelled",
		Message: fmt.Sprintf("Order #%s cancelled: %s", order.ID.String(), reason),
	}
	s.notifications = append(s.notifications, notification)
	log.Printf("Mock: Order cancellation sent for order %s", order.ID)
	return nil
}

// SendPaymentFailed sends payment failure (mock implementation)
func (s *MockNotificationService) SendPaymentFailed(ctx context.Context, order *models.Order, payment *models.Payment, reason string) error {
	notification := NotificationRequest{
		UserID:  order.UserID,
		Type:    "payment_failed",
		Title:   "Payment Failed",
		Message: fmt.Sprintf("Payment failed for order #%s: %s", order.ID.String(), reason),
	}
	s.notifications = append(s.notifications, notification)
	log.Printf("Mock: Payment failure sent for order %s", order.ID)
	return nil
}

// SendCouponRedeemed sends coupon redemption (mock implementation)
func (s *MockNotificationService) SendCouponRedeemed(ctx context.Context, userID uuid.UUID, coupon *models.Coupon, orderID uuid.UUID) error {
	notification := NotificationRequest{
		UserID:  userID,
		Type:    "coupon_redeemed",
		Title:   "Coupon Applied",
		Message: fmt.Sprintf("Coupon '%s' applied to order %s", coupon.Code, orderID.String()),
	}
	s.notifications = append(s.notifications, notification)
	log.Printf("Mock: Coupon redemption sent for coupon %s", coupon.Code)
	return nil
}

// SendLowBalanceWarning sends low balance warning (mock implementation)
func (s *MockNotificationService) SendLowBalanceWarning(ctx context.Context, userID uuid.UUID) error {
	notification := NotificationRequest{
		UserID:  userID,
		Type:    "low_balance_warning",
		Title:   "Low Balance Warning",
		Message: "Your account balance is low",
	}
	s.notifications = append(s.notifications, notification)
	log.Printf("Mock: Low balance warning sent for user %s", userID)
	return nil
}

// SendNewRefundRequest sends refund request notification (mock implementation)
func (s *MockNotificationService) SendNewRefundRequest(ctx context.Context, refundRequest *models.RefundRequest) error {
	notification := NotificationRequest{
		UserID:  refundRequest.UserID,
		Type:    "refund_request_submitted",
		Title:   "New Refund Request",
		Message: fmt.Sprintf("Refund request for order %s", refundRequest.OrderID.String()),
	}
	s.notifications = append(s.notifications, notification)
	log.Printf("Mock: Refund request notification sent for refund %s", refundRequest.ID)
	return nil
}

// SendRefundProcessed sends refund processed notification (mock implementation)
func (s *MockNotificationService) SendRefundProcessed(ctx context.Context, refundRequest *models.RefundRequest, stripeRefund *stripe.Refund) error {
	notification := NotificationRequest{
		UserID:  refundRequest.UserID,
		Type:    "refund_processed",
		Title:   "Refund Processed",
		Message: fmt.Sprintf("Refund processed for order %s", refundRequest.OrderID.String()),
	}
	s.notifications = append(s.notifications, notification)
	log.Printf("Mock: Refund processed notification sent for refund %s", refundRequest.ID)
	return nil
}

// SendRefundRejected sends refund rejected notification (mock implementation)
func (s *MockNotificationService) SendRefundRejected(ctx context.Context, refundRequest *models.RefundRequest, adminReason string) error {
	notification := NotificationRequest{
		UserID:  refundRequest.UserID,
		Type:    "refund_rejected",
		Title:   "Refund Rejected",
		Message: fmt.Sprintf("Refund rejected for order %s: %s", refundRequest.OrderID.String(), adminReason),
	}
	s.notifications = append(s.notifications, notification)
	log.Printf("Mock: Refund rejected notification sent for refund %s", refundRequest.ID)
	return nil
}

// SendRefundCompleted sends refund completed notification (mock implementation)
func (s *MockNotificationService) SendRefundCompleted(ctx context.Context, refundRequest *models.RefundRequest, stripeRefund *stripe.Refund) error {
	notification := NotificationRequest{
		UserID:  refundRequest.UserID,
		Type:    "refund_completed",
		Title:   "Refund Completed",
		Message: fmt.Sprintf("Refund completed for order %s", refundRequest.OrderID.String()),
	}
	s.notifications = append(s.notifications, notification)
	log.Printf("Mock: Refund completed notification sent for refund %s", refundRequest.ID)
	return nil
}
