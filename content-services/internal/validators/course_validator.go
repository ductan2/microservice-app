package validators

import (
	"errors"
	"strings"

	"content-services/internal/dto"
	"content-services/internal/utils"
)

// ValidateCreateCourseRequest validates the input for creating a course
func ValidateCreateCourseRequest(req *dto.CreateCourseRequest) error {
	if req == nil {
		return errors.New("course request is required")
	}

	// Title is required
	if strings.TrimSpace(req.Title) == "" {
		return errors.New("title is required")
	}

	// Title length validation
	if len(req.Title) > 255 {
		return errors.New("title cannot be longer than 255 characters")
	}

	// Description length validation
	if req.Description != nil && len(strings.TrimSpace(*req.Description)) > 2000 {
		return errors.New("description cannot be longer than 2000 characters")
	}

	// Validate UUID fields
	if req.TopicID != nil {
		if _, err := utils.ValidateUUID(*req.TopicID); err != nil {
			return errors.New("invalid topic ID format")
		}
	}

	if req.LevelID != nil {
		if _, err := utils.ValidateUUID(*req.LevelID); err != nil {
			return errors.New("invalid level ID format")
		}
	}

	if req.InstructorID != nil {
		if _, err := utils.ValidateUUID(*req.InstructorID); err != nil {
			return errors.New("invalid instructor ID format")
		}
	}

	// URL validation for thumbnail
	if req.ThumbnailURL != nil {
		thumbnailURL := strings.TrimSpace(*req.ThumbnailURL)
		if thumbnailURL != "" && (len(thumbnailURL) > 500 || !strings.HasPrefix(thumbnailURL, "http")) {
			return errors.New("invalid thumbnail URL format")
		}
	}

	// Price validation
	if req.Price != nil && *req.Price < 0 {
		return errors.New("price cannot be negative")
	}

	// Duration validation
	if req.DurationHours != nil && *req.DurationHours < 0 {
		return errors.New("duration hours cannot be negative")
	}

	return nil
}

// ValidateUpdateCourseRequest validates the input for updating a course
func ValidateUpdateCourseRequest(req *dto.UpdateCourseRequest) error {
	if req == nil {
		return errors.New("course update request is required")
	}

	// Title validation if provided
	if req.Title != nil {
		title := strings.TrimSpace(*req.Title)
		if title == "" {
			return errors.New("title cannot be empty")
		}
		if len(title) > 255 {
			return errors.New("title cannot be longer than 255 characters")
		}
	}

	// Description validation if provided
	if req.Description != nil {
		description := strings.TrimSpace(*req.Description)
		if len(description) > 2000 {
			return errors.New("description cannot be longer than 2000 characters")
		}
	}

	// Validate UUID fields if provided
	if req.TopicID != nil {
		if _, err := utils.ValidateUUID(*req.TopicID); err != nil {
			return errors.New("invalid topic ID format")
		}
	}

	if req.LevelID != nil {
		if _, err := utils.ValidateUUID(*req.LevelID); err != nil {
			return errors.New("invalid level ID format")
		}
	}

	if req.InstructorID != nil {
		if _, err := utils.ValidateUUID(*req.InstructorID); err != nil {
			return errors.New("invalid instructor ID format")
		}
	}

	// URL validation for thumbnail if provided
	if req.ThumbnailURL != nil {
		thumbnailURL := strings.TrimSpace(*req.ThumbnailURL)
		if thumbnailURL != "" && (len(thumbnailURL) > 500 || !strings.HasPrefix(thumbnailURL, "http")) {
			return errors.New("invalid thumbnail URL format")
		}
	}

	// Price validation if provided
	if req.Price != nil && *req.Price < 0 {
		return errors.New("price cannot be negative")
	}

	// Duration validation if provided
	if req.DurationHours != nil && *req.DurationHours < 0 {
		return errors.New("duration hours cannot be negative")
	}

	return nil
}

// ValidateCourseReviewRequest validates course review input
func ValidateCourseReviewRequest(req *dto.CreateCourseReviewRequest) error {
	if req == nil {
		return errors.New("course review request is required")
	}

	// Course ID is required
	if strings.TrimSpace(req.CourseID) == "" {
		return errors.New("course ID is required")
	}

	// Validate course ID format
	if _, err := utils.ValidateUUID(req.CourseID); err != nil {
		return errors.New("invalid course ID format")
	}

	// Rating validation
	if req.Rating < 1 || req.Rating > 5 {
		return errors.New("rating must be between 1 and 5")
	}

	// Comment length validation if provided
	if req.Comment != nil {
		comment := strings.TrimSpace(*req.Comment)
		if len(comment) > 1000 {
			return errors.New("comment cannot be longer than 1000 characters")
		}
	}

	return nil
}

// ValidateAddCourseLessonRequest validates adding a lesson to a course
func ValidateAddCourseLessonRequest(req *dto.AddCourseLessonRequest) error {
	if req == nil {
		return errors.New("add course lesson request is required")
	}

	// Course ID is required
	if strings.TrimSpace(req.CourseID) == "" {
		return errors.New("course ID is required")
	}

	// Lesson ID is required
	if strings.TrimSpace(req.LessonID) == "" {
		return errors.New("lesson ID is required")
	}

	// Validate UUID formats
	if _, err := utils.ValidateUUID(req.CourseID); err != nil {
		return errors.New("invalid course ID format")
	}

	if _, err := utils.ValidateUUID(req.LessonID); err != nil {
		return errors.New("invalid lesson ID format")
	}

	// Order validation
	if req.Order < 0 {
		return errors.New("order must be non-negative")
	}

	return nil
}

// ValidateCourseFilterRequest validates course filter input
func ValidateCourseFilterRequest(req *dto.CourseFilterRequest) error {
	if req == nil {
		return nil
	}

	// Validate UUID fields if provided
	if req.TopicID != nil {
		if _, err := utils.ValidateUUID(*req.TopicID); err != nil {
			return errors.New("invalid topic ID format")
		}
	}

	if req.LevelID != nil {
		if _, err := utils.ValidateUUID(*req.LevelID); err != nil {
			return errors.New("invalid level ID format")
		}
	}

	if req.InstructorID != nil {
		if _, err := utils.ValidateUUID(*req.InstructorID); err != nil {
			return errors.New("invalid instructor ID format")
		}
	}

	// Search length validation
	if req.Search != nil {
		search := strings.TrimSpace(*req.Search)
		if len(search) > 100 {
			return errors.New("search term cannot be longer than 100 characters")
		}
	}

	return nil
}