package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	ussdapp "github.com/marketpay/backend/internal/ussd/application"
	ussdmodel "github.com/marketpay/backend/internal/ussd/domain/model"
)

// Handler handles inbound USSD gateway callbacks.
type Handler struct {
	svc *ussdapp.USSDService
}

func NewHandler(svc *ussdapp.USSDService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/ussd", h.Handle)
}

// Handle godoc
// @Summary Handle USSD request from gateway
// @Tags ussd
// @Accept x-www-form-urlencoded
// @Produce plain
// @Param sessionId formData string true "Session ID"
// @Param serviceCode formData string true "Service Code"
// @Param phoneNumber formData string true "Phone Number"
// @Param text formData string false "Accumulated input"
// @Success 200 {string} string "CON ... or END ..."
// @Router /ussd [post]
func (h *Handler) Handle(c *gin.Context) {
	req := ussdmodel.USSDRequest{
		SessionID:   c.PostForm("sessionId"),
		ServiceCode: c.PostForm("serviceCode"),
		PhoneNumber: c.PostForm("phoneNumber"),
		Text:        c.PostForm("text"),
	}

	if req.SessionID == "" || req.PhoneNumber == "" {
		c.String(http.StatusBadRequest, "END Invalid request.")
		return
	}

	response, _ := h.svc.Process(c.Request.Context(), req)
	c.String(http.StatusOK, response.Message)
}
