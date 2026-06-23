package http

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	loanapp "github.com/marketpay/backend/internal/loan/application"
	monimeinfra "github.com/marketpay/backend/internal/monime/infrastructure"
	paymentapp "github.com/marketpay/backend/internal/payment/application"
	repayapp "github.com/marketpay/backend/internal/repayment/application"
	"github.com/marketpay/backend/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type WebhookHandler struct {
	db         *gorm.DB
	paymentSvc *paymentapp.PaymentService
	adapter    *monimeinfra.MonimeAdapter
	loanSvc    *loanapp.LoanService
	repaySvc   *repayapp.RepaymentService
	log        *logger.Logger
}

func NewWebhookHandler(db *gorm.DB, paymentSvc *paymentapp.PaymentService, adapter *monimeinfra.MonimeAdapter, loanSvc *loanapp.LoanService, repaySvc *repayapp.RepaymentService, log *logger.Logger) *WebhookHandler {
	return &WebhookHandler{db: db, paymentSvc: paymentSvc, adapter: adapter, loanSvc: loanSvc, repaySvc: repaySvc, log: log}
}

func (h *WebhookHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/monime/webhook", h.Handle)
	rg.POST("/monime/webhooks/payout", h.Handle)
}

// monimeWebhookV2 matches the actual Monime v2 webhook payload format.
type monimeWebhookV2 struct {
	APIVersion string          `json:"apiVersion"`
	Event      monimeWebhookEvent `json:"event"`
	Object     monimeWebhookObject `json:"object"`
	Data       json.RawMessage    `json:"data"`
}

type monimeWebhookEvent struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Timestamp string `json:"timestamp"`
}

type monimeWebhookObject struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

// payoutWebhookData is the data payload for payout events.
type payoutWebhookData struct {
	Amount         *payoutAmount    `json:"amount,omitempty"`
	Destination    *payoutDestination `json:"destination,omitempty"`
	FailureDetail  *payoutFailure   `json:"failureDetail,omitempty"`
	Status         string           `json:"status"`
	Source         *payoutSource    `json:"source,omitempty"`
}

type payoutAmount struct {
	Currency string  `json:"currency"`
	Value    float64 `json:"value"`
}

type payoutDestination struct {
	PhoneNumber          string  `json:"phoneNumber"`
	ProviderID           string  `json:"providerId"`
	TransactionReference *string `json:"transactionReference"`
	Type                 string  `json:"type"`
}

type payoutFailure struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type payoutSource struct {
	FinancialAccountID   string `json:"financialAccountId"`
	TransactionReference string `json:"transactionReference"`
}

