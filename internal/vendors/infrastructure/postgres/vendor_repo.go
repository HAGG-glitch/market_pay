package postgres

import (
	"context"

	"github.com/google/uuid"
	vendormodel "github.com/marketpay/backend/internal/vendors/domain/model"
	"gorm.io/gorm"
)

// VendorRepo is the GORM implementation of VendorRepository.
type VendorRepo struct {
	db *gorm.DB
}

// NewVendorRepo constructs a VendorRepo.
func NewVendorRepo(db *gorm.DB) *VendorRepo {
	return &VendorRepo{db: db}
}

func (r *VendorRepo) Create(ctx context.Context, vendor *vendormodel.Vendor) error {
	return r.db.WithContext(ctx).Create(vendor).Error
}

func (r *VendorRepo) FindByID(ctx context.Context, id uuid.UUID) (*vendormodel.Vendor, error) {
	var vendor vendormodel.Vendor
	err := r.db.WithContext(ctx).
		Preload("MarketAssociation").
		First(&vendor, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &vendor, nil
}

func (r *VendorRepo) FindByPhone(ctx context.Context, phone string) (*vendormodel.Vendor, error) {
	var vendor vendormodel.Vendor
	err := r.db.WithContext(ctx).Where("phone = ?", phone).First(&vendor).Error
	if err != nil {
		return nil, err
	}
	return &vendor, nil
}

func (r *VendorRepo) FindByUserID(ctx context.Context, userID uuid.UUID) (*vendormodel.Vendor, error) {
	var vendor vendormodel.Vendor
	err := r.db.WithContext(ctx).
		Preload("MarketAssociation").
		Where("user_id = ?", userID).
		First(&vendor).Error
	if err != nil {
		return nil, err
	}
	return &vendor, nil
}

func (r *VendorRepo) Update(ctx context.Context, vendor *vendormodel.Vendor) error {
	return r.db.WithContext(ctx).Save(vendor).Error
}

func (r *VendorRepo) List(ctx context.Context, offset, limit int) ([]*vendormodel.Vendor, int64, error) {
	var vendors []*vendormodel.Vendor
	var count int64

	r.db.WithContext(ctx).Model(&vendormodel.Vendor{}).Count(&count)

	err := r.db.WithContext(ctx).
		Preload("MarketAssociation").
		Order("created_at DESC").
		Offset(offset).Limit(limit).
		Find(&vendors).Error

	return vendors, count, err
}

func (r *VendorRepo) FindMarketAssociation(ctx context.Context, id uuid.UUID) (*vendormodel.MarketAssociation, error) {
	var ma vendormodel.MarketAssociation
	err := r.db.WithContext(ctx).First(&ma, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &ma, nil
}

func (r *VendorRepo) ListMarketAssociations(ctx context.Context) ([]*vendormodel.MarketAssociation, error) {
	var mas []*vendormodel.MarketAssociation
	err := r.db.WithContext(ctx).Order("name ASC").Find(&mas).Error
	return mas, err
}
