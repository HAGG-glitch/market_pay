package exchange

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	loanapp "github.com/marketpay/backend/internal/loan/application"
	loanmodel "github.com/marketpay/backend/internal/loan/domain/model"
	paymentapp "github.com/marketpay/backend/internal/payment/application"
	repayapp "github.com/marketpay/backend/internal/repayment/application"
	vendorapp "github.com/marketpay/backend/internal/vendors/application"
	vendormodel "github.com/marketpay/backend/internal/vendors/domain/model"
	"github.com/marketpay/backend/pkg/monimeexchange"
	"github.com/marketpay/backend/pkg/realtime"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Notifier publishes in-app realtime events.
type Notifier interface {
	NotifyRole(ctx context.Context, role, eventType, title, body string, isDemo bool)
	NotifyUser(ctx context.Context, userID uuid.UUID, eventType, title, body string, isDemo bool)
}

// Service processes Monime USSD exchange actions.
type Service struct {
	db         *gorm.DB
	crypto     *monimeexchange.Crypto
	vendorSvc  *vendorapp.VendorService
	loanSvc    *loanapp.LoanService
	paymentSvc *paymentapp.PaymentService
	repaySvc   *repayapp.RepaymentService
	notifier   Notifier
	log        *zap.Logger
}

func NewService(
	db *gorm.DB,
	crypto *monimeexchange.Crypto,
	vendorSvc *vendorapp.VendorService,
	loanSvc *loanapp.LoanService,
	paymentSvc *paymentapp.PaymentService,
	repaySvc *repayapp.RepaymentService,
	notifier Notifier,
	log *zap.Logger,
) *Service {
	return &Service{
		db: db, crypto: crypto, vendorSvc: vendorSvc,
		loanSvc: loanSvc, paymentSvc: paymentSvc, repaySvc: repaySvc, notifier: notifier, log: log,
	}
}

func (s *Service) Handle(ctx context.Context, raw monimeexchange.EncryptedRequest) (string, error) {
	payload, aesKey, err := s.crypto.DecryptRequest(raw)
	if err != nil {
		return "", err
	}

	s.log.Info("incoming request", zap.String("page", payload.CurrentPage), zap.String("subscriber_id", payload.Global.SubscriberID), zap.Int("subscriber_id_len", len(payload.Global.SubscriberID)))

	sessionID := payload.Global.SessionID
	currentPage := payload.CurrentPage
	idempotencyKey := sessionID + "-" + currentPage

	// Check idempotency — return cached response if already processed
	if idempotencyKey != "-" {
		var existing struct{ ResponseData string }
		if err := s.db.Raw(`SELECT response_data FROM monime_exchange_sessions WHERE session_id = ? AND response_data != ''`, idempotencyKey).Scan(&existing).Error; err == nil && existing.ResponseData != "" {
			return existing.ResponseData, nil
		}
		// Claim the slot to prevent concurrent processing
		s.db.Exec(`INSERT INTO monime_exchange_sessions (session_id) VALUES (?) ON CONFLICT DO NOTHING`, idempotencyKey)
	}

	resp, err := s.route(ctx, payload)
	if err != nil {
		s.log.Error("exchange route", zap.Error(err), zap.String("page", currentPage))
		stop := monimeexchange.StopResponse{
			Action:  "stop",
			Message: "We could not complete your request. Please try again later.",
		}
		encrypted, encErr := s.crypto.EncryptResponse(stop, aesKey)
		if encErr != nil {
			return "", encErr
		}
		if idempotencyKey != "-" {
			s.db.Exec(`UPDATE monime_exchange_sessions SET response_data = ? WHERE session_id = ?`, encrypted, idempotencyKey)
		}
		return encrypted, nil
	}

	encrypted, err := s.crypto.EncryptResponse(resp, aesKey)
	if err != nil {
		return "", err
	}

	if idempotencyKey != "-" {
		s.db.Exec(`UPDATE monime_exchange_sessions SET response_data = ? WHERE session_id = ?`, encrypted, idempotencyKey)
	}
	return encrypted, nil
}

