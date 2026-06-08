package http

import (
	"context"
	"net/http"
	"time"

	"marketpay/internal/ussd"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// MarketPayUSSDHandler handles MarketPay USSD HTTP requests using Gin
type MarketPayUSSDHandler struct {
	flowService *ussd.MarketPayFlowService
}

// NewMarketPayUSSDHandler creates a new MarketPay USSD HTTP handler
func NewMarketPayUSSDHandler(flowService *ussd.MarketPayFlowService) *MarketPayUSSDHandler {
	return &MarketPayUSSDHandler{flowService: flowService}
}

// AdvanceRequest is the HTTP request payload for advancing the flow
type AdvanceRequest struct {
	SessionID   string            `json:"session_id" binding:"required"`
	CurrentPage string            `json:"current_page" binding:"required"`
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

// HealthResponse is the health check response
type HealthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
	Time    string `json:"time"`
}

// Advance handles the HTTP endpoint for advancing the USSD flow
func (h *MarketPayUSSDHandler) Advance(c *gin.Context) {
	var req AdvanceRequest

	// Bind and validate request
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error().Err(err).Msg("failed to bind MarketPay USSD request")
		c.JSON(http.StatusBadRequest, AdvanceResponse{
			Error: "Invalid request: " + err.Error(),
		})
		return
	}

	log.Debug().
		Str("session_id", req.SessionID).
		Str("current_page", req.CurrentPage).
		Msg("MarketPay USSD HTTP request received")

	// Create context with timeout
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

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
		c.JSON(http.StatusInternalServerError, AdvanceResponse{
			Error: "Failed to process request",
		})
		return
	}

	response := AdvanceResponse{
		Action:   string(result.Action),
		NextPage: string(result.NextPage),
		Message:  result.Message,
		Data:     result.Data,
	}

	log.Debug().
		Str("session_id", req.SessionID).
		Str("action", string(result.Action)).
		Str("next_page", string(result.NextPage)).
		Msg("MarketPay USSD HTTP response sent")

	c.JSON(http.StatusOK, response)
}

// HealthCheck handles the health check endpoint
func (h *MarketPayUSSDHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, HealthResponse{
		Status:  "ok",
		Service: "marketpay-ussd",
		Time:    time.Now().UTC().Format(time.RFC3339),
	})
}

// SetupRoutes sets up all USSD flow routes
func (h *MarketPayUSSDHandler) SetupRoutes(router *gin.Engine) {
	// Health check endpoints
	router.GET("/health", h.HealthCheck)
	router.GET("/healthz", h.HealthCheck)

	// USSD API v1
	v1 := router.Group("/api/v1")
	{
		ussd := v1.Group("/ussd")
		{
			ussd.POST("/advance", h.Advance)
		}
	}

	// Legacy USSD endpoint
	router.POST("/api/ussd/advance", h.Advance)

	// API info endpoint
	router.GET("/api/info", h.APIInfo)
}

// APIInfo returns API information
func (h *MarketPayUSSDHandler) APIInfo(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"service":    "MarketPay USSD Service",
		"version":    "1.0.0",
		"endpoints":  []string{"/api/v1/ussd/advance", "/api/ussd/advance", "/health"},
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
		"git_commit": "main",
	})
}
