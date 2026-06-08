package model_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/marketpay/backend/internal/vendors/domain/model"
	shared "github.com/marketpay/backend/internal/shared/domain/model"
	"github.com/stretchr/testify/assert"
)

func makeAdultVendor() *model.Vendor {
	dob := time.Now().AddDate(-25, 0, 0) // 25 years old
	firstTx := time.Now().AddDate(0, -2, 0) // 2 months ago
	return &model.Vendor{
		FirstName:           "Aminata",
		LastName:            "Koroma",
		Phone:               "+23276123456",
		NationalIDNumber:    "NID123",
		NationalIDType:      "NATIONAL_ID",
		DateOfBirth:         dob,
		MarketAssociationID: uuid.New(),
		KYCStatus:           shared.KYCVerified,
		Status:              model.VendorStatusActive,
		FirstTransactionAt:  &firstTx,
	}
}

func TestVendor_Age(t *testing.T) {
	v := makeAdultVendor()
	assert.Equal(t, 25, v.Age())
}

func TestVendor_IsAdult_True(t *testing.T) {
	v := makeAdultVendor()
	assert.True(t, v.IsAdult())
}

func TestVendor_IsAdult_False(t *testing.T) {
	dob := time.Now().AddDate(-16, 0, 0)
	v := &model.Vendor{DateOfBirth: dob}
	assert.False(t, v.IsAdult())
}

func TestVendor_HasSufficientTransactionHistory_True(t *testing.T) {
	past := time.Now().AddDate(0, -2, 0)
	v := &model.Vendor{FirstTransactionAt: &past}
	assert.True(t, v.HasSufficientTransactionHistory())
}

func TestVendor_HasSufficientTransactionHistory_TooRecent(t *testing.T) {
	recent := time.Now().AddDate(0, 0, -10) // only 10 days
	v := &model.Vendor{FirstTransactionAt: &recent}
	assert.False(t, v.HasSufficientTransactionHistory())
}

func TestVendor_HasSufficientTransactionHistory_NoHistory(t *testing.T) {
	v := &model.Vendor{}
	assert.False(t, v.HasSufficientTransactionHistory())
}

func TestVendor_IsEligibleForLoan_AllGood(t *testing.T) {
	v := makeAdultVendor()
	err := v.IsEligibleForLoan()
	assert.NoError(t, err)
}

func TestVendor_IsEligibleForLoan_PendingKYC(t *testing.T) {
	v := makeAdultVendor()
	v.KYCStatus = shared.KYCPending
	err := v.IsEligibleForLoan()
	assert.Error(t, err)
}

func TestVendor_IsEligibleForLoan_Suspended(t *testing.T) {
	v := makeAdultVendor()
	v.Status = model.VendorStatusSuspended
	err := v.IsEligibleForLoan()
	assert.Error(t, err)
}

func TestVendor_IsEligibleForLoan_InsufficientHistory(t *testing.T) {
	v := makeAdultVendor()
	recent := time.Now().AddDate(0, 0, -5)
	v.FirstTransactionAt = &recent
	err := v.IsEligibleForLoan()
	assert.Error(t, err)
}

func TestVendor_KYCCompletenessScore_Full(t *testing.T) {
	v := makeAdultVendor()
	assert.Equal(t, 100.0, v.KYCCompletenessScore())
}

func TestVendor_KYCCompletenessScore_Partial(t *testing.T) {
	v := &model.Vendor{
		NationalIDNumber: "NID123",
		NationalIDType:   "NATIONAL_ID",
		// Missing market association and DOB
	}
	assert.Equal(t, 50.0, v.KYCCompletenessScore())
}

func TestVendor_FullName(t *testing.T) {
	v := &model.Vendor{FirstName: "Aminata", LastName: "Koroma"}
	assert.Equal(t, "Aminata Koroma", v.FullName())
}

func TestVendor_Activate_Suspend(t *testing.T) {
	v := &model.Vendor{Status: model.VendorStatusPending}
	v.Activate()
	assert.Equal(t, model.VendorStatusActive, v.Status)
	v.Suspend()
	assert.Equal(t, model.VendorStatusSuspended, v.Status)
}
