package model

import (
	"time"

	"github.com/google/uuid"
	shared "github.com/marketpay/backend/internal/shared/domain/model"
)

const (
	RepaymentStatusPending   = "PENDING"
	RepaymentStatusCompleted = "COMPLETED"
	RepaymentStatusFailed    = "FAILED"
)

type LoanRepayment struct {
	shared.BaseModel
	LoanID     uuid.UUID              `gorm:"type:uuid;not null;index" json:"loan_id"`
	VendorID   uuid.UUID              `gorm:"type:uuid;not null;index" json:"vendor_id"`
	Amount     float64                `gorm:"type:decimal(15,2);not null" json:"amount"`
	MonimeRef  string                 `gorm:"type:varchar(255);not null;index" json:"monime_ref"`
	PaymentRef string                 `gorm:"type:varchar(255);not null;uniqueIndex" json:"payment_ref"`
	Metadata   map[string]interface{} `gorm:"type:jsonb;serializer:json;default:'{}'" json:"metadata"`
	Status     string                 `gorm:"type:varchar(50);not null;default:'PENDING'" json:"status"`
	PaidAt     *time.Time             `json:"paid_at,omitempty"`
}

func NewLoanRepayment(loanID, vendorID uuid.UUID, amount float64, monimeRef, paymentRef string, metadata map[string]interface{}) *LoanRepayment {
	return &LoanRepayment{
		LoanID:     loanID,
		VendorID:   vendorID,
		Amount:     amount,
		MonimeRef:  monimeRef,
		PaymentRef: paymentRef,
		Metadata:   metadata,
		Status:     RepaymentStatusPending,
	}
}

func (r *LoanRepayment) Confirm() {
	r.Status = RepaymentStatusCompleted
	now := time.Now()
	r.PaidAt = &now
}

func (r *LoanRepayment) Fail() {
	r.Status = RepaymentStatusFailed
}
