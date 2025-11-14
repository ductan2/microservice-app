package dto

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"order-services/internal/models"
)

// CreatePaymentIntentRequest represents the request to create a payment intent
type CreatePaymentIntentRequest struct {
	// OrderID is extracted from the URL path, no need to include in body
	PaymentMethod *string `json:"payment_method,omitempty" validate:"omitempty"` // Payment method ID for immediate confirmation
	Confirm       *bool   `json:"confirm,omitempty" validate:"omitempty"`         // Whether to confirm the payment immediately
}

// CreatePaymentIntentResponse represents the response for creating a payment intent
type CreatePaymentIntentResponse struct {
	ClientSecret    string `json:"client_secret"`
	PaymentIntentID string `json:"payment_intent_id"`
	Status          string `json:"status"`
	Amount          int64  `json:"amount"`          // in cents
	Currency        string `json:"currency"`
	NextAction      *NextAction `json:"next_action,omitempty"`
}

// NextAction represents the next action required for payment
type NextAction struct {
	Type         string `json:"type"`
	RedirectURL  string `json:"redirect_url,omitempty"`
	UseStripeSDK *bool  `json:"use_stripe_sdk,omitempty"`
}

// ConfirmPaymentRequest represents the request to confirm a payment
type ConfirmPaymentRequest struct {
	PaymentMethodID string `json:"payment_method_id" validate:"required"` // Payment method ID to use
}

// ConfirmPaymentResponse represents the response for confirming a payment
type ConfirmPaymentResponse struct {
	PaymentIntentID string               `json:"payment_intent_id"`
	Status          string               `json:"status"`
	Amount          int64                `json:"amount"`          // in cents
	Currency        string               `json:"currency"`
	NextAction      *NextAction          `json:"next_action,omitempty"`
	ChargeID        *string              `json:"charge_id,omitempty"`
	ReceiptURL      *string              `json:"receipt_url,omitempty"`
	FailureReason   *string              `json:"failure_reason,omitempty"`
	Payment         PaymentResponse      `json:"payment"`
}

// PaymentMethodsResponse represents available payment methods
type PaymentMethodsResponse struct {
	PaymentMethods []PaymentMethod `json:"payment_methods"`
}

// PaymentMethod represents a payment method
type PaymentMethod struct {
	ID           string      `json:"id"`
	Type         string      `json:"type"`
	Card         *CardInfo   `json:"card,omitempty"`
	CreatedAt    time.Time   `json:"created_at"`
	IsDefault    bool        `json:"is_default"`
}

// CardInfo represents card information
type CardInfo struct {
	Brand        string `json:"brand"`
	Last4        string `json:"last4"`
	ExpiryMonth  int    `json:"expiry_month"`
	ExpiryYear   int    `json:"expiry_year"`
	Country      string `json:"country"`
	Fingerprint  string `json:"fingerprint"`
	Funding      string `json:"funding"` // credit, debit, prepaid, unknown
}

// StripeConfigResponse represents Stripe configuration for the frontend
type StripeConfigResponse struct {
	PublishableKey    string   `json:"publishable_key"`
	Currency          string   `json:"currency"`
	Country           string   `json:"country"`
	PaymentMethodTypes []string `json:"payment_method_types"`
	MinimumAmount     int64    `json:"minimum_amount"`     // in cents
	MaximumAmount     int64    `json:"maximum_amount"`     // in cents
	AcceptedCurrencies []string `json:"accepted_currencies"`
}

// WebhookEventResponse represents the response for webhook processing
type WebhookEventResponse struct {
	EventID    string `json:"event_id"`
	EventType  string `json:"event_type"`
	Processed  bool   `json:"processed"`
	Message    string `json:"message,omitempty"`
}

// RefundRequest represents the request to process a refund
type RefundRequest struct {
	Reason   string `json:"reason" validate:"required,min=1,max=500"`
	Amount   *int64 `json:"amount,omitempty" validate:"omitempty,min=0"` // in cents, if not specified refund full amount
}

// RefundResponse represents the response for a refund
type RefundResponse struct {
	RefundID      string    `json:"refund_id"`
	PaymentIntentID string   `json:"payment_intent_id"`
	Amount        int64     `json:"amount"`         // in cents
	Currency      string    `json:"currency"`
	Status        string    `json:"status"`
	Reason        string    `json:"reason"`
	CreatedAt     time.Time `json:"created_at"`
	ReceiptURL    *string   `json:"receipt_url,omitempty"`
}