func (s *Service) route(ctx context.Context, p *monimeexchange.ExchangePayload) (interface{}, error) {
	sc := sessionContext(p)

	switch p.CurrentPage {
	case "mp_collect_market_name", "mp_pub_reg_confirm":
		return s.registerVendor(ctx, p, sc)
	case "mp_confirm_payment_receipt":
		return s.validatePayment(ctx, sc)
	case "mp_collect_payment_pin":
		return s.processPayment(ctx, p, sc)
	case "mp_balance_exchange":
		return s.checkBalance(ctx, p)
	case "mp_loan_eligibility_exchange":
		return s.checkLoanEligibility(ctx, p)
	case "mp_confirm_loan_application":
		return s.applyLoan(ctx, p, sc)
	case "mp_credit_score_exchange":
		return s.handleCreditScore(ctx, p)
	case "mp_repay_loan_exchange":
		return s.handleRepayLoan(ctx, p)
	case "mp_confirm_repayment":
		return s.handleConfirmRepayment(ctx, sc)
	case "mp_collect_repayment_pin":
		return s.handleRepaymentResult(ctx, p, sc)
	case "mp_transaction_history_exchange":
		return s.handleTransactionHistory(ctx, p)
	case "mp_pub_confirm_payment":
		return s.handlePublicConfirmPayment(ctx, sc)
	case "mp_pub_collect_payment_pin":
		return s.handlePublicPaymentResult(ctx, p, sc)
	case "mp_access_gate_exchange":
		return s.handleAccessGateExchange(ctx, p)
	default:
		return monimeexchange.StopResponse{
			Action:  "stop",
			Message: "Unknown service. Please dial again.",
		}, nil
	}
}

func (s *Service) registerVendor(ctx context.Context, p *monimeexchange.ExchangePayload, sc map[string]interface{}) (interface{}, error) {
	name := stringValue(sc["registration_vendor_name"])
	market := stringValue(sc["registration_market_name"])
	confirmed := stringValue(sc["registration_confirmed"])
	s.log.Info("vendor registration", zap.String("session", p.Global.SessionID), zap.String("name", name), zap.String("market", market), zap.String("confirmed", confirmed))

	if confirmed == "false" {
		return monimeexchange.NavigateResponse{
			Action: "navigate", PageID: "mp_pub_show_reg_cancelled",
			PageData: map[string]interface{}{"message": "Registration cancelled."},
		}, nil
	}

	if name == "" || market == "" {
		return monimeexchange.StopResponse{Action: "stop", Message: "Registration failed. Name and market are required."}, nil
	}

	// Collect phone from USSD text input, normalize to +232 format
	rawPhone := stringValue(sc["registration_phone"])
	phone := normalizePhone(rawPhone)
	if len(phone) < 10 {
		return monimeexchange.StopResponse{Action: "stop", Message: "Invalid phone number. Please try again."}, nil
	}

	parts := strings.Fields(name)
	first := parts[0]
	last := "Vendor"
	if len(parts) > 1 {
		last = strings.Join(parts[1:], " ")
	}

	userID := uuid.New()
	syntheticEmail := fmt.Sprintf("%s@ussd.marketpay.sl", strings.TrimPrefix(phone, "+"))
	pinHash := "$2a$10$WOx9GopNZF933jMGRn16/.1IyKAE3087DTJLUfhmJHOYArxQf/Rgq"

	// Use existing user ID if the user was already created (re-registration or partial failure)
	var existingUser struct{ Id string }
	if err := s.db.Raw(`SELECT id::text FROM users WHERE phone = ?`, phone).Scan(&existingUser).Error; err == nil && existingUser.Id != "" {
		if parsed, parseErr := uuid.Parse(existingUser.Id); parseErr == nil {
			userID = parsed
		}
	} else {
		s.db.Exec(`INSERT INTO users (id, email, phone, password_hash, role, is_active, is_verified, is_demo, display_name)
			VALUES (?, ?, ?, ?, 'VENDOR', false, false, false, ?) ON CONFLICT DO NOTHING`,
			userID, syntheticEmail, phone, pinHash, name)
	}

	// Find the default loan officer to assign as field agent
	var fieldAgentID *uuid.UUID
	var loanOfficer struct{ Id string }
	if err := s.db.Raw(`SELECT id::text FROM users WHERE role = 'LOAN_OFFICER' AND is_active = true ORDER BY created_at ASC LIMIT 1`).Scan(&loanOfficer).Error; err == nil && loanOfficer.Id != "" {
		if parsed, parseErr := uuid.Parse(loanOfficer.Id); parseErr == nil {
			fieldAgentID = &parsed
		}
	}

	_, err := s.vendorSvc.RegisterFromUSSD(ctx, vendorapp.USSDRegisterInput{
		FirstName:    first,
		LastName:     last,
		Phone:        phone,
		MarketName:   market,
		NationalID:   "USSD-" + strings.TrimPrefix(phone, "+"),
		BusinessType: "TRADER",
		PIN:          "0000",
		UserID:       userID,
		FieldAgentID: fieldAgentID,
		IsDemo:       false,
	})
	if err != nil {
		if strings.Contains(err.Error(), "exists") {
			return monimeexchange.StopResponse{Action: "stop", Message: "This phone number is already registered."}, nil
		}
		return nil, err
	}

	vendor, findErr := s.vendorSvc.GetByUserID(ctx, userID)
	if findErr != nil {
		s.log.Warn("vendor lookup after registration", zap.Error(findErr))
	} else {
		s.upsertSubscriber(ctx, p.Global.SubscriberID, vendor.ID, p.Global.SubscriberMsisdn)
		// Grant USSD access so the vendor can use the menu after registration
		s.db.Exec(`INSERT INTO ussd_allowed_subscribers (subscriber_id_hash, is_active) VALUES (?, true) ON CONFLICT (subscriber_id_hash) DO UPDATE SET is_active = true, updated_at = NOW()`,
			p.Global.SubscriberID)
	}

	s.notifier.NotifyRole(ctx, "LOAN_OFFICER", "VendorCreated",
		"New vendor registered",
		fmt.Sprintf("%s registered via USSD at %s (phone: %s)", name, market, phone), false)

	return monimeexchange.NavigateResponse{
		Action: "navigate",
		PageID: "mp_pub_show_reg_result",
		PageData: map[string]interface{}{
			"vendor_name": name,
			"market_name": market,
			"message":     fmt.Sprintf("Registration submitted for %s at %s.\nPhone: %s\nA field agent will contact you to verify and activate your account.", name, market, phone),
		},
	}, nil
}

