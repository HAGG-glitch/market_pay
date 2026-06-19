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
('a81aa760bee43a675750076012f74d4e43248898b077b934547d3f6dceefdc68', 'Nelson', TRUE),
('82a05af957af142cd35af548d76596ba0cf21730023e032adaef2c7c279d7270', 'Moses Moore', TRUE),
('b7e8504f4419ded90df8d15d2846814894b1fdfb0ad84f4974709ee0d94c3e87', 'Mr Cyril', TRUE),
('bdb7b76c58a77bd2dc791a01e19651276e820b4e0ad8f9568482dde3b0e0579d', 'Sanah', TRUE),
('f6a5d604c84acfb1a0904456d7b753407c4b7aad81f2bd18e5f2df6a885d5b95', 'Joshua Yoki', TRUE),
('82afa84053cecdef23cebb5b4ef37bf710f3c5d648cb6022bdb6bb5b774bce0a', 'alphakan', TRUE)
ON CONFLICT (subscriber_id_hash)
DO UPDATE SET
    label = EXCLUDED.label,
    is_active = EXCLUDED.is_active,
    updated_at = NOW();
