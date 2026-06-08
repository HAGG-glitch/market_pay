package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BaseModel provides common fields for all entities.
type BaseModel struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// BeforeCreate sets a UUID if not already set.
func (b *BaseModel) BeforeCreate(tx *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}

// Role represents the user's system role.
type Role string

const (
	RoleSuperAdmin  Role = "SUPER_ADMIN"
	RoleAdmin       Role = "ADMIN"
	RoleLoanOfficer Role = "LOAN_OFFICER"
	RoleFieldAgent  Role = "FIELD_AGENT"
	RoleVendor      Role = "VENDOR"
	RoleCustomer    Role = "CUSTOMER"
	RoleMFIPartner  Role = "MFI_PARTNER"
)

// Money is a value object representing a monetary amount in SLE.
type Money struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

// NewMoney constructs a Money value in SLE.
func NewMoney(amount float64) Money {
	return Money{Amount: amount, Currency: "SLE"}
}

// Add returns the sum of two Money values.
func (m Money) Add(other Money) Money {
	return Money{Amount: m.Amount + other.Amount, Currency: m.Currency}
}

// Subtract returns the difference of two Money values.
func (m Money) Subtract(other Money) Money {
	return Money{Amount: m.Amount - other.Amount, Currency: m.Currency}
}

// Multiply scales a Money value.
func (m Money) Multiply(factor float64) Money {
	return Money{Amount: m.Amount * factor, Currency: m.Currency}
}

// IsPositive checks if the amount is greater than zero.
func (m Money) IsPositive() bool {
	return m.Amount > 0
}

// IsZero checks if the amount is zero.
func (m Money) IsZero() bool {
	return m.Amount == 0
}

// PhoneNumber is a value object representing a Sierra Leonean phone number.
type PhoneNumber struct {
	Number      string `json:"number"`
	CountryCode string `json:"country_code"`
}

// NewSLPhoneNumber constructs a SL phone number.
func NewSLPhoneNumber(number string) PhoneNumber {
	return PhoneNumber{Number: number, CountryCode: "+232"}
}

// FullNumber returns the full international number.
func (p PhoneNumber) FullNumber() string {
	return p.CountryCode + p.Number
}

// Address represents a physical address.
type Address struct {
	Street   string `json:"street"`
	City     string `json:"city"`
	District string `json:"district"`
	Country  string `json:"country"`
}

// KYCStatus represents the KYC verification state.
type KYCStatus string

const (
	KYCPending   KYCStatus = "PENDING"
	KYCVerified  KYCStatus = "VERIFIED"
	KYCRejected  KYCStatus = "REJECTED"
)

// AuditLog records state changes for compliance.
type AuditLog struct {
	BaseModel
	ActorID    uuid.UUID `gorm:"type:uuid;not null" json:"actor_id"`
	ActorRole  Role      `gorm:"type:varchar(50);not null" json:"actor_role"`
	Action     string    `gorm:"type:varchar(100);not null" json:"action"`
	Resource   string    `gorm:"type:varchar(100);not null" json:"resource"`
	ResourceID string    `gorm:"type:varchar(255);not null" json:"resource_id"`
	OldState   string    `gorm:"type:text" json:"old_state,omitempty"`
	NewState   string    `gorm:"type:text" json:"new_state,omitempty"`
	IPAddress  string    `gorm:"type:varchar(50)" json:"ip_address,omitempty"`
	UserAgent  string    `gorm:"type:text" json:"user_agent,omitempty"`
}

// OutboxEvent represents a domain event pending publication.
type OutboxEvent struct {
	BaseModel
	EventType   string    `gorm:"type:varchar(100);not null;index" json:"event_type"`
	AggregateID string    `gorm:"type:varchar(255);not null;index" json:"aggregate_id"`
	Payload     string    `gorm:"type:jsonb;not null" json:"payload"`
	Status      string    `gorm:"type:varchar(50);not null;default:'PENDING'" json:"status"`
	RetryCount  int       `gorm:"default:0" json:"retry_count"`
	NextRetryAt time.Time `json:"next_retry_at"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
	Error       string    `gorm:"type:text" json:"error,omitempty"`
}

// Notification represents an outgoing notification.
type Notification struct {
	BaseModel
	RecipientID   uuid.UUID `gorm:"type:uuid;not null;index" json:"recipient_id"`
	RecipientPhone string   `gorm:"type:varchar(20)" json:"recipient_phone"`
	RecipientEmail string   `gorm:"type:varchar(255)" json:"recipient_email"`
	Channel       string    `gorm:"type:varchar(20);not null" json:"channel"` // sms, whatsapp, email
	EventType     string    `gorm:"type:varchar(100);not null" json:"event_type"`
	Subject       string    `gorm:"type:varchar(255)" json:"subject"`
	Body          string    `gorm:"type:text;not null" json:"body"`
	Status        string    `gorm:"type:varchar(50);not null;default:'PENDING'" json:"status"`
	SentAt        *time.Time `json:"sent_at,omitempty"`
	Error         string    `gorm:"type:text" json:"error,omitempty"`
}
