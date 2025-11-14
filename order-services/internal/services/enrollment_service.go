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

	"order-services/internal/config"
	"order-services/internal/models"
)

// EnrollmentService handles integration with the lesson enrollment service
type EnrollmentService interface {
	CreateEnrollmentsForPaidOrder(ctx context.Context, order *models.Order) error
	GetUserEnrollments(ctx context.Context, userID uuid.UUID) ([]Enrollment, error)
	CheckExistingEnrollment(ctx context.Context, userID, courseID uuid.UUID) (bool, error)
}

// Enrollment represents a user's enrollment in a course
type Enrollment struct {
	ID         uuid.UUID `json:"id"`
	UserID     uuid.UUID `json:"user_id"`
	CourseID   uuid.UUID `json:"course_id"`
	OrderID    uuid.UUID `json:"order_id"`
	Status     string    `json:"status"` // active, completed, suspended, cancelled
	EnrolledAt time.Time `json:"enrolled_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// enrollmentService implements EnrollmentService
type enrollmentService struct {
	baseURL    string
	httpClient *http.Client
	config     *config.Config
}

// NewEnrollmentService creates a new enrollment service instance
func NewEnrollmentService(config *config.Config) EnrollmentService {
	return &enrollmentService{
		baseURL: getEnrollmentServiceURL(config),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		config: config,
	}
}

// CreateEnrollmentsForPaidOrder creates enrollments for all courses in a paid order
func (s *enrollmentService) CreateEnrollmentsForPaidOrder(ctx context.Context, order *models.Order) error {
	if order.Status != models.OrderStatusPaid {
		return fmt.Errorf("order is not paid: %s", order.Status)
	}

	if len(order.OrderItems) == 0 {
		return fmt.Errorf("no items in order")
	}

	// Create enrollment requests for each course item
	enrollmentRequests := make([]CreateEnrollmentRequest, 0, len(order.OrderItems))
	for _, item := range order.OrderItems {
		if item.ItemType != models.OrderItemTypeCourse {
			continue // Skip non-course items
		}

		// Check if user is already enrolled
		alreadyEnrolled, err := s.CheckExistingEnrollment(ctx, order.UserID, item.CourseID)
		if err != nil {
			log.Printf("Warning: Failed to check existing enrollment for user %s, course %s: %v",
				order.UserID, item.CourseID, err)
			// Continue with enrollment attempt
		} else if alreadyEnrolled {
			log.Printf("Info: User %s already enrolled in course %s", order.UserID, item.CourseID)
			continue // Skip existing enrollment
		}

		enrollmentRequests = append(enrollmentRequests, CreateEnrollmentRequest{
			UserID:   order.UserID,
			CourseID: item.CourseID,
			OrderID:  order.ID,
			Source:   "order_payment",
		})
	}

	if len(enrollmentRequests) == 0 {
		log.Printf("Info: No new enrollments to create for order %s", order.ID)
		return nil
	}

	// Create batch enrollment request
	batchRequest := BatchCreateEnrollmentRequest{
		Enrollments: enrollmentRequests,
	}

	return s.createBatchEnrollments(ctx, batchRequest)
}

// GetUserEnrollments retrieves all enrollments for a user
func (s *enrollmentService) GetUserEnrollments(ctx context.Context, userID uuid.UUID) ([]Enrollment, error) {
	url := fmt.Sprintf("%s/api/v1/enrollments/user/%s", s.baseURL, userID.String())

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Internal-Service", "order-service")
	req.Header.Set("Authorization", "Bearer "+s.getInternalAuthToken())

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get enrollments: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("enrollment service returned status: %d", resp.StatusCode)
	}

	var response struct {
		Enrollments []Enrollment `json:"enrollments"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode enrollment response: %w", err)
	}

	return response.Enrollments, nil
}

