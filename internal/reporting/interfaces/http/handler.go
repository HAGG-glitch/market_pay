package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	shared "github.com/marketpay/backend/internal/shared/domain/model"
	"github.com/marketpay/backend/pkg/middleware"
	"gorm.io/gorm"
)

// Handler handles reporting requests.
type Handler struct {
	db *gorm.DB
}

func NewHandler(db *gorm.DB) *Handler {
	return &Handler{db: db}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, auth gin.HandlerFunc) {
	r := rg.Group("/reports")
	r.Use(auth)
	{
		r.GET("/portfolio", middleware.RequireRoles(shared.RoleAdmin, shared.RoleSuperAdmin, shared.RoleMFIPartner), h.Portfolio)
		r.GET("/repayment-rate", middleware.RequireRoles(shared.RoleAdmin, shared.RoleSuperAdmin, shared.RoleMFIPartner), h.RepaymentRate)
		r.GET("/default-rate", middleware.RequireRoles(shared.RoleAdmin, shared.RoleSuperAdmin), h.DefaultRate)
		r.GET("/disbursement-volume", middleware.RequireRoles(shared.RoleAdmin, shared.RoleSuperAdmin), h.DisbursementVolume)
		r.GET("/vendor-distribution", middleware.RequireRoles(shared.RoleAdmin, shared.RoleSuperAdmin), h.VendorDistribution)
		r.GET("/partner-summary", middleware.RequireRoles(shared.RoleMFIPartner, shared.RoleAdmin, shared.RoleSuperAdmin), h.PartnerSummary)
		r.GET("/officer-queue", middleware.RequireRoles(shared.RoleLoanOfficer, shared.RoleAdmin), h.OfficerQueue)
	}
}

// Portfolio godoc
// @Summary Portfolio outstanding report
// @Tags reports
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /reports/portfolio [get]
func (h *Handler) Portfolio(c *gin.Context) {
	var result struct {
		TotalLoans       int64   `json:"total_loans"`
		ActiveLoans      int64   `json:"active_loans"`
		TotalDisbursed   float64 `json:"total_disbursed"`
		TotalOutstanding float64 `json:"total_outstanding"`
		TotalInterest    float64 `json:"total_interest_earned"`
	}

	h.db.Raw(`
		SELECT
			COUNT(*) as total_loans,
			COUNT(CASE WHEN state = 'ACTIVE' THEN 1 END) as active_loans,
			COALESCE(SUM(principal_amount), 0) as total_disbursed,
			COALESCE(SUM(outstanding_amount), 0) as total_outstanding,
			COALESCE(SUM(total_amount - principal_amount), 0) as total_interest_earned
		FROM loans
		WHERE deleted_at IS NULL AND state IN ('ACTIVE','CLOSED','DEFAULTED')
	`).Scan(&result)

	c.JSON(http.StatusOK, result)
}

// RepaymentRate godoc
// @Summary Repayment rate report
// @Tags reports
// @Security BearerAuth
// @Router /reports/repayment-rate [get]
func (h *Handler) RepaymentRate(c *gin.Context) {
	var result struct {
		TotalScheduled int64   `json:"total_scheduled"`
		TotalPaid      int64   `json:"total_paid"`
		RepaymentRate  float64 `json:"repayment_rate_pct"`
	}

	h.db.Raw(`
		SELECT
			COUNT(*) as total_scheduled,
			COUNT(CASE WHEN status = 'PAID' THEN 1 END) as total_paid,
			ROUND(COUNT(CASE WHEN status = 'PAID' THEN 1 END)::numeric / NULLIF(COUNT(*),0) * 100, 2) as repayment_rate_pct
		FROM repayment_schedules
		WHERE deleted_at IS NULL
	`).Scan(&result)

	c.JSON(http.StatusOK, result)
}

