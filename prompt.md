Phase 1 completed (foundation)

1. Demo vs Live mode
   Database: is_demo on users, vendors, loans, groups, payments, etc.
   Migration 000004: Marks seed users as Demo Data and adds demo accounts for all 7 roles:
   superadmin@marketpay.sl, admin.demo@marketpay.sl, mfi.demo@marketpay.sl, officer@marketpay.sl, agent@marketpay.sl, vendor.demo@marketpay.sl, customer.demo@marketpay.sl
   Password: password
   Backend: X-MarketPay-Mode: demo|live header via middleware
   Frontend: Demo/Live toggle on login + dashboard; Quick Login only in Demo mode; auto dev-login removed in Live mode
2. Monime USSD Exchange (encrypted)
   pkg/monimeexchange/crypto.go — RSA-OAEP SHA256 + AES-128-GCM per Monime spec
   POST /api/v1/monime/exchange — decrypts request, runs business logic, returns encrypted text/plain
   Exchange actions wired:
   Vendor registration (after market name)
   Payment validation & processing
   Balance check
   Loan eligibility
   Loan application + Loan Officer/MFI notifications
   Updated flow: configs/monime-ussd-flow.json — replace YOUR_API_HOST with your public API URL in Monime dashboard
   Set RSA key:

$env:MONIME_RSA_PRIVATE_KEY = Get-Content path\to\private.pem -Raw
docker compose up --build -d 3. Real-time notifications (SSE)
in_app_notifications table
GET /api/v1/notifications — list (scoped by demo/live mode)
GET /api/v1/notifications/stream — SSE stream
Events fire on: VendorCreated, LoanRequested, RepaymentReceived, etc. 4. Schema extensions
field_agent_id, vendor_code, freeze fields on vendors
freeze_history table
monime_exchange_sessions for idempotency
Apply migrations & restart
cd c:\Users\joshu\marketpay-backend\marketpay
docker compose up --build -d
Phase 2 still needed (your full spec)
Area Status
Role dashboards with charts (Super Admin, MFI PAR, etc.)
Partial — analytics page still mock
Field Agent → vendor assignment UI
Schema ready, APIs/UI pending
Freeze/unfreeze HTTP endpoints + history
Schema ready, handlers pending
Group lending workflow UI
Mock data still
Vendor phone-only auth (no email)
USSD path done; dashboard pending
Monime payment collection API (not just validation)
Needs Monime adapter wiring
Full audit log API
Table exists, no HTTP routes
Reporting filtered by is_demo
Middleware ready, queries need update
How to test now
Frontend: http://localhost:3000 → set Demo mode → Quick Login → any role
Live mode: Sign in with real credentials only (no quick login)
Monime: Upload configs/monime-ussd-flow.json to Monime dashboard (update API host), add your RSA public key, set MONIME_RSA_PRIVATE_KEY on the API
Notifications: After login, open GET /api/v1/notifications/stream with Bearer token
Recommended next step
Pick one area to finish next:

Super Admin + MFI dashboards with real charts from /reports/\*
Field Agent workflow (register vendor, assign to group, notify Loan Officer)
Freeze/unfreeze endpoints + notification chain
Full Monime payment collection via Exchange after PIN page
Tell me which phase-2 area you want first and I'll continue from there.
