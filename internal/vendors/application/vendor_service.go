package application

import (
	"context"
	"time"

	"github.com/google/uuid"
	vendormodel "github.com/marketpay/backend/internal/vendors/domain/model"
	shared "github.com/marketpay/backend/internal/shared/domain/model"
	apperrors "github.com/marketpay/backend/pkg/errors"
	"github.com/marketpay/backend/pkg/logger"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// VendorRepository defines persistence for vendors.
type VendorRepository interface {
	Create(ctx context.Context, vendor *vendormodel.Vendor) error
	FindByID(ctx context.Context, id uuid.UUID) (*vendormodel.Vendor, error)
	FindByPhone(ctx context.Context, phone string) (*vendormodel.Vendor, error)
	FindByUserID(ctx context.Context, userID uuid.UUID) (*vendormodel.Vendor, error)
	Update(ctx context.Context, vendor *vendormodel.Vendor) error
	List(ctx context.Context, isDemo bool, fieldAgentID *uuid.UUID, offset, limit int) ([]*vendormodel.Vendor, int64, error)
	FindMarketAssociation(ctx context.Context, id uuid.UUID) (*vendormodel.MarketAssociation, error)
	ListMarketAssociations(ctx context.Context) ([]*vendormodel.MarketAssociation, error)
	LogFreezeHistory(ctx context.Context, entityType string, entityID, actorID uuid.UUID, action, reason, actorRole string, isDemo bool) error
}

// EventPublisher publishes domain events to the outbox.
type EventPublisher interface {
	Publish(ctx context.Context, eventType, aggregateID string, payload interface{}) error
}

// VendorService handles vendor registration and management.
type VendorService struct {
	vendors   VendorRepository
	events    EventPublisher
	log       *logger.Logger
}

// NewVendorService constructs a VendorService.
func NewVendorService(vendors VendorRepository, events EventPublisher, log *logger.Logger) *VendorService {
	return &VendorService{vendors: vendors, events: events, log: log}
}

// CreateVendorInput holds the vendor registration payload.
type CreateVendorInput struct {
	UserID              uuid.UUID
	FirstName           string
	LastName            string
	Phone               string
	NationalIDNumber    string
	NationalIDType      string
	DateOfBirth         time.Time
	Address             string
	MarketAssociationID uuid.UUID
	BusinessName        string
	BusinessType        string
	PIN                 string
	IsDemo              bool
	FieldAgentID        *uuid.UUID
}

// Create registers a new vendor.
func (s *VendorService) Create(ctx context.Context, input CreateVendorInput) (*vendormodel.Vendor, error) {
	// Validate minimum age
	age := int(time.Since(input.DateOfBirth).Hours() / (24 * 365))
	if age < 18 {
		return nil, apperrors.ErrBadRequest("vendor must be at least 18 years old")
	}

	// Check phone uniqueness
	existing, _ := s.vendors.FindByPhone(ctx, input.Phone)
	if existing != nil {
		return nil, apperrors.ErrAlreadyExists("vendor with this phone")
	}

	// Hash PIN
	pinHash, err := bcrypt.GenerateFromPassword([]byte(input.PIN), bcrypt.DefaultCost)
	if err != nil {
		return nil, apperrors.ErrInternalServer(err)
	}

	vendor := &vendormodel.Vendor{
		UserID:              input.UserID,
		FirstName:           input.FirstName,
		LastName:            input.LastName,
		Phone:               input.Phone,
		NationalIDNumber:    input.NationalIDNumber,
		NationalIDType:      input.NationalIDType,
		DateOfBirth:         input.DateOfBirth,
		Address:             input.Address,
		MarketAssociationID: input.MarketAssociationID,
		BusinessName:        input.BusinessName,
		BusinessType:        input.BusinessType,
		KYCStatus:           shared.KYCPending,
		Status:              vendormodel.VendorStatusPending,
		PINHash:             string(pinHash),
		IsDemo:              input.IsDemo,
		FieldAgentID:        input.FieldAgentID,
	}

	if err := s.vendors.Create(ctx, vendor); err != nil {
		return nil, apperrors.ErrInternalServer(err)
	}

	payload := map[string]interface{}{
		"vendor_id": vendor.ID.String(),
		"phone":     vendor.Phone,
		"name":      vendor.FullName(),
		"is_demo":   vendor.IsDemo,
	}
	_ = s.events.Publish(ctx, "VendorRegistered", vendor.ID.String(), payload)
	_ = s.events.Publish(ctx, "VendorCreated", vendor.ID.String(), payload)

	s.log.Info("vendor registered", zap.String("vendor_id", vendor.ID.String()))
	return vendor, nil
}

// ApproveKYC marks the vendor's KYC as verified and activates them.
func (s *VendorService) ApproveKYC(ctx context.Context, vendorID uuid.UUID, approverID uuid.UUID) (*vendormodel.Vendor, error) {
	vendor, err := s.vendors.FindByID(ctx, vendorID)
	if err != nil || vendor == nil {
		return nil, apperrors.ErrNotFound("vendor")
	}

	now := time.Now()
	vendor.KYCStatus = shared.KYCVerified
	vendor.KYCVerifiedAt = &now
	vendor.Activate()

	if err := s.vendors.Update(ctx, vendor); err != nil {
		return nil, apperrors.ErrInternalServer(err)
	}

	return vendor, nil
}

// GetByID retrieves a vendor by ID.
func (s *VendorService) GetByID(ctx context.Context, id uuid.UUID) (*vendormodel.Vendor, error) {
	vendor, err := s.vendors.FindByID(ctx, id)
	if err != nil {
		return nil, apperrors.ErrNotFound("vendor")
	}
	return vendor, nil
}

// CheckEligibility evaluates loan eligibility for a vendor.
func (s *VendorService) CheckEligibility(ctx context.Context, vendorID uuid.UUID) error {
	vendor, err := s.vendors.FindByID(ctx, vendorID)
	if err != nil {
		return apperrors.ErrNotFound("vendor")
	}
	return vendor.IsEligibleForLoan()
}

// VerifyPIN checks if the supplied PIN matches the vendor's stored hash.
func (s *VendorService) VerifyPIN(ctx context.Context, vendorID uuid.UUID, pin string) error {
	vendor, err := s.vendors.FindByID(ctx, vendorID)
	if err != nil {
		return apperrors.ErrNotFound("vendor")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(vendor.PINHash), []byte(pin)); err != nil {
		return apperrors.ErrInvalidPIN
	}
	return nil
}

// List returns a paginated list of vendors scoped by demo mode.
func (s *VendorService) List(ctx context.Context, isDemo bool, fieldAgentID *uuid.UUID, offset, limit int) ([]*vendormodel.Vendor, int64, error) {
	return s.vendors.List(ctx, isDemo, fieldAgentID, offset, limit)
}
