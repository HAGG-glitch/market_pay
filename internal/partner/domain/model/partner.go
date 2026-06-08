package model

import (
	shared "github.com/marketpay/backend/internal/shared/domain/model"
)

// PartnerType categorizes the capital provider.
type PartnerType string

const (
	PartnerTypeMFI  PartnerType = "MFI_PARTNER"
	PartnerTypeNGO  PartnerType = "NGO_FUND"
	PartnerTypeBank PartnerType = "BANK_PARTNER"
)

// Partner represents an MFI, NGO, or bank providing capital.
type Partner struct {
	shared.BaseModel
	Name              string      `gorm:"type:varchar(255);not null;uniqueIndex" json:"name"`
	Type              PartnerType `gorm:"type:varchar(50);not null" json:"type"`
	ContactEmail      string      `gorm:"type:varchar(255)" json:"contact_email"`
	ContactPhone      string      `gorm:"type:varchar(20)" json:"contact_phone"`
	CommissionRate    float64     `gorm:"type:decimal(5,4);not null" json:"commission_rate"`
	AvailableFunds    float64     `gorm:"type:decimal(20,2);default:0" json:"available_funds"`
	TotalDisbursed    float64     `gorm:"type:decimal(20,2);default:0" json:"total_disbursed"`
	TotalRepaid       float64     `gorm:"type:decimal(20,2);default:0" json:"total_repaid"`
	TotalCommission   float64     `gorm:"type:decimal(20,2);default:0" json:"total_commission"`
	IsActive          bool        `gorm:"default:true" json:"is_active"`
	APIKey            string      `gorm:"type:varchar(255);uniqueIndex" json:"-"`
}
