package model

import (
	"time"

	shared "github.com/marketpay/backend/internal/shared/domain/model"
)

// MonimeEventType categorizes Monime webhook events.
type MonimeEventType string

const (
	EventDisbursementRequested MonimeEventType = "DisbursementRequested"
	EventDisbursementSucceeded MonimeEventType = "DisbursementSucceeded"
	EventDisbursementFailed    MonimeEventType = "DisbursementFailed"
	EventCollectionRequested   MonimeEventType = "CollectionRequested"
	EventCollectionReceived    MonimeEventType = "CollectionReceived"
	EventCollectionFailed      MonimeEventType = "CollectionFailed"
)

// MonimeTransactionStatus tracks outgoing/incoming payment state.
type MonimeTransactionStatus string

const (
	MonimeStatusPending   MonimeTransactionStatus = "PENDING"
	MonimeStatusSuccess   MonimeTransactionStatus = "SUCCESS"
	MonimeStatusFailed    MonimeTransactionStatus = "FAILED"
	MonimeStatusRetrying  MonimeTransactionStatus = "RETRYING"
	MonimeStatusManual    MonimeTransactionStatus = "MANUAL_REVIEW"
)

// MonimeTransaction persists every call to the Monime API.
type MonimeTransaction struct {
	shared.BaseModel
	Reference       string                  `gorm:"type:varchar(255);not null;uniqueIndex" json:"reference"`
	ExternalRef     string                  `gorm:"type:varchar(255);index" json:"external_ref"`
	Type            string                  `gorm:"type:varchar(50);not null" json:"type"` // DISBURSEMENT | COLLECTION
	Amount          float64                 `gorm:"type:decimal(15,2);not null" json:"amount"`
	Currency        string                  `gorm:"type:varchar(10);not null;default:'SLE'" json:"currency"`
	Phone           string                  `gorm:"type:varchar(20);not null" json:"phone"`
	Status          MonimeTransactionStatus `gorm:"type:varchar(50);not null;default:'PENDING'" json:"status"`
	RetryCount      int                     `gorm:"default:0" json:"retry_count"`
	NextRetryAt     *time.Time              `json:"next_retry_at,omitempty"`
	LastError       string                  `gorm:"type:text" json:"last_error,omitempty"`
	WebhookReceived bool                    `gorm:"default:false" json:"webhook_received"`
	WebhookPayload  string                  `gorm:"type:jsonb" json:"webhook_payload,omitempty"`
	EntityID        string                  `gorm:"type:varchar(255);index" json:"entity_id"` // loan ID or payment ID
	EntityType      string                  `gorm:"type:varchar(50)" json:"entity_type"`
}

// DisbursementRequest is sent to Monime to disburse a loan.
type DisbursementRequest struct {
	Reference   string  `json:"reference"`
	Phone       string  `json:"phone"`
	Amount      float64 `json:"amount"`
	Currency    string  `json:"currency"`
	Description string  `json:"description"`
}

// DisbursementResponse is received from Monime.
type DisbursementResponse struct {
	ExternalRef string `json:"external_ref"`
	Status      string `json:"status"`
	Message     string `json:"message"`
}

// CollectionRequest is sent to Monime to collect repayment.
type CollectionRequest struct {
	Reference   string  `json:"reference"`
	Phone       string  `json:"phone"`
	Amount      float64 `json:"amount"`
	Currency    string  `json:"currency"`
	Description string  `json:"description"`
}

// CollectionResponse is received from Monime.
type CollectionResponse struct {
	ExternalRef string `json:"external_ref"`
	Status      string `json:"status"`
	Message     string `json:"message"`
}

// WebhookEvent is the inbound webhook from Monime.
type WebhookEvent struct {
	EventType   MonimeEventType `json:"event_type"`
	Reference   string          `json:"reference"`
	ExternalRef string          `json:"external_ref"`
	Amount      float64         `json:"amount"`
	Currency    string          `json:"currency"`
	Timestamp   time.Time       `json:"timestamp"`
	Payload     map[string]interface{} `json:"payload"`
}
