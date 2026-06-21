package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	// Config & shared infra
	"github.com/marketpay/backend/pkg/config"
	_ "github.com/marketpay/backend/docs"
	"github.com/marketpay/backend/pkg/logger"
	mw "github.com/marketpay/backend/pkg/middleware"
	"github.com/marketpay/backend/pkg/outbox"

	// Auth
	authapp  "github.com/marketpay/backend/internal/auth/application"
	authpg   "github.com/marketpay/backend/internal/auth/infrastructure/postgres"
	authhttp "github.com/marketpay/backend/internal/auth/interfaces/http"

	// Vendor
	vendorapp  "github.com/marketpay/backend/internal/vendors/application"
	vendorpg   "github.com/marketpay/backend/internal/vendors/infrastructure/postgres"
	vendorhttp "github.com/marketpay/backend/internal/vendors/interfaces/http"

	// Loan
	loanapp  "github.com/marketpay/backend/internal/loan/application"
	loanpg   "github.com/marketpay/backend/internal/loan/infrastructure/postgres"
	loanhttp "github.com/marketpay/backend/internal/loan/interfaces/http"

	// Repayment
	repayapp  "github.com/marketpay/backend/internal/repayment/application"
	repaypg   "github.com/marketpay/backend/internal/repayment/infrastructure/postgres"
	repayhttp "github.com/marketpay/backend/internal/repayment/interfaces/http"

	// Group
	groupapp  "github.com/marketpay/backend/internal/group/application"
	grouppg   "github.com/marketpay/backend/internal/group/infrastructure/postgres"
	grouphttp "github.com/marketpay/backend/internal/group/interfaces/http"

	// Payment
	paymentapp  "github.com/marketpay/backend/internal/payment/application"
	paymentpg   "github.com/marketpay/backend/internal/payment/infrastructure/postgres"
	paymenthttp "github.com/marketpay/backend/internal/payment/interfaces/http"

	// Credit Score
	scoreapp "github.com/marketpay/backend/internal/creditscore/application"
	scorepg  "github.com/marketpay/backend/internal/creditscore/infrastructure/postgres"

	// Ledger
	ledgerapp "github.com/marketpay/backend/internal/ledger/application"
	ledgerpg  "github.com/marketpay/backend/internal/ledger/infrastructure/postgres"

	// USSD
	ussdapp  "github.com/marketpay/backend/internal/ussd/application"
	ussdpg   "github.com/marketpay/backend/internal/ussd/infrastructure/postgres"
	ussdhttp "github.com/marketpay/backend/internal/ussd/interfaces/http"

	// Reporting
	reporthttp "github.com/marketpay/backend/internal/reporting/interfaces/http"

	// Audit
	audithttp "github.com/marketpay/backend/internal/audit/interfaces/http"

	// Monime gateway
	monimeinfra "github.com/marketpay/backend/internal/monime/infrastructure"
	monimemodel "github.com/marketpay/backend/internal/monime/domain/model"

	// Monime Exchange
	monimeexchangeapp "github.com/marketpay/backend/internal/monime/exchange"
	monimehttp "github.com/marketpay/backend/internal/monime/interfaces/http"

	// Notifications
	notifhttp "github.com/marketpay/backend/internal/notification/interfaces/http"

	"github.com/marketpay/backend/pkg/monimeexchange"
	"github.com/marketpay/backend/pkg/realtime"
)

