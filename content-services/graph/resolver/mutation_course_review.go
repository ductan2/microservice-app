package resolver

import (
	"content-services/graph/model"
	"context"

	"github.com/google/uuid"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

// SubmitCourseReview is the resolver for the submitCourseReview field.
func (r *mutationResolver) SubmitCourseReview(ctx context.Context, input model.SubmitCourseReviewInput) (*model.CourseReview, error) {
	if r.CourseReviewService == nil {
		return nil, gqlerror.Errorf("course review service not configured")
	}

	userID, err := userIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	courseID, err := uuid.Parse(input.CourseID)
	if err != nil {
		return nil, gqlerror.Errorf("invalid course ID: %v", err)
	}

	review, err := r.CourseReviewService.SubmitReview(ctx, courseID, userID, input.Rating, derefString(input.Comment))
	if err != nil {
		return nil, mapCourseReviewError(err)
	}

	return mapCourseReview(review), nil
}

// DeleteCourseReview is the resolver for the deleteCourseReview field.
func (r *mutationResolver) DeleteCourseReview(ctx context.Context, courseID string) (bool, error) {
	if r.CourseReviewService == nil {
		return false, gqlerror.Errorf("course review service not configured")
	}

	userID, err := userIDFromContext(ctx)
	if err != nil {
		return false, err
	}

	id, err := uuid.Parse(courseID)
	if err != nil {
		return false, gqlerror.Errorf("invalid course ID: %v", err)
	}

	if err := r.CourseReviewService.DeleteReview(ctx, id, userID); err != nil {
		return false, mapCourseReviewError(err)
	}

	return true, nil
}
