package outbox

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	shared "github.com/marketpay/backend/internal/shared/domain/model"
	"github.com/marketpay/backend/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Publisher saves domain events to the outbox table atomically.
type Publisher struct {
	db  *gorm.DB
	log *logger.Logger
}

// NewPublisher constructs a Publisher.
func NewPublisher(db *gorm.DB, log *logger.Logger) *Publisher {
	return &Publisher{db: db, log: log}
}

// Publish persists an event to the outbox table within the current transaction.
func (p *Publisher) Publish(ctx context.Context, eventType, aggregateID string, payload interface{}) error {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		p.log.Error("failed to marshal event payload",
			zap.String("event_type", eventType),
			zap.Error(err))
		return err
	}

	event := &shared.OutboxEvent{
		BaseModel: shared.BaseModel{
			ID: uuid.New(),
		},
		EventType:   eventType,
		AggregateID: aggregateID,
		Payload:     string(payloadBytes),
		Status:      "PENDING",
		NextRetryAt: time.Now(),
	}

	return p.db.WithContext(ctx).Create(event).Error
}

// Worker processes pending outbox events and publishes them.
type Worker struct {
	db          *gorm.DB
	log         *logger.Logger
	retryDelays []time.Duration
	handlers    map[string]EventHandler
}

// EventHandler is a function that handles a specific event type.
type EventHandler func(ctx context.Context, event *shared.OutboxEvent) error

// NewWorker constructs an outbox Worker.
func NewWorker(db *gorm.DB, log *logger.Logger, retryDelays []time.Duration) *Worker {
	return &Worker{
		db:          db,
		log:         log,
		retryDelays: retryDelays,
		handlers:    make(map[string]EventHandler),
	}
}

// Register registers a handler for a specific event type.
func (w *Worker) Register(eventType string, handler EventHandler) {
	w.handlers[eventType] = handler
}

// Run starts the outbox worker loop.
func (w *Worker) Run(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.log.Info("outbox worker stopping")
			return
		case <-ticker.C:
			w.processEvents(ctx)
		}
	}
}

func (w *Worker) processEvents(ctx context.Context) {
	var events []shared.OutboxEvent
	err := w.db.WithContext(ctx).
		Where("status = ? AND next_retry_at <= ?", "PENDING", time.Now()).
		Order("created_at ASC").
		Limit(100).
		Find(&events).Error

	if err != nil {
		w.log.Error("failed to fetch outbox events", zap.Error(err))
		return
	}

	for _, event := range events {
		ev := event
		w.handleEvent(ctx, &ev)
	}
}

func (w *Worker) handleEvent(ctx context.Context, event *shared.OutboxEvent) {
	handler, ok := w.handlers[event.EventType]
	if !ok {
		// No handler registered — mark as manual review after max retries
		w.log.Warn("no handler for event type", zap.String("event_type", event.EventType))
		w.markManualReview(ctx, event, "no handler registered")
		return
	}

	if err := handler(ctx, event); err != nil {
		w.log.Error("event handler failed",
			zap.String("event_id", event.ID.String()),
			zap.String("event_type", event.EventType),
			zap.Error(err))

		event.RetryCount++
		event.Error = err.Error()

		maxRetries := len(w.retryDelays)
		if event.RetryCount >= maxRetries {
			w.markManualReview(ctx, event, err.Error())
			return
		}

		nextDelay := w.retryDelays[event.RetryCount]
		nextRetry := time.Now().Add(nextDelay)
		event.NextRetryAt = nextRetry

		w.db.WithContext(ctx).Save(event)
		return
	}

	// Success
	now := time.Now()
	event.Status = "PUBLISHED"
	event.PublishedAt = &now
	w.db.WithContext(ctx).Save(event)
}

func (w *Worker) markManualReview(ctx context.Context, event *shared.OutboxEvent, reason string) {
	event.Status = "MANUAL_REVIEW"
	event.Error = reason
	w.db.WithContext(ctx).Save(event)
	w.log.Error("event moved to manual review",
		zap.String("event_id", event.ID.String()),
		zap.String("event_type", event.EventType),
		zap.String("reason", reason))
}
