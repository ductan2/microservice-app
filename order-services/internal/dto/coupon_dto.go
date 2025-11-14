package dto

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"order-services/internal/models"
)

// CouponResponse represents a coupon in the response
type CouponResponse struct {
	ID                   uuid.UUID                `json:"id"`
	Code                 string                   `json:"code"`
	Name                 string                   `json:"name"`
	Description          *string                  `json:"description,omitempty"`
	Type                 string                   `json:"type"`                   // percentage, fixed_amount
	PercentOff           *int                     `json:"percent_off,omitempty"`
	AmountOff            *int64                   `json:"amount_off,omitempty"`    // in cents
	Currency             *string                  `json:"currency,omitempty"`
	MaxRedemptions       *int                     `json:"max_redemptions,omitempty"`
	PerUserLimit         *int                     `json:"per_user_limit,omitempty"`
	RedemptionCount      int                      `json:"redemption_count"`
	MinimumAmount        *int64                   `json:"minimum_amount,omitempty"` // in cents
	ApplicableCourseIDs  []uuid.UUID              `json:"applicable_course_ids,omitempty"`
	ApplicableCourseType string                   `json:"applicable_course_type"`   // all, specific, category
	FirstTimeOnly        bool                     `json:"first_time_only"`
	ValidFrom            time.Time                `json:"valid_from"`
	ExpiresAt            *time.Time               `json:"expires_at,omitempty"`
	IsActive             bool                     `json:"is_active"`
	CreatedAt            time.Time                `json:"created_at"`
	UpdatedAt            time.Time                `json:"updated_at"`
	UserRedemptionCount  *int                     `json:"user_redemption_count,omitempty"` // Per-user usage
	CanRedeem            *bool                    `json:"can_redeem,omitempty"`           // Whether current user can redeem
	RedemptionMessage    *string                  `json:"redemption_message,omitempty"`   // Message about redemption status
}

// ValidateCouponRequest represents the request to validate a coupon
type ValidateCouponRequest struct {
	Code       string    `json:"code" validate:"required,min=3,max=50"`
	UserID     uuid.UUID `json:"user_id" validate:"required,uuid4"`
	OrderAmount int64    `json:"order_amount" validate:"required,min=0"` // in cents
	CourseIDs  []uuid.UUID `json:"course_ids,omitempty" validate:"omitempty"`
}

// ValidateCouponResponse represents the response for coupon validation
type ValidateCouponResponse struct {
	Valid         bool               `json:"valid"`
	Coupon        *CouponResponse    `json:"coupon,omitempty"`
	DiscountAmount int64             `json:"discount_amount"` // in cents
	FinalAmount    int64             `json:"final_amount"`    // order amount after discount
	Message        string            `json:"message"`
	AppliedCourses []uuid.UUID       `json:"applied_courses,omitempty"` // Courses the coupon applies to
}

// CouponListResponse represents a paginated list of coupons
type CouponListResponse struct {
	Coupons []CouponResponse `json:"coupons"`
	Total   int64            `json:"total"`
	Limit   int              `json:"limit"`
	Offset  int              `json:"offset"`
}

// CreateCouponRequest represents the request to create a coupon (admin only)
type CreateCouponRequest struct {
	Code                 string                   `json:"code" validate:"required,min=3,max=50,alphanum"`
	Name                 string                   `json:"name" validate:"required,min=1,max=100"`
	Description          *string                  `json:"description,omitempty" validate:"omitempty,max=500"`
	Type                 string                   `json:"type" validate:"required,oneof=percentage fixed_amount"`
	PercentOff           *int                     `json:"percent_off,omitempty" validate:"omitempty,min=1,max=100"`
	AmountOff            *int64                   `json:"amount_off,omitempty" validate:"omitempty,min=1"` // in cents
	Currency             *string                  `json:"currency,omitempty" validate:"omitempty,len=3"`
	MaxRedemptions       *int                     `json:"max_redemptions,omitempty" validate:"omitempty,min=1"`
	PerUserLimit         *int                     `json:"per_user_limit,omitempty" validate:"omitempty,min=1"`
	MinimumAmount        *int64                   `json:"minimum_amount,omitempty" validate:"omitempty,min=1"` // in cents
	ApplicableCourseIDs  []uuid.UUID              `json:"applicable_course_ids,omitempty"`
	ApplicableCourseType string                   `json:"applicable_course_type" validate:"required,oneof=all specific category"`
	FirstTimeOnly        bool                     `json:"first_time_only"`
	ValidFrom            time.Time                `json:"valid_from" validate:"required"`
	ExpiresAt            *time.Time               `json:"expires_at,omitempty"`
	IsActive             bool                     `json:"is_active"`
}

