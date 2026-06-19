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
	svc         *exchange.Service
	keyLoaded   bool
}

func NewHandler(svc *exchange.Service, keyLoaded bool) *Handler {
	return &Handler{svc: svc, keyLoaded: keyLoaded}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/monime/exchange", h.HandleExchange)
	rg.GET("/monime/exchange/health", h.Health)
}

func (h *Handler) Health(c *gin.Context) {
	c.JSON(200, gin.H{
		"status":           "ok",
		"exchange_enabled": true,
		"key_loaded":       h.keyLoaded,
	})
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
