package model

import (
	"github.com/google/uuid"
	shared "github.com/marketpay/backend/internal/shared/domain/model"
)

// PaymentStatus represents the lifecycle of a payment.
type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "PENDING"
	PaymentStatusCompleted PaymentStatus = "SUCCESS"
	PaymentStatusFailed    PaymentStatus = "FAILED"
	PaymentStatusRefunded  PaymentStatus = "REFUNDED"
)

const FeeRate = 0.01 // 1%

// Payment represents a customer-to-vendor payment.
type Payment struct {
	shared.BaseModel
	CustomerID      uuid.UUID     `gorm:"type:uuid;not null;index" json:"customer_id"`
	VendorID        uuid.UUID     `gorm:"type:uuid;not null;index" json:"vendor_id"`
	Amount          float64       `gorm:"type:decimal(15,2);not null" json:"amount"`
	Fee             float64       `gorm:"type:decimal(15,2);not null" json:"fee"`
	NetAmount       float64       `gorm:"type:decimal(15,2);not null" json:"net_amount"`
	Status          PaymentStatus `gorm:"type:varchar(50);not null;default:'PENDING'" json:"status"`
	MonimeReference string        `gorm:"type:varchar(255);index" json:"monime_reference,omitempty"`
	Description     string        `gorm:"type:text" json:"description,omitempty"`
	Currency        string        `gorm:"type:varchar(10);not null;default:'SLE'" json:"currency"`
	IsDemo          bool          `gorm:"default:false" json:"is_demo"`
}

// NewPayment creates a Payment with fee and net amount computed.
func NewPayment(customerID, vendorID uuid.UUID, amount float64) *Payment {
	fee := amount * FeeRate
	return &Payment{
		CustomerID: customerID,
		VendorID:   vendorID,
		Amount:     amount,
		Fee:        fee,
		NetAmount:  amount - fee,
		Status:     PaymentStatusPending,
		Currency:   "SLE",
	}
}

// Complete marks the payment as completed.
func (p *Payment) Complete(monimeRef string) {
	p.Status = PaymentStatusCompleted
	p.MonimeReference = monimeRef
}

// Fail marks the payment as failed.
func (p *Payment) Fail() {
	p.Status = PaymentStatusFailed
}
