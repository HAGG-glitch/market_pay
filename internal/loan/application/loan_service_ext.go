package application

import (
	"context"

	loanmodel "github.com/marketpay/backend/internal/loan/domain/model"
)

// ListByState returns paginated loans filtered by state.
func (s *LoanService) ListByState(ctx context.Context, state loanmodel.LoanState, offset, limit int) ([]*loanmodel.Loan, int64, error) {
	return s.loans.ListByState(ctx, state, offset, limit)
}
