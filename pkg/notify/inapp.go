package notify

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RoleTargets maps outbox event types to in-app notification recipients.
var RoleTargets = map[string][]string{
	"VendorCreated":       {"LOAN_OFFICER"},
	"VendorRegistered":    {"LOAN_OFFICER"},
	"GroupCreated":        {"LOAN_OFFICER"},
	"LoanApplied":         {"LOAN_OFFICER", "MFI_PARTNER"},
	"LoanRequested":       {"LOAN_OFFICER", "MFI_PARTNER"},
	"LoanApproved":        {"VENDOR"},
	"LoanRejected":        {"VENDOR"},
	"RepaymentReceived":   {"LOAN_OFFICER", "VENDOR"},
	"AccountFrozen":       {"FIELD_AGENT"},
	"AccountUnfrozen":     {"FIELD_AGENT"},
	"GroupFrozen":         {"FIELD_AGENT", "LOAN_OFFICER"},
	"PaymentCompleted":    {"VENDOR", "CUSTOMER"},
}

// Titles for in-app notifications.
func Title(eventType string) string {
	switch eventType {
	case "VendorCreated", "VendorRegistered":
		return "New Vendor Registered"
	case "GroupCreated":
		return "New Group Created"
	case "LoanApplied", "LoanRequested":
		return "Loan Application"
	case "LoanApproved":
		return "Loan Approved"
	case "LoanRejected":
		return "Loan Rejected"
	case "RepaymentReceived":
		return "Repayment Received"
	case "AccountFrozen":
		return "Account Frozen"
	case "AccountUnfrozen":
		return "Account Unfrozen"
	case "GroupFrozen":
		return "Group Frozen"
	case "PaymentCompleted":
		return "Payment Completed"
	default:
		return eventType
	}
}

// BodyFromPayload builds a short notification body from event payload.
func BodyFromPayload(eventType string, payload string) string {
	var data map[string]interface{}
	_ = json.Unmarshal([]byte(payload), &data)
	switch eventType {
	case "VendorCreated", "VendorRegistered":
		if name, ok := data["name"].(string); ok {
			return "Vendor " + name + " registered for review"
		}
	case "GroupCreated":
		if name, ok := data["name"].(string); ok {
			return "Group \"" + name + "\" created"
		}
	case "LoanApplied", "LoanRequested":
		if amt, ok := data["amount"].(float64); ok {
			return "New loan request for " + formatAmt(amt)
		}
	case "AccountFrozen":
		if reason, ok := data["reason"].(string); ok && reason != "" {
			return "Account frozen: " + reason
		}
		return "A vendor account has been frozen"
	case "RepaymentReceived":
		if amt, ok := data["amount"].(float64); ok {
			return "Repayment of " + formatAmt(amt) + " received"
		}
	}
	return Title(eventType)
}

func formatAmt(v float64) string {
	return fmt.Sprintf("%.2f SLE", v)
}

// DispatchOutboxEvent creates in_app_notifications for role-targeted events.
func DispatchOutboxEvent(ctx context.Context, db *gorm.DB, eventType string, payload string, isDemo bool) error {
	roles, ok := RoleTargets[eventType]
	if !ok {
		return nil
	}
	title := Title(eventType)
	body := BodyFromPayload(eventType, payload)

	for _, role := range roles {
		var userIDs []uuid.UUID
		db.WithContext(ctx).Raw(
			`SELECT id FROM users WHERE role = ? AND is_active = true AND is_demo = ?`,
			role, isDemo,
		).Scan(&userIDs)
		for _, id := range userIDs {
			db.WithContext(ctx).Exec(
				`INSERT INTO in_app_notifications (recipient_id, event_type, title, body, is_demo)
				 VALUES (?, ?, ?, ?, ?)`,
				id, eventType, title, body, isDemo,
			)
		}
	}
	return nil
}

// IsDemoFromPayload extracts is_demo from outbox payload when present.
func IsDemoFromPayload(payload string) bool {
	var data map[string]interface{}
	if json.Unmarshal([]byte(payload), &data) != nil {
		return false
	}
	if v, ok := data["is_demo"].(bool); ok {
		return v
	}
	return false
}
