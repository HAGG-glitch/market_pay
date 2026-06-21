CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS ussd_allowed_subscribers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    subscriber_id_hash TEXT NOT NULL UNIQUE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    label TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ussd_allowed_subscribers_hash_active
ON ussd_allowed_subscribers (subscriber_id_hash, is_active);

INSERT INTO ussd_allowed_subscribers (subscriber_id_hash, label, is_active)
VALUES
('3d051666-0faf-57ee-acd7-3850e5f46501', 'Nelson', TRUE),
('82769d25-da42-5389-b795-35fb7d9c4410', 'Moses Moore', TRUE),
('b7e8504f4419ded90df8d15d2846814894b1fdfb0ad84f4974709ee0d94c3e87', 'Mr Cyril', TRUE),
('bdb7b76c58a77bd2dc791a01e19651276e820b4e0ad8f9568482dde3b0e0579d', 'Sanah', TRUE),
('f6a5d604c84acfb1a0904456d7b753407c4b7aad81f2bd18e5f2df6a885d5b95', 'Joshua Yoki', TRUE),
('ea30f380-bbb1-5bf0-9439-94ecfb63171e', 'alphakan', TRUE),
('c6636869-eed2-56e4-8900-9f12beca1aec', 'Araia', TRUE)
ON CONFLICT (subscriber_id_hash)
DO UPDATE SET
    label = EXCLUDED.label,
    is_active = EXCLUDED.is_active,
    updated_at = NOW();
