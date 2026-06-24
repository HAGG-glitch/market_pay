package model

import (
	"time"

	"github.com/google/uuid"
	shared "github.com/marketpay/backend/internal/shared/domain/model"
	"golang.org/x/crypto/bcrypt"
)

// User is the auth aggregate root.
type User struct {
	shared.BaseModel
	Email         string          `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	Phone         string          `gorm:"type:varchar(20);uniqueIndex" json:"phone"`
	PasswordHash  string          `gorm:"type:varchar(255);not null" json:"-"`
	Role          shared.Role     `gorm:"type:varchar(50);not null" json:"role"`
	IsActive      bool            `gorm:"default:true" json:"is_active"`
	IsVerified    bool            `gorm:"default:false" json:"is_verified"`
	IsDemo        bool            `gorm:"default:false" json:"is_demo"`
	DisplayName   string          `gorm:"type:varchar(255)" json:"display_name,omitempty"`
	LastLoginAt   *time.Time      `json:"last_login_at,omitempty"`
	RefreshTokens []RefreshToken  `gorm:"foreignKey:UserID" json:"-"`
}

// RefreshToken stores issued refresh tokens.
type RefreshToken struct {
	shared.BaseModel
	UserID    uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	Token     string    `gorm:"type:varchar(512);uniqueIndex;not null" json:"-"`
	ExpiresAt time.Time `json:"expires_at"`
	Revoked   bool      `gorm:"default:false" json:"revoked"`
	User      *User     `gorm:"foreignKey:UserID" json:"-"`
}

// SetPassword hashes and stores a plaintext password.
func (u *User) SetPassword(plaintext string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintext), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.PasswordHash = string(hash)
	return nil
}

// CheckPassword verifies a plaintext password against the stored hash.
func (u *User) CheckPassword(plaintext string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(plaintext))
	return err == nil
}

// IsExpired checks if a refresh token has expired.
func (rt *RefreshToken) IsExpired() bool {
	return time.Now().After(rt.ExpiresAt)
}

// IsValid checks if a refresh token is usable.
func (rt *RefreshToken) IsValid() bool {
	return !rt.Revoked && !rt.IsExpired()
}

// TokenClaims holds JWT claim data.
type TokenClaims struct {
	UserID       uuid.UUID   `json:"user_id"`
	Email        string      `json:"email"`
	Role         shared.Role `json:"role"`
	VendorStatus string      `json:"vendor_status,omitempty"`
	KYCStatus    string      `json:"kyc_status,omitempty"`
}
