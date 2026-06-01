package ussd

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
)

// MarketPayFlowService handles the MarketPay USSD flow
type MarketPayFlowService struct {
	store StateStore
}

// NewMarketPayFlowService creates a new MarketPay flow service
func NewMarketPayFlowService(store StateStore) *MarketPayFlowService {
	return &MarketPayFlowService{store: store}
}

// Advance processes the USSD flow independent of HTTP or gRPC transport details
func (s *MarketPayFlowService) Advance(ctx context.Context, input AdvanceFlowInput) (*FlowResult, error) {
	if ctx.Err() != nil {
		log.Error().Err(ctx.Err()).
			Str("session_id", input.SessionID).
			Str("current_page", string(input.CurrentPage)).
			Msg("context error before MarketPay USSD advance")
		return nil, ctx.Err()
	}

	currentPage := input.CurrentPage
	if strings.TrimSpace(string(currentPage)) == "" {
		currentPage = PageSelectService
		log.Debug().Str("session_id", input.SessionID).Msg("defaulting to PageSelectService")
	}

	log.Debug().
		Str("session_id", input.SessionID).
		Str("current_page", string(currentPage)).
		Int("input_values_count", len(input.Values)).
		Msg("MarketPay USSD flow advance started")

	data, err := s.loadState(ctx, input.SessionID)
	if err != nil {
		log.Error().Err(err).
			Str("session_id", input.SessionID).
			Msg("failed to load MarketPay USSD session state")
		return nil, err
	}
	mergeValues(data, input.Values)

	// Handle page-based flow structure
	switch currentPage {

	case PageSelectService:
		return s.handleSelectService(ctx, input.SessionID, data)

	case PageCollectVendorName:
		return s.handleCollectVendorName(ctx, input.SessionID, data)

	case PageCollectMarketName:
		return s.handleCollectMarketName(ctx, input.SessionID, data)

	case PageSubmitVendorRegistration:
		return s.handleSubmitVendorRegistration(ctx, input.SessionID, data)

	case PageShowVendorRegistrationResult:
		return s.handleShowVendorRegistrationResult(ctx, input.SessionID, data)

	case PageCollectPaymentVendorCode:
		return s.handleCollectPaymentVendorCode(ctx, input.SessionID, data)

	case PageCollectPaymentAmount:
		return s.handleCollectPaymentAmount(ctx, input.SessionID, data)

	case PageConfirmPaymentChoice:
		return s.handleConfirmPaymentChoice(ctx, input.SessionID, data)

	case PageSubmitVendorPayment:
		return s.handleSubmitVendorPayment(ctx, input.SessionID, data)

	case PageShowPaymentResult:
		return s.handleShowPaymentResult(ctx, input.SessionID, data)

	case PageShowPaymentCancelled:
		return s.handleShowPaymentCancelled(ctx, input.SessionID, data)

	case PageFetchTransactionHistory:
		return s.handleFetchTransactionHistory(ctx, input.SessionID, data)

	case PageShowTransactionHistory:
		return s.handleShowTransactionHistory(ctx, input.SessionID, data)

	case PageFetchBalance:
		return s.handleFetchBalance(ctx, input.SessionID, data)

	case PageShowBalance:
		return s.handleShowBalance(ctx, input.SessionID, data)

	case PageFetchLoanEligibility:
		return s.handleFetchLoanEligibility(ctx, input.SessionID, data)

	case PageShowLoanEligibility:
		return s.handleShowLoanEligibility(ctx, input.SessionID, data)

	case PageCollectLoanAmount:
		return s.handleCollectLoanAmount(ctx, input.SessionID, data)

	case PageConfirmLoanApplication:
		return s.handleConfirmLoanApplication(ctx, input.SessionID, data)

	case PageSubmitLoanApplication:
		return s.handleSubmitLoanApplication(ctx, input.SessionID, data)

	case PageShowLoanApplicationResult:
		return s.handleShowLoanApplicationResult(ctx, input.SessionID, data)

	case PageShowLoanApplicationCancelled:
		return s.handleShowLoanApplicationCancelled(ctx, input.SessionID, data)

	case PageExitService:
		return s.handleExitService(ctx, input.SessionID, data)

	default:
		log.Error().
			Str("session_id", input.SessionID).
			Str("unknown_page", string(currentPage)).
			Msg("unknown page in MarketPay USSD flow")
		return &FlowResult{
			Action:   ActionStop,
			Message:  fmt.Sprintf("We could not process page %q. Please try again or restart the session.", currentPage),
			Data:     cloneMap(data),
		}, nil
	}
}

