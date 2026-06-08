package model

import (
	"time"

	"github.com/google/uuid"
	apperrors "github.com/marketpay/backend/pkg/errors"
	shared "github.com/marketpay/backend/internal/shared/domain/model"
)

// VendorStatus represents the lifecycle state of a vendor.
type VendorStatus string

const (
	VendorStatusPending   VendorStatus = "PENDING"
	VendorStatusActive    VendorStatus = "ACTIVE"
	VendorStatusSuspended VendorStatus = "SUSPENDED"
	VendorStatusBlacklist VendorStatus = "BLACKLISTED"
)

// MarketAssociation represents the market the vendor belongs to.
type MarketAssociation struct {
	shared.BaseModel
	Name     string `gorm:"type:varchar(255);not null;uniqueIndex" json:"name"`
	Location string `gorm:"type:varchar(255);not null" json:"location"`
	District string `gorm:"type:varchar(100);not null" json:"district"`
}

// Vendor is the aggregate root for the vendor bounded context.
type Vendor struct {
	shared.BaseModel
	UserID              uuid.UUID          `gorm:"type:uuid;not null;uniqueIndex" json:"user_id"`
	FirstName           string             `gorm:"type:varchar(100);not null" json:"first_name"`
	LastName            string             `gorm:"type:varchar(100);not null" json:"last_name"`
	Phone               string             `gorm:"type:varchar(20);not null;uniqueIndex" json:"phone"`
	NationalIDNumber    string             `gorm:"type:varchar(50);not null;uniqueIndex" json:"national_id_number"`
	NationalIDType      string             `gorm:"type:varchar(50);not null" json:"national_id_type"`
	DateOfBirth         time.Time          `gorm:"not null" json:"date_of_birth"`
	Address             string             `gorm:"type:text" json:"address"`
	MarketAssociationID uuid.UUID          `gorm:"type:uuid;not null" json:"market_association_id"`
	MarketAssociation   *MarketAssociation `gorm:"foreignKey:MarketAssociationID" json:"market_association,omitempty"`
	BusinessName        string             `gorm:"type:varchar(255)" json:"business_name"`
	BusinessType        string             `gorm:"type:varchar(100)" json:"business_type"`
	KYCStatus           shared.KYCStatus   `gorm:"type:varchar(50);not null;default:'PENDING'" json:"kyc_status"`
	KYCVerifiedAt       *time.Time         `json:"kyc_verified_at,omitempty"`
	Status              VendorStatus       `gorm:"type:varchar(50);not null;default:'PENDING'" json:"status"`
	PINHash             string             `gorm:"type:varchar(255);not null" json:"-"`
	TransactionCount    int                `gorm:"default:0" json:"transaction_count"`
	FirstTransactionAt  *time.Time         `json:"first_transaction_at,omitempty"`
	CreditScore         float64            `gorm:"default:0" json:"credit_score"`
	GroupID             *uuid.UUID         `gorm:"type:uuid;index" json:"group_id,omitempty"`
}

// FullName returns the vendor's full name.
func (v *Vendor) FullName() string {
	return v.FirstName + " " + v.LastName
}

// Age calculates the vendor's current age.
func (v *Vendor) Age() int {
	now := time.Now()
	years := now.Year() - v.DateOfBirth.Year()
	if now.YearDay() < v.DateOfBirth.YearDay() {
		years--
	}
	return years
}

// IsAdult checks if the vendor meets the minimum age requirement.
func (v *Vendor) IsAdult() bool {
	return v.Age() >= 18
}

// HasSufficientTransactionHistory checks the 30-day minimum rule.
func (v *Vendor) HasSufficientTransactionHistory() bool {
	if v.FirstTransactionAt == nil {
		return false
	}
	return time.Since(*v.FirstTransactionAt).Hours() >= 24*30
}

// IsEligibleForLoan checks all loan eligibility criteria.
func (v *Vendor) IsEligibleForLoan() error {
	if v.Status != VendorStatusActive {
		return apperrors.ErrVendorNotEligible
	}
	if v.KYCStatus != shared.KYCVerified {
		return apperrors.ErrVendorNotEligible
	}
	if !v.HasSufficientTransactionHistory() {
		return apperrors.ErrInsufficientTransactionHistory
	}
	return nil
}

// IsKYCComplete checks if all KYC fields are filled.
func (v *Vendor) IsKYCComplete() bool {
	return v.NationalIDNumber != "" &&
		v.NationalIDType != "" &&
		v.MarketAssociationID != uuid.Nil &&
		!v.DateOfBirth.IsZero()
}

// KYCCompletenessScore returns a 0-100 score for KYC completeness.
func (v *Vendor) KYCCompletenessScore() float64 {
	score := 0.0
	if v.NationalIDNumber != "" {
		score += 25
	}
	if v.NationalIDType != "" {
		score += 25
	}
	if v.MarketAssociationID != uuid.Nil {
		score += 25
	}
	if !v.DateOfBirth.IsZero() {
		score += 25
	}
	return score
}

// Activate sets the vendor status to active.
func (v *Vendor) Activate() {
	v.Status = VendorStatusActive
}

// Suspend sets the vendor status to suspended.
func (v *Vendor) Suspend() {
	v.Status = VendorStatusSuspended
}
