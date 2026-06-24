package application

import (
	"context"

	"github.com/google/uuid"
	shared "github.com/marketpay/backend/internal/shared/domain/model"
	vendormodel "github.com/marketpay/backend/internal/vendors/domain/model"
	apperrors "github.com/marketpay/backend/pkg/errors"
)

// LoginUserByID issues tokens for an already-authenticated user.
func (s *AuthService) LoginUserByID(ctx context.Context, userID uuid.UUID) (*TokenPair, error) {
	user, err := s.users.FindByID(ctx, userID)
	if err != nil || user == nil {
		return nil, apperrors.ErrUnauthorized("user not found")
	}
	// Vendor authorization is managed via vendor.Status, not user.IsActive.
	// This allows pending vendors (registered via USSD) to log in and view
	// their status while waiting for loan officer approval.
	if !user.IsActive && user.Role != shared.RoleVendor {
		return nil, apperrors.ErrUnauthorized("account is suspended")
	}
	return s.issueTokenPair(ctx, user)
}

// LoginVendorByID issues tokens for a verified vendor, embedding vendor status in the JWT.
func (s *AuthService) LoginVendorByID(ctx context.Context, userID uuid.UUID, vendor *vendormodel.Vendor) (*TokenPair, error) {
	user, err := s.users.FindByID(ctx, userID)
	if err != nil || user == nil {
		return nil, apperrors.ErrUnauthorized("user not found")
	}
	if !user.IsActive && user.Role != shared.RoleVendor {
		return nil, apperrors.ErrUnauthorized("account is suspended")
	}
	return s.issueTokenPair(ctx, user, map[string]interface{}{
		"vendor_status": string(vendor.Status),
		"kyc_status":    string(vendor.KYCStatus),
	})
}
