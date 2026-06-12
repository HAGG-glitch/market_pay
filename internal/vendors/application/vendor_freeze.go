package application

import (
	"context"
	"time"

	"github.com/google/uuid"
	shared "github.com/marketpay/backend/internal/shared/domain/model"
	vendormodel "github.com/marketpay/backend/internal/vendors/domain/model"
	apperrors "github.com/marketpay/backend/pkg/errors"
)

// FreezeVendor freezes a vendor account and records history.
func (s *VendorService) FreezeVendor(ctx context.Context, vendorID, actorID uuid.UUID, actorRole shared.Role, reason string) error {
	vendor, err := s.vendors.FindByID(ctx, vendorID)
	if err != nil || vendor == nil {
		return apperrors.ErrNotFound("vendor")
	}
	now := time.Now()
	vendor.FrozenAt = &now
	vendor.FrozenBy = &actorID
	vendor.FreezeReason = reason
	vendor.Status = vendormodel.VendorStatusSuspended

	if err := s.vendors.Update(ctx, vendor); err != nil {
		return apperrors.ErrInternalServer(err)
	}

	_ = s.events.Publish(ctx, "AccountFrozen", vendorID.String(), map[string]interface{}{
		"vendor_id": vendorID.String(),
		"reason":    reason,
		"actor_id":  actorID.String(),
		"is_demo":   vendor.IsDemo,
	})

	s.logFreezeHistory(ctx, "vendor", vendorID, actorID, "FREEZE", reason, actorRole, vendor.IsDemo)
	return nil
}

// UnfreezeVendor reactivates a frozen vendor.
func (s *VendorService) UnfreezeVendor(ctx context.Context, vendorID, actorID uuid.UUID, actorRole shared.Role) error {
	vendor, err := s.vendors.FindByID(ctx, vendorID)
	if err != nil || vendor == nil {
		return apperrors.ErrNotFound("vendor")
	}
	vendor.FrozenAt = nil
	vendor.FrozenBy = nil
	vendor.FreezeReason = ""
	vendor.Activate()

	if err := s.vendors.Update(ctx, vendor); err != nil {
		return apperrors.ErrInternalServer(err)
	}

	_ = s.events.Publish(ctx, "AccountUnfrozen", vendorID.String(), map[string]interface{}{
		"vendor_id": vendorID.String(),
		"actor_id":  actorID.String(),
		"is_demo":   vendor.IsDemo,
	})

	s.logFreezeHistory(ctx, "vendor", vendorID, actorID, "UNFREEZE", "", actorRole, vendor.IsDemo)
	return nil
}

func (s *VendorService) logFreezeHistory(ctx context.Context, entityType string, entityID, actorID uuid.UUID, action, reason string, actorRole shared.Role, isDemo bool) {
	_ = s.vendors.LogFreezeHistory(ctx, entityType, entityID, actorID, action, reason, string(actorRole), isDemo)
}
