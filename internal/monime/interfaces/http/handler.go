package http

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/marketpay/backend/internal/monime/exchange"
	"github.com/marketpay/backend/pkg/monimeexchange"
)

// Handler serves Monime encrypted exchange endpoints.
type Handler struct {
	svc *exchange.Service
}

func NewHandler(svc *exchange.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/monime/exchange", h.HandleExchange)
}

// HandleExchange decrypts Monime exchange requests and returns encrypted text/plain.
func (h *Handler) HandleExchange(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.String(http.StatusBadRequest, "")
		return
	}

	var req monimeexchange.EncryptedRequest
	if err := json.Unmarshal(body, &req); err != nil || req.EncryptedAesKey == "" {
		c.String(http.StatusBadRequest, "")
		return
	}

	encrypted, err := h.svc.Handle(c.Request.Context(), req)
	if err != nil {
		c.String(http.StatusInternalServerError, "")
		return
	}

	c.Header("Content-Type", "text/plain")
	c.String(http.StatusOK, encrypted)
}
