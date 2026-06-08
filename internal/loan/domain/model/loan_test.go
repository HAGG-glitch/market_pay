package model_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/marketpay/backend/internal/loan/domain/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoan_Transition_ValidPaths(t *testing.T) {
	tests := []struct {
		name string
		from model.LoanState
		to   model.LoanState
	}{
		{"draft to pending", model.LoanStateDraft, model.LoanStatePendingReview},
		{"pending to auto approved", model.LoanStatePendingReview, model.LoanStateAutoApproved},
		{"pending to under review", model.LoanStatePendingReview, model.LoanStateUnderReview},
		{"pending to rejected", model.LoanStatePendingReview, model.LoanStateRejected},
		{"under review to approved", model.LoanStateUnderReview, model.LoanStateApproved},
		{"under review to rejected", model.LoanStateUnderReview, model.LoanStateRejected},
		{"approved to disbursed", model.LoanStateApproved, model.LoanStateDisbursed},
		{"auto approved to disbursed", model.LoanStateAutoApproved, model.LoanStateDisbursed},
		{"disbursed to active", model.LoanStateDisbursed, model.LoanStateActive},
		{"active to closed", model.LoanStateActive, model.LoanStateClosed},
		{"active to defaulted", model.LoanStateActive, model.LoanStateDefaulted},
		{"defaulted to closed", model.LoanStateDefaulted, model.LoanStateClosed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loan := &model.Loan{State: tt.from}
			err := loan.Transition(tt.to)
			require.NoError(t, err)
			assert.Equal(t, tt.to, loan.State)
		})
	}
}

func TestLoan_Transition_InvalidPaths(t *testing.T) {
	tests := []struct {
		name string
		from model.LoanState
		to   model.LoanState
	}{
		{"draft to approved", model.LoanStateDraft, model.LoanStateApproved},
		{"closed to active", model.LoanStateClosed, model.LoanStateActive},
		{"rejected to approved", model.LoanStateRejected, model.LoanStateApproved},
		{"active to pending", model.LoanStateActive, model.LoanStatePendingReview},
		{"draft to disbursed", model.LoanStateDraft, model.LoanStateDisbursed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loan := &model.Loan{State: tt.from}
			err := loan.Transition(tt.to)
			assert.Error(t, err, "expected transition %s -> %s to fail", tt.from, tt.to)
			assert.Equal(t, tt.from, loan.State, "state should not change on invalid transition")
		})
	}
}

func TestLoan_CalculateTotalAmount_FlatRate(t *testing.T) {
	loan := &model.Loan{
		PrincipalAmount: 1000,
		InterestRate:    0.05,
		InterestType:    model.InterestTypeFlat,
	}
	total := loan.CalculateTotalAmount()
	assert.Equal(t, 1050.0, total)
}

func TestLoan_CalculateTotalAmount_EmergencyAdvance(t *testing.T) {
	loan := &model.Loan{
		PrincipalAmount: 500,
		InterestRate:    0.05, // 5% flat
		InterestType:    model.InterestTypeFlat,
	}
	total := loan.CalculateTotalAmount()
	assert.Equal(t, 525.0, total)
}

func TestLoan_GenerateSchedule_Biweekly(t *testing.T) {
	now := time.Now()
	loan := &model.Loan{
		PrincipalAmount: 1000,
		InterestRate:    0.08,
		InterestType:    model.InterestTypeFlat,
		TermWeeks:       8,
		Frequency:       model.RepaymentFrequencyBiweekly,
		DisbursedAt:     &now,
	}
	loan.TotalAmount = loan.CalculateTotalAmount()

	schedules := loan.GenerateSchedule()
	assert.Len(t, schedules, 4, "8 weeks biweekly = 4 installments")

	for i, s := range schedules {
		assert.Equal(t, i+1, s.InstallmentNo)
		assert.Equal(t, "PENDING", s.Status)
		assert.Greater(t, s.TotalDue, 0.0)
	}
}

func TestLoan_GenerateSchedule_Monthly(t *testing.T) {
	now := time.Now()
	loan := &model.Loan{
		PrincipalAmount: 2000,
		InterestRate:    0.08,
		InterestType:    model.InterestTypeFlat,
		TermWeeks:       8,
		Frequency:       model.RepaymentFrequencyMonthly,
		DisbursedAt:     &now,
	}
	loan.TotalAmount = loan.CalculateTotalAmount()

	schedules := loan.GenerateSchedule()
	assert.Len(t, schedules, 2, "8 weeks monthly = 2 installments")
}

func TestLoan_IsOverdue(t *testing.T) {
	past := time.Now().Add(-48 * time.Hour)
	future := time.Now().Add(48 * time.Hour)

	overdueActive := &model.Loan{State: model.LoanStateActive, DueDate: &past}
	assert.True(t, overdueActive.IsOverdue())

	notDue := &model.Loan{State: model.LoanStateActive, DueDate: &future}
	assert.False(t, notDue.IsOverdue())

	closedOverdue := &model.Loan{State: model.LoanStateClosed, DueDate: &past}
	assert.False(t, closedOverdue.IsOverdue())
}

func TestLoan_IsInGracePeriod(t *testing.T) {
	recentPast := time.Now().Add(-2 * 24 * time.Hour) // 2 days ago
	farPast := time.Now().Add(-10 * 24 * time.Hour)   // 10 days ago

	loan := &model.Loan{DueDate: &recentPast}
	assert.True(t, loan.IsInGracePeriod(7))

	loan2 := &model.Loan{DueDate: &farPast}
	assert.False(t, loan2.IsInGracePeriod(7))
}

func TestLoan_CanBeAutoApproved(t *testing.T) {
	emergencyLoan := &model.Loan{LoanType: model.LoanTypeEmergencyAdvance}
	assert.True(t, emergencyLoan.CanBeAutoApproved(75.0, 80.0))
	assert.False(t, emergencyLoan.CanBeAutoApproved(75.0, 70.0))

	starterLoan := &model.Loan{LoanType: model.LoanTypeStarterLoan}
	assert.False(t, starterLoan.CanBeAutoApproved(75.0, 80.0))
}

func TestLoan_IDs_NotNil(t *testing.T) {
	loan := &model.Loan{
		VendorID:        uuid.New(),
		PrincipalAmount: 100,
		InterestRate:    0.05,
		InterestType:    model.InterestTypeFlat,
		LoanType:        model.LoanTypeEmergencyAdvance,
		State:           model.LoanStateDraft,
	}
	assert.NotEqual(t, uuid.Nil, loan.VendorID)
}
