package response

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Response represents a standardized API response
type Response struct {
	Status    string      `json:"status"`
	Message   string      `json:"message,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
	RequestID string      `json:"request_id,omitempty"`
}

// Meta represents pagination metadata
type Meta struct {
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// PaginatedResponse represents a paginated API response
type PaginatedResponse struct {
	Response
	Meta Meta `json:"meta,omitempty"`
}

// Success sends a successful response with data
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Status:    "success",
		Data:      data,
		Timestamp: time.Now(),
	})
}

// SuccessWithMessage sends a successful response with message
func SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Status:    "success",
		Message:   message,
		Data:      data,
		Timestamp: time.Now(),
	})
}

// Created sends a created response with data
func Created(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusCreated, Response{
		Status:    "success",
		Message:   message,
		Data:      data,
		Timestamp: time.Now(),
	})
}

// Accepted sends an accepted response (for async operations)
func Accepted(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusAccepted, Response{
		Status:    "success",
		Message:   message,
		Data:      data,
		Timestamp: time.Now(),
	})
}

// NoContent sends a 204 No Content response
func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// Paginated sends a paginated response
func Paginated(c *gin.Context, data interface{}, meta Meta) {
	c.JSON(http.StatusOK, PaginatedResponse{
		Response: Response{
			Status:    "success",
			Data:      data,
			Timestamp: time.Now(),
		},
		Meta: meta,
	})
}

// BadRequest sends a 400 Bad Request response
func BadRequest(c *gin.Context, message string, details interface{}) {
	c.JSON(http.StatusBadRequest, gin.H{
		"status":    "error",
		"message":   message,
		"details":   details,
		"timestamp": time.Now(),
	})
}

// Unauthorized sends a 401 Unauthorized response
func Unauthorized(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, gin.H{
		"status":    "error",
		"message":   message,
		"timestamp": time.Now(),
	})
}

// Forbidden sends a 403 Forbidden response
func Forbidden(c *gin.Context, message string) {
	c.JSON(http.StatusForbidden, gin.H{
		"status":    "error",
		"message":   message,
		"timestamp": time.Now(),
	})
}

// NotFound sends a 404 Not Found response
func NotFound(c *gin.Context, message string) {
	c.JSON(http.StatusNotFound, gin.H{
		"status":    "error",
		"message":   message,
		"timestamp": time.Now(),
	})
}

// Conflict sends a 409 Conflict response
func Conflict(c *gin.Context, message string) {
	c.JSON(http.StatusConflict, gin.H{
		"status":    "error",
		"message":   message,
		"timestamp": time.Now(),
	})
}

// TooManyRequests sends a 429 Too Many Requests response
func TooManyRequests(c *gin.Context, message string) {
	c.JSON(http.StatusTooManyRequests, gin.H{
		"status":    "error",
		"message":   message,
		"timestamp": time.Now(),
	})
}

// InternalServerError sends a 500 Internal Server Error response
func InternalServerError(c *gin.Context, message string) {
	c.JSON(http.StatusInternalServerError, gin.H{
		"status":    "error",
		"message":   message,
		"timestamp": time.Now(),
	})
}

// ServiceUnavailable sends a 503 Service Unavailable response
func ServiceUnavailable(c *gin.Context, message string) {
	c.JSON(http.StatusServiceUnavailable, gin.H{
		"status":    "error",
		"message":   message,
		"timestamp": time.Now(),
	})
}

// ValidationError sends a validation error response
func ValidationError(c *gin.Context, details interface{}) {
	c.JSON(http.StatusBadRequest, gin.H{
		"status":    "error",
		"message":   "Validation failed",
		"details":   details,
		"timestamp": time.Now(),
	})
}