package application_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	loanapp  "github.com/marketpay/backend/internal/loan/application"
	loanmodel "github.com/marketpay/backend/internal/loan/domain/model"
	shared    "github.com/marketpay/backend/internal/shared/domain/model"
	"github.com/marketpay/backend/pkg/config"
	"github.com/marketpay/backend/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// ── Mocks ─────────────────────────────────────────────────────────────────────

type mockLoanRepo struct{ mock.Mock }

func (m *mockLoanRepo) Create(ctx context.Context, loan *loanmodel.Loan) error {
	args := m.Called(ctx, loan)
	loan.ID = uuid.New() // simulate DB-assigned ID
	return args.Error(0)
}
func (m *mockLoanRepo) FindByID(ctx context.Context, id uuid.UUID) (*loanmodel.Loan, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil { return nil, args.Error(1) }
	return args.Get(0).(*loanmodel.Loan), args.Error(1)
}
func (m *mockLoanRepo) FindByVendorID(ctx context.Context, vendorID uuid.UUID, offset, limit int) ([]*loanmodel.Loan, int64, error) {
	args := m.Called(ctx, vendorID, offset, limit)
	return args.Get(0).([]*loanmodel.Loan), args.Get(1).(int64), args.Error(2)
}
func (m *mockLoanRepo) Update(ctx context.Context, loan *loanmodel.Loan) error {
	return m.Called(ctx, loan).Error(0)
}
func (m *mockLoanRepo) SaveSchedules(ctx context.Context, schedules []loanmodel.RepaymentSchedule) error {
	return m.Called(ctx, schedules).Error(0)
}
func (m *mockLoanRepo) FindSchedulesByLoanID(ctx context.Context, loanID uuid.UUID) ([]loanmodel.RepaymentSchedule, error) {
	args := m.Called(ctx, loanID)
	return args.Get(0).([]loanmodel.RepaymentSchedule), args.Error(1)
}
func (m *mockLoanRepo) UpdateSchedule(ctx context.Context, s *loanmodel.RepaymentSchedule) error {
	return m.Called(ctx, s).Error(0)
}
func (m *mockLoanRepo) ListByState(ctx context.Context, state loanmodel.LoanState, isDemo bool, offset, limit int) ([]*loanmodel.Loan, int64, error) {
	args := m.Called(ctx, state, isDemo, offset, limit)
	return args.Get(0).([]*loanmodel.Loan), args.Get(1).(int64), args.Error(2)
}
func (m *mockLoanRepo) FindByMonimeReference(ctx context.Context, ref string) *loanmodel.Loan {
	args := m.Called(ctx, ref)
	if args.Get(0) == nil { return nil }
	return args.Get(0).(*loanmodel.Loan)
}

type mockAuditRepo struct{ mock.Mock }

func (m *mockAuditRepo) Log(ctx context.Context, log *shared.AuditLog) error {
	return m.Called(ctx, log).Error(0)
}

type mockEventPublisher struct{ mock.Mock }

func (m *mockEventPublisher) Publish(ctx context.Context, eventType, aggregateID string, payload interface{}) error {
	return m.Called(ctx, eventType, aggregateID, payload).Error(0)
}

type mockCreditScoreService struct{ mock.Mock }

func (m *mockCreditScoreService) GetScore(ctx context.Context, vendorID uuid.UUID) (float64, bool, error) {
	args := m.Called(ctx, vendorID)
	return args.Get(0).(float64), args.Get(1).(bool), args.Error(2)
}

type mockEligibilityChecker struct{ mock.Mock }

