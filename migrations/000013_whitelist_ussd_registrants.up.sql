-- 000013: Create USSD-registered vendors (Maurice, Sanah) under Joshua Yoki and whitelist them.

-- Maurice Bangura (VENDOR)
INSERT INTO users (id, email, phone, password_hash, role, is_active, is_verified, is_demo, display_name)
VALUES (
    'a1b2c3d4-e5f6-7890-abcd-ef1234567890',
    'maurice+23276889012@ussd.marketpay.sl',
    '+23276889012',
    crypt('password123', gen_salt('bf')),
    'VENDOR', TRUE, FALSE, FALSE,
    'Maurice Bangura'
)
ON CONFLICT (id) DO NOTHING;

INSERT INTO vendors (user_id, first_name, last_name, phone, national_id_number, national_id_type, date_of_birth, market_association_id, business_type, kyc_status, status, pin_hash, vendor_code, is_demo, field_agent_id)
VALUES (
    'a1b2c3d4-e5f6-7890-abcd-ef1234567890',
    'Maurice', 'Bangura',
    '+23276889012',
    'NID-BANGURA-001',
    'NATIONAL_ID',
    '1990-01-01',
    (SELECT id FROM market_associations ORDER BY name LIMIT 1),
    'TRADER',
    'PENDING', 'PENDING',
    crypt('2000', gen_salt('bf')),
    'MP00600',
    FALSE,
    'b8f1077b-c186-40d1-8889-e3c10cad7fa8'
)
ON CONFLICT (phone) DO NOTHING;

-- Link Maurice's subscriber IDs to his vendor record
INSERT INTO ussd_subscribers (subscriber_id, vendor_id, masked_msisdn)
SELECT s.subscriber_id, v.id, v.phone
FROM (VALUES ('be3a4ef5-f55e-5f78-b09c-a25dc69c5c5c'), ('e6b9c01c-827d-5305-bcdb-fe935c7e7796')) AS s(subscriber_id)
CROSS JOIN vendors v WHERE v.phone = '+23276889012'
ON CONFLICT (subscriber_id) DO UPDATE SET vendor_id = EXCLUDED.vendor_id, masked_msisdn = EXCLUDED.masked_msisdn, updated_at = NOW();

-- Sanah Marah (VENDOR)
INSERT INTO users (id, email, phone, password_hash, role, is_active, is_verified, is_demo, display_name)
VALUES (
    'b2c3d4e5-f6a7-8901-bcde-f12345678901',
    'sanah+23276889013@ussd.marketpay.sl',
    '+23276889013',
    crypt('password123', gen_salt('bf')),
    'VENDOR', TRUE, FALSE, FALSE,
    'Sanah Marah'
)
ON CONFLICT (id) DO NOTHING;

INSERT INTO vendors (user_id, first_name, last_name, phone, national_id_number, national_id_type, date_of_birth, market_association_id, business_type, kyc_status, status, pin_hash, vendor_code, is_demo, field_agent_id)
VALUES (
    'b2c3d4e5-f6a7-8901-bcde-f12345678901',
    'Sanah', 'Marah',
    '+23276889013',
    'NID-MARAH-001',
    'NATIONAL_ID',
    '1990-01-01',
    (SELECT id FROM market_associations ORDER BY name LIMIT 1),
    'TRADER',
    'PENDING', 'PENDING',
    crypt('2000', gen_salt('bf')),
    'MP00601',
    FALSE,
    'b8f1077b-c186-40d1-8889-e3c10cad7fa8'
)
ON CONFLICT (phone) DO NOTHING;

-- Link Sanah's subscriber ID to her vendor record
INSERT INTO ussd_subscribers (subscriber_id, vendor_id, masked_msisdn)
SELECT 'c08d98dc-4851-5ccb-8d13-75725ca49a17', v.id, v.phone
FROM vendors v WHERE v.phone = '+23276889013'
ON CONFLICT (subscriber_id) DO UPDATE SET vendor_id = EXCLUDED.vendor_id, masked_msisdn = EXCLUDED.masked_msisdn, updated_at = NOW();

-- Whitelist for USSD access
INSERT INTO ussd_allowed_subscribers (subscriber_id_hash, label, is_active)
VALUES
  ('be3a4ef5-f55e-5f78-b09c-a25dc69c5c5c', 'Maurice Bangura', TRUE),
  ('e6b9c01c-827d-5305-bcdb-fe935c7e7796', 'Maurice Bangura', TRUE),
  ('c08d98dc-4851-5ccb-8d13-75725ca49a17', 'Sanah Marah', TRUE)
ON CONFLICT (subscriber_id_hash)
DO UPDATE SET
    label     = EXCLUDED.label,
    is_active = EXCLUDED.is_active,
    updated_at = NOW();

-- Credit scores for loan eligibility
INSERT INTO credit_scores (vendor_id, total_score, transaction_volume_score, transaction_consistency_score, repayment_history_score, market_association_score, kyc_completeness_score, group_bonus, is_eligible, can_auto_approve, version)
SELECT v.id, 50, 10, 10, 10, 10, 5, 5, true, false, 1
FROM vendors v
WHERE v.field_agent_id = 'b8f1077b-c186-40d1-8889-e3c10cad7fa8'
  AND v.phone IN ('+23276889012', '+23276889013')
  AND NOT EXISTS (SELECT 1 FROM credit_scores cs WHERE cs.vendor_id = v.id);
