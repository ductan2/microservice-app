package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"order-services/internal/models"
	"order-services/internal/repositories"
)

var (
	ErrCouponNotFound      = errors.New("coupon not found")
	ErrCouponExpired       = errors.New("coupon has expired")
	ErrCouponInactive      = errors.New("coupon is not active")
	ErrCouponNotStarted    = errors.New("coupon has not started yet")
	ErrCouponUsageExceeded = errors.New("coupon usage limit exceeded")
	ErrUserUsageExceeded   = errors.New("user has exceeded coupon usage limit")
	ErrMinimumAmountNotMet = errors.New("minimum order amount not met")
	ErrFirstTimeOnly       = errors.New("coupon is for first-time customers only")
	ErrCourseNotApplicable = errors.New("coupon not applicable to this course")
)

// CouponService defines the business logic interface for coupon management
type CouponService interface {
	ValidateCoupon(ctx context.Context, code string, userID uuid.UUID, orderAmount int64, courseIDs []uuid.UUID) (*models.Coupon, int64, error)
	GetCouponByCode(ctx context.Context, code string) (*models.Coupon, error)
	GetUserCouponUsage(ctx context.Context, couponID, userID uuid.UUID) (int, error)
	CheckCouponAvailability(ctx context.Context, couponID uuid.UUID) error
	CalculateDiscount(ctx context.Context, coupon *models.Coupon, orderAmount int64) (int64, error)
	ListAvailableCoupons(ctx context.Context, userID uuid.UUID) ([]models.Coupon, error)
	GetCoupon(ctx context.Context, couponID uuid.UUID) (*models.Coupon, error)
}

// couponService implements the coupon business logic
type couponService struct {
	couponRepo repositories.CouponRepository
	orderRepo  repositories.OrderRepository
}

// NewCouponService creates a new coupon service instance
func NewCouponService(
	couponRepo repositories.CouponRepository,
	orderRepo repositories.OrderRepository,
) CouponService {
	return &couponService{
		couponRepo: couponRepo,
		orderRepo:  orderRepo,
	}
}

// ValidateCoupon validates a coupon code and calculates the discount
func (s *couponService) ValidateCoupon(ctx context.Context, code string, userID uuid.UUID, orderAmount int64, courseIDs []uuid.UUID) (*models.Coupon, int64, error) {
	// Get coupon by code
	coupon, err := s.couponRepo.GetByCode(ctx, code)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get coupon: %w", err)
	}

	if coupon == nil {
		return nil, 0, ErrCouponNotFound
	}

	// Validate coupon availability
	if err := s.CheckCouponAvailability(ctx, coupon.ID); err != nil {
		return nil, 0, err
	}

	// Validate user-specific restrictions
	if err := s.validateUserRestrictions(ctx, coupon, userID, orderAmount, courseIDs); err != nil {
		return nil, 0, err
	}

	// Calculate discount amount
	discountAmount, err := s.CalculateDiscount(ctx, coupon, orderAmount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to calculate discount: %w", err)
	}

	return coupon, discountAmount, nil
}

// GetCouponByCode retrieves a coupon by its code
func (s *couponService) GetCouponByCode(ctx context.Context, code string) (*models.Coupon, error) {
	coupon, err := s.couponRepo.GetByCode(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to get coupon by code: %w", err)
	}

	return coupon, nil
}

// GetUserCouponUsage gets the number of times a user has used a specific coupon
func (s *couponService) GetUserCouponUsage(ctx context.Context, couponID, userID uuid.UUID) (int, error) {
	usage, err := s.couponRepo.GetUserRedemptionCount(ctx, couponID, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to get user coupon usage: %w", err)
	}

	return usage, nil
}

