package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	loanmodel "github.com/marketpay/backend/internal/loan/domain/model"
	shared "github.com/marketpay/backend/internal/shared/domain/model"
	apperrors "github.com/marketpay/backend/pkg/errors"
	"github.com/marketpay/backend/pkg/config"
	"github.com/marketpay/backend/pkg/logger"
	"go.uber.org/zap"
)

// LoanRepository defines persistence for loans.
type LoanRepository interface {
	Create(ctx context.Context, loan *loanmodel.Loan) error
	FindByID(ctx context.Context, id uuid.UUID) (*loanmodel.Loan, error)
	FindByVendorID(ctx context.Context, vendorID uuid.UUID, offset, limit int) ([]*loanmodel.Loan, int64, error)
	Update(ctx context.Context, loan *loanmodel.Loan) error
	SaveSchedules(ctx context.Context, schedules []loanmodel.RepaymentSchedule) error
	FindSchedulesByLoanID(ctx context.Context, loanID uuid.UUID) ([]loanmodel.RepaymentSchedule, error)
	UpdateSchedule(ctx context.Context, schedule *loanmodel.RepaymentSchedule) error
	ListByState(ctx context.Context, state loanmodel.LoanState, isDemo bool, offset, limit int) ([]*loanmodel.Loan, int64, error)
}

// AuditRepository persists audit log entries.
type AuditRepository interface {
	Log(ctx context.Context, log *shared.AuditLog) error
}

// EventPublisher publishes domain events.
type EventPublisher interface {
	Publish(ctx context.Context, eventType, aggregateID string, payload interface{}) error
}

// CreditScoreService calculates and retrieves vendor credit scores.
type CreditScoreService interface {
	GetScore(ctx context.Context, vendorID uuid.UUID) (float64, bool, error)
}

// VendorEligibilityChecker checks vendor loan eligibility.
type VendorEligibilityChecker interface {
	CheckEligibility(ctx context.Context, vendorID uuid.UUID) error
}

// VendorPhoneFinder looks up a vendor's phone number by ID.
type VendorPhoneFinder interface {
	FindPhoneByID(ctx context.Context, vendorID uuid.UUID) (string, error)
}

// MonimePayoutDisburser sends money to a vendor via Monime Payout API.
type MonimePayoutDisburser interface {
	Disburse(ctx context.Context, phone string, amount float64) (string, error)
}

// LoanService handles loan origination and lifecycle.
type LoanService struct {
	loans            LoanRepository
	audit            AuditRepository
	events           EventPublisher
	creditScore      CreditScoreService
	eligibility      VendorEligibilityChecker
	vendorPhone      VendorPhoneFinder
	monimePayout     MonimePayoutDisburser
	cfg              config.LoanProductsConfig
	scoreCfg         config.CreditScoreConfig
	log              *logger.Logger
}

// NewLoanService constructs a LoanService.
func NewLoanService(
	loans LoanRepository,
	audit AuditRepository,
	events EventPublisher,
	creditScore CreditScoreService,
	eligibility VendorEligibilityChecker,
	vendorPhone VendorPhoneFinder,
	monimePayout MonimePayoutDisburser,
	cfg config.LoanProductsConfig,
	scoreCfg config.CreditScoreConfig,
	log *logger.Logger,
) *LoanService {
	return &LoanService{
		loans:        loans,
		audit:        audit,
		events:       events,
		creditScore:  creditScore,
		eligibility:  eligibility,
		vendorPhone:  vendorPhone,
		monimePayout: monimePayout,
		cfg:          cfg,
		scoreCfg:     scoreCfg,
		log:          log,
	}
}

// ApplyInput holds the loan application payload.
type ApplyInput struct {
	VendorID  uuid.UUID
	LoanType  loanmodel.LoanType
	Amount    float64
	TermWeeks int
	Frequency loanmodel.RepaymentFrequency
	GroupID   *uuid.UUID
	FundedBy  loanmodel.FundingSource
	PartnerID *uuid.UUID
	IsDemo    bool
}

