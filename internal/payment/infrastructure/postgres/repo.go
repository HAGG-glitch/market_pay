package postgres

import (
	"context"

	"github.com/google/uuid"
	paymentmodel "github.com/marketpay/backend/internal/payment/domain/model"
	"gorm.io/gorm"
)

// PaymentRepo implements PaymentRepository.
type PaymentRepo struct {
	db *gorm.DB
}

func NewPaymentRepo(db *gorm.DB) *PaymentRepo {
	return &PaymentRepo{db: db}
}

func (r *PaymentRepo) Create(ctx context.Context, payment *paymentmodel.Payment) error {
	return r.db.WithContext(ctx).Create(payment).Error
}

func (r *PaymentRepo) FindByID(ctx context.Context, id uuid.UUID) (*paymentmodel.Payment, error) {
	var payment paymentmodel.Payment
	err := r.db.WithContext(ctx).First(&payment, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

func (r *PaymentRepo) FindByVendorID(ctx context.Context, vendorID uuid.UUID, offset, limit int) ([]*paymentmodel.Payment, int64, error) {
	var payments []*paymentmodel.Payment
	var count int64
	r.db.WithContext(ctx).Model(&paymentmodel.Payment{}).Where("vendor_id = ?", vendorID).Count(&count)
	err := r.db.WithContext(ctx).
		Where("vendor_id = ?", vendorID).
		Order("created_at DESC").
		Offset(offset).Limit(limit).
		Find(&payments).Error
	return payments, count, err
}

func (r *PaymentRepo) FindByCustomerID(ctx context.Context, customerID uuid.UUID, offset, limit int) ([]*paymentmodel.Payment, int64, error) {
	var payments []*paymentmodel.Payment
	var count int64
	r.db.WithContext(ctx).Model(&paymentmodel.Payment{}).Where("customer_id = ?", customerID).Count(&count)
	err := r.db.WithContext(ctx).
		Where("customer_id = ?", customerID).
		Order("created_at DESC").
		Offset(offset).Limit(limit).
		Find(&payments).Error
	return payments, count, err
}

func (r *PaymentRepo) Update(ctx context.Context, payment *paymentmodel.Payment) error {
	return r.db.WithContext(ctx).Save(payment).Error
}
