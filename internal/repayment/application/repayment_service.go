package application

import (
	"context"
	"time"

	"github.com/google/uuid"
	loanmodel "github.com/marketpay/backend/internal/loan/domain/model"
	repaymodel "github.com/marketpay/backend/internal/repayment/domain/model"
	shared "github.com/marketpay/backend/internal/shared/domain/model"
	apperrors "github.com/marketpay/backend/pkg/errors"
	"github.com/marketpay/backend/pkg/config"
	"github.com/marketpay/backend/pkg/logger"
	"go.uber.org/zap"
)

// RepaymentRepository handles repayment persistence.
type RepaymentRepository interface {
	FindScheduleByID(ctx context.Context, id uuid.UUID) (*loanmodel.RepaymentSchedule, error)
	UpdateSchedule(ctx context.Context, s *loanmodel.RepaymentSchedule) error
	FindLoanByID(ctx context.Context, id uuid.UUID) (*loanmodel.Loan, error)
	UpdateLoan(ctx context.Context, loan *loanmodel.Loan) error
	FindOverdueSchedules(ctx context.Context) ([]loanmodel.RepaymentSchedule, error)
	CreateRepayment(ctx context.Context, repayment *repaymodel.LoanRepayment) error
	FindRepaymentByMonimeRef(ctx context.Context, monimeRef string) (*repaymodel.LoanRepayment, error)
	FindRepaymentByPaymentRef(ctx context.Context, paymentRef string) (*repaymodel.LoanRepayment, error)
	UpdateRepayment(ctx context.Context, repayment *repaymodel.LoanRepayment) error
}

// AuditRepository logs events.
type AuditRepository interface {
	Log(ctx context.Context, log *shared.AuditLog) error
}

// EventPublisher publishes domain events.
type EventPublisher interface {
	Publish(ctx context.Context, eventType, aggregateID string, payload interface{}) error
}

// RepaymentService handles loan repayments.
type RepaymentService struct {
	repo      RepaymentRepository
	audit     AuditRepository
	events    EventPublisher
	cfg       config.RepaymentConfig
	log       *logger.Logger
}

// NewRepaymentService constructs a RepaymentService.
func NewRepaymentService(
	repo RepaymentRepository,
	audit AuditRepository,
	events EventPublisher,
	cfg config.RepaymentConfig,
	log *logger.Logger,
) *RepaymentService {
	return &RepaymentService{repo: repo, audit: audit, events: events, cfg: cfg, log: log}
}

// RepayInput holds repayment data.
type RepayInput struct {
	LoanID     uuid.UUID
	VendorID   uuid.UUID
	Amount     float64
	MonimeRef  string
}

// Repay processes a repayment against the oldest outstanding schedule.
func (s *RepaymentService) Repay(ctx context.Context, input RepayInput) (*loanmodel.Loan, error) {
	loan, err := s.repo.FindLoanByID(ctx, input.LoanID)
	if err != nil {
		return nil, apperrors.ErrNotFound("loan")
	}

	if loan.State != loanmodel.LoanStateActive {
		return nil, apperrors.ErrBadRequest("loan is not active")
	}

	remaining := input.Amount

	// Fetch all pending schedules sorted by due date
	schedules, err := s.repo.FindOverdueSchedules(ctx)
	if err != nil {
		return nil, apperrors.ErrInternalServer(err)
	}

	// Filter to this loan's schedules (include PARTIAL — still has outstanding balance)
	var loanSchedules []loanmodel.RepaymentSchedule
	for _, sc := range schedules {
		if sc.LoanID == input.LoanID && sc.Status != "PAID" {
			loanSchedules = append(loanSchedules, sc)
		}
	}

	for i := range loanSchedules {
		if remaining <= 0 {
			break
		}

		sc := &loanSchedules[i]
		outstanding := sc.TotalDue - sc.AmountPaid

		// Apply penalty if overdue beyond grace period
		if time.Now().After(sc.DueDate.Add(time.Duration(s.cfg.GracePeriodDays) * 24 * time.Hour)) {
			penalty := outstanding * s.cfg.DefaultPenaltyRate
			sc.PenaltyAmount += penalty
			outstanding += penalty
		}

		payment := remaining
		if payment > outstanding {
			payment = outstanding
		}

		sc.AmountPaid += payment
		remaining -= payment

		if sc.AmountPaid >= sc.TotalDue+sc.PenaltyAmount {
			sc.Status = "PAID"
			now := time.Now()
			sc.PaidAt = &now
		} else {
			sc.Status = "PARTIAL"
		}

		if err := s.repo.UpdateSchedule(ctx, sc); err != nil {
			s.log.Error("failed to update schedule", zap.Error(err))
		}
	}

	// Reduce outstanding loan balance
	loan.OutstandingAmount -= (input.Amount - remaining)
	if loan.OutstandingAmount < 0 {
		loan.OutstandingAmount = 0
	}

	// Check if fully repaid
	if loan.OutstandingAmount == 0 {
		_ = loan.Transition(loanmodel.LoanStateClosed)
		s.log.Info("loan fully repaid and closed", zap.String("loan_id", loan.ID.String()))
	}

	if err := s.repo.UpdateLoan(ctx, loan); err != nil {
		return nil, apperrors.ErrInternalServer(err)
	}

	_ = s.audit.Log(ctx, &shared.AuditLog{
		ActorID:    input.VendorID,
		ActorRole:  shared.RoleVendor,
		Action:     "REPAYMENT_RECEIVED",
		Resource:   "loan",
		ResourceID: loan.ID.String(),
		NewState:   string(loan.State),
	})

	_ = s.events.Publish(ctx, "RepaymentReceived", loan.ID.String(), map[string]interface{}{
		"loan_id":    loan.ID.String(),
		"vendor_id":  loan.VendorID.String(),
		"amount":     input.Amount,
		"monime_ref": input.MonimeRef,
	})

	return loan, nil
}

