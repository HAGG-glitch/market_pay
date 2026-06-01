package http

import (
	"context"
	"encoding/json"
	"net/http"

	"marketpay/internal/ussd"

	"github.com/rs/zerolog/log"
)

// MarketPayUSSDHandler handles MarketPay USSD HTTP requests
type MarketPayUSSDHandler struct {
	flowService *ussd.MarketPayFlowService
}

// NewMarketPayUSSDHandler creates a new MarketPay USSD HTTP handler
func NewMarketPayUSSDHandler(flowService *ussd.MarketPayFlowService) *MarketPayUSSDHandler {
	return &MarketPayUSSDHandler{flowService: flowService}
}

// AdvanceRequest is the HTTP request payload for advancing the flow
type AdvanceRequest struct {
	SessionID   string            `json:"session_id"`
	CurrentPage string            `json:"current_page"`
	Values      map[string]string `json:"values"`
}

// AdvanceResponse is the HTTP response payload from the flow
type AdvanceResponse struct {
	Action   string            `json:"action"`
	NextPage string            `json:"next_page"`
	Message  string            `json:"message"`
	Data     map[string]string `json:"data"`
	Error    string            `json:"error,omitempty"`
}

// Advance handles the HTTP endpoint for advancing the USSD flow
func (h *MarketPayUSSDHandler) Advance(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 30*1000) // 30 seconds
	defer cancel()

	var req AdvanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error().Err(err).Msg("failed to decode MarketPay USSD request")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(AdvanceResponse{
			Error: "Invalid request payload",
		})
		return
	}

	log.Debug().
		Str("session_id", req.SessionID).
		Str("current_page", req.CurrentPage).
		Msg("MarketPay USSD HTTP request received")

	flowInput := ussd.AdvanceFlowInput{
		SessionID:   req.SessionID,
		CurrentPage: ussd.FlowPage(req.CurrentPage),
		Values:      req.Values,
	}

	result, err := h.flowService.Advance(ctx, flowInput)
	if err != nil {
		log.Error().Err(err).
			Str("session_id", req.SessionID).
			Msg("MarketPay USSD flow failed")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(AdvanceResponse{
			Error: "Failed to process request",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(AdvanceResponse{
		Action:   string(result.Action),
		NextPage: string(result.NextPage),
		Message:  result.Message,
		Data:     result.Data,
	})

	log.Debug().
		Str("session_id", req.SessionID).
		Str("action", string(result.Action)).
		Str("next_page", string(result.NextPage)).
		Msg("MarketPay USSD HTTP response sent")
}

// HealthCheck handles the health check endpoint
func (h *MarketPayUSSDHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"service": "marketpay-ussd",
	})
}
