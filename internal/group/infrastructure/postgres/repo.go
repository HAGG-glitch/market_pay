package postgres

import (
	"context"

	"github.com/google/uuid"
	groupmodel "github.com/marketpay/backend/internal/group/domain/model"
	"gorm.io/gorm"
)

// GroupRepo implements GroupRepository.
type GroupRepo struct {
	db *gorm.DB
}

func NewGroupRepo(db *gorm.DB) *GroupRepo {
	return &GroupRepo{db: db}
}

func (r *GroupRepo) Create(ctx context.Context, group *groupmodel.Group) error {
	return r.db.WithContext(ctx).Create(group).Error
}

func (r *GroupRepo) FindByID(ctx context.Context, id uuid.UUID) (*groupmodel.Group, error) {
	var group groupmodel.Group
	err := r.db.WithContext(ctx).
		Preload("Members").
		First(&group, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &group, nil
}

func (r *GroupRepo) FindByVendorID(ctx context.Context, vendorID uuid.UUID) (*groupmodel.Group, error) {
	var member groupmodel.GroupMember
	if err := r.db.WithContext(ctx).Where("vendor_id = ?", vendorID).First(&member).Error; err != nil {
		return nil, err
	}
	return r.FindByID(ctx, member.GroupID)
}

func (r *GroupRepo) Update(ctx context.Context, group *groupmodel.Group) error {
	return r.db.WithContext(ctx).Save(group).Error
}

func (r *GroupRepo) AddMember(ctx context.Context, member *groupmodel.GroupMember) error {
	return r.db.WithContext(ctx).Create(member).Error
}

func (r *GroupRepo) RemoveMember(ctx context.Context, groupID, vendorID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("group_id = ? AND vendor_id = ?", groupID, vendorID).
		Delete(&groupmodel.GroupMember{}).Error
}

func (r *GroupRepo) List(ctx context.Context, offset, limit int) ([]*groupmodel.Group, int64, error) {
	var groups []*groupmodel.Group
	var count int64
	r.db.WithContext(ctx).Model(&groupmodel.Group{}).Count(&count)
	err := r.db.WithContext(ctx).
		Preload("Members").
		Order("created_at DESC").
		Offset(offset).Limit(limit).
		Find(&groups).Error
	return groups, count, err
}
