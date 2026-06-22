-- 000009: Add source column to loans table to distinguish USSD vs Web origin

ALTER TABLE loans ADD COLUMN IF NOT EXISTS source VARCHAR(20) NOT NULL DEFAULT 'WEB';

-- Backfill existing loans: those created via USSD exchange have no explicit
-- source marker yet, default to WEB is correct for all existing records.
