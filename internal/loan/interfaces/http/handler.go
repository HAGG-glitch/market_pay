package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	loanapp "github.com/marketpay/backend/internal/loan/application"
	loanmodel "github.com/marketpay/backend/internal/loan/domain/model"
	shared "github.com/marketpay/backend/internal/shared/domain/model"
	apperrors "github.com/marketpay/backend/pkg/errors"
	"github.com/marketpay/backend/pkg/democtx"
	"github.com/marketpay/backend/pkg/middleware"
	"github.com/marketpay/backend/pkg/pagination"
)

// Handler handles loan HTTP requests.
type Handler struct {
	loanSvc *loanapp.LoanService
}

// NewHandler constructs a loan Handler.
func NewHandler(loanSvc *loanapp.LoanService) *Handler {
	return &Handler{loanSvc: loanSvc}
}

// RegisterRoutes mounts loan routes.
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, auth gin.HandlerFunc) {
	loans := rg.Group("/loans")
	loans.Use(auth)
	{
		loans.POST("", middleware.RequireRoles(shared.RoleVendor), h.Apply)
		loans.GET("", middleware.RequireRoles(shared.RoleAdmin, shared.RoleSuperAdmin, shared.RoleLoanOfficer), h.ListByState)
		loans.GET("/:id", h.GetByID)
		loans.GET("/:id/schedule", h.GetSchedule)
		loans.PUT("/:id/approve", middleware.RequireRoles(shared.RoleLoanOfficer, shared.RoleAdmin, shared.RoleSuperAdmin), h.Approve)
		loans.PUT("/:id/reject", middleware.RequireRoles(shared.RoleLoanOfficer, shared.RoleAdmin, shared.RoleSuperAdmin), h.Reject)
		loans.PUT("/:id/disburse", middleware.RequireRoles(shared.RoleAdmin, shared.RoleSuperAdmin), h.Disburse)
		loans.GET("/vendor/:vendor_id", h.GetVendorLoans)
	}
}

type applyLoanRequest struct {
	LoanType  loanmodel.LoanType           `json:"loan_type" binding:"required"`
	Amount    float64                       `json:"amount" binding:"required,gt=0"`
	TermWeeks int                           `json:"term_weeks" binding:"required,gt=0"`
	Frequency loanmodel.RepaymentFrequency `json:"frequency" binding:"required"`
	GroupID   *string                       `json:"group_id"`
	FundedBy  loanmodel.FundingSource      `json:"funded_by" binding:"required"`
	PartnerID *string                       `json:"partner_id"`
}

type reviewLoanRequest struct {
	Note string `json:"note"`
}

type rejectLoanRequest struct {
	Reason string `json:"reason" binding:"required"`
}

type disburseLoanRequest struct {
	MonimeReference string `json:"monime_reference"`
}

