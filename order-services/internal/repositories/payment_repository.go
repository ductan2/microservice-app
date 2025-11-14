package repositories

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/google/uuid"
	"order-services/internal/models"
)

// PaymentRepository interface for payment data access
type PaymentRepository interface {
	Create(ctx context.Context, payment *models.Payment) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Payment, error)
	GetByOrderID(ctx context.Context, orderID uuid.UUID) (*models.Payment, error)
	GetByStripePaymentIntentID(ctx context.Context, paymentIntentID string) (*models.Payment, error)
	Update(ctx context.Context, payment *models.Payment) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
	UpdateStatusByStripePaymentIntentID(ctx context.Context, paymentIntentID, status string) error
	MarkAsFailed(ctx context.Context, id uuid.UUID, failureMessage, failureCode string) error
	GetPaymentsByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.Payment, int64, error)
	GetPaymentStats(ctx context.Context, userID *uuid.UUID, timeRange *TimeRange) (*PaymentStats, error)
	GetFailedPayments(ctx context.Context, olderThan time.Time) ([]models.Payment, error)
}

// PaymentStats represents aggregated payment statistics
type PaymentStats struct {
	TotalPayments    int64 `json:"total_payments"`
	SuccessfulAmount int64 `json:"successful_amount"` // in cents
	FailedAmount     int64 `json:"failed_amount"`     // in cents
	PendingAmount    int64 `json:"pending_amount"`    // in cents
	SuccessRate      float64 `json:"success_rate"`     // percentage
	AverageAmount    int64 `json:"average_amount"`    // in cents
}

// paymentRepository implements PaymentRepository
type paymentRepository struct {
	db *gorm.DB
}

// NewPaymentRepository creates a new payment repository
func NewPaymentRepository(db *gorm.DB) PaymentRepository {
	return &paymentRepository{db: db}
}

// Create creates a new payment record
func (r *paymentRepository) Create(ctx context.Context, payment *models.Payment) error {
	return r.db.WithContext(ctx).Create(payment).Error
}

// GetByID retrieves a payment by ID
func (r *paymentRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Payment, error) {
	var payment models.Payment
	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&payment).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &payment, nil
}

// GetByOrderID retrieves the most recent payment for an order
func (r *paymentRepository) GetByOrderID(ctx context.Context, orderID uuid.UUID) (*models.Payment, error) {
	var payment models.Payment
	err := r.db.WithContext(ctx).
		Where("order_id = ?", orderID).
		Order("created_at DESC").
		First(&payment).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &payment, nil
}

// GetByStripePaymentIntentID retrieves a payment by Stripe payment intent ID
func (r *paymentRepository) GetByStripePaymentIntentID(ctx context.Context, paymentIntentID string) (*models.Payment, error) {
	var payment models.Payment
	err := r.db.WithContext(ctx).
		Where("stripe_payment_intent_id = ?", paymentIntentID).
		First(&payment).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &payment, nil
}

// Update updates a payment record
func (r *paymentRepository) Update(ctx context.Context, payment *models.Payment) error {
	return r.db.WithContext(ctx).Save(payment).Error
}

// UpdateStatus updates the status of a payment
func (r *paymentRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	updates := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}

	// Set processed_at for terminal states
	if status == models.PaymentStatusSucceeded || status == models.PaymentStatusFailed || status == models.PaymentStatusCanceled {
		updates["processed_at"] = time.Now()
	}

	return r.db.WithContext(ctx).Model(&models.Payment{}).
		Where("id = ?", id).
		Updates(updates).Error
}

// UpdateStatusByStripePaymentIntentID updates payment status by Stripe payment intent ID
func (r *paymentRepository) UpdateStatusByStripePaymentIntentID(ctx context.Context, paymentIntentID, status string) error {
	updates := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}

	// Set processed_at for terminal states
	if status == models.PaymentStatusSucceeded || status == models.PaymentStatusFailed || status == models.PaymentStatusCanceled {
		updates["processed_at"] = time.Now()
	}

	return r.db.WithContext(ctx).Model(&models.Payment{}).
		Where("stripe_payment_intent_id = ?", paymentIntentID).
		Updates(updates).Error
}

