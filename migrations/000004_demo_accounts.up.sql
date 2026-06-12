-- Mark existing seed accounts as demo and add full role demo set

UPDATE users SET is_demo = true, display_name = 'Super Admin Demo'
WHERE email = 'superadmin@marketpay.sl';

UPDATE users SET is_demo = true, display_name = 'Loan Officer Demo'
WHERE email = 'officer@marketpay.sl';

UPDATE users SET is_demo = true, display_name = 'Field Agent Demo'
WHERE email = 'agent@marketpay.sl';

UPDATE partners SET is_demo = true
WHERE name = 'Sierra Leone Microfinance Trust';

INSERT INTO users (email, phone, password_hash, role, is_active, is_verified, is_demo, display_name) VALUES
  ('admin.demo@marketpay.sl', '+23276000010',
   '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi',
   'ADMIN', true, true, true, 'Admin Demo'),
  ('mfi.demo@marketpay.sl', '+23276000011',
   '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi',
   'MFI_PARTNER', true, true, true, 'MFI Demo'),
  ('vendor.demo@marketpay.sl', '+23276000012',
   '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi',
   'VENDOR', true, true, true, 'Vendor Demo'),
  ('customer.demo@marketpay.sl', '+23276000013',
   '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi',
   'CUSTOMER', true, true, true, 'Customer Demo')
ON CONFLICT (email) DO UPDATE SET
  is_demo = true,
  display_name = EXCLUDED.display_name;
