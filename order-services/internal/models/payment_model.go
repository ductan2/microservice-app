package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// Payment represents payment attempts for an order
type Payment struct {
	ID                    uuid.UUID    `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	OrderID               uuid.UUID    `gorm:"type:uuid;not null;index:payments_order_idx;constraint:OnDelete:CASCADE" json:"order_id"`
	StripePaymentIntentID string       `gorm:"type:text;uniqueIndex;not null" json:"stripe_payment_intent_id"`
	Amount                int64        `gorm:"type:bigint;not null" json:"amount"` // in cents
	Currency              string       `gorm:"type:varchar(3);default:'USD';not null" json:"currency"`
	Status                string       `gorm:"type:varchar(50);not null;check:status IN ('requires_payment_method','requires_confirmation','requires_action','processing','succeeded','canceled','failed')" json:"status"`
	PaymentMethod         string       `gorm:"type:text" json:"payment_method,omitempty"`
	PaymentMethodType     string       `gorm:"type:varchar(50);json:"payment_method_type,omitempty"` // card, ideal, etc.
	StripeChargeID        string       `gorm:"type:text;index:payments_charge_idx" json:"stripe_charge_id,omitempty"`
	StripeReceiptURL      string       `gorm:"type:text" json:"stripe_receipt_url,omitempty"`
	FailureMessage        string       `gorm:"type:text" json:"failure_message,omitempty"`
	FailureCode           string       `gorm:"type:text" json:"failure_code,omitempty"`
	ProcessedAt           sql.NullTime `gorm:"type:timestamptz" json:"processed_at,omitempty"`
	CreatedAt             time.Time    `json:"created_at"`
	UpdatedAt             time.Time    `json:"updated_at"`
}

// Payment Status Constants
const (
	PaymentStatusRequiresPaymentMethod = "requires_payment_method"
	PaymentStatusRequiresConfirmation  = "requires_confirmation"
	PaymentStatusRequiresAction         = "requires_action"
	PaymentStatusProcessing             = "processing"
	PaymentStatusSucceeded              = "succeeded"
	PaymentStatusCanceled               = "canceled"
	PaymentStatusFailed                 = "failed"
)

// WebhookEvent represents incoming Stripe webhook events for idempotency
type WebhookEvent struct {
	ID           uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	StripeEventID string   `gorm:"type:text;uniqueIndex;not null" json:"stripe_event_id"`
	Type         string   `gorm:"type:text;not null" json:"type"`
	Payload      []byte   `gorm:"type:jsonb;not null" json:"payload"`
	Processed    bool     `gorm:"default:false;not null" json:"processed"`
	ProcessedAt  sql.NullTime `gorm:"type:timestamptz" json:"processed_at,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}