// CheckExistingEnrollment checks if a user is already enrolled in a course
func (s *enrollmentService) CheckExistingEnrollment(ctx context.Context, userID, courseID uuid.UUID) (bool, error) {
	url := fmt.Sprintf("%s/api/v1/enrollments/check", s.baseURL)

	request := CheckEnrollmentRequest{
		UserID:   userID,
		CourseID: courseID,
	}

	// Marshal request body
	body, err := json.Marshal(request)
	if err != nil {
		return false, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return false, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Internal-Service", "order-service")
	req.Header.Set("Authorization", "Bearer "+s.getInternalAuthToken())

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to check enrollment: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return false, nil // Not enrolled
	}

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("enrollment service returned status: %d", resp.StatusCode)
	}

	var response CheckEnrollmentResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return false, fmt.Errorf("failed to decode response: %w", err)
	}

	return response.Enrolled, nil
}

// Internal request/response structs

type CreateEnrollmentRequest struct {
	UserID   uuid.UUID `json:"user_id"`
	CourseID uuid.UUID `json:"course_id"`
	OrderID  uuid.UUID `json:"order_id"`
	Source   string    `json:"source"` // order_payment, admin_add, etc.
}

type BatchCreateEnrollmentRequest struct {
	Enrollments []CreateEnrollmentRequest `json:"enrollments"`
}

type CheckEnrollmentRequest struct {
	UserID   uuid.UUID `json:"user_id"`
	CourseID uuid.UUID `json:"course_id"`
}

type CheckEnrollmentResponse struct {
	Enrolled   bool        `json:"enrolled"`
	Enrollment *Enrollment `json:"enrollment,omitempty"`
}

// Helper methods

func (s *enrollmentService) createBatchEnrollments(ctx context.Context, request BatchCreateEnrollmentRequest) error {
	url := fmt.Sprintf("%s/api/v1/enrollments/batch", s.baseURL)

	// Marshal request body
	body, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal enrollment request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create enrollment request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Internal-Service", "order-service")
	req.Header.Set("Authorization", "Bearer "+s.getInternalAuthToken())

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to create enrollments: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("enrollment service returned status: %d", resp.StatusCode)
	}

	return nil
}

func (s *enrollmentService) getInternalAuthToken() string {
	// In production, this would use proper internal service authentication
	// For now, return a mock token
	return "internal-service-token"
}

func getEnrollmentServiceURL(config *config.Config) string {
	// In production, this would come from environment variables or service discovery
	// For now, return a default URL
	return "http://lesson-services:8005"
}

// MockEnrollmentService implements a mock enrollment service for testing
type MockEnrollmentService struct {
	enrollments map[string]bool // user_id:course_id -> enrolled
}

// NewMockEnrollmentService creates a new mock enrollment service
func NewMockEnrollmentService() *MockEnrollmentService {
	return &MockEnrollmentService{
		enrollments: make(map[string]bool),
	}
}

// CreateEnrollmentsForPaidOrder creates enrollments for a paid order (mock implementation)
func (s *MockEnrollmentService) CreateEnrollmentsForPaidOrder(ctx context.Context, order *models.Order) error {
	if order.Status != models.OrderStatusPaid {
		return fmt.Errorf("order is not paid: %s", order.Status)
	}

	for _, item := range order.OrderItems {
		if item.ItemType == models.OrderItemTypeCourse {
			key := fmt.Sprintf("%s:%s", order.UserID, item.CourseID)
			s.enrollments[key] = true
			log.Printf("Mock: Created enrollment for user %s in course %s", order.UserID, item.CourseID)
		}
	}

	return nil
}

// GetUserEnrollments retrieves enrollments for a user (mock implementation)
func (s *MockEnrollmentService) GetUserEnrollments(ctx context.Context, userID uuid.UUID) ([]Enrollment, error) {
	var enrollments []Enrollment

	for key := range s.enrollments {
		if s.enrollments[key] {
			// Parse user_id from key
			// This is a simplified implementation
			enrollments = append(enrollments, Enrollment{
				ID:         uuid.New(),
				UserID:     userID,
				CourseID:   uuid.New(), // Mock course ID
				Status:     "active",
				EnrolledAt: time.Now(),
				UpdatedAt:  time.Now(),
			})
		}
	}

	return enrollments, nil
}

// CheckExistingEnrollment checks if user is enrolled in course (mock implementation)
func (s *MockEnrollmentService) CheckExistingEnrollment(ctx context.Context, userID, courseID uuid.UUID) (bool, error) {
	key := fmt.Sprintf("%s:%s", userID, courseID)
	return s.enrollments[key], nil
}
