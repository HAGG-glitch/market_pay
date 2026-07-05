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
	"github.com/marketpay/backend/pkg/monimepayout"
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

	// ── Ensure vendors under Joshua are eligible for testing ─────────────
	db.Exec(`UPDATE vendors SET status = 'ACTIVE', kyc_status = 'VERIFIED', first_transaction_at = NOW() - INTERVAL '60 days', credit_score = 80, updated_at = NOW() WHERE field_agent_id = 'b8f1077b-c186-40d1-8889-e3c10cad7fa8' AND deleted_at IS NULL`)
	log.Info("vendor eligibility updated for testing")

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

	// ── Monime Payout ────────────────────────────────────────────────────
	payoutClient := monimepayout.NewClient(monimepayout.Config{
		BaseURL:            cfg.Monime.Payout.BaseURL,
		APIKey:             cfg.Monime.Payout.APIKey,
		SpaceID:            cfg.Monime.Payout.SpaceID,
		FinancialAccountID: cfg.Monime.Payout.FinancialAccountID,
		ProviderID:         cfg.Monime.Payout.ProviderID,
		Timeout:            cfg.Monime.Payout.Timeout,
	})

	// ── Loan ──────────────────────────────────────────────────────────────
	loanRepo    := loanpg.NewLoanRepo(db)
	vendorPhoneAdapter := &vendorPhoneFinderAdapter{vendorSvc: vendorSvc}
	payoutAdapter := &monimePayoutAdapter{
		client: payoutClient,
		cfg:    cfg.Monime.Payout,
	}
	loanSvc := loanapp.NewLoanService(loanRepo, auditRepo, outboxPub, scoreSvc, vendorSvc, vendorPhoneAdapter, payoutAdapter, cfg.Loans, cfg.CreditScore, log)

	// ── Repayment ─────────────────────────────────────────────────────────
	repayRepo    := repaypg.NewRepaymentRepo(db)
	repaySvc     := repayapp.NewRepaymentService(repayRepo, auditRepo, outboxPub, cfg.Repayment, log)
	repayHandler := repayhttp.NewHandler(repaySvc)

	loanHandler := loanhttp.NewHandler(loanSvc, vendorSvc, repaySvc)

	// ── Group (coming soon) ──────────────────────────────────────────────
	// Temporarily disabled until the group pay feature is properly planned.

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
	monimeWebhook := monimehttp.NewWebhookHandler(db, paymentSvc, monimeAdapter, loanSvc, repaySvc, log)

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
			exchangeSvc := monimeexchangeapp.NewService(db, crypto, vendorSvc, loanSvc, paymentSvc, repaySvc, inAppNotifier, log.Logger)
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
		AllowWildcard:    true,
		MaxAge:           12 * time.Hour,
	}))
	router.NoRoute(mw.NotFound())
	router.NoMethod(mw.MethodNotAllowed())

	// Root — status page with USSD cooldown timer and keep-alive tracker
	startTime := time.Now()
	router.GET("/", func(c *gin.Context) {
		c.Header("Content-Type", "text/html; charset=utf-8")
		startUnix := startTime.UnixMilli()
		appName := cfg.App.Name
		appVersion := cfg.App.Version
		appEnv := cfg.App.Env
		c.String(http.StatusOK, `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>`+appName+` Status</title>
<link rel="icon" href="data:image/svg+xml,<svg xmlns=%22http://www.w3.org/2000/svg%22 viewBox=%220 0 100 100%22><text y=%22.9em%22 font-size=%2290%22>📊</text></svg>">
<style>
  *{margin:0;padding:0;box-sizing:border-box}
  body{font-family:-apple-system,BlinkMacSystemFont,"Segoe UI","Noto Sans",Helvetica,Arial,sans-serif;background:#0d1117;color:#e6edf3;display:flex;flex-direction:column;align-items:center;justify-content:center;min-height:100vh;padding:20px}
  .card{background:#161b22;border:1px solid #30363d;border-radius:8px;padding:40px;max-width:620px;width:100%;text-align:center}
  .status-dot{display:inline-block;width:14px;height:14px;border-radius:50%;margin-right:8px;vertical-align:middle}
  .dot-green{background:#3fb950;box-shadow:0 0 8px #3fb95088}
  .dot-yellow{background:#d29922;box-shadow:0 0 8px #d2992288}
  .dot-red{background:#f85149;box-shadow:0 0 8px #f8514988}
  h1{font-size:24px;font-weight:600;margin-bottom:4px}
  .subtitle{color:#8b949e;font-size:14px;margin-bottom:24px}
  .ussd-section{background:#1c2333;border:1px solid #30363d;border-radius:8px;padding:24px;margin-bottom:20px}
  .ussd-section h2{font-size:16px;font-weight:600;margin-bottom:12px;color:#e6edf3}
  .timer{font-size:42px;font-weight:700;font-variant-numeric:tabular-nums;letter-spacing:2px}
  .timer-green{color:#3fb950}
  .timer-amber{color:#d29922}
  .ussd-ready-text{font-size:15px;color:#8b949e;margin-bottom:12px}
  .ussd-code{display:inline-block;background:#0d1117;border:1px solid #30363d;border-radius:6px;padding:12px 24px;font-size:22px;font-weight:700;font-family:"SFMono-Regular",Consolas,"Liberation Mono",Menlo,Courier,monospace;color:#3fb950;letter-spacing:1px;margin-top:8px}
  .component{display:flex;align-items:center;justify-content:space-between;padding:12px 16px;border:1px solid #30363d;border-radius:6px;margin-bottom:8px;text-align:left}
  .component-name{font-size:14px;font-weight:500}
  .component-status{font-size:13px}
  .keepalive-row{display:flex;align-items:center;justify-content:space-between;padding:10px 16px;background:#1c2333;border:1px solid #30363d;border-radius:6px;margin-top:16px;font-size:13px;color:#8b949e}
   .codes-table{width:100%;border-collapse:collapse;margin-top:12px;font-size:13px}
   .codes-table th{text-align:left;padding:8px 12px;font-size:11px;text-transform:uppercase;letter-spacing:0.5px;color:#8b949e;border-bottom:1px solid #30363d}
   .codes-table td{padding:10px 12px;border-bottom:1px solid #21262d}
   .codes-table tr:last-child td{border-bottom:none}
   .codes-table .code{font-family:"SFMono-Regular",Consolas,"Liberation Mono",Menlo,Courier,monospace;color:#3fb950;font-weight:600;letter-spacing:0.5px}
   .codes-table .purpose{color:#8b949e;font-size:12px}
   .keepalive-row span.timer-sm{font-family:"SFMono-Regular",Consolas,"Liberation Mono",Menlo,Courier,monospace;color:#e6edf3;font-weight:600}
   .footer{margin-top:20px;font-size:12px;color:#484f58}
</style>
</head>
<body>
<div class="card">
  <div><span class="status-dot dot-green"></span></div>
  <h1>`+appName+`</h1>
  <p class="subtitle" id="serverTime">`+time.Now().UTC().Format("Jan 2, 2006 15:04 UTC")+`</p>

   <div class="ussd-section">
     <h2>📱 USSD Gateway</h2>
     <div id="ussdTimer" class="timer timer-amber">--:--</div>
     <div id="ussdReady" style="display:none">
       <p class="ussd-ready-text">USSD is ready! Dial now:</p>
       <table class="codes-table">
         <thead>
           <tr><th>Name</th><th>Short Code</th><th>Purpose</th></tr>
         </thead>
         <tbody>
           <tr><td>Market Pay Public</td><td class="code">*715*4563143#</td><td class="purpose">Vendor Registration</td></tr>
           <tr><td>Market Pay</td><td class="code">*715*965#</td><td class="purpose">Lending Services</td></tr>
         </tbody>
       </table>
     </div>
   </div>

  <div class="component">
    <span class="component-name">API Server</span>
    <span class="component-status"><span class="status-dot dot-green"></span>Operational</span>
  </div>
  <div class="component">
    <span class="component-name">USSD Gateway</span>
    <span class="component-status"><span class="status-dot dot-green"></span>Operational</span>
  </div>
  <div class="component">
    <span class="component-name">Monime Webhooks</span>
    <span class="component-status"><span class="status-dot dot-green"></span>Operational</span>
  </div>
  <div class="component">
    <span class="component-name">PostgreSQL</span>
    <span class="component-status"><span class="status-dot dot-green"></span>Operational</span>
  </div>

  <div class="keepalive-row">
    <span>Next keep-alive ping</span>
    <span class="timer-sm" id="pingTimer">--:--</span>
  </div>

  <div class="footer">`+appName+` v`+appVersion+` · `+appEnv+`</div>
</div>

<script>
  const START = `+fmt.Sprintf("%d", startUnix)+`;
  const COOLDOWN_MS = 5 * 60 * 1000;
  const PING_INTERVAL = 5 * 60 * 1000;

  function pad(n){return n.toString().padStart(2,'0')}
  function fmt(ms){
    if(ms <= 0) return "0:00";
    const m = Math.floor(ms / 60000);
    const s = Math.floor((ms % 60000) / 1000);
    return m + ":" + pad(s);
  }

  function tick(){
    const now = Date.now();
    const elapsed = Math.max(0, now - START);
    const cooldown = Math.max(0, COOLDOWN_MS - elapsed);

    // USSD timer
    if(cooldown > 0){
      document.getElementById('ussdTimer').textContent = fmt(cooldown);
      document.getElementById('ussdTimer').className = 'timer timer-amber';
      document.getElementById('ussdReady').style.display = 'none';
      document.getElementById('ussdTimer').style.display = 'block';
    } else {
      document.getElementById('ussdTimer').style.display = 'none';
      document.getElementById('ussdReady').style.display = 'block';
    }

    // Ping timer
    const cycles = Math.floor(elapsed / PING_INTERVAL);
    const nextPing = (cycles + 1) * PING_INTERVAL;
    const pingRemaining = Math.max(0, nextPing - elapsed);
    document.getElementById('pingTimer').textContent = fmt(pingRemaining);

    // Server time
    const d = new Date(now);
    document.getElementById('serverTime').textContent =
      d.toLocaleDateString('en-US',{month:'short',day:'numeric',year:'numeric'}) + ' ' +
      pad(d.getUTCHours()) + ':' + pad(d.getUTCMinutes()) + ' UTC';
  }

  tick();
  setInterval(tick, 1000);
</script>
</body>
</html>`)
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

	// ── Keep-alive: prevent Render free-tier sleep ─────────────────────────
	keepAliveCtx, keepAliveStop := context.WithCancel(context.Background())
	defer keepAliveStop()
	if cfg.App.PublicURL != "" {
		go func() {
			client := &http.Client{Timeout: 10 * time.Second}
			ticker := time.NewTicker(5 * time.Minute)
			defer ticker.Stop()
			log.Info("keep-alive started", zap.String("url", cfg.App.PublicURL))
			for {
				select {
				case <-ticker.C:
					resp, err := client.Get(cfg.App.PublicURL)
					if err != nil {
						log.Warn("keep-alive ping failed", zap.Error(err))
						continue
					}
					resp.Body.Close()
					log.Debug("keep-alive ping ok", zap.Int("status", resp.StatusCode))
				case <-keepAliveCtx.Done():
					log.Info("keep-alive stopped")
					return
				}
			}
		}()
	} else {
		log.Warn("keep-alive disabled — public_url not set")
	}

	// ── USSD ready notification: log after 5-min cooldown ─────────────────
	if cfg.App.PublicURL != "" {
		go func() {
			select {
			case <-time.After(5 * time.Minute):
				log.Info("USSD ready — you can start dialing *715*1913660# now")
			case <-keepAliveCtx.Done():
			}
		}()
	}

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
	keepAliveStop()
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

// vendorPhoneFinderAdapter adapts VendorService to loanapp.VendorPhoneFinder.
type vendorPhoneFinderAdapter struct {
	vendorSvc *vendorapp.VendorService
}

func (a *vendorPhoneFinderAdapter) FindPhoneByID(ctx context.Context, vendorID uuid.UUID) (string, error) {
	v, err := a.vendorSvc.GetByID(ctx, vendorID)
	if err != nil {
		return "", err
	}
	return v.Phone, nil
}

// monimePayoutAdapter adapts monimepayout.Client to loanapp.MonimePayoutDisburser.
type monimePayoutAdapter struct {
	client *monimepayout.Client
	cfg    config.MonimePayoutConfig
}

func (a *monimePayoutAdapter) Disburse(ctx context.Context, phone string, amount float64) (string, error) {
	providerID := a.cfg.ProviderForPhone(phone)
	req := monimepayout.PayoutRequest{
		Amount:      monimepayout.SLEAmount(amount),
		Destination: monimepayout.MomoDestination(phone, providerID),
		Source: &monimepayout.Source{
			FinancialAccountID: a.cfg.FinancialAccountID,
		},
		Metadata: map[string]string{
			"purpose": "loan_disbursement",
		},
	}
	resp, err := a.client.CreatePayout(ctx, req)
	if err != nil {
		return "", err
	}
	return resp.Result.ID, nil
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
