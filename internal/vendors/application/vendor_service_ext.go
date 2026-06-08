package application

import (
	"context"

	"github.com/google/uuid"
	vendormodel "github.com/marketpay/backend/internal/vendors/domain/model"
	apperrors "github.com/marketpay/backend/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

// VerifyPINByPhone looks up a vendor by phone and verifies their PIN.
func (s *VendorService) VerifyPINByPhone(ctx context.Context, phone, pin string) error {
	vendor, err := s.vendors.FindByPhone(ctx, phone)
	if err != nil {
		return apperrors.ErrNotFound("vendor")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(vendor.PINHash), []byte(pin)); err != nil {
		return apperrors.ErrInvalidPIN
	}
	return nil
}

// ListMarketAssociations returns all market associations.
func (s *VendorService) ListMarketAssociations(ctx context.Context) ([]*vendormodel.MarketAssociation, error) {
	return s.vendors.ListMarketAssociations(ctx)
}

// UpdateCreditScore persists the latest credit score on the vendor record.
func (s *VendorService) UpdateCreditScore(ctx context.Context, vendorID uuid.UUID, score float64) error {
	vendor, err := s.vendors.FindByID(ctx, vendorID)
	if err != nil {
		return apperrors.ErrNotFound("vendor")
	}
	vendor.CreditScore = score
	return s.vendors.Update(ctx, vendor)
}
