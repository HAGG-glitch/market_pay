package main

import (
	"context"
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

	// Register event handlers
	worker.Register("VendorRegistered", func(ctx context.Context, event *sharedmodel.OutboxEvent) error {
		log.Info("handling VendorRegistered event",
			zap.String("event_id", event.ID.String()),
			zap.String("aggregate_id", event.AggregateID),
		)
		// TODO: send SMS/WhatsApp notification via notification service
		return nil
	})

	worker.Register("LoanApproved", func(ctx context.Context, event *sharedmodel.OutboxEvent) error {
		log.Info("handling LoanApproved event",
			zap.String("loan_id", event.AggregateID),
		)
		// TODO: send loan approval notification
		return nil
	})

	worker.Register("LoanRejected", func(ctx context.Context, event *sharedmodel.OutboxEvent) error {
		log.Info("handling LoanRejected event",
			zap.String("loan_id", event.AggregateID),
		)
		return nil
	})

	worker.Register("LoanDisbursed", func(ctx context.Context, event *sharedmodel.OutboxEvent) error {
		log.Info("handling LoanDisbursed event",
			zap.String("loan_id", event.AggregateID),
		)
		// TODO: post journal entries via ledger service
		return nil
	})

	worker.Register("RepaymentReceived", func(ctx context.Context, event *sharedmodel.OutboxEvent) error {
		log.Info("handling RepaymentReceived",
			zap.String("loan_id", event.AggregateID),
		)
		return nil
	})

	worker.Register("LoanDefaulted", func(ctx context.Context, event *sharedmodel.OutboxEvent) error {
		log.Info("handling LoanDefaulted",
			zap.String("loan_id", event.AggregateID),
		)
		return nil
	})

	worker.Register("GroupFrozen", func(ctx context.Context, event *sharedmodel.OutboxEvent) error {
		log.Info("handling GroupFrozen",
			zap.String("group_id", event.AggregateID),
		)
		return nil
	})

	worker.Register("PaymentCompleted", func(ctx context.Context, event *sharedmodel.OutboxEvent) error {
		log.Info("handling PaymentCompleted",
			zap.String("payment_id", event.AggregateID),
		)
		return nil
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
