package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// AppError represents a structured application error.
// The Code field is for internal/logging use only — never shown to end users.
type AppError struct {
	Code    string `json:"-"`
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
		return New("NOT_FOUND", fmt.Sprintf("We couldn't find this %s. It may have been removed or the link may be incorrect.", resource), http.StatusNotFound)
	}

	ErrAlreadyExists = func(resource string) *AppError {
		return New("ALREADY_EXISTS", fmt.Sprintf("This %s is already registered. Please use a different one or sign in instead.", resource), http.StatusConflict)
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
		return Wrap(err, "INTERNAL_ERROR", "Something went wrong on our end. Please try again in a moment.", http.StatusInternalServerError)
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
			fmt.Sprintf("This loan cannot move from '%s' to '%s'. Only certain status changes are allowed at this stage.", from, to),
			http.StatusBadRequest)
	}

	ErrInsufficientCreditScore = New("INSUFFICIENT_CREDIT_SCORE",
		"This vendor's credit score is too low to qualify for a loan at this time.", http.StatusBadRequest)

	ErrGroupFrozen = New("GROUP_FROZEN",
		"This group has been frozen due to a member default. No actions can be taken until it is unfrozen.", http.StatusBadRequest)

	ErrGroupFull = New("GROUP_FULL",
		"This group has reached its maximum number of members. No more vendors can be added.", http.StatusBadRequest)

	ErrGroupMinSize = New("GROUP_MIN_SIZE",
		"This group does not have enough members yet. Please add more vendors to meet the minimum size.", http.StatusBadRequest)

	ErrVendorNotEligible = New("VENDOR_NOT_ELIGIBLE",
		"This vendor does not meet the requirements to apply for a loan. Please check their KYC status and transaction history.", http.StatusBadRequest)

	ErrInsufficientTransactionHistory = New("INSUFFICIENT_TRANSACTION_HISTORY",
		"This vendor needs at least 30 days of transaction history before they can apply for a loan.", http.StatusBadRequest)

	ErrInvalidPIN = New("INVALID_PIN",
		"The PIN you entered is incorrect. Please try again.", http.StatusUnauthorized)

	ErrUSSDSessionExpired = New("SESSION_EXPIRED",
		"Your USSD session has expired. Please start again from the beginning.", http.StatusGone)

	ErrLedgerUnbalanced = New("LEDGER_UNBALANCED",
		"The transaction could not be completed because the journal entries are unbalanced. Please contact support.", http.StatusBadRequest)

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
