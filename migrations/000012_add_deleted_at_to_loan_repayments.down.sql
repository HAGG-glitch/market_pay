DROP INDEX IF EXISTS idx_loan_repayments_deleted_at;
ALTER TABLE loan_repayments DROP COLUMN IF EXISTS deleted_at;
