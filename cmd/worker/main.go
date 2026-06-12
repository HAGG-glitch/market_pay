package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/marketpay/backend/pkg/config"
	"github.com/marketpay/backend/pkg/logger"
	"github.com/marketpay/backend/pkg/notify"
	"github.com/marketpay/backend/pkg/outbox"
	sharedmodel "github.com/marketpay/backend/internal/shared/domain/model"
)

func main() {
	cfgPath := flag.String("config", "configs/config.yaml", "path to config file")
	flag.Parse()

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load config: %v\n", err)
		os.Exit(1)
	}

	log, err := logger.New(cfg.Logging.Level, cfg.Logging.Format)
	if err != nil {
		fmt.Fprintf(os.Stderr, "init logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Sync()

	db, err := gorm.Open(postgres.Open(cfg.Database.DSN()), &gorm.Config{})
	if err != nil {
		log.Fatal("connect database", zap.Error(err))
	}

	retryDelays := []time.Duration{0, time.Hour, 24 * time.Hour}
	if len(cfg.Outbox.RetryIntervals) > 0 {
		retryDelays = cfg.Outbox.RetryIntervals
	}

	worker := outbox.NewWorker(db, log, retryDelays)

	inAppHandler := func(eventType string) outbox.EventHandler {
		return func(ctx context.Context, event *sharedmodel.OutboxEvent) error {
			isDemo := notify.IsDemoFromPayload(event.Payload)
			return notify.DispatchOutboxEvent(ctx, db, eventType, event.Payload, isDemo)
		}
	}

	for _, evt := range []string{
		"VendorCreated", "VendorRegistered", "GroupCreated",
		"LoanApplied", "LoanRequested", "LoanApproved", "LoanRejected",
		"RepaymentReceived", "AccountFrozen", "AccountUnfrozen",
		"GroupFrozen", "PaymentCompleted",
	} {
		worker.Register(evt, inAppHandler(evt))
	}

	worker.Register("LoanDisbursed", func(ctx context.Context, event *sharedmodel.OutboxEvent) error {
		log.Info("handling LoanDisbursed", zap.String("loan_id", event.AggregateID))
		return inAppHandler("LoanDisbursed")(ctx, event)
	})

	worker.Register("LoanDefaulted", func(ctx context.Context, event *sharedmodel.OutboxEvent) error {
		log.Info("handling LoanDefaulted", zap.String("loan_id", event.AggregateID))
		var data map[string]interface{}
		_ = json.Unmarshal([]byte(event.Payload), &data)
		isDemo, _ := data["is_demo"].(bool)
		return notify.DispatchOutboxEvent(ctx, db, "LoanDefaulted", event.Payload, isDemo)
	})

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		log.Info("worker shutting down...")
		cancel()
	}()

	interval := cfg.Outbox.WorkerInterval
	if interval == 0 {
		interval = 10 * time.Second
	}

	log.Info("outbox worker started", zap.Duration("interval", interval))
	worker.Run(ctx, interval)
	log.Info("outbox worker stopped")
}