// loadState loads session state from the store
func (s *MarketPayFlowService) loadState(ctx context.Context, sessionID string) (map[string]string, error) {
	if s.store == nil || strings.TrimSpace(sessionID) == "" {
		log.Debug().
			Str("session_id", sessionID).
			Bool("store_nil", s.store == nil).
			Msg("returning empty state: no store or session ID")
		return map[string]string{}, nil
	}
	data, err := s.store.Load(ctx, sessionID)
	if err != nil {
		log.Debug().Str("session_id", sessionID).Msg("session not found in store, starting new")
		return map[string]string{}, nil
	}
	if data == nil {
		log.Debug().Str("session_id", sessionID).Msg("loaded nil state, initializing empty")
		return map[string]string{}, nil
	}
	log.Debug().
		Str("session_id", sessionID).
		Int("state_keys", len(data)).
		Msg("successfully loaded state from store")
	return cloneMap(data), nil
}

// persistState saves session state to the store
func (s *MarketPayFlowService) persistState(ctx context.Context, sessionID string, data map[string]string) error {
	if s.store == nil || strings.TrimSpace(sessionID) == "" {
		log.Debug().
			Str("session_id", sessionID).
			Bool("store_nil", s.store == nil).
			Msg("skipping state persistence: no store or session ID")
		return nil
	}
	log.Debug().
		Str("session_id", sessionID).
		Int("state_keys", len(data)).
		Msg("persisting state to store")
	if err := s.store.Save(ctx, sessionID, cloneMap(data)); err != nil {
		log.Error().Err(err).
			Str("session_id", sessionID).
			Int("state_keys", len(data)).
			Msg("failed to persist state to store")
		return err
	}
	log.Debug().Str("session_id", sessionID).Msg("state successfully persisted")
	return nil
}

// navigateToNext saves state and navigates to next page
func (s *MarketPayFlowService) navigateToNext(ctx context.Context, sessionID string, data map[string]string, nextPage FlowPage, message string) (*FlowResult, error) {
	if err := s.persistState(ctx, sessionID, data); err != nil {
		log.Error().Err(err).
			Str("session_id", sessionID).
			Str("next_page", string(nextPage)).
			Msg("failed to persist state during navigation")
		return nil, err
	}
	log.Info().
		Str("session_id", sessionID).
		Str("next_page", string(nextPage)).
		Msg("MarketPay USSD flow navigating")
	return &FlowResult{
		Action:   ActionNavigate,
		NextPage: nextPage,
		Message:  message,
		Data:     cloneMap(data),
	}, nil
}

// stopFlow stops the flow and returns a message
func (s *MarketPayFlowService) stopFlow(ctx context.Context, sessionID string, data map[string]string, message string) (*FlowResult, error) {
	if err := s.persistState(ctx, sessionID, data); err != nil {
		log.Warn().Err(err).
			Str("session_id", sessionID).
			Msg("failed to persist state during stop")
	}
	return &FlowResult{
		Action:  ActionStop,
		Message: message,
		Data:    cloneMap(data),
	}, nil
}

