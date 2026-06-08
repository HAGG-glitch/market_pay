package model

import (
	"time"

	"github.com/google/uuid"
	shared "github.com/marketpay/backend/internal/shared/domain/model"
)

// USSDMenuState tracks which menu the user is at.
type USSDMenuState string

const (
	MenuStateMain           USSDMenuState = "MAIN"
	MenuStateRegister       USSDMenuState = "REGISTER"
	MenuStatePayVendor      USSDMenuState = "PAY_VENDOR"
	MenuStateSalesHistory   USSDMenuState = "SALES_HISTORY"
	MenuStateLoanEligibility USSDMenuState = "LOAN_ELIGIBILITY"
	MenuStateApplyLoan      USSDMenuState = "APPLY_LOAN"
	MenuStateLoanBalance    USSDMenuState = "LOAN_BALANCE"
	MenuStateRepaySchedule  USSDMenuState = "REPAY_SCHEDULE"
	MenuStateRepayLoan      USSDMenuState = "REPAY_LOAN"
	MenuStateGroupInfo      USSDMenuState = "GROUP_INFO"
	MenuStatePINEntry       USSDMenuState = "PIN_ENTRY"
	MenuStateConfirm        USSDMenuState = "CONFIRM"
)

// USSDSession represents an active USSD session.
type USSDSession struct {
	shared.BaseModel
	SessionID   string        `gorm:"type:varchar(255);not null;uniqueIndex" json:"session_id"`
	PhoneNumber string        `gorm:"type:varchar(20);not null;index" json:"phone_number"`
	UserID      *uuid.UUID    `gorm:"type:uuid;index" json:"user_id,omitempty"`
	MenuState   USSDMenuState `gorm:"type:varchar(50);not null" json:"menu_state"`
	StateData   string        `gorm:"type:jsonb" json:"state_data"` // arbitrary JSON context
	PINVerified bool          `gorm:"default:false" json:"pin_verified"`
	LastInput   string        `gorm:"type:varchar(255)" json:"last_input"`
	ExpiresAt   time.Time     `gorm:"not null" json:"expires_at"`
	IsActive    bool          `gorm:"default:true" json:"is_active"`
}

// IsExpired checks if the session has timed out.
func (s *USSDSession) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// USSDRequest is the inbound USSD payload.
type USSDRequest struct {
	SessionID   string `json:"session_id" form:"sessionId"`
	ServiceCode string `json:"service_code" form:"serviceCode"`
	PhoneNumber string `json:"phone_number" form:"phoneNumber"`
	Text        string `json:"text" form:"text"` // accumulated input
}

// USSDResponse is returned to the USSD gateway.
type USSDResponse struct {
	Message    string `json:"message"`
	ContinueSession bool `json:"continue_session"`
}

// CON continues the session; END terminates it.
func ContinueResponse(msg string) USSDResponse {
	return USSDResponse{Message: "CON " + msg, ContinueSession: true}
}

func EndResponse(msg string) USSDResponse {
	return USSDResponse{Message: "END " + msg, ContinueSession: false}
}
