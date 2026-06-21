-- 000007: Seed Loan Officer and Vendors
-- Creates Joshua Yoki as LOAN_OFFICER and links Araia, Nelson, Moses Moore, Bernard as vendors under him.

-- Joshua Yoki (LOAN_OFFICER)
INSERT INTO users (id, email, phone, password_hash, role, is_active, is_verified, is_demo, display_name)
VALUES (
    'b8f1077b-c186-40d1-8889-e3c10cad7fa8',
    'joshua@marketpay.sl',
    '+23273608360',
    crypt('password123', gen_salt('bf')),
    'LOAN_OFFICER',
    TRUE, TRUE, FALSE,
    'Joshua Yoki'
)
ON CONFLICT (email) DO UPDATE SET
    role = 'LOAN_OFFICER',
    display_name = 'Joshua Yoki',
    phone = '+23273608360',
    is_active = TRUE,
    is_verified = TRUE;

-- Abraham Araia (PIN: 2006)
INSERT INTO users (id, email, phone, password_hash, role, is_active, is_verified, is_demo, display_name)
VALUES ('aec6a745-4b99-485a-b71a-5577551198d7', 'araia+23276902971@ussd.marketpay.sl', '+23276902971', crypt('password123', gen_salt('bf')), 'VENDOR', TRUE, FALSE, FALSE, 'Abraham Araia')
ON CONFLICT (id) DO NOTHING;

INSERT INTO vendors (user_id, first_name, last_name, phone, national_id_number, national_id_type, date_of_birth, market_association_id, business_type, kyc_status, status, pin_hash, vendor_code, is_demo, field_agent_id)
VALUES (
    'aec6a745-4b99-485a-b71a-5577551198d7',
    'Abraham', 'Araia',
    '+23276902971',
    'NID-ARAIA-001',
    'NATIONAL_ID',
    '1990-01-01',
    (SELECT id FROM market_associations ORDER BY name LIMIT 1),
    'TRADER',
    'PENDING', 'PENDING',
    crypt('2006', gen_salt('bf')),
    'MP' || LPAD(CAST(floor(random() * 100000)::int AS text), 5, '0'),
    FALSE,
    'b8f1077b-c186-40d1-8889-e3c10cad7fa8'
)
ON CONFLICT (phone) DO NOTHING;

-- Emmanuel Nelson (PIN: 2005)
INSERT INTO users (id, email, phone, password_hash, role, is_active, is_verified, is_demo, display_name)
VALUES ('648774d1-3b5b-4326-a700-c14245736db1', 'nelson+23279826564@ussd.marketpay.sl', '+23279826564', crypt('password123', gen_salt('bf')), 'VENDOR', TRUE, FALSE, FALSE, 'Emmanuel Nelson')
ON CONFLICT (id) DO NOTHING;

INSERT INTO vendors (user_id, first_name, last_name, phone, national_id_number, national_id_type, date_of_birth, market_association_id, business_type, kyc_status, status, pin_hash, vendor_code, is_demo, field_agent_id)
VALUES (
    '648774d1-3b5b-4326-a700-c14245736db1',
    'Emmanuel', 'Nelson',
    '+23279826564',
    'NID-NELSON-001',
    'NATIONAL_ID',
    '1990-01-01',
    (SELECT id FROM market_associations ORDER BY name LIMIT 1),
    'TRADER',
    'PENDING', 'PENDING',
    crypt('2005', gen_salt('bf')),
    'MP' || LPAD(CAST(floor(random() * 100000)::int AS text), 5, '0'),
    FALSE,
    'b8f1077b-c186-40d1-8889-e3c10cad7fa8'
)
ON CONFLICT (phone) DO NOTHING;

-- Moses Moore (PIN: 2004)
INSERT INTO users (id, email, phone, password_hash, role, is_active, is_verified, is_demo, display_name)
VALUES ('78de1c58-e7af-4e0c-a1d7-26af075c8d8b', 'moore+23233346989@ussd.marketpay.sl', '+23233346989', crypt('password123', gen_salt('bf')), 'VENDOR', TRUE, FALSE, FALSE, 'Moses Moore')
ON CONFLICT (id) DO NOTHING;

INSERT INTO vendors (user_id, first_name, last_name, phone, national_id_number, national_id_type, date_of_birth, market_association_id, business_type, kyc_status, status, pin_hash, vendor_code, is_demo, field_agent_id)
VALUES (
    '78de1c58-e7af-4e0c-a1d7-26af075c8d8b',
    'Moses', 'Moore',
    '+23233346989',
    'NID-MOORE-001',
    'NATIONAL_ID',
    '1990-01-01',
    (SELECT id FROM market_associations ORDER BY name LIMIT 1),
    'TRADER',
    'PENDING', 'PENDING',
    crypt('2004', gen_salt('bf')),
    'MP' || LPAD(CAST(floor(random() * 100000)::int AS text), 5, '0'),
    FALSE,
    'b8f1077b-c186-40d1-8889-e3c10cad7fa8'
)
ON CONFLICT (phone) DO NOTHING;

-- Bernard Gamanga (PIN: 2005)
INSERT INTO users (id, email, phone, password_hash, role, is_active, is_verified, is_demo, display_name)
VALUES ('3d7966a4-cb35-4023-b89b-c7b69abcd8fc', 'gamanga+23272571067@ussd.marketpay.sl', '+23272571067', crypt('password123', gen_salt('bf')), 'VENDOR', TRUE, FALSE, FALSE, 'Bernard Gamanga')
ON CONFLICT (id) DO NOTHING;

INSERT INTO vendors (user_id, first_name, last_name, phone, national_id_number, national_id_type, date_of_birth, market_association_id, business_type, kyc_status, status, pin_hash, vendor_code, is_demo, field_agent_id)
VALUES (
    '3d7966a4-cb35-4023-b89b-c7b69abcd8fc',
    'Bernard', 'Gamanga',
    '+23272571067',
    'NID-GAMANGA-001',
    'NATIONAL_ID',
    '1990-01-01',
    (SELECT id FROM market_associations ORDER BY name LIMIT 1),
    'TRADER',
    'PENDING', 'PENDING',
    crypt('2005', gen_salt('bf')),
    'MP' || LPAD(CAST(floor(random() * 100000)::int AS text), 5, '0'),
    FALSE,
    'b8f1077b-c186-40d1-8889-e3c10cad7fa8'
)
ON CONFLICT (phone) DO NOTHING;

-- Link Monime subscriber UUIDs to vendor accounts
INSERT INTO ussd_subscribers (subscriber_id, vendor_id, masked_msisdn)
SELECT 'c6636869-eed2-56e4-8900-9f12beca1aec', id, '+23276902971' FROM vendors WHERE phone = '+23276902971'
UNION ALL
SELECT '3d051666-0faf-57ee-acd7-3850e5f46501', id, '+23279826564' FROM vendors WHERE phone = '+23279826564'
UNION ALL
SELECT '82769d25-da42-5389-b795-35fb7d9c4410', id, '+23233346989' FROM vendors WHERE phone = '+23233346989'
UNION ALL
SELECT 'c2366079-4414-59f8-b662-7537af83ac6f', id, '+23272571067' FROM vendors WHERE phone = '+23272571067'
ON CONFLICT (subscriber_id) DO UPDATE SET vendor_id = EXCLUDED.vendor_id, masked_msisdn = EXCLUDED.masked_msisdn, updated_at = NOW();
