package postgres

import (
	"context"

	"github.com/google/uuid"
	scoremodel  "github.com/marketpay/backend/internal/creditscore/domain/model"
	shared      "github.com/marketpay/backend/internal/shared/domain/model"
	vendormodel "github.com/marketpay/backend/internal/vendors/domain/model"
	"gorm.io/gorm"
)

// ── ScoreRepo ────────────────────────────────────────────────────────────────

// ScoreRepo implements ScoreRepository.
type ScoreRepo struct {
	db *gorm.DB
}

func NewScoreRepo(db *gorm.DB) *ScoreRepo {
	return &ScoreRepo{db: db}
}

func (r *ScoreRepo) Save(ctx context.Context, score *scoremodel.CreditScore) error {
	return r.db.WithContext(ctx).Create(score).Error
}

func (r *ScoreRepo) FindLatestByVendorID(ctx context.Context, vendorID uuid.UUID) (*scoremodel.CreditScore, error) {
	var score scoremodel.CreditScore
	err := r.db.WithContext(ctx).
		Where("vendor_id = ?", vendorID).
		Order("created_at DESC").
		First(&score).Error
	if err != nil {
		return nil, err
	}
	return &score, nil
}

// ── FactorRepo ───────────────────────────────────────────────────────────────

// FactorRepo computes scoring factors from live DB data.
type FactorRepo struct {
	db *gorm.DB
}

func NewFactorRepo(db *gorm.DB) *FactorRepo {
	return &FactorRepo{db: db}
}

// GetFactors implements FactorProvider. It queries multiple tables to assemble
// the raw inputs needed by the credit score calculator.
func (r *FactorRepo) GetFactors(ctx context.Context, vendorID uuid.UUID) (*scoremodel.CreditScoreFactors, error) {
	var vendor vendormodel.Vendor
	if err := r.db.WithContext(ctx).First(&vendor, "id = ?", vendorID).Error; err != nil {
		return nil, err
	}

	// Transaction count (completed payments received) in last 90 days
	var txCount int64
	r.db.WithContext(ctx).
		Table("payments").
		Where("vendor_id = ? AND status = 'COMPLETED' AND created_at >= NOW() - INTERVAL '90 days'", vendorID).
		Count(&txCount)

	// Consistency: distinct days with ≥1 completed payment in last 30 days
	var consistencyDays int64
	r.db.WithContext(ctx).
		Raw(`SELECT COUNT(DISTINCT DATE(created_at))
		     FROM payments
		     WHERE vendor_id = ? AND status = 'COMPLETED'
		     AND created_at >= NOW() - INTERVAL '30 days'`, vendorID).
		Scan(&consistencyDays)

	// Repayment history
	var successfulRepayments int64
	r.db.WithContext(ctx).
		Table("repayment_schedules").
		Joins("JOIN loans ON loans.id = repayment_schedules.loan_id").
		Where("loans.vendor_id = ? AND repayment_schedules.status = 'PAID'", vendorID).
		Count(&successfulRepayments)

	var missedRepayments int64
	r.db.WithContext(ctx).
		Table("repayment_schedules").
		Joins("JOIN loans ON loans.id = repayment_schedules.loan_id").
		Where("loans.vendor_id = ? AND repayment_schedules.status = 'PENDING' AND repayment_schedules.due_date < NOW()", vendorID).
		Count(&missedRepayments)

	// Active group membership
	var groupCount int64
	r.db.WithContext(ctx).
		Table("group_members").
		Joins("JOIN groups ON groups.id = group_members.group_id").
		Where("group_members.vendor_id = ? AND groups.status = 'ACTIVE' AND group_members.deleted_at IS NULL", vendorID).
		Count(&groupCount)

	// Total payment volume
	var totalVolume float64
	r.db.WithContext(ctx).
		Table("payments").
		Select("COALESCE(SUM(net_amount), 0)").
		Where("vendor_id = ? AND status = 'COMPLETED'", vendorID).
		Scan(&totalVolume)

	return &scoremodel.CreditScoreFactors{
		VendorID:              vendorID,
		TotalTransactions:     int(txCount),
		TransactionVolumeSLE:  totalVolume,
		ConsistencyDays:       int(consistencyDays),
		SuccessfulRepayments:  int(successfulRepayments),
		MissedRepayments:      int(missedRepayments),
		HasMarketAssociation:  vendor.MarketAssociationID != uuid.Nil,
		IsKYCComplete:         vendor.KYCStatus == shared.KYCVerified,
		IsInActiveGroup:       groupCount > 0,
	}, nil
}

// ── AuditRepo ────────────────────────────────────────────────────────────────

// AuditRepo handles audit log persistence.
type AuditRepo struct {
	db *gorm.DB
}

func NewAuditRepo(db *gorm.DB) *AuditRepo {
	return &AuditRepo{db: db}
}

func (r *AuditRepo) Log(ctx context.Context, log *shared.AuditLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}
