package types

import "errors"

var (
	// ErrCourseReviewNotFound indicates that the requested review could not be located.
	ErrCourseReviewNotFound = errors.New("course review: not found")
	// ErrCourseReviewNotEnrolled indicates the user hasn't enrolled in the course yet.
	ErrCourseReviewNotEnrolled = errors.New("course review: user not enrolled")
	// ErrCourseReviewInvalidRating indicates the provided rating is outside the accepted range.
	ErrCourseReviewInvalidRating = errors.New("course review: invalid rating")
)