func (s *Service) validatePayment(ctx context.Context, sc map[string]interface{}) (interface{}, error) {
	code := stringValue(sc["payment_vendor_code"])
	amount := stringValue(sc["payment_amount"])
	confirmed := stringValue(sc["payment_confirmed"]) == "true"
	s.log.Info("validate payment", zap.String("code", code), zap.String("amount", amount), zap.Bool("confirmed", confirmed))

	if !confirmed {
		return monimeexchange.NavigateResponse{Action: "navigate", PageID: "mp_show_payment_cancelled"}, nil
	}
	if !strings.HasPrefix(code, "MP") {
		return monimeexchange.StopResponse{Action: "stop", Message: "Invalid vendor code."}, nil
	}
	if amount == "" {
		return monimeexchange.StopResponse{Action: "stop", Message: "Invalid payment amount."}, nil
	}
	return monimeexchange.NavigateResponse{Action: "navigate", PageID: "mp_collect_payment_pin"}, nil
}

func (s *Service) processPayment(ctx context.Context, p *monimeexchange.ExchangePayload, sc map[string]interface{}) (interface{}, error) {
	code := stringValue(sc["payment_vendor_code"])
	amountStr := stringValue(sc["payment_amount"])
	sendSMS := stringValue(sc["payment_send_sms_receipt"]) == "true"
	s.log.Info("process payment", zap.String("session", p.Global.SessionID), zap.String("code", code), zap.String("amount", amountStr))

	var amount float64
	fmt.Sscanf(amountStr, "%f", &amount)
	if amount <= 0 {
		return monimeexchange.StopResponse{Action: "stop", Message: "Invalid amount."}, nil
	}

	ref := fmt.Sprintf("USSD-%s-%d", p.Global.SessionID, time.Now().Unix())

	// Look up vendor by code to validate it exists
	vendor, err := s.vendorSvc.GetByCode(ctx, code)
	if err != nil || vendor == nil {
		return monimeexchange.StopResponse{Action: "stop", Message: "Vendor code not found."}, nil
	}

	receipt := fmt.Sprintf("Payment SLE %.2f to %s. Ref: %s", amount, code, ref)
	if sendSMS {
		receipt += "\nSMS receipt will be sent."
	}

	s.notifier.NotifyRole(ctx, "LOAN_OFFICER", "RepaymentReceived",
		"USSD payment received", receipt, false)

	pageID := "mp_show_payment_result_no_sms"
	if sendSMS {
		pageID = "mp_show_payment_result_sms"
	}
	return monimeexchange.NavigateResponse{
		Action: "navigate",
		PageID: pageID,
		PageData: map[string]interface{}{
			"amount":      amountStr,
			"vendor_code": code,
			"reference":   ref,
		},
	}, nil
}

