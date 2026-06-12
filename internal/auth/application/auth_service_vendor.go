package application

import (
	"context"

	"github.com/google/uuid"
	apperrors "github.com/marketpay/backend/pkg/errors"
)

// LoginUserByID issues tokens for an already-authenticated user.
func (s *AuthService) LoginUserByID(ctx context.Context, userID uuid.UUID) (*TokenPair, error) {
	user, err := s.users.FindByID(ctx, userID)
	if err != nil || user == nil {
		return nil, apperrors.ErrUnauthorized("user not found")
	}
	if !user.IsActive {
		return nil, apperrors.ErrUnauthorized("account is suspended")
	}
	return s.issueTokenPair(ctx, user)
}
