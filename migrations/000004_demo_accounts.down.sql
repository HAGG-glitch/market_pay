DELETE FROM users WHERE email IN (
  'admin.demo@marketpay.sl',
  'mfi.demo@marketpay.sl',
  'vendor.demo@marketpay.sl',
  'customer.demo@marketpay.sl'
);

UPDATE users SET is_demo = false, display_name = NULL
WHERE email IN ('superadmin@marketpay.sl', 'officer@marketpay.sl', 'agent@marketpay.sl');

UPDATE partners SET is_demo = false
WHERE name = 'Sierra Leone Microfinance Trust';