func (s *Service) checkBalance(ctx context.Context, p *monimeexchange.ExchangePayload) (interface{}, error) {
	s.log.Info("balance check", zap.String("session", p.Global.SessionID))
	vendor, err := s.findVendorBySubscriber(ctx, p.Global.SubscriberID)
	if err != nil || vendor == nil {
		return monimeexchange.NavigateResponse{
			Action: "navigate",
			PageID: "mp_show_balance_result",
			PageData: map[string]interface{}{
				"balance": "0.00",
				"message": "No vendor account found for this number. Please register first.",
			},
		}, nil
	}

	var balance float64
	s.db.Raw(`SELECT COALESCE(SUM(outstanding_amount), 0) FROM loans WHERE vendor_id = ? AND deleted_at IS NULL AND state IN ('ACTIVE','DISBURSED')`,
		vendor.ID).Scan(&balance)
	return monimeexchange.NavigateResponse{
		Action: "navigate",
		PageID: "mp_show_balance_result",
		PageData: map[string]interface{}{
			"balance": fmt.Sprintf("%.2f", balance),
			"message": fmt.Sprintf("Outstanding loan balance: SLE %.2f", balance),
		},
	}, nil
}

func (s *Service) checkLoanEligibility(ctx context.Context, p *monimeexchange.ExchangePayload) (interface{}, error) {
	s.log.Info("loan eligibility check", zap.String("session", p.Global.SessionID))
	vendor, err := s.findVendorBySubscriber(ctx, p.Global.SubscriberID)
	if err != nil || vendor == nil {
		return monimeexchange.NavigateResponse{
			Action: "navigate",
			PageID: "mp_show_loan_eligibility_result",
			PageData: map[string]interface{}{
				"eligible": false,
				"message":  "Register as a vendor first before checking loan eligibility.",
			},
		}, nil
	}

	eligible, reason, err := s.vendorSvc.CheckEligibilityByPhone(ctx, vendor.Phone)
	if err != nil {
		return nil, err
	}

	msg := "You are eligible for a loan."
	if !eligible {
		msg = reason
	}
	return monimeexchange.NavigateResponse{
		Action: "navigate",
		PageID: "mp_show_loan_eligibility_result",
		PageData: map[string]interface{}{
			"eligible": eligible,
			"message":  msg,
		},
	}, nil
}

func (s *Service) applyLoan(ctx context.Context, p *monimeexchange.ExchangePayload, sc map[string]interface{}) (interface{}, error) {
	s.log.Info("loan application", zap.String("session", p.Global.SessionID), zap.String("amount", stringValue(sc["loan_amount"])), zap.String("confirmed", stringValue(sc["loan_confirmed"])))
	if stringValue(sc["loan_confirmed"]) != "true" {
		return monimeexchange.NavigateResponse{Action: "navigate", PageID: "mp_show_loan_application_cancelled"}, nil
	}

	amountStr := stringValue(sc["loan_amount"])
	var amount float64
	fmt.Sscanf(amountStr, "%f", &amount)
	if amount <= 0 {
		return monimeexchange.StopResponse{Action: "stop", Message: "Invalid loan amount."}, nil
	}

	vendor, err := s.findVendorBySubscriber(ctx, p.Global.SubscriberID)
	if err != nil || vendor == nil {
		return monimeexchange.StopResponse{Action: "stop", Message: "Register as a vendor before applying for a loan."}, nil
	}

	loan, err := s.loanSvc.ApplyFromUSSD(ctx, loanapp.USSDApplyInput{
		VendorID:  vendor.ID,
		LoanType:  loanmodel.LoanTypeStarterLoan,
		Amount:    amount,
		TermWeeks: 4,
		Frequency: loanmodel.RepaymentFrequencyBiweekly,
		FundedBy:  loanmodel.FundingSourceMFIPartner,
		IsDemo:    false,
	})
	if err != nil {
		return monimeexchange.StopResponse{Action: "stop", Message: err.Error()}, nil
	}

	tracking := loan.ID.String()[:8]
	s.notifier.NotifyRole(ctx, "LOAN_OFFICER", "LoanRequested",
		"New loan application", fmt.Sprintf("Vendor %s applied for SLE %.2f. Tracking: %s", vendor.FullName(), amount, tracking), false)
	s.notifier.NotifyRole(ctx, "MFI_PARTNER", "LoanRequested",
		"New loan for review", fmt.Sprintf("Loan %s pending review.", tracking), false)

	return monimeexchange.NavigateResponse{
		Action: "navigate",
		PageID: "mp_show_loan_application_result",
		PageData: map[string]interface{}{
			"tracking_number": tracking,
			"message":         fmt.Sprintf("Loan application submitted. Tracking number: %s", tracking),
		},
	}, nil
}

