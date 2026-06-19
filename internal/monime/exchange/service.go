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
	notifier   Notifier
	log        *zap.Logger
}

func NewService(
	db *gorm.DB,
	crypto *monimeexchange.Crypto,
	vendorSvc *vendorapp.VendorService,
	loanSvc *loanapp.LoanService,
	paymentSvc *paymentapp.PaymentService,
	notifier Notifier,
	log *zap.Logger,
) *Service {
	return &Service{
		db: db, crypto: crypto, vendorSvc: vendorSvc,
		loanSvc: loanSvc, paymentSvc: paymentSvc, notifier: notifier, log: log,
	}
}

func (s *Service) Handle(ctx context.Context, raw monimeexchange.EncryptedRequest) (string, error) {
	payload, aesKey, err := s.crypto.DecryptRequest(raw)
	if err != nil {
		return "", err
	}

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
	case "mp_collect_market_name":
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
	phone := normalizePhone(p.Global.SubscriberMsisdn)

	if name == "" || market == "" {
		return monimeexchange.StopResponse{Action: "stop", Message: "Registration failed. Name and market are required."}, nil
	}

	parts := strings.Fields(name)
	first := parts[0]
	last := "Vendor"
	if len(parts) > 1 {
		last = strings.Join(parts[1:], " ")
	}

	userID := uuid.New()
	syntheticEmail := fmt.Sprintf("%s@ussd.marketpay.sl", strings.TrimPrefix(phone, "+"))
	pinHash := "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi"
	s.db.Exec(`INSERT INTO users (id, email, phone, password_hash, role, is_active, is_verified, is_demo, display_name)
		VALUES (?, ?, ?, ?, 'VENDOR', true, false, false, ?) ON CONFLICT DO NOTHING`,
		userID, syntheticEmail, phone, pinHash, name)

	_, err := s.vendorSvc.RegisterFromUSSD(ctx, vendorapp.USSDRegisterInput{
		FirstName:    first,
		LastName:     last,
		Phone:        phone,
		MarketName:   market,
		NationalID:   "USSD-" + strings.TrimPrefix(phone, "+"),
		BusinessType: "TRADER",
		PIN:          "0000",
		UserID:       userID,
		IsDemo:       false,
	})
	if err != nil {
		if strings.Contains(err.Error(), "exists") {
			return monimeexchange.StopResponse{Action: "stop", Message: "This phone number is already registered."}, nil
		}
		return nil, err
	}

	// Find the created vendor by user ID to link subscriber_id
	vendor, findErr := s.vendorSvc.GetByUserID(ctx, userID)
	if findErr != nil {
		s.log.Warn("vendor lookup after registration", zap.Error(findErr))
	} else {
		s.upsertSubscriber(ctx, p.Global.SubscriberID, vendor.ID, p.Global.SubscriberMsisdn)
	}

	s.notifier.NotifyRole(ctx, "LOAN_OFFICER", "VendorCreated",
		"New vendor registered", fmt.Sprintf("%s registered via USSD at %s", name, market), false)

	return monimeexchange.NavigateResponse{
		Action: "navigate",
		PageID: "mp_show_vendor_registration_result",
		PageData: map[string]interface{}{
			"vendor_name": name,
			"market_name": market,
			"message":     fmt.Sprintf("Vendor %s registered at %s. You will receive SMS confirmation.", name, market),
		},
	}, nil
}

func (s *Service) validatePayment(_ context.Context, sc map[string]interface{}) (interface{}, error) {
	code := stringValue(sc["payment_vendor_code"])
	amount := stringValue(sc["payment_amount"])
	confirmed := stringValue(sc["payment_confirmed"]) == "true"

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
func (s *Service) handleAccessGateExchange(ctx context.Context, p *monimeexchange.ExchangePayload) (interface{}, error) {
	subscriberID := p.Global.SubscriberID
	if subscriberID == "" {
		return monimeexchange.StopResponse{
			Action:  "stop",
			Message: "Flow doesn't exist.",
		}, nil
	}

	hash := normalizeSubscriberHash(subscriberID)

	var allowed struct{ Count int }
	err := s.db.Raw(`SELECT COUNT(*) AS count FROM ussd_allowed_subscribers WHERE subscriber_id_hash = ? AND is_active = true`, hash).Scan(&allowed).Error
	if err != nil || allowed.Count == 0 {
		return monimeexchange.StopResponse{
			Action:  "stop",
			Message: "Flow doesn't exist.",
		}, nil
	}

	return monimeexchange.NavigateResponse{
		Action: "navigate",
		PageID: "mp_select_service",
	}, nil
}

// normalizeSubscriberHash returns the SHA-256 hex digest of subscriberID.
// It MUST only be called once per value — never re-hash a previously hashed value.
func normalizeSubscriberHash(subscriberID string) string {
	h := sha256.Sum256([]byte(subscriberID))
	return hex.EncodeToString(h[:])
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