// PaymentHistoryResponse represents payment history for a user
type PaymentHistoryResponse struct {
	Payments []PaymentHistoryItem `json:"payments"`
	Total    int64                `json:"total"`
	Limit    int                  `json:"limit"`
	Offset   int                  `json:"offset"`
}

// PaymentHistoryItem represents a payment in history
type PaymentHistoryItem struct {
	ID                uuid.UUID  `json:"id"`
	OrderID           uuid.UUID  `json:"order_id"`
	Amount            int64      `json:"amount"`            // in cents
	Currency          string     `json:"currency"`
	Status            string     `json:"status"`
	PaymentMethodType *string    `json:"payment_method_type,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	ProcessedAt       *time.Time `json:"processed_at,omitempty"`
	Order             OrderResponse `json:"order"`
}

// PaymentStatsResponse represents payment statistics
type PaymentStatsResponse struct {
	TotalPayments    int64 `json:"total_payments"`
	SuccessfulAmount int64 `json:"successful_amount"` // in cents
	FailedAmount     int64 `json:"failed_amount"`     // in cents
	PendingAmount    int64 `json:"pending_amount"`    // in cents
	SuccessRate      float64 `json:"success_rate"`     // percentage
	AverageAmount    int64 `json:"average_amount"`    // in cents
}

// Conversion functions

// FromStripePaymentIntent converts a Stripe PaymentIntent to CreatePaymentIntentResponse
func (r *CreatePaymentIntentResponse) FromStripePaymentIntent(pi interface{}) {
	// This would be implemented with the actual Stripe PaymentIntent struct
	// For now, we'll use a generic interface approach
	// In production, you would use stripe.PaymentIntent directly

	// Example implementation (pseudo-code):
	/*
	stripePI := pi.(stripe.PaymentIntent)
	r.ClientSecret = stripePI.ClientSecret
	r.PaymentIntentID = stripePI.ID
	r.Status = string(stripePI.Status)
	r.Amount = stripePI.Amount
	r.Currency = string(stripePI.Currency)

	if stripePI.NextAction != nil {
		r.NextAction = &NextAction{
			Type: string(stripePI.NextAction.Type),
		}
		if stripePI.NextAction.RedirectToURL != nil {
			r.NextAction.RedirectURL = stripePI.NextAction.RedirectToURL.URL
		}
		if stripePI.NextAction.UseStripeSDK != nil {
			r.NextAction.UseStripeSDK = stripePI.NextAction.UseStripeSDK
		}
	}
	*/
}

// FromPaymentModel converts a Payment model to PaymentHistoryItem
func (r *PaymentHistoryItem) FromPaymentModel(payment *models.Payment, order *models.Order) {
	r.ID = payment.ID
	r.OrderID = payment.OrderID
	r.Amount = payment.Amount
	r.Currency = payment.Currency
	r.Status = payment.Status
	r.CreatedAt = payment.CreatedAt

	if payment.PaymentMethodType != "" {
		r.PaymentMethodType = &payment.PaymentMethodType
	}
	if payment.ProcessedAt.Valid {
		r.ProcessedAt = &payment.ProcessedAt.Time
	}

	// Convert order
	if order != nil {
		r.Order = OrderResponse{}
		r.Order.FromModel(order)
	}
}

// DefaultStripeConfig returns default Stripe configuration
func DefaultStripeConfig() StripeConfigResponse {
	return StripeConfigResponse{
		Currency:      "USD",
		Country:       "US",
		PaymentMethodTypes: []string{
			"card",
			"apple_pay",
			"google_pay",
			"ideal",
			"sepa_debit",
		},
		MinimumAmount:      50,  // $0.50 minimum
		MaximumAmount:      999999, // $9,999.99 maximum
		AcceptedCurrencies: []string{"USD", "EUR", "GBP"},
	}
}

// ValidatePaymentAmount validates payment amount against configured limits
func ValidatePaymentAmount(amount int64, config StripeConfigResponse) error {
	if amount < config.MinimumAmount {
		return fmt.Errorf("minimum amount is %d cents", config.MinimumAmount)
	}
	if amount > config.MaximumAmount {
		return fmt.Errorf("maximum amount is %d cents", config.MaximumAmount)
	}
	return nil
}