// findVendorBySubscriber looks up a vendor through the subscriber_id mapping.
func (s *Service) findVendorBySubscriber(ctx context.Context, subscriberID string) (*vendormodel.Vendor, error) {
	if subscriberID == "" {
		return nil, fmt.Errorf("empty subscriber id")
	}
	var sub struct{ VendorID uuid.UUID }
	err := s.db.Raw(`SELECT vendor_id FROM ussd_subscribers WHERE subscriber_id = ?`, subscriberID).Scan(&sub).Error
	if err != nil || sub.VendorID == uuid.Nil {
		return nil, fmt.Errorf("no vendor linked to subscriber")
	}
	return s.vendorSvc.GetByID(ctx, sub.VendorID)
}

// upsertSubscriber creates or updates a subscriber identity mapping.
func (s *Service) upsertSubscriber(ctx context.Context, subscriberID string, vendorID uuid.UUID, maskedMsisdn string) {
	if subscriberID == "" {
		return
	}
	s.db.Exec(`INSERT INTO ussd_subscribers (subscriber_id, vendor_id, masked_msisdn) VALUES (?, ?, ?)
		ON CONFLICT (subscriber_id) DO UPDATE SET vendor_id = EXCLUDED.vendor_id, masked_msisdn = EXCLUDED.masked_msisdn, updated_at = NOW()`,
		subscriberID, vendorID, maskedMsisdn)
}

// handleAccessGateExchange checks if the subscriber is allowed to use the USSD flow.
// subscriberID from Monime is already a SHA-256 hash — compare directly, do NOT re-hash.
func (s *Service) handleAccessGateExchange(ctx context.Context, p *monimeexchange.ExchangePayload) (interface{}, error) {
	subscriberID := p.Global.SubscriberID
	subHash := subscriberID
	if len(subHash) > 8 {
		subHash = subHash[:8]
	}
	if subscriberID == "" {
		return monimeexchange.StopResponse{
			Action:  "stop",
			Message: "Flow doesn't exist.",
		}, nil
	}

	var allowed struct{ Count int }
	err := s.db.Raw(`SELECT COUNT(*) AS count FROM ussd_allowed_subscribers WHERE subscriber_id_hash = ? AND is_active = true`, subscriberID).Scan(&allowed).Error
	if err != nil || allowed.Count == 0 {
		s.log.Warn("access gate denied", zap.String("subscriber_id", subscriberID))
		return monimeexchange.StopResponse{
			Action:  "stop",
			Message: "Flow doesn't exist.",
		}, nil
	}

	s.log.Info("access gate allowed", zap.String("subscriber_id", subscriberID))

	return monimeexchange.NavigateResponse{
		Action: "navigate",
		PageID: "mp_select_service",
	}, nil
}

func (s *Service) handleCreditScore(ctx context.Context, p *monimeexchange.ExchangePayload) (interface{}, error) {
	s.log.Info("credit score request", zap.String("session", p.Global.SessionID))
	vendor, err := s.findVendorBySubscriber(ctx, p.Global.SubscriberID)
	if err != nil || vendor == nil {
		return monimeexchange.NavigateResponse{
			Action: "navigate",
			PageID: "mp_show_credit_score",
			PageData: map[string]interface{}{
				"message": "No vendor account found. Please contact support.",
			},
		}, nil
	}
	return monimeexchange.NavigateResponse{
		Action: "navigate",
		PageID: "mp_show_credit_score",
		PageData: map[string]interface{}{
			"credit_score": vendor.CreditScore,
			"message":      fmt.Sprintf("Your credit score: %.0f\nMaintain good repayment history to improve your score.", vendor.CreditScore),
		},
	}, nil
}

