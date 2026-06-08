DELETE FROM partners WHERE name = 'Sierra Leone Microfinance Trust';
DELETE FROM users WHERE email IN ('superadmin@marketpay.sl','officer@marketpay.sl','agent@marketpay.sl');
DELETE FROM market_associations WHERE name IN (
  'Big Market Freetown','Kissy Market','Lumley Market','Congo Market',
  'Waterloo Market','Bo Market','Kenema Central Market','Makeni Market'
);
DELETE FROM accounts WHERE type IN (
  'LOAN_RECEIVABLE','PARTNER_LIABILITY','INTEREST_INCOME',
  'PENALTY_INCOME','COMMISSION_INCOME','TRANSACTION_FEE_INCOME','MONIME_FLOAT'
);
