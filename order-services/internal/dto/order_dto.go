package dto

import (
	"time"

	"github.com/google/uuid"
	"order-services/internal/models"
)

// CreateOrderRequest represents the request to create a new order
type CreateOrderRequest struct {
	Items         []CreateOrderItemRequest `json:"items" validate:"required,min=1"`
	CouponCode    *string                  `json:"coupon_code,omitempty" validate:"omitempty,min=3,max=50"`
	CustomerEmail string                   `json:"customer_email" validate:"required,email,max=255"`
	CustomerName  *string                  `json:"customer_name,omitempty" validate:"omitempty,max=100"`
	Metadata      map[string]interface{}   `json:"metadata,omitempty"`
}

// CreateOrderItemRequest represents an item in the create order request
type CreateOrderItemRequest struct {
	CourseID      uuid.UUID `json:"course_id" validate:"required,uuid4"`
	Quantity      int       `json:"quantity" validate:"required,min=1,max=10"`
	PriceSnapshot *int64    `json:"price_snapshot,omitempty" validate:"omitempty,min=0"` // in cents
}

// OrderResponse represents the response for an order
type OrderResponse struct {
	ID                uuid.UUID               `json:"id"`
	UserID            uuid.UUID               `json:"user_id"`
	TotalAmount       int64                   `json:"total_amount"`        // in cents
	OriginalAmount    int64                   `json:"original_amount"`     // in cents (before discount)
	DiscountAmount    int64                   `json:"discount_amount"`     // in cents
	Currency          string                  `json:"currency"`
	Status            string                  `json:"status"`
	PaymentIntentID   *string                 `json:"payment_intent_id,omitempty"`
	StripeCheckoutID  *string                 `json:"stripe_checkout_id,omitempty"`
	CustomerEmail     string                  `json:"customer_email"`
	CustomerName      *string                 `json:"customer_name,omitempty"`
	ExpiresAt         *time.Time              `json:"expires_at,omitempty"`
	PaidAt            *time.Time              `json:"paid_at,omitempty"`
	CancelledAt       *time.Time              `json:"cancelled_at,omitempty"`
	FailedAt          *time.Time              `json:"failed_at,omitempty"`
	RefundedAt        *time.Time              `json:"refunded_at,omitempty"`
	FailureReason     *string                 `json:"failure_reason,omitempty"`
	Metadata          map[string]interface{}   `json:"metadata,omitempty"`
	CreatedAt         time.Time               `json:"created_at"`
	UpdatedAt         time.Time               `json:"updated_at"`
	OrderItems        []OrderItemResponse     `json:"order_items"`
	Payments          []PaymentResponse       `json:"payments,omitempty"`
	CouponRedemptions []CouponRedemptionResponse `json:"coupon_redemptions,omitempty"`
}

