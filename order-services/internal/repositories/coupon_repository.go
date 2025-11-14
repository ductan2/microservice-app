package repositories

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/google/uuid"
	"order-services/internal/models"
)

// CouponRepository interface for coupon data access
type CouponRepository interface {
	Create(ctx context.Context, coupon *models.Coupon) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Coupon, error)
	GetByCode(ctx context.Context, code string) (*models.Coupon, error)
	GetActiveCoupons(ctx context.Context) ([]models.Coupon, error)
	Update(ctx context.Context, coupon *models.Coupon) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetUserRedemptionCount(ctx context.Context, couponID, userID uuid.UUID) (int, error)
	CreateRedemption(ctx context.Context, redemption *models.CouponRedemption) error
	GetRedemptionsByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.CouponRedemption, int64, error)
	GetRedemptionsByCoupon(ctx context.Context, couponID uuid.UUID, limit, offset int) ([]models.CouponRedemption, int64, error)
	UpdateUsageCount(ctx context.Context, couponID uuid.UUID) error
	GetCouponStats(ctx context.Context, timeRange *TimeRange) (*CouponStats, error)
}

// CouponStats represents aggregated coupon statistics
type CouponStats struct {
	TotalCoupons        int64 `json:"total_coupons"`
	ActiveCoupons       int64 `json:"active_coupons"`
	TotalRedemptions    int64 `json:"total_redemptions"`
	TotalDiscountAmount int64 `json:"total_discount_amount"` // in cents
	AverageDiscount     int64 `json:"average_discount"`     // in cents
}

// couponRepository implements CouponRepository
type couponRepository struct {
	db *gorm.DB
}

// NewCouponRepository creates a new coupon repository
func NewCouponRepository(db *gorm.DB) CouponRepository {
	return &couponRepository{db: db}
}

// Create creates a new coupon
func (r *couponRepository) Create(ctx context.Context, coupon *models.Coupon) error {
	return r.db.WithContext(ctx).Create(coupon).Error
}

// GetByID retrieves a coupon by ID
func (r *couponRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Coupon, error) {
	var coupon models.Coupon
	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&coupon).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &coupon, nil
}

// GetByCode retrieves a coupon by code (case-insensitive)
func (r *couponRepository) GetByCode(ctx context.Context, code string) (*models.Coupon, error) {
	var coupon models.Coupon
	err := r.db.WithContext(ctx).
		Where("code = ?", code).
		First(&coupon).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &coupon, nil
}

// GetActiveCoupons retrieves all currently active coupons
func (r *couponRepository) GetActiveCoupons(ctx context.Context) ([]models.Coupon, error) {
	var coupons []models.Coupon
	now := time.Now()

	err := r.db.WithContext(ctx).
		Where("is_active = ? AND valid_from <= ? AND (expires_at IS NULL OR expires_at > ?)",
			true, now, now).
		Order("created_at DESC").
		Find(&coupons).Error

	return coupons, err
}

// Update updates a coupon
func (r *couponRepository) Update(ctx context.Context, coupon *models.Coupon) error {
	return r.db.WithContext(ctx).Save(coupon).Error
}

// Delete deletes a coupon by ID
func (r *couponRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.Coupon{}, id).Error
}

// GetUserRedemptionCount gets the number of times a user has used a specific coupon
func (r *couponRepository) GetUserRedemptionCount(ctx context.Context, couponID, userID uuid.UUID) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.CouponRedemption{}).
		Where("coupon_id = ? AND user_id = ?", couponID, userID).
		Count(&count).Error

	return int(count), err
}

// CreateRedemption creates a new coupon redemption record
func (r *couponRepository) CreateRedemption(ctx context.Context, redemption *models.CouponRedemption) error {
	return r.db.WithContext(ctx).Create(redemption).Error
}

// GetRedemptionsByUser retrieves coupon redemptions for a user with pagination
func (r *couponRepository) GetRedemptionsByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.CouponRedemption, int64, error) {
	var redemptions []models.CouponRedemption
	var total int64

	err := r.db.WithContext(ctx).Model(&models.CouponRedemption{}).
		Where("user_id = ?", userID).
		Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.WithContext(ctx).
		Preload("Coupon").
		Where("user_id = ?", userID).
		Order("redeemed_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&redemptions).Error

	return redemptions, total, err
}

// GetRedemptionsByCoupon retrieves redemptions for a specific coupon with pagination
func (r *couponRepository) GetRedemptionsByCoupon(ctx context.Context, couponID uuid.UUID, limit, offset int) ([]models.CouponRedemption, int64, error) {
	var redemptions []models.CouponRedemption
	var total int64

	err := r.db.WithContext(ctx).Model(&models.CouponRedemption{}).
		Where("coupon_id = ?", couponID).
		Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.WithContext(ctx).
		Where("coupon_id = ?", couponID).
		Order("redeemed_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&redemptions).Error

	return redemptions, total, err
}

// UpdateUsageCount increments the redemption count for a coupon
func (r *couponRepository) UpdateUsageCount(ctx context.Context, couponID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&models.Coupon{}).
		Where("id = ?", couponID).
		UpdateColumn("redemption_count", gorm.Expr("redemption_count + ?", 1)).Error
}

// GetCouponStats retrieves aggregated coupon statistics
func (r *couponRepository) GetCouponStats(ctx context.Context, timeRange *TimeRange) (*CouponStats, error) {
	var stats CouponStats

	// Get total coupons
	err := r.db.WithContext(ctx).Model(&models.Coupon{}).Count(&stats.TotalCoupons).Error
	if err != nil {
		return nil, err
	}

	// Get active coupons
	now := time.Now()
	err = r.db.WithContext(ctx).Model(&models.Coupon{}).
		Where("is_active = ? AND valid_from <= ? AND (expires_at IS NULL OR expires_at > ?)",
			true, now, now).
		Count(&stats.ActiveCoupons).Error
	if err != nil {
		return nil, err
	}

	// Get total redemptions
	err = r.db.WithContext(ctx).Model(&models.CouponRedemption{}).Count(&stats.TotalRedemptions).Error
	if err != nil {
		return nil, err
	}

	// Get total discount amount
	var totalDiscount int64
	err = r.db.WithContext(ctx).Model(&models.CouponRedemption{}).
		Select("COALESCE(SUM(discount_amount), 0)").
		Scan(&totalDiscount).Error
	if err != nil {
		return nil, err
	}
	stats.TotalDiscountAmount = totalDiscount

	// Calculate average discount
	if stats.TotalRedemptions > 0 {
		stats.AverageDiscount = totalDiscount / stats.TotalRedemptions
	}

	return &stats, nil
}