package application_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	authapp   "github.com/marketpay/backend/internal/auth/application"
	authmodel "github.com/marketpay/backend/internal/auth/domain/model"
	shared    "github.com/marketpay/backend/internal/shared/domain/model"
	"github.com/marketpay/backend/pkg/config"
	"github.com/marketpay/backend/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// ── Mock ──────────────────────────────────────────────────────────────────────

type mockUserRepo struct{ mock.Mock }

func (m *mockUserRepo) Create(ctx context.Context, user *authmodel.User) error {
	args := m.Called(ctx, user)
	user.ID = uuid.New()
	return args.Error(0)
}
func (m *mockUserRepo) FindByID(ctx context.Context, id uuid.UUID) (*authmodel.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil { return nil, args.Error(1) }
	return args.Get(0).(*authmodel.User), args.Error(1)
}
func (m *mockUserRepo) FindByEmail(ctx context.Context, email string) (*authmodel.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil { return nil, args.Error(1) }
	return args.Get(0).(*authmodel.User), args.Error(1)
}
func (m *mockUserRepo) FindByPhone(ctx context.Context, phone string) (*authmodel.User, error) {
	args := m.Called(ctx, phone)
	if args.Get(0) == nil { return nil, args.Error(1) }
	return args.Get(0).(*authmodel.User), args.Error(1)
}
func (m *mockUserRepo) Update(ctx context.Context, user *authmodel.User) error {
	return m.Called(ctx, user).Error(0)
}
func (m *mockUserRepo) SaveRefreshToken(ctx context.Context, token *authmodel.RefreshToken) error {
	return m.Called(ctx, token).Error(0)
}
func (m *mockUserRepo) FindRefreshToken(ctx context.Context, token string) (*authmodel.RefreshToken, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil { return nil, args.Error(1) }
	return args.Get(0).(*authmodel.RefreshToken), args.Error(1)
}
func (m *mockUserRepo) RevokeRefreshToken(ctx context.Context, token string) error {
	return m.Called(ctx, token).Error(0)
}
func (m *mockUserRepo) RevokeAllUserTokens(ctx context.Context, userID uuid.UUID) error {
	return m.Called(ctx, userID).Error(0)
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func newTestAuthService(repo authapp.UserRepository) *authapp.AuthService {
	cfg := config.JWTConfig{
		AccessSecret:  "test-access-secret",
		RefreshSecret: "test-refresh-secret",
		AccessExpiry:  15 * time.Minute,
		RefreshExpiry: 7 * 24 * time.Hour,
	}
	return authapp.NewAuthService(repo, cfg, logger.NewNop())
}

// ── Tests ─────────────────────────────────────────────────────────────────────

func TestAuthService_Register_Success(t *testing.T) {
	repo := &mockUserRepo{}
	ctx  := context.Background()

	repo.On("FindByEmail", ctx, "test@marketpay.sl").Return(nil, nil)
	repo.On("Create", ctx, mock.AnythingOfType("*model.User")).Return(nil)

	svc  := newTestAuthService(repo)
	user, err := svc.Register(ctx, authapp.RegisterInput{
		Email:    "test@marketpay.sl",
		Phone:    "+23276123456",
		Password: "SecurePass123",
		Role:     shared.RoleVendor,
	})

	require.NoError(t, err)
	assert.Equal(t, "test@marketpay.sl", user.Email)
	assert.Equal(t, shared.RoleVendor, user.Role)
	repo.AssertExpectations(t)
}

func TestAuthService_Register_DuplicateEmail(t *testing.T) {
	repo := &mockUserRepo{}
	ctx  := context.Background()

	existing := &authmodel.User{Email: "test@marketpay.sl"}
	repo.On("FindByEmail", ctx, "test@marketpay.sl").Return(existing, nil)

	svc := newTestAuthService(repo)
	_, err := svc.Register(ctx, authapp.RegisterInput{
		Email:    "test@marketpay.sl",
		Password: "SecurePass123",
		Role:     shared.RoleVendor,
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
	repo.AssertNotCalled(t, "Create")
}

func TestAuthService_Login_Success(t *testing.T) {
	repo := &mockUserRepo{}
	ctx  := context.Background()

	user := &authmodel.User{
		Email:    "test@marketpay.sl",
		Role:     shared.RoleVendor,
		IsActive: true,
	}
	user.ID = uuid.New()
	_ = user.SetPassword("SecurePass123")

	repo.On("FindByEmail", ctx, "test@marketpay.sl").Return(user, nil)
	repo.On("Update", ctx, mock.Anything).Return(nil)
	repo.On("SaveRefreshToken", ctx, mock.Anything).Return(nil)

	svc  := newTestAuthService(repo)
	pair, err := svc.Login(ctx, authapp.LoginInput{
		Email:    "test@marketpay.sl",
		Password: "SecurePass123",
	})

	require.NoError(t, err)
	assert.NotEmpty(t, pair.AccessToken)
	assert.NotEmpty(t, pair.RefreshToken)
	assert.Equal(t, int64(900), pair.ExpiresIn) // 15 min = 900 sec
}

func TestAuthService_Login_WrongPassword(t *testing.T) {
	repo := &mockUserRepo{}
	ctx  := context.Background()

	user := &authmodel.User{Email: "test@marketpay.sl", IsActive: true}
	user.ID = uuid.New()
	_ = user.SetPassword("CorrectPassword")

	repo.On("FindByEmail", ctx, "test@marketpay.sl").Return(user, nil)

	svc := newTestAuthService(repo)
	_, err := svc.Login(ctx, authapp.LoginInput{
		Email:    "test@marketpay.sl",
		Password: "WrongPassword",
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid credentials")
}

func TestAuthService_Login_InactiveUser(t *testing.T) {
	repo := &mockUserRepo{}
	ctx  := context.Background()

	user := &authmodel.User{Email: "test@marketpay.sl", IsActive: false}
	user.ID = uuid.New()
	_ = user.SetPassword("SecurePass123")

	repo.On("FindByEmail", ctx, "test@marketpay.sl").Return(user, nil)

	svc := newTestAuthService(repo)
	_, err := svc.Login(ctx, authapp.LoginInput{
		Email:    "test@marketpay.sl",
		Password: "SecurePass123",
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "suspended")
}

func TestAuthService_ValidateToken_Valid(t *testing.T) {
	repo := &mockUserRepo{}
	ctx  := context.Background()

	user := &authmodel.User{
		Email:    "test@marketpay.sl",
		Role:     shared.RoleAdmin,
		IsActive: true,
	}
	user.ID = uuid.New()
	_ = user.SetPassword("SecurePass123")

	repo.On("FindByEmail", ctx, "test@marketpay.sl").Return(user, nil)
	repo.On("Update", ctx, mock.Anything).Return(nil)
	repo.On("SaveRefreshToken", ctx, mock.Anything).Return(nil)

	svc  := newTestAuthService(repo)
	pair, err := svc.Login(ctx, authapp.LoginInput{
		Email:    "test@marketpay.sl",
		Password: "SecurePass123",
	})
	require.NoError(t, err)

	claims, err := svc.ValidateAccessToken(pair.AccessToken)
	require.NoError(t, err)
	assert.Equal(t, user.ID, claims.UserID)
	assert.Equal(t, shared.RoleAdmin, claims.Role)
}

func TestAuthService_ValidateToken_Invalid(t *testing.T) {
	svc := newTestAuthService(&mockUserRepo{})
	_, err := svc.ValidateAccessToken("not.a.real.token")
	assert.Error(t, err)
}