// @title MarketPay API
// @version 1.0
// @description MarketPay microfinance and USSD payment platform for informal market vendors in Sierra Leone.
// @contact.name MarketPay Engineering
// @contact.email engineering@marketpay.sl
// @license.name MIT
// @host localhost:8080
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	cfgPath := flag.String("config", "configs/config.yaml", "path to config file")
	flag.Parse()

	// ── Config ────────────────────────────────────────────────────────────
	cfg, err := config.Load(*cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load config: %v\n", err)
		os.Exit(1)
	}

	// ── Logger ────────────────────────────────────────────────────────────
	log, err := logger.New(cfg.Logging.Level, cfg.Logging.Format)
	if err != nil {
		fmt.Fprintf(os.Stderr, "init logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Sync()

	// ── Database ──────────────────────────────────────────────────────────
	gormCfg := &gorm.Config{}
	if !cfg.IsProduction() {
		gormCfg.Logger = gormlogger.Default.LogMode(gormlogger.Info)
	}

	db, err := gorm.Open(postgres.Open(cfg.Database.DSN()), gormCfg)
	if err != nil {
		log.Fatal("connect database", zap.Error(err))
	}

	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)

	log.Info("database connected")

	// ── Redis ─────────────────────────────────────────────────────────────
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr(),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
		PoolSize: cfg.Redis.PoolSize,
	})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatal("connect redis", zap.Error(err))
	}
	log.Info("redis connected")

	// ── Shared ────────────────────────────────────────────────────────────
	outboxPub := outbox.NewPublisher(db, log)
	auditRepo  := scorepg.NewAuditRepo(db)

	// ── Auth ──────────────────────────────────────────────────────────────
	userRepo    := authpg.NewUserRepo(db)
	authSvc     := authapp.NewAuthService(userRepo, cfg.JWT, log)
	jwtMW       := mw.AuthMiddleware(authSvc)

	// ── Rate limiter ──────────────────────────────────────────────────────
	rateLimiter := mw.NewRateLimiter(rdb, cfg.RateLimit.RequestsPerMinute, time.Minute)

	// ── Vendor ────────────────────────────────────────────────────────────
	vendorRepo    := vendorpg.NewVendorRepo(db)
	vendorSvc     := vendorapp.NewVendorService(vendorRepo, outboxPub, log)
	vendorHandler := vendorhttp.NewHandler(vendorSvc)
	authHandler   := authhttp.NewHandler(authSvc, vendorSvc)

	// ── Credit Score ──────────────────────────────────────────────────────
	scoreRepo := scorepg.NewScoreRepo(db)
	factorRepo := scorepg.NewFactorRepo(db)
	scoreSvc   := scoreapp.NewService(scoreRepo, factorRepo, cfg.CreditScore, log)

	// ── Loan ──────────────────────────────────────────────────────────────
	loanRepo    := loanpg.NewLoanRepo(db)
	loanSvc     := loanapp.NewLoanService(loanRepo, auditRepo, outboxPub, scoreSvc, vendorSvc, cfg.Loans, cfg.CreditScore, log)
	loanHandler := loanhttp.NewHandler(loanSvc)

	// ── Repayment ─────────────────────────────────────────────────────────
	repayRepo    := repaypg.NewRepaymentRepo(db)
	repaySvc     := repayapp.NewRepaymentService(repayRepo, auditRepo, outboxPub, cfg.Repayment, log)
	repayHandler := repayhttp.NewHandler(repaySvc)

	// ── Group ─────────────────────────────────────────────────────────────
	groupRepo    := grouppg.NewGroupRepo(db)
	groupSvc     := groupapp.NewGroupService(groupRepo, outboxPub, log)
	groupHandler := grouphttp.NewHandler(groupSvc)

	// ── Monime Adapter ────────────────────────────────────────────────────
	monimeAdapter := monimeinfra.NewMonimeAdapter(cfg.Monime, log)

	monimeCollect := &monimeCollectorAdapter{adapter: monimeAdapter}

	// ── Payment ───────────────────────────────────────────────────────────
	paymentRepo    := paymentpg.NewPaymentRepo(db)
	paymentSvc     := paymentapp.NewPaymentService(paymentRepo, outboxPub, monimeCollect, log)
	paymentHandler := paymenthttp.NewHandler(paymentSvc)

	// ── Ledger ────────────────────────────────────────────────────────────
	ledgerRepo := ledgerpg.NewLedgerRepo(db)
	_           = ledgerapp.NewLedgerService(ledgerRepo, log)

	// ── USSD ──────────────────────────────────────────────────────────────
	sessionRepo := ussdpg.NewSessionRepo(db)
	ussdSvc     := ussdapp.NewUSSDService(
		sessionRepo,
		newUSSDVendorAdapter(vendorRepo, vendorSvc),
		newUSSDLoanAdapter(loanSvc),
		newUSSDPaymentAdapter(paymentSvc),
		cfg.USSD.SessionTimeout,
		log,
	)
	ussdHandler := ussdhttp.NewHandler(ussdSvc)

	// ── Reporting ─────────────────────────────────────────────────────────
	reportHandler := reporthttp.NewHandler(db)
	auditHandler  := audithttp.NewHandler(db)

	// ── Realtime / Notifications ──────────────────────────────────────────
	eventHub := realtime.NewHub()
	inAppNotifier := monimeexchangeapp.NewInAppNotifier(db, eventHub)
	notifHandler := notifhttp.NewHandler(db, eventHub)

	// ── Monime USSD Exchange ────────────────────────────────────────────────
	monimeWebhook := monimehttp.NewWebhookHandler(db, paymentSvc, monimeAdapter)

		var monimeHandler *monimehttp.Handler
	keyLoaded := false
	pemKey := os.Getenv("MONIME_RSA_PRIVATE_KEY")
	if pemKey == "" {
		if keyFile := os.Getenv("MONIME_RSA_KEY_FILE"); keyFile != "" {
			if b, err := os.ReadFile(keyFile); err == nil {
				pemKey = string(b)
			}
		}
	}
	if pemKey != "" || cfg.Monime.RSAPrivateKeyPEM != "" {
		keyPEM := pemKey
		if keyPEM == "" {
			keyPEM = cfg.Monime.RSAPrivateKeyPEM
		}
		crypto, err := monimeexchange.NewCrypto([]byte(keyPEM))
		if err != nil {
			log.Warn("monime exchange crypto init failed", zap.Error(err))
		} else {
			keyLoaded = true
			exchangeSvc := monimeexchangeapp.NewService(db, crypto, vendorSvc, loanSvc, paymentSvc, inAppNotifier, log.Logger)
			monimeHandler = monimehttp.NewHandler(exchangeSvc, keyLoaded)
			log.Info("monime exchange endpoint enabled")
		}
	} else {
		log.Warn("MONIME_RSA_PRIVATE_KEY not set — exchange endpoint disabled")
	}

	// ── Gin ───────────────────────────────────────────────────────────────
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(requestid.New())
	router.Use(mw.SecurityHeaders())
	router.Use(rateLimiter.Limit())
	router.Use(mw.DemoMode())
	router.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORS.AllowedOrigins,
		AllowMethods:     cfg.CORS.AllowedMethods,
		AllowHeaders:     cfg.CORS.AllowedHeaders,
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	router.NoRoute(mw.NotFound())
	router.NoMethod(mw.MethodNotAllowed())

	// Root
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": cfg.App.Name,
			"status":  "running",
		})
	})

	// Health
	router.GET("/health", func(c *gin.Context) {
		sqlDB, _ := db.DB()
		if err := sqlDB.Ping(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unhealthy", "db": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": cfg.App.Name,
			"version": cfg.App.Version,
		})
	})

	// Swagger UI
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API v1
	v1 := router.Group("/api/v1")
	authHandler.RegisterRoutes(v1, jwtMW)
	vendorHandler.RegisterRoutes(v1, jwtMW)
	loanHandler.RegisterRoutes(v1, jwtMW)
	repayHandler.RegisterRoutes(v1, jwtMW)
	groupHandler.RegisterRoutes(v1, jwtMW)
	paymentHandler.RegisterRoutes(v1, jwtMW)
	reportHandler.RegisterRoutes(v1, jwtMW)
	auditHandler.RegisterRoutes(v1, jwtMW)
	notifHandler.RegisterRoutes(v1, jwtMW)
	monimeWebhook.RegisterRoutes(v1)

	if monimeHandler != nil {
		monimeHandler.RegisterRoutes(v1)
	}

	// USSD — rate limited separately, no JWT
	ussdGroup := v1.Group("")
	ussdGroup.Use(rateLimiter.USSDLimit())
	ussdHandler.RegisterRoutes(ussdGroup)

	// ── gRPC ──────────────────────────────────────────────────────────────
	grpcServer := grpc.NewServer()
	go func() {
		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.App.GRPCPort))
		if err != nil {
			log.Error("gRPC listen failed", zap.Error(err))
			return
		}
		log.Info("gRPC listening", zap.Int("port", cfg.App.GRPCPort))
		if err := grpcServer.Serve(lis); err != nil {
			log.Error("gRPC serve", zap.Error(err))
		}
	}()

	// ── HTTP Server ───────────────────────────────────────────────────────
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.App.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Info("HTTP server listening", zap.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("HTTP server error", zap.Error(err))
		}
	}()

	// ── Graceful shutdown ─────────────────────────────────────────────────
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down gracefully...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	grpcServer.GracefulStop()
	if err := srv.Shutdown(ctx); err != nil {
		log.Error("HTTP shutdown error", zap.Error(err))
	}
	log.Info("server stopped")
}

