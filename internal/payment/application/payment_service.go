package application

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	paymentmodel "github.com/marketpay/backend/internal/payment/domain/model"
	apperrors "github.com/marketpay/backend/pkg/errors"
	"github.com/marketpay/backend/pkg/logger"
	"go.uber.org/zap"
)

// PaymentRepository defines payment persistence.
type PaymentRepository interface {
	Create(ctx context.Context, payment *paymentmodel.Payment) error
	FindByID(ctx context.Context, id uuid.UUID) (*paymentmodel.Payment, error)
	FindAll(ctx context.Context, isDemo bool, offset, limit int) ([]*paymentmodel.Payment, int64, error)
	FindByVendorID(ctx context.Context, vendorID uuid.UUID, offset, limit int) ([]*paymentmodel.Payment, int64, error)
	FindByCustomerID(ctx context.Context, customerID uuid.UUID, offset, limit int) ([]*paymentmodel.Payment, int64, error)
	Update(ctx context.Context, payment *paymentmodel.Payment) error
}

// EventPublisher publishes domain events.
type EventPublisher interface {
	Publish(ctx context.Context, eventType, aggregateID string, payload interface{}) error
}

// MonimeCollector calls the Monime payment gateway to collect money.
type MonimeCollector interface {
	Collect(ctx context.Context, phone, amount string) (string, error)
}

// PaymentService handles vendor payments.
type PaymentService struct {
	payments PaymentRepository
	events   EventPublisher
	monime   MonimeCollector
	log      *logger.Logger
}

func NewPaymentService(payments PaymentRepository, events EventPublisher, monime MonimeCollector, log *logger.Logger) *PaymentService {
	return &PaymentService{payments: payments, events: events, monime: monime, log: log}
}

// InitiateInput holds payment initiation data.
type InitiateInput struct {
	CustomerID uuid.UUID
	VendorID   uuid.UUID
	Amount     float64
	MonimeRef  string
	IsDemo     bool
}

// Initiate creates a new payment record and triggers Monime collection.
func (s *PaymentService) Initiate(ctx context.Context, input InitiateInput) (*paymentmodel.Payment, error) {
	if input.Amount <= 0 {
		return nil, apperrors.ErrInvalidAmount("amount must be positive")
	}

	payment := paymentmodel.NewPayment(input.CustomerID, input.VendorID, input.Amount)
	payment.IsDemo = input.IsDemo

	if err := s.payments.Create(ctx, payment); err != nil {
		return nil, apperrors.ErrInternalServer(err)
	}

	if s.monime != nil && !input.IsDemo {
		phone := "" // resolved from customer record upstream
		ref, err := s.monime.Collect(ctx, phone, fmt.Sprintf("%.2f", input.Amount))
		if err != nil {
			s.log.Error("monime collection failed", zap.Error(err))
			payment.Fail()
			s.payments.Update(ctx, payment)
		} else {
			payment.MonimeReference = ref
			s.payments.Update(ctx, payment)
		}
	}

	return payment, nil
}

// Complete marks a payment as completed.
func (s *PaymentService) Complete(ctx context.Context, paymentID uuid.UUID, monimeRef string) (*paymentmodel.Payment, error) {
	payment, err := s.payments.FindByID(ctx, paymentID)
	if err != nil {
		return nil, apperrors.ErrNotFound("payment")
	}

	payment.Complete(monimeRef)
	if err := s.payments.Update(ctx, payment); err != nil {
		return nil, apperrors.ErrInternalServer(err)
	}

	_ = s.events.Publish(ctx, "PaymentCompleted", payment.ID.String(), map[string]interface{}{
		"payment_id":  payment.ID.String(),
		"customer_id": payment.CustomerID.String(),
		"vendor_id":   payment.VendorID.String(),
		"amount":      payment.Amount,
		"fee":         payment.Fee,
		"net_amount":  payment.NetAmount,
	})

	return payment, nil
}

// List returns paginated payments filtered by demo mode.
func (s *PaymentService) List(ctx context.Context, isDemo bool, offset, limit int) ([]*paymentmodel.Payment, int64, error) {
	return s.payments.FindAll(ctx, isDemo, offset, limit)
}

// GetVendorPayments returns paginated payments received by a vendor.
func (s *PaymentService) GetVendorPayments(ctx context.Context, vendorID uuid.UUID, offset, limit int) ([]*paymentmodel.Payment, int64, error) {
	return s.payments.FindByVendorID(ctx, vendorID, offset, limit)
}
