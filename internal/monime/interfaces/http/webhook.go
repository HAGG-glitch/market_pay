package http

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	loanapp "github.com/marketpay/backend/internal/loan/application"
	monimeinfra "github.com/marketpay/backend/internal/monime/infrastructure"
	paymentapp "github.com/marketpay/backend/internal/payment/application"
	"github.com/marketpay/backend/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// WebhookHandler handles Monime payment webhooks.
type WebhookHandler struct {
	db         *gorm.DB
	paymentSvc *paymentapp.PaymentService
	adapter    *monimeinfra.MonimeAdapter
	loanSvc    *loanapp.LoanService
	log        *logger.Logger
}

func NewWebhookHandler(db *gorm.DB, paymentSvc *paymentapp.PaymentService, adapter *monimeinfra.MonimeAdapter, loanSvc *loanapp.LoanService, log *logger.Logger) *WebhookHandler {
	return &WebhookHandler{db: db, paymentSvc: paymentSvc, adapter: adapter, loanSvc: loanSvc, log: log}
}

func (h *WebhookHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/monime/webhook", h.Handle)
	rg.POST("/monime/webhooks/payout", h.Handle)
}

type monimeWebhookPayload struct {
	Event     string `json:"event"`
	Reference string `json:"reference"`
	Status    string `json:"status"`
	Amount    string `json:"amount"`
	Phone     string `json:"phone"`
}

func (h *WebhookHandler) Handle(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

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

	var payload monimeWebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}

	if payload.Reference == "" {
		c.JSON(http.StatusOK, gin.H{"received": true})
		return
	}

	switch payload.Event {
	case "DisbursementSucceeded", "payout.completed":
		if h.loanSvc != nil {
			if err := h.loanSvc.ConfirmDisbursement(c.Request.Context(), payload.Reference); err != nil {
				h.log.Error("confirm disbursement failed", zap.Error(err))
				h.db.Exec(`UPDATE loans SET state = 'ACTIVE' WHERE monime_reference = ? AND state = 'DISBURSED'`, payload.Reference)
				h.db.Exec(`UPDATE loans SET state = 'ACTIVE' WHERE monime_reference = ? AND state = 'DISBURSEMENT_PENDING'`, payload.Reference)
			}
		} else {
			h.db.Exec(`UPDATE loans SET state = 'ACTIVE' WHERE monime_reference = ? AND state = 'DISBURSED'`, payload.Reference)
			h.db.Exec(`UPDATE loans SET state = 'ACTIVE' WHERE monime_reference = ? AND state = 'DISBURSEMENT_PENDING'`, payload.Reference)
		}
	case "DisbursementFailed", "payout.failed":
		if h.loanSvc != nil {
			failureReason := ""
			if payload.Status != "" {
				failureReason = payload.Status
			}
			if err := h.loanSvc.FailDisbursement(c.Request.Context(), payload.Reference, failureReason); err != nil {
				h.log.Error("fail disbursement failed", zap.Error(err))
				h.db.Exec(`UPDATE loans SET state = 'APPROVED', monime_reference = NULL WHERE monime_reference = ? AND state = 'DISBURSEMENT_PENDING'`, payload.Reference)
				h.db.Exec(`UPDATE loans SET state = 'APPROVED' WHERE monime_reference = ? AND state = 'DISBURSED'`, payload.Reference)
			}
		} else {
			h.db.Exec(`UPDATE loans SET state = 'APPROVED', monime_reference = NULL WHERE monime_reference = ? AND state = 'DISBURSEMENT_PENDING'`, payload.Reference)
			h.db.Exec(`UPDATE loans SET state = 'APPROVED' WHERE monime_reference = ? AND state = 'DISBURSED'`, payload.Reference)
		}
		h.db.Exec(`INSERT INTO loan_events (loan_id, event_type, payload) SELECT id, 'PAYOUT_FAILED', ? FROM loans WHERE monime_reference = ?`, string(body), payload.Reference)
	default:
		if payload.Status == "SUCCESS" {
			var paymentID uuid.UUID
			h.db.Raw(
				`SELECT id FROM payments WHERE monime_reference = ? AND status != 'SUCCESS' LIMIT 1`,
				payload.Reference,
			).Scan(&paymentID)
			if paymentID != uuid.Nil {
				h.paymentSvc.Complete(c.Request.Context(), paymentID, payload.Reference)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"received": true})
}