// handleSelectService handles the main service selection menu
func (s *MarketPayFlowService) handleSelectService(ctx context.Context, sessionID string, data map[string]string) (*FlowResult, error) {
	log.Debug().Str("session_id", sessionID).Msg("processing PageSelectService")

	selectedService := strings.TrimSpace(data["selected_service"])
	if selectedService == "" {
		// Show menu for initial request
		return s.navigateToNext(ctx, sessionID, data, PageSelectService,
			"Welcome to MarketPay\nSelect service\n"+
				"1. Register vendor\n"+
				"2. Pay vendor\n"+
				"3. Transaction history\n"+
				"4. Check balance\n"+
				"5. Loan eligibility\n"+
				"6. Loan application\n"+
				"7. Exit")
	}

	// Route based on selected service
	switch selectedService {
	case "register_vendor":
		return s.navigateToNext(ctx, sessionID, data, PageCollectVendorName, "Enter vendor name")
	case "pay_vendor":
		return s.navigateToNext(ctx, sessionID, data, PageCollectPaymentVendorCode, "Enter vendor code")
	case "transaction_history":
		return s.navigateToNext(ctx, sessionID, data, PageFetchTransactionHistory, "Loading transactions, please wait")
	case "balance_check":
		return s.navigateToNext(ctx, sessionID, data, PageFetchBalance, "Checking balance, please wait")
	case "loan_eligibility":
		return s.navigateToNext(ctx, sessionID, data, PageFetchLoanEligibility, "Checking eligibility, please wait")
	case "loan_application":
		return s.navigateToNext(ctx, sessionID, data, PageCollectLoanAmount, "Enter loan amount in SLE")
	case "exit":
		return s.navigateToNext(ctx, sessionID, data, PageExitService, "Thanks for using MarketPay.")
	default:
		return s.stopFlow(ctx, sessionID, data, "Invalid service selection. Please try again.")
	}
}

// handleCollectVendorName collects vendor name for registration
func (s *MarketPayFlowService) handleCollectVendorName(ctx context.Context, sessionID string, data map[string]string) (*FlowResult, error) {
	log.Debug().Str("session_id", sessionID).Msg("processing PageCollectVendorName")

	vendorName := strings.TrimSpace(data["registration_vendor_name"])
	if vendorName == "" {
		return s.navigateToNext(ctx, sessionID, data, PageCollectVendorName, "Enter vendor name")
	}

	if !ValidateVendorName(vendorName) {
		log.Warn().Str("session_id", sessionID).Str("vendor_name", maskSensitive(vendorName)).
			Msg("validation failed: invalid vendor name format")
		return s.stopFlow(ctx, sessionID, data, GetValidationMessage("vendor_name"))
	}

	data["registration_vendor_name"] = vendorName
	return s.navigateToNext(ctx, sessionID, data, PageCollectMarketName, "Enter market name")
}

// handleCollectMarketName collects market name for registration
func (s *MarketPayFlowService) handleCollectMarketName(ctx context.Context, sessionID string, data map[string]string) (*FlowResult, error) {
	log.Debug().Str("session_id", sessionID).Msg("processing PageCollectMarketName")

	marketName := strings.TrimSpace(data["registration_market_name"])
	if marketName == "" {
		return s.navigateToNext(ctx, sessionID, data, PageCollectMarketName, "Enter market name")
	}

	if !ValidateMarketName(marketName) {
		log.Warn().Str("session_id", sessionID).Str("market_name", maskSensitive(marketName)).
			Msg("validation failed: invalid market name format")
		return s.stopFlow(ctx, sessionID, data, GetValidationMessage("market_name"))
	}

	data["registration_market_name"] = marketName
	return s.navigateToNext(ctx, sessionID, data, PageSubmitVendorRegistration, "Registering vendor, please wait")
}

// handleSubmitVendorRegistration submits vendor registration
func (s *MarketPayFlowService) handleSubmitVendorRegistration(ctx context.Context, sessionID string, data map[string]string) (*FlowResult, error) {
	log.Debug().Str("session_id", sessionID).Msg("processing PageSubmitVendorRegistration")

	// TODO: Call external API to register vendor
	// For now, simulate success
	vendorName := firstNonEmpty(data, "registration_vendor_name")
	data["registration_status"] = "success"
	data["vendor_id"] = "V" + sessionID[:10]

	message := fmt.Sprintf("Vendor %s registered successfully.\nVendor ID: %s", vendorName, data["vendor_id"])
	return s.navigateToNext(ctx, sessionID, data, PageShowVendorRegistrationResult, message)
}

