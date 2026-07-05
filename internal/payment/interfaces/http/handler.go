package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	paymentapp "github.com/marketpay/backend/internal/payment/application"
	shared "github.com/marketpay/backend/internal/shared/domain/model"
	"github.com/marketpay/backend/pkg/democtx"
	"github.com/marketpay/backend/pkg/middleware"
	"github.com/marketpay/backend/pkg/pagination"
	"github.com/marketpay/backend/pkg/response"
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
	p.Use(auth, middleware.RequireActiveVendor())
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
		response.Error(c, http.StatusInternalServerError, "Something went wrong while fetching payments. Please try again.")
		return
	}
	response.Paginated(c, payments, total, params.Page, params.Limit, "Payments retrieved successfully.")
}

func (h *Handler) Initiate(c *gin.Context) {
	var req initiatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Please check the information you entered and try again.")
		return
	}

	customerID, err := uuid.Parse(req.CustomerID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Please check the information you entered and try again.")
		return
	}
	vendorID, err := uuid.Parse(req.VendorID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "We couldn't find this vendor. The link may be incorrect or the vendor may have been removed.")
		return
	}

	payment, err := h.svc.Initiate(c.Request.Context(), paymentapp.InitiateInput{
		CustomerID: customerID,
		VendorID:   vendorID,
		Amount:     req.Amount,
		IsDemo:     democtx.FromGin(c),
	})
	if err != nil {
		response.ErrorFromAppError(c, err)
		return
	}
	response.Created(c, payment, "Payment initiated successfully.")
}

func (h *Handler) Complete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Please check the information you entered and try again.")
		return
	}

	var req completePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Please check the information you entered and try again.")
		return
	}

	payment, err := h.svc.Complete(c.Request.Context(), id, req.MonimeReference)
	if err != nil {
		response.ErrorFromAppError(c, err)
		return
	}
	response.Success(c, payment, "Payment completed successfully.")
}

func (h *Handler) GetVendorPayments(c *gin.Context) {
	vendorID, err := uuid.Parse(c.Param("vendor_id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "We couldn't find this vendor. The link may be incorrect or the vendor may have been removed.")
		return
	}

	params := pagination.FromQuery(c)
	payments, total, err := h.svc.GetVendorPayments(c.Request.Context(), vendorID, params.Offset(), params.Limit)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Something went wrong while fetching vendor payments. Please try again.")
		return
	}
	response.Paginated(c, payments, total, params.Page, params.Limit, "Vendor payments retrieved successfully.")
}

// GetUserID used in handler
var _ = middleware.GetUserID
