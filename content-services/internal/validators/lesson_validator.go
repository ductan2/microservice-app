package validators

import (
	"errors"
	"strings"

	"content-services/internal/dto"
	"content-services/internal/utils"
)

// ValidateCreateLessonRequest validates the input for creating a lesson
func ValidateCreateLessonRequest(req *dto.CreateLessonRequest) error {
	if req == nil {
		return errors.New("lesson request is required")
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
	if req.Description != nil && len(strings.TrimSpace(*req.Description)) > 1000 {
		return errors.New("description cannot be longer than 1000 characters")
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

	if req.CreatedBy != nil {
		if _, err := utils.ValidateUUID(*req.CreatedBy); err != nil {
			return errors.New("invalid created_by ID format")
		}
	}

	// Code validation if provided
	if req.Code != nil && strings.TrimSpace(*req.Code) == "" {
		return errors.New("code cannot be empty if provided")
	}

	return nil
}

// ValidateUpdateLessonRequest validates the input for updating a lesson
func ValidateUpdateLessonRequest(req *dto.UpdateLessonRequest) error {
	if req == nil {
		return errors.New("lesson update request is required")
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
		if len(description) > 1000 {
			return errors.New("description cannot be longer than 1000 characters")
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

	return nil
}

// ValidateLessonSectionRequest validates lesson section input
func ValidateLessonSectionRequest(req *dto.CreateLessonSectionRequest) error {
	if req == nil {
		return errors.New("lesson section request is required")
	}

	// Lesson ID is required
	if strings.TrimSpace(req.LessonID) == "" {
		return errors.New("lesson ID is required")
	}

	// Validate lesson ID format
	if _, err := utils.ValidateUUID(req.LessonID); err != nil {
		return errors.New("invalid lesson ID format")
	}

	// Order must be non-negative
	if req.Order < 0 {
		return errors.New("order must be non-negative")
	}

	// Type is required
	if strings.TrimSpace(req.Type) == "" {
		return errors.New("section type is required")
	}

	// Validate section type
	validTypes := map[string]bool{
		"text": true, "dialog": true, "audio": true, "image": true, "exercise": true,
	}
	if !validTypes[strings.ToLower(req.Type)] {
		return errors.New("invalid section type, must be: text, dialog, audio, image, or exercise")
	}

	// Body validation
	if req.Body == nil {
		return errors.New("section body is required")
	}

	return nil
}

// ValidateLessonFilterRequest validates lesson filter input
func ValidateLessonFilterRequest(req *dto.LessonFilterRequest) error {
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

	if req.CreatedBy != nil {
		if _, err := utils.ValidateUUID(*req.CreatedBy); err != nil {
			return errors.New("invalid created_by ID format")
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