package postgres

import (
	"context"

	"github.com/google/uuid"
	partnermodel "github.com/marketpay/backend/internal/partner/domain/model"
	"gorm.io/gorm"
)

// PartnerRepo is the GORM implementation of PartnerRepository.
type PartnerRepo struct {
	db *gorm.DB
}

// NewPartnerRepo constructs a PartnerRepo.
func NewPartnerRepo(db *gorm.DB) *PartnerRepo {
	return &PartnerRepo{db: db}
}

func (r *PartnerRepo) Create(ctx context.Context, p *partnermodel.Partner) error {
	return r.db.WithContext(ctx).Create(p).Error
}

func (r *PartnerRepo) FindByID(ctx context.Context, id uuid.UUID) (*partnermodel.Partner, error) {
	var p partnermodel.Partner
	err := r.db.WithContext(ctx).First(&p, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *PartnerRepo) FindByAPIKey(ctx context.Context, apiKey string) (*partnermodel.Partner, error) {
	var p partnermodel.Partner
	err := r.db.WithContext(ctx).Where("api_key = ? AND is_active = true", apiKey).First(&p).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *PartnerRepo) Update(ctx context.Context, p *partnermodel.Partner) error {
	return r.db.WithContext(ctx).Save(p).Error
}

func (r *PartnerRepo) List(ctx context.Context, offset, limit int) ([]*partnermodel.Partner, int64, error) {
	var partners []*partnermodel.Partner
	var count int64
	r.db.WithContext(ctx).Model(&partnermodel.Partner{}).Count(&count)
	err := r.db.WithContext(ctx).
		Order("created_at DESC").
		Offset(offset).Limit(limit).
		Find(&partners).Error
	return partners, count, err
}
