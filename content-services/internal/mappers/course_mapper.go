package mappers

import (
	"content-services/graph/model"
	"content-services/internal/models"
	"content-services/internal/utils"
	"strings"
)

// CourseToGraphQL converts models.Course to model.Course
func CourseToGraphQL(course *models.Course) *model.Course {
	if course == nil {
		return nil
	}

	var topicID *string
	if course.TopicID != nil {
		id := course.TopicID.String()
		topicID = &id
	}

	var levelID *string
	if course.LevelID != nil {
		id := course.LevelID.String()
		levelID = &id
	}

	var instructorID *string
	if course.InstructorID != nil {
		id := course.InstructorID.String()
		instructorID = &id
	}

	mapped := &model.Course{
		ID:            course.ID.String(),
		Title:         course.Title,
		Description:   utils.ToStringPtr(course.Description),
		TopicID:       topicID,
		LevelID:       levelID,
		InstructorID:  instructorID,
		ThumbnailURL:  utils.ToStringPtr(course.ThumbnailURL),
		IsPublished:   course.IsPublished,
		IsFeatured:    course.IsFeatured,
		Price:         utils.ToFloat64Ptr(course.Price),
		DurationHours: utils.ToIntPtr(course.DurationHours),
		ReviewCount:   course.ReviewCount,
		CreatedAt:     course.CreatedAt,
		UpdatedAt:     course.UpdatedAt,
	}

	if course.PublishedAt.Valid {
		mapped.PublishedAt = &course.PublishedAt.Time
	}

	if course.ReviewCount > 0 {
		mapped.AverageRating = utils.ToFloat64Ptr(course.AverageRating)
	}

	return mapped
}

// CourseLessonToGraphQL converts models.CourseLesson to model.CourseLesson
func CourseLessonToGraphQL(lesson *models.CourseLesson) *model.CourseLesson {
	if lesson == nil {
		return nil
	}

	return &model.CourseLesson{
		ID:         lesson.ID.String(),
		CourseID:   lesson.CourseID.String(),
		LessonID:   lesson.LessonID.String(),
		Ord:        lesson.Ord,
		IsRequired: lesson.IsRequired,
		CreatedAt:  lesson.CreatedAt,
	}
}

// CourseLessonsToGraphQL converts slice of models.CourseLesson to GraphQL models
func CourseLessonsToGraphQL(lessons []models.CourseLesson) []*model.CourseLesson {
	items := make([]*model.CourseLesson, 0, len(lessons))
	for i := range lessons {
		items = append(items, CourseLessonToGraphQL(&lessons[i]))
	}
	return items
}

// CourseReviewToGraphQL converts models.CourseReview to model.CourseReview
func CourseReviewToGraphQL(review *models.CourseReview) *model.CourseReview {
	if review == nil {
		return nil
	}

	comment := utils.ToStringPtr(review.Comment)
	if comment != nil {
		trimmed := strings.TrimSpace(*comment)
		if trimmed == "" {
			comment = nil
		} else {
			comment = &trimmed
		}
	}

	return &model.CourseReview{
		ID:        review.ID.String(),
		CourseID:  review.CourseID.String(),
		UserID:    review.UserID.String(),
		Rating:    review.Rating,
		Comment:   comment,
		CreatedAt: review.CreatedAt,
		UpdatedAt:  review.UpdatedAt,
	}
}

// CourseReviewsToGraphQL converts slice of models.CourseReview to GraphQL models
func CourseReviewsToGraphQL(reviews []models.CourseReview) []*model.CourseReview {
	out := make([]*model.CourseReview, 0, len(reviews))
	for i := range reviews {
		out = append(out, CourseReviewToGraphQL(&reviews[i]))
	}
	return out
}