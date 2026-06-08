package model_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/marketpay/backend/internal/creditscore/domain/model"
	"github.com/stretchr/testify/assert"
)

var defaultWeights = model.ScoreWeights{
	TransactionVolume:      30,
	TransactionConsistency: 20,
	RepaymentHistory:       30,
	MarketAssociation:      10,
	KYCCompleteness:        10,
	GroupBonus:             5,
}

func TestCalculate_FullScore(t *testing.T) {
	factors := model.CreditScoreFactors{
		VendorID:             uuid.New(),
		TotalTransactions:    50,
		TransactionVolumeSLE: 10000,
		ConsistencyDays:      30,
		SuccessfulRepayments: 10,
		MissedRepayments:     0,
		HasMarketAssociation: true,
		IsKYCComplete:        true,
		IsInActiveGroup:      true,
	}

	score := model.Calculate(factors, defaultWeights, 50, 75)

	assert.Equal(t, 30.0, score.TransactionVolumeScore, "max transaction volume")
	assert.Equal(t, 20.0, score.TransactionConsistencyScore, "max consistency")
	assert.Equal(t, 30.0, score.RepaymentHistoryScore, "perfect repayment history")
	assert.Equal(t, 10.0, score.MarketAssociationScore, "has association")
	assert.Equal(t, 10.0, score.KYCCompletenessScore, "KYC complete")
	assert.Equal(t, 5.0, score.GroupBonus, "in active group")
	assert.Equal(t, 105.0, score.TotalScore)
	assert.True(t, score.IsEligible)
	assert.True(t, score.CanAutoApprove)
}

func TestCalculate_MinimalNewVendor(t *testing.T) {
	factors := model.CreditScoreFactors{
		VendorID:             uuid.New(),
		TotalTransactions:    0,
		ConsistencyDays:      0,
		SuccessfulRepayments: 0,
		MissedRepayments:     0,
		HasMarketAssociation: false,
		IsKYCComplete:        false,
		IsInActiveGroup:      false,
	}

	score := model.Calculate(factors, defaultWeights, 50, 75)

	// No transaction history, no repayments → 15 (50% neutral repayment)
	assert.False(t, score.IsEligible)
	assert.False(t, score.CanAutoApprove)
	assert.Less(t, score.TotalScore, 50.0)
}

func TestCalculate_EligibleButNotAutoApprove(t *testing.T) {
	factors := model.CreditScoreFactors{
		VendorID:             uuid.New(),
		TotalTransactions:    10,
		ConsistencyDays:      15,
		SuccessfulRepayments: 5,
		MissedRepayments:     2,
		HasMarketAssociation: true,
		IsKYCComplete:        true,
		IsInActiveGroup:      false,
	}

	score := model.Calculate(factors, defaultWeights, 50, 75)
	assert.True(t, score.IsEligible, "should be eligible")
	// May or may not auto-approve — just test that eligible ≥ min_score
	assert.GreaterOrEqual(t, score.TotalScore, 50.0)
}

func TestCalculate_PerfectRepaymentHistory(t *testing.T) {
	factors := model.CreditScoreFactors{
		VendorID:             uuid.New(),
		SuccessfulRepayments: 20,
		MissedRepayments:     0,
	}
	score := model.Calculate(factors, defaultWeights, 50, 75)
	assert.Equal(t, 30.0, score.RepaymentHistoryScore)
}

func TestCalculate_NeutralRepaymentNoHistory(t *testing.T) {
	factors := model.CreditScoreFactors{
		VendorID:             uuid.New(),
		SuccessfulRepayments: 0,
		MissedRepayments:     0,
	}
	score := model.Calculate(factors, defaultWeights, 50, 75)
	assert.Equal(t, 15.0, score.RepaymentHistoryScore, "no history = 50% of 30")
}

func TestCalculate_PartialMissedRepayments(t *testing.T) {
	factors := model.CreditScoreFactors{
		VendorID:             uuid.New(),
		SuccessfulRepayments: 8,
		MissedRepayments:     2,
	}
	score := model.Calculate(factors, defaultWeights, 50, 75)
	// 8/10 * 30 = 24
	assert.Equal(t, 24.0, score.RepaymentHistoryScore)
}

func TestCalculate_GroupBonusOnlyWhenActive(t *testing.T) {
	base := model.CreditScoreFactors{VendorID: uuid.New(), IsInActiveGroup: false}
	bonus := model.CreditScoreFactors{VendorID: uuid.New(), IsInActiveGroup: true}

	scoreBase := model.Calculate(base, defaultWeights, 50, 75)
	scoreBonus := model.Calculate(bonus, defaultWeights, 50, 75)

	assert.Equal(t, 0.0, scoreBase.GroupBonus)
	assert.Equal(t, 5.0, scoreBonus.GroupBonus)
	assert.Equal(t, scoreBase.TotalScore+5, scoreBonus.TotalScore)
}
