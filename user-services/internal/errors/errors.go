package errors

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrorType represents different categories of errors
type ErrorType string

const (
	ErrorTypeValidation    ErrorType = "validation"
	ErrorTypeAuthentication ErrorType = "authentication"
	ErrorTypeAuthorization  ErrorType = "authorization"
	ErrorTypeNotFound      ErrorType = "not_found"
	ErrorTypeConflict      ErrorType = "conflict"
	ErrorTypeRateLimit     ErrorType = "rate_limit"
	ErrorTypeInternal      ErrorType = "internal"
	ErrorTypeExternal      ErrorType = "external"
)

// AppError represents a structured application error
type AppError struct {
	Type       ErrorType     `json:"type"`
	Message    string        `json:"message"`
	Code       string        `json:"code,omitempty"`
	HTTPStatus int           `json:"-"`
	Details    interface{}   `json:"details,omitempty"`
	Cause      error         `json:"-"`
	Context    map[string]interface{} `json:"-"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Type, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// Unwrap returns the underlying cause
func (e *AppError) Unwrap() error {
	return e.Cause
}

// NewAppError creates a new application error
func NewAppError(errorType ErrorType, message string, httpStatus int) *AppError {
	return &AppError{
		Type:       errorType,
		Message:    message,
		HTTPStatus: httpStatus,
		Context:    make(map[string]interface{}),
	}
}

// WithCode adds an error code
func (e *AppError) WithCode(code string) *AppError {
	e.Code = code
	return e
}

// WithCause adds a cause error
func (e *AppError) WithCause(cause error) *AppError {
	e.Cause = cause
	return e
}

// WithDetails adds error details
func (e *AppError) WithDetails(details interface{}) *AppError {
	e.Details = details
	return e
}

// WithContext adds context information
func (e *AppError) WithContext(key string, value interface{}) *AppError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// Predefined error constructors
func NewValidationError(message string) *AppError {
	return NewAppError(ErrorTypeValidation, message, http.StatusBadRequest)
}

func NewAuthenticationError(message string) *AppError {
	return NewAppError(ErrorTypeAuthentication, message, http.StatusUnauthorized)
}

func NewAuthorizationError(message string) *AppError {
	return NewAppError(ErrorTypeAuthorization, message, http.StatusForbidden)
}

func NewNotFoundError(resource string) *AppError {
	return NewAppError(ErrorTypeNotFound, fmt.Sprintf("%s not found", resource), http.StatusNotFound)
}

func NewConflictError(message string) *AppError {
	return NewAppError(ErrorTypeConflict, message, http.StatusConflict)
}

func NewRateLimitError(message string) *AppError {
	return NewAppError(ErrorTypeRateLimit, message, http.StatusTooManyRequests)
}

func NewInternalError(message string) *AppError {
	return NewAppError(ErrorTypeInternal, message, http.StatusInternalServerError)
}

func NewExternalServiceError(service string, message string) *AppError {
	return NewAppError(ErrorTypeExternal, fmt.Sprintf("%s service error: %s", service, message), http.StatusBadGateway)
}

// Common application errors
var (
	ErrInvalidCredentials     = NewAuthenticationError("Invalid email or password").WithCode("INVALID_CREDENTIALS")
	ErrEmailNotVerified       = NewAuthenticationError("Email address not verified").WithCode("EMAIL_NOT_VERIFIED")
	ErrInvalidMFACode         = NewAuthenticationError("Invalid or expired MFA code").WithCode("INVALID_MFA_CODE")
	ErrMFANotSetup           = NewAuthenticationError("MFA not setup for this account").WithCode("MFA_NOT_SETUP")
	ErrSessionExpired        = NewAuthenticationError("Session has expired").WithCode("SESSION_EXPIRED")
	ErrTokenInvalid          = NewAuthenticationError("Invalid authentication token").WithCode("TOKEN_INVALID")
	ErrAccountLocked         = NewAuthenticationError("Account has been locked").WithCode("ACCOUNT_LOCKED")
	ErrAccountDisabled       = NewAuthenticationError("Account has been disabled").WithCode("ACCOUNT_DISABLED")

	ErrEmailExists           = NewConflictError("Email address already exists").WithCode("EMAIL_EXISTS")
	ErrWeakPassword          = NewValidationError("Password does not meet security requirements").WithCode("WEAK_PASSWORD")
	ErrInvalidEmail          = NewValidationError("Invalid email address format").WithCode("INVALID_EMAIL")
	ErrPasswordMismatch      = NewValidationError("Passwords do not match").WithCode("PASSWORD_MISMATCH")
	InvalidVerificationToken = NewValidationError("Invalid or expired verification token").WithCode("INVALID_VERIFICATION_TOKEN")
	InvalidPasswordResetToken = NewValidationError("Invalid or expired password reset token").WithCode("INVALID_PASSWORD_RESET_TOKEN")

	ErrUserNotFound          = NewNotFoundError("User").WithCode("USER_NOT_FOUND")
	ErrSessionNotFound       = NewNotFoundError("Session").WithCode("SESSION_NOT_FOUND")
	ErrMFAMethodNotFound     = NewNotFoundError("MFA method").WithCode("MFA_METHOD_NOT_FOUND")

	ErrDatabaseConnection    = NewInternalError("Database connection failed").WithCode("DATABASE_CONNECTION_ERROR")
	ErrCacheConnection       = NewInternalError("Cache connection failed").WithCode("CACHE_CONNECTION_ERROR")
	ErrQueueConnection       = NewInternalError("Message queue connection failed").WithCode("QUEUE_CONNECTION_ERROR")
	ErrEmailServiceFailed    = NewExternalServiceError("email", "Failed to send email").WithCode("EMAIL_SERVICE_FAILED")
)

// IsAppError checks if an error is an AppError
func IsAppError(err error) bool {
	_, ok := err.(*AppError)
	return ok
}

// GetAppError converts any error to AppError, wrapping if necessary
func GetAppError(err error) *AppError {
	if appErr, ok := err.(*AppError); ok {
		return appErr
	}
	return NewInternalError("An unexpected error occurred").WithCause(err)
}

// ErrorHandler middleware for Gin
func ErrorHandler() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		var err error

		switch e := recovered.(type) {
		case string:
			err = NewInternalError(e)
		case error:
			err = GetAppError(e)
		default:
			err = NewInternalError("An unexpected error occurred")
		}

		SendError(c, err)
	})
}

// SendError sends a standardized error response
func SendError(c *gin.Context, err error) {
	appErr := GetAppError(err)

	// Log internal errors with context
	if appErr.Type == ErrorTypeInternal || appErr.Type == ErrorTypeExternal {
		// In production, you'd want to use a proper logging framework
		// with structured logging and correlation IDs
		fmt.Printf("Internal Error: %+v\n", appErr)

		// Don't expose internal error details to clients
		if appErr.Type == ErrorTypeInternal {
			appErr.Message = "An internal error occurred"
			appErr.Details = nil
		}
	}

	response := gin.H{
		"status": "error",
		"message": appErr.Message,
	}

	if appErr.Code != "" {
		response["error_code"] = appErr.Code
	}

	if appErr.Details != nil {
		response["details"] = appErr.Details
	}

	c.JSON(appErr.HTTPStatus, response)
}