// MarkAsFailed marks a payment as failed with detailed error information
func (r *paymentRepository) MarkAsFailed(ctx context.Context, id uuid.UUID, failureMessage, failureCode string) error {
	updates := map[string]interface{}{
		"status":          models.PaymentStatusFailed,
		"failure_message": failureMessage,
		"failure_code":    failureCode,
		"processed_at":    time.Now(),
		"updated_at":      time.Now(),
	}

	return r.db.WithContext(ctx).Model(&models.Payment{}).
		Where("id = ?", id).
		Updates(updates).Error
}

// GetPaymentsByUserID retrieves payments for a user with pagination
func (r *paymentRepository) GetPaymentsByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.Payment, int64, error) {
	var payments []models.Payment
	var total int64

	// Count total payments for user
	err := r.db.WithContext(ctx).Model(&models.Payment{}).
		Joins("JOIN orders ON payments.order_id = orders.id").
		Where("orders.user_id = ?", userID).
		Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// Get paginated payments
	err = r.db.WithContext(ctx).
		Joins("JOIN orders ON payments.order_id = orders.id").
		Preload("Order").
		Where("orders.user_id = ?", userID).
		Order("payments.created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&payments).Error

	return payments, total, err
}

// GetPaymentStats retrieves aggregated payment statistics
func (r *paymentRepository) GetPaymentStats(ctx context.Context, userID *uuid.UUID, timeRange *TimeRange) (*PaymentStats, error) {
	var stats PaymentStats

	// Build base query
	baseQuery := r.db.WithContext(ctx).Model(&models.Payment{}).
		Joins("JOIN orders ON payments.order_id = orders.id")

	if userID != nil {
		baseQuery = baseQuery.Where("orders.user_id = ?", *userID)
	}

	if timeRange != nil {
		baseQuery = baseQuery.Where("payments.created_at >= ? AND payments.created_at <= ?", timeRange.From, timeRange.To)
	}

	// Get total payments
	err := baseQuery.Count(&stats.TotalPayments).Error
	if err != nil {
		return nil, err
	}

	// Get successful payments amount
	var successfulAmount int64
	err = baseQuery.Where("payments.status = ?", models.PaymentStatusSucceeded).
		Select("COALESCE(SUM(payments.amount), 0)").
		Scan(&successfulAmount).Error
	if err != nil {
		return nil, err
	}
	stats.SuccessfulAmount = successfulAmount

	// Get failed payments amount
	var failedAmount int64
	err = baseQuery.Where("payments.status = ?", models.PaymentStatusFailed).
		Select("COALESCE(SUM(payments.amount), 0)").
		Scan(&failedAmount).Error
	if err != nil {
		return nil, err
	}
	stats.FailedAmount = failedAmount

	// Get pending payments amount
	var pendingAmount int64
	err = baseQuery.Where("payments.status IN ?", []string{
		models.PaymentStatusRequiresPaymentMethod,
		models.PaymentStatusRequiresConfirmation,
		models.PaymentStatusRequiresAction,
		models.PaymentStatusProcessing,
	}).
		Select("COALESCE(SUM(payments.amount), 0)").
		Scan(&pendingAmount).Error
	if err != nil {
		return nil, err
	}
	stats.PendingAmount = pendingAmount

	// Calculate success rate and average amount
	if stats.TotalPayments > 0 {
		var successfulCount int64
		err = baseQuery.Where("payments.status = ?", models.PaymentStatusSucceeded).
			Count(&successfulCount).Error
		if err != nil {
			return nil, err
		}
		stats.SuccessRate = float64(successfulCount) / float64(stats.TotalPayments) * 100

		// Average amount (excluding failed payments)
		var totalValidAmount int64
		var validCount int64
		err = baseQuery.Where("payments.status != ?", models.PaymentStatusFailed).
			Select("COALESCE(SUM(payments.amount), 0), COUNT(*)").
			Row().Scan(&totalValidAmount, &validCount)
		if err != nil {
			return nil, err
		}
		if validCount > 0 {
			stats.AverageAmount = totalValidAmount / validCount
		}
	}

	return &stats, nil
}

// GetFailedPayments retrieves payments that are stuck in processing state
func (r *paymentRepository) GetFailedPayments(ctx context.Context, olderThan time.Time) ([]models.Payment, error) {
	var payments []models.Payment
	err := r.db.WithContext(ctx).
		Where("status = ? AND created_at < ?", models.PaymentStatusProcessing, olderThan).
		Find(&payments).Error

	return payments, err
}

// WithTransaction executes a function within a database transaction
func (r *paymentRepository) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txCtx := context.WithValue(ctx, "tx", tx)
		return fn(txCtx)
	})
}