// CheckCouponAvailability checks if a coupon is currently available for use
func (s *couponService) CheckCouponAvailability(ctx context.Context, couponID uuid.UUID) error {
	coupon, err := s.couponRepo.GetByID(ctx, couponID)
	if err != nil {
		return fmt.Errorf("failed to get coupon: %w", err)
	}

	if coupon == nil {
		return ErrCouponNotFound
	}

	// Check if coupon is active
	if !coupon.IsActive {
		return ErrCouponInactive
	}

	// Check if coupon has started
	now := time.Now()
	if coupon.ValidFrom.After(now) {
		return ErrCouponNotStarted
	}

	// Check if coupon has expired
	if coupon.ExpiresAt.Valid && coupon.ExpiresAt.Time.Before(now) {
		return ErrCouponExpired
	}

	// Check global usage limit
	if coupon.MaxRedemptions != nil && coupon.RedemptionCount >= *coupon.MaxRedemptions {
		return ErrCouponUsageExceeded
	}

	return nil
}

// CalculateDiscount calculates the discount amount for a coupon
func (s *couponService) CalculateDiscount(ctx context.Context, coupon *models.Coupon, orderAmount int64) (int64, error) {
	if coupon == nil {
		return 0, ErrCouponNotFound
	}

	var discountAmount int64

	switch coupon.Type {
	case models.CouponTypePercentage:
		if coupon.PercentOff == nil {
			return 0, fmt.Errorf("percentage coupon missing percent_off value")
		}
		discountAmount = (orderAmount * int64(*coupon.PercentOff)) / 100

	case models.CouponTypeFixedAmount:
		if coupon.AmountOff == nil {
			return 0, fmt.Errorf("fixed amount coupon missing amount_off value")
		}
		discountAmount = *coupon.AmountOff

	default:
		return 0, fmt.Errorf("unknown coupon type: %s", coupon.Type)
	}

	// Ensure discount doesn't exceed order amount
	if discountAmount > orderAmount {
		discountAmount = orderAmount
	}

	return discountAmount, nil
}

// ListAvailableCoupons returns a list of coupons currently available for a user
func (s *couponService) ListAvailableCoupons(ctx context.Context, userID uuid.UUID) ([]models.Coupon, error) {
	coupons, err := s.couponRepo.GetActiveCoupons(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get active coupons: %w", err)
	}

	var availableCoupons []models.Coupon

	for _, coupon := range coupons {
		// Check basic availability
		if err := s.CheckCouponAvailability(ctx, coupon.ID); err != nil {
			continue
		}

		// Check user-specific restrictions
		if coupon.FirstTimeOnly {
			hasOrders, err := s.orderRepo.UserHasPreviousOrders(ctx, userID)
			if err != nil {
				continue // Skip this coupon if we can't check
			}
			if hasOrders {
				continue
			}
		}

		// Check per-user limit
		if coupon.PerUserLimit != nil {
			usage, err := s.GetUserCouponUsage(ctx, coupon.ID, userID)
			if err != nil {
				continue
			}
			if usage >= *coupon.PerUserLimit {
				continue
			}
		}

		availableCoupons = append(availableCoupons, coupon)
	}

	return availableCoupons, nil
}

// GetCoupon retrieves a coupon by ID
func (s *couponService) GetCoupon(ctx context.Context, couponID uuid.UUID) (*models.Coupon, error) {
	coupon, err := s.couponRepo.GetByID(ctx, couponID)
	if err != nil {
		return nil, fmt.Errorf("failed to get coupon: %w", err)
	}

	return coupon, nil
}

// Helper methods

