package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	loanmodel "github.com/marketpay/backend/internal/loan/domain/model"
	repayapp "github.com/marketpay/backend/internal/repayment/application"
	shared "github.com/marketpay/backend/internal/shared/domain/model"
	"github.com/marketpay/backend/pkg/middleware"
	"github.com/marketpay/backend/pkg/response"
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
		response.Error(c, http.StatusBadRequest, "Please check the information you entered and try again.")
		return
	}

	loanID, err := uuid.Parse(req.LoanID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Please check the information you entered and try again.")
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
		response.ErrorFromAppError(c, err)
		return
	}
	response.Success(c, loan, "Repayment processed successfully.")
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
		response.Error(c, http.StatusBadRequest, "Please check the information you entered and try again.")
		return
	}

	if err := h.svc.MarkDefaulted(c.Request.Context(), loanID); err != nil {
		response.ErrorFromAppError(c, err)
		return
	}
	response.Success(c, nil, "Loan marked as defaulted successfully.")
}