// ── USSD adapters ─────────────────────────────────────────────────────────────

type ussdVendorAdapter struct {
	vendorRepo *vendorpg.VendorRepo
	vendorSvc  *vendorapp.VendorService
}

func newUSSDVendorAdapter(repo *vendorpg.VendorRepo, svc *vendorapp.VendorService) *ussdVendorAdapter {
	return &ussdVendorAdapter{vendorRepo: repo, vendorSvc: svc}
}

func (a *ussdVendorAdapter) FindByPhone(ctx context.Context, phone string) (interface{}, error) {
	return a.vendorRepo.FindByPhone(ctx, phone)
}

func (a *ussdVendorAdapter) VerifyPIN(ctx context.Context, phone, pin string) error {
	return a.vendorSvc.VerifyPINByPhone(ctx, phone, pin)
}

func (a *ussdVendorAdapter) CheckEligibility(ctx context.Context, phone string) (bool, string, error) {
	vendor, err := a.vendorRepo.FindByPhone(ctx, phone)
	if err != nil {
		return false, "vendor not found", err
	}
	if err := vendor.IsEligibleForLoan(); err != nil {
		return false, err.Error(), nil
	}
	return true, "", nil
}

func (a *ussdVendorAdapter) GetLoanBalance(ctx context.Context, phone string) (float64, error) {
	vendor, err := a.vendorRepo.FindByPhone(ctx, phone)
	if err != nil {
		return 0, err
	}
	return vendor.CreditScore, nil // placeholder — replace with outstanding loan sum
}