func (s *Service) handleRepayLoan(ctx context.Context, p *monimeexchange.ExchangePayload) (interface{}, error) {
	s.log.Info("repay loan request", zap.String("session", p.Global.SessionID))
	vendor, err := s.findVendorBySubscriber(ctx, p.Global.SubscriberID)
	if err != nil || vendor == nil {
		return monimeexchange.NavigateResponse{
			Action: "navigate",
			PageID: "mp_collect_repayment_amount",
			PageData: map[string]interface{}{
				"balance": "0.00",
			},
		}, nil
	}
	var balance float64
	s.db.Raw(`SELECT COALESCE(SUM(outstanding_amount), 0) FROM loans WHERE vendor_id = ? AND deleted_at IS NULL AND state IN ('ACTIVE','DISBURSED')`, vendor.ID).Scan(&balance)
	return monimeexchange.NavigateResponse{
		Action: "navigate",
		PageID: "mp_collect_repayment_amount",
		PageData: map[string]interface{}{
			"balance": fmt.Sprintf("%.2f", balance),
		},
	}, nil
}

func (s *Service) handleConfirmRepayment(ctx context.Context, sc map[string]interface{}) (interface{}, error) {
	confirmed := stringValue(sc["repayment_confirmed"]) == "true"
	s.log.Info("confirm repayment", zap.Bool("confirmed", confirmed), zap.String("amount", stringValue(sc["payment_amount"])))
	if !confirmed {
		return monimeexchange.NavigateResponse{
			Action: "navigate",
			PageID: "mp_show_repayment_cancelled",
			PageData: map[string]interface{}{
				"message": "Repayment cancelled.",
			},
		}, nil
	}
	return monimeexchange.NavigateResponse{Action: "navigate", PageID: "mp_collect_repayment_pin"}, nil
}

func (s *Service) handleRepaymentResult(ctx context.Context, p *monimeexchange.ExchangePayload, sc map[string]interface{}) (interface{}, error) {
	amountStr := stringValue(sc["payment_amount"])
	ref := fmt.Sprintf("REPAY-%s-%d", p.Global.SessionID, time.Now().Unix())
	s.log.Info("repayment result", zap.String("session", p.Global.SessionID), zap.String("amount", amountStr), zap.String("ref", ref))

	vendor, err := s.findVendorBySubscriber(ctx, p.Global.SubscriberID)
	if err == nil && vendor != nil {
		s.notifier.NotifyRole(ctx, "LOAN_OFFICER", "RepaymentReceived",
			"USSD loan repayment",
			fmt.Sprintf("Vendor %s repaid SLE %s. Ref: %s", vendor.FullName(), amountStr, ref), false)

		// Record and confirm the repayment immediately.
		// The exchange callback confirms the subscriber was debited successfully,
		// so we trust it rather than waiting for a webhook with no per-transaction ref.
		amount := 0.0
		fmt.Sscanf(amountStr, "%f", &amount)
		loanIDs := s.activeLoanIDs(ctx, vendor.ID)
		for _, loanID := range loanIDs {
			rec, err := s.repaySvc.RecordRepayment(ctx, repayapp.RecordRepaymentInput{
				LoanID:     loanID,
				VendorID:   vendor.ID,
				Amount:     amount,
				MonimeRef:  ref,
				PaymentRef: ref,
				Metadata: map[string]interface{}{
					"source":        "ussd_repayment",
					"session_id":    p.Global.SessionID,
					"masked_phone":  p.Global.SubscriberMsisdn,
					"subscriber_id": p.Global.SubscriberID,
					"monime_event":  p.ExportedData,
				},
			})
			if err != nil {
				s.log.Error("failed to record repayment", zap.Error(err), zap.String("ref", ref))
				break
			}
			s.log.Info("repayment recorded", zap.String("id", rec.ID.String()), zap.String("ref", ref))
			if err := s.repaySvc.ConfirmRepayment(ctx, ref); err != nil {
				s.log.Error("failed to confirm repayment", zap.Error(err), zap.String("ref", ref))
			}
			break // one record per repayment
		}
	}

	return monimeexchange.NavigateResponse{
		Action: "navigate",
		PageID: "mp_show_repayment_result",
		PageData: map[string]interface{}{
			"reference": ref,
			"message":   fmt.Sprintf("Repayment of SLE %s successful.\nRef: %s\nThank you.", amountStr, ref),
		},
	}, nil
}

// activeLoanIDs returns IDs of active loans for a vendor.
func (s *Service) activeLoanIDs(ctx context.Context, vendorID uuid.UUID) []uuid.UUID {
	var ids []uuid.UUID
	s.db.Raw(`SELECT id FROM loans WHERE vendor_id = ? AND state = 'ACTIVE' AND deleted_at IS NULL LIMIT 1`, vendorID).Scan(&ids)
	return ids
}