// RecordRepaymentInput holds data to create a new loan repayment record.
type RecordRepaymentInput struct {
	LoanID     uuid.UUID
	VendorID   uuid.UUID
	Amount     float64
	MonimeRef  string
	PaymentRef string
	Metadata   map[string]interface{}
}

// RecordRepayment creates a pending LoanRepayment record for webhook reconciliation.
func (s *RepaymentService) RecordRepayment(ctx context.Context, input RecordRepaymentInput) (*repaymodel.LoanRepayment, error) {
	repayment := repaymodel.NewLoanRepayment(
		input.LoanID, input.VendorID, input.Amount,
		input.MonimeRef, input.PaymentRef, input.Metadata,
	)
	if err := s.repo.CreateRepayment(ctx, repayment); err != nil {
		return nil, apperrors.ErrInternalServer(err)
	}
	s.log.Info("repayment recorded",
		zap.String("payment_ref", input.PaymentRef),
		zap.String("monime_ref", input.MonimeRef),
		zap.Float64("amount", input.Amount),
	)
	return repayment, nil
}

// ConfirmRepayment marks a LoanRepayment as completed and applies it to the loan.
func (s *RepaymentService) ConfirmRepayment(ctx context.Context, monimeRef string) error {
	repayment, err := s.repo.FindRepaymentByMonimeRef(ctx, monimeRef)
	if err != nil {
		return apperrors.ErrNotFound("repayment")
	}
	if repayment.Status != repaymodel.RepaymentStatusPending {
		s.log.Warn("repayment not in PENDING state",
			zap.String("payment_ref", repayment.PaymentRef),
			zap.String("status", repayment.Status),
		)
		return nil
	}

	repayment.Confirm()
	if err := s.repo.UpdateRepayment(ctx, repayment); err != nil {
		return apperrors.ErrInternalServer(err)
	}

	// Apply to loan schedules
	_, err = s.Repay(ctx, RepayInput{
		LoanID:    repayment.LoanID,
		VendorID:  repayment.VendorID,
		Amount:    repayment.Amount,
		MonimeRef: repayment.MonimeRef,
	})
	if err != nil {
		s.log.Error("failed to apply repayment to loan",
			zap.String("loan_id", repayment.LoanID.String()),
			zap.Error(err),
		)
	}

	_ = s.audit.Log(ctx, &shared.AuditLog{
		ActorID:    repayment.VendorID,
		ActorRole:  shared.RoleVendor,
		Action:     "REPAYMENT_CONFIRMED",
		Resource:   "loan_repayment",
		ResourceID: repayment.ID.String(),
		NewState:   repaymodel.RepaymentStatusCompleted,
	})

	return nil
}

// FailRepayment marks a LoanRepayment as failed.
func (s *RepaymentService) FailRepayment(ctx context.Context, monimeRef string) error {
	repayment, err := s.repo.FindRepaymentByMonimeRef(ctx, monimeRef)
	if err != nil {
		return apperrors.ErrNotFound("repayment")
	}
	repayment.Fail()
	if err := s.repo.UpdateRepayment(ctx, repayment); err != nil {
		return apperrors.ErrInternalServer(err)
	}
	s.log.Warn("repayment failed",
		zap.String("payment_ref", repayment.PaymentRef),
		zap.String("monime_ref", monimeRef),
	)
	return nil
}

// MarkDefaulted flags a loan as defaulted and freezes the group if applicable.
func (s *RepaymentService) MarkDefaulted(ctx context.Context, loanID uuid.UUID) error {
	loan, err := s.repo.FindLoanByID(ctx, loanID)
	if err != nil {
		return apperrors.ErrNotFound("loan")
	}

	if err := loan.Transition(loanmodel.LoanStateDefaulted); err != nil {
		return err
	}

	if err := s.repo.UpdateLoan(ctx, loan); err != nil {
		return apperrors.ErrInternalServer(err)
	}

	_ = s.events.Publish(ctx, "LoanDefaulted", loan.ID.String(), map[string]interface{}{
		"loan_id":   loan.ID.String(),
		"vendor_id": loan.VendorID.String(),
		"group_id":  loan.GroupID,
	})

	return nil
}
