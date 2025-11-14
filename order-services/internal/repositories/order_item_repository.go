package repositories

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/google/uuid"
	"order-services/internal/models"
)

// OrderItemRepository interface for order item data access
type OrderItemRepository interface {
	Create(ctx context.Context, item *models.OrderItem) error
	CreateBatch(ctx context.Context, items []models.OrderItem) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.OrderItem, error)
	GetByOrderID(ctx context.Context, orderID uuid.UUID) ([]models.OrderItem, error)
	GetByCourseID(ctx context.Context, courseID uuid.UUID, limit int, offset int) ([]models.OrderItem, int64, error)
	Update(ctx context.Context, item *models.OrderItem) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByOrderID(ctx context.Context, orderID uuid.UUID) error
	GetCourseSalesStats(ctx context.Context, courseID uuid.UUID) (*CourseSalesStats, error)
}

// CourseSalesStats represents sales statistics for a course
type CourseSalesStats struct {
	TotalSales     int64 `json:"total_sales"`
	TotalRevenue   int64 `json:"total_revenue"`
	TotalStudents  int64 `json:"total_students"`
	AveragePrice   int64 `json:"average_price"`
	RecentSales    int64 `json:"recent_sales"` // sales in last 30 days
}

// orderItemRepository implements OrderItemRepository
type orderItemRepository struct {
	db *gorm.DB
}

// NewOrderItemRepository creates a new order item repository
func NewOrderItemRepository(db *gorm.DB) OrderItemRepository {
	return &orderItemRepository{db: db}
}

// Create creates a new order item
func (r *orderItemRepository) Create(ctx context.Context, item *models.OrderItem) error {
	return r.db.WithContext(ctx).Create(item).Error
}

// CreateBatch creates multiple order items in a single transaction
func (r *orderItemRepository) CreateBatch(ctx context.Context, items []models.OrderItem) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, item := range items {
			if err := tx.Create(&item).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// GetByID retrieves an order item by ID
func (r *orderItemRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.OrderItem, error) {
	var item models.OrderItem
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&item).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

// GetByOrderID retrieves all order items for an order
func (r *orderItemRepository) GetByOrderID(ctx context.Context, orderID uuid.UUID) ([]models.OrderItem, error) {
	var items []models.OrderItem
	err := r.db.WithContext(ctx).
		Where("order_id = ?", orderID).
		Order("created_at").
		Find(&items).Error
	return items, err
}

// GetByCourseID retrieves order items for a course with pagination
func (r *orderItemRepository) GetByCourseID(ctx context.Context, courseID uuid.UUID, limit int, offset int) ([]models.OrderItem, int64, error) {
	var items []models.OrderItem
	var total int64

	err := r.db.WithContext(ctx).Model(&models.OrderItem{}).
		Where("course_id = ?", courseID).
		Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.WithContext(ctx).
		Preload("Order"). // Include order information
		Where("course_id = ?", courseID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&items).Error

	return items, total, err
}

// Update updates an order item
func (r *orderItemRepository) Update(ctx context.Context, item *models.OrderItem) error {
	return r.db.WithContext(ctx).Save(item).Error
}

// Delete deletes an order item by ID
func (r *orderItemRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.OrderItem{}, id).Error
}

// DeleteByOrderID deletes all order items for an order
func (r *orderItemRepository) DeleteByOrderID(ctx context.Context, orderID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("order_id = ?", orderID).Delete(&models.OrderItem{}).Error
}

// GetCourseSalesStats retrieves sales statistics for a specific course
func (r *orderItemRepository) GetCourseSalesStats(ctx context.Context, courseID uuid.UUID) (*CourseSalesStats, error) {
	var stats struct {
		TotalSales    int64 `json:"total_sales"`
		TotalRevenue  int64 `json:"total_revenue"`
		TotalStudents int64 `json:"total_students"`
		RecentSales   int64 `json:"recent_sales"`
	}

	// Get total sales and revenue from paid orders only
	query := r.db.WithContext(ctx).Table("order_items oi").
		Select("COUNT(*) as total_sales, COALESCE(SUM(oi.price_snapshot), 0) as total revenue").
		Joins("INNER JOIN orders o ON oi.order_id = o.id").
		Where("oi.course_id = ? AND o.status = ?", courseID, models.OrderStatusPaid)

	if err := query.Row().Scan(&stats.TotalSales, &stats.TotalRevenue); err != nil {
		return nil, err
	}

	// Get unique student count (users who actually purchased the course)
	var studentCount int64
	if err := r.db.WithContext(ctx).Table("order_items oi").
		Select("COUNT(DISTINCT o.user_id)").
		Joins("INNER JOIN orders o ON oi.order_id = o.id").
		Where("oi.course_id = ? AND o.status = ?", courseID, models.OrderStatusPaid).
		Scan(&studentCount).Error; err != nil {
		return nil, err
	}
	stats.TotalStudents = studentCount

	// Get recent sales (last 30 days)
	if err := r.db.WithContext(ctx).Table("order_items oi").
		Count(&stats.RecentSales).
		Joins("INNER JOIN orders o ON oi.order_id = o.id").
		Where("oi.course_id = ? AND o.status = ? AND o.created_at >= NOW() - INTERVAL '30 days'",
			courseID, models.OrderStatusPaid).Error; err != nil {
		return nil, err
	}

	// Calculate average price
	averagePrice := int64(0)
	if stats.TotalSales > 0 {
		averagePrice = stats.TotalRevenue / stats.TotalSales
	}

	result := &CourseSalesStats{
		TotalSales:    stats.TotalSales,
		TotalRevenue:  stats.TotalRevenue,
		TotalStudents: stats.TotalStudents,
		AveragePrice:  averagePrice,
		RecentSales:   stats.RecentSales,
	}

	return result, nil
}