func (s *Service) handleTransactionHistory(ctx context.Context, p *monimeexchange.ExchangePayload) (interface{}, error) {
	s.log.Info("transaction history request", zap.String("session", p.Global.SessionID))
	vendor, err := s.findVendorBySubscriber(ctx, p.Global.SubscriberID)
	if err != nil || vendor == nil {
		return monimeexchange.NavigateResponse{
			Action: "navigate",
			PageID: "mp_show_transaction_history_result",
			PageData: map[string]interface{}{
				"message": "No vendor account found.",
			},
		}, nil
	}

	type LoanRecord struct {
		Principal float64
		State     string
		CreatedAt time.Time
	}
	var loans []LoanRecord
	s.db.Raw(`SELECT principal_amount AS principal, state, created_at FROM loans WHERE vendor_id = ? AND deleted_at IS NULL ORDER BY created_at DESC LIMIT 10`, vendor.ID).Scan(&loans)

	if len(loans) == 0 {
		return monimeexchange.NavigateResponse{
			Action: "navigate",
			PageID: "mp_show_transaction_history_result",
			PageData: map[string]interface{}{
				"message": "No transaction history found.",
			},
		}, nil
	}

	var sb strings.Builder
	sb.WriteString("Recent transactions:\n")
	for i, l := range loans {
		friendlyState := l.State
		switch l.State {
		case "ACTIVE", "DISBURSED":
			friendlyState = "Active"
		case "PAID":
			friendlyState = "Paid"
		case "DRAFT":
			friendlyState = "Pending"
		case "REJECTED":
			friendlyState = "Rejected"
		}
		sb.WriteString(fmt.Sprintf("%d. SLE %.2f - %s\n", i+1, l.Principal, friendlyState))
	}

	return monimeexchange.NavigateResponse{
		Action: "navigate",
		PageID: "mp_show_transaction_history_result",
		PageData: map[string]interface{}{
			"message": sb.String(),
		},
	}, nil
}

func (s *Service) handlePublicConfirmPayment(ctx context.Context, sc map[string]interface{}) (interface{}, error) {
	code := stringValue(sc["payment_vendor_code"])
	amount := stringValue(sc["payment_amount"])
	confirmed := stringValue(sc["payment_confirmed"]) == "true"
	s.log.Info("public confirm payment", zap.String("code", code), zap.String("amount", amount), zap.Bool("confirmed", confirmed))

	if !confirmed {
		return monimeexchange.NavigateResponse{Action: "navigate", PageID: "mp_pub_show_payment_cancelled"}, nil
	}
	if !strings.HasPrefix(code, "MP") {
		return monimeexchange.StopResponse{Action: "stop", Message: "Invalid vendor code."}, nil
	}
	if amount == "" {
		return monimeexchange.StopResponse{Action: "stop", Message: "Invalid payment amount."}, nil
	}
	return monimeexchange.NavigateResponse{Action: "navigate", PageID: "mp_pub_collect_payment_pin"}, nil
}

func (s *Service) handlePublicPaymentResult(ctx context.Context, p *monimeexchange.ExchangePayload, sc map[string]interface{}) (interface{}, error) {
	code := stringValue(sc["payment_vendor_code"])
	amountStr := stringValue(sc["payment_amount"])
	ref := fmt.Sprintf("USSD-%s-%d", p.Global.SessionID, time.Now().Unix())
	s.log.Info("public payment result", zap.String("session", p.Global.SessionID), zap.String("code", code), zap.String("amount", amountStr), zap.String("ref", ref))

	s.notifier.NotifyRole(ctx, "LOAN_OFFICER", "PaymentReceived",
		"USSD public payment",
		fmt.Sprintf("Customer paid SLE %s to %s. Ref: %s", amountStr, code, ref), false)

	// Record and confirm the repayment immediately
	vendor, err := s.vendorSvc.GetByCode(ctx, code)
	if err == nil && vendor != nil {
		amount := 0.0
		fmt.Sscanf(amountStr, "%f", &amount)
		loanIDs := s.activeLoanIDs(ctx, vendor.ID)
		for _, loanID := range loanIDs {
			rec, err := s.repaySvc.RecordRepayment(ctx, repayapp.RecordRepaymentInput{
				LoanID:     loanID,
				VendorID:   vendor.ID,
				Amount:     amount,
				MonimeRef:  ref,
				PaymentRef: ref,
				Metadata: map[string]interface{}{
					"source":       "ussd_public_payment",
					"vendor_code":  code,
					"session_id":   p.Global.SessionID,
					"masked_phone": p.Global.SubscriberMsisdn,
				},
			})
			if err != nil {
				s.log.Error("failed to record public repayment", zap.Error(err), zap.String("ref", ref))
				break
			}
			s.log.Info("public repayment recorded", zap.String("id", rec.ID.String()), zap.String("ref", ref))
			if err := s.repaySvc.ConfirmRepayment(ctx, ref); err != nil {
				s.log.Error("failed to confirm public repayment", zap.Error(err), zap.String("ref", ref))
			}
			break
		}
	}

	return monimeexchange.NavigateResponse{
		Action: "navigate",
		PageID: "mp_pub_show_payment_result",
		PageData: map[string]interface{}{
			"reference": ref,
			"message":   fmt.Sprintf("Payment of SLE %s to %s successful.\nRef: %s\nThank you.", amountStr, code, ref),
		},
	}, nil
}

