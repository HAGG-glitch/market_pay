package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	authapp "github.com/marketpay/backend/internal/auth/application"
	shared "github.com/marketpay/backend/internal/shared/domain/model"
	apperrors "github.com/marketpay/backend/pkg/errors"
	"net/http"
)

const (
	ContextKeyUserID       = "user_id"
	ContextKeyRole         = "user_role"
	ContextKeyClaims       = "claims"
	ContextKeyVendorStatus = "vendor_status"
	ContextKeyKYCStatus    = "kyc_status"
)

// AuthMiddleware validates JWT access tokens.
func AuthMiddleware(authSvc *authapp.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			respondError(c, apperrors.ErrUnauthorized("missing authorization header"))
			return
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
			respondError(c, apperrors.ErrUnauthorized("invalid authorization header format"))
			return
		}

		claims, err := authSvc.ValidateAccessToken(parts[1])
		if err != nil {
			respondError(c, err)
			return
		}

		c.Set(ContextKeyUserID, claims.UserID.String())
		c.Set(ContextKeyRole, string(claims.Role))
		c.Set(ContextKeyClaims, claims)
		c.Set(ContextKeyVendorStatus, claims.VendorStatus)
		c.Set(ContextKeyKYCStatus, claims.KYCStatus)
		c.Next()
	}
}

// RequireRoles enforces role-based access control.
func RequireRoles(roles ...shared.Role) gin.HandlerFunc {
	allowed := make(map[shared.Role]bool, len(roles))
	for _, r := range roles {
		allowed[r] = true
	}

	return func(c *gin.Context) {
		roleStr, exists := c.Get(ContextKeyRole)
		if !exists {
			respondError(c, apperrors.ErrUnauthorized("no role in context"))
			return
		}

		role := shared.Role(roleStr.(string))
		if !allowed[role] {
			respondError(c, apperrors.ErrForbidden("insufficient permissions"))
			return
		}
		c.Next()
	}
}

// RequireActiveVendor blocks pending vendors from accessing the route.
// Vendors with status PENDING can only reach this point if explicitly allowed.
func RequireActiveVendor() gin.HandlerFunc {
	return func(c *gin.Context) {
		role := GetRole(c)
		if role != shared.RoleVendor {
			c.Next()
			return
		}
		vendorStatus, _ := c.Get(ContextKeyVendorStatus)
		if vendorStatus == "PENDING" {
			respondError(c, apperrors.ErrForbidden("account pending approval"))
			return
		}
		c.Next()
	}
}

func respondError(c *gin.Context, err error) {
	status := apperrors.HTTPStatus(err)
	c.AbortWithStatusJSON(status, gin.H{"error": err.Error()})
}

// SecurityHeaders adds security-related HTTP headers.
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Next()
	}
}

// RequestLogger logs incoming requests using the Gin default logger fields.
func RequestLogger() gin.HandlerFunc {
	return gin.Logger()
}

// GetUserID extracts user ID string from gin context.
func GetUserID(c *gin.Context) string {
	id, _ := c.Get(ContextKeyUserID)
	if id == nil {
		return ""
	}
	return id.(string)
}

// GetRole extracts role from gin context.
func GetRole(c *gin.Context) shared.Role {
	r, _ := c.Get(ContextKeyRole)
	if r == nil {
		return ""
	}
	return shared.Role(r.(string))
}

// NotFound handles 404s uniformly.
func NotFound() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "route not found"})
	}
}

// MethodNotAllowed handles 405s uniformly.
func MethodNotAllowed() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "method not allowed"})
	}
}
