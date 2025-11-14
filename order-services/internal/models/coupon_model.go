package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// Coupon represents discount coupons that can be applied to orders
type Coupon struct {
	ID                    uuid.UUID    `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	Code                  string       `gorm:"type:text;uniqueIndex;not null" json:"code"`
	Name                  string       `gorm:"type:text;not null" json:"name"`
	Description           string       `gorm:"type:text" json:"description,omitempty"`
	Type                  string       `gorm:"type:varchar(20);not null;check:type IN ('percentage','fixed_amount')" json:"type"`
	PercentOff            *int         `gorm:"type:int;check:percent_off > 0 AND percent_off <= 100" json:"percent_off,omitempty"` // percentage discount
	AmountOff             *int64       `gorm:"type:bigint;check:amount_off > 0" json:"amount_off,omitempty"` // fixed amount discount in cents
	Currency              string       `gorm:"type:varchar(3);default:'USD'" json:"currency,omitempty"`
	MaxRedemptions        *int         `gorm:"type:int;check:max_redemptions > 0" json:"max_redemptions,omitempty"` // global usage limit
	PerUserLimit          *int         `gorm:"type:int;check:per_user_limit > 0" json:"per_user_limit,omitempty"` // per user limit
	RedemptionCount       int          `gorm:"type:int;default:0;not null" json:"redemption_count"`
	MinimumAmount         *int64       `gorm:"type:bigint;check:minimum_amount > 0" json:"minimum_amount,omitempty"` // minimum order amount in cents
	ApplicableCourseIDs   []uuid.UUID  `gorm:"type:jsonb;serializer:json" json:"applicable_course_ids,omitempty"` // specific courses this applies to
	ApplicableCourseType  string       `gorm:"type:varchar(20);check:applicable_course_type IN ('all','specific','category')" json:"applicable_course_type"` // all, specific courses, or category
	FirstTimeOnly         bool         `gorm:"default:false;not null" json:"first_time_only"` // only for first-time customers
	ValidFrom             time.Time    `gorm:"not null" json:"valid_from"`
	ExpiresAt             sql.NullTime `gorm:"type:timestamptz" json:"expires_at,omitempty"`
	IsActive              bool         `gorm:"default:true;not null" json:"is_active"`
	CreatedAt             time.Time    `json:"created_at"`
	UpdatedAt             time.Time    `json:"updated_at"`

	// Relationships
	CouponRedemptions     []CouponRedemption `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;foreignKey:CouponID;references:ID" json:"coupon_redemptions,omitempty"`
}

// Coupon Type Constants
const (
	CouponTypePercentage  = "percentage"
	CouponTypeFixedAmount = "fixed_amount"
)

// Coupon Applicability Constants
const (
	CouponApplicabilityAll      = "all"
	CouponApplicabilitySpecific = "specific"
	CouponApplicabilityCategory = "category"
)

// CouponRedemption tracks when a coupon is used
type CouponRedemption struct {
	ID         uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	CouponID   uuid.UUID `gorm:"type:uuid;not null;index:coupon_redemptions_coupon_idx;constraint:OnDelete:CASCADE" json:"coupon_id"`
	UserID     uuid.UUID `gorm:"type:uuid;not null;index:coupon_redemptions_user_idx" json:"user_id"`
	OrderID    uuid.UUID `gorm:"type:uuid;not null;uniqueIndex;constraint:OnDelete:CASCADE" json:"order_id"`
	DiscountAmount int64  `gorm:"type:bigint;not null" json:"discount_amount"` // discount applied in cents
	RedeemedAt  time.Time `gorm:"default:now();not null" json:"redeemed_at"`

	// Relationships
	Coupon     Coupon     `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;foreignKey:CouponID;references:ID" json:"coupon,omitempty"`
}