package application

import (
	"context"
	"time"

	"github.com/google/uuid"
	loanmodel "github.com/marketpay/backend/internal/loan/domain/model"
	apperrors "github.com/marketpay/backend/pkg/errors"
	shared "github.com/marketpay/backend/internal/shared/domain/model"
)

// PaymentPlanSummary is a loan with aggregated schedule data for the payment plans view.
type PaymentPlanSummary struct {
	*loanmodel.Loan
	TotalPaid       float64    `json:"total_paid"`
	Remaining       float64    `json:"remaining"`
	ProgressPercent float64    `json:"progress_percent"`
	NextDueDate     *time.Time `json:"next_due_date"`
	ScheduleCount   int        `json:"schedule_count"`
	PaidCount       int        `json:"paid_count"`
}

// ListPaymentPlans returns loans with computed schedule summaries.
func (s *LoanService) ListPaymentPlans(ctx context.Context, state loanmodel.LoanState, isDemo bool, offset, limit int) ([]*PaymentPlanSummary, int64, error) {
	loans, total, err := s.loans.ListPaymentPlans(ctx, state, isDemo, offset, limit)
	if err != nil {
		return nil, 0, err
	}

	summaries := make([]*PaymentPlanSummary, len(loans))
	for i, loan := range loans {
		var totalPaid float64
		var paidCount int
		var nextDueDate *time.Time

		for _, sch := range loan.Schedules {
			totalPaid += sch.AmountPaid
			if sch.Status == "PAID" {
				paidCount++
			} else if nextDueDate == nil || sch.DueDate.Before(*nextDueDate) {
				d := sch.DueDate
				nextDueDate = &d
			}
		}

		remaining := loan.TotalAmount - totalPaid
		if remaining < 0 {
			remaining = 0
		}

		var progress float64
		if loan.TotalAmount > 0 {
			progress = (totalPaid / loan.TotalAmount) * 100
		}

		summaries[i] = &PaymentPlanSummary{
			Loan:            loan,
			TotalPaid:       totalPaid,
			Remaining:       remaining,
			ProgressPercent: progress,
			NextDueDate:     nextDueDate,
			ScheduleCount:   len(loan.Schedules),
			PaidCount:       paidCount,
		}
	}

	return summaries, total, nil
}

// ListByState returns paginated loans filtered by state and demo mode.
func (s *LoanService) ListByState(ctx context.Context, state loanmodel.LoanState, isDemo bool, offset, limit int) ([]*loanmodel.Loan, int64, error) {
	return s.loans.ListByState(ctx, state, isDemo, offset, limit)
}

// RevertDisbursement manually reverts an ACTIVE loan back to APPROVED.
// Used when payout.completed was a false positive but money never arrived.
func (s *LoanService) RevertDisbursement(ctx context.Context, loanID uuid.UUID) error {
	loan, err := s.loans.FindByID(ctx, loanID)
	if err != nil {
		return apperrors.ErrNotFound("loan")
	}

	if loan.State != loanmodel.LoanStateActive {
		return apperrors.ErrInvalidLoanState(string(loan.State), string(loanmodel.LoanStateApproved))
	}

	if err := loan.Transition(loanmodel.LoanStateApproved); err != nil {
		return err
	}

	loan.MonimeReference = ""
	loan.PayoutID = ""
	loan.ProviderRef = ""
	loan.FailureReason = "Manually reverted — payout did not complete"
	loan.DueDate = nil

	if err := s.loans.Update(ctx, loan); err != nil {
		return apperrors.ErrInternalServer(err)
	}

	_ = s.loans.DeleteSchedulesByLoanID(ctx, loan.ID)

	_ = s.audit.Log(ctx, &shared.AuditLog{
		Action:     "LOAN_DISBURSEMENT_REVERTED",
		Resource:   "loan",
		ResourceID: loan.ID.String(),
		OldState:   string(loanmodel.LoanStateActive),
		NewState:   string(loanmodel.LoanStateApproved),
	})

	return nil
}
