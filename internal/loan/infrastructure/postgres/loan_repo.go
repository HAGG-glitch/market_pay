package postgres

import (
	"context"

	"github.com/google/uuid"
	loanmodel "github.com/marketpay/backend/internal/loan/domain/model"
	"gorm.io/gorm"
)

// LoanRepo is the GORM implementation of LoanRepository.
type LoanRepo struct {
	db *gorm.DB
}

// NewLoanRepo constructs a LoanRepo.
func NewLoanRepo(db *gorm.DB) *LoanRepo {
	return &LoanRepo{db: db}
}

func (r *LoanRepo) Create(ctx context.Context, loan *loanmodel.Loan) error {
	return r.db.WithContext(ctx).Create(loan).Error
}

func (r *LoanRepo) FindByID(ctx context.Context, id uuid.UUID) (*loanmodel.Loan, error) {
	var loan loanmodel.Loan
	err := r.db.WithContext(ctx).
		Model(&loanmodel.Loan{}).
		Select("loans.*, COALESCE(CONCAT(vendors.first_name, ' ', vendors.last_name), '') as vendor_name").
		Joins("LEFT JOIN vendors ON vendors.id = loans.vendor_id").
		Where("loans.id = ?", id).
		Preload("Schedules").
		First(&loan).Error
	if err != nil {
		return nil, err
	}
	return &loan, nil
}

func (r *LoanRepo) FindByVendorID(ctx context.Context, vendorID uuid.UUID, offset, limit int) ([]*loanmodel.Loan, int64, error) {
	var loans []*loanmodel.Loan
	var count int64

	r.db.WithContext(ctx).Model(&loanmodel.Loan{}).Where("vendor_id = ?", vendorID).Count(&count)

	err := r.db.WithContext(ctx).
		Model(&loanmodel.Loan{}).
		Select("loans.*, COALESCE(CONCAT(vendors.first_name, ' ', vendors.last_name), '') as vendor_name").
		Joins("LEFT JOIN vendors ON vendors.id = loans.vendor_id").
		Where("loans.vendor_id = ?", vendorID).
		Order("loans.created_at DESC").
		Offset(offset).Limit(limit).
		Find(&loans).Error

	return loans, count, err
}

func (r *LoanRepo) Update(ctx context.Context, loan *loanmodel.Loan) error {
	return r.db.WithContext(ctx).Save(loan).Error
}

func (r *LoanRepo) SaveSchedules(ctx context.Context, schedules []loanmodel.RepaymentSchedule) error {
	return r.db.WithContext(ctx).CreateInBatches(schedules, 100).Error
}

func (r *LoanRepo) FindSchedulesByLoanID(ctx context.Context, loanID uuid.UUID) ([]loanmodel.RepaymentSchedule, error) {
	var schedules []loanmodel.RepaymentSchedule
	err := r.db.WithContext(ctx).
		Where("loan_id = ?", loanID).
		Order("installment_no ASC").
		Find(&schedules).Error
	return schedules, err
}

func (r *LoanRepo) UpdateSchedule(ctx context.Context, schedule *loanmodel.RepaymentSchedule) error {
	return r.db.WithContext(ctx).Save(schedule).Error
}

func (r *LoanRepo) FindByMonimeReference(ctx context.Context, ref string) *loanmodel.Loan {
	var loan loanmodel.Loan
	err := r.db.WithContext(ctx).
		Model(&loanmodel.Loan{}).
		Select("loans.*, COALESCE(CONCAT(vendors.first_name, ' ', vendors.last_name), '') as vendor_name").
		Joins("LEFT JOIN vendors ON vendors.id = loans.vendor_id").
		Where("loans.monime_reference = ?", ref).
		Preload("Schedules").
		First(&loan).Error
	if err != nil {
		return nil
	}
	return &loan
}

func (r *LoanRepo) ListByState(ctx context.Context, state loanmodel.LoanState, isDemo bool, offset, limit int) ([]*loanmodel.Loan, int64, error) {
	var loans []*loanmodel.Loan
	var count int64

	countQuery := r.db.WithContext(ctx).Model(&loanmodel.Loan{}).Where("is_demo = ?", isDemo)
	if state != "" {
		countQuery = countQuery.Where("state = ?", state)
	}
	countQuery.Count(&count)

	query := r.db.WithContext(ctx).
		Model(&loanmodel.Loan{}).
		Select("loans.*, COALESCE(CONCAT(vendors.first_name, ' ', vendors.last_name), '') as vendor_name").
		Joins("LEFT JOIN vendors ON vendors.id = loans.vendor_id").
		Where("loans.is_demo = ?", isDemo)
	if state != "" {
		query = query.Where("loans.state = ?", state)
	}
	err := query.
		Order("loans.created_at DESC").
		Offset(offset).Limit(limit).
		Find(&loans).Error

	return loans, count, err
}
