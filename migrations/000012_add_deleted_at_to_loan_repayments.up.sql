ALTER TABLE loan_repayments ADD COLUMN deleted_at TIMESTAMPTZ;
CREATE INDEX idx_loan_repayments_deleted_at ON loan_repayments(deleted_at);
