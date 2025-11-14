package repositories

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"gorm.io/gorm"

	"order-services/internal/models"

	"github.com/google/uuid"
)

// OrderRepository interface for order data access
type OrderRepository interface {
	Create(ctx context.Context, order *models.Order) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Order, error)
	GetByUserID(ctx context.Context, userID uuid.UUID, limit int, offset int) ([]models.Order, int64, error)
	GetByPaymentIntentID(ctx context.Context, paymentIntentID string) (*models.Order, error)
	GetPendingOrders(ctx context.Context, olderThan time.Time) ([]models.Order, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string, timestamp *sql.NullTime, reason string) error
	UpdatePaymentIntentID(ctx context.Context, id uuid.UUID, paymentIntentID string) error
	MarkAsPaid(ctx context.Context, id uuid.UUID, paymentIntentID string) error
	MarkAsFailed(ctx context.Context, id uuid.UUID, reason string) error
	MarkAsCancelled(ctx context.Context, id uuid.UUID, reason string) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetOrderStats(ctx context.Context, userID *uuid.UUID, timeRange *TimeRange) (*OrderStats, error)
	WithTx(ctx context.Context, fn func(ctx context.Context) error) error
	UserHasPreviousOrders(ctx context.Context, userID uuid.UUID) (bool, error)
}

// TimeRange represents a time period for queries
type TimeRange struct {
	From time.Time
	To   time.Time
}

// OrderStats represents aggregated order statistics
type OrderStats struct {
	TotalOrders       int64 `json:"total_orders"`
	TotalRevenue      int64 `json:"total_revenue"`
	PendingOrders     int64 `json:"pending_orders"`
	CompletedOrders   int64 `json:"completed_orders"`
	CancelledOrders   int64 `json:"cancelled_orders"`
	FailedOrders      int64 `json:"failed_orders"`
	RefundedOrders    int64 `json:"refunded_orders"`
	AverageOrderValue int64 `json:"average_order_value"`
}

// orderRepository implements OrderRepository
type orderRepository struct {
	db *gorm.DB
}

// NewOrderRepository creates a new order repository
func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &orderRepository{db: db}
}

// getDB gets the database connection from context if available (for transactions),
// otherwise returns the default repository DB connection
func (r *orderRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := ctx.Value("tx").(*gorm.DB); ok {
		return tx
	}
	return r.db
}

// Create creates a new order
func (r *orderRepository) Create(ctx context.Context, order *models.Order) error {
	return r.getDB(ctx).WithContext(ctx).Create(order).Error
}

// GetByID retrieves an order by ID
func (r *orderRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Order, error) {
	var order models.Order
	err := r.getDB(ctx).WithContext(ctx).
		Preload("OrderItems").
		Preload("Payments").
		Preload("CouponRedemptions").
		Where("id = ?", id).
		First(&order).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &order, nil
}

// GetByUserID retrieves orders for a user with pagination
func (r *orderRepository) GetByUserID(ctx context.Context, userID uuid.UUID, limit int, offset int) ([]models.Order, int64, error) {
	var orders []models.Order
	var total int64

	err := r.getDB(ctx).WithContext(ctx).Model(&models.Order{}).
		Where("user_id = ?", userID).
		Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.getDB(ctx).WithContext(ctx).
		Preload("OrderItems").
		Preload("Payments").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&orders).Error

	return orders, total, err
}

// GetByPaymentIntentID retrieves an order by Stripe payment intent ID
func (r *orderRepository) GetByPaymentIntentID(ctx context.Context, paymentIntentID string) (*models.Order, error) {
	var order models.Order
	err := r.getDB(ctx).WithContext(ctx).
		Preload("OrderItems").
		Preload("Payments").
		Where("payment_intent_id = ?", paymentIntentID).
		First(&order).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &order, nil
}

// GetPendingOrders retrieves orders that are stuck in pending_payment status
func (r *orderRepository) GetPendingOrders(ctx context.Context, olderThan time.Time) ([]models.Order, error) {
	var orders []models.Order
	err := r.getDB(ctx).WithContext(ctx).
		Where("status = ? AND created_at < ?", models.OrderStatusPendingPayment, olderThan).
		Find(&orders).Error
	return orders, err
}

// UpdateStatus updates the order status and optionally a timestamp
func (r *orderRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string, timestamp *sql.NullTime, reason string) error {
	updates := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}

	// Add timestamp based on status
	if timestamp != nil {
		switch status {
		case models.OrderStatusPaid:
			updates["paid_at"] = timestamp
		case models.OrderStatusCancelled:
			updates["cancelled_at"] = timestamp
		case models.OrderStatusFailed:
			updates["failed_at"] = timestamp
		case models.OrderStatusRefunded:
			updates["refunded_at"] = timestamp
		}
	}

	// Add failure reason if provided
	if reason != "" {
		updates["failure_reason"] = reason
	}

	return r.getDB(ctx).WithContext(ctx).Model(&models.Order{}).
		Where("id = ?", id).
		Updates(updates).Error
}

// UpdatePaymentIntentID updates the Stripe payment intent ID for an order
func (r *orderRepository) UpdatePaymentIntentID(ctx context.Context, id uuid.UUID, paymentIntentID string) error {
	return r.getDB(ctx).WithContext(ctx).Model(&models.Order{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"payment_intent_id": paymentIntentID,
			"status":            models.OrderStatusPendingPayment,
			"updated_at":        time.Now(),
		}).Error
}

