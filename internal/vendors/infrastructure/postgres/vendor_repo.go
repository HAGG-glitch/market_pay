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

func (r *VendorRepo) List(ctx context.Context, isDemo bool, fieldAgentID *uuid.UUID, offset, limit int) ([]*vendormodel.Vendor, int64, error) {
	var vendors []*vendormodel.Vendor
	var count int64

	q := r.db.WithContext(ctx).Model(&vendormodel.Vendor{}).Where("is_demo = ?", isDemo)
	if fieldAgentID != nil {
		q = q.Where("field_agent_id = ?", *fieldAgentID)
	}
	q.Count(&count)

	q2 := r.db.WithContext(ctx).Where("is_demo = ?", isDemo)
	if fieldAgentID != nil {
		q2 = q2.Where("field_agent_id = ?", *fieldAgentID)
	}
	err := q2.Preload("MarketAssociation").
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

// LogFreezeHistory records a freeze/unfreeze action.
func (r *VendorRepo) LogFreezeHistory(ctx context.Context, entityType string, entityID, actorID uuid.UUID, action, reason, actorRole string, isDemo bool) error {
	return r.db.WithContext(ctx).Exec(`
		INSERT INTO freeze_history (entity_type, entity_id, action, reason, actor_id, actor_role, is_demo)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		entityType, entityID, action, reason, actorID, actorRole, isDemo).Error
}
