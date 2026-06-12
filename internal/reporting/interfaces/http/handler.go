package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	shared "github.com/marketpay/backend/internal/shared/domain/model"
	"github.com/marketpay/backend/pkg/democtx"
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
		r.GET("/disbursement-volume", middleware.RequireRoles(shared.RoleAdmin, shared.RoleSuperAdmin, shared.RoleMFIPartner), h.DisbursementVolume)
		r.GET("/vendor-distribution", middleware.RequireRoles(shared.RoleAdmin, shared.RoleSuperAdmin), h.VendorDistribution)
		r.GET("/partner-summary", middleware.RequireRoles(shared.RoleMFIPartner, shared.RoleAdmin, shared.RoleSuperAdmin), h.PartnerSummary)
		r.GET("/officer-queue", middleware.RequireRoles(shared.RoleLoanOfficer, shared.RoleAdmin, shared.RoleSuperAdmin), h.OfficerQueue)
		r.GET("/dashboard-summary", middleware.RequireRoles(shared.RoleAdmin, shared.RoleSuperAdmin, shared.RoleMFIPartner, shared.RoleLoanOfficer), h.DashboardSummary)
	}
}

func (h *Handler) demo(c *gin.Context) bool {
	return democtx.FromGin(c)
}

func (h *Handler) Portfolio(c *gin.Context) {
	isDemo := h.demo(c)
	var result struct {
		TotalLoans       int64   `json:"total_loans"`
		ActiveLoans      int64   `json:"active_loans"`
		OverdueLoans     int64   `json:"overdue_loans"`
		TotalDisbursed   float64 `json:"total_disbursed"`
		TotalOutstanding float64 `json:"total_outstanding"`
		TotalInterest    float64 `json:"total_interest_earned"`
		TotalVendors     int64   `json:"total_vendors"`
		TotalCustomers   int64   `json:"total_customers"`
	}

	h.db.Raw(`
		SELECT
			COUNT(*) as total_loans,
			COUNT(CASE WHEN state = 'ACTIVE' THEN 1 END) as active_loans,
			COUNT(CASE WHEN state = 'DEFAULTED' THEN 1 END) as overdue_loans,
			COALESCE(SUM(principal_amount), 0) as total_disbursed,
			COALESCE(SUM(outstanding_amount), 0) as total_outstanding,
			COALESCE(SUM(total_amount - principal_amount), 0) as total_interest_earned
		FROM loans
		WHERE deleted_at IS NULL AND is_demo = ? AND state IN ('ACTIVE','CLOSED','DEFAULTED','DISBURSED')
	`, isDemo).Scan(&result)

	h.db.Raw(`SELECT COUNT(*) FROM vendors WHERE deleted_at IS NULL AND is_demo = ?`, isDemo).Scan(&result.TotalVendors)
	h.db.Raw(`SELECT COUNT(*) FROM customers WHERE deleted_at IS NULL AND is_demo = ?`, isDemo).Scan(&result.TotalCustomers)

	c.JSON(http.StatusOK, result)
}

func (h *Handler) RepaymentRate(c *gin.Context) {
	isDemo := h.demo(c)
	var result struct {
		TotalScheduled int64   `json:"total_scheduled"`
		TotalPaid      int64   `json:"total_paid"`
		RepaymentRate  float64 `json:"repayment_rate_pct"`
	}

	h.db.Raw(`
		SELECT
			COUNT(rs.*) as total_scheduled,
			COUNT(CASE WHEN rs.status = 'PAID' THEN 1 END) as total_paid,
			ROUND(COUNT(CASE WHEN rs.status = 'PAID' THEN 1 END)::numeric / NULLIF(COUNT(rs.*),0) * 100, 2) as repayment_rate_pct
		FROM repayment_schedules rs
		JOIN loans l ON l.id = rs.loan_id
		WHERE rs.deleted_at IS NULL AND l.is_demo = ?
	`, isDemo).Scan(&result)

	c.JSON(http.StatusOK, result)
}

func (h *Handler) DefaultRate(c *gin.Context) {
	isDemo := h.demo(c)
	var result struct {
		TotalLoans     int64   `json:"total_loans"`
		DefaultedLoans int64   `json:"defaulted_loans"`
		DefaultRate    float64 `json:"default_rate_pct"`
	}

	h.db.Raw(`
		SELECT
			COUNT(*) as total_loans,
			COUNT(CASE WHEN state = 'DEFAULTED' THEN 1 END) as defaulted_loans,
			ROUND(COUNT(CASE WHEN state = 'DEFAULTED' THEN 1 END)::numeric / NULLIF(COUNT(*),0) * 100, 2) as default_rate_pct
		FROM loans
		WHERE deleted_at IS NULL AND is_demo = ? AND state NOT IN ('DRAFT','PENDING_REVIEW','UNDER_REVIEW','AUTO_APPROVED','REJECTED')
	`, isDemo).Scan(&result)

	c.JSON(http.StatusOK, result)
}

