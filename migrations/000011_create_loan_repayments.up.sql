CREATE TABLE loan_repayments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    loan_id UUID NOT NULL REFERENCES loans(id),
    vendor_id UUID NOT NULL REFERENCES vendors(id),
    amount DECIMAL(15,2) NOT NULL,
    monime_ref VARCHAR(255) NOT NULL,
    payment_ref VARCHAR(255) NOT NULL UNIQUE,
    metadata JSONB DEFAULT '{}',
    status VARCHAR(50) NOT NULL DEFAULT 'PENDING',
    paid_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_loan_repayments_monime_ref ON loan_repayments(monime_ref);
CREATE INDEX idx_loan_repayments_loan_id ON loan_repayments(loan_id);
CREATE INDEX idx_loan_repayments_payment_ref ON loan_repayments(payment_ref);