// UpdateCouponRequest represents the request to update a coupon (admin only)
type UpdateCouponRequest struct {
	Name                 *string                  `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	Description          *string                  `json:"description,omitempty" validate:"omitempty,max=500"`
	MaxRedemptions       *int                     `json:"max_redemptions,omitempty" validate:"omitempty,min=1"`
	MinimumAmount        *int64                   `json:"minimum_amount,omitempty" validate:"omitempty,min=1"` // in cents
	ApplicableCourseIDs  *[]uuid.UUID             `json:"applicable_course_ids,omitempty"`
	ApplicableCourseType *string                  `json:"applicable_course_type,omitempty" validate:"omitempty,oneof=all specific category"`
	IsActive             *bool                    `json:"is_active,omitempty"`
	ExpiresAt            *time.Time               `json:"expires_at,omitempty"`
}

// UserCouponUsageResponse represents a user's coupon usage history
type UserCouponUsageResponse struct {
	UserID       uuid.UUID              `json:"user_id"`
	CouponUsages []CouponUsageResponse  `json:"coupon_usages"`
	TotalSavings int64                  `json:"total_savings"` // in cents
}

// CouponUsageResponse represents a single coupon usage
type CouponUsageResponse struct {
	Coupon          CouponResponse    `json:"coupon"`
	RedemptionCount int               `json:"redemption_count"`
	TotalDiscount   int64             `json:"total_discount"` // in cents
	LastUsedAt      *time.Time        `json:"last_used_at,omitempty"`
	CanRedeem       bool              `json:"can_redeem"`
	NextRedeemAt    *time.Time        `json:"next_redeem_at,omitempty"`
}

// CouponStatsResponse represents coupon statistics (admin)
type CouponStatsResponse struct {
	TotalCoupons        int64 `json:"total_coupons"`
	ActiveCoupons       int64 `json:"active_coupons"`
	TotalRedemptions    int64 `json:"total_redemptions"`
	TotalDiscountGiven  int64 `json:"total_discount_given"` // in cents
	AverageDiscount     int64 `json:"average_discount"`     // in cents
	TopUsedCoupons      []CouponUsageStats `json:"top_used_coupons"`
}

// CouponUsageStats represents usage statistics for a specific coupon
type CouponUsageStats struct {
	CouponID       uuid.UUID `json:"coupon_id"`
	Code           string    `json:"code"`
	Name           string    `json:"name"`
	RedemptionCount int      `json:"redemption_count"`
	TotalDiscount   int64     `json:"total_discount"` // in cents
	UniqueUsers     int       `json:"unique_users"`
}

// BulkCouponRequest represents a request to create multiple coupons (admin)
type BulkCouponRequest struct {
	BaseCoupon     CreateCouponRequest `json:"base_coupon" validate:"required"`
	Quantity       int                `json:"quantity" validate:"required,min=1,max=1000"`
	Prefix         *string            `json:"prefix,omitempty" validate:"omitempty,max=10"`
	CodePattern    string             `json:"code_pattern" validate:"required,oneof=random sequential custom"`
	CustomCodes    []string           `json:"custom_codes,omitempty" validate:"omitempty,max=1000"`
}

// BulkCouponResponse represents the response for bulk coupon creation
type BulkCouponResponse struct {
	CreatedCoupons []CouponResponse `json:"created_coupons"`
	FailedCodes    []string         `json:"failed_codes,omitempty"`
	TotalCreated   int              `json:"total_created"`
	TotalFailed    int              `json:"total_failed"`
}

// Conversion functions

// FromModel converts a Coupon model to CouponResponse
func (r *CouponResponse) FromModel(coupon *models.Coupon) {
	if coupon == nil {
		return
	}

	r.ID = coupon.ID
	r.Code = coupon.Code
	r.Name = coupon.Name
	r.Type = coupon.Type
	r.RedemptionCount = coupon.RedemptionCount
	r.ApplicableCourseType = coupon.ApplicableCourseType
	r.FirstTimeOnly = coupon.FirstTimeOnly
	r.ValidFrom = coupon.ValidFrom
	r.IsActive = coupon.IsActive
	r.CreatedAt = coupon.CreatedAt
	r.UpdatedAt = coupon.UpdatedAt

	// Copy applicable course IDs
	if len(coupon.ApplicableCourseIDs) > 0 {
		r.ApplicableCourseIDs = make([]uuid.UUID, len(coupon.ApplicableCourseIDs))
		copy(r.ApplicableCourseIDs, coupon.ApplicableCourseIDs)
	}

	// Optional fields
	if coupon.Description != "" {
		r.Description = &coupon.Description
	}
	if coupon.PercentOff != nil {
		r.PercentOff = coupon.PercentOff
	}
	if coupon.AmountOff != nil {
		r.AmountOff = coupon.AmountOff
	}
	if coupon.Currency != "" {
		r.Currency = &coupon.Currency
	}
	if coupon.MaxRedemptions != nil {
		r.MaxRedemptions = coupon.MaxRedemptions
	}
	if coupon.PerUserLimit != nil {
		r.PerUserLimit = coupon.PerUserLimit
	}
	if coupon.MinimumAmount != nil {
		r.MinimumAmount = coupon.MinimumAmount
	}
	if coupon.ExpiresAt.Valid {
		r.ExpiresAt = &coupon.ExpiresAt.Time
	}
}

// WithUserUsage adds user-specific usage information to the coupon response
func (r *CouponResponse) WithUserUsage(userRedemptionCount int, canRedeem bool, message string) {
	r.UserRedemptionCount = &userRedemptionCount
	r.CanRedeem = &canRedeem
	r.RedemptionMessage = &message
}

// Validate validates the CreateCouponRequest
func (req *CreateCouponRequest) Validate() error {
	// Validate coupon type specific fields
	switch req.Type {
	case "percentage":
		if req.PercentOff == nil {
			return fmt.Errorf("percent_off is required for percentage coupons")
		}
		if *req.PercentOff < 1 || *req.PercentOff > 100 {
			return fmt.Errorf("percent_off must be between 1 and 100")
		}
		if req.AmountOff != nil {
			return fmt.Errorf("amount_off should not be set for percentage coupons")
		}
	case "fixed_amount":
		if req.AmountOff == nil {
			return fmt.Errorf("amount_off is required for fixed amount coupons")
		}
		if *req.AmountOff < 1 {
			return fmt.Errorf("amount_off must be greater than 0")
		}
		if req.PercentOff != nil {
			return fmt.Errorf("percent_off should not be set for fixed amount coupons")
		}
		if req.Currency == nil {
			return fmt.Errorf("currency is required for fixed amount coupons")
		}
	default:
		return fmt.Errorf("invalid coupon type: %s", req.Type)
	}

	// Validate course applicability
	switch req.ApplicableCourseType {
	case "specific":
		if len(req.ApplicableCourseIDs) == 0 {
			return fmt.Errorf("applicable_course_ids is required when type is 'specific'")
		}
	case "all", "category":
		// No additional validation needed
	default:
		return fmt.Errorf("invalid applicable_course_type: %s", req.ApplicableCourseType)
	}

	// Validate dates
	if req.ExpiresAt != nil && !req.ExpiresAt.After(req.ValidFrom) {
		return fmt.Errorf("expires_at must be after valid_from")
	}

	return nil
}

// CalculateDiscount calculates the discount amount for this coupon
func (r *CouponResponse) CalculateDiscount(orderAmount int64) (int64, error) {
	if r.Type == models.CouponTypePercentage {
		if r.PercentOff == nil {
			return 0, fmt.Errorf("percentage coupon missing percent_off")
		}
		discount := (orderAmount * int64(*r.PercentOff)) / 100
		if discount > orderAmount {
			discount = orderAmount
		}
		return discount, nil
	} else if r.Type == models.CouponTypeFixedAmount {
		if r.AmountOff == nil {
			return 0, fmt.Errorf("fixed amount coupon missing amount_off")
		}
		if *r.AmountOff > orderAmount {
			return orderAmount, nil
		}
		return *r.AmountOff, nil
	}

	return 0, fmt.Errorf("unknown coupon type: %s", r.Type)
}