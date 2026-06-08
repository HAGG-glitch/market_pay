-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Users
CREATE TABLE IF NOT EXISTS users (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email         VARCHAR(255) UNIQUE NOT NULL,
    phone         VARCHAR(20) UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    role          VARCHAR(50) NOT NULL,
    is_active     BOOLEAN DEFAULT true,
    is_verified   BOOLEAN DEFAULT false,
    last_login_at TIMESTAMPTZ,
    created_at    TIMESTAMPTZ DEFAULT NOW(),
    updated_at    TIMESTAMPTZ DEFAULT NOW(),
    deleted_at    TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_role  ON users(role);

-- Refresh Tokens
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token      VARCHAR(512) UNIQUE NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked    BOOLEAN DEFAULT false,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens(user_id);

-- Market Associations
CREATE TABLE IF NOT EXISTS market_associations (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name       VARCHAR(255) UNIQUE NOT NULL,
    location   VARCHAR(255) NOT NULL,
    district   VARCHAR(100) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- Vendors
CREATE TABLE IF NOT EXISTS vendors (
    id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id               UUID UNIQUE NOT NULL REFERENCES users(id),
    first_name            VARCHAR(100) NOT NULL,
    last_name             VARCHAR(100) NOT NULL,
    phone                 VARCHAR(20)  UNIQUE NOT NULL,
    national_id_number    VARCHAR(50)  UNIQUE NOT NULL,
    national_id_type      VARCHAR(50)  NOT NULL,
    date_of_birth         DATE         NOT NULL,
    address               TEXT,
    market_association_id UUID         NOT NULL REFERENCES market_associations(id),
    business_name         VARCHAR(255),
    business_type         VARCHAR(100),
    kyc_status            VARCHAR(50)  NOT NULL DEFAULT 'PENDING',
    kyc_verified_at       TIMESTAMPTZ,
    status                VARCHAR(50)  NOT NULL DEFAULT 'PENDING',
    pin_hash              VARCHAR(255) NOT NULL,
    transaction_count     INT          DEFAULT 0,
    first_transaction_at  TIMESTAMPTZ,
    credit_score          DECIMAL(5,2) DEFAULT 0,
    group_id              UUID,
    created_at            TIMESTAMPTZ  DEFAULT NOW(),
    updated_at            TIMESTAMPTZ  DEFAULT NOW(),
    deleted_at            TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_vendors_phone      ON vendors(phone);
CREATE INDEX IF NOT EXISTS idx_vendors_user_id    ON vendors(user_id);
CREATE INDEX IF NOT EXISTS idx_vendors_kyc_status ON vendors(kyc_status);
CREATE INDEX IF NOT EXISTS idx_vendors_status     ON vendors(status);

-- Customers
CREATE TABLE IF NOT EXISTS customers (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    VARCHAR(255) UNIQUE,
    first_name VARCHAR(100) NOT NULL,
    last_name  VARCHAR(100) NOT NULL,
    phone      VARCHAR(20)  UNIQUE NOT NULL,
    pin_hash   VARCHAR(255),
    kyc_status VARCHAR(50)  DEFAULT 'PENDING',
    is_active  BOOLEAN      DEFAULT true,
    created_at TIMESTAMPTZ  DEFAULT NOW(),
    updated_at TIMESTAMPTZ  DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_customers_phone ON customers(phone);

-- Partners
CREATE TABLE IF NOT EXISTS partners (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name             VARCHAR(255) UNIQUE NOT NULL,
    type             VARCHAR(50)  NOT NULL,
    contact_email    VARCHAR(255),
    contact_phone    VARCHAR(20),
    commission_rate  DECIMAL(5,4) NOT NULL DEFAULT 0.05,
    available_funds  DECIMAL(20,2) DEFAULT 0,
    total_disbursed  DECIMAL(20,2) DEFAULT 0,
    total_repaid     DECIMAL(20,2) DEFAULT 0,
    total_commission DECIMAL(20,2) DEFAULT 0,
    is_active        BOOLEAN DEFAULT true,
    api_key          VARCHAR(255) UNIQUE,
    created_at       TIMESTAMPTZ DEFAULT NOW(),
    updated_at       TIMESTAMPTZ DEFAULT NOW(),
    deleted_at       TIMESTAMPTZ
);

-- Groups
CREATE TABLE IF NOT EXISTS groups (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name          VARCHAR(255) UNIQUE NOT NULL,
    description   TEXT,
    status        VARCHAR(50) NOT NULL DEFAULT 'ACTIVE',
    leader_id     UUID NOT NULL REFERENCES vendors(id),
    freeze_reason TEXT,
    created_at    TIMESTAMPTZ DEFAULT NOW(),
    updated_at    TIMESTAMPTZ DEFAULT NOW(),
    deleted_at    TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_groups_status ON groups(status);

-- Group Members
CREATE TABLE IF NOT EXISTS group_members (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id   UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    vendor_id  UUID NOT NULL REFERENCES vendors(id),
    is_leader  BOOLEAN DEFAULT false,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    UNIQUE(group_id, vendor_id)
);
CREATE INDEX IF NOT EXISTS idx_group_members_group_id  ON group_members(group_id);
CREATE INDEX IF NOT EXISTS idx_group_members_vendor_id ON group_members(vendor_id);

-- Loans
CREATE TABLE IF NOT EXISTS loans (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    vendor_id            UUID NOT NULL REFERENCES vendors(id),
    group_id             UUID REFERENCES groups(id),
    loan_type            VARCHAR(50)  NOT NULL,
    state                VARCHAR(50)  NOT NULL DEFAULT 'DRAFT',
    principal_amount     DECIMAL(15,2) NOT NULL,
    interest_rate        DECIMAL(5,4)  NOT NULL,
    interest_type        VARCHAR(30)   NOT NULL,
    total_amount         DECIMAL(15,2) NOT NULL,
    outstanding_amount   DECIMAL(15,2),
    term_weeks           INT NOT NULL,
    frequency            VARCHAR(20) NOT NULL,
    disbursed_at         TIMESTAMPTZ,
    due_date             TIMESTAMPTZ,
    credit_score_at_time DECIMAL(5,2),
    funded_by            VARCHAR(50),
    partner_id           UUID REFERENCES partners(id),
    commission_rate      DECIMAL(5,4) DEFAULT 0,
    commission_paid      BOOLEAN DEFAULT false,
    reviewed_by          UUID REFERENCES users(id),
    review_note          TEXT,
    rejection_reason     TEXT,
    monime_reference     VARCHAR(255),
    currency             VARCHAR(10) NOT NULL DEFAULT 'SLE',
    created_at           TIMESTAMPTZ DEFAULT NOW(),
    updated_at           TIMESTAMPTZ DEFAULT NOW(),
    deleted_at           TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_loans_vendor_id        ON loans(vendor_id);
CREATE INDEX IF NOT EXISTS idx_loans_state            ON loans(state);
CREATE INDEX IF NOT EXISTS idx_loans_partner_id       ON loans(partner_id);
CREATE INDEX IF NOT EXISTS idx_loans_monime_reference ON loans(monime_reference);
CREATE INDEX IF NOT EXISTS idx_loans_group_id         ON loans(group_id);

-- Repayment Schedules
CREATE TABLE IF NOT EXISTS repayment_schedules (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    loan_id         UUID NOT NULL REFERENCES loans(id) ON DELETE CASCADE,
    installment_no  INT NOT NULL,
    due_date        TIMESTAMPTZ NOT NULL,
    principal_due   DECIMAL(15,2) NOT NULL,
    interest_due    DECIMAL(15,2) NOT NULL,
    total_due       DECIMAL(15,2) NOT NULL,
    amount_paid     DECIMAL(15,2) DEFAULT 0,
    penalty_amount  DECIMAL(15,2) DEFAULT 0,
    status          VARCHAR(50) NOT NULL DEFAULT 'PENDING',
    paid_at         TIMESTAMPTZ,
    is_grace_period BOOLEAN DEFAULT false,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_repayment_schedules_loan_id  ON repayment_schedules(loan_id);
CREATE INDEX IF NOT EXISTS idx_repayment_schedules_due_date ON repayment_schedules(due_date);
CREATE INDEX IF NOT EXISTS idx_repayment_schedules_status   ON repayment_schedules(status);

-- Payments
CREATE TABLE IF NOT EXISTS payments (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id      UUID NOT NULL REFERENCES customers(id),
    vendor_id        UUID NOT NULL REFERENCES vendors(id),
    amount           DECIMAL(15,2) NOT NULL,
    fee              DECIMAL(15,2) NOT NULL,
    net_amount       DECIMAL(15,2) NOT NULL,
    status           VARCHAR(50) NOT NULL DEFAULT 'PENDING',
    monime_reference VARCHAR(255),
    description      TEXT,
    currency         VARCHAR(10) NOT NULL DEFAULT 'SLE',
    created_at       TIMESTAMPTZ DEFAULT NOW(),
    updated_at       TIMESTAMPTZ DEFAULT NOW(),
    deleted_at       TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_payments_customer_id ON payments(customer_id);
CREATE INDEX IF NOT EXISTS idx_payments_vendor_id   ON payments(vendor_id);
CREATE INDEX IF NOT EXISTS idx_payments_status      ON payments(status);

-- Credit Scores
CREATE TABLE IF NOT EXISTS credit_scores (
    id                            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    vendor_id                     UUID NOT NULL REFERENCES vendors(id),
    total_score                   DECIMAL(5,2) NOT NULL,
    transaction_volume_score      DECIMAL(5,2),
    transaction_consistency_score DECIMAL(5,2),
    repayment_history_score       DECIMAL(5,2),
    market_association_score      DECIMAL(5,2),
    kyc_completeness_score        DECIMAL(5,2),
    group_bonus                   DECIMAL(5,2),
    is_eligible                   BOOLEAN NOT NULL,
    can_auto_approve              BOOLEAN NOT NULL,
    version                       INT DEFAULT 1,
    created_at                    TIMESTAMPTZ DEFAULT NOW(),
    updated_at                    TIMESTAMPTZ DEFAULT NOW(),
    deleted_at                    TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_credit_scores_vendor_id ON credit_scores(vendor_id);

-- Ledger Accounts
CREATE TABLE IF NOT EXISTS accounts (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type        VARCHAR(50) UNIQUE NOT NULL,
    name        VARCHAR(255) NOT NULL,
    balance     DECIMAL(20,2) DEFAULT 0,
    currency    VARCHAR(10)   NOT NULL DEFAULT 'SLE',
    description TEXT,
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    updated_at  TIMESTAMPTZ DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);

-- Journal Entries
CREATE TABLE IF NOT EXISTS journal_entries (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    reference   VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    entry_date  TIMESTAMPTZ NOT NULL,
    posted_by   UUID REFERENCES users(id),
    is_posted   BOOLEAN DEFAULT false,
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    updated_at  TIMESTAMPTZ DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_journal_entries_reference  ON journal_entries(reference);
CREATE INDEX IF NOT EXISTS idx_journal_entries_entry_date ON journal_entries(entry_date);

-- Journal Lines
CREATE TABLE IF NOT EXISTS journal_lines (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    journal_entry_id UUID NOT NULL REFERENCES journal_entries(id) ON DELETE CASCADE,
    account_id       UUID NOT NULL REFERENCES accounts(id),
    entry_type       VARCHAR(10)   NOT NULL,
    amount           DECIMAL(15,2) NOT NULL,
    description      TEXT,
    created_at       TIMESTAMPTZ DEFAULT NOW(),
    updated_at       TIMESTAMPTZ DEFAULT NOW(),
    deleted_at       TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_journal_lines_journal_entry_id ON journal_lines(journal_entry_id);
CREATE INDEX IF NOT EXISTS idx_journal_lines_account_id       ON journal_lines(account_id);

-- USSD Sessions
CREATE TABLE IF NOT EXISTS ussd_sessions (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id   VARCHAR(255) UNIQUE NOT NULL,
    phone_number VARCHAR(20)  NOT NULL,
    user_id      UUID REFERENCES users(id),
    menu_state   VARCHAR(50)  NOT NULL,
    state_data   JSONB,
    pin_verified BOOLEAN DEFAULT false,
    last_input   VARCHAR(255),
    expires_at   TIMESTAMPTZ NOT NULL,
    is_active    BOOLEAN DEFAULT true,
    created_at   TIMESTAMPTZ DEFAULT NOW(),
    updated_at   TIMESTAMPTZ DEFAULT NOW(),
    deleted_at   TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_ussd_sessions_session_id ON ussd_sessions(session_id);
CREATE INDEX IF NOT EXISTS idx_ussd_sessions_phone      ON ussd_sessions(phone_number);

-- Monime Transactions
CREATE TABLE IF NOT EXISTS monime_transactions (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    reference        VARCHAR(255) UNIQUE NOT NULL,
    external_ref     VARCHAR(255),
    type             VARCHAR(50)   NOT NULL,
    amount           DECIMAL(15,2) NOT NULL,
    currency         VARCHAR(10)   NOT NULL DEFAULT 'SLE',
    phone            VARCHAR(20)   NOT NULL,
    status           VARCHAR(50)   NOT NULL DEFAULT 'PENDING',
    retry_count      INT DEFAULT 0,
    next_retry_at    TIMESTAMPTZ,
    last_error       TEXT,
    webhook_received BOOLEAN DEFAULT false,
    webhook_payload  JSONB,
    entity_id        VARCHAR(255),
    entity_type      VARCHAR(50),
    created_at       TIMESTAMPTZ DEFAULT NOW(),
    updated_at       TIMESTAMPTZ DEFAULT NOW(),
    deleted_at       TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_monime_transactions_reference ON monime_transactions(reference);
CREATE INDEX IF NOT EXISTS idx_monime_transactions_entity_id ON monime_transactions(entity_id);

-- Outbox Events
CREATE TABLE IF NOT EXISTS outbox_events (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_type    VARCHAR(100) NOT NULL,
    aggregate_id  VARCHAR(255) NOT NULL,
    payload       JSONB NOT NULL,
    status        VARCHAR(50) NOT NULL DEFAULT 'PENDING',
    retry_count   INT DEFAULT 0,
    next_retry_at TIMESTAMPTZ DEFAULT NOW(),
    published_at  TIMESTAMPTZ,
    error         TEXT,
    created_at    TIMESTAMPTZ DEFAULT NOW(),
    updated_at    TIMESTAMPTZ DEFAULT NOW(),
    deleted_at    TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_outbox_events_status        ON outbox_events(status, next_retry_at);
CREATE INDEX IF NOT EXISTS idx_outbox_events_event_type    ON outbox_events(event_type);

-- Notifications
CREATE TABLE IF NOT EXISTS notifications (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    recipient_id    UUID NOT NULL,
    recipient_phone VARCHAR(20),
    recipient_email VARCHAR(255),
    channel         VARCHAR(20)  NOT NULL,
    event_type      VARCHAR(100) NOT NULL,
    subject         VARCHAR(255),
    body            TEXT NOT NULL,
    status          VARCHAR(50) NOT NULL DEFAULT 'PENDING',
    sent_at         TIMESTAMPTZ,
    error           TEXT,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_notifications_recipient_id ON notifications(recipient_id);
CREATE INDEX IF NOT EXISTS idx_notifications_status       ON notifications(status);

-- Audit Logs
CREATE TABLE IF NOT EXISTS audit_logs (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    actor_id    UUID NOT NULL,
    actor_role  VARCHAR(50)  NOT NULL,
    action      VARCHAR(100) NOT NULL,
    resource    VARCHAR(100) NOT NULL,
    resource_id VARCHAR(255) NOT NULL,
    old_state   TEXT,
    new_state   TEXT,
    ip_address  VARCHAR(50),
    user_agent  TEXT,
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    updated_at  TIMESTAMPTZ DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_audit_logs_actor_id   ON audit_logs(actor_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_resource   ON audit_logs(resource, resource_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at);
