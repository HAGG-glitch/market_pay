package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// AppError represents a structured application error.
type AppError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Status  int    `json:"-"`
	Err     error  `json:"-"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// Standard error constructors.
func New(code, message string, status int) *AppError {
	return &AppError{Code: code, Message: message, Status: status}
}

func Wrap(err error, code, message string, status int) *AppError {
	return &AppError{Code: code, Message: message, Status: status, Err: err}
}

// Common domain errors.
var (
	ErrNotFound = func(resource string) *AppError {
		return New("NOT_FOUND", fmt.Sprintf("%s not found", resource), http.StatusNotFound)
	}

	ErrAlreadyExists = func(resource string) *AppError {
		return New("ALREADY_EXISTS", fmt.Sprintf("%s already exists", resource), http.StatusConflict)
	}

	ErrUnauthorized = func(msg string) *AppError {
		return New("UNAUTHORIZED", msg, http.StatusUnauthorized)
	}

	ErrForbidden = func(msg string) *AppError {
		return New("FORBIDDEN", msg, http.StatusForbidden)
	}

	ErrBadRequest = func(msg string) *AppError {
		return New("BAD_REQUEST", msg, http.StatusBadRequest)
	}

	ErrInternalServer = func(err error) *AppError {
		return Wrap(err, "INTERNAL_ERROR", "An internal server error occurred", http.StatusInternalServerError)
	}

	ErrValidation = func(msg string) *AppError {
		return New("VALIDATION_ERROR", msg, http.StatusUnprocessableEntity)
	}

	ErrConflict = func(msg string) *AppError {
		return New("CONFLICT", msg, http.StatusConflict)
	}

	ErrUnprocessable = func(msg string) *AppError {
		return New("UNPROCESSABLE", msg, http.StatusUnprocessableEntity)
	}
)

// Business rule errors.
var (
	ErrInvalidLoanState = func(from, to string) *AppError {
		return New("INVALID_LOAN_TRANSITION",
			fmt.Sprintf("cannot transition loan from %s to %s", from, to),
			http.StatusBadRequest)
	}

	ErrInsufficientCreditScore = New("INSUFFICIENT_CREDIT_SCORE",
		"credit score does not meet minimum requirement", http.StatusBadRequest)

	ErrGroupFrozen = New("GROUP_FROZEN",
		"group is frozen due to member default", http.StatusBadRequest)

	ErrGroupFull = New("GROUP_FULL",
		"group has reached maximum member limit", http.StatusBadRequest)

	ErrGroupMinSize = New("GROUP_MIN_SIZE",
		"group does not meet minimum size requirement", http.StatusBadRequest)

	ErrVendorNotEligible = New("VENDOR_NOT_ELIGIBLE",
		"vendor does not meet loan eligibility requirements", http.StatusBadRequest)

	ErrInsufficientTransactionHistory = New("INSUFFICIENT_TRANSACTION_HISTORY",
		"vendor requires at least 30 days of transaction history", http.StatusBadRequest)

	ErrInvalidPIN = New("INVALID_PIN", "invalid PIN", http.StatusUnauthorized)

	ErrUSSDSessionExpired = New("SESSION_EXPIRED", "USSD session has expired", http.StatusGone)

	ErrLedgerUnbalanced = New("LEDGER_UNBALANCED",
		"journal entries do not balance", http.StatusBadRequest)

	ErrInvalidAmount = func(msg string) *AppError {
		return New("INVALID_AMOUNT", msg, http.StatusBadRequest)
	}
)

// IsNotFound checks if the error is a not-found error.
func IsNotFound(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Status == http.StatusNotFound
	}
	return false
}

// IsConflict checks if the error is a conflict error.
func IsConflict(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Status == http.StatusConflict
	}
	return false
}

// HTTPStatus extracts the HTTP status from an error.
func HTTPStatus(err error) int {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Status
	}
	return http.StatusInternalServerError
}