func (h *Handler) DisbursementVolume(c *gin.Context) {
	isDemo := h.demo(c)
	var result []struct {
		Month  string  `json:"month"`
		Count  int64   `json:"count"`
		Volume float64 `json:"volume"`
	}

	h.db.Raw(`
		SELECT
			TO_CHAR(disbursed_at, 'YYYY-MM') as month,
			COUNT(*) as count,
			COALESCE(SUM(principal_amount), 0) as volume
		FROM loans
		WHERE deleted_at IS NULL AND is_demo = ? AND disbursed_at IS NOT NULL
		GROUP BY month
		ORDER BY month DESC
		LIMIT 12
	`, isDemo).Scan(&result)

	c.JSON(http.StatusOK, result)
}

func (h *Handler) VendorDistribution(c *gin.Context) {
	isDemo := h.demo(c)
	var result []struct {
		MarketName  string `json:"market_name"`
		VendorCount int64  `json:"vendor_count"`
	}

	h.db.Raw(`
		SELECT ma.name as market_name, COUNT(v.id) as vendor_count
		FROM vendors v
		JOIN market_associations ma ON ma.id = v.market_association_id
		WHERE v.deleted_at IS NULL AND v.is_demo = ?
		GROUP BY ma.name
		ORDER BY vendor_count DESC
	`, isDemo).Scan(&result)

	c.JSON(http.StatusOK, result)
}

func (h *Handler) PartnerSummary(c *gin.Context) {
	isDemo := h.demo(c)
	var result []struct {
		PartnerID      string  `json:"partner_id"`
		LoansIssued    int64   `json:"loans_issued"`
		TotalDisbursed float64 `json:"total_disbursed"`
		TotalRepaid    float64 `json:"total_repaid"`
		CommissionOwed float64 `json:"commission_owed"`
	}

	h.db.Raw(`
		SELECT
			partner_id::text,
			COUNT(*) as loans_issued,
			COALESCE(SUM(principal_amount), 0) as total_disbursed,
			COALESCE(SUM(principal_amount - outstanding_amount), 0) as total_repaid,
			COALESCE(SUM(principal_amount * commission_rate), 0) as commission_owed
		FROM loans
		WHERE deleted_at IS NULL AND is_demo = ? AND partner_id IS NOT NULL
		GROUP BY partner_id
	`, isDemo).Scan(&result)

	c.JSON(http.StatusOK, result)
}

func (h *Handler) OfficerQueue(c *gin.Context) {
	isDemo := h.demo(c)
	var result struct {
		PendingReview int64 `json:"pending_review"`
		UnderReview   int64 `json:"under_review"`
	}

	h.db.Raw(`
		SELECT
			COUNT(CASE WHEN state = 'PENDING_REVIEW' THEN 1 END) as pending_review,
			COUNT(CASE WHEN state = 'UNDER_REVIEW' THEN 1 END) as under_review
		FROM loans WHERE deleted_at IS NULL AND is_demo = ?
	`, isDemo).Scan(&result)

	c.JSON(http.StatusOK, result)
}

func (h *Handler) DashboardSummary(c *gin.Context) {
	isDemo := h.demo(c)
	var summary struct {
		TotalVendors   int64   `json:"total_vendors"`
		ActiveLoans    int64   `json:"active_loans"`
		OverdueLoans   int64   `json:"overdue_loans"`
		PortfolioValue float64 `json:"portfolio_value"`
		RepaymentRate  float64 `json:"repayment_rate"`
		DefaultRate    float64 `json:"default_rate"`
	}

	h.db.Raw(`SELECT COUNT(*) FROM vendors WHERE deleted_at IS NULL AND is_demo = ?`, isDemo).Scan(&summary.TotalVendors)
	h.db.Raw(`SELECT COUNT(*) FROM loans WHERE deleted_at IS NULL AND is_demo = ? AND state = 'ACTIVE'`, isDemo).Scan(&summary.ActiveLoans)
	h.db.Raw(`SELECT COUNT(*) FROM loans WHERE deleted_at IS NULL AND is_demo = ? AND state = 'DEFAULTED'`, isDemo).Scan(&summary.OverdueLoans)
	h.db.Raw(`SELECT COALESCE(SUM(outstanding_amount),0) FROM loans WHERE deleted_at IS NULL AND is_demo = ? AND state IN ('ACTIVE','DISBURSED')`, isDemo).Scan(&summary.PortfolioValue)

	var repay struct{ Rate float64 }
	h.db.Raw(`
		SELECT ROUND(COUNT(CASE WHEN rs.status = 'PAID' THEN 1 END)::numeric / NULLIF(COUNT(rs.*),0) * 100, 2) as rate
		FROM repayment_schedules rs JOIN loans l ON l.id = rs.loan_id
		WHERE rs.deleted_at IS NULL AND l.is_demo = ?
	`, isDemo).Scan(&repay)
	summary.RepaymentRate = repay.Rate

	var def struct{ Rate float64 }
	h.db.Raw(`
		SELECT ROUND(COUNT(CASE WHEN state = 'DEFAULTED' THEN 1 END)::numeric / NULLIF(COUNT(*),0) * 100, 2) as rate
		FROM loans WHERE deleted_at IS NULL AND is_demo = ? AND state NOT IN ('DRAFT','PENDING_REVIEW','UNDER_REVIEW','REJECTED')
	`, isDemo).Scan(&def)
	summary.DefaultRate = def.Rate

	c.JSON(http.StatusOK, summary)
}
