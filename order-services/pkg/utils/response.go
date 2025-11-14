package utils

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"order-services/internal/dto"
)

// APIResponseWithMeta represents an API response with metadata
type APIResponseWithMeta struct {
	dto.APIResponse
	Meta interface{} `json:"meta,omitempty"`
}

// GetCurrentTimestamp returns the current timestamp
func GetCurrentTimestamp() time.Time {
	return time.Now()
}

// SuccessResponse sends a successful JSON response
func SuccessResponse(c *gin.Context, statusCode int, data interface{}, messages ...string) {
	response := dto.APIResponse{
		Success:   true,
		Data:      data,
		Timestamp: GetCurrentTimestamp(),
	}

	if len(messages) > 0 && messages[0] != "" {
		response.Message = messages[0]
	}

	c.JSON(statusCode, response)
}

// SuccessResponseWithMeta sends a successful JSON response with metadata
func SuccessResponseWithMeta(c *gin.Context, statusCode int, data interface{}, meta interface{}) {
	response := APIResponseWithMeta{
		APIResponse: dto.APIResponse{
			Success:   true,
			Data:      data,
			Timestamp: GetCurrentTimestamp(),
		},
		Meta: meta,
	}

	c.JSON(statusCode, response)
}

// ErrorResponse sends an error JSON response
func ErrorResponse(c *gin.Context, statusCode int, errorCode, message string) {
	response := dto.APIResponse{
		Success:   false,
		Error:     &dto.ErrorInfo{Code: errorCode, Message: message},
		Timestamp: GetCurrentTimestamp(),
	}

	c.JSON(statusCode, response)
}

// ValidationError handles validation errors
func ValidationError(c *gin.Context, err error) {
	ErrorResponse(c, http.StatusBadRequest, dto.ErrCodeValidationFailed, err.Error())
}

// Error type checking functions
func IsValidationError(err error) bool {
	return false // Placeholder - would check for specific validation error types
}

func IsNotFoundError(err error) bool {
	return false // Placeholder - would check for specific not found error types
}

func IsUnauthorizedError(err error) bool {
	return false // Placeholder - would check for specific unauthorized error types
}

func IsConflictError(err error) bool {
	return false // Placeholder - would check for specific conflict error types
}

func IsExpiredError(err error) bool {
	return false // Placeholder - would check for specific expired error types
}

func IsInactiveError(err error) bool {
	return false // Placeholder - would check for specific inactive error types
}

// IsAdmin checks if the current user has admin role
func IsAdmin(c *gin.Context) bool {
	role, exists := c.Get("user_role")
	return exists && role == "admin"
}

// Custom error types for better error handling
type (
	AppValidationError struct {
		Field   string
		Message string
		Value   interface{}
	}

	NotFoundError struct {
		Entity string
		ID     string
	}

	UnauthorizedError struct {
		Message string
	}

	ConflictError struct {
		Message string
	}

	ExpiredError struct {
		Entity string
	}

	InactiveError struct {
		Entity string
	}
)

func (e AppValidationError) Error() string {
	return e.Message
}

func (e NotFoundError) Error() string {
	return e.Entity + " not found"
}

func (e UnauthorizedError) Error() string {
	return e.Message
}

func (e ConflictError) Error() string {
	return e.Message
}

func (e ExpiredError) Error() string {
	return e.Entity + " has expired"
}

func (e InactiveError) Error() string {
	return e.Entity + " is inactive"
}

// Error checking functions with proper type checking
func IsValidationErrorType(err error) bool {
	var validationErr AppValidationError
	return errors.As(err, &validationErr)
}

func IsNotFoundErrorType(err error) bool {
	var notFoundErr NotFoundError
	return errors.As(err, &notFoundErr)
}

func IsUnauthorizedErrorType(err error) bool {
	var unauthorizedErr UnauthorizedError
	return errors.As(err, &unauthorizedErr)
}

func IsConflictErrorType(err error) bool {
	var conflictErr ConflictError
	return errors.As(err, &conflictErr)
}

func IsExpiredErrorType(err error) bool {
	var expiredErr ExpiredError
	return errors.As(err, &expiredErr)
}

func IsInactiveErrorType(err error) bool {
	var inactiveErr InactiveError
	return errors.As(err, &inactiveErr)
}