// handleShowVendorRegistrationResult displays registration result
func (s *MarketPayFlowService) handleShowVendorRegistrationResult(ctx context.Context, sessionID string, data map[string]string) (*FlowResult, error) {
	log.Debug().Str("session_id", sessionID).Msg("processing PageShowVendorRegistrationResult")
	
	if err := s.persistState(ctx, sessionID, data); err != nil {
		log.Warn().Err(err).Str("session_id", sessionID).Msg("failed to persist state")
	}

	return &FlowResult{
		Action:   ActionNavigate,
		NextPage: PageSelectService,
		Message:  "Registration complete. Select another service.",
		Data:     cloneMap(data),
	}, nil
}

// handleCollectPaymentVendorCode collects vendor code for payment
func (s *MarketPayFlowService) handleCollectPaymentVendorCode(ctx context.Context, sessionID string, data map[string]string) (*FlowResult, error) {
	log.Debug().Str("session_id", sessionID).Msg("processing PageCollectPaymentVendorCode")

	vendorCode := strings.TrimSpace(data["payment_vendor_code"])
	if vendorCode == "" {
		return s.navigateToNext(ctx, sessionID, data, PageCollectPaymentVendorCode, "Enter vendor code\nUse a code like MP12345")
	}

	if !ValidateVendorCode(vendorCode) {
		log.Warn().Str("session_id", sessionID).Str("vendor_code", maskVendorCode(vendorCode)).
			Msg("validation failed: invalid vendor code format")
		return s.stopFlow(ctx, sessionID, data, GetValidationMessage("vendor_code"))
	}

	data["payment_vendor_code"] = vendorCode
	return s.navigateToNext(ctx, sessionID, data, PageCollectPaymentAmount, "Enter amount in SLE")
}

// handleCollectPaymentAmount collects payment amount
func (s *MarketPayFlowService) handleCollectPaymentAmount(ctx context.Context, sessionID string, data map[string]string) (*FlowResult, error) {
	log.Debug().Str("session_id", sessionID).Msg("processing PageCollectPaymentAmount")

	amountStr := strings.TrimSpace(data["payment_amount"])
	if amountStr == "" {
		return s.navigateToNext(ctx, sessionID, data, PageCollectPaymentAmount, "Enter amount in SLE\nMin: 1, Max: 10,000,000")
	}

	amount, err := strconv.ParseInt(amountStr, 10, 64)
	if err != nil {
		return s.stopFlow(ctx, sessionID, data, "Please enter a valid amount")
	}

	if !ValidateAmount(amount, 1, 10000000) {
		if amount < 1 {
			return s.stopFlow(ctx, sessionID, data, GetValidationMessage("amount_min"))
		}
		return s.stopFlow(ctx, sessionID, data, GetValidationMessage("amount_max"))
	}

	data["payment_amount"] = amountStr
	vendor := maskVendorCode(data["payment_vendor_code"])
	return s.navigateToNext(ctx, sessionID, data, PageConfirmPaymentChoice,
		fmt.Sprintf("Send %s SLE to %s?\n1. Send with SMS receipt\n2. Send without SMS receipt\n3. Cancel", amountStr, vendor))
}

// handleConfirmPaymentChoice handles payment confirmation
func (s *MarketPayFlowService) handleConfirmPaymentChoice(ctx context.Context, sessionID string, data map[string]string) (*FlowResult, error) {
	log.Debug().Str("session_id", sessionID).Msg("processing PageConfirmPaymentChoice")

	choice := strings.TrimSpace(data["payment_confirmed"])
	if choice == "" {
		return s.navigateToNext(ctx, sessionID, data, PageConfirmPaymentChoice, "Confirm payment")
	}

	if choice == "cancel" || choice == "3" {
		data["payment_confirmed"] = "false"
		return s.navigateToNext(ctx, sessionID, data, PageShowPaymentCancelled, "Payment cancelled")
	}

	data["payment_confirmed"] = "true"
	if choice == "pay_sms" || choice == "1" {
		data["payment_send_sms_receipt"] = "true"
	} else {
		data["payment_send_sms_receipt"] = "false"
	}
	return s.navigateToNext(ctx, sessionID, data, PageSubmitVendorPayment, "Submitting payment, please wait")
}

