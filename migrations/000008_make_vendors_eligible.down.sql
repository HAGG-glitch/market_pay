-- 000008: Revert vendor eligibility

UPDATE vendors
SET status = 'PENDING',
    kyc_status = 'PENDING',
    first_transaction_at = NULL,
    credit_score = 0,
    updated_at = NOW()
WHERE field_agent_id = 'b8f1077b-c186-40d1-8889-e3c10cad7fa8'
  AND deleted_at IS NULL;
