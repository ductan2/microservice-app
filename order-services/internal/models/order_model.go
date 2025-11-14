package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// Order represents a customer's order for course purchases
type Order struct {
	ID                uuid.UUID    `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	UserID            uuid.UUID    `gorm:"type:uuid;not null;index:orders_user_idx;constraint:OnDelete:CASCADE" json:"user_id"`
	TotalAmount       int64        `gorm:"type:bigint;not null" json:"total_amount"` // in cents
	Currency          string       `gorm:"type:varchar(3);default:'USD';not null" json:"currency"`
	Status            string       `gorm:"type:varchar(50);default:'created';not null;check:status IN ('created','pending_payment','paid','failed','cancelled','refunded')" json:"status"`
	PaymentIntentID   string       `gorm:"type:text;index:orders_payment_intent_idx" json:"payment_intent_id,omitempty"`
	StripeCheckoutID  string       `gorm:"type:text" json:"stripe_checkout_id,omitempty"`
	CustomerEmail     string       `gorm:"type:text;not null" json:"customer_email"`
	CustomerName      string       `gorm:"type:text" json:"customer_name,omitempty"`
	ExpiresAt         sql.NullTime `gorm:"type:timestamptz" json:"expires_at,omitempty"`
	PaidAt            sql.NullTime `gorm:"type:timestamptz" json:"paid_at,omitempty"`
	CancelledAt       sql.NullTime `gorm:"type:timestamptz" json:"cancelled_at,omitempty"`
	FailedAt          sql.NullTime `gorm:"type:timestamptz" json:"failed_at,omitempty"`
	RefundedAt        sql.NullTime `gorm:"type:timestamptz" json:"refunded_at,omitempty"`
	FailureReason     string       `gorm:"type:text" json:"failure_reason,omitempty"`
	Metadata          map[string]any `gorm:"type:jsonb;default:'{}'" json:"metadata"`
	CreatedAt         time.Time    `json:"created_at"`
	UpdatedAt         time.Time    `json:"updated_at"`
	DeletedAt         sql.NullTime `gorm:"type:timestamptz" json:"deleted_at,omitempty"`

	// Relationships
	OrderItems        []OrderItem `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;foreignKey:OrderID;references:ID" json:"order_items,omitempty"`
	Payments          []Payment   `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;foreignKey:OrderID;references:ID" json:"payments,omitempty"`
	CouponRedemptions []CouponRedemption `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;foreignKey:OrderID;references:ID" json:"coupon_redemptions,omitempty"`
	RefundRequests    []RefundRequest `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;foreignKey:OrderID;references:ID" json:"refund_requests,omitempty"`
	Invoices          []Invoice   `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;foreignKey:OrderID;references:ID" json:"invoices,omitempty"`
	FraudLogs         []FraudLog  `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;foreignKey:OrderID;references:ID" json:"fraud_logs,omitempty"`
}

// Order Status Constants
const (
	OrderStatusCreated        = "created"
	OrderStatusPendingPayment = "pending_payment"
	OrderStatusPaid           = "paid"
	OrderStatusFailed         = "failed"
	OrderStatusCancelled      = "cancelled"
	OrderStatusRefunded       = "refunded"
)

// OrderItem represents individual items within an order
type OrderItem struct {
	ID            uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	OrderID       uuid.UUID `gorm:"type:uuid;not null;index:order_items_order_idx;constraint:OnDelete:CASCADE" json:"order_id"`
	CourseID      uuid.UUID `gorm:"type:uuid;not null;index:order_items_course_idx" json:"course_id"`
	CourseTitle   string    `gorm:"type:text;not null" json:"course_title"`
	CourseDescription string  `gorm:"type:text" json:"course_description,omitempty"`
	InstructorID  uuid.UUID `gorm:"type:uuid;index:order_items_instructor_idx" json:"instructor_id,omitempty"`
	InstructorName string   `gorm:"type:text" json:"instructor_name,omitempty"`
	PriceSnapshot int64     `gorm:"type:bigint;not null" json:"price_snapshot"` // in cents
	OriginalPrice int64     `gorm:"type:bigint;not null" json:"original_price"` // in cents
	Quantity      int       `gorm:"type:int;not null;default:1;check:quantity > 0" json:"quantity"`
	ItemType      string    `gorm:"type:varchar(50);default:'course';not null;check:item_type IN ('course','bundle','subscription')" json:"item_type"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// OrderItem Type Constants
const (
	OrderItemTypeCourse      = "course"
	OrderItemTypeBundle      = "bundle"
	OrderItemTypeSubscription = "subscription"
)