// handleSubmitVendorPayment submits the vendor payment
func (s *MarketPayFlowService) handleSubmitVendorPayment(ctx context.Context, sessionID string, data map[string]string) (*FlowResult, error) {
	log.Debug().Str("session_id", sessionID).Msg("processing PageSubmitVendorPayment")

	// TODO: Call external API to process payment
	// For now, simulate success
	data["payment_status"] = "success"
	data["payment_ref"] = "MP" + sessionID[:12]

	amount := firstNonEmpty(data, "payment_amount")
	vendor := maskVendorCode(data["payment_vendor_code"])
	message := fmt.Sprintf("Payment of %s SLE to %s successful.\nRef: %s", amount, vendor, data["payment_ref"])

	return s.navigateToNext(ctx, sessionID, data, PageShowPaymentResult, message)
}

// handleShowPaymentResult displays payment result
func (s *MarketPayFlowService) handleShowPaymentResult(ctx context.Context, sessionID string, data map[string]string) (*FlowResult, error) {
	log.Debug().Str("session_id", sessionID).Msg("processing PageShowPaymentResult")
	
	if err := s.persistState(ctx, sessionID, data); err != nil {
		log.Warn().Err(err).Str("session_id", sessionID).Msg("failed to persist state")
	}

	return &FlowResult{
		Action:   ActionNavigate,
		NextPage: PageSelectService,
		Message:  "Payment complete. Select another service.",
		Data:     cloneMap(data),
	}, nil
}

// handleShowPaymentCancelled displays payment cancelled message
func (s *MarketPayFlowService) handleShowPaymentCancelled(ctx context.Context, sessionID string, data map[string]string) (*FlowResult, error) {
	log.Debug().Str("session_id", sessionID).Msg("processing PageShowPaymentCancelled")
	
	if err := s.persistState(ctx, sessionID, data); err != nil {
		log.Warn().Err(err).Str("session_id", sessionID).Msg("failed to persist state")
	}

	return &FlowResult{
		Action:   ActionNavigate,
		NextPage: PageSelectService,
		Message:  "Payment cancelled. Select another service.",
		Data:     cloneMap(data),
	}, nil
}

// handleFetchTransactionHistory fetches transaction history
func (s *MarketPayFlowService) handleFetchTransactionHistory(ctx context.Context, sessionID string, data map[string]string) (*FlowResult, error) {
	log.Debug().Str("session_id", sessionID).Msg("processing PageFetchTransactionHistory")

	// TODO: Call external API to fetch transaction history
	// For now, simulate response
	data["transaction_history"] = "3 transactions found"
	return s.navigateToNext(ctx, sessionID, data, PageShowTransactionHistory, "Loading transactions, please wait")
}

// handleShowTransactionHistory displays transaction history
func (s *MarketPayFlowService) handleShowTransactionHistory(ctx context.Context, sessionID string, data map[string]string) (*FlowResult, error) {
	log.Debug().Str("session_id", sessionID).Msg("processing PageShowTransactionHistory")
	
	history := firstNonEmpty(data, "transaction_history")
	if err := s.persistState(ctx, sessionID, data); err != nil {
		log.Warn().Err(err).Str("session_id", sessionID).Msg("failed to persist state")
	}

	return &FlowResult{
		Action:   ActionNavigate,
		NextPage: PageSelectService,
		Message:  fmt.Sprintf("Your transactions:\n%s\n\nSelect another service.", history),
		Data:     cloneMap(data),
	}, nil
}