// Apply creates a loan application and triggers auto-approval if eligible.
func (s *LoanService) Apply(ctx context.Context, input ApplyInput) (*loanmodel.Loan, error) {
	if err := s.eligibility.CheckEligibility(ctx, input.VendorID); err != nil {
		return nil, err
	}

	if err := s.validateAmount(input.LoanType, input.Amount); err != nil {
		return nil, err
	}

	score, canAutoApprove, err := s.creditScore.GetScore(ctx, input.VendorID)
	if err != nil {
		return nil, err
	}

	if score < s.scoreCfg.MinScore {
		return nil, apperrors.ErrInsufficientCreditScore
	}

	interestType, rate := s.productConfig(input.LoanType)

	loan := &loanmodel.Loan{
		VendorID:          input.VendorID,
		GroupID:           input.GroupID,
		LoanType:          input.LoanType,
		State:             loanmodel.LoanStateDraft,
		PrincipalAmount:   input.Amount,
		InterestRate:      rate,
		InterestType:      interestType,
		TermWeeks:         input.TermWeeks,
		Frequency:         input.Frequency,
		CreditScoreAtTime: score,
		FundedBy:          input.FundedBy,
		PartnerID:         input.PartnerID,
		IsDemo:            input.IsDemo,
	}
	loan.TotalAmount = loan.CalculateTotalAmount()
	loan.OutstandingAmount = loan.TotalAmount

	if err := s.loans.Create(ctx, loan); err != nil {
		return nil, apperrors.ErrInternalServer(err)
	}

	_ = loan.Transition(loanmodel.LoanStatePendingReview)

	if input.LoanType == loanmodel.LoanTypeEmergencyAdvance && canAutoApprove {
		_ = loan.Transition(loanmodel.LoanStateAutoApproved)
		s.log.Info("loan auto-approved", zap.String("loan_id", loan.ID.String()))
	}

	if err := s.loans.Update(ctx, loan); err != nil {
		return nil, apperrors.ErrInternalServer(err)
	}

	_ = s.audit.Log(ctx, &shared.AuditLog{
		ActorID:    input.VendorID,
		ActorRole:  shared.RoleVendor,
		Action:     "LOAN_APPLIED",
		Resource:   "loan",
		ResourceID: loan.ID.String(),
		NewState:   string(loan.State),
	})

	payload := map[string]interface{}{
		"loan_id":   loan.ID.String(),
		"vendor_id": loan.VendorID.String(),
		"amount":    loan.PrincipalAmount,
		"type":      string(loan.LoanType),
		"state":     string(loan.State),
		"is_demo":   loan.IsDemo,
	}
	_ = s.events.Publish(ctx, "LoanApplied", loan.ID.String(), payload)
	_ = s.events.Publish(ctx, "LoanRequested", loan.ID.String(), payload)

	return loan, nil
}

// Approve moves a loan to APPROVED state.
func (s *LoanService) Approve(ctx context.Context, loanID, officerID uuid.UUID, note string) (*loanmodel.Loan, error) {
	loan, err := s.loans.FindByID(ctx, loanID)
	if err != nil {
		return nil, apperrors.ErrNotFound("loan")
	}

	oldState := loan.State
	if err := loan.Transition(loanmodel.LoanStateApproved); err != nil {
		return nil, err
	}
	loan.ReviewedBy = &officerID
	loan.ReviewNote = note

	if err := s.loans.Update(ctx, loan); err != nil {
		return nil, apperrors.ErrInternalServer(err)
	}

	_ = s.audit.Log(ctx, &shared.AuditLog{
		ActorID:    officerID,
		ActorRole:  shared.RoleLoanOfficer,
		Action:     "LOAN_APPROVED",
		Resource:   "loan",
		ResourceID: loan.ID.String(),
		OldState:   string(oldState),
		NewState:   string(loan.State),
	})

	_ = s.events.Publish(ctx, "LoanApproved", loan.ID.String(), map[string]interface{}{
		"loan_id":   loan.ID.String(),
		"vendor_id": loan.VendorID.String(),
	})

	return loan, nil
}

// Reject moves a loan to REJECTED state.
func (s *LoanService) Reject(ctx context.Context, loanID, officerID uuid.UUID, reason string) (*loanmodel.Loan, error) {
	loan, err := s.loans.FindByID(ctx, loanID)
	if err != nil {
		return nil, apperrors.ErrNotFound("loan")
	}

	oldState := loan.State
	if err := loan.Transition(loanmodel.LoanStateRejected); err != nil {
		return nil, err
	}
	loan.ReviewedBy = &officerID
	loan.RejectionReason = reason

	if err := s.loans.Update(ctx, loan); err != nil {
		return nil, apperrors.ErrInternalServer(err)
	}

	_ = s.audit.Log(ctx, &shared.AuditLog{
		ActorID:    officerID,
		ActorRole:  shared.RoleLoanOfficer,
		Action:     "LOAN_REJECTED",
		Resource:   "loan",
		ResourceID: loan.ID.String(),
		OldState:   string(oldState),
		NewState:   string(loan.State),
	})

	_ = s.events.Publish(ctx, "LoanRejected", loan.ID.String(), map[string]interface{}{
		"loan_id": loan.ID.String(),
		"reason":  reason,
	})

	return loan, nil
}

