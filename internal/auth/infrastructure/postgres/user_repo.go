package postgres

import (
	"context"

	"github.com/google/uuid"
	authmodel "github.com/marketpay/backend/internal/auth/domain/model"
	"gorm.io/gorm"
)

// UserRepo is the GORM implementation of UserRepository.
type UserRepo struct {
	db *gorm.DB
}

// NewUserRepo constructs a UserRepo.
func NewUserRepo(db *gorm.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) Create(ctx context.Context, user *authmodel.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *UserRepo) FindByID(ctx context.Context, id uuid.UUID) (*authmodel.User, error) {
	var user authmodel.User
	err := r.db.WithContext(ctx).First(&user, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepo) FindByEmail(ctx context.Context, email string) (*authmodel.User, error) {
	var user authmodel.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepo) FindByPhone(ctx context.Context, phone string) (*authmodel.User, error) {
	var user authmodel.User
	err := r.db.WithContext(ctx).Where("phone = ?", phone).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepo) Update(ctx context.Context, user *authmodel.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *UserRepo) SaveRefreshToken(ctx context.Context, token *authmodel.RefreshToken) error {
	return r.db.WithContext(ctx).Create(token).Error
}

func (r *UserRepo) FindRefreshToken(ctx context.Context, token string) (*authmodel.RefreshToken, error) {
	var rt authmodel.RefreshToken
	err := r.db.WithContext(ctx).Where("token = ? AND revoked = false", token).First(&rt).Error
	if err != nil {
		return nil, err
	}
	return &rt, nil
}

func (r *UserRepo) RevokeRefreshToken(ctx context.Context, token string) error {
	return r.db.WithContext(ctx).
		Model(&authmodel.RefreshToken{}).
		Where("token = ?", token).
		Update("revoked", true).Error
}

func (r *UserRepo) RevokeAllUserTokens(ctx context.Context, userID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&authmodel.RefreshToken{}).
		Where("user_id = ?", userID).
		Update("revoked", true).Error
}
