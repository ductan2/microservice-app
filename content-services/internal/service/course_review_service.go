package service

import (
	"content-services/internal/models"
	"content-services/internal/repository"
	"content-services/internal/types"
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
)

// CourseReviewService defines operations for handling course reviews.
type CourseReviewService interface {
	SubmitReview(ctx context.Context, courseID, userID uuid.UUID, rating int, comment string) (*models.CourseReview, error)
	DeleteReview(ctx context.Context, courseID, userID uuid.UUID) error
	ListReviews(ctx context.Context, courseID uuid.UUID, page, pageSize int) ([]models.CourseReview, int64, error)
	GetReviewByUser(ctx context.Context, courseID, userID uuid.UUID) (*models.CourseReview, error)
}

type courseReviewService struct {
	reviewRepo     repository.CourseReviewRepository
	courseRepo     repository.CourseRepository
	enrollmentRepo repository.CourseEnrollmentRepository
}

// NewCourseReviewService constructs a CourseReviewService instance.
func NewCourseReviewService(reviewRepo repository.CourseReviewRepository, courseRepo repository.CourseRepository, enrollmentRepo repository.CourseEnrollmentRepository) CourseReviewService {
	return &courseReviewService{
		reviewRepo:     reviewRepo,
		courseRepo:     courseRepo,
		enrollmentRepo: enrollmentRepo,
	}
}

func (s *courseReviewService) SubmitReview(ctx context.Context, courseID, userID uuid.UUID, rating int, comment string) (*models.CourseReview, error) {
	if rating < 1 || rating > 5 {
		return nil, types.ErrCourseReviewInvalidRating
	}

	if _, err := s.courseRepo.GetByID(ctx, courseID); err != nil {
		return nil, err
	}

	if s.enrollmentRepo != nil {
		enrolled, err := s.enrollmentRepo.IsUserEnrolled(ctx, courseID, userID)
		if err != nil {
			return nil, err
		}
		if !enrolled {
			return nil, types.ErrCourseReviewNotEnrolled
		}
	}

	existing, err := s.reviewRepo.GetByCourseAndUser(ctx, courseID, userID)
	if err != nil && err != types.ErrCourseReviewNotFound {
		return nil, err
	}

	now := time.Now().UTC()
	review := &models.CourseReview{
		CourseID:  courseID,
		UserID:    userID,
		Rating:    rating,
		Comment:   strings.TrimSpace(comment),
		UpdatedAt: now,
		CreatedAt: now,
	}

	if existing != nil {
		review.ID = existing.ID
		review.CreatedAt = existing.CreatedAt
	}

	if review.ID == uuid.Nil {
		review.ID = uuid.New()
	}

	if err := s.reviewRepo.Upsert(ctx, review); err != nil {
		return nil, err
	}

	return s.reviewRepo.GetByCourseAndUser(ctx, courseID, userID)
}

func (s *courseReviewService) DeleteReview(ctx context.Context, courseID, userID uuid.UUID) error {
	return s.reviewRepo.DeleteByCourseAndUser(ctx, courseID, userID)
}

func (s *courseReviewService) ListReviews(ctx context.Context, courseID uuid.UUID, page, pageSize int) ([]models.CourseReview, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	return s.reviewRepo.ListByCourse(ctx, courseID, pageSize, offset)
}

func (s *courseReviewService) GetReviewByUser(ctx context.Context, courseID, userID uuid.UUID) (*models.CourseReview, error) {
	review, err := s.reviewRepo.GetByCourseAndUser(ctx, courseID, userID)
	if err != nil {
		if err == types.ErrCourseReviewNotFound {
			return nil, nil
		}
		return nil, err
	}
	return review, nil
}
