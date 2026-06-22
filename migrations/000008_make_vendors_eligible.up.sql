-- 000008: Make all vendors under Joshua eligible for loans

UPDATE vendors
SET status = 'ACTIVE',
    kyc_status = 'VERIFIED',
    first_transaction_at = NOW() - INTERVAL '60 days',
    credit_score = 80,
    updated_at = NOW()
WHERE field_agent_id = 'b8f1077b-c186-40d1-8889-e3c10cad7fa8'
  AND deleted_at IS NULL;
