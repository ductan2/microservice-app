package repositories

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"order-services/internal/models"
	"order-services/internal/types"
)

// RefundRepository interface for refund data access
type RefundRepository interface {
	Create(ctx context.Context, refund *models.RefundRequest) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.RefundRequest, error)
	GetByOrderID(ctx context.Context, orderID uuid.UUID) (*models.RefundRequest, error)
	GetByStripeRefundID(ctx context.Context, stripeRefundID string) (*models.RefundRequest, error)
	GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.RefundRequest, int64, error)
	GetAll(ctx context.Context, limit, offset int) ([]models.RefundRequest, int64, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status, adminReason string) error
	UpdateStripeRefundID(ctx context.Context, id uuid.UUID, stripeRefundID string) error
	GetStats(ctx context.Context, timeRange *types.TimeRange) (*RefundStats, error)
}

// RefundStats represents refund statistics
type RefundStats struct {
	TotalRefunds     int64 `json:"total_refunds"`
	TotalAmount      int64 `json:"total_amount"`      // in cents
	ProcessedRefunds int64 `json:"processed_refunds"`
	PendingRefunds   int64 `json:"pending_refunds"`
	RejectedRefunds  int64 `json:"rejected_refunds"`
	FailedRefunds    int64 `json:"failed_refunds"`
	AverageAmount    int64 `json:"average_amount"`    // in cents
}

// refundRepository implements RefundRepository
type refundRepository struct {
	db *sql.DB
}

// NewRefundRepository creates a new refund repository
func NewRefundRepository(db *sql.DB) RefundRepository {
	return &refundRepository{db: db}
}

// Create creates a new refund request
func (r *refundRepository) Create(ctx context.Context, refund *models.RefundRequest) error {
	query := `
		INSERT INTO refund_requests (order_id, user_id, amount, reason, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id`

	err := r.db.QueryRowContext(ctx, query,
		refund.OrderID,
		refund.UserID,
		refund.Amount,
		refund.Reason,
		refund.Status,
		time.Now(),
		time.Now(),
	).Scan(&refund.ID)

	return err
}

// GetByID retrieves a refund request by ID
func (r *refundRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.RefundRequest, error) {
	query := `
		SELECT id, order_id, user_id, amount, reason, status, admin_reason,
		       stripe_refund_id, created_at, updated_at, processed_at
		FROM refund_requests
		WHERE id = $1`

	var refund models.RefundRequest
	var adminReason, stripeRefundID sql.NullString
	var processedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&refund.ID,
		&refund.OrderID,
		&refund.UserID,
		&refund.Amount,
		&refund.Reason,
		&refund.Status,
		&adminReason,
		&stripeRefundID,
		&refund.CreatedAt,
		&refund.UpdatedAt,
		&processedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if adminReason.Valid {
		refund.AdminReason = &adminReason.String
	}

	if stripeRefundID.Valid {
		refund.StripeRefundID = &stripeRefundID.String
	}

	if processedAt.Valid {
		refund.ProcessedAt = processedAt
	}

	return &refund, nil
}

// GetByOrderID retrieves a refund request by order ID
func (r *refundRepository) GetByOrderID(ctx context.Context, orderID uuid.UUID) (*models.RefundRequest, error) {
	query := `
		SELECT id, order_id, user_id, amount, reason, status, admin_reason,
		       stripe_refund_id, created_at, updated_at, processed_at
		FROM refund_requests
		WHERE order_id = $1
		ORDER BY created_at DESC
		LIMIT 1`

	var refund models.RefundRequest
	var adminReason, stripeRefundID sql.NullString
	var processedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, orderID).Scan(
		&refund.ID,
		&refund.OrderID,
		&refund.UserID,
		&refund.Amount,
		&refund.Reason,
		&refund.Status,
		&adminReason,
		&stripeRefundID,
		&refund.CreatedAt,
		&refund.UpdatedAt,
		&processedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if adminReason.Valid {
		refund.AdminReason = &adminReason.String
	}

	if stripeRefundID.Valid {
		refund.StripeRefundID = &stripeRefundID.String
	}

	if processedAt.Valid {
		refund.ProcessedAt = processedAt
	}

	return &refund, nil
}

// GetByStripeRefundID retrieves a refund request by Stripe refund ID
func (r *refundRepository) GetByStripeRefundID(ctx context.Context, stripeRefundID string) (*models.RefundRequest, error) {
	query := `
		SELECT id, order_id, user_id, amount, reason, status, admin_reason,
		       stripe_refund_id, created_at, updated_at, processed_at
		FROM refund_requests
		WHERE stripe_refund_id = $1`

	var refund models.RefundRequest
	var adminReason sql.NullString
	var stripeRefundIDNull sql.NullString
	var processedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, stripeRefundID).Scan(
		&refund.ID,
		&refund.OrderID,
		&refund.UserID,
		&refund.Amount,
		&refund.Reason,
		&refund.Status,
		&adminReason,
		&stripeRefundIDNull,
		&refund.CreatedAt,
		&refund.UpdatedAt,
		&processedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if adminReason.Valid {
		refund.AdminReason = &adminReason.String
	}

	if stripeRefundIDNull.Valid {
		refund.StripeRefundID = &stripeRefundIDNull.String
	}

	if processedAt.Valid {
		refund.ProcessedAt = processedAt
	}

	return &refund, nil
}

