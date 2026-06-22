package http

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	monimeinfra "github.com/marketpay/backend/internal/monime/infrastructure"
	paymentapp "github.com/marketpay/backend/internal/payment/application"
	"gorm.io/gorm"
)

// WebhookHandler handles Monime payment webhooks.
type WebhookHandler struct {
	db         *gorm.DB
	paymentSvc *paymentapp.PaymentService
	adapter    *monimeinfra.MonimeAdapter
}

func NewWebhookHandler(db *gorm.DB, paymentSvc *paymentapp.PaymentService, adapter *monimeinfra.MonimeAdapter) *WebhookHandler {
	return &WebhookHandler{db: db, paymentSvc: paymentSvc, adapter: adapter}
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
	if h.adapter != nil && !h.adapter.ValidateWebhook(body, signature) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid signature"})
		return
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
		h.db.Exec(`UPDATE loans SET state = 'ACTIVE' WHERE monime_reference = ? AND state = 'DISBURSED'`, payload.Reference)
	case "DisbursementFailed", "payout.failed":
		h.db.Exec(`UPDATE loans SET state = 'DISBURSED' WHERE monime_reference = ? AND state = 'ACTIVE'`, payload.Reference)
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
