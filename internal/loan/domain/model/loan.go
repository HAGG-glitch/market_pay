package model

import (
	"math"
	"time"

	"github.com/google/uuid"
	apperrors "github.com/marketpay/backend/pkg/errors"
	shared "github.com/marketpay/backend/internal/shared/domain/model"
)

// LoanType categorizes the loan product.
type LoanType string

const (
	LoanTypeEmergencyAdvance LoanType = "EMERGENCY_ADVANCE"
	LoanTypeStarterLoan      LoanType = "STARTER_LOAN"
	LoanTypeGrowthLoan       LoanType = "GROWTH_LOAN"
)

// LoanState represents the loan lifecycle.
type LoanState string

const (
	LoanStateDraft          LoanState = "DRAFT"
	LoanStatePendingReview  LoanState = "PENDING_REVIEW"
	LoanStateAutoApproved   LoanState = "AUTO_APPROVED"
	LoanStateUnderReview    LoanState = "UNDER_REVIEW"
	LoanStateApproved       LoanState = "APPROVED"
	LoanStateRejected       LoanState = "REJECTED"
	LoanStateDisbursed      LoanState = "DISBURSED"
	LoanStateActive         LoanState = "ACTIVE"
	LoanStateClosed         LoanState = "CLOSED"
	LoanStateDefaulted      LoanState = "DEFAULTED"
)

// InterestType determines how interest is calculated.
type InterestType string

const (
	InterestTypeFlat             InterestType = "FLAT"
	InterestTypeDecliningBalance InterestType = "DECLINING_BALANCE"
)

// FundingSource identifies who funds the loan.
type FundingSource string

const (
	FundingSourceMFIPartner  FundingSource = "MFI_PARTNER"
	FundingSourceNGOFund     FundingSource = "NGO_FUND"
	FundingSourceBankPartner FundingSource = "BANK_PARTNER"
)

// RepaymentFrequency is how often repayments are made.
type RepaymentFrequency string

const (
	RepaymentFrequencyBiweekly RepaymentFrequency = "BIWEEKLY"
	RepaymentFrequencyMonthly  RepaymentFrequency = "MONTHLY"
)

// validTransitions defines the state machine rules.
var validTransitions = map[LoanState][]LoanState{
	LoanStateDraft:         {LoanStatePendingReview},
	LoanStatePendingReview: {LoanStateAutoApproved, LoanStateUnderReview, LoanStateRejected},
	LoanStateAutoApproved:  {LoanStateDisbursed, LoanStateRejected},
	LoanStateUnderReview:   {LoanStateApproved, LoanStateRejected},
	LoanStateApproved:      {LoanStateDisbursed},
	LoanStateRejected:      {},
	LoanStateDisbursed:     {LoanStateActive},
	LoanStateActive:        {LoanStateClosed, LoanStateDefaulted},
	LoanStateClosed:        {},
	LoanStateDefaulted:     {LoanStateClosed},
}

// Loan is the aggregate root for the loan bounded context.
type Loan struct {
	shared.BaseModel
	VendorID          uuid.UUID          `gorm:"type:uuid;not null;index" json:"vendor_id"`
	GroupID           *uuid.UUID         `gorm:"type:uuid;index" json:"group_id,omitempty"`
	LoanType          LoanType           `gorm:"type:varchar(50);not null" json:"loan_type"`
	State             LoanState          `gorm:"type:varchar(50);not null;default:'DRAFT'" json:"state"`
	PrincipalAmount   float64            `gorm:"type:decimal(15,2);not null" json:"principal_amount"`
	InterestRate      float64            `gorm:"type:decimal(5,4);not null" json:"interest_rate"`
	InterestType      InterestType       `gorm:"type:varchar(30);not null" json:"interest_type"`
	TotalAmount       float64            `gorm:"type:decimal(15,2);not null" json:"total_amount"`
	OutstandingAmount float64            `gorm:"type:decimal(15,2)" json:"outstanding_amount"`
	TermWeeks         int                `gorm:"not null" json:"term_weeks"`
	Frequency         RepaymentFrequency `gorm:"type:varchar(20);not null" json:"frequency"`
	DisbursedAt       *time.Time         `json:"disbursed_at,omitempty"`
	DueDate           *time.Time         `json:"due_date,omitempty"`
	CreditScoreAtTime float64            `gorm:"type:decimal(5,2)" json:"credit_score_at_time"`
	FundedBy          FundingSource      `gorm:"type:varchar(50)" json:"funded_by"`
	PartnerID         *uuid.UUID         `gorm:"type:uuid;index" json:"partner_id,omitempty"`
	CommissionRate    float64            `gorm:"type:decimal(5,4)" json:"commission_rate"`
	CommissionPaid    bool               `gorm:"default:false" json:"commission_paid"`
	ReviewedBy        *uuid.UUID         `gorm:"type:uuid" json:"reviewed_by,omitempty"`
	ReviewNote        string             `gorm:"type:text" json:"review_note,omitempty"`
	RejectionReason   string             `gorm:"type:text" json:"rejection_reason,omitempty"`
	MonimeReference   string             `gorm:"type:varchar(255);index" json:"monime_reference,omitempty"`
	Schedules         []RepaymentSchedule `gorm:"foreignKey:LoanID" json:"schedules,omitempty"`
	Currency          string             `gorm:"type:varchar(10);not null;default:'SLE'" json:"currency"`
}

