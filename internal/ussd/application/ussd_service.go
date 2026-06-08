package application

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	ussdmodel "github.com/marketpay/backend/internal/ussd/domain/model"
	apperrors "github.com/marketpay/backend/pkg/errors"
	"github.com/marketpay/backend/pkg/logger"
	"go.uber.org/zap"
)

// SessionRepository manages USSD sessions.
type SessionRepository interface {
	CreateOrUpdate(ctx context.Context, session *ussdmodel.USSDSession) error
	FindBySessionID(ctx context.Context, sessionID string) (*ussdmodel.USSDSession, error)
	Deactivate(ctx context.Context, sessionID string) error
}

// VendorLookup finds vendors by phone.
type VendorLookup interface {
	FindByPhone(ctx context.Context, phone string) (interface{}, error)
	VerifyPIN(ctx context.Context, phone, pin string) error
	CheckEligibility(ctx context.Context, phone string) (bool, string, error)
	GetLoanBalance(ctx context.Context, phone string) (float64, error)
	GetRepaymentSchedule(ctx context.Context, phone string) (string, error)
	GetSalesHistory(ctx context.Context, phone string) (string, error)
	GetGroupInfo(ctx context.Context, phone string) (string, error)
}

// LoanApplicator applies for loans via USSD.
type LoanApplicator interface {
	ApplyUSSD(ctx context.Context, phone, loanType string, amount float64) (string, error)
}

// PaymentInitiator initiates payments via USSD.
type PaymentInitiator interface {
	InitiateUSSD(ctx context.Context, fromPhone, toVendorCode string, amount float64) (string, error)
}

// USSDService processes USSD flows.
type USSDService struct {
	sessions SessionRepository
	vendors  VendorLookup
	loans    LoanApplicator
	payments PaymentInitiator
	timeout  time.Duration
	log      *logger.Logger
}

// NewUSSDService constructs a USSDService.
func NewUSSDService(
	sessions SessionRepository,
	vendors VendorLookup,
	loans LoanApplicator,
	payments PaymentInitiator,
	timeout time.Duration,
	log *logger.Logger,
) *USSDService {
	return &USSDService{
		sessions: sessions,
		vendors:  vendors,
		loans:    loans,
		payments: payments,
		timeout:  timeout,
		log:      log,
	}
}

// Process handles an inbound USSD request and returns the response.
func (s *USSDService) Process(ctx context.Context, req ussdmodel.USSDRequest) (ussdmodel.USSDResponse, error) {
	session, err := s.sessions.FindBySessionID(ctx, req.SessionID)
	if err != nil {
		// New session
		session = &ussdmodel.USSDSession{
			SessionID:   req.SessionID,
			PhoneNumber: req.PhoneNumber,
			MenuState:   ussdmodel.MenuStateMain,
			ExpiresAt:   time.Now().Add(s.timeout),
			IsActive:    true,
		}
	}

	if session.IsExpired() {
		_ = s.sessions.Deactivate(ctx, req.SessionID)
		return ussdmodel.EndResponse("Session expired. Please try again."), apperrors.ErrUSSDSessionExpired
	}

	// Split accumulated text to get current input
	parts := strings.Split(req.Text, "*")
	currentInput := ""
	if len(parts) > 0 {
		currentInput = parts[len(parts)-1]
	}

	response, nextState := s.handleMenu(ctx, session, currentInput, req)

	session.MenuState = nextState
	session.LastInput = currentInput
	session.ExpiresAt = time.Now().Add(s.timeout)
	_ = s.sessions.CreateOrUpdate(ctx, session)

	return response, nil
}

