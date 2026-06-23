package application

import (
	"context"

	"github.com/google/uuid"
	loanmodel "github.com/marketpay/backend/internal/loan/domain/model"
	apperrors "github.com/marketpay/backend/pkg/errors"
	shared "github.com/marketpay/backend/internal/shared/domain/model"
)

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
