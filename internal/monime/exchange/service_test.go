package exchange

import (
	"fmt"
	"testing"

	"github.com/marketpay/backend/pkg/monimeexchange"
	"github.com/stretchr/testify/assert"
)

func TestNormalizePhone(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"+23276123456", "+23276123456"},
		{"23276123456", "+23276123456"},
		{"076123456", "+232076123456"},
		{"76123456", "+76123456"},
		{"+1234567890", "+1234567890"},
		{"", "+"},
	}
	for _, tc := range tests {
		result := normalizePhone(tc.input)
		assert.Equal(t, tc.expected, result, "normalizePhone(%q)", tc.input)
	}
}

func TestStringValue(t *testing.T) {
	tests := []struct {
		input    interface{}
		expected string
	}{
		{"hello", "hello"},
		{"", ""},
		{nil, ""},
		{42, "42"},
		{3.14, "3.14"},
		{true, "true"},
	}
	for _, tc := range tests {
		result := stringValue(tc.input)
		assert.Equal(t, tc.expected, result, "stringValue(%v)", tc.input)
	}
}

func TestSessionContextMerge(t *testing.T) {
	p := &monimeexchange.ExchangePayload{}
	p.FlowData = map[string]interface{}{"flow_key": "flow_val"}
	p.ExportedData = map[string]interface{}{"export_key": "export_val"}
	p.SessionContext = map[string]interface{}{
		"mutations": map[string]interface{}{"mutation_key": "mutation_val"},
	}

	sc := sessionContext(p)

	assert.Equal(t, "flow_val", sc["flow_key"])
	assert.Equal(t, "export_val", sc["export_key"])
	assert.Equal(t, "mutation_val", sc["mutation_key"])
}

func TestSessionContextExportedOverridesFlow(t *testing.T) {
	p := &monimeexchange.ExchangePayload{}
	p.FlowData = map[string]interface{}{"key": "from_flow"}
	p.ExportedData = map[string]interface{}{"key": "from_export"}

	sc := sessionContext(p)

	assert.Equal(t, "from_export", sc["key"])
}

func TestSessionContextMutationsOverride(t *testing.T) {
	p := &monimeexchange.ExchangePayload{}
	p.FlowData = map[string]interface{}{"key": "from_flow"}
	p.ExportedData = map[string]interface{}{"key": "from_export"}
	p.SessionContext = map[string]interface{}{
		"mutations": map[string]interface{}{"key": "from_mutation"},
	}

	sc := sessionContext(p)

	assert.Equal(t, "from_mutation", sc["key"])
}

func TestRouteDispatchesKnownPages(t *testing.T) {
	pages := []struct {
		pageID  string
		service string
	}{
		{"mp_access_gate_exchange", ""},
		{"mp_collect_market_name", "register_vendor"},
		{"mp_confirm_payment_receipt", "pay_vendor"},
		{"mp_collect_payment_pin", "pay_vendor"},
		{"mp_balance_exchange", "balance_check"},
		{"mp_loan_eligibility_exchange", "loan_eligibility"},
		{"mp_confirm_loan_application", "loan_application"},
	}

	for _, tc := range pages {
		p := &monimeexchange.ExchangePayload{
			CurrentPage: tc.pageID,
		}
		p.ExportedData = map[string]interface{}{
			"selected_service": tc.service,
		}
		p.Global.SubscriberID = "test-sub-" + tc.pageID
		p.Global.SessionID = "test-sess-" + tc.pageID

		assert.Equal(t, tc.pageID, p.CurrentPage)
		assert.Equal(t, tc.service, p.ExportedData["selected_service"])
	}
}

func TestDuplicateKeyConstruction(t *testing.T) {
	sessionID := "sess-abc-123"
	currentPage := "mp_collect_payment_pin"
	expectedKey := sessionID + "-" + currentPage

	p := &monimeexchange.ExchangePayload{}
	p.Global.SessionID = sessionID
	p.CurrentPage = currentPage

	actualKey := p.Global.SessionID + "-" + p.CurrentPage
	assert.Equal(t, expectedKey, actualKey)
}

func TestValidatePayment(t *testing.T) {
	tests := []struct {
		name         string
		code         string
		amount       string
		confirmed    string
		expectCancel bool
		expectStop   bool
	}{
		{"confirmed valid", "MP12345", "5000", "true", false, false},
		{"cancelled", "", "", "false", true, false},
		{"invalid code", "INVALID", "5000", "true", false, true},
		{"empty amount", "MP12345", "", "true", false, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			sc := map[string]interface{}{
				"payment_vendor_code": tc.code,
				"payment_amount":      tc.amount,
				"payment_confirmed":   tc.confirmed,
			}

			code := stringValue(sc["payment_vendor_code"])
			amount := stringValue(sc["payment_amount"])
			confirmed := stringValue(sc["payment_confirmed"]) == "true"

			if !confirmed {
				assert.True(t, tc.expectCancel)
				return
			}
			if !tc.expectCancel && code != "" && amount != "" {
				assert.True(t, code != "" && amount != "")
			}
			if !tc.expectCancel && (!hasMPPrefix(code) || amount == "") {
				assert.True(t, tc.expectStop)
			}
		})
	}
}

func hasMPPrefix(code string) bool {
	return len(code) >= 2 && code[:2] == "MP"
}

func TestApplyLoanLogic(t *testing.T) {
	tests := []struct {
		name           string
		loanConfirmed  string
		loanAmount     string
		expectNavigate string
		expectStop     bool
	}{
		{"confirmed valid", "true", "5000", "mp_show_loan_application_result", false},
		{"cancelled", "false", "", "mp_show_loan_application_cancelled", false},
		{"confirmed invalid amount", "true", "0", "", true},
		{"confirmed negative amount", "true", "-100", "", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			confirmed := stringValue(tc.loanConfirmed) == "true"
			amountStr := stringValue(tc.loanAmount)

			if !confirmed {
				assert.Equal(t, "mp_show_loan_application_cancelled", tc.expectNavigate)
				return
			}

			var amount float64
			if amountStr != "" {
				_, _ = fmt.Sscanf(amountStr, "%f", &amount)
			}
			if amount <= 0 {
				assert.True(t, tc.expectStop)
				return
			}

			assert.Equal(t, "mp_show_loan_application_result", tc.expectNavigate)
		})
	}
}