func (h *WebhookHandler) Handle(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	// ── Signature validation (best-effort) ───────────────────────────────
	signature := c.GetHeader("X-Monime-Signature")
	if h.adapter != nil {
		secret := h.adapter.GetWebhookSecret()
		if secret == "" {
			h.log.Warn("webhook secret is empty — skipping signature validation")
		} else if !h.adapter.ValidateWebhook(body, signature) {
			h.log.Warn("webhook signature validation failed",
				zap.String("signature", signature),
				zap.ByteString("body", body),
			)
		}
	}

	// ── Parse the v2 webhook ─────────────────────────────────────────────
	var v2 monimeWebhookV2
	if err := json.Unmarshal(body, &v2); err != nil {
		h.log.Error("webhook parse failed", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}

	eventName := v2.Event.Name
	payoutID := v2.Object.ID

	h.log.Info("webhook received",
		zap.String("event", eventName),
		zap.String("payout_id", payoutID),
	)

	switch eventName {
	case "payout.completed":
		h.handlePayoutCompleted(c, payoutID, v2.Data)
	case "payout.failed":
		h.handlePayoutFailed(c, payoutID, v2.Data)
	case "financial_account.credited":
		h.handleAccountCredited(c, payoutID, v2.Data)
	default:
		h.log.Debug("ignoring webhook event", zap.String("event", eventName))
	}

	c.JSON(http.StatusOK, gin.H{"received": true})
}

func (h *WebhookHandler) handlePayoutCompleted(c *gin.Context, payoutID string, rawData json.RawMessage) {
	var data payoutWebhookData
	if rawData != nil {
		_ = json.Unmarshal(rawData, &data)
	}

	providerRef := ""
	if data.Destination != nil && data.Destination.TransactionReference != nil {
		providerRef = *data.Destination.TransactionReference
	}

	if h.loanSvc != nil {
		if err := h.loanSvc.ConfirmDisbursement(c.Request.Context(), payoutID); err != nil {
			h.log.Error("confirm disbursement failed (fallback to raw SQL)", zap.Error(err), zap.String("payout_id", payoutID))
			h.db.Exec(`UPDATE loans SET state = 'ACTIVE', provider_ref = ? WHERE monime_reference = ? AND state = 'DISBURSEMENT_PENDING'`, providerRef, payoutID)
			h.db.Exec(`UPDATE loans SET state = 'ACTIVE', provider_ref = ? WHERE monime_reference = ? AND state = 'DISBURSED'`, providerRef, payoutID)
		} else if providerRef != "" {
			h.db.Exec(`UPDATE loans SET provider_ref = ? WHERE monime_reference = ?`, providerRef, payoutID)
		}
	} else {
		h.db.Exec(`UPDATE loans SET state = 'ACTIVE', provider_ref = ? WHERE monime_reference = ? AND state = 'DISBURSEMENT_PENDING'`, providerRef, payoutID)
		h.db.Exec(`UPDATE loans SET state = 'ACTIVE', provider_ref = ? WHERE monime_reference = ? AND state = 'DISBURSED'`, providerRef, payoutID)
	}
}

func (h *WebhookHandler) handlePayoutFailed(c *gin.Context, payoutID string, rawData json.RawMessage) {
	var data payoutWebhookData
	failureReason := ""
	if rawData != nil {
		_ = json.Unmarshal(rawData, &data)
		if data.FailureDetail != nil {
			failureReason = data.FailureDetail.Code + ": " + data.FailureDetail.Message
		}
	}

	h.log.Warn("payout failed",
		zap.String("payout_id", payoutID),
		zap.String("reason", failureReason),
	)

	if h.loanSvc != nil {
		if err := h.loanSvc.FailDisbursement(c.Request.Context(), payoutID, failureReason); err != nil {
			h.log.Error("fail disbursement failed (fallback to raw SQL)", zap.Error(err), zap.String("payout_id", payoutID))
			h.db.Exec(`UPDATE loans SET state = 'APPROVED', monime_reference = NULL, payout_id = NULL, provider_ref = NULL, due_date = NULL, failure_reason = ? WHERE monime_reference = ? AND state IN ('DISBURSEMENT_PENDING','ACTIVE')`, failureReason, payoutID)
			h.db.Exec(`UPDATE loans SET state = 'APPROVED', failure_reason = ? WHERE monime_reference = ? AND state = 'DISBURSED'`, failureReason, payoutID)
			h.db.Exec(`DELETE FROM repayment_schedules WHERE loan_id IN (SELECT id FROM loans WHERE monime_reference = ? AND state = 'APPROVED')`, payoutID)
		}
	} else {
		h.db.Exec(`UPDATE loans SET state = 'APPROVED', monime_reference = NULL, payout_id = NULL, provider_ref = NULL, due_date = NULL, failure_reason = ? WHERE monime_reference = ? AND state IN ('DISBURSEMENT_PENDING','ACTIVE')`, failureReason, payoutID)
		h.db.Exec(`UPDATE loans SET state = 'APPROVED', failure_reason = ? WHERE monime_reference = ? AND state = 'DISBURSED'`, failureReason, payoutID)
		h.db.Exec(`DELETE FROM repayment_schedules WHERE loan_id IN (SELECT id FROM loans WHERE monime_reference = ? AND state = 'APPROVED')`, payoutID)
	}

}

// creditWebhookData is the data payload for financial_account.credited events.
type creditWebhookData struct {
	Amount   *creditAmount   `json:"amount,omitempty"`
	Source   *creditSource   `json:"source,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

type creditAmount struct {
	Currency string  `json:"currency"`
	Value    float64 `json:"value"`
}

type creditSource struct {
	PhoneNumber          string `json:"phoneNumber"`
	ProviderID           string `json:"providerId"`
	TransactionReference string `json:"transactionReference"`
	Type                 string `json:"type"`
}

func (h *WebhookHandler) handleAccountCredited(c *gin.Context, collectionRef string, rawData json.RawMessage) {
	var data creditWebhookData
	if rawData != nil {
		_ = json.Unmarshal(rawData, &data)
	}

	h.log.Info("account credited",
		zap.String("collection_ref", collectionRef),
		zap.Float64("amount", safeAmount(data.Amount)),
	)

	// Try matching by Monime collection reference
	if err := h.repaySvc.ConfirmRepayment(c.Request.Context(), collectionRef); err == nil {
		h.log.Info("repayment confirmed via monime_ref", zap.String("collection_ref", collectionRef))
		return
	}

	// Fallback: try matching by payment_ref in metadata
	if ref, ok := data.Metadata["payment_ref"].(string); ok && ref != "" {
		if err := h.repaySvc.ConfirmRepayment(c.Request.Context(), ref); err == nil {
			h.log.Info("repayment confirmed via metadata.payment_ref", zap.String("payment_ref", ref))
			return
		}
		h.log.Warn("no repayment record matched", zap.String("collection_ref", collectionRef), zap.String("payment_ref", ref))
	} else {
		h.log.Warn("no repayment matched by collection_ref and no payment_ref in metadata", zap.String("collection_ref", collectionRef))
	}
}

func safeAmount(a *creditAmount) float64 {
	if a == nil {
		return 0
	}
	return a.Value
}

// Payment webhook handling (legacy — collection webhooks).
func (h *WebhookHandler) handleLegacyPayment(body []byte, status string, reference string) {
	if status == "SUCCESS" && reference != "" {
		var paymentID uuid.UUID
		h.db.Raw(
			`SELECT id FROM payments WHERE monime_reference = ? AND status != 'SUCCESS' LIMIT 1`,
			reference,
		).Scan(&paymentID)
		if paymentID != uuid.Nil {
			_, _ = h.paymentSvc.Complete(contextTODO(), paymentID, reference)
		}
	}
}

func contextTODO() context.Context {
	// minimal context for db operations during webhook
	return context.Background()
}
