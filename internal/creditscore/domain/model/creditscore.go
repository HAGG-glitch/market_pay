package model

import (
	"github.com/google/uuid"
	shared "github.com/marketpay/backend/internal/shared/domain/model"
)

// CreditScore stores a vendor's scored breakdown.
type CreditScore struct {
	shared.BaseModel
	VendorID                    uuid.UUID `gorm:"type:uuid;not null;index" json:"vendor_id"`
	TotalScore                  float64   `gorm:"type:decimal(5,2);not null" json:"total_score"`
	TransactionVolumeScore      float64   `gorm:"type:decimal(5,2)" json:"transaction_volume_score"`
	TransactionConsistencyScore float64   `gorm:"type:decimal(5,2)" json:"transaction_consistency_score"`
	RepaymentHistoryScore       float64   `gorm:"type:decimal(5,2)" json:"repayment_history_score"`
	MarketAssociationScore      float64   `gorm:"type:decimal(5,2)" json:"market_association_score"`
	KYCCompletenessScore        float64   `gorm:"type:decimal(5,2)" json:"kyc_completeness_score"`
	GroupBonus                  float64   `gorm:"type:decimal(5,2)" json:"group_bonus"`
	IsEligible                  bool      `gorm:"not null" json:"is_eligible"`
	CanAutoApprove              bool      `gorm:"not null" json:"can_auto_approve"`
	Version                     int       `gorm:"default:1" json:"version"`
}

// CreditScoreFactors holds raw data used to compute the score.
type CreditScoreFactors struct {
	VendorID              uuid.UUID
	TotalTransactions     int
	TransactionVolumeSLE  float64
	ConsistencyDays       int // days with at least one transaction in last 30 days
	SuccessfulRepayments  int
	MissedRepayments      int
	HasMarketAssociation  bool
	IsKYCComplete         bool
	IsInActiveGroup       bool
}

// ScoreWeights defines the scoring weights (must sum to 100 + bonus).
type ScoreWeights struct {
	TransactionVolume     int // 30
	TransactionConsistency int // 20
	RepaymentHistory      int // 30
	MarketAssociation     int // 10
	KYCCompleteness       int // 10
	GroupBonus            int // 5
}

// DefaultWeights are the system defaults from business rules.
var DefaultWeights = ScoreWeights{
	TransactionVolume:      30,
	TransactionConsistency: 20,
	RepaymentHistory:       30,
	MarketAssociation:      10,
	KYCCompleteness:        10,
	GroupBonus:             5,
}

// Calculate computes the credit score from raw factors using weights.
func Calculate(factors CreditScoreFactors, weights ScoreWeights, minScore, autoApproveScore float64) *CreditScore {
	// Transaction volume score (0-30): scale based on total transactions
	volumeScore := float64(weights.TransactionVolume)
	if factors.TotalTransactions < 10 {
		volumeScore = float64(weights.TransactionVolume) * float64(factors.TotalTransactions) / 10.0
	}

	// Consistency score (0-20): days with transactions in last 30 days
	consistencyScore := float64(weights.TransactionConsistency) * float64(factors.ConsistencyDays) / 30.0
	if consistencyScore > float64(weights.TransactionConsistency) {
		consistencyScore = float64(weights.TransactionConsistency)
	}

	// Repayment history (0-30)
	repaymentScore := 0.0
	totalRepayments := factors.SuccessfulRepayments + factors.MissedRepayments
	if totalRepayments > 0 {
		repaymentScore = float64(weights.RepaymentHistory) *
			float64(factors.SuccessfulRepayments) / float64(totalRepayments)
	} else {
		// No history: neutral score
		repaymentScore = float64(weights.RepaymentHistory) * 0.5
	}

	// Market association (0-10)
	assocScore := 0.0
	if factors.HasMarketAssociation {
		assocScore = float64(weights.MarketAssociation)
	}

	// KYC completeness (0-10)
	kycScore := 0.0
	if factors.IsKYCComplete {
		kycScore = float64(weights.KYCCompleteness)
	}

	// Group bonus (0-5)
	groupBonus := 0.0
	if factors.IsInActiveGroup {
		groupBonus = float64(weights.GroupBonus)
	}

	totalScore := volumeScore + consistencyScore + repaymentScore + assocScore + kycScore + groupBonus

	return &CreditScore{
		VendorID:                    factors.VendorID,
		TotalScore:                  totalScore,
		TransactionVolumeScore:      volumeScore,
		TransactionConsistencyScore: consistencyScore,
		RepaymentHistoryScore:       repaymentScore,
		MarketAssociationScore:      assocScore,
		KYCCompletenessScore:        kycScore,
		GroupBonus:                  groupBonus,
		IsEligible:                  totalScore >= minScore,
		CanAutoApprove:              totalScore >= autoApproveScore,
	}
}
