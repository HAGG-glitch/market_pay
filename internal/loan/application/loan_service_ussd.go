package application

import (
	"context"

	"github.com/google/uuid"
	loanmodel "github.com/marketpay/backend/internal/loan/domain/model"
)

// USSDApplyInput holds loan application from USSD exchange.
type USSDApplyInput struct {
	VendorID  uuid.UUID
	LoanType  loanmodel.LoanType
	Amount    float64
	TermWeeks int
	Frequency loanmodel.RepaymentFrequency
	FundedBy  loanmodel.FundingSource
	IsDemo    bool
}

// ApplyFromUSSD creates a loan application from USSD flow.
func (s *LoanService) ApplyFromUSSD(ctx context.Context, input USSDApplyInput) (*loanmodel.Loan, error) {
	loan, err := s.Apply(ctx, ApplyInput{
		VendorID:  input.VendorID,
		LoanType:  input.LoanType,
		Amount:    input.Amount,
		TermWeeks: input.TermWeeks,
		Frequency: input.Frequency,
		FundedBy:  input.FundedBy,
	})
	if err != nil {
		return nil, err
	}
	if input.IsDemo {
		loan.IsDemo = true
		_ = s.loans.Update(ctx, loan)
	}
	return loan, nil
}
