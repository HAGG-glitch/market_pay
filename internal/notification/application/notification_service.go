package application

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	shared "github.com/marketpay/backend/internal/shared/domain/model"
	"github.com/marketpay/backend/pkg/logger"
	"go.uber.org/zap"
)

// NotificationRepository persists notifications.
type NotificationRepository interface {
	Save(ctx context.Context, n *shared.Notification) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status string, errMsg string) error
}

// SMSSender sends SMS messages.
type SMSSender interface {
	Send(ctx context.Context, phone, body string) error
}

// WhatsAppSender sends WhatsApp messages.
type WhatsAppSender interface {
	Send(ctx context.Context, phone, body string) error
}

// EmailSender sends email messages.
type EmailSender interface {
	Send(ctx context.Context, to, subject, body string) error
}

// NotificationService dispatches notifications across channels.
// Priority: SMS first, then WhatsApp, then Email.
type NotificationService struct {
	repo      NotificationRepository
	sms       SMSSender
	whatsapp  WhatsAppSender
	email     EmailSender
	log       *logger.Logger
}

// NewNotificationService constructs a NotificationService.
func NewNotificationService(
	repo NotificationRepository,
	sms SMSSender,
	whatsapp WhatsAppSender,
	email EmailSender,
	log *logger.Logger,
) *NotificationService {
	return &NotificationService{repo: repo, sms: sms, whatsapp: whatsapp, email: email, log: log}
}

// SendInput holds notification dispatch data.
type SendInput struct {
	RecipientID    uuid.UUID
	RecipientPhone string
	RecipientEmail string
	EventType      string
	Subject        string
	Body           string
}

// Send dispatches a notification using the highest-priority available channel.
func (s *NotificationService) Send(ctx context.Context, input SendInput) error {
	// Try SMS first
	if input.RecipientPhone != "" && s.sms != nil {
		return s.sendViaSMS(ctx, input)
	}
	// Fall back to WhatsApp
	if input.RecipientPhone != "" && s.whatsapp != nil {
		return s.sendViaWhatsApp(ctx, input)
	}
	// Fall back to email
	if input.RecipientEmail != "" && s.email != nil {
		return s.sendViaEmail(ctx, input)
	}
	s.log.Warn("no channel available for notification",
		zap.String("recipient_id", input.RecipientID.String()),
		zap.String("event_type", input.EventType),
	)
	return fmt.Errorf("no notification channel available for recipient %s", input.RecipientID)
}

func (s *NotificationService) sendViaSMS(ctx context.Context, input SendInput) error {
	n := &shared.Notification{
		RecipientID:    input.RecipientID,
		RecipientPhone: input.RecipientPhone,
		Channel:        "sms",
		EventType:      input.EventType,
		Body:           input.Body,
		Status:         "PENDING",
	}
	_ = s.repo.Save(ctx, n)

	if err := s.sms.Send(ctx, input.RecipientPhone, input.Body); err != nil {
		_ = s.repo.UpdateStatus(ctx, n.ID, "FAILED", err.Error())
		s.log.Error("SMS send failed", zap.Error(err))
		return err
	}

	_ = s.repo.UpdateStatus(ctx, n.ID, "SENT", "")
	return nil
}

func (s *NotificationService) sendViaWhatsApp(ctx context.Context, input SendInput) error {
	n := &shared.Notification{
		RecipientID:    input.RecipientID,
		RecipientPhone: input.RecipientPhone,
		Channel:        "whatsapp",
		EventType:      input.EventType,
		Body:           input.Body,
		Status:         "PENDING",
	}
	_ = s.repo.Save(ctx, n)

	if err := s.whatsapp.Send(ctx, input.RecipientPhone, input.Body); err != nil {
		_ = s.repo.UpdateStatus(ctx, n.ID, "FAILED", err.Error())
		return err
	}
	_ = s.repo.UpdateStatus(ctx, n.ID, "SENT", "")
	return nil
}

func (s *NotificationService) sendViaEmail(ctx context.Context, input SendInput) error {
	n := &shared.Notification{
		RecipientID:    input.RecipientID,
		RecipientEmail: input.RecipientEmail,
		Channel:        "email",
		EventType:      input.EventType,
		Subject:        input.Subject,
		Body:           input.Body,
		Status:         "PENDING",
	}
	_ = s.repo.Save(ctx, n)

	if err := s.email.Send(ctx, input.RecipientEmail, input.Subject, input.Body); err != nil {
		_ = s.repo.UpdateStatus(ctx, n.ID, "FAILED", err.Error())
		return err
	}
	_ = s.repo.UpdateStatus(ctx, n.ID, "SENT", "")
	return nil
}

// Templates returns the body for each known event type.
func NotificationBody(eventType string, data map[string]string) string {
	switch eventType {
	case "VendorRegistered":
		return fmt.Sprintf("Welcome to MarketPay, %s! Your account has been created. Dial *737# to get started.", data["name"])
	case "LoanApproved":
		return fmt.Sprintf("Congratulations! Your loan of %s SLE has been approved. Disbursement is being processed.", data["amount"])
	case "LoanRejected":
		return fmt.Sprintf("Your loan application has been declined. Reason: %s. Contact support for assistance.", data["reason"])
	case "LoanDisbursed":
		return fmt.Sprintf("Your loan of %s SLE has been disbursed to your mobile money account. Reference: %s", data["amount"], data["reference"])
	case "RepaymentDue":
		return fmt.Sprintf("Reminder: Your loan repayment of %s SLE is due on %s. Dial *737# to repay.", data["amount"], data["due_date"])
	case "RepaymentReceived":
		return fmt.Sprintf("Payment received: %s SLE. Outstanding balance: %s SLE. Thank you!", data["amount"], data["outstanding"])
	case "LatePaymentWarning":
		return fmt.Sprintf("Your loan payment of %s SLE is overdue. Please repay immediately to avoid penalties.", data["amount"])
	case "LoanDefaulted":
		return fmt.Sprintf("Your loan has been marked as defaulted. Please contact MarketPay support immediately.")
	case "GroupMemberDefaulted":
		return fmt.Sprintf("A member of your group has defaulted. Your group has been temporarily frozen.")
	default:
		return fmt.Sprintf("MarketPay notification: %s", eventType)
	}
}
