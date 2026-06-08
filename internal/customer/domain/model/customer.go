package model

import (
	shared "github.com/marketpay/backend/internal/shared/domain/model"
)

// Customer is registered by vendors or self-registration via USSD.
type Customer struct {
	shared.BaseModel
	UserID    string           `gorm:"type:varchar(255);uniqueIndex" json:"user_id"`
	FirstName string           `gorm:"type:varchar(100);not null" json:"first_name"`
	LastName  string           `gorm:"type:varchar(100);not null" json:"last_name"`
	Phone     string           `gorm:"type:varchar(20);not null;uniqueIndex" json:"phone"`
	PINHash   string           `gorm:"type:varchar(255)" json:"-"`
	KYCStatus shared.KYCStatus `gorm:"type:varchar(50);not null;default:'PENDING'" json:"kyc_status"`
	IsActive  bool             `gorm:"default:true" json:"is_active"`
}

// FullName returns the customer's full name.
func (c *Customer) FullName() string {
	return c.FirstName + " " + c.LastName
}
