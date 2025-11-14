package dto

import (
	"time"
)

// APIResponse represents a standard API response wrapper
type APIResponse struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Error     *ErrorInfo  `json:"error,omitempty"`
	Message   string      `json:"message,omitempty"`
	Meta      *Meta       `json:"meta,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
	RequestID string      `json:"request_id,omitempty"`
}

// ErrorInfo represents error information in API responses
type ErrorInfo struct {
	Code       string      `json:"code"`                 // Error code for programmatic handling
	Message    string      `json:"message"`              // User-friendly error message
	Details    interface{} `json:"details,omitempty"`     // Additional error details
	StackTrace string      `json:"stack_trace,omitempty"` // Stack trace (development only)
	Field      string      `json:"field,omitempty"`       // Field name for validation errors
}

// Meta represents pagination metadata
type Meta struct {
	Total       int64  `json:"total,omitempty"`
	Page        int    `json:"page,omitempty"`
	Limit       int    `json:"limit,omitempty"`
	Offset      int    `json:"offset,omitempty"`
	TotalPages  int    `json:"total_pages,omitempty"`
	HasNext     bool   `json:"has_next,omitempty"`
	HasPrevious bool   `json:"has_previous,omitempty"`
}

// ValidationError represents a single validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   interface{} `json:"value,omitempty"`
}

// ValidationErrorResponse represents a validation error response
type ValidationErrorResponse struct {
	Code    string           `json:"code"`
	Message string           `json:"message"`
	Errors  []ValidationError `json:"errors"`
}

// HealthResponse represents health check response
type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Version   string            `json:"version"`
	Services  map[string]string `json:"services"`
	Checks    map[string]bool   `json:"checks"`
}

// StatusResponse represents a simple status response
type StatusResponse struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// PaginationParams represents pagination parameters for requests
type PaginationParams struct {
	Limit  int `json:"limit" form:"limit" query:"limit" validate:"omitempty,min=1,max=100"`
	Offset int `json:"offset" form:"offset" query:"offset" validate:"omitempty,min=0"`
	Page   int `json:"page" form:"page" query:"page" validate:"omitempty,min=1"`
}

// SortParams represents sorting parameters for requests
type SortParams struct {
	SortBy string `json:"sort_by" form:"sort_by" query:"sort_by" validate:"omitempty"`
	SortOrder string `json:"sort_order" form:"sort_order" query:"sort_order" validate:"omitempty,oneof=asc desc"`
}

// DateRangeParams represents date range filtering parameters
type DateRangeParams struct {
	StartDate *time.Time `json:"start_date" form:"start_date" query:"start_date"`
	EndDate   *time.Time `json:"end_date" form:"end_date" query:"end_date"`
}

// FilterParams represents common filtering parameters
type FilterParams struct {
	Status    []string `json:"status" form:"status" query:"status"`
	UserID    *string  `json:"user_id" form:"user_id" query:"user_id"`
	CourseID  *string  `json:"course_id" form:"course_id" query:"course_id"`
	Currency  *string  `json:"currency" form:"currency" query:"currency"`
}

// RequestParams combines common request parameters
type RequestParams struct {
	PaginationParams `json:"pagination"`
	SortParams       `json:"sort"`
	DateRangeParams  `json:"date_range"`
	FilterParams     `json:"filter"`
}

// SuccessResponse creates a successful API response
func SuccessResponse(data interface{}, message ...string) *APIResponse {
	response := &APIResponse{
		Success:   true,
		Data:      data,
		Timestamp: time.Now(),
	}

	if len(message) > 0 {
		response.Message = message[0]
	}

	return response
}

// SuccessResponseWithMeta creates a successful API response with metadata
func SuccessResponseWithMeta(data interface{}, meta *Meta, message ...string) *APIResponse {
	response := SuccessResponse(data, message...)
	response.Meta = meta
	return response
}

// ErrorResponse creates an error API response
func ErrorResponse(errorCode, message string, details ...interface{}) *APIResponse {
	errorInfo := &ErrorInfo{
		Code:    errorCode,
		Message: message,
	}

	if len(details) > 0 {
		errorInfo.Details = details[0]
	}

	return &APIResponse{
		Success:   false,
		Error:     errorInfo,
		Timestamp: time.Now(),
	}
}

// CreateValidationErrorResponse creates a validation error response
func CreateValidationErrorResponse(message string, errors []ValidationError) *APIResponse {
	errorInfo := &ErrorInfo{
		Code:    "VALIDATION_ERROR",
		Message: message,
		Details: ValidationErrorResponse{
			Code:    "VALIDATION_ERROR",
			Message: message,
			Errors:  errors,
		},
	}

	return &APIResponse{
		Success:   false,
		Error:     errorInfo,
		Timestamp: time.Now(),
	}
}

// WithRequestID adds a request ID to the response
func (r *APIResponse) WithRequestID(requestID string) *APIResponse {
	r.RequestID = requestID
	return r
}

// Pagination utilities

// CalculatePagination calculates pagination values from limit and offset
func CalculatePagination(total, limit, offset int) *Meta {
	if limit <= 0 {
		limit = 20
	}

	totalPages := (total + limit - 1) / limit
	currentPage := (offset / limit) + 1

	return &Meta{
		Total:       int64(total),
		Limit:       limit,
		Offset:      offset,
		TotalPages:  totalPages,
		Page:        currentPage,
		HasNext:     currentPage < totalPages,
		HasPrevious: currentPage > 1,
	}
}

// CalculatePaginationFromPage calculates pagination values from page and limit
func CalculatePaginationFromPage(total, page, limit int) *Meta {
	if limit <= 0 {
		limit = 20
	}
	if page <= 0 {
		page = 1
	}

	totalPages := (total + limit - 1) / limit
	offset := (page - 1) * limit

	return &Meta{
		Total:       int64(total),
		Limit:       limit,
		Offset:      offset,
		TotalPages:  totalPages,
		Page:        page,
		HasNext:     page < totalPages,
		HasPrevious: page > 1,
	}
}

// Validation utilities

// NewValidationError creates a new validation error
func NewValidationError(field, message string, value interface{}) ValidationError {
	return ValidationError{
		Field:   field,
		Message: message,
		Value:   value,
	}
}

// Error codes
const (
	// General error codes
	ErrCodeInternalError      = "INTERNAL_ERROR"
	ErrCodeBadRequest         = "BAD_REQUEST"
	ErrCodeUnauthorized       = "UNAUTHORIZED"
	ErrCodeForbidden          = "FORBIDDEN"
	ErrCodeNotFound           = "NOT_FOUND"
	ErrCodeConflict           = "CONFLICT"
	ErrCodeValidationFailed   = "VALIDATION_FAILED"
	ErrCodeRateLimitExceeded  = "RATE_LIMIT_EXCEEDED"
	ErrCodeServiceUnavailable = "SERVICE_UNAVAILABLE"

	// Order-specific error codes
	ErrCodeOrderNotFound       = "ORDER_NOT_FOUND"
	ErrCodeOrderExpired        = "ORDER_EXPIRED"
	ErrCodeOrderCannotCancel   = "ORDER_CANNOT_CANCEL"
	ErrCodeInvalidOrderStatus  = "INVALID_ORDER_STATUS"
	ErrCodeEmptyOrder          = "EMPTY_ORDER"
	ErrCodeInvalidCourse       = "INVALID_COURSE"
	ErrCodePaymentRequired     = "PAYMENT_REQUIRED"

	// Payment-specific error codes
	ErrCodePaymentNotFound     = "PAYMENT_NOT_FOUND"
	ErrCodePaymentFailed       = "PAYMENT_FAILED"
	ErrCodeInvalidPaymentIntent = "INVALID_PAYMENT_INTENT"
	ErrCodeWebhookSignature    = "WEBHOOK_SIGNATURE_INVALID"
	ErrCodeDuplicateWebhook    = "DUPLICATE_WEBHOOK"
	ErrCodeStripeError         = "STRIPE_ERROR"

	// Coupon-specific error codes
	ErrCodeCouponNotFound      = "COUPON_NOT_FOUND"
	ErrCodeCouponExpired       = "COUPON_EXPIRED"
	ErrCodeCouponInactive      = "COUPON_INACTIVE"
	ErrCodeCouponNotStarted    = "COUPON_NOT_STARTED"
	ErrCodeCouponUsageExceeded = "COUPON_USAGE_EXCEEDED"
	ErrCodeUserUsageExceeded   = "USER_USAGE_EXCEEDED"
	ErrCodeMinimumAmountNotMet = "MINIMUM_AMOUNT_NOT_MET"
	ErrCodeFirstTimeOnly       = "FIRST_TIME_ONLY"
	ErrCodeCourseNotApplicable = "COURSE_NOT_APPLICABLE"

	// Event-specific error codes
	ErrCodeEventNotFound       = "EVENT_NOT_FOUND"
	ErrCodeEventPublishFailed  = "EVENT_PUBLISH_FAILED"
	ErrCodeQueueConnection     = "QUEUE_CONNECTION_FAILED"
	ErrCodeQueueChannel        = "QUEUE_CHANNEL_FAILED"
)

// Common validation messages
const (
	MsgRequiredField     = "This field is required"
	MsgInvalidUUID       = "Invalid UUID format"
	MsgInvalidEmail      = "Invalid email format"
	MsgInvalidDate       = "Invalid date format"
	MsgInvalidNumber     = "Invalid number format"
	MsgMinLength         = "Must be at least %d characters"
	MsgMaxLength         = "Must be no more than %d characters"
	MsgMinValue          = "Must be at least %d"
	MsgMaxValue          = "Must be no more than %d"
	MsgOneOf             = "Must be one of: %v"
	MsgInvalidFormat     = "Invalid format"
)