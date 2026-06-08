package application

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	authmodel "github.com/marketpay/backend/internal/auth/domain/model"
	shared "github.com/marketpay/backend/internal/shared/domain/model"
	"github.com/marketpay/backend/pkg/config"
	apperrors "github.com/marketpay/backend/pkg/errors"
	"github.com/marketpay/backend/pkg/logger"
	"go.uber.org/zap"
)

// UserRepository defines persistence operations for users.
type UserRepository interface {
	Create(ctx context.Context, user *authmodel.User) error
	FindByID(ctx context.Context, id uuid.UUID) (*authmodel.User, error)
	FindByEmail(ctx context.Context, email string) (*authmodel.User, error)
	FindByPhone(ctx context.Context, phone string) (*authmodel.User, error)
	Update(ctx context.Context, user *authmodel.User) error
	SaveRefreshToken(ctx context.Context, token *authmodel.RefreshToken) error
	FindRefreshToken(ctx context.Context, token string) (*authmodel.RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, token string) error
	RevokeAllUserTokens(ctx context.Context, userID uuid.UUID) error
}

// AuthService handles authentication and token management.
type AuthService struct {
	users  UserRepository
	cfg    config.JWTConfig
	log    *logger.Logger
}

// NewAuthService constructs a new AuthService.
func NewAuthService(users UserRepository, cfg config.JWTConfig, log *logger.Logger) *AuthService {
	return &AuthService{users: users, cfg: cfg, log: log}
}

// RegisterInput holds registration data.
type RegisterInput struct {
	Email    string
	Phone    string
	Password string
	Role     shared.Role
}

// LoginInput holds login credentials.
type LoginInput struct {
	Email    string
	Password string
}

// TokenPair holds an access and refresh token.
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

// Register creates a new user account.
func (s *AuthService) Register(ctx context.Context, input RegisterInput) (*authmodel.User, error) {
	existing, _ := s.users.FindByEmail(ctx, input.Email)
	if existing != nil {
		return nil, apperrors.ErrAlreadyExists("user")
	}

	user := &authmodel.User{
		Email: input.Email,
		Phone: input.Phone,
		Role:  input.Role,
	}
	if err := user.SetPassword(input.Password); err != nil {
		return nil, apperrors.ErrInternalServer(err)
	}

	if err := s.users.Create(ctx, user); err != nil {
		return nil, apperrors.ErrInternalServer(err)
	}

	s.log.Info("user registered", zap.String("user_id", user.ID.String()), zap.String("role", string(user.Role)))
	return user, nil
}

// Login authenticates a user and returns a token pair.
func (s *AuthService) Login(ctx context.Context, input LoginInput) (*TokenPair, error) {
	user, err := s.users.FindByEmail(ctx, input.Email)
	if err != nil || user == nil {
		return nil, apperrors.ErrUnauthorized("invalid credentials")
	}

	if !user.IsActive {
		return nil, apperrors.ErrUnauthorized("account is suspended")
	}

	if !user.CheckPassword(input.Password) {
		return nil, apperrors.ErrUnauthorized("invalid credentials")
	}

	pair, err := s.issueTokenPair(ctx, user)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	user.LastLoginAt = &now
	_ = s.users.Update(ctx, user)

	return pair, nil
}

// Refresh exchanges a valid refresh token for a new token pair.
func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (*TokenPair, error) {
	rt, err := s.users.FindRefreshToken(ctx, refreshToken)
	if err != nil || rt == nil {
		return nil, apperrors.ErrUnauthorized("invalid refresh token")
	}

	if !rt.IsValid() {
		return nil, apperrors.ErrUnauthorized("refresh token expired or revoked")
	}

	user, err := s.users.FindByID(ctx, rt.UserID)
	if err != nil || user == nil {
		return nil, apperrors.ErrUnauthorized("user not found")
	}

	// Rotate the token
	if err := s.users.RevokeRefreshToken(ctx, refreshToken); err != nil {
		return nil, apperrors.ErrInternalServer(err)
	}

	return s.issueTokenPair(ctx, user)
}

// Logout revokes all tokens for a user.
func (s *AuthService) Logout(ctx context.Context, userID uuid.UUID) error {
	return s.users.RevokeAllUserTokens(ctx, userID)
}

// ValidateAccessToken parses and validates a JWT access token.
func (s *AuthService) ValidateAccessToken(tokenString string) (*authmodel.TokenClaims, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(s.cfg.AccessSecret), nil
	})

	if err != nil || !token.Valid {
		return nil, apperrors.ErrUnauthorized("invalid access token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, apperrors.ErrUnauthorized("invalid token claims")
	}

	userIDStr, _ := claims["user_id"].(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, apperrors.ErrUnauthorized("invalid user ID in token")
	}

	return &authmodel.TokenClaims{
		UserID: userID,
		Email:  claims["email"].(string),
		Role:   shared.Role(claims["role"].(string)),
	}, nil
}

func (s *AuthService) issueTokenPair(ctx context.Context, user *authmodel.User) (*TokenPair, error) {
	now := time.Now()
	accessExpiry := now.Add(s.cfg.AccessExpiry)

	// Access token
	accessClaims := jwt.MapClaims{
		"user_id": user.ID.String(),
		"email":   user.Email,
		"role":    string(user.Role),
		"iat":     now.Unix(),
		"exp":     accessExpiry.Unix(),
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenStr, err := accessToken.SignedString([]byte(s.cfg.AccessSecret))
	if err != nil {
		return nil, apperrors.ErrInternalServer(err)
	}

	// Refresh token
	refreshTokenID := uuid.New().String()
	refreshClaims := jwt.MapClaims{
		"jti":     refreshTokenID,
		"user_id": user.ID.String(),
		"iat":     now.Unix(),
		"exp":     now.Add(s.cfg.RefreshExpiry).Unix(),
	}
	refreshJWT := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenStr, err := refreshJWT.SignedString([]byte(s.cfg.RefreshSecret))
	if err != nil {
		return nil, apperrors.ErrInternalServer(err)
	}

	rt := &authmodel.RefreshToken{
		UserID:    user.ID,
		Token:     refreshTokenStr,
		ExpiresAt: now.Add(s.cfg.RefreshExpiry),
	}
	if err := s.users.SaveRefreshToken(ctx, rt); err != nil {
		return nil, apperrors.ErrInternalServer(err)
	}

	return &TokenPair{
		AccessToken:  accessTokenStr,
		RefreshToken: refreshTokenStr,
		ExpiresIn:    int64(s.cfg.AccessExpiry.Seconds()),
	}, nil
}