// handleFetchBalance fetches account balance
func (s *MarketPayFlowService) handleFetchBalance(ctx context.Context, sessionID string, data map[string]string) (*FlowResult, error) {
	log.Debug().Str("session_id", sessionID).Msg("processing PageFetchBalance")

	// TODO: Call external API to fetch balance
	// For now, simulate response
	data["balance"] = "5,000,000 SLE"
	return s.navigateToNext(ctx, sessionID, data, PageShowBalance, "Checking balance, please wait")
}

// handleShowBalance displays account balance
func (s *MarketPayFlowService) handleShowBalance(ctx context.Context, sessionID string, data map[string]string) (*FlowResult, error) {
	log.Debug().Str("session_id", sessionID).Msg("processing PageShowBalance")
	
	balance := firstNonEmpty(data, "balance")
	if err := s.persistState(ctx, sessionID, data); err != nil {
		log.Warn().Err(err).Str("session_id", sessionID).Msg("failed to persist state")
	}

	return &FlowResult{
		Action:   ActionNavigate,
		NextPage: PageSelectService,
		Message:  fmt.Sprintf("Your balance: %s\n\nSelect another service.", balance),
		Data:     cloneMap(data),
	}, nil
}

// handleFetchLoanEligibility checks loan eligibility
func (s *MarketPayFlowService) handleFetchLoanEligibility(ctx context.Context, sessionID string, data map[string]string) (*FlowResult, error) {
	log.Debug().Str("session_id", sessionID).Msg("processing PageFetchLoanEligibility")

	// TODO: Call external API to check eligibility
	// For now, simulate response
	data["loan_eligible"] = "true"
	data["max_loan_amount"] = "2,000,000"
	return s.navigateToNext(ctx, sessionID, data, PageShowLoanEligibility, "Checking eligibility, please wait")
}

// handleShowLoanEligibility displays loan eligibility
func (s *MarketPayFlowService) handleShowLoanEligibility(ctx context.Context, sessionID string, data map[string]string) (*FlowResult, error) {
	log.Debug().Str("session_id", sessionID).Msg("processing PageShowLoanEligibility")
	
	eligible := firstNonEmpty(data, "loan_eligible")
	maxAmount := firstNonEmpty(data, "max_loan_amount")
	
	var message string
	if eligible == "true" {
		message = fmt.Sprintf("You are eligible for a loan.\nMax amount: %s SLE", maxAmount)
	} else {
		message = "You are not currently eligible for a loan."
	}
	
	if err := s.persistState(ctx, sessionID, data); err != nil {
		log.Warn().Err(err).Str("session_id", sessionID).Msg("failed to persist state")
	}

	return &FlowResult{
		Action:   ActionNavigate,
		NextPage: PageSelectService,
		Message:  message + "\n\nSelect another service.",
		Data:     cloneMap(data),
	}, nil
}

// handleCollectLoanAmount collects loan amount
func (s *MarketPayFlowService) handleCollectLoanAmount(ctx context.Context, sessionID string, data map[string]string) (*FlowResult, error) {
	log.Debug().Str("session_id", sessionID).Msg("processing PageCollectLoanAmount")

	loanAmountStr := strings.TrimSpace(data["loan_amount"])
	if loanAmountStr == "" {
		return s.navigateToNext(ctx, sessionID, data, PageCollectLoanAmount, "Enter loan amount in SLE\nMin: 1, Max: 5,000,000")
	}

	amount, err := strconv.ParseInt(loanAmountStr, 10, 64)
	if err != nil {
		return s.stopFlow(ctx, sessionID, data, "Please enter a valid amount")
	}

	if !ValidateAmount(amount, 1, 5000000) {
		if amount < 1 {
			return s.stopFlow(ctx, sessionID, data, GetValidationMessage("loan_amount_min"))
		}
		return s.stopFlow(ctx, sessionID, data, GetValidationMessage("loan_amount_max"))
	}

	data["loan_amount"] = loanAmountStr
	return s.navigateToNext(ctx, sessionID, data, PageConfirmLoanApplication,
		fmt.Sprintf("Submit loan application for %s SLE?\n1. Submit\n2. Cancel", loanAmountStr))
}

