-- 000008: Make all vendors under Joshua eligible for loans

UPDATE vendors
SET status = 'ACTIVE',
    kyc_status = 'VERIFIED',
    first_transaction_at = NOW() - INTERVAL '60 days',
    credit_score = 80,
    updated_at = NOW()
WHERE field_agent_id = 'b8f1077b-c186-40d1-8889-e3c10cad7fa8'
  AND deleted_at IS NULL;

INSERT INTO credit_scores (id, vendor_id, total_score, transaction_volume_score, transaction_consistency_score, repayment_history_score, market_association_score, kyc_completeness_score, group_bonus, is_eligible, can_auto_approve, version, created_at, updated_at)
SELECT gen_random_uuid(), id, 80, 24, 16, 15, 10, 10, 5, true, true, 1, NOW(), NOW()
FROM vendors
WHERE field_agent_id = 'b8f1077b-c186-40d1-8889-e3c10cad7fa8'
  AND deleted_at IS NULL
  AND id NOT IN (SELECT vendor_id FROM credit_scores WHERE deleted_at IS NULL);
