package model_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/marketpay/backend/internal/payment/domain/model"
	"github.com/stretchr/testify/assert"
)

func TestNewPayment_FeeCalculation(t *testing.T) {
	customerID := uuid.New()
	vendorID   := uuid.New()

	p := model.NewPayment(customerID, vendorID, 1000.0)

	assert.Equal(t, 1000.0, p.Amount)
	assert.Equal(t, 10.0, p.Fee)       // 1% of 1000
	assert.Equal(t, 990.0, p.NetAmount) // 1000 - 10
	assert.Equal(t, model.PaymentStatusPending, p.Status)
	assert.Equal(t, "SLE", p.Currency)
}

func TestNewPayment_SmallAmount(t *testing.T) {
	p := model.NewPayment(uuid.New(), uuid.New(), 50.0)
	assert.Equal(t, 0.5, p.Fee)
	assert.Equal(t, 49.5, p.NetAmount)
}

func TestPayment_Complete(t *testing.T) {
	p := model.NewPayment(uuid.New(), uuid.New(), 500.0)
	p.Complete("MONIME-REF-123")

	assert.Equal(t, model.PaymentStatusCompleted, p.Status)
	assert.Equal(t, "MONIME-REF-123", p.MonimeReference)
}

func TestPayment_Fail(t *testing.T) {
	p := model.NewPayment(uuid.New(), uuid.New(), 500.0)
	p.Fail()
	assert.Equal(t, model.PaymentStatusFailed, p.Status)
}

func TestPayment_IDs_Set(t *testing.T) {
	customerID := uuid.New()
	vendorID   := uuid.New()
	p := model.NewPayment(customerID, vendorID, 100.0)

	assert.Equal(t, customerID, p.CustomerID)
	assert.Equal(t, vendorID, p.VendorID)
}
