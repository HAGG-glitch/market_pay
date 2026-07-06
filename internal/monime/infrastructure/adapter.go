package infrastructure

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	monimemodel "github.com/marketpay/backend/internal/monime/domain/model"
	"github.com/marketpay/backend/pkg/config"
	"github.com/marketpay/backend/pkg/logger"
	"go.uber.org/zap"
)

// MonimeAdapter implements PaymentGateway using the Monime API.
type MonimeAdapter struct {
	cfg    config.MonimeConfig
	client *http.Client
	log    *logger.Logger
}

// NewMonimeAdapter constructs a new MonimeAdapter.
func NewMonimeAdapter(cfg config.MonimeConfig, log *logger.Logger) *MonimeAdapter {
	return &MonimeAdapter{
		cfg: cfg,
		client: &http.Client{
			Timeout: cfg.Timeout,
		},
		log: log,
	}
}

func (m *MonimeAdapter) Disburse(ctx context.Context, req monimemodel.DisbursementRequest) (*monimemodel.DisbursementResponse, error) {
	body, _ := json.Marshal(req)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		m.cfg.BaseURL+"/disbursements", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build disburse request: %w", err)
	}

	m.setHeaders(httpReq)

	resp, err := m.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("disburse request: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		m.log.Error("monime disburse error",
			zap.Int("status", resp.StatusCode),
			zap.String("body", string(respBody)))
		return nil, fmt.Errorf("monime returned %d: %s", resp.StatusCode, string(respBody))
	}

	var result monimemodel.DisbursementResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("parse disburse response: %w", err)
	}

	return &result, nil
}

func (m *MonimeAdapter) Collect(ctx context.Context, req monimemodel.CollectionRequest) (*monimemodel.CollectionResponse, error) {
	body, _ := json.Marshal(req)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		m.cfg.BaseURL+"/collections", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build collect request: %w", err)
	}

	m.setHeaders(httpReq)

	resp, err := m.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("collect request: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("monime returned %d: %s", resp.StatusCode, string(respBody))
	}

	var result monimemodel.CollectionResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("parse collect response: %w", err)
	}

	return &result, nil
}

func (m *MonimeAdapter) GetWebhookSecret() string {
	return m.cfg.WebhookSecret
}

func (m *MonimeAdapter) ValidateWebhook(payload []byte, signature string) bool {
	m.log.Info("validating webhook signature",
		zap.String("signature", signature),
		zap.ByteString("body", payload),
	)

	if ts, sig, ok := parseMonimeSignature(signature); ok {
		signed := fmt.Sprintf("%d.%s", ts, string(payload))
		mac := hmac.New(sha256.New, []byte(m.cfg.WebhookSecret))
		mac.Write([]byte(signed))
		expected := hex.EncodeToString(mac.Sum(nil))
		valid := hmac.Equal([]byte(expected), []byte(sig))
		if !valid {
			m.log.Warn("timestamped signature mismatch",
				zap.Int64("ts", ts),
				zap.String("expected", expected),
				zap.String("received", sig),
			)
			return false
		}

		now := time.Now().Unix()
		diff := now - ts
		if diff < 0 {
			diff = -diff
		}
		if diff > 300 {
			m.log.Warn("webhook timestamp outside tolerance", zap.Int64("ts", ts), zap.Int64("now", now))
			return false
		}
		return true
	}

	mac := hmac.New(sha256.New, []byte(m.cfg.WebhookSecret))
	mac.Write(payload)
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signature))
}

// parseMonimeSignature parses "t=<ts>,s=<hex>" or "t=<ts>,v1=<hex>" format.
// Returns timestamp (seconds), hex signature, and whether parsing succeeded.
func parseMonimeSignature(sig string) (int64, string, bool) {
	var ts int64
	var hexVal string

	parts := strings.Split(sig, ",")
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if strings.HasPrefix(p, "t=") {
			if n, err := strconv.ParseInt(p[2:], 10, 64); err == nil {
				ts = n
			}
		} else if strings.HasPrefix(p, "s=") || strings.HasPrefix(p, "v1=") {
			hexVal = strings.TrimPrefix(p, "s=")
			hexVal = strings.TrimPrefix(hexVal, "v1=")
		}
	}

	if ts == 0 || hexVal == "" {
		return 0, "", false
	}
	return ts, hexVal, true
}

func (m *MonimeAdapter) GetTransaction(ctx context.Context, reference string) (*monimemodel.MonimeTransaction, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet,
		m.cfg.BaseURL+"/transactions/"+reference, nil)
	if err != nil {
		return nil, err
	}
	m.setHeaders(httpReq)

	resp, err := m.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var tx monimemodel.MonimeTransaction
	if err := json.NewDecoder(resp.Body).Decode(&tx); err != nil {
		return nil, err
	}
	return &tx, nil
}

func (m *MonimeAdapter) setHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+m.cfg.APIKey)
	req.Header.Set("X-Request-Time", time.Now().UTC().Format(time.RFC3339))
}
