package ussd

import "context"

type FlowAction string

type FlowPage string

const (
	ActionNavigate FlowAction = "navigate"
	ActionStop     FlowAction = "stop"

	// Main menu page
	PageSelectService FlowPage = "mp_select_service"

	// Vendor registration pages
	PageCollectVendorName            FlowPage = "mp_collect_vendor_name"
	PageCollectMarketName            FlowPage = "mp_collect_market_name"
	PageSubmitVendorRegistration     FlowPage = "mp_submit_vendor_registration"
	PageShowVendorRegistrationResult FlowPage = "mp_show_vendor_registration_result"

	// Vendor payment pages
	PageCollectPaymentVendorCode FlowPage = "mp_collect_payment_vendor_code"
	PageCollectPaymentAmount     FlowPage = "mp_collect_payment_amount"
	PageConfirmPaymentChoice     FlowPage = "mp_confirm_payment_choice"
	PageSubmitVendorPayment      FlowPage = "mp_submit_vendor_payment"
	PageShowPaymentResult        FlowPage = "mp_show_payment_result"
	PageShowPaymentCancelled     FlowPage = "mp_show_payment_cancelled"

	// Transaction history pages
	PageFetchTransactionHistory FlowPage = "mp_fetch_transaction_history"
	PageShowTransactionHistory  FlowPage = "mp_show_transaction_history"

	// Balance check pages
	PageFetchBalance FlowPage = "mp_fetch_balance"
	PageShowBalance  FlowPage = "mp_show_balance"

	// Loan eligibility pages
	PageFetchLoanEligibility FlowPage = "mp_fetch_loan_eligibility"
	PageShowLoanEligibility  FlowPage = "mp_show_loan_eligibility"

	// Loan application pages
	PageCollectLoanAmount            FlowPage = "mp_collect_loan_amount"
	PageConfirmLoanApplication       FlowPage = "mp_confirm_loan_application"
	PageSubmitLoanApplication        FlowPage = "mp_submit_loan_application"
	PageShowLoanApplicationResult    FlowPage = "mp_show_loan_application_result"
	PageShowLoanApplicationCancelled FlowPage = "mp_show_loan_application_cancelled"

	// Exit page
	PageExitService FlowPage = "mp_exit_service"
)

var pageSequence = map[FlowPage]FlowPage{
	PageSelectService:                PageSelectService, // stays at menu until selection made
	PageCollectVendorName:            PageCollectMarketName,
	PageCollectMarketName:            PageSubmitVendorRegistration,
	PageSubmitVendorRegistration:     PageShowVendorRegistrationResult,
	PageShowVendorRegistrationResult: PageSelectService,
	PageCollectPaymentVendorCode:     PageCollectPaymentAmount,
	PageCollectPaymentAmount:         PageConfirmPaymentChoice,
	PageConfirmPaymentChoice:         PageSubmitVendorPayment,
	PageSubmitVendorPayment:          PageShowPaymentResult,
	PageShowPaymentResult:            PageSelectService,
	PageShowPaymentCancelled:         PageSelectService,
	PageFetchTransactionHistory:      PageShowTransactionHistory,
	PageShowTransactionHistory:       PageSelectService,
	PageFetchBalance:                 PageShowBalance,
	PageShowBalance:                  PageSelectService,
	PageFetchLoanEligibility:         PageShowLoanEligibility,
	PageShowLoanEligibility:          PageSelectService,
	PageCollectLoanAmount:            PageConfirmLoanApplication,
	PageConfirmLoanApplication:       PageSubmitLoanApplication,
	PageSubmitLoanApplication:        PageShowLoanApplicationResult,
	PageShowLoanApplicationResult:    PageSelectService,
	PageShowLoanApplicationCancelled: PageSelectService,
	PageExitService:                  PageExitService,
}

// StateStore is an application port for keeping multi-step USSD state between requests.
type StateStore interface {
	Load(ctx context.Context, sessionID string) (map[string]string, error)
	Save(ctx context.Context, sessionID string, data map[string]string) error
}

// AdvanceFlowInput is transport-agnostic input into the MarketPay USSD use case.
type AdvanceFlowInput struct {
	SessionID   string
	CurrentPage FlowPage
	Values      map[string]string
}

// FlowResult is transport-agnostic output from the MarketPay USSD use case.
type FlowResult struct {
	Action   FlowAction
	NextPage FlowPage
	Message  string
	Data     map[string]string
}

var ErrSessionNotFound = "ussd session not found"