// Disburse marks a loan as disbursed and generates the repayment schedule.
func (s *LoanService) Disburse(ctx context.Context, loanID uuid.UUID, monimeRef string) (*loanmodel.Loan, error) {
	loan, err := s.loans.FindByID(ctx, loanID)
	if err != nil {
		return nil, apperrors.ErrNotFound("loan")
	}

	oldState := loan.State
	if err := loan.Transition(loanmodel.LoanStateDisbursed); err != nil {
		return nil, err
	}

	now := time.Now()
	dueDate := now.Add(time.Duration(loan.TermWeeks) * 7 * 24 * time.Hour)
	loan.DisbursedAt = &now
	loan.DueDate = &dueDate
	loan.MonimeReference = monimeRef

	if err := s.loans.Update(ctx, loan); err != nil {
		return nil, apperrors.ErrInternalServer(err)
	}

	schedules := loan.GenerateSchedule()
	if len(schedules) > 0 {
		_ = s.loans.SaveSchedules(ctx, schedules)
	}

	_ = loan.Transition(loanmodel.LoanStateActive)
	_ = s.loans.Update(ctx, loan)

	_ = s.audit.Log(ctx, &shared.AuditLog{
		Action:     "LOAN_DISBURSED",
		Resource:   "loan",
		ResourceID: loan.ID.String(),
		OldState:   string(oldState),
		NewState:   string(loan.State),
	})

	_ = s.events.Publish(ctx, "LoanDisbursed", loan.ID.String(), map[string]interface{}{
		"loan_id":          loan.ID.String(),
		"vendor_id":        loan.VendorID.String(),
		"amount":           loan.PrincipalAmount,
		"monime_reference": monimeRef,
	})

	return loan, nil
}

// DisburseWithPayout calls Monime Payout API and records the disbursement.
func (s *LoanService) DisburseWithPayout(ctx context.Context, loanID uuid.UUID) (*loanmodel.Loan, error) {
	loan, err := s.loans.FindByID(ctx, loanID)
	if err != nil {
		return nil, apperrors.ErrNotFound("loan")
	}

	phone, err := s.vendorPhone.FindPhoneByID(ctx, loan.VendorID)
	if err != nil {
		return nil, apperrors.ErrInternalServer(fmt.Errorf("lookup vendor phone: %w", err))
	}

	monimeRef, err := s.monimePayout.Disburse(ctx, phone, loan.PrincipalAmount)
	if err != nil {
		return nil, apperrors.ErrInternalServer(fmt.Errorf("monime payout failed: %w", err))
	}

	return s.Disburse(ctx, loanID, monimeRef)
}

// GetByID retrieves a loan by ID.
func (s *LoanService) GetByID(ctx context.Context, id uuid.UUID) (*loanmodel.Loan, error) {
	loan, err := s.loans.FindByID(ctx, id)
	if err != nil {
		return nil, apperrors.ErrNotFound("loan")
	}
	return loan, nil
}

// GetVendorLoans returns paginated loans for a vendor.
func (s *LoanService) GetVendorLoans(ctx context.Context, vendorID uuid.UUID, offset, limit int) ([]*loanmodel.Loan, int64, error) {
	return s.loans.FindByVendorID(ctx, vendorID, offset, limit)
}

// GetSchedule returns the repayment schedule for a loan.
func (s *LoanService) GetSchedule(ctx context.Context, loanID uuid.UUID) ([]loanmodel.RepaymentSchedule, error) {
	return s.loans.FindSchedulesByLoanID(ctx, loanID)
}

func (s *LoanService) validateAmount(loanType loanmodel.LoanType, amount float64) error {
	var min, max float64
	switch loanType {
	case loanmodel.LoanTypeEmergencyAdvance:
		min = s.cfg.EmergencyAdvance.MinAmount
		max = s.cfg.EmergencyAdvance.MaxAmount
	case loanmodel.LoanTypeStarterLoan:
		min = s.cfg.StarterLoan.MinAmount
		max = s.cfg.StarterLoan.MaxAmount
	case loanmodel.LoanTypeGrowthLoan:
		min = s.cfg.GrowthLoan.MinAmount
		max = s.cfg.GrowthLoan.MaxAmount
	default:
		return apperrors.ErrBadRequest("unknown loan type")
	}

	if amount < min || amount > max {
		return apperrors.ErrInvalidAmount(
			fmt.Sprintf("amount must be between %.2f and %.2f SLE", min, max),
		)
	}
	return nil
}

func (s *LoanService) productConfig(loanType loanmodel.LoanType) (loanmodel.InterestType, float64) {
	switch loanType {
	case loanmodel.LoanTypeEmergencyAdvance:
		return loanmodel.InterestTypeFlat, s.cfg.EmergencyAdvance.InterestRate
	case loanmodel.LoanTypeStarterLoan:
		return loanmodel.InterestTypeFlat, s.cfg.StarterLoan.InterestRate
	case loanmodel.LoanTypeGrowthLoan:
		return loanmodel.InterestTypeDecliningBalance, s.cfg.GrowthLoan.InterestRate
	default:
		return loanmodel.InterestTypeFlat, 0.05
	}
}
