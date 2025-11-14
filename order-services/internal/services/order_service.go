package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"order-services/internal/config"
	"order-services/internal/models"
	"order-services/internal/repositories"

	"github.com/google/uuid"
)

var (
	ErrOrderNotFound      = errors.New("order not found")
	ErrInvalidOrderStatus = errors.New("invalid order status")
	ErrOrderExpired       = errors.New("order has expired")
	ErrEmptyOrder         = errors.New("order must contain at least one item")
	ErrUnauthorizedOrder  = errors.New("unauthorized to access this order")
	ErrInvalidCourse      = errors.New("invalid course data")
	ErrPaymentRequired    = errors.New("payment required for this operation")
	ErrInvalidCoupon      = errors.New("invalid coupon")
)

// OrderService defines the business logic interface for order management
type OrderService interface {
	CreateOrder(ctx context.Context, req *CreateOrderRequest) (*models.Order, error)
	GetOrder(ctx context.Context, orderID, userID uuid.UUID) (*models.Order, error)
	ListUserOrders(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.Order, int64, error)
	CancelOrder(ctx context.Context, orderID, userID uuid.UUID, reason string) error
	UpdateOrderStatus(ctx context.Context, orderID uuid.UUID, status string, reason string) error
	ProcessExpiredOrders(ctx context.Context) error
	ValidateOrderAccess(ctx context.Context, orderID, userID uuid.UUID) error
}