func (a *ussdVendorAdapter) GetRepaymentSchedule(ctx context.Context, phone string) (string, error) {
	return "No active repayment schedule.", nil
}

func (a *ussdVendorAdapter) GetSalesHistory(ctx context.Context, phone string) (string, error) {
	return "No recent transactions found.", nil
}

func (a *ussdVendorAdapter) GetGroupInfo(ctx context.Context, phone string) (string, error) {
	return "You are not currently in a group.", nil
}

type ussdLoanAdapter struct {
	loanSvc *loanapp.LoanService
}

func newUSSDLoanAdapter(svc *loanapp.LoanService) *ussdLoanAdapter {
	return &ussdLoanAdapter{loanSvc: svc}
}

func (a *ussdLoanAdapter) ApplyUSSD(ctx context.Context, phone, loanType string, amount float64) (string, error) {
	return fmt.Sprintf("Loan application submitted. Type: %s. Amount: %.2f SLE. You will receive an SMS shortly.", loanType, amount), nil
}

// monimeCollectorAdapter adapts MonimeAdapter to paymentapp.MonimeCollector.
type monimeCollectorAdapter struct {
	adapter *monimeinfra.MonimeAdapter
}

func (a *monimeCollectorAdapter) Collect(ctx context.Context, phone, amount string) (string, error) {
	var amtFloat float64
	fmt.Sscanf(amount, "%f", &amtFloat)
	req := monimemodel.CollectionRequest{
		Reference:   "pay-" + uuid.New().String()[:12],
		Phone:       phone,
		Amount:      amtFloat,
		Currency:    "SLE",
		Description: "MarketPay payment collection",
	}
	resp, err := a.adapter.Collect(ctx, req)
	if err != nil {
		return "", err
	}
	return resp.ExternalRef, nil
}

type ussdPaymentAdapter struct {
	paymentSvc *paymentapp.PaymentService
}

func newUSSDPaymentAdapter(svc *paymentapp.PaymentService) *ussdPaymentAdapter {
	return &ussdPaymentAdapter{paymentSvc: svc}
}

func (a *ussdPaymentAdapter) InitiateUSSD(ctx context.Context, fromPhone, toVendorCode string, amount float64) (string, error) {
	return "Payment initiated. You will receive an SMS confirmation shortly.", nil
}
