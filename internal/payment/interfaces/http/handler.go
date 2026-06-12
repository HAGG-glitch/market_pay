package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	paymentapp "github.com/marketpay/backend/internal/payment/application"
	shared "github.com/marketpay/backend/internal/shared/domain/model"
	apperrors "github.com/marketpay/backend/pkg/errors"
	"github.com/marketpay/backend/pkg/democtx"
	"github.com/marketpay/backend/pkg/middleware"
	"github.com/marketpay/backend/pkg/pagination"
)

// Handler handles payment HTTP requests.
type Handler struct {
	svc *paymentapp.PaymentService
}

func NewHandler(svc *paymentapp.PaymentService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, auth gin.HandlerFunc) {
	p := rg.Group("/payments")
	p.Use(auth)
	{
		p.GET("", middleware.RequireRoles(shared.RoleAdmin, shared.RoleSuperAdmin), h.List)
		p.POST("", middleware.RequireRoles(shared.RoleCustomer, shared.RoleVendor), h.Initiate)
		p.PUT("/:id/complete", middleware.RequireRoles(shared.RoleAdmin, shared.RoleSuperAdmin), h.Complete)
		p.GET("/vendor/:vendor_id", h.GetVendorPayments)
	}
}

type initiatePaymentRequest struct {
	CustomerID string  `json:"customer_id" binding:"required"`
	VendorID   string  `json:"vendor_id" binding:"required"`
	Amount     float64 `json:"amount" binding:"required,gt=0"`
}

type completePaymentRequest struct {
	MonimeReference string `json:"monime_reference" binding:"required"`
}

func (h *Handler) List(c *gin.Context) {
	params := pagination.FromQuery(c)
	isDemo := democtx.FromGin(c)

	payments, total, err := h.svc.List(c.Request.Context(), isDemo, params.Offset(), params.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, pagination.NewResponse(payments, total, params))
}

func (h *Handler) Initiate(c *gin.Context) {
	var req initiatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	customerID, err := uuid.Parse(req.CustomerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid customer_id"})
		return
	}
	vendorID, err := uuid.Parse(req.VendorID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid vendor_id"})
		return
	}

	payment, err := h.svc.Initiate(c.Request.Context(), paymentapp.InitiateInput{
		CustomerID: customerID,
		VendorID:   vendorID,
		Amount:     req.Amount,
		IsDemo:     democtx.FromGin(c),
	})
	if err != nil {
		c.JSON(apperrors.HTTPStatus(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, payment)
}

func (h *Handler) Complete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payment ID"})
		return
	}

	var req completePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	payment, err := h.svc.Complete(c.Request.Context(), id, req.MonimeReference)
	if err != nil {
		c.JSON(apperrors.HTTPStatus(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, payment)
}

func (h *Handler) GetVendorPayments(c *gin.Context) {
	vendorID, err := uuid.Parse(c.Param("vendor_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid vendor ID"})
		return
	}

	params := pagination.FromQuery(c)
	payments, total, err := h.svc.GetVendorPayments(c.Request.Context(), vendorID, params.Offset(), params.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, pagination.NewResponse(payments, total, params))
}

// GetUserID used in handler
var _ = middleware.GetUserID