// OrderItemResponse represents an order item in the response
type OrderItemResponse struct {
	ID               uuid.UUID  `json:"id"`
	CourseID         uuid.UUID  `json:"course_id"`
	CourseTitle      string     `json:"course_title"`
	CourseDescription *string   `json:"course_description,omitempty"`
	InstructorID     *uuid.UUID `json:"instructor_id,omitempty"`
	InstructorName   *string    `json:"instructor_name,omitempty"`
	PriceSnapshot    int64      `json:"price_snapshot"`    // in cents
	OriginalPrice    int64      `json:"original_price"`    // in cents
	Quantity         int        `json:"quantity"`
	ItemType         string     `json:"item_type"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// OrderListResponse represents a paginated list of orders
type OrderListResponse struct {
	Orders []OrderResponse `json:"orders"`
	Total  int64           `json:"total"`
	Limit  int             `json:"limit"`
	Offset int             `json:"offset"`
}

// UpdateOrderRequest represents the request to update an order (admin only)
type UpdateOrderRequest struct {
	Status        *string                `json:"status,omitempty" validate:"omitempty,oneof=created pending_payment paid failed cancelled refunded"`
	FailureReason *string                `json:"failure_reason,omitempty" validate:"omitempty,max=500"`
	Metadata      *map[string]interface{} `json:"metadata,omitempty"`
}

// CancelOrderRequest represents the request to cancel an order
type CancelOrderRequest struct {
	Reason string `json:"reason" validate:"required,min=1,max=500"`
}

// PaymentResponse represents a payment in the response
type PaymentResponse struct {
	ID                    uuid.UUID  `json:"id"`
	OrderID               uuid.UUID  `json:"order_id"`
	StripePaymentIntentID string     `json:"stripe_payment_intent_id"`
	Amount                int64      `json:"amount"`                // in cents
	Currency              string     `json:"currency"`
	Status                string     `json:"status"`
	PaymentMethod         *string    `json:"payment_method,omitempty"`
	PaymentMethodType     *string    `json:"payment_method_type,omitempty"`
	StripeChargeID        *string    `json:"stripe_charge_id,omitempty"`
	StripeReceiptURL      *string    `json:"stripe_receipt_url,omitempty"`
	FailureMessage        *string    `json:"failure_message,omitempty"`
	FailureCode           *string    `json:"failure_code,omitempty"`
	ProcessedAt           *time.Time `json:"processed_at,omitempty"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
}

// CouponRedemptionResponse represents a coupon redemption in the response
type CouponRedemptionResponse struct {
	ID             uuid.UUID `json:"id"`
	CouponID       uuid.UUID `json:"coupon_id"`
	Coupon         CouponResponse `json:"coupon"`
	UserID         uuid.UUID `json:"user_id"`
	OrderID        uuid.UUID `json:"order_id"`
	DiscountAmount int64     `json:"discount_amount"` // in cents
	RedeemedAt     time.Time `json:"redeemed_at"`
}

// OrderStatsResponse represents order statistics
type OrderStatsResponse struct {
	TotalOrders        int64 `json:"total_orders"`
	TotalRevenue       int64 `json:"total_revenue"`       // in cents
	PendingOrders      int64 `json:"pending_orders"`
	CompletedOrders    int64 `json:"completed_orders"`
	CancelledOrders    int64 `json:"cancelled_orders"`
	FailedOrders       int64 `json:"failed_orders"`
	RefundedOrders     int64 `json:"refunded_orders"`
	AverageOrderValue  int64 `json:"average_order_value"`  // in cents
}

// Conversion functions

// FromModel converts an Order model to OrderResponse
func (r *OrderResponse) FromModel(order *models.Order) {
	if order == nil {
		return
	}

	r.ID = order.ID
	r.UserID = order.UserID
	r.TotalAmount = order.TotalAmount
	r.Currency = order.Currency
	r.Status = order.Status
	r.CustomerEmail = order.CustomerEmail
	r.Metadata = order.Metadata
	r.CreatedAt = order.CreatedAt
	r.UpdatedAt = order.UpdatedAt

	// Optional fields
	if order.PaymentIntentID != "" {
		r.PaymentIntentID = &order.PaymentIntentID
	}
	if order.StripeCheckoutID != "" {
		r.StripeCheckoutID = &order.StripeCheckoutID
	}
	if order.CustomerName != "" {
		r.CustomerName = &order.CustomerName
	}
	if order.ExpiresAt.Valid {
		r.ExpiresAt = &order.ExpiresAt.Time
	}
	if order.PaidAt.Valid {
		r.PaidAt = &order.PaidAt.Time
	}
	if order.CancelledAt.Valid {
		r.CancelledAt = &order.CancelledAt.Time
	}
	if order.FailedAt.Valid {
		r.FailedAt = &order.FailedAt.Time
	}
	if order.RefundedAt.Valid {
		r.RefundedAt = &order.RefundedAt.Time
	}
	if order.FailureReason != "" {
		r.FailureReason = &order.FailureReason
	}

	// Convert order items
	r.OrderItems = make([]OrderItemResponse, len(order.OrderItems))
	for i, item := range order.OrderItems {
		r.OrderItems[i] = OrderItemResponse{}.FromModel(&item)
	}

	// Convert payments
	r.Payments = make([]PaymentResponse, len(order.Payments))
	for i, payment := range order.Payments {
		r.Payments[i] = PaymentResponse{}.FromModel(&payment)
	}

	// Convert coupon redemptions and calculate discount
	r.CouponRedemptions = make([]CouponRedemptionResponse, len(order.CouponRedemptions))
	for i, redemption := range order.CouponRedemptions {
		r.CouponRedemptions[i] = CouponRedemptionResponse{}.FromModel(&redemption)
		r.DiscountAmount += redemption.DiscountAmount
	}

	// Calculate original amount
	r.OriginalAmount = r.TotalAmount + r.DiscountAmount
}