// MarkAsPaid marks an order as paid
func (r *orderRepository) MarkAsPaid(ctx context.Context, id uuid.UUID, paymentIntentID string) error {
	return r.getDB(ctx).WithContext(ctx).Model(&models.Order{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":            models.OrderStatusPaid,
			"payment_intent_id": paymentIntentID,
			"paid_at":           time.Now(),
			"updated_at":        time.Now(),
		}).Error
}

// MarkAsFailed marks an order as failed
func (r *orderRepository) MarkAsFailed(ctx context.Context, id uuid.UUID, reason string) error {
	return r.getDB(ctx).WithContext(ctx).Model(&models.Order{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":         models.OrderStatusFailed,
			"failure_reason": reason,
			"failed_at":      time.Now(),
			"updated_at":     time.Now(),
		}).Error
}

// MarkAsCancelled marks an order as cancelled
func (r *orderRepository) MarkAsCancelled(ctx context.Context, id uuid.UUID, reason string) error {
	return r.getDB(ctx).WithContext(ctx).Model(&models.Order{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":         models.OrderStatusCancelled,
			"failure_reason": reason,
			"cancelled_at":   time.Now(),
			"updated_at":     time.Now(),
		}).Error
}

// Delete deletes an order by ID
func (r *orderRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.getDB(ctx).WithContext(ctx).Delete(&models.Order{}, id).Error
}

// GetOrderStats retrieves aggregated order statistics
func (r *orderRepository) GetOrderStats(ctx context.Context, userID *uuid.UUID, timeRange *TimeRange) (*OrderStats, error) {
	var result struct {
		TotalOrders     sql.NullInt64 `json:"total_orders"`
		TotalRevenue    sql.NullInt64 `json:"total_revenue"`
		PendingOrders   sql.NullInt64 `json:"pending_orders"`
		CompletedOrders sql.NullInt64 `json:"completed_orders"`
		CancelledOrders sql.NullInt64 `json:"cancelled_orders"`
		FailedOrders    sql.NullInt64 `json:"failed_orders"`
		RefundedOrders  sql.NullInt64 `json:"refunded_orders"`
	}

	// Build query with optional filters
	query := r.getDB(ctx).WithContext(ctx).Model(&models.Order{})

	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}

	if timeRange != nil {
		query = query.Where("created_at >= ? AND created_at <= ?", timeRange.From, timeRange.To)
	}

	// Get total orders
	var totalOrders int64
	if err := query.Count(&totalOrders).Error; err != nil {
		return nil, err
	}
	result.TotalOrders = sql.NullInt64{Int64: totalOrders, Valid: true}

	// Get total revenue from paid orders
	var revenue int64
	if err := query.Where("status = ?", models.OrderStatusPaid).
		Select("COALESCE(SUM(total_amount), 0)").
		Scan(&revenue).Error; err != nil {
		return nil, err
	}
	result.TotalRevenue = sql.NullInt64{Int64: revenue, Valid: true}

	// Get status counts
	statusCounts := map[string]int64{
		models.OrderStatusPendingPayment: 0,
		models.OrderStatusPaid:           0,
		models.OrderStatusCancelled:      0,
		models.OrderStatusFailed:         0,
		models.OrderStatusRefunded:       0,
	}

	for status := range statusCounts {
		var count int64
		if err := query.Where("status = ?", status).Count(&count).Error; err != nil {
			return nil, err
		}
		statusCounts[status] = count
	}

	result.PendingOrders = sql.NullInt64{Int64: statusCounts[models.OrderStatusPendingPayment], Valid: true}
	result.CompletedOrders = sql.NullInt64{Int64: statusCounts[models.OrderStatusPaid], Valid: true}
	result.CancelledOrders = sql.NullInt64{Int64: statusCounts[models.OrderStatusCancelled], Valid: true}
	result.FailedOrders = sql.NullInt64{Int64: statusCounts[models.OrderStatusFailed], Valid: true}
	result.RefundedOrders = sql.NullInt64{Int64: statusCounts[models.OrderStatusRefunded], Valid: true}

	stats := &OrderStats{
		TotalOrders:     result.TotalOrders.Int64,
		TotalRevenue:    result.TotalRevenue.Int64,
		PendingOrders:   result.PendingOrders.Int64,
		CompletedOrders: result.CompletedOrders.Int64,
		CancelledOrders: result.CancelledOrders.Int64,
		FailedOrders:    result.FailedOrders.Int64,
		RefundedOrders:  result.RefundedOrders.Int64,
	}

	// Calculate average order value
	if stats.CompletedOrders > 0 {
		stats.AverageOrderValue = stats.TotalRevenue / stats.CompletedOrders
	}

	return stats, nil
}

// WithTx executes a function within a database transaction
func (r *orderRepository) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txCtx := context.WithValue(ctx, "tx", tx)
		return fn(txCtx)
	})
}

// UserHasPreviousOrders checks if the user has at least one successfully completed order
func (r *orderRepository) UserHasPreviousOrders(ctx context.Context, userID uuid.UUID) (bool, error) {
	var count int64

	err := r.getDB(ctx).WithContext(ctx).Model(&models.Order{}).
		Where("user_id = ? AND status IN ?", userID, []string{models.OrderStatusPaid, models.OrderStatusRefunded}).
		Count(&count).Error
	if err != nil {
		return false, err
	}

	return count > 0, nil
}