func (s *USSDService) handleMenu(
	ctx context.Context,
	session *ussdmodel.USSDSession,
	input string,
	req ussdmodel.USSDRequest,
) (ussdmodel.USSDResponse, ussdmodel.USSDMenuState) {

	switch session.MenuState {

	case ussdmodel.MenuStateMain:
		if req.Text == "" || input == "" {
			return ussdmodel.ContinueResponse(
				"Welcome to MarketPay\n" +
					"1. Register\n" +
					"2. Pay Vendor\n" +
					"3. Sales History\n" +
					"4. Loan Eligibility\n" +
					"5. Apply for Loan\n" +
					"6. Loan Balance\n" +
					"7. Repayment Schedule\n" +
					"8. Repay Loan\n" +
					"9. Group Info",
			), ussdmodel.MenuStateMain
		}

		switch input {
		case "1":
			return ussdmodel.ContinueResponse("Enter your full name:"), ussdmodel.MenuStateRegister
		case "2":
			return ussdmodel.ContinueResponse("Enter vendor code:"), ussdmodel.MenuStatePayVendor
		case "3":
			history, _ := s.vendors.GetSalesHistory(ctx, req.PhoneNumber)
			return ussdmodel.EndResponse(history), ussdmodel.MenuStateMain
		case "4":
			eligible, reason, _ := s.vendors.CheckEligibility(ctx, req.PhoneNumber)
			if eligible {
				return ussdmodel.EndResponse("You are eligible for a loan."), ussdmodel.MenuStateMain
			}
			return ussdmodel.EndResponse("Not eligible: " + reason), ussdmodel.MenuStateMain
		case "5":
			return ussdmodel.ContinueResponse("Select loan type:\n1. Emergency Advance (50-500 SLE)\n2. Starter Loan (500-2000 SLE)\n3. Growth Loan (2000-5000 SLE)"), ussdmodel.MenuStateApplyLoan
		case "6":
			balance, _ := s.vendors.GetLoanBalance(ctx, req.PhoneNumber)
			return ussdmodel.EndResponse(fmt.Sprintf("Outstanding balance: %.2f SLE", balance)), ussdmodel.MenuStateMain
		case "7":
			schedule, _ := s.vendors.GetRepaymentSchedule(ctx, req.PhoneNumber)
			return ussdmodel.EndResponse(schedule), ussdmodel.MenuStateMain
		case "8":
			return ussdmodel.ContinueResponse("Enter repayment amount (SLE):"), ussdmodel.MenuStateRepayLoan
		case "9":
			info, _ := s.vendors.GetGroupInfo(ctx, req.PhoneNumber)
			return ussdmodel.EndResponse(info), ussdmodel.MenuStateMain
		default:
			return ussdmodel.EndResponse("Invalid option. Goodbye."), ussdmodel.MenuStateMain
		}

	case ussdmodel.MenuStatePINEntry:
		if err := s.vendors.VerifyPIN(ctx, req.PhoneNumber, input); err != nil {
			return ussdmodel.EndResponse("Invalid PIN. Please try again."), ussdmodel.MenuStateMain
		}
		// Restore previous state from session data
		var stateData map[string]string
		_ = json.Unmarshal([]byte(session.StateData), &stateData)
		session.PINVerified = true
		return ussdmodel.ContinueResponse("PIN verified. Enter amount:"), ussdmodel.MenuStatePayVendor

	case ussdmodel.MenuStatePayVendor:
		if !session.PINVerified {
			saveStateData(session, "vendor_code", input)
			return ussdmodel.ContinueResponse("Enter your PIN:"), ussdmodel.MenuStatePINEntry
		}
		result, _ := s.payments.InitiateUSSD(ctx, req.PhoneNumber, input, 0)
		return ussdmodel.EndResponse(result), ussdmodel.MenuStateMain

	case ussdmodel.MenuStateApplyLoan:
		var loanType string
		switch input {
		case "1":
			loanType = "EMERGENCY_ADVANCE"
		case "2":
			loanType = "STARTER_LOAN"
		case "3":
			loanType = "GROWTH_LOAN"
		default:
			return ussdmodel.EndResponse("Invalid selection."), ussdmodel.MenuStateMain
		}
		saveStateData(session, "loan_type", loanType)
		return ussdmodel.ContinueResponse("Enter amount in SLE:"), ussdmodel.MenuStateConfirm

	case ussdmodel.MenuStateRepayLoan:
		return ussdmodel.EndResponse("Repayment request submitted. You will receive confirmation via SMS."), ussdmodel.MenuStateMain

	default:
		return ussdmodel.EndResponse("Service unavailable. Please try again."), ussdmodel.MenuStateMain
	}
}

func saveStateData(session *ussdmodel.USSDSession, key, value string) {
	data := make(map[string]string)
	_ = json.Unmarshal([]byte(session.StateData), &data)
	data[key] = value
	b, _ := json.Marshal(data)
	session.StateData = string(b)
	_ = session.StateData
}

func (s *USSDService) logRequest(req ussdmodel.USSDRequest) {
	s.log.Info("USSD request",
		zap.String("session_id", req.SessionID),
		zap.String("phone", req.PhoneNumber),
		zap.String("text", req.Text),
	)
}