// CreateOrderRequest represents the request to create a new order
type CreateOrderRequest struct {
	UserID        uuid.UUID              `json:"user_id" validate:"required"`
	Items         []OrderItemRequest     `json:"items" validate:"required,min=1"`
	CouponCode    *string                `json:"coupon_code,omitempty"`
	CustomerEmail string                 `json:"customer_email" validate:"required,email"`
	CustomerName  *string                `json:"customer_name,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// OrderItemRequest represents an item in the order
type OrderItemRequest struct {
	CourseID      uuid.UUID `json:"course_id" validate:"required"`
	Quantity      int       `json:"quantity" validate:"required,min=1"`
	PriceSnapshot int64     `json:"price_snapshot,omitempty"` // in cents, if not provided will fetch from course service
}

// OrderService implements the business logic for orders
type orderService struct {
	orderRepo     repositories.OrderRepository
	orderItemRepo repositories.OrderItemRepository
	couponRepo    repositories.CouponRepository
	courseRepo    repositories.CourseRepository
	outboxRepo    repositories.OutboxRepository
	config        *config.Config
}

// NewOrderService creates a new order service instance
func NewOrderService(
	orderRepo repositories.OrderRepository,
	orderItemRepo repositories.OrderItemRepository,
	couponRepo repositories.CouponRepository,
	courseRepo repositories.CourseRepository,
	outboxRepo repositories.OutboxRepository,
	config *config.Config,
) OrderService {
	return &orderService{
		orderRepo:     orderRepo,
		orderItemRepo: orderItemRepo,
		couponRepo:    couponRepo,
		courseRepo:    courseRepo,
		outboxRepo:    outboxRepo,
		config:        config,
	}
}

// CreateOrder creates a new order with validation and business logic
func (s *orderService) CreateOrder(ctx context.Context, req *CreateOrderRequest) (*models.Order, error) {
	// Validate request
	if err := s.validateCreateOrderRequest(ctx, req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Calculate total amount and validate items
	totalAmount, orderItems, err := s.calculateOrderTotal(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate order total: %w", err)
	}

	// Apply coupon discount if provided
	var coupon *models.Coupon
	var discountAmount int64 = 0
	if req.CouponCode != nil {
		coupon, discountAmount, err = s.validateAndApplyCoupon(ctx, *req.CouponCode, req.UserID, totalAmount)
		if err != nil {
			return nil, fmt.Errorf("coupon validation failed: %w", err)
		}
	}

	finalAmount := totalAmount - discountAmount
	if finalAmount < 0 {
		finalAmount = 0
	}

	// Set expiration time
	expiresAt := time.Now().Add(time.Duration(s.config.OrderExpiresIn) * time.Hour)
	expiresAtSql := sql.NullTime{
		Time:  expiresAt,
		Valid: true,
	}

	// Create order
	order := &models.Order{
		UserID:        req.UserID,
		TotalAmount:   finalAmount,
		Currency:      "USD",
		Status:        models.OrderStatusCreated,
		CustomerEmail: req.CustomerEmail,
		ExpiresAt:     expiresAtSql,
		Metadata:      req.Metadata,
	}

	if req.CustomerName != nil {
		order.CustomerName = *req.CustomerName
	}

	// Start transaction
	err = s.orderRepo.WithTx(ctx, func(ctx context.Context) error {
		// Create order
		if err := s.orderRepo.Create(ctx, order); err != nil {
			return fmt.Errorf("failed to create order: %w", err)
		}

		// Create order items
		for _, item := range orderItems {
			item.OrderID = order.ID
			if err := s.orderItemRepo.Create(ctx, &item); err != nil {
				return fmt.Errorf("failed to create order item: %w", err)
			}
		}

		// Create coupon redemption if coupon was used
		if coupon != nil {
			redemption := &models.CouponRedemption{
				CouponID:       coupon.ID,
				UserID:         req.UserID,
				OrderID:        order.ID,
				DiscountAmount: discountAmount,
			}
			if err := s.couponRepo.CreateRedemption(ctx, redemption); err != nil {
				return fmt.Errorf("failed to create coupon redemption: %w", err)
			}
		}

		// Create order.created event
		event := &models.Outbox{
			AggregateID: order.ID,
			Topic:       "order.events",
			Type:        "order.created",
			Payload:     s.createOrderEventPayload(order, orderItems, coupon),
		}
		if err := s.outboxRepo.Create(ctx, event); err != nil {
			return fmt.Errorf("failed to create outbox event: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Load relationships before returning
	order, err = s.orderRepo.GetByID(ctx, order.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve created order: %w", err)
	}

	return order, nil
}

// GetOrder retrieves an order by ID with authorization check
func (s *orderService) GetOrder(ctx context.Context, orderID, userID uuid.UUID) (*models.Order, error) {
	if err := s.ValidateOrderAccess(ctx, orderID, userID); err != nil {
		return nil, err
	}

	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	if order == nil {
		return nil, ErrOrderNotFound
	}

	return order, nil
}

// ListUserOrders retrieves paginated orders for a user
func (s *orderService) ListUserOrders(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.Order, int64, error) {
	if limit <= 0 || limit > 100 {
		limit = 20 // default limit
	}
	if offset < 0 {
		offset = 0
	}

	orders, total, err := s.orderRepo.GetByUserID(ctx, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list user orders: %w", err)
	}

	return orders, total, nil
}

// CancelOrder cancels an order with validation
func (s *orderService) CancelOrder(ctx context.Context, orderID, userID uuid.UUID, reason string) error {
	// Validate access and get current order
	order, err := s.GetOrder(ctx, orderID, userID)
	if err != nil {
		return err
	}

	// Check if order can be cancelled
	if !s.canCancelOrder(order) {
		return fmt.Errorf("%w: order status is %s", ErrInvalidOrderStatus, order.Status)
	}

	// Start transaction for cancellation
	return s.orderRepo.WithTx(ctx, func(ctx context.Context) error {
		// Mark order as cancelled
		if err := s.orderRepo.MarkAsCancelled(ctx, orderID, reason); err != nil {
			return fmt.Errorf("failed to mark order as cancelled: %w", err)
		}

		// Create order.cancelled event
		event := &models.Outbox{
			AggregateID: orderID,
			Topic:       "order.events",
			Type:        "order.cancelled",
			Payload:     s.createOrderCancelledEventPayload(order, reason),
		}
		if err := s.outboxRepo.Create(ctx, event); err != nil {
			return fmt.Errorf("failed to create cancellation event: %w", err)
		}

		return nil
	})
}

// UpdateOrderStatus updates the status of an order (admin/system operation)
func (s *orderService) UpdateOrderStatus(ctx context.Context, orderID uuid.UUID, status string, reason string) error {
	// Validate status transition
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to get order: %w", err)
	}

	if order == nil {
		return ErrOrderNotFound
	}

	if !s.isValidStatusTransition(order.Status, status) {
		return fmt.Errorf("%w: cannot transition from %s to %s", ErrInvalidOrderStatus, order.Status, status)
	}

	// Update status with appropriate timestamp
	var timestamp *sql.NullTime
	now := time.Now()
	switch status {
	case models.OrderStatusPaid:
		timestamp = &sql.NullTime{Time: now, Valid: true}
	case models.OrderStatusFailed:
		timestamp = &sql.NullTime{Time: now, Valid: true}
	case models.OrderStatusCancelled:
		timestamp = &sql.NullTime{Time: now, Valid: true}
	case models.OrderStatusRefunded:
		timestamp = &sql.NullTime{Time: now, Valid: true}
	}

	if err := s.orderRepo.UpdateStatus(ctx, orderID, status, timestamp, reason); err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	// Create status change event
	eventType := fmt.Sprintf("order.%s", status)
	event := &models.Outbox{
		AggregateID: orderID,
		Topic:       "order.events",
		Type:        eventType,
		Payload:     s.createOrderStatusEventPayload(order, status, reason),
	}

	return s.outboxRepo.Create(ctx, event)
}

// ProcessExpiredOrders finds and processes orders that have expired
func (s *orderService) ProcessExpiredOrders(ctx context.Context) error {
	expiredTime := time.Now().Add(-time.Hour) // Orders expired more than 1 hour ago
	expiredOrders, err := s.orderRepo.GetPendingOrders(ctx, expiredTime)
	if err != nil {
		return fmt.Errorf("failed to get expired orders: %w", err)
	}

	for _, order := range expiredOrders {
		if err := s.UpdateOrderStatus(ctx, order.ID, models.OrderStatusCancelled, "Order expired"); err != nil {
			// Log error but continue processing other orders
			// TODO: Add proper logging
			continue
		}
	}

	return nil
}

// ValidateOrderAccess checks if a user has access to an order
func (s *orderService) ValidateOrderAccess(ctx context.Context, orderID, userID uuid.UUID) error {
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to get order for access validation: %w", err)
	}

	if order == nil {
		return ErrOrderNotFound
	}

	if order.UserID != userID {
		return ErrUnauthorizedOrder
	}

	return nil
}

// Helper methods

func (s *orderService) validateCreateOrderRequest(ctx context.Context, req *CreateOrderRequest) error {
	if len(req.Items) == 0 {
		return ErrEmptyOrder
	}

	// Validate each item
	for _, item := range req.Items {
		if item.Quantity <= 0 {
			return fmt.Errorf("%w: invalid quantity for item %s", ErrInvalidCourse, item.CourseID)
		}
	}

	return nil
}

func (s *orderService) calculateOrderTotal(ctx context.Context, req *CreateOrderRequest) (int64, []models.OrderItem, error) {
	var total int64
	var orderItems []models.OrderItem

	for _, itemReq := range req.Items {
		// Get course information
		course, err := s.courseRepo.GetByID(ctx, itemReq.CourseID)
		if err != nil {
			return 0, nil, fmt.Errorf("failed to get course %s: %w", itemReq.CourseID, err)
		}

		if course == nil || !course.IsActive {
			return 0, nil, fmt.Errorf("%w: course %s not found or inactive", ErrInvalidCourse, itemReq.CourseID)
		}

		price := course.Price
		if itemReq.PriceSnapshot > 0 {
			price = itemReq.PriceSnapshot
		}

		itemTotal := price * int64(itemReq.Quantity)
		total += itemTotal

		orderItem := models.OrderItem{
			CourseID:          itemReq.CourseID,
			CourseTitle:       course.Title,
			CourseDescription: course.Description,
			InstructorID:      course.InstructorID,
			InstructorName:    course.InstructorName,
			PriceSnapshot:     price,
			OriginalPrice:     course.Price,
			Quantity:          itemReq.Quantity,
			ItemType:          models.OrderItemTypeCourse,
		}

		orderItems = append(orderItems, orderItem)
	}

	return total, orderItems, nil
}

func (s *orderService) validateAndApplyCoupon(ctx context.Context, code string, userID uuid.UUID, orderAmount int64) (*models.Coupon, int64, error) {
	coupon, err := s.couponRepo.GetByCode(ctx, code)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get coupon: %w", err)
	}

	if coupon == nil {
		return nil, 0, fmt.Errorf("%w: coupon not found", ErrInvalidCoupon)
	}

	// Validate coupon availability
	if err := s.validateCouponAvailability(coupon); err != nil {
		return nil, 0, err
	}

	// Validate user restrictions
	if err := s.validateCouponUserRestrictions(ctx, coupon, userID, orderAmount); err != nil {
		return nil, 0, err
	}

	// Calculate discount
	discount := s.calculateDiscount(coupon, orderAmount)

	return coupon, discount, nil
}

// validateCouponAvailability validates basic coupon availability (active, dates, usage limits)
func (s *orderService) validateCouponAvailability(coupon *models.Coupon) error {
	// Check if coupon is active
	if !coupon.IsActive {
		return fmt.Errorf("%w: coupon is not active", ErrInvalidCoupon)
	}

	// Check if coupon has started
	now := time.Now()
	if coupon.ValidFrom.After(now) {
		return fmt.Errorf("%w: coupon has not started yet", ErrInvalidCoupon)
	}

	// Check if coupon has expired
	if coupon.ExpiresAt.Valid && coupon.ExpiresAt.Time.Before(now) {
		return fmt.Errorf("%w: coupon has expired", ErrInvalidCoupon)
	}

	// Check global usage limit
	if coupon.MaxRedemptions != nil && coupon.RedemptionCount >= *coupon.MaxRedemptions {
		return fmt.Errorf("%w: coupon usage limit exceeded", ErrInvalidCoupon)
	}

	return nil
}

// validateCouponUserRestrictions validates user-specific coupon restrictions
func (s *orderService) validateCouponUserRestrictions(ctx context.Context, coupon *models.Coupon, userID uuid.UUID, orderAmount int64) error {
	// Check minimum order amount
	if coupon.MinimumAmount != nil && orderAmount < *coupon.MinimumAmount {
		return fmt.Errorf("%w: minimum order amount not met", ErrInvalidCoupon)
	}

	// Check per-user usage limit
	if coupon.PerUserLimit != nil {
		usage, err := s.couponRepo.GetUserRedemptionCount(ctx, coupon.ID, userID)
		if err != nil {
			return fmt.Errorf("failed to check user coupon usage: %w", err)
		}
		if usage >= *coupon.PerUserLimit {
			return fmt.Errorf("%w: user coupon usage limit exceeded", ErrInvalidCoupon)
		}
	}

	return nil
}

func (s *orderService) calculateDiscount(coupon *models.Coupon, orderAmount int64) int64 {
	var discount int64

	switch coupon.Type {
	case models.CouponTypePercentage:
		if coupon.PercentOff != nil {
			discount = (orderAmount * int64(*coupon.PercentOff)) / 100
		}
	case models.CouponTypeFixedAmount:
		if coupon.AmountOff != nil {
			discount = *coupon.AmountOff
		}
	}

	// Ensure discount doesn't exceed order amount
	if discount > orderAmount {
		discount = orderAmount
	}

	return discount
}

func (s *orderService) canCancelOrder(order *models.Order) bool {
	return order.Status == models.OrderStatusCreated || order.Status == models.OrderStatusPendingPayment
}

func (s *orderService) isValidStatusTransition(currentStatus, newStatus string) bool {
	validTransitions := map[string][]string{
		models.OrderStatusCreated:        {models.OrderStatusPendingPayment, models.OrderStatusCancelled},
		models.OrderStatusPendingPayment: {models.OrderStatusPaid, models.OrderStatusFailed, models.OrderStatusCancelled},
		models.OrderStatusPaid:           {models.OrderStatusRefunded},
		models.OrderStatusFailed:         {models.OrderStatusCancelled},
		models.OrderStatusCancelled:      {},
		models.OrderStatusRefunded:       {},
	}

	allowedStatuses, exists := validTransitions[currentStatus]
	if !exists {
		return false
	}

	for _, status := range allowedStatuses {
		if status == newStatus {
			return true
		}
	}

	return false
}

// Event payload creation methods
func (s *orderService) createOrderEventPayload(order *models.Order, items []models.OrderItem, coupon *models.Coupon) []byte {
	payload := map[string]interface{}{
		"order_id":     order.ID,
		"user_id":      order.UserID,
		"total_amount": order.TotalAmount,
		"currency":     order.Currency,
		"status":       order.Status,
		"items":        items,
		"created_at":   order.CreatedAt,
		"expires_at":   order.ExpiresAt,
	}

	if coupon != nil {
		payload["coupon"] = map[string]interface{}{
			"id":   coupon.ID,
			"code": coupon.Code,
		}
	}

	// Convert to JSON (simplified for example)
	// In production, use proper JSON marshaling
	return s.marshalPayload(payload)
}

func (s *orderService) createOrderCancelledEventPayload(order *models.Order, reason string) []byte {
	payload := map[string]interface{}{
		"order_id":     order.ID,
		"user_id":      order.UserID,
		"status":       order.Status,
		"reason":       reason,
		"cancelled_at": time.Now(),
	}

	// Convert to JSON (simplified for example)
	return s.marshalPayload(payload)
}

func (s *orderService) createOrderStatusEventPayload(order *models.Order, status, reason string) []byte {
	payload := map[string]interface{}{
		"order_id":        order.ID,
		"user_id":         order.UserID,
		"previous_status": order.Status,
		"new_status":      status,
		"reason":          reason,
		"updated_at":      time.Now(),
	}

	// Convert to JSON (simplified for example)
	return s.marshalPayload(payload)
}

func (s *orderService) marshalPayload(payload map[string]interface{}) []byte {
	data, err := json.Marshal(payload)
	if err != nil {
		return []byte(`{"error": "failed to marshal payload"}`)
	}
	return data
}
