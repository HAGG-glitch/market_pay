package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	apperrors "github.com/marketpay/backend/pkg/errors"
	shared "github.com/marketpay/backend/internal/shared/domain/model"
)

// AccountType defines the ledger account categories.
type AccountType string

const (
	AccountLoanReceivable   AccountType = "LOAN_RECEIVABLE"
	AccountPartnerLiability AccountType = "PARTNER_LIABILITY"
	AccountInterestIncome   AccountType = "INTEREST_INCOME"
	AccountPenaltyIncome    AccountType = "PENALTY_INCOME"
	AccountCommissionIncome AccountType = "COMMISSION_INCOME"
	AccountTransactionFee   AccountType = "TRANSACTION_FEE_INCOME"
	AccountMonimeFloat      AccountType = "MONIME_FLOAT"
)

// EntryType is debit or credit.
type EntryType string

const (
	EntryTypeDebit  EntryType = "DEBIT"
	EntryTypeCredit EntryType = "CREDIT"
)

// Account represents a ledger account.
type Account struct {
	shared.BaseModel
	Type        AccountType `gorm:"type:varchar(50);not null;uniqueIndex" json:"type"`
	Name        string      `gorm:"type:varchar(255);not null" json:"name"`
	Balance     float64     `gorm:"type:decimal(20,2);default:0" json:"balance"`
	Currency    string      `gorm:"type:varchar(10);not null;default:'SLE'" json:"currency"`
	Description string      `gorm:"type:text" json:"description"`
}

// JournalEntry is a balanced double-entry transaction.
type JournalEntry struct {
	shared.BaseModel
	Reference   string         `gorm:"type:varchar(255);not null;index" json:"reference"`
	Description string         `gorm:"type:text;not null" json:"description"`
	EntryDate   time.Time      `gorm:"not null" json:"entry_date"`
	PostedBy    uuid.UUID      `gorm:"type:uuid" json:"posted_by"`
	Lines       []JournalLine  `gorm:"foreignKey:JournalEntryID" json:"lines"`
	IsPosted    bool           `gorm:"default:false" json:"is_posted"`
}

// JournalLine is one side of a journal entry (debit or credit).
type JournalLine struct {
	shared.BaseModel
	JournalEntryID uuid.UUID   `gorm:"type:uuid;not null;index" json:"journal_entry_id"`
	AccountID      uuid.UUID   `gorm:"type:uuid;not null;index" json:"account_id"`
	Account        *Account    `gorm:"foreignKey:AccountID" json:"account,omitempty"`
	EntryType      EntryType   `gorm:"type:varchar(10);not null" json:"entry_type"`
	Amount         float64     `gorm:"type:decimal(15,2);not null" json:"amount"`
	Description    string      `gorm:"type:text" json:"description"`
}

// Validate ensures the journal entry is balanced (debits == credits).
func (j *JournalEntry) Validate() error {
	var totalDebits, totalCredits float64
	for _, line := range j.Lines {
		switch line.EntryType {
		case EntryTypeDebit:
			totalDebits += line.Amount
		case EntryTypeCredit:
			totalCredits += line.Amount
		}
	}

	// Allow for floating point rounding
	diff := totalDebits - totalCredits
	if diff < 0 {
		diff = -diff
	}
	if diff > 0.01 {
		return fmt.Errorf("%w: debits=%.2f credits=%.2f",
			apperrors.ErrLedgerUnbalanced, totalDebits, totalCredits)
	}
	return nil
}

// TotalDebits returns the sum of all debit lines.
func (j *JournalEntry) TotalDebits() float64 {
	total := 0.0
	for _, line := range j.Lines {
		if line.EntryType == EntryTypeDebit {
			total += line.Amount
		}
	}
	return total
}

// TotalCredits returns the sum of all credit lines.
func (j *JournalEntry) TotalCredits() float64 {
	total := 0.0
	for _, line := range j.Lines {
		if line.EntryType == EntryTypeCredit {
			total += line.Amount
		}
	}
	return total
}
