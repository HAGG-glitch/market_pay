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
	"github.com/marketpay/backend/pkg/response"
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
		response.Error(c, http.StatusBadRequest, "Please check your information. Make sure your email is valid and your password is at least 8 characters.")
		return
	}

	user, err := h.authSvc.Register(c.Request.Context(), authapp.RegisterInput{
		Email:    req.Email,
		Phone:    req.Phone,
		Password: req.Password,
		Role:     req.Role,
	})
	if err != nil {
		if apperrors.IsConflict(err) {
			response.Error(c, http.StatusConflict, "An account with this email already exists. Please sign in instead.")
			return
		}
		response.ErrorFromAppError(c, err)
		return
	}

	response.Created(c, gin.H{
		"id":    user.ID,
		"email": user.Email,
		"role":  user.Role,
	}, "Account created successfully. You can now sign in.")
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
		response.Error(c, http.StatusBadRequest, "Please enter a valid email address and password.")
		return
	}

	pair, err := h.authSvc.Login(c.Request.Context(), authapp.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		status := apperrors.HTTPStatus(err)
		if status == http.StatusUnauthorized {
			msg := err.Error()
			if msg == "account is suspended" {
				response.Error(c, http.StatusUnauthorized, "Your account has been suspended. Please contact support for assistance.")
			} else {
				response.Error(c, http.StatusUnauthorized, "The email or password you entered is incorrect. Please try again.")
			}
			return
		}
		response.ErrorFromAppError(c, err)
		return
	}

	response.Success(c, pair, "Welcome back! You are now signed in.")
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
		response.Error(c, http.StatusBadRequest, "Please provide your phone number and 4-digit PIN.")
		return
	}

	userID, err := h.vendorSvc.AuthenticateByPhonePIN(c.Request.Context(), req.Phone, req.PIN)
	if err != nil {
		status := apperrors.HTTPStatus(err)
		if status == http.StatusUnauthorized {
			response.Error(c, http.StatusUnauthorized, "The phone number or PIN you entered is incorrect. Please try again.")
		} else {
			response.ErrorFromAppError(c, err)
		}
		return
	}

	vendor, err := h.vendorSvc.GetByUserID(c.Request.Context(), userID)
	if err != nil {
		response.ErrorFromAppError(c, err)
		return
	}

	pair, err := h.authSvc.LoginVendorByID(c.Request.Context(), userID, vendor)
	if err != nil {
		response.ErrorFromAppError(c, err)
		return
	}

	response.Success(c, vendorLoginResponse{
		TokenPair:    *pair,
		VendorID:     vendor.ID.String(),
		VendorStatus: string(vendor.Status),
		KYCStatus:    string(vendor.KYCStatus),
	}, "Welcome! You are now signed in.")
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
		response.Error(c, http.StatusBadRequest, "A refresh token is required.")
		return
	}

	pair, err := h.authSvc.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Your session has expired. Please sign in again.")
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
		response.Error(c, http.StatusBadRequest, "There was a problem with your session. Please try signing out again.")
		return
	}

	if err := h.authSvc.Logout(c.Request.Context(), userID); err != nil {
		response.Error(c, http.StatusInternalServerError, "Something went wrong. Please try again.")
		return
	}

	response.Success(c, nil, "You have been signed out successfully.")
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
		response.Error(c, http.StatusBadRequest, "There was a problem with your session.")
		return
	}
	user, err := h.authSvc.GetUserByID(c.Request.Context(), userID)
	if err != nil || user == nil {
		response.Success(c, gin.H{
			"user_id": userIDStr,
			"role":    middleware.GetRole(c),
		}, "")
		return
	}
	response.Success(c, gin.H{
		"user_id":      user.ID.String(),
		"email":        user.Email,
		"phone":        user.Phone,
		"role":         user.Role,
		"is_demo":      user.IsDemo,
		"display_name": user.DisplayName,
	}, "")
}

// ListUsersByRole retrieves users filtered by role query parameter.
func (h *Handler) ListUsersByRole(c *gin.Context) {
	roleStr := c.Query("role")
	if roleStr == "" {
		response.Error(c, http.StatusBadRequest, "Please specify a role to filter users by.")
		return
	}

	role := shared.Role(roleStr)
	isDemo := democtx.FromGin(c)

	users, err := h.authSvc.ListUsersByRole(c.Request.Context(), role)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Could not load users at this time. Please try again.")
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

	response.Success(c, filtered, "")
}
