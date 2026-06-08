package service

import (
	"context"

	monimemodel "github.com/marketpay/backend/internal/monime/domain/model"
)

// PaymentGateway is the port interface for any payment provider.
// Implementations include MonimeAdapter, MockAdapter, etc.
type PaymentGateway interface {
	// Disburse sends funds to a mobile money number.
	Disburse(ctx context.Context, req monimemodel.DisbursementRequest) (*monimemodel.DisbursementResponse, error)

	// Collect initiates a collection from a mobile money number.
	Collect(ctx context.Context, req monimemodel.CollectionRequest) (*monimemodel.CollectionResponse, error)

	// ValidateWebhook validates HMAC signature on an incoming webhook.
	ValidateWebhook(payload []byte, signature string) bool

	// GetTransaction fetches a transaction status by reference.
	GetTransaction(ctx context.Context, reference string) (*monimemodel.MonimeTransaction, error)
}
