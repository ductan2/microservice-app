package dto

// CreateOrderRequest represents the payload to create a new order via the BFF.
type CreateOrderRequest struct {
	Items         []CreateOrderItemRequest `json:"items" binding:"required,min=1"`
	CouponCode    *string                  `json:"coupon_code,omitempty"`
	CustomerEmail string                   `json:"customer_email" binding:"required,email"`
	CustomerName  *string                  `json:"customer_name,omitempty"`
	Metadata      map[string]interface{}   `json:"metadata,omitempty"`
}

// CreateOrderItemRequest represents a single item in an order creation payload.
type CreateOrderItemRequest struct {
	CourseID      string `json:"course_id" binding:"required"`
	Quantity      int    `json:"quantity" binding:"required,min=1"`
	PriceSnapshot *int64 `json:"price_snapshot,omitempty"`
}

// CancelOrderRequest captures the payload to cancel an order.
type CancelOrderRequest struct {
	Reason string `json:"reason" binding:"required"`
}

// OrderListQuery captures the supported query parameters for listing orders.
type OrderListQuery struct {
	Limit     int    `form:"limit"`
	Offset    int    `form:"offset"`
	Page      int    `form:"page"`
	Status    string `form:"status"`
	SortBy    string `form:"sort_by"`
	SortOrder string `form:"sort_order"`
}

// CreatePaymentIntentRequest mirrors the upstream request to create a payment intent.
type CreatePaymentIntentRequest struct {
	PaymentMethod *string `json:"payment_method,omitempty"`
	Confirm       *bool   `json:"confirm,omitempty"`
}

// ConfirmPaymentRequest mirrors the upstream confirm payment payload.
type ConfirmPaymentRequest struct {
	PaymentMethodID string `json:"payment_method_id" binding:"required"`
}

// PaymentHistoryQuery captures query params for fetching payment history.
type PaymentHistoryQuery struct {
	Limit  int    `form:"limit"`
	Offset int    `form:"offset"`
	Page   int    `form:"page"`
	Status string `form:"status"`
}

// ValidateCouponRequest represents the payload to validate a coupon.
// The user ID is injected by the controller before forwarding to the order service.
type ValidateCouponRequest struct {
	Code        string   `json:"code" binding:"required"`
	OrderAmount int64    `json:"order_amount" binding:"required,min=0"`
	CourseIDs   []string `json:"course_ids,omitempty"`
	UserID      string   `json:"user_id,omitempty"`
}

// CouponListQuery captures pagination query params for coupons.
type CouponListQuery struct {
	Limit  int `form:"limit"`
	Offset int `form:"offset"`
	Page   int `form:"page"`
}
