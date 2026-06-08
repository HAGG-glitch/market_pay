-- Seed ledger accounts
INSERT INTO accounts (id, type, name, currency, description) VALUES
  (gen_random_uuid(), 'LOAN_RECEIVABLE',    'Loan Receivable',       'SLE', 'Outstanding loan principal owed to MarketPay'),
  (gen_random_uuid(), 'PARTNER_LIABILITY',  'Partner Liability',     'SLE', 'Funds received from MFI/NGO partners pending deployment'),
  (gen_random_uuid(), 'INTEREST_INCOME',    'Interest Income',       'SLE', 'Interest earned on loans'),
  (gen_random_uuid(), 'PENALTY_INCOME',     'Penalty Income',        'SLE', 'Late payment penalties collected'),
  (gen_random_uuid(), 'COMMISSION_INCOME',  'Commission Income',     'SLE', 'Commission earned from partners'),
  (gen_random_uuid(), 'TRANSACTION_FEE_INCOME', 'Transaction Fee Income', 'SLE', '1% fee on vendor payments'),
  (gen_random_uuid(), 'MONIME_FLOAT',       'Monime Float',          'SLE', 'Cash held in Monime mobile money float')
ON CONFLICT (type) DO NOTHING;

-- Seed market associations (Freetown markets)
INSERT INTO market_associations (id, name, location, district) VALUES
  (gen_random_uuid(), 'Big Market Freetown',       'Wallace Johnson Street, Freetown', 'Western Area Urban'),
  (gen_random_uuid(), 'Kissy Market',               'Kissy Road, Freetown',             'Western Area Urban'),
  (gen_random_uuid(), 'Lumley Market',              'Lumley Beach Road, Freetown',      'Western Area Urban'),
  (gen_random_uuid(), 'Congo Market',               'Congo Cross, Freetown',            'Western Area Urban'),
  (gen_random_uuid(), 'Waterloo Market',            'Waterloo, Freetown',               'Western Area Rural'),
  (gen_random_uuid(), 'Bo Market',                  'Bo Town Centre',                   'Bo'),
  (gen_random_uuid(), 'Kenema Central Market',      'Kenema',                           'Kenema'),
  (gen_random_uuid(), 'Makeni Market',              'Makeni',                           'Bombali')
ON CONFLICT (name) DO NOTHING;

-- Seed super admin user (password: Admin@1234)
-- bcrypt hash of "Admin@1234"
INSERT INTO users (id, email, phone, password_hash, role, is_active, is_verified) VALUES
  (gen_random_uuid(),
   'superadmin@marketpay.sl',
   '+23276000001',
   '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi',
   'SUPER_ADMIN',
   true,
   true)
ON CONFLICT (email) DO NOTHING;

-- Seed demo MFI partner
INSERT INTO partners (id, name, type, contact_email, contact_phone, commission_rate, available_funds, is_active, api_key) VALUES
  (gen_random_uuid(),
   'Sierra Leone Microfinance Trust',
   'MFI_PARTNER',
   'partner@slmt.sl',
   '+23276000002',
   0.05,
   5000000.00,
   true,
   'slmt-api-key-demo-2024')
ON CONFLICT (name) DO NOTHING;

-- Seed demo loan officer
INSERT INTO users (id, email, phone, password_hash, role, is_active, is_verified) VALUES
  (gen_random_uuid(),
   'officer@marketpay.sl',
   '+23276000003',
   '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi',
   'LOAN_OFFICER',
   true,
   true)
ON CONFLICT (email) DO NOTHING;

-- Seed demo field agent
INSERT INTO users (id, email, phone, password_hash, role, is_active, is_verified) VALUES
  (gen_random_uuid(),
   'agent@marketpay.sl',
   '+23276000004',
   '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi',
   'FIELD_AGENT',
   true,
   true)
ON CONFLICT (email) DO NOTHING;