// Apply godoc
// @Summary Apply for a loan
// @Tags loans
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body applyLoanRequest true "Loan application"
// @Success 201 {object} loanmodel.Loan
// @Router /loans [post]
func (h *Handler) Apply(c *gin.Context) {
	var req applyLoanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	vendorIDStr := middleware.GetUserID(c)
	vendorID, _ := uuid.Parse(vendorIDStr)

	input := loanapp.ApplyInput{
		VendorID:  vendorID,
		LoanType:  req.LoanType,
		Amount:    req.Amount,
		TermWeeks: req.TermWeeks,
		Frequency: req.Frequency,
		FundedBy:  req.FundedBy,
		IsDemo:    democtx.FromGin(c),
	}

	if req.GroupID != nil {
		gid, err := uuid.Parse(*req.GroupID)
		if err == nil {
			input.GroupID = &gid
		}
	}
	if req.PartnerID != nil {
		pid, err := uuid.Parse(*req.PartnerID)
		if err == nil {
			input.PartnerID = &pid
		}
	}

	loan, err := h.loanSvc.Apply(c.Request.Context(), input)
	if err != nil {
		c.JSON(apperrors.HTTPStatus(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, loan)
}

// GetByID godoc
// @Summary Get loan by ID
// @Tags loans
// @Security BearerAuth
// @Param id path string true "Loan ID"
// @Success 200 {object} loanmodel.Loan
// @Router /loans/{id} [get]
func (h *Handler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid loan ID"})
		return
	}

	loan, err := h.loanSvc.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(apperrors.HTTPStatus(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, loan)
}

// GetSchedule godoc
// @Summary Get repayment schedule for a loan
// @Tags loans
// @Security BearerAuth
// @Param id path string true "Loan ID"
// @Success 200 {array} loanmodel.RepaymentSchedule
// @Router /loans/{id}/schedule [get]
func (h *Handler) GetSchedule(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid loan ID"})
		return
	}

	schedules, err := h.loanSvc.GetSchedule(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, schedules)
}

// Approve godoc
// @Summary Approve a loan
// @Tags loans
// @Security BearerAuth
// @Param id path string true "Loan ID"
// @Param body body reviewLoanRequest false "Review note"
// @Success 200 {object} loanmodel.Loan
// @Router /loans/{id}/approve [put]
func (h *Handler) Approve(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid loan ID"})
		return
	}

	var req reviewLoanRequest
	_ = c.ShouldBindJSON(&req)

	officerIDStr := middleware.GetUserID(c)
	officerID, _ := uuid.Parse(officerIDStr)

	loan, err := h.loanSvc.Approve(c.Request.Context(), id, officerID, req.Note)
	if err != nil {
		c.JSON(apperrors.HTTPStatus(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, loan)
}

// Reject godoc
// @Summary Reject a loan
// @Tags loans
// @Security BearerAuth
// @Param id path string true "Loan ID"
// @Param body body rejectLoanRequest true "Rejection reason"
// @Success 200 {object} loanmodel.Loan
// @Router /loans/{id}/reject [put]
func (h *Handler) Reject(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid loan ID"})
		return
	}

	var req rejectLoanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	officerIDStr := middleware.GetUserID(c)
	officerID, _ := uuid.Parse(officerIDStr)

	loan, err := h.loanSvc.Reject(c.Request.Context(), id, officerID, req.Reason)
	if err != nil {
		c.JSON(apperrors.HTTPStatus(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, loan)
}

// Disburse godoc
// @Summary Disburse an approved loan
// @Tags loans
// @Security BearerAuth
// @Param id path string true "Loan ID"
// @Param body body disburseLoanRequest false "Optional manual Monime reference"
// @Success 200 {object} loanmodel.Loan
// @Router /loans/{id}/disburse [put]
func (h *Handler) Disburse(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid loan ID"})
		return
	}

	var req disburseLoanRequest
	_ = c.ShouldBindJSON(&req)

	var loan *loanmodel.Loan
	if req.MonimeReference != "" {
		loan, err = h.loanSvc.Disburse(c.Request.Context(), id, req.MonimeReference)
	} else {
		loan, err = h.loanSvc.DisburseWithPayout(c.Request.Context(), id)
	}
	if err != nil {
		c.JSON(apperrors.HTTPStatus(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, loan)
}

// ListByState godoc
// @Summary List loans by state
// @Tags loans
// @Security BearerAuth
// @Param state query string false "Loan state"
// @Param page query int false "Page"
// @Param limit query int false "Limit"
// @Success 200 {object} pagination.Response[loanmodel.Loan]
// @Router /loans [get]
func (h *Handler) ListByState(c *gin.Context) {
	stateStr := c.DefaultQuery("state", string(loanmodel.LoanStateUnderReview))
	state := loanmodel.LoanState(stateStr)
	params := pagination.FromQuery(c)

	loans, total, err := h.loanSvc.ListByState(c.Request.Context(), state, democtx.FromGin(c), params.Offset(), params.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, pagination.NewResponse(loans, total, params))
}

// GetVendorLoans godoc
// @Summary List loans for a specific vendor
// @Tags loans
// @Security BearerAuth
// @Param vendor_id path string true "Vendor ID"
// @Success 200 {object} pagination.Response[loanmodel.Loan]
// @Router /loans/vendor/{vendor_id} [get]
func (h *Handler) GetVendorLoans(c *gin.Context) {
	vendorID, err := uuid.Parse(c.Param("vendor_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid vendor ID"})
		return
	}

	params := pagination.FromQuery(c)
	loans, total, err := h.loanSvc.GetVendorLoans(c.Request.Context(), vendorID, params.Offset(), params.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, pagination.NewResponse(loans, total, params))
}
