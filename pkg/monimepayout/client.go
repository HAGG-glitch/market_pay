package monimepayout

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type Config struct {
	BaseURL            string
	APIKey             string
	SpaceID            string
	FinancialAccountID string
	ProviderID         string
	Timeout            time.Duration
}

type Amount struct {
	Currency string `json:"currency"`
	Value    int    `json:"value"`
}

type Destination struct {
	Type        string `json:"type"`
	ProviderID  string `json:"providerId"`
	PhoneNumber string `json:"phoneNumber"`
}

type Source struct {
	FinancialAccountID string `json:"financialAccountId"`
}

type PayoutRequest struct {
	Amount      Amount            `json:"amount"`
	Destination Destination       `json:"destination"`
	Source      *Source           `json:"source,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

type PayoutResponse struct {
	Success  bool         `json:"success"`
	Messages []string     `json:"messages"`
	Result   PayoutResult `json:"result"`
}

type PayoutResult struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

type Client struct {
	cfg    Config
	client *http.Client
}

func NewClient(cfg Config) *Client {
	return &Client{
		cfg: cfg,
		client: &http.Client{
			Timeout: cfg.Timeout,
		},
	}
}

func (c *Client) CreatePayout(ctx context.Context, req PayoutRequest) (*PayoutResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal payout request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.BaseURL+"/v1/payouts", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build payout request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.cfg.APIKey)
	httpReq.Header.Set("Monime-Space-Id", c.cfg.SpaceID)
	httpReq.Header.Set("Idempotency-Key", uuid.New().String())

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("payout request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read payout response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("monime returned %d: %s", resp.StatusCode, string(respBody))
	}

	var payoutResp PayoutResponse
	if err := json.Unmarshal(respBody, &payoutResp); err != nil {
		return nil, fmt.Errorf("parse payout response: %w", err)
	}

	return &payoutResp, nil
}

func SLEAmount(amount float64) Amount {
	return Amount{
		Currency: "SLE",
		Value:    int(amount * 100),
	}
}

func MomoDestination(phoneNumber, providerID string) Destination {
	return Destination{
		Type:        "momo",
		ProviderID:  providerID,
		PhoneNumber: phoneNumber,
	}
}
