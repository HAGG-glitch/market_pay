package application

import (
	"context"
	"time"

	"github.com/google/uuid"
	scoremodel "github.com/marketpay/backend/internal/creditscore/domain/model"
	apperrors "github.com/marketpay/backend/pkg/errors"
	"github.com/marketpay/backend/pkg/config"
	"github.com/marketpay/backend/pkg/logger"
)

// ScoreRepository persists credit scores.
type ScoreRepository interface {
	Save(ctx context.Context, score *scoremodel.CreditScore) error
	FindLatestByVendorID(ctx context.Context, vendorID uuid.UUID) (*scoremodel.CreditScore, error)
}

// FactorProvider retrieves raw scoring factors for a vendor.
type FactorProvider interface {
	GetFactors(ctx context.Context, vendorID uuid.UUID) (*scoremodel.CreditScoreFactors, error)
}

// Service computes and caches credit scores.
type Service struct {
	scores  ScoreRepository
	factors FactorProvider
	cfg     config.CreditScoreConfig
	log     *logger.Logger
}

// NewService constructs a credit score Service.
func NewService(scores ScoreRepository, factors FactorProvider, cfg config.CreditScoreConfig, log *logger.Logger) *Service {
	return &Service{scores: scores, factors: factors, cfg: cfg, log: log}
}

// Compute recalculates the credit score for a vendor.
func (s *Service) Compute(ctx context.Context, vendorID uuid.UUID) (*scoremodel.CreditScore, error) {
	factors, err := s.factors.GetFactors(ctx, vendorID)
	if err != nil {
		return nil, apperrors.ErrInternalServer(err)
	}

	weights := scoremodel.ScoreWeights{
		TransactionVolume:      s.cfg.TransactionVolumeWeight,
		TransactionConsistency: s.cfg.TransactionConsistencyWeight,
		RepaymentHistory:       s.cfg.RepaymentHistoryWeight,
		MarketAssociation:      s.cfg.MarketAssociationWeight,
		KYCCompleteness:        s.cfg.KYCCompletenessWeight,
		GroupBonus:             s.cfg.GroupBonus,
	}

	score := scoremodel.Calculate(*factors, weights, s.cfg.MinScore, s.cfg.AutoApproveScore)
	score.VendorID = vendorID

	if err := s.scores.Save(ctx, score); err != nil {
		return nil, apperrors.ErrInternalServer(err)
	}

	return score, nil
}

// GetScore returns the latest score for a vendor (cached or freshly computed).
// Returns (score, canAutoApprove, error).
func (s *Service) GetScore(ctx context.Context, vendorID uuid.UUID) (float64, bool, error) {
	existing, err := s.scores.FindLatestByVendorID(ctx, vendorID)
	if err == nil && existing != nil {
		// Use cached score if less than 24 hours old
		if time.Since(existing.CreatedAt) < 24*time.Hour {
			return existing.TotalScore, existing.CanAutoApprove, nil
		}
	}

	// Recompute
	score, err := s.Compute(ctx, vendorID)
	if err != nil {
		return 0, false, err
	}

	return score.TotalScore, score.CanAutoApprove, nil
}