func sessionContext(p *monimeexchange.ExchangePayload) map[string]interface{} {
	merged := map[string]interface{}{}
	for k, v := range p.FlowData {
		merged[k] = v
	}
	for k, v := range p.ExportedData {
		merged[k] = v
	}
	for k, v := range p.SessionContext {
		merged[k] = v
	}
	if sc, ok := p.SessionContext["mutations"].(map[string]interface{}); ok {
		for k, v := range sc {
			merged[k] = v
		}
	}
	return merged
}

func stringValue(v interface{}) string {
	if v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return t
	default:
		return fmt.Sprintf("%v", t)
	}
}

func normalizePhone(msisdn string) string {
	digits := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, msisdn)
	if strings.HasPrefix(digits, "232") {
		return "+" + digits
	}
	if len(digits) == 9 {
		return "+232" + digits
	}
	return "+" + digits
}

func (s *Service) isDuplicate(sessionID string, p *monimeexchange.ExchangePayload) bool {
	if sessionID == "" {
		return false
	}
	h := sha256.Sum256([]byte(sessionID + p.CurrentPage + fmt.Sprint(p.FlowData)))
	hash := hex.EncodeToString(h[:])
	var existing string
	err := s.db.Raw(`SELECT response_hash FROM monime_exchange_sessions WHERE session_id = ?`, sessionID+"-"+p.CurrentPage).Scan(&existing).Error
	return err == nil && existing == hash
}

func (s *Service) recordSession(sessionID string, resp interface{}) {
	if sessionID == "" {
		return
	}
	h := sha256.Sum256([]byte(fmt.Sprint(resp)))
	hash := hex.EncodeToString(h[:])
	s.db.Exec(`INSERT INTO monime_exchange_sessions (session_id, response_hash) VALUES (?, ?)
		ON CONFLICT (session_id) DO UPDATE SET response_hash = EXCLUDED.response_hash`,
		sessionID, hash)
}

// InAppNotifier bridges exchange events to SSE hub + DB.
type InAppNotifier struct {
	db  *gorm.DB
	hub *realtime.Hub
}

func NewInAppNotifier(db *gorm.DB, hub *realtime.Hub) *InAppNotifier {
	return &InAppNotifier{db: db, hub: hub}
}

func (n *InAppNotifier) NotifyRole(ctx context.Context, role, eventType, title, body string, isDemo bool) {
	n.hub.PublishRole(role, realtime.Event{Type: eventType, Payload: map[string]string{
		"title": title, "body": body, "role": role,
	}})
	var userIDs []uuid.UUID
	n.db.Raw(`SELECT id FROM users WHERE role = ? AND is_active = true AND is_demo = ?`, role, isDemo).Scan(&userIDs)
	for _, id := range userIDs {
		n.NotifyUser(ctx, id, eventType, title, body, isDemo)
	}
}

func (n *InAppNotifier) NotifyUser(ctx context.Context, userID uuid.UUID, eventType, title, body string, isDemo bool) {
	n.db.Exec(`INSERT INTO in_app_notifications (recipient_id, event_type, title, body, is_demo)
		VALUES (?, ?, ?, ?, ?)`, userID, eventType, title, body, isDemo)
	n.hub.Publish(userID.String(), realtime.Event{Type: eventType, Payload: map[string]string{
		"title": title, "body": body,
	}})
}
