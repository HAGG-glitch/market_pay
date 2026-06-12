-- Demo/live data isolation and platform workflow extensions

ALTER TABLE users ADD COLUMN IF NOT EXISTS is_demo BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE users ADD COLUMN IF NOT EXISTS display_name VARCHAR(255);
ALTER TABLE users ADD COLUMN IF NOT EXISTS field_agent_id UUID REFERENCES users(id);

ALTER TABLE vendors ADD COLUMN IF NOT EXISTS is_demo BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE vendors ADD COLUMN IF NOT EXISTS vendor_code VARCHAR(20) UNIQUE;
ALTER TABLE vendors ADD COLUMN IF NOT EXISTS field_agent_id UUID REFERENCES users(id);
ALTER TABLE vendors ADD COLUMN IF NOT EXISTS frozen_at TIMESTAMPTZ;
ALTER TABLE vendors ADD COLUMN IF NOT EXISTS frozen_by UUID REFERENCES users(id);
ALTER TABLE vendors ADD COLUMN IF NOT EXISTS freeze_reason TEXT;

ALTER TABLE customers ADD COLUMN IF NOT EXISTS is_demo BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE partners ADD COLUMN IF NOT EXISTS is_demo BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE groups ADD COLUMN IF NOT EXISTS is_demo BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE groups ADD COLUMN IF NOT EXISTS field_agent_id UUID REFERENCES users(id);
ALTER TABLE loans ADD COLUMN IF NOT EXISTS is_demo BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE payments ADD COLUMN IF NOT EXISTS is_demo BOOLEAN NOT NULL DEFAULT false;

ALTER TABLE notifications ADD COLUMN IF NOT EXISTS is_demo BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE notifications ADD COLUMN IF NOT EXISTS is_read BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE notifications ADD COLUMN IF NOT EXISTS title VARCHAR(255);

CREATE TABLE IF NOT EXISTS freeze_history (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entity_type   VARCHAR(50) NOT NULL,
    entity_id     UUID NOT NULL,
    action        VARCHAR(20) NOT NULL,
    reason        TEXT,
    actor_id      UUID REFERENCES users(id),
    actor_role    VARCHAR(50),
    is_demo       BOOLEAN NOT NULL DEFAULT false,
    created_at    TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_freeze_history_entity ON freeze_history(entity_type, entity_id);

CREATE TABLE IF NOT EXISTS in_app_notifications (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    recipient_id  UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    event_type    VARCHAR(100) NOT NULL,
    title         VARCHAR(255) NOT NULL,
    body          TEXT NOT NULL,
    is_read       BOOLEAN NOT NULL DEFAULT false,
    is_demo       BOOLEAN NOT NULL DEFAULT false,
    metadata      JSONB,
    created_at    TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_in_app_notifications_recipient ON in_app_notifications(recipient_id, is_read);

CREATE TABLE IF NOT EXISTS monime_exchange_sessions (
    session_id    VARCHAR(255) PRIMARY KEY,
    response_hash VARCHAR(64),
    created_at    TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_users_is_demo ON users(is_demo);
CREATE INDEX IF NOT EXISTS idx_vendors_is_demo ON vendors(is_demo);
CREATE INDEX IF NOT EXISTS idx_loans_is_demo ON loans(is_demo);
