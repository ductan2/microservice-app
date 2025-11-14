package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// Invoice represents billing invoices for completed orders
type Invoice struct {
	ID             uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	OrderID        uuid.UUID `gorm:"type:uuid;not null;uniqueIndex;index:invoices_order_idx;constraint:OnDelete:CASCADE" json:"order_id"`
	UserID         uuid.UUID `gorm:"type:uuid;not null;index:invoices_user_idx" json:"user_id"`
	InvoiceNumber  string    `gorm:"type:text;uniqueIndex;not null" json:"invoice_number"`
	TotalAmount    int64     `gorm:"type:bigint;not null" json:"total_amount"` // in cents
	Currency       string    `gorm:"type:varchar(3);default:'USD';not null" json:"currency"`
	BillingAddress string    `gorm:"type:jsonb" json:"billing_address,omitempty"`
	TaxAmount      int64     `gorm:"type:bigint;default:0" json:"tax_amount"` // in cents
	TaxBreakdown   string    `gorm:"type:jsonb" json:"tax_breakdown,omitempty"` // detailed tax info
	PDFURL         string    `gorm:"type:text" json:"pdf_url,omitempty"`
	StripeChargeID string    `gorm:"type:text;index:invoices_charge_idx" json:"stripe_charge_id,omitempty"`
	Status         string    `gorm:"type:varchar(20);default:'draft';check:status IN ('draft','sent','paid','void')" json:"status"`
	IssuedAt       time.Time `gorm:"default:now();not null" json:"issued_at"`
	DueAt          time.Time `gorm:"not null" json:"due_at"`
	PaidAt         sql.NullTime `gorm:"type:timestamptz" json:"paid_at,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`

	// Relationships
	Order          Order     `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;foreignKey:OrderID;references:ID" json:"order,omitempty"`
}

// Invoice Status Constants
const (
	InvoiceStatusDraft = "draft"
	InvoiceStatusSent  = "sent"
	InvoiceStatusPaid  = "paid"
	InvoiceStatusVoid  = "void"
)

// RefundRequest represents customer refund requests
type RefundRequest struct {
	ID             uuid.UUID    `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	OrderID        uuid.UUID    `gorm:"type:uuid;not null;index:refund_requests_order_idx;constraint:OnDelete:CASCADE" json:"order_id"`
	UserID         uuid.UUID    `gorm:"type:uuid;not null;index:refund_requests_user_idx" json:"user_id"`
	Reason         string       `gorm:"type:text;not null" json:"reason"`
	ReasonCategory string       `gorm:"type:varchar(50);not null;check:reason_category IN ('technical','content','accidental','duplicate','quality','other')" json:"reason_category"`
	Amount         int64        `gorm:"type:bigint;not null" json:"amount"` // refund amount in cents
	Status         string       `gorm:"type:varchar(20);default:'pending';check:status IN ('pending','approved','rejected','processed','failed','cancelled')" json:"status"`
	AdminReason    *string      `gorm:"type:text" json:"admin_reason,omitempty"`
	AdminNotes     string       `gorm:"type:text" json:"admin_notes,omitempty"`
	ProcessedBy    *uuid.UUID   `gorm:"type:uuid" json:"processed_by,omitempty"` // admin who processed
	ProcessedAt    sql.NullTime `gorm:"type:timestamptz" json:"processed_at,omitempty"`
	StripeRefundID *string      `gorm:"type:text;index:refund_requests_stripe_idx" json:"stripe_refund_id,omitempty"`
	CreatedAt      time.Time    `json:"created_at"`
	UpdatedAt      time.Time    `json:"updated_at"`

	// Relationships
	Order          Order `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;foreignKey:OrderID;references:ID" json:"order,omitempty"`
}

// RefundRequest Status Constants
const (
	RefundStatusPending   = "pending"
	RefundStatusApproved  = "approved"
	RefundStatusRejected  = "rejected"
	RefundStatusProcessed = "processed"
	RefundStatusFailed    = "failed"
	RefundStatusCancelled = "cancelled"
)

// RefundRequest Reason Category Constants
const (
	RefundReasonTechnical  = "technical"
	RefundReasonContent    = "content"
	RefundReasonAccidental = "accidental"
	RefundReasonDuplicate  = "duplicate"
	RefundReasonQuality    = "quality"
	RefundReasonOther      = "other"
)

// FraudLog tracks potential fraudulent activities
type FraudLog struct {
	ID         uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	OrderID    uuid.UUID `gorm:"type:uuid;not null;index:fraud_logs_order_idx;constraint:OnDelete:CASCADE" json:"order_id"`
	UserID     uuid.UUID `gorm:"type:uuid;not null;index:fraud_logs_user_idx" json:"user_id"`
	RiskLevel  string    `gorm:"type:varchar(20);not null;check:risk_level IN ('low','medium','high','critical')" json:"risk_level"`
	RiskScore  float64   `gorm:"type:decimal(5,2);not null;check:risk_score >= 0 AND risk_score <= 100" json:"risk_score"`
	Reasons    string    `gorm:"type:jsonb" json:"reasons"` // array of reasons for the risk score
	Details    string    `gorm:"type:jsonb" json:"details"` // detailed analysis (IP, device, etc.)
	Action     string    `gorm:"type:varchar(20);not null;check:action IN ('none','review','block','manual_review')" json:"action"`
	ReviewedBy *uuid.UUID `gorm:"type:uuid" json:"reviewed_by,omitempty"` // admin who reviewed
	ReviewedAt sql.NullTime `gorm:"type:timestamptz" json:"reviewed_at,omitempty"`
	IPAddress  string    `gorm:"type:inet" json:"ip_address,omitempty"`
	UserAgent  string    `gorm:"type:text" json:"user_agent,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`

	// Relationships
	Order      Order `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;foreignKey:OrderID;references:ID" json:"order,omitempty"`
}

// FraudLog Risk Level Constants
const (
	FraudRiskLevelLow      = "low"
	FraudRiskLevelMedium   = "medium"
	FraudRiskLevelHigh     = "high"
	FraudRiskLevelCritical = "critical"
)

// FraudLog Action Constants
const (
	FraudActionNone        = "none"
	FraudActionReview      = "review"
	FraudActionBlock       = "block"
	FraudActionManualReview = "manual_review"
)