// validateUserRestrictions validates user-specific coupon restrictions
func (s *couponService) validateUserRestrictions(ctx context.Context, coupon *models.Coupon, userID uuid.UUID, orderAmount int64, courseIDs []uuid.UUID) error {
	// Check minimum order amount
	if coupon.MinimumAmount != nil && orderAmount < *coupon.MinimumAmount {
		return fmt.Errorf("%w: minimum amount is %d cents", ErrMinimumAmountNotMet, *coupon.MinimumAmount)
	}

	// Check first-time customer restriction
	if coupon.FirstTimeOnly {
		hasOrders, err := s.orderRepo.UserHasPreviousOrders(ctx, userID)
		if err != nil {
			return fmt.Errorf("failed to check user order history: %w", err)
		}
		if hasOrders {
			return ErrFirstTimeOnly
		}
	}

	// Check per-user usage limit
	if coupon.PerUserLimit != nil {
		usage, err := s.GetUserCouponUsage(ctx, coupon.ID, userID)
		if err != nil {
			return fmt.Errorf("failed to check user coupon usage: %w", err)
		}
		if usage >= *coupon.PerUserLimit {
			return fmt.Errorf("%w: limit is %d", ErrUserUsageExceeded, *coupon.PerUserLimit)
		}
	}

	// Check course applicability
	if len(courseIDs) > 0 {
		if err := s.validateCourseApplicability(coupon, courseIDs); err != nil {
			return err
		}
	}

	return nil
}

// validateCourseApplicability checks if the coupon applies to the given courses
func (s *couponService) validateCourseApplicability(coupon *models.Coupon, courseIDs []uuid.UUID) error {
	switch coupon.ApplicableCourseType {
	case models.CouponApplicabilityAll:
		// Coupon applies to all courses
		return nil

	case models.CouponApplicabilitySpecific:
		// Check if any of the order courses match the applicable courses
		if len(coupon.ApplicableCourseIDs) == 0 {
			return ErrCourseNotApplicable
		}

		courseSet := make(map[uuid.UUID]bool)
		for _, id := range coupon.ApplicableCourseIDs {
			courseSet[id] = true
		}

		for _, orderCourseID := range courseIDs {
			if courseSet[orderCourseID] {
				return nil // At least one course is applicable
			}
		}

		return ErrCourseNotApplicable

	case models.CouponApplicabilityCategory:
		// TODO: Implement category-based validation when course categories are available
		// For now, we'll allow category-based coupons
		return nil

	default:
		return fmt.Errorf("unknown course applicability type: %s", coupon.ApplicableCourseType)
	}
}

// GetCouponWithUsage returns a coupon with detailed usage information
func (s *couponService) GetCouponWithUsage(ctx context.Context, couponID, userID uuid.UUID) (*models.Coupon, int, error) {
	coupon, err := s.GetCoupon(ctx, couponID)
	if err != nil {
		return nil, 0, err
	}

	userUsage, err := s.GetUserCouponUsage(ctx, couponID, userID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get user usage: %w", err)
	}

	return coupon, userUsage, nil
}

// IsCouponApplicableToCourse checks if a coupon can be applied to a specific course
func (s *couponService) IsCouponApplicableToCourse(ctx context.Context, coupon *models.Coupon, courseID uuid.UUID) error {
	if coupon == nil {
		return ErrCouponNotFound
	}

	switch coupon.ApplicableCourseType {
	case models.CouponApplicabilityAll:
		return nil

	case models.CouponApplicabilitySpecific:
		for _, applicableID := range coupon.ApplicableCourseIDs {
			if applicableID == courseID {
				return nil
			}
		}
		return ErrCourseNotApplicable

	case models.CouponApplicabilityCategory:
		// TODO: Implement category checking when course categories are available
		return nil

	default:
		return fmt.Errorf("unknown course applicability type: %s", coupon.ApplicableCourseType)
	}
}

// GetBestCoupon finds the best applicable coupon for an order
func (s *couponService) GetBestCoupon(ctx context.Context, userID uuid.UUID, orderAmount int64, courseIDs []uuid.UUID, couponCodes []string) (*models.Coupon, int64, error) {
	if len(couponCodes) == 0 {
		return nil, 0, nil
	}

	var bestCoupon *models.Coupon
	var bestDiscount int64

	for _, code := range couponCodes {
		coupon, discount, err := s.ValidateCoupon(ctx, code, userID, orderAmount, courseIDs)
		if err != nil {
			continue // Skip invalid coupons
		}

		if discount > bestDiscount {
			bestCoupon = coupon
			bestDiscount = discount
		}
	}

	return bestCoupon, bestDiscount, nil
}
