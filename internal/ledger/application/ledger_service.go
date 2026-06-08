package application

import (
	"context"
	"time"

	"github.com/google/uuid"
	ledgermodel "github.com/marketpay/backend/internal/ledger/domain/model"
	apperrors "github.com/marketpay/backend/pkg/errors"
	"github.com/marketpay/backend/pkg/logger"
)

// LedgerRepository defines ledger persistence.
type LedgerRepository interface {
	FindAccount(ctx context.Context, accountType ledgermodel.AccountType) (*ledgermodel.Account, error)
	CreateJournalEntry(ctx context.Context, entry *ledgermodel.JournalEntry) error
	UpdateAccountBalance(ctx context.Context, accountID uuid.UUID, delta float64) error
	ListEntries(ctx context.Context, offset, limit int) ([]*ledgermodel.JournalEntry, int64, error)
}

// LedgerService handles double-entry bookkeeping.
type LedgerService struct {
	repo LedgerRepository
	log  *logger.Logger
}

func NewLedgerService(repo LedgerRepository, log *logger.Logger) *LedgerService {
	return &LedgerService{repo: repo, log: log}
}

// RecordDisbursement posts journal entries for a loan disbursement.
// DR Loan Receivable / CR Partner Liability
func (s *LedgerService) RecordDisbursement(ctx context.Context, loanID uuid.UUID, amount float64, partnerID uuid.UUID) error {
	loanReceivable, err := s.repo.FindAccount(ctx, ledgermodel.AccountLoanReceivable)
	if err != nil {
		return apperrors.ErrInternalServer(err)
	}
	partnerLiability, err := s.repo.FindAccount(ctx, ledgermodel.AccountPartnerLiability)
	if err != nil {
		return apperrors.ErrInternalServer(err)
	}

	entry := &ledgermodel.JournalEntry{
		Reference:   "DISB-" + loanID.String(),
		Description: "Loan disbursement",
		EntryDate:   time.Now(),
		Lines: []ledgermodel.JournalLine{
			{AccountID: loanReceivable.ID, EntryType: ledgermodel.EntryTypeDebit, Amount: amount, Description: "Loan receivable"},
			{AccountID: partnerLiability.ID, EntryType: ledgermodel.EntryTypeCredit, Amount: amount, Description: "Partner liability"},
		},
		IsPosted: true,
	}

	if err := entry.Validate(); err != nil {
		return err
	}

	return s.repo.CreateJournalEntry(ctx, entry)
}

// RecordRepayment posts journal entries for a loan repayment.
// DR Monime Float / CR Loan Receivable + CR Interest Income
func (s *LedgerService) RecordRepayment(ctx context.Context, loanID uuid.UUID, principal, interest float64) error {
	monimeFloat, _ := s.repo.FindAccount(ctx, ledgermodel.AccountMonimeFloat)
	loanReceivable, _ := s.repo.FindAccount(ctx, ledgermodel.AccountLoanReceivable)
	interestIncome, _ := s.repo.FindAccount(ctx, ledgermodel.AccountInterestIncome)

	entry := &ledgermodel.JournalEntry{
		Reference:   "REPAY-" + loanID.String() + "-" + time.Now().Format("20060102150405"),
		Description: "Loan repayment",
		EntryDate:   time.Now(),
		Lines: []ledgermodel.JournalLine{
			{AccountID: monimeFloat.ID, EntryType: ledgermodel.EntryTypeDebit, Amount: principal + interest},
			{AccountID: loanReceivable.ID, EntryType: ledgermodel.EntryTypeCredit, Amount: principal},
			{AccountID: interestIncome.ID, EntryType: ledgermodel.EntryTypeCredit, Amount: interest},
		},
		IsPosted: true,
	}

	if err := entry.Validate(); err != nil {
		return err
	}

	return s.repo.CreateJournalEntry(ctx, entry)
}

// RecordPaymentFee records the 1% transaction fee income.
// DR Monime Float / CR Transaction Fee Income
func (s *LedgerService) RecordPaymentFee(ctx context.Context, paymentID uuid.UUID, fee float64) error {
	monimeFloat, _ := s.repo.FindAccount(ctx, ledgermodel.AccountMonimeFloat)
	feeIncome, _ := s.repo.FindAccount(ctx, ledgermodel.AccountTransactionFee)

	entry := &ledgermodel.JournalEntry{
		Reference:   "FEE-" + paymentID.String(),
		Description: "Payment transaction fee",
		EntryDate:   time.Now(),
		Lines: []ledgermodel.JournalLine{
			{AccountID: monimeFloat.ID, EntryType: ledgermodel.EntryTypeDebit, Amount: fee},
			{AccountID: feeIncome.ID, EntryType: ledgermodel.EntryTypeCredit, Amount: fee},
		},
		IsPosted: true,
	}

	if err := entry.Validate(); err != nil {
		return err
	}

	return s.repo.CreateJournalEntry(ctx, entry)
}

// ListEntries returns paginated journal entries.
func (s *LedgerService) ListEntries(ctx context.Context, offset, limit int) ([]*ledgermodel.JournalEntry, int64, error) {
	return s.repo.ListEntries(ctx, offset, limit)
}
