package model_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/marketpay/backend/internal/ledger/domain/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"time"
)

func makeAccount(t model.AccountType) model.Account {
	return model.Account{Type: t, Currency: "SLE"}
}

func TestJournalEntry_Validate_Balanced(t *testing.T) {
	dr := uuid.New()
	cr := uuid.New()

	entry := &model.JournalEntry{
		Reference:   "TEST-001",
		Description: "Test entry",
		EntryDate:   time.Now(),
		Lines: []model.JournalLine{
			{AccountID: dr, EntryType: model.EntryTypeDebit, Amount: 1000},
			{AccountID: cr, EntryType: model.EntryTypeCredit, Amount: 1000},
		},
	}

	err := entry.Validate()
	require.NoError(t, err)
}

func TestJournalEntry_Validate_Unbalanced(t *testing.T) {
	dr := uuid.New()
	cr := uuid.New()

	entry := &model.JournalEntry{
		Reference: "TEST-002",
		Lines: []model.JournalLine{
			{AccountID: dr, EntryType: model.EntryTypeDebit, Amount: 1000},
			{AccountID: cr, EntryType: model.EntryTypeCredit, Amount: 999},
		},
	}

	err := entry.Validate()
	assert.Error(t, err, "unbalanced entry should fail")
	assert.Contains(t, err.Error(), "balance")
}

func TestJournalEntry_Validate_MultiLineBalanced(t *testing.T) {
	a1 := uuid.New()
	a2 := uuid.New()
	a3 := uuid.New()

	entry := &model.JournalEntry{
		Reference: "TEST-003",
		Lines: []model.JournalLine{
			{AccountID: a1, EntryType: model.EntryTypeDebit, Amount: 1000},
			{AccountID: a2, EntryType: model.EntryTypeCredit, Amount: 800},
			{AccountID: a3, EntryType: model.EntryTypeCredit, Amount: 200},
		},
	}

	err := entry.Validate()
	require.NoError(t, err)
}

func TestJournalEntry_TotalDebits(t *testing.T) {
	entry := &model.JournalEntry{
		Lines: []model.JournalLine{
			{EntryType: model.EntryTypeDebit, Amount: 500},
			{EntryType: model.EntryTypeDebit, Amount: 300},
			{EntryType: model.EntryTypeCredit, Amount: 800},
		},
	}
	assert.Equal(t, 800.0, entry.TotalDebits())
}

func TestJournalEntry_TotalCredits(t *testing.T) {
	entry := &model.JournalEntry{
		Lines: []model.JournalLine{
			{EntryType: model.EntryTypeDebit, Amount: 800},
			{EntryType: model.EntryTypeCredit, Amount: 500},
			{EntryType: model.EntryTypeCredit, Amount: 300},
		},
	}
	assert.Equal(t, 800.0, entry.TotalCredits())
}

func TestJournalEntry_Validate_FloatingPointTolerance(t *testing.T) {
	dr := uuid.New()
	cr := uuid.New()

	entry := &model.JournalEntry{
		Reference: "TEST-FLOAT",
		Lines: []model.JournalLine{
			{AccountID: dr, EntryType: model.EntryTypeDebit, Amount: 100.001},
			{AccountID: cr, EntryType: model.EntryTypeCredit, Amount: 100.000},
		},
	}
	// Within 0.01 tolerance
	err := entry.Validate()
	require.NoError(t, err)
}

func TestJournalEntry_Validate_EmptyLines(t *testing.T) {
	entry := &model.JournalEntry{
		Reference: "EMPTY",
		Lines:     []model.JournalLine{},
	}
	// 0 == 0, should pass
	err := entry.Validate()
	require.NoError(t, err)
}

func TestAccount_Types(t *testing.T) {
	types := []model.AccountType{
		model.AccountLoanReceivable,
		model.AccountPartnerLiability,
		model.AccountInterestIncome,
		model.AccountPenaltyIncome,
		model.AccountCommissionIncome,
		model.AccountTransactionFee,
		model.AccountMonimeFloat,
	}
	assert.Len(t, types, 7, "should have 7 ledger account types")
}