// GetByUserID retrieves refund requests for a user with pagination
func (r *refundRepository) GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.RefundRequest, int64, error) {
	var refunds []models.RefundRequest
	var total int64

	// Count total refunds for user
	countQuery := `SELECT COUNT(*) FROM refund_requests WHERE user_id = $1`
	err := r.db.QueryRowContext(ctx, countQuery, userID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get paginated refunds
	query := `
		SELECT id, order_id, user_id, amount, reason, status, admin_reason,
		       stripe_refund_id, created_at, updated_at, processed_at
		FROM refund_requests
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var refund models.RefundRequest
		var adminReason, stripeRefundID sql.NullString
		var processedAt sql.NullTime

		err := rows.Scan(
			&refund.ID,
			&refund.OrderID,
			&refund.UserID,
			&refund.Amount,
			&refund.Reason,
			&refund.Status,
			&adminReason,
			&stripeRefundID,
			&refund.CreatedAt,
			&refund.UpdatedAt,
			&processedAt,
		)
		if err != nil {
			return nil, 0, err
		}

		if adminReason.Valid {
			refund.AdminReason = &adminReason.String
		}

		if stripeRefundID.Valid {
			refund.StripeRefundID = &stripeRefundID.String
		}

		if processedAt.Valid {
			refund.ProcessedAt = processedAt
		}

		refunds = append(refunds, refund)
	}

	return refunds, total, rows.Err()
}

// GetAll retrieves all refund requests with pagination
func (r *refundRepository) GetAll(ctx context.Context, limit, offset int) ([]models.RefundRequest, int64, error) {
	var refunds []models.RefundRequest
	var total int64

	// Count total refunds
	countQuery := `SELECT COUNT(*) FROM refund_requests`
	err := r.db.QueryRowContext(ctx, countQuery).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get paginated refunds
	query := `
		SELECT id, order_id, user_id, amount, reason, status, admin_reason,
		       stripe_refund_id, created_at, updated_at, processed_at
		FROM refund_requests
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var refund models.RefundRequest
		var adminReason, stripeRefundID sql.NullString
		var processedAt sql.NullTime

		err := rows.Scan(
			&refund.ID,
			&refund.OrderID,
			&refund.UserID,
			&refund.Amount,
			&refund.Reason,
			&refund.Status,
			&adminReason,
			&stripeRefundID,
			&refund.CreatedAt,
			&refund.UpdatedAt,
			&processedAt,
		)
		if err != nil {
			return nil, 0, err
		}

		if adminReason.Valid {
			refund.AdminReason = &adminReason.String
		}

		if stripeRefundID.Valid {
			refund.StripeRefundID = &stripeRefundID.String
		}

		if processedAt.Valid {
			refund.ProcessedAt = processedAt
		}

		refunds = append(refunds, refund)
	}

	return refunds, total, rows.Err()
}

// UpdateStatus updates the status of a refund request
func (r *refundRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status, adminReason string) error {
	query := `
		UPDATE refund_requests
		SET status = $1, admin_reason = $2, updated_at = $3
		WHERE id = $4`

	_, err := r.db.ExecContext(ctx, query, status, adminReason, time.Now(), id)
	return err
}

// UpdateStripeRefundID updates the Stripe refund ID for a refund request
func (r *refundRepository) UpdateStripeRefundID(ctx context.Context, id uuid.UUID, stripeRefundID string) error {
	query := `
		UPDATE refund_requests
		SET stripe_refund_id = $1, updated_at = $2
		WHERE id = $3`

	_, err := r.db.ExecContext(ctx, query, stripeRefundID, time.Now(), id)
	return err
}

// GetStats retrieves aggregated refund statistics
func (r *refundRepository) GetStats(ctx context.Context, timeRange *types.TimeRange) (*RefundStats, error) {
	var stats RefundStats

	var err error

	// Get total refunds
	if timeRange != nil {
		err = r.db.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM refund_requests WHERE created_at >= $1 AND created_at <= $2`,
			timeRange.From, timeRange.To).Scan(&stats.TotalRefunds)
	} else {
		err = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM refund_requests`).Scan(&stats.TotalRefunds)
	}
	if err != nil {
		return nil, err
	}

	// Get total amount
	var totalAmount int64
	if timeRange != nil {
		err = r.db.QueryRowContext(ctx,
			`SELECT COALESCE(SUM(amount), 0) FROM refund_requests WHERE created_at >= $1 AND created_at <= $2`,
			timeRange.From, timeRange.To).Scan(&totalAmount)
	} else {
		err = r.db.QueryRowContext(ctx, `SELECT COALESCE(SUM(amount), 0) FROM refund_requests`).Scan(&totalAmount)
	}
	if err != nil {
		return nil, err
	}
	stats.TotalAmount = totalAmount

	// Get status counts
	statuses := map[string]*int64{
		models.RefundStatusProcessed: &stats.ProcessedRefunds,
		models.RefundStatusPending:   &stats.PendingRefunds,
		models.RefundStatusRejected:  &stats.RejectedRefunds,
		models.RefundStatusFailed:    &stats.FailedRefunds,
	}

	for status, count := range statuses {
		if timeRange != nil {
			err = r.db.QueryRowContext(ctx,
				`SELECT COUNT(*) FROM refund_requests WHERE status = $1 AND created_at >= $2 AND created_at <= $3`,
				status, timeRange.From, timeRange.To).Scan(count)
		} else {
			err = r.db.QueryRowContext(ctx,
				`SELECT COUNT(*) FROM refund_requests WHERE status = $1`,
				status).Scan(count)
		}
		if err != nil {
			return nil, err
		}
	}

	// Calculate average amount
	if stats.TotalRefunds > 0 {
		stats.AverageAmount = totalAmount / stats.TotalRefunds
	}

	return &stats, nil
}