// RepaymentSchedule is a single installment on a loan.
type RepaymentSchedule struct {
	shared.BaseModel
	LoanID         uuid.UUID  `gorm:"type:uuid;not null;index" json:"loan_id"`
	InstallmentNo  int        `gorm:"not null" json:"installment_no"`
	DueDate        time.Time  `gorm:"not null" json:"due_date"`
	PrincipalDue   float64    `gorm:"type:decimal(15,2);not null" json:"principal_due"`
	InterestDue    float64    `gorm:"type:decimal(15,2);not null" json:"interest_due"`
	TotalDue       float64    `gorm:"type:decimal(15,2);not null" json:"total_due"`
	AmountPaid     float64    `gorm:"type:decimal(15,2);default:0" json:"amount_paid"`
	PenaltyAmount  float64    `gorm:"type:decimal(15,2);default:0" json:"penalty_amount"`
	Status         string     `gorm:"type:varchar(50);not null;default:'PENDING'" json:"status"`
	PaidAt         *time.Time `json:"paid_at,omitempty"`
	IsGracePeriod  bool       `gorm:"default:false" json:"is_grace_period"`
}

// Transition moves the loan to a new state, enforcing the state machine.
func (l *Loan) Transition(to LoanState) error {
	allowed, ok := validTransitions[l.State]
	if !ok {
		return apperrors.ErrInvalidLoanState(string(l.State), string(to))
	}
	for _, s := range allowed {
		if s == to {
			l.State = to
			return nil
		}
	}
	return apperrors.ErrInvalidLoanState(string(l.State), string(to))
}

// CanBeAutoApproved checks if this loan type supports auto-approval.
func (l *Loan) CanBeAutoApproved(autoApproveScore float64, vendorScore float64) bool {
	return l.LoanType == LoanTypeEmergencyAdvance && vendorScore >= autoApproveScore
}

// CalculateTotalAmount computes total repayment based on interest type.
func (l *Loan) CalculateTotalAmount() float64 {
	switch l.InterestType {
	case InterestTypeFlat:
		return l.PrincipalAmount * (1 + l.InterestRate)
	case InterestTypeDecliningBalance:
		// Monthly declining balance calculation
		months := l.TermWeeks / 4
		if months == 0 {
			months = 1
		}
		monthsF := float64(months)
		monthly := l.PrincipalAmount * l.InterestRate / monthsF /
			(1 - pow(1+l.InterestRate/monthsF, -monthsF))
		return monthly * monthsF
	default:
		return l.PrincipalAmount * (1 + l.InterestRate)
	}
}

// GenerateSchedule creates installment schedule entries.
func (l *Loan) GenerateSchedule() []RepaymentSchedule {
	if l.DisbursedAt == nil {
		return nil
	}

	var installments int
	var interval time.Duration

	switch l.Frequency {
	case RepaymentFrequencyBiweekly:
		installments = l.TermWeeks / 2
		interval = 14 * 24 * time.Hour
	case RepaymentFrequencyMonthly:
		installments = l.TermWeeks / 4
		interval = 28 * 24 * time.Hour
	default:
		installments = l.TermWeeks / 2
		interval = 14 * 24 * time.Hour
	}

	if installments == 0 {
		installments = 1
	}

	totalAmount := l.CalculateTotalAmount()
	installmentAmount := totalAmount / float64(installments)
	principalPerInstallment := l.PrincipalAmount / float64(installments)
	interestPerInstallment := installmentAmount - principalPerInstallment

	schedules := make([]RepaymentSchedule, installments)
	for i := 0; i < installments; i++ {
		dueDate := l.DisbursedAt.Add(interval * time.Duration(i+1))
		schedules[i] = RepaymentSchedule{
			LoanID:        l.ID,
			InstallmentNo: i + 1,
			DueDate:       dueDate,
			PrincipalDue:  principalPerInstallment,
			InterestDue:   interestPerInstallment,
			TotalDue:      installmentAmount,
			Status:        "PENDING",
		}
	}
	return schedules
}

// IsOverdue checks if the loan has missed payments.
func (l *Loan) IsOverdue() bool {
	if l.DueDate == nil {
		return false
	}
	return time.Now().After(*l.DueDate) && l.State == LoanStateActive
}

// IsInGracePeriod checks if the loan is within the grace period window.
func (l *Loan) IsInGracePeriod(gracePeriodDays int) bool {
	if l.DueDate == nil {
		return false
	}
	graceCutoff := l.DueDate.Add(time.Duration(gracePeriodDays) * 24 * time.Hour)
	return time.Now().Before(graceCutoff)
}

func pow(base, exp float64) float64 {
	return math.Pow(base, exp)
}