func (m *mockEligibilityChecker) CheckEligibility(ctx context.Context, vendorID uuid.UUID) error {
	return m.Called(ctx, vendorID).Error(0)
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func newTestLoanService(
	repo loanapp.LoanRepository,
	audit loanapp.AuditRepository,
	events loanapp.EventPublisher,
	score loanapp.CreditScoreService,
	eligibility loanapp.VendorEligibilityChecker,
) *loanapp.LoanService {
	cfg := config.LoanProductsConfig{
		EmergencyAdvance: config.LoanProductConfig{MinAmount: 50, MaxAmount: 500, InterestRate: 0.05, AutoApprove: true},
		StarterLoan:      config.LoanProductConfig{MinAmount: 500, MaxAmount: 2000, InterestRate: 0.08},
		GrowthLoan:       config.LoanProductConfig{MinAmount: 2000, MaxAmount: 5000, InterestRate: 0.10},
	}
	scoreCfg := config.CreditScoreConfig{MinScore: 50, AutoApproveScore: 75}
	log := logger.NewNop()
	return loanapp.NewLoanService(repo, audit, events, score, eligibility, nil, nil, cfg, scoreCfg, log)
}

// ── Tests ─────────────────────────────────────────────────────────────────────

func TestLoanService_Apply_Success_AutoApprove(t *testing.T) {
	repo      := &mockLoanRepo{}
	audit     := &mockAuditRepo{}
	events    := &mockEventPublisher{}
	score     := &mockCreditScoreService{}
	eligible  := &mockEligibilityChecker{}

	vendorID := uuid.New()
	ctx := context.Background()

	eligible.On("CheckEligibility", ctx, vendorID).Return(nil)
	score.On("GetScore", ctx, vendorID).Return(80.0, true, nil) // above 75 → auto-approve

	repo.On("Create", ctx, mock.AnythingOfType("*model.Loan")).Return(nil)
	repo.On("Update", ctx, mock.AnythingOfType("*model.Loan")).Return(nil)
	audit.On("Log", ctx, mock.Anything).Return(nil)
	events.On("Publish", ctx, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	svc := newTestLoanService(repo, audit, events, score, eligible)

	loan, err := svc.Apply(ctx, loanapp.ApplyInput{
		VendorID:  vendorID,
		LoanType:  loanmodel.LoanTypeEmergencyAdvance,
		Amount:    200,
		TermWeeks: 2,
		Frequency: loanmodel.RepaymentFrequencyBiweekly,
		FundedBy:  loanmodel.FundingSourceMFIPartner,
	})

	require.NoError(t, err)
	assert.Equal(t, loanmodel.LoanStateAutoApproved, loan.State)
	assert.Equal(t, 200.0, loan.PrincipalAmount)
	assert.Equal(t, 210.0, loan.TotalAmount) // 200 * 1.05
	repo.AssertExpectations(t)
}

func TestLoanService_Apply_InsufficientScore(t *testing.T) {
	repo     := &mockLoanRepo{}
	audit    := &mockAuditRepo{}
	events   := &mockEventPublisher{}
	score    := &mockCreditScoreService{}
	eligible := &mockEligibilityChecker{}

	vendorID := uuid.New()
	ctx := context.Background()

	eligible.On("CheckEligibility", ctx, vendorID).Return(nil)
	score.On("GetScore", ctx, vendorID).Return(40.0, false, nil) // below min 50

	svc := newTestLoanService(repo, audit, events, score, eligible)

	_, err := svc.Apply(ctx, loanapp.ApplyInput{
		VendorID:  vendorID,
		LoanType:  loanmodel.LoanTypeEmergencyAdvance,
		Amount:    200,
		TermWeeks: 2,
		Frequency: loanmodel.RepaymentFrequencyBiweekly,
		FundedBy:  loanmodel.FundingSourceMFIPartner,
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "credit score")
	repo.AssertNotCalled(t, "Create")
}

func TestLoanService_Apply_AmountOutOfRange(t *testing.T) {
	repo     := &mockLoanRepo{}
	audit    := &mockAuditRepo{}
	events   := &mockEventPublisher{}
	score    := &mockCreditScoreService{}
	eligible := &mockEligibilityChecker{}

	vendorID := uuid.New()
	ctx := context.Background()

	eligible.On("CheckEligibility", ctx, vendorID).Return(nil)
	score.On("GetScore", ctx, vendorID).Return(80.0, true, nil)

	svc := newTestLoanService(repo, audit, events, score, eligible)

	_, err := svc.Apply(ctx, loanapp.ApplyInput{
		VendorID:  vendorID,
		LoanType:  loanmodel.LoanTypeEmergencyAdvance,
		Amount:    10000, // way above 500 max
		TermWeeks: 2,
		Frequency: loanmodel.RepaymentFrequencyBiweekly,
		FundedBy:  loanmodel.FundingSourceMFIPartner,
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "amount")
}

func TestLoanService_Apply_EligibilityFails(t *testing.T) {
	repo     := &mockLoanRepo{}
	audit    := &mockAuditRepo{}
	events   := &mockEventPublisher{}
	score    := &mockCreditScoreService{}
	eligible := &mockEligibilityChecker{}

	vendorID := uuid.New()
	ctx := context.Background()

	eligible.On("CheckEligibility", ctx, vendorID).Return(errors.New("vendor not eligible"))

	svc := newTestLoanService(repo, audit, events, score, eligible)

	_, err := svc.Apply(ctx, loanapp.ApplyInput{
		VendorID: vendorID,
		LoanType: loanmodel.LoanTypeEmergencyAdvance,
		Amount:   200,
	})

	assert.Error(t, err)
	repo.AssertNotCalled(t, "Create")
}

func TestLoanService_Approve_Success(t *testing.T) {
	repo     := &mockLoanRepo{}
	audit    := &mockAuditRepo{}
	events   := &mockEventPublisher{}
	score    := &mockCreditScoreService{}
	eligible := &mockEligibilityChecker{}

	loanID   := uuid.New()
	officerID := uuid.New()
	ctx := context.Background()

	existingLoan := &loanmodel.Loan{
		State:           loanmodel.LoanStateUnderReview,
		PrincipalAmount: 500,
	}
	existingLoan.ID = loanID

	repo.On("FindByID", ctx, loanID).Return(existingLoan, nil)
	repo.On("Update", ctx, mock.AnythingOfType("*model.Loan")).Return(nil)
	audit.On("Log", ctx, mock.Anything).Return(nil)
	events.On("Publish", ctx, "LoanApproved", loanID.String(), mock.Anything).Return(nil)

	svc := newTestLoanService(repo, audit, events, score, eligible)

	loan, err := svc.Approve(ctx, loanID, officerID, "looks good")
	require.NoError(t, err)
	assert.Equal(t, loanmodel.LoanStateApproved, loan.State)
	assert.Equal(t, &officerID, loan.ReviewedBy)
}

func TestLoanService_Disburse_SetsDisbursementPending(t *testing.T) {
	repo     := &mockLoanRepo{}
	audit    := &mockAuditRepo{}
	events   := &mockEventPublisher{}
	score    := &mockCreditScoreService{}
	eligible := &mockEligibilityChecker{}

	loanID := uuid.New()
	ctx := context.Background()

	existingLoan := &loanmodel.Loan{
		State:           loanmodel.LoanStateApproved,
		PrincipalAmount: 1000,
		InterestRate:    0.05,
		InterestType:    loanmodel.InterestTypeFlat,
		TermWeeks:       4,
		Frequency:       loanmodel.RepaymentFrequencyBiweekly,
	}
	existingLoan.ID = loanID
	existingLoan.TotalAmount = existingLoan.CalculateTotalAmount()

	repo.On("FindByID", ctx, loanID).Return(existingLoan, nil)
	repo.On("Update", ctx, mock.AnythingOfType("*model.Loan")).Return(nil)
	audit.On("Log", ctx, mock.Anything).Return(nil)
	events.On("Publish", ctx, "LoanDisbursed", loanID.String(), mock.Anything).Return(nil)

	svc := newTestLoanService(repo, audit, events, score, eligible)

	loan, err := svc.Disburse(ctx, loanID, "MONIME-REF-001")
	require.NoError(t, err)
	assert.Equal(t, loanmodel.LoanStateDisbursementPending, loan.State)
	assert.NotNil(t, loan.DisbursedAt)
	assert.Equal(t, "MONIME-REF-001", loan.MonimeReference)
}
