package ussd

import (
	"regexp"
)

// ValidateVendorName validates vendor name format
func ValidateVendorName(name string) bool {
	if len(name) < 2 || len(name) > 50 {
		return false
	}
	// Pattern: ^[A-Za-z0-9][A-Za-z0-9 ,.&-]{1,49}$
	pattern := `^[A-Za-z0-9][A-Za-z0-9 ,.&-]{1,49}$`
	matched, _ := regexp.MatchString(pattern, name)
	return matched
}

// ValidateMarketName validates market name format
func ValidateMarketName(name string) bool {
	if len(name) < 2 || len(name) > 50 {
		return false
	}
	// Pattern: ^[A-Za-z0-9][A-Za-z0-9 ,.&-]{1,49}$
	pattern := `^[A-Za-z0-9][A-Za-z0-9 ,.&-]{1,49}$`
	matched, _ := regexp.MatchString(pattern, name)
	return matched
}

// ValidateVendorCode validates vendor code format
func ValidateVendorCode(code string) bool {
	if len(code) < 7 || len(code) > 12 {
		return false
	}
	// Pattern: ^MP[0-9]{5,10}$
	pattern := `^MP[0-9]{5,10}$`
	matched, _ := regexp.MatchString(pattern, code)
	return matched
}

// ValidateAmount validates payment amount
func ValidateAmount(amount int64, minValue, maxValue int64) bool {
	return amount >= minValue && amount <= maxValue
}

// GetValidationMessage returns appropriate error message for validation failure
func GetValidationMessage(validationType string) string {
	messages := map[string]string{
		"vendor_name":     "Enter a valid name",
		"market_name":     "Enter a valid market name",
		"vendor_code":     "Use a code like MP12345",
		"amount_min":      "Amount must be at least 1 SLE",
		"amount_max":      "Amount exceeds maximum limit",
		"loan_amount_min": "Loan amount must be at least 1 SLE",
		"loan_amount_max": "Loan amount exceeds maximum limit",
	}
	if msg, exists := messages[validationType]; exists {
		return msg
	}
	return "Invalid input"
}
