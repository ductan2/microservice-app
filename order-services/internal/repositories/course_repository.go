package repositories

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// Course represents course information for order processing
type Course struct {
	ID          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Price       int64     `json:"price"`       // in cents
	InstructorID uuid.UUID `json:"instructor_id"`
	InstructorName string  `json:"instructor_name"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CourseRepository interface for course data access
type CourseRepository interface {
	GetByID(ctx context.Context, courseID uuid.UUID) (*Course, error)
	GetByIDs(ctx context.Context, courseIDs []uuid.UUID) ([]Course, error)
	ValidateCourses(ctx context.Context, courseIDs []uuid.UUID) error
}

// courseRepository implements CourseRepository using HTTP client to communicate with course service
type courseRepository struct {
	baseURL    string
	httpClient *http.Client
}

// NewCourseRepository creates a new course repository
func NewCourseRepository(baseURL string) CourseRepository {
	return &courseRepository{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetByID retrieves a course by ID from the course service
func (r *courseRepository) GetByID(ctx context.Context, courseID uuid.UUID) (*Course, error) {
	url := fmt.Sprintf("%s/api/v1/courses/%s", r.baseURL, courseID.String())

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	// Add internal service authentication if needed
	req.Header.Set("X-Internal-Service", "order-service")

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get course: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("course not found")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("course service returned status: %d", resp.StatusCode)
	}

	// Parse response (simplified - in production, use proper JSON unmarshaling)
	// var course Course
	// TODO: Implement proper response parsing
	// err = json.NewDecoder(resp.Body).Decode(&course)
	// if err != nil {
	//     return nil, fmt.Errorf("failed to parse course response: %w", err)
	// }

	// For now, return a mock course
	return &Course{
		ID:          courseID,
		Title:       "Mock Course",
		Description: "Mock course description",
		Price:       9999, // $99.99
		InstructorID: uuid.New(),
		InstructorName: "Mock Instructor",
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}, nil
}

// GetByIDs retrieves multiple courses by their IDs
func (r *courseRepository) GetByIDs(ctx context.Context, courseIDs []uuid.UUID) ([]Course, error) {
	if len(courseIDs) == 0 {
		return []Course{}, nil
	}

	// url := fmt.Sprintf("%s/api/v1/courses/batch", r.baseURL)

	// Create request body with course IDs
	// requestBody := map[string]interface{}{
		// "course_ids": courseIDs,
	// }

	// TODO: Implement proper HTTP request with JSON body
	// For now, return mock courses
	courses := make([]Course, len(courseIDs))
	for i, courseID := range courseIDs {
		courses[i] = Course{
			ID:          courseID,
			Title:       fmt.Sprintf("Course %d", i+1),
			Description: "Mock course description",
			Price:       9999,
			InstructorID: uuid.New(),
			InstructorName: "Mock Instructor",
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
	}

	return courses, nil
}

// ValidateCourses checks if all courses exist and are active
func (r *courseRepository) ValidateCourses(ctx context.Context, courseIDs []uuid.UUID) error {
	if len(courseIDs) == 0 {
		return fmt.Errorf("no course IDs provided")
	}

	courses, err := r.GetByIDs(ctx, courseIDs)
	if err != nil {
		return fmt.Errorf("failed to validate courses: %w", err)
	}

	if len(courses) != len(courseIDs) {
		return fmt.Errorf("some courses not found")
	}

	// Check if all courses are active
	for _, course := range courses {
		if !course.IsActive {
			return fmt.Errorf("course %s is not active", course.ID)
		}
	}

	return nil
}

// MockCourseRepository implements a mock course repository for testing
type MockCourseRepository struct {
	courses map[uuid.UUID]Course
}

// NewMockCourseRepository creates a new mock course repository
func NewMockCourseRepository() *MockCourseRepository {
	return &MockCourseRepository{
		courses: make(map[uuid.UUID]Course),
	}
}

// AddCourse adds a course to the mock repository
func (r *MockCourseRepository) AddCourse(course Course) {
	r.courses[course.ID] = course
}

// GetByID retrieves a course by ID from the mock repository
func (r *MockCourseRepository) GetByID(ctx context.Context, courseID uuid.UUID) (*Course, error) {
	course, exists := r.courses[courseID]
	if !exists {
		return nil, fmt.Errorf("course not found")
	}
	return &course, nil
}

// GetByIDs retrieves multiple courses by their IDs from the mock repository
func (r *MockCourseRepository) GetByIDs(ctx context.Context, courseIDs []uuid.UUID) ([]Course, error) {
	courses := make([]Course, 0, len(courseIDs))
	for _, courseID := range courseIDs {
		course, exists := r.courses[courseID]
		if exists {
			courses = append(courses, course)
		}
	}
	return courses, nil
}

// ValidateCourses checks if all courses exist and are active in the mock repository
func (r *MockCourseRepository) ValidateCourses(ctx context.Context, courseIDs []uuid.UUID) error {
	if len(courseIDs) == 0 {
		return fmt.Errorf("no course IDs provided")
	}

	for _, courseID := range courseIDs {
		course, exists := r.courses[courseID]
		if !exists {
			return fmt.Errorf("course %s not found", courseID)
		}
		if !course.IsActive {
			return fmt.Errorf("course %s is not active", courseID)
		}
	}

	return nil
}