// FromModel converts an OrderItem model to OrderItemResponse
func (r OrderItemResponse) FromModel(item *models.OrderItem) OrderItemResponse {
	response := OrderItemResponse{
		ID:            item.ID,
		CourseID:      item.CourseID,
		CourseTitle:   item.CourseTitle,
		PriceSnapshot: item.PriceSnapshot,
		OriginalPrice: item.OriginalPrice,
		Quantity:      item.Quantity,
		ItemType:      item.ItemType,
		CreatedAt:     item.CreatedAt,
		UpdatedAt:     item.UpdatedAt,
	}

	if item.CourseDescription != "" {
		response.CourseDescription = &item.CourseDescription
	}
	if item.InstructorID != uuid.Nil {
		response.InstructorID = &item.InstructorID
	}
	if item.InstructorName != "" {
		response.InstructorName = &item.InstructorName
	}

	return response
}

// FromModel converts a Payment model to PaymentResponse
func (r PaymentResponse) FromModel(payment *models.Payment) PaymentResponse {
	response := PaymentResponse{
		ID:                    payment.ID,
		OrderID:               payment.OrderID,
		StripePaymentIntentID: payment.StripePaymentIntentID,
		Amount:                payment.Amount,
		Currency:              payment.Currency,
		Status:                payment.Status,
		CreatedAt:             payment.CreatedAt,
		UpdatedAt:             payment.UpdatedAt,
	}

	if payment.PaymentMethod != "" {
		response.PaymentMethod = &payment.PaymentMethod
	}
	if payment.PaymentMethodType != "" {
		response.PaymentMethodType = &payment.PaymentMethodType
	}
	if payment.StripeChargeID != "" {
		response.StripeChargeID = &payment.StripeChargeID
	}
	if payment.StripeReceiptURL != "" {
		response.StripeReceiptURL = &payment.StripeReceiptURL
	}
	if payment.FailureMessage != "" {
		response.FailureMessage = &payment.FailureMessage
	}
	if payment.FailureCode != "" {
		response.FailureCode = &payment.FailureCode
	}
	if payment.ProcessedAt.Valid {
		response.ProcessedAt = &payment.ProcessedAt.Time
	}

	return response
}

// FromModel converts a CouponRedemption model to CouponRedemptionResponse
func (r CouponRedemptionResponse) FromModel(redemption *models.CouponRedemption) CouponRedemptionResponse {
	response := CouponRedemptionResponse{
		ID:             redemption.ID,
		CouponID:       redemption.CouponID,
		UserID:         redemption.UserID,
		OrderID:        redemption.OrderID,
		DiscountAmount: redemption.DiscountAmount,
		RedeemedAt:     redemption.RedeemedAt,
	}

	// Convert coupon if available
	if redemption.Coupon.Code != "" {
		var couponResponse CouponResponse
		couponResponse.FromModel(&redemption.Coupon)
		response.Coupon = couponResponse
	}

	return response
}