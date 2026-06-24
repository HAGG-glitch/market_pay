package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	loanmodel "github.com/marketpay/backend/internal/loan/domain/model"
	repayapp "github.com/marketpay/backend/internal/repayment/application"
	shared "github.com/marketpay/backend/internal/shared/domain/model"
	apperrors "github.com/marketpay/backend/pkg/errors"
	"github.com/marketpay/backend/pkg/middleware"
)

// Swagger schema references.
var _ loanmodel.Loan

// Handler handles repayment requests.
type Handler struct {
	svc *repayapp.RepaymentService
}

// NewHandler constructs a repayment Handler.
func NewHandler(svc *repayapp.RepaymentService) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes mounts repayment routes.
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, auth gin.HandlerFunc) {
	r := rg.Group("/repayments")
	r.Use(auth, middleware.RequireActiveVendor())
	{
		r.POST("", middleware.RequireRoles(shared.RoleVendor, shared.RoleAdmin), h.Repay)
		r.PUT("/loans/:id/default", middleware.RequireRoles(shared.RoleAdmin, shared.RoleSuperAdmin), h.MarkDefault)
	}
}

type repayRequest struct {
	LoanID    string  `json:"loan_id" binding:"required"`
	Amount    float64 `json:"amount" binding:"required,gt=0"`
	MonimeRef string  `json:"monime_reference"`
}

// Repay godoc
// @Summary Make a loan repayment
// @Tags repayments
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body repayRequest true "Repayment data"
// @Success 200 {object} loanmodel.Loan
// @Router /repayments [post]
func (h *Handler) Repay(c *gin.Context) {
	var req repayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	loanID, err := uuid.Parse(req.LoanID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid loan_id"})
		return
	}

	vendorIDStr := middleware.GetUserID(c)
	vendorID, _ := uuid.Parse(vendorIDStr)

	loan, err := h.svc.Repay(c.Request.Context(), repayapp.RepayInput{
		LoanID:    loanID,
		VendorID:  vendorID,
		Amount:    req.Amount,
		MonimeRef: req.MonimeRef,
	})
	if err != nil {
		c.JSON(apperrors.HTTPStatus(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, loan)
}

// MarkDefault godoc
// @Summary Mark a loan as defaulted
// @Tags repayments
// @Security BearerAuth
// @Param id path string true "Loan ID"
// @Success 200 {object} map[string]string
// @Router /repayments/loans/{id}/default [put]
func (h *Handler) MarkDefault(c *gin.Context) {
	loanID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid loan ID"})
		return
	}

	if err := h.svc.MarkDefaulted(c.Request.Context(), loanID); err != nil {
		c.JSON(apperrors.HTTPStatus(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "loan marked as defaulted"})
}
