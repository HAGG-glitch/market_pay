package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	authapp "github.com/marketpay/backend/internal/auth/application"
	vendorapp "github.com/marketpay/backend/internal/vendors/application"
	shared "github.com/marketpay/backend/internal/shared/domain/model"
	"github.com/marketpay/backend/pkg/democtx"
	apperrors "github.com/marketpay/backend/pkg/errors"
	"github.com/marketpay/backend/pkg/middleware"
	"github.com/google/uuid"
)

// Handler handles auth HTTP requests.
type Handler struct {
	authSvc   *authapp.AuthService
	vendorSvc *vendorapp.VendorService
}

// NewHandler constructs an auth Handler.
func NewHandler(authSvc *authapp.AuthService, vendorSvc *vendorapp.VendorService) *Handler {
	return &Handler{authSvc: authSvc, vendorSvc: vendorSvc}
}

// RegisterRoutes mounts auth routes onto a router group.
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	auth := rg.Group("/auth")
	{
		auth.POST("/register", h.Register)
		auth.POST("/login", h.Login)
		auth.POST("/vendor-login", h.VendorLogin)
		auth.POST("/refresh", h.Refresh)
		auth.POST("/logout", authMiddleware, h.Logout)
		auth.GET("/me", authMiddleware, h.Me)
		auth.GET("/users", authMiddleware, middleware.RequireRoles(shared.RoleAdmin, shared.RoleSuperAdmin, shared.RoleLoanOfficer), h.ListUsersByRole)
	}
}

// registerRequest is the registration payload.
type registerRequest struct {
	Email    string      `json:"email" binding:"required,email"`
	Phone    string      `json:"phone" binding:"required"`
	Password string      `json:"password" binding:"required,min=8"`
	Role     shared.Role `json:"role" binding:"required"`
}

// loginRequest is the login payload.
type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type vendorLoginRequest struct {
	Phone string `json:"phone" binding:"required"`
	PIN   string `json:"pin" binding:"required,len=4"`
}

type vendorLoginResponse struct {
	authapp.TokenPair
	VendorID     string `json:"vendor_id"`
	VendorStatus string `json:"vendor_status"`
	KYCStatus    string `json:"kyc_status"`
}

// refreshRequest is the refresh payload.
type refreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// Register godoc
// @Summary Register a new user
// @Tags auth
// @Accept json
// @Produce json
// @Param body body registerRequest true "Registration data"
// @Success 201 {object} map[string]interface{}
// @Router /auth/register [post]
func (h *Handler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.authSvc.Register(c.Request.Context(), authapp.RegisterInput{
		Email:    req.Email,
		Phone:    req.Phone,
		Password: req.Password,
		Role:     req.Role,
	})
	if err != nil {
		status := apperrors.HTTPStatus(err)
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":    user.ID,
		"email": user.Email,
		"role":  user.Role,
	})
}

// Login godoc
// @Summary Login with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param body body loginRequest true "Login credentials"
// @Success 200 {object} authapp.TokenPair
// @Router /auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pair, err := h.authSvc.Login(c.Request.Context(), authapp.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		status := apperrors.HTTPStatus(err)
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, pair)
}

// VendorLogin godoc
// @Summary Vendor login with phone and PIN
// @Tags auth
// @Accept json
// @Produce json
// @Param body body vendorLoginRequest true "Vendor credentials"
// @Success 200 {object} vendorLoginResponse
// @Router /auth/vendor-login [post]
func (h *Handler) VendorLogin(c *gin.Context) {
	var req vendorLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := h.vendorSvc.AuthenticateByPhonePIN(c.Request.Context(), req.Phone, req.PIN)
	if err != nil {
		c.JSON(apperrors.HTTPStatus(err), gin.H{"error": err.Error()})
		return
	}

	vendor, err := h.vendorSvc.GetByUserID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(apperrors.HTTPStatus(err), gin.H{"error": err.Error()})
		return
	}

	pair, err := h.authSvc.LoginUserByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(apperrors.HTTPStatus(err), gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, vendorLoginResponse{
		TokenPair:    *pair,
		VendorID:     vendor.ID.String(),
		VendorStatus: string(vendor.Status),
		KYCStatus:    string(vendor.KYCStatus),
	})
}

// Refresh godoc
// @Summary Refresh access token
// @Tags auth
// @Accept json
// @Produce json
// @Param body body refreshRequest true "Refresh token"
// @Success 200 {object} authapp.TokenPair
// @Router /auth/refresh [post]
func (h *Handler) Refresh(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pair, err := h.authSvc.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		status := apperrors.HTTPStatus(err)
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, pair)
}

// Logout godoc
// @Summary Logout and revoke all tokens
// @Tags auth
// @Security BearerAuth
// @Success 200 {object} map[string]string
// @Router /auth/logout [post]
func (h *Handler) Logout(c *gin.Context) {
	userIDStr := middleware.GetUserID(c)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	if err := h.authSvc.Logout(c.Request.Context(), userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "logged out successfully"})
}

// Me godoc
// @Summary Get current user info
// @Tags auth
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /auth/me [get]
func (h *Handler) Me(c *gin.Context) {
	userIDStr := middleware.GetUserID(c)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}
	user, err := h.authSvc.GetUserByID(c.Request.Context(), userID)
	if err != nil || user == nil {
		c.JSON(http.StatusOK, gin.H{
			"user_id": userIDStr,
			"role":    middleware.GetRole(c),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"user_id":      user.ID.String(),
		"email":        user.Email,
		"phone":        user.Phone,
		"role":         user.Role,
		"is_demo":      user.IsDemo,
		"display_name": user.DisplayName,
	})
}

// ListUsersByRole retrieves users filtered by role query parameter.
func (h *Handler) ListUsersByRole(c *gin.Context) {
	roleStr := c.Query("role")
	if roleStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "role query parameter is required"})
		return
	}

	role := shared.Role(roleStr)
	isDemo := democtx.FromGin(c)

	users, err := h.authSvc.ListUsersByRole(c.Request.Context(), role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var filtered []gin.H
	for _, u := range users {
		if u.IsDemo == isDemo {
			filtered = append(filtered, gin.H{
				"id":           u.ID.String(),
				"email":        u.Email,
				"phone":        u.Phone,
				"role":         u.Role,
				"display_name": u.DisplayName,
				"is_demo":      u.IsDemo,
			})
		}
	}

	c.JSON(http.StatusOK, filtered)
}
