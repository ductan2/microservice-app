package mappers

import (
	"strings"

	"content-services/graph/model"
	"content-services/internal/repository"
	"content-services/internal/utils"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

// BuildLessonFilter converts GraphQL filter input to repository filter
func BuildLessonFilter(input *model.LessonFilterInput) (*repository.LessonFilter, error) {
	if input == nil {
		return nil, nil
	}

	filter := &repository.LessonFilter{}

	if input.TopicID != nil && *input.TopicID != "" {
		topicID, err := utils.ValidateUUID(*input.TopicID)
		if err != nil {
			return nil, gqlerror.Errorf("invalid topic ID: %v", err)
		}
		filter.TopicID = &topicID
	}

	if input.LevelID != nil && *input.LevelID != "" {
		levelID, err := utils.ValidateUUID(*input.LevelID)
		if err != nil {
			return nil, gqlerror.Errorf("invalid level ID: %v", err)
		}
		filter.LevelID = &levelID
	}

	if input.IsPublished != nil {
		filter.IsPublished = input.IsPublished
	}

	if input.Search != nil {
		filter.Search = strings.TrimSpace(*input.Search)
	}

	if input.CreatedBy != nil && *input.CreatedBy != "" {
		createdBy, err := utils.ValidateUUID(*input.CreatedBy)
		if err != nil {
			return nil, gqlerror.Errorf("invalid createdBy: %v", err)
		}
		filter.CreatedBy = &createdBy
	}

	return filter, nil
}

// BuildLessonOrder converts GraphQL order input to repository sort option
func BuildLessonOrder(input *model.LessonOrderInput) *repository.SortOption {
	if input == nil {
		return nil
	}
	option := &repository.SortOption{Direction: MapOrderDirection(input.Direction)}
	switch input.Field {
	case model.LessonOrderFieldPublishedAt:
		option.Field = "published_at"
	case model.LessonOrderFieldVersion:
		option.Field = "version"
	default:
		option.Field = "created_at"
	}
	return option
}

// BuildLessonSectionFilter converts GraphQL section filter input to repository filter
func BuildLessonSectionFilter(input *model.LessonSectionFilterInput) *repository.LessonSectionFilter {
	if input == nil {
		return nil
	}
	filter := &repository.LessonSectionFilter{}
	if input.Type != nil {
		sectionType := NormalizeLessonSectionType(*input.Type)
		filter.Type = &sectionType
	}
	return filter
}

// BuildLessonSectionOrder converts GraphQL section order input to repository sort option
func BuildLessonSectionOrder(input *model.LessonSectionOrderInput) *repository.SortOption {
	if input == nil {
		return nil
	}
	option := &repository.SortOption{Direction: MapOrderDirection(input.Direction)}
	switch input.Field {
	case model.LessonSectionOrderFieldCreatedAt:
		option.Field = "created_at"
	default:
		option.Field = "ord"
	}
	return option
}

// BuildCourseFilter converts GraphQL filter input to repository filter
func BuildCourseFilter(input *model.CourseFilterInput) (*repository.CourseFilter, error) {
	if input == nil {
		return nil, nil
	}

	filter := &repository.CourseFilter{}

	if input.TopicID != nil && *input.TopicID != "" {
		id, err := utils.ValidateUUID(*input.TopicID)
		if err != nil {
			return nil, gqlerror.Errorf("invalid topicId: %v", err)
		}
		filter.TopicID = &id
	}

	if input.LevelID != nil && *input.LevelID != "" {
		id, err := utils.ValidateUUID(*input.LevelID)
		if err != nil {
			return nil, gqlerror.Errorf("invalid levelId: %v", err)
		}
		filter.LevelID = &id
	}

	if input.InstructorID != nil && *input.InstructorID != "" {
		id, err := utils.ValidateUUID(*input.InstructorID)
		if err != nil {
			return nil, gqlerror.Errorf("invalid instructorId: %v", err)
		}
		filter.InstructorID = &id
	}

	if input.IsPublished != nil {
		filter.IsPublished = input.IsPublished
	}

	if input.IsFeatured != nil {
		filter.IsFeatured = input.IsFeatured
	}

	if input.Search != nil {
		filter.Search = strings.TrimSpace(*input.Search)
	}

	return filter, nil
}

// BuildCourseOrder converts GraphQL order input to repository sort option
func BuildCourseOrder(input *model.CourseOrderInput) *repository.SortOption {
	if input == nil {
		return nil
	}
	option := &repository.SortOption{Direction: MapOrderDirection(input.Direction)}
	switch input.Field {
	case model.CourseOrderFieldUpdatedAt:
		option.Field = "updated_at"
	case model.CourseOrderFieldPublishedAt:
		option.Field = "published_at"
	case model.CourseOrderFieldTitle:
		option.Field = "title"
	case model.CourseOrderFieldPrice:
		option.Field = "price"
	default:
		option.Field = "created_at"
	}
	return option
}

// BuildCourseLessonFilter converts GraphQL course lesson filter input to repository filter
func BuildCourseLessonFilter(input *model.CourseLessonFilterInput) *repository.CourseLessonFilter {
	if input == nil {
		return nil
	}
	filter := &repository.CourseLessonFilter{}
	if input.IsRequired != nil {
		filter.IsRequired = input.IsRequired
	}
	return filter
}

// BuildCourseLessonOrder converts GraphQL course lesson order input to repository sort option
func BuildCourseLessonOrder(input *model.CourseLessonOrderInput) *repository.SortOption {
	if input == nil {
		return nil
	}
	option := &repository.SortOption{Direction: MapOrderDirection(input.Direction)}
	switch input.Field {
	case model.CourseLessonOrderFieldCreatedAt:
		option.Field = "created_at"
	default:
		option.Field = "ord"
	}
	return option
}

// BuildFlashcardSetFilter converts GraphQL filter input to repository filter
func BuildFlashcardSetFilter(input *model.FlashcardSetFilterInput) (*repository.FlashcardSetFilter, error) {
	if input == nil {
		return nil, nil
	}
	filter := &repository.FlashcardSetFilter{}
	if input.TopicID != nil && *input.TopicID != "" {
		id, err := utils.ValidateUUID(*input.TopicID)
		if err != nil {
			return nil, gqlerror.Errorf("invalid topicId: %v", err)
		}
		filter.TopicID = &id
	}
	if input.LevelID != nil && *input.LevelID != "" {
		id, err := utils.ValidateUUID(*input.LevelID)
		if err != nil {
			return nil, gqlerror.Errorf("invalid levelId: %v", err)
		}
		filter.LevelID = &id
	}
	if input.CreatedBy != nil && *input.CreatedBy != "" {
		id, err := utils.ValidateUUID(*input.CreatedBy)
		if err != nil {
			return nil, gqlerror.Errorf("invalid createdBy: %v", err)
		}
		filter.CreatedBy = &id
	}
	if input.Search != nil {
		filter.Search = strings.TrimSpace(*input.Search)
	}
	return filter, nil
}

// BuildFlashcardSetOrder converts GraphQL order input to repository sort option
func BuildFlashcardSetOrder(input *model.FlashcardSetOrderInput) *repository.SortOption {
	if input == nil {
		return nil
	}
	option := &repository.SortOption{Direction: MapOrderDirection(input.Direction)}
	switch input.Field {
	case model.FlashcardSetOrderFieldCardCount:
		option.Field = "card_count"
	default:
		option.Field = "created_at"
	}
	return option
}

// MapOrderDirection maps GraphQL order direction to repository sort direction
func MapOrderDirection(direction model.OrderDirection) repository.SortDirection {
	if direction == model.OrderDirectionAsc {
		return repository.SortAscending
	}
	return repository.SortDescending
}