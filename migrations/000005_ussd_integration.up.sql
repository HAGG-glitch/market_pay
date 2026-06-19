-- USSD subscriber identity mapping
CREATE TABLE IF NOT EXISTS ussd_subscribers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    subscriber_id TEXT NOT NULL UNIQUE,
    vendor_id UUID REFERENCES vendors(id),
    masked_msisdn TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_ussd_subscribers_subscriber_id ON ussd_subscribers(subscriber_id);

-- Add response_data to exchange sessions for idempotent response caching
ALTER TABLE monime_exchange_sessions ADD COLUMN IF NOT EXISTS response_data TEXT;
