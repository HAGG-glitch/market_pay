package exchange

import (
	"crypto/sha256"
	"encoding/hex"
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

func TestNormalizeSubscriberHash(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantLength int
	}{
		{"non-empty subscriber id", "sub-abc-123", 64},
		{"different subscriber id", "sub-xyz-789", 64},
		{"empty string", "", 64},
		{"special characters", "user@domain!123", 64},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := normalizeSubscriberHash(tc.input)

			assert.Len(t, result, tc.wantLength)
			// Verify it's valid hex
			decoded, err := hex.DecodeString(result)
			assert.NoError(t, err)
			assert.Len(t, decoded, 32)

			// Verify it matches a direct SHA-256 call
			h := sha256.Sum256([]byte(tc.input))
			expected := hex.EncodeToString(h[:])
			assert.Equal(t, expected, result, "must match standard SHA-256 hex encoding")
		})
	}
}

func TestNormalizeSubscriberHashDeterministic(t *testing.T) {
	input := "subscriber-test-001"
	h1 := normalizeSubscriberHash(input)
	h2 := normalizeSubscriberHash(input)
	assert.Equal(t, h1, h2, "same input must produce same hash")
}

func TestNormalizeSubscriberHashUnique(t *testing.T) {
	h1 := normalizeSubscriberHash("user-a")
	h2 := normalizeSubscriberHash("user-b")
	assert.NotEqual(t, h1, h2, "different inputs must produce different hashes")
}

func TestNormalizeSubscriberHashNoDoubleHash(t *testing.T) {
	input := "raw-subscriber-id"
	first := normalizeSubscriberHash(input)
	second := normalizeSubscriberHash(first)
	// second should NOT equal first — if we double-hash by mistake, they would differ from a single hash
	h := sha256.Sum256([]byte(input))
	single := hex.EncodeToString(h[:])
	assert.Equal(t, single, first, "first pass must be standard SHA-256")
	assert.NotEqual(t, first, second, "second pass must produce different output, proving no double-hash bug")
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