// handleConfirmLoanApplication handles loan application confirmation
func (s *MarketPayFlowService) handleConfirmLoanApplication(ctx context.Context, sessionID string, data map[string]string) (*FlowResult, error) {
	log.Debug().Str("session_id", sessionID).Msg("processing PageConfirmLoanApplication")

	choice := strings.TrimSpace(data["loan_confirmed"])
	if choice == "" {
		return s.navigateToNext(ctx, sessionID, data, PageConfirmLoanApplication, "Confirm application")
	}

	if choice == "cancel" || choice == "2" {
		data["loan_confirmed"] = "false"
		return s.navigateToNext(ctx, sessionID, data, PageShowLoanApplicationCancelled, "Loan application cancelled")
	}

	data["loan_confirmed"] = "true"
	return s.navigateToNext(ctx, sessionID, data, PageSubmitLoanApplication, "Submitting loan request, please wait")
}

// handleSubmitLoanApplication submits the loan application
func (s *MarketPayFlowService) handleSubmitLoanApplication(ctx context.Context, sessionID string, data map[string]string) (*FlowResult, error) {
	log.Debug().Str("session_id", sessionID).Msg("processing PageSubmitLoanApplication")

	// TODO: Call external API to submit loan application
	// For now, simulate success
	data["loan_status"] = "submitted"
	data["loan_ref"] = "LN" + sessionID[:12]

	amount := firstNonEmpty(data, "loan_amount")
	message := fmt.Sprintf("Loan application for %s SLE submitted.\nRef: %s\nYou will receive updates via SMS.", amount, data["loan_ref"])

	return s.navigateToNext(ctx, sessionID, data, PageShowLoanApplicationResult, message)
}

// handleShowLoanApplicationResult displays loan application result
func (s *MarketPayFlowService) handleShowLoanApplicationResult(ctx context.Context, sessionID string, data map[string]string) (*FlowResult, error) {
	log.Debug().Str("session_id", sessionID).Msg("processing PageShowLoanApplicationResult")
	
	if err := s.persistState(ctx, sessionID, data); err != nil {
		log.Warn().Err(err).Str("session_id", sessionID).Msg("failed to persist state")
	}

	return &FlowResult{
		Action:   ActionNavigate,
		NextPage: PageSelectService,
		Message:  "Application submitted. Select another service.",
		Data:     cloneMap(data),
	}, nil
}

// handleShowLoanApplicationCancelled displays loan application cancelled message
func (s *MarketPayFlowService) handleShowLoanApplicationCancelled(ctx context.Context, sessionID string, data map[string]string) (*FlowResult, error) {
	log.Debug().Str("session_id", sessionID).Msg("processing PageShowLoanApplicationCancelled")
	
	if err := s.persistState(ctx, sessionID, data); err != nil {
		log.Warn().Err(err).Str("session_id", sessionID).Msg("failed to persist state")
	}

	return &FlowResult{
		Action:   ActionNavigate,
		NextPage: PageSelectService,
		Message:  "Application cancelled. Select another service.",
		Data:     cloneMap(data),
	}, nil
}

// handleExitService handles exit from the service
func (s *MarketPayFlowService) handleExitService(ctx context.Context, sessionID string, data map[string]string) (*FlowResult, error) {
	log.Debug().Str("session_id", sessionID).Msg("processing PageExitService")
	log.Info().Str("session_id", sessionID).Msg("MarketPay USSD flow completed - user exiting")
	
	if err := s.persistState(ctx, sessionID, data); err != nil {
		log.Warn().Err(err).Str("session_id", sessionID).Msg("failed to persist state at exit")
	}

	return &FlowResult{
		Action:  ActionStop,
		Message: "Thanks for using MarketPay.",
		Data:    cloneMap(data),
	}, nil
}
