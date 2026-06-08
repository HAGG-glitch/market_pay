package postgres

import (
	"context"

	"github.com/google/uuid"
	loanmodel "github.com/marketpay/backend/internal/loan/domain/model"
	"gorm.io/gorm"
)

// RepaymentRepo implements RepaymentRepository using GORM.
type RepaymentRepo struct {
	db *gorm.DB
}

// NewRepaymentRepo constructs a RepaymentRepo.
func NewRepaymentRepo(db *gorm.DB) *RepaymentRepo {
	return &RepaymentRepo{db: db}
}

func (r *RepaymentRepo) FindScheduleByID(ctx context.Context, id uuid.UUID) (*loanmodel.RepaymentSchedule, error) {
	var s loanmodel.RepaymentSchedule
	err := r.db.WithContext(ctx).First(&s, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *RepaymentRepo) UpdateSchedule(ctx context.Context, s *loanmodel.RepaymentSchedule) error {
	return r.db.WithContext(ctx).Save(s).Error
}

func (r *RepaymentRepo) FindLoanByID(ctx context.Context, id uuid.UUID) (*loanmodel.Loan, error) {
	var loan loanmodel.Loan
	err := r.db.WithContext(ctx).First(&loan, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &loan, nil
}

func (r *RepaymentRepo) UpdateLoan(ctx context.Context, loan *loanmodel.Loan) error {
	return r.db.WithContext(ctx).Save(loan).Error
}

func (r *RepaymentRepo) FindOverdueSchedules(ctx context.Context) ([]loanmodel.RepaymentSchedule, error) {
	var schedules []loanmodel.RepaymentSchedule
	err := r.db.WithContext(ctx).
		Where("status IN ('PENDING','PARTIAL')").
		Order("due_date ASC").
		Find(&schedules).Error
	return schedules, err
}