// DefaultRate godoc
// @Summary Default rate report
// @Tags reports
// @Security BearerAuth
// @Router /reports/default-rate [get]
func (h *Handler) DefaultRate(c *gin.Context) {
	var result struct {
		TotalLoans   int64   `json:"total_loans"`
		DefaultedLoans int64 `json:"defaulted_loans"`
		DefaultRate  float64 `json:"default_rate_pct"`
	}

	h.db.Raw(`
		SELECT
			COUNT(*) as total_loans,
			COUNT(CASE WHEN state = 'DEFAULTED' THEN 1 END) as defaulted_loans,
			ROUND(COUNT(CASE WHEN state = 'DEFAULTED' THEN 1 END)::numeric / NULLIF(COUNT(*),0) * 100, 2) as default_rate_pct
		FROM loans
		WHERE deleted_at IS NULL AND state NOT IN ('DRAFT','PENDING_REVIEW','UNDER_REVIEW','AUTO_APPROVED','REJECTED')
	`).Scan(&result)

	c.JSON(http.StatusOK, result)
}

// DisbursementVolume godoc
// @Summary Disbursement volume by month
// @Tags reports
// @Security BearerAuth
// @Router /reports/disbursement-volume [get]
func (h *Handler) DisbursementVolume(c *gin.Context) {
	var result []struct {
		Month  string  `json:"month"`
		Count  int64   `json:"count"`
		Volume float64 `json:"volume"`
	}

	h.db.Raw(`
		SELECT
			TO_CHAR(disbursed_at, 'YYYY-MM') as month,
			COUNT(*) as count,
			SUM(principal_amount) as volume
		FROM loans
		WHERE deleted_at IS NULL AND disbursed_at IS NOT NULL
		GROUP BY month
		ORDER BY month DESC
		LIMIT 12
	`).Scan(&result)

	c.JSON(http.StatusOK, result)
}

// VendorDistribution godoc
// @Summary Vendor distribution by market
// @Tags reports
// @Security BearerAuth
// @Router /reports/vendor-distribution [get]
func (h *Handler) VendorDistribution(c *gin.Context) {
	var result []struct {
		MarketName  string `json:"market_name"`
		VendorCount int64  `json:"vendor_count"`
	}

	h.db.Raw(`
		SELECT ma.name as market_name, COUNT(v.id) as vendor_count
		FROM vendors v
		JOIN market_associations ma ON ma.id = v.market_association_id
		WHERE v.deleted_at IS NULL
		GROUP BY ma.name
		ORDER BY vendor_count DESC
	`).Scan(&result)

	c.JSON(http.StatusOK, result)
}

// PartnerSummary godoc
// @Summary Partner loans and commission summary
// @Tags reports
// @Security BearerAuth
// @Router /reports/partner-summary [get]
func (h *Handler) PartnerSummary(c *gin.Context) {
	var result []struct {
		PartnerID       string  `json:"partner_id"`
		LoansIssued     int64   `json:"loans_issued"`
		TotalDisbursed  float64 `json:"total_disbursed"`
		TotalRepaid     float64 `json:"total_repaid"`
		CommissionOwed  float64 `json:"commission_owed"`
	}

	h.db.Raw(`
		SELECT
			partner_id::text,
			COUNT(*) as loans_issued,
			SUM(principal_amount) as total_disbursed,
			SUM(principal_amount - outstanding_amount) as total_repaid,
			SUM(principal_amount * commission_rate) as commission_owed
		FROM loans
		WHERE deleted_at IS NULL AND partner_id IS NOT NULL
		GROUP BY partner_id
	`).Scan(&result)

	c.JSON(http.StatusOK, result)
}

// OfficerQueue godoc
// @Summary Loan officer pending review queue
// @Tags reports
// @Security BearerAuth
// @Router /reports/officer-queue [get]
func (h *Handler) OfficerQueue(c *gin.Context) {
	var result struct {
		PendingReview int64 `json:"pending_review"`
		UnderReview   int64 `json:"under_review"`
	}

	h.db.Raw(`
		SELECT
			COUNT(CASE WHEN state = 'PENDING_REVIEW' THEN 1 END) as pending_review,
			COUNT(CASE WHEN state = 'UNDER_REVIEW' THEN 1 END) as under_review
		FROM loans WHERE deleted_at IS NULL
	`).Scan(&result)

	c.JSON(http.StatusOK, result)
}

// ensure middleware import used
var _ = middleware.GetUserID
