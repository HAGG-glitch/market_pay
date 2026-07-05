package operatorlookup

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLookup_Orange(t *testing.T) {
	prefixes := []string{"72", "73", "74", "75", "76", "78", "79"}
	for _, p := range prefixes {
		phone := "+232" + p + "123456"
		r := Lookup(phone)
		if assert.NotNil(t, r, "prefix %s should match Orange", p) {
			assert.Equal(t, "m17", r.MonimeCode)
			assert.Equal(t, "Orange", r.Company)
			assert.Equal(t, "Orange Money", r.MobileMoney)
		}
	}
}

func TestLookup_Africell(t *testing.T) {
	prefixes := []string{"88", "77", "90", "99", "30", "33"}
	for _, p := range prefixes {
		phone := "+232" + p + "123456"
		r := Lookup(phone)
		if assert.NotNil(t, r, "prefix %s should match Africell", p) {
			assert.Equal(t, "m18", r.MonimeCode)
			assert.Equal(t, "Africell", r.Company)
			assert.Equal(t, "Afrimoney", r.MobileMoney)
		}
	}
}

func TestLookup_Qcell(t *testing.T) {
	prefixes := []string{"31", "34"}
	for _, p := range prefixes {
		phone := "+232" + p + "123456"
		r := Lookup(phone)
		if assert.NotNil(t, r, "prefix %s should match Qcell", p) {
			assert.Equal(t, "m13", r.MonimeCode)
			assert.Equal(t, "Qcell", r.Company)
			assert.Equal(t, "Qcell Money", r.MobileMoney)
		}
	}
}

func TestLookup_FormatVariations(t *testing.T) {
	tests := []struct {
		input  string
		expect string
	}{
		{"+23276902971", "m17"},
		{"23276902971", "m17"},
		{"076902971", "m17"},
		{"76902971", "m17"},
	}
	for _, tt := range tests {
		r := Lookup(tt.input)
		if assert.NotNil(t, r, "input %s", tt.input) {
			assert.Equal(t, tt.expect, r.MonimeCode, "input %s", tt.input)
		}
	}
}

func TestLookup_Unknown(t *testing.T) {
	phones := []string{"+23200000000", "+23211111111", "+1234567890", ""}
	for _, p := range phones {
		assert.Nil(t, Lookup(p), "input %s", p)
	}
}

func TestLookupProviderID(t *testing.T) {
	assert.Equal(t, "m17", LookupProviderID("+23276902971"))
	assert.Equal(t, "m18", LookupProviderID("+23233902971"))
	assert.Equal(t, "m13", LookupProviderID("+23231902971"))
	assert.Equal(t, "", LookupProviderID("+23200000000"))
	assert.Equal(t, "", LookupProviderID(""))
}
