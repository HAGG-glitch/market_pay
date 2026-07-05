# MarketPay — Testing Guide & Login Credentials

> **Base URL (prod):** `https://market-pay-fpsi.onrender.com`  
> **Base URL (local):** `http://localhost:8080`  
> **Frontend:** `https://market-pay.vercel.app`  
> **Swagger UI:** `http://localhost:8080/swagger/index.html`

---

## 1. Login Credentials by Role

### Web Login (email + password)

All passwords below are the **actual working passwords** matching the bcrypt hashes in the seed data.

| Role | Email | Password | Phone |
|---|---|---|---|
| **SUPER_ADMIN** | `superadmin@marketpay.sl` | `password` | `+23276000001` |
| **ADMIN** | `admin.demo@marketpay.sl` | `password` | `+23276000010` |
| **LOAN_OFFICER** | `officer@marketpay.sl` | `password` | `+23276000003` |
| **LOAN_OFFICER** | `joshua@marketpay.sl` | `password123` | `+23273608360` |
| **FIELD_AGENT** | `agent@marketpay.sl` | `password` | `+23276000004` |
| **MFI_PARTNER** | `mfi.demo@marketpay.sl` | `password` | `+23276000011` |
| **VENDOR** | `vendor.demo@marketpay.sl` | `password` | `+23276000012` |
| **CUSTOMER** | `customer.demo@marketpay.sl` | `password` | `+23276000013` |

### Vendor Login (phone + PIN via `/auth/vendor-login`)

| Name | Phone | PIN | Mobile Network |
|---|---|---|---|
| Abraham Araia | `+23276902971` | `2006` | Orange SL (m17) |
| Emmanuel Nelson | `+23279826564` | `2005` | Orange SL (m17) |
| Moses Moore | `+23233346989` | `2004` | Africell (m18) |
| Bernard Gamanga | `+23272571067` | `2005` | Orange SL (m17) |

> **Note:** The seed migration file says `Admin@1234` in the comments, but the actual bcrypt hash matches `password`.

---

## 2. API Endpoint Reference

### Authentication

| Method | Endpoint | Auth | Body |
|---|---|---|---|
| POST | `/api/v1/auth/register` | No | `{email, phone, password, role}` |
| POST | `/api/v1/auth/login` | No | `{email, password}` |
| POST | `/api/v1/auth/vendor-login` | No | `{phone, pin}` |
| POST | `/api/v1/auth/refresh` | No | `{refresh_token}` |
| POST | `/api/v1/auth/logout` | Yes | — |
| GET | `/api/v1/auth/me` | Yes | — |
| GET | `/api/v1/auth/users?role=LOAN_OFFICER` | Yes | — |

### Vendors

| Method | Endpoint | Required Roles | Notes |
|---|---|---|---|
| POST | `/api/v1/vendors` | ADMIN / SUPER_ADMIN / FIELD_AGENT | Register vendor |
| GET | `/api/v1/vendors` | ADMIN / SUPER_ADMIN / LOAN_OFFICER / FIELD_AGENT | List all |
| GET | `/api/v1/vendors/:id` | Yes | Get one |
| GET | `/api/v1/vendors/:id/eligibility` | Yes | Check loan eligibility |
| PUT | `/api/v1/vendors/:id/kyc/approve` | ADMIN / SUPER_ADMIN / LOAN_OFFICER | Approve KYC |
| PUT | `/api/v1/vendors/:id/freeze` | LOAN_OFFICER / ADMIN / SUPER_ADMIN | Freeze vendor |
| PUT | `/api/v1/vendors/:id/unfreeze` | LOAN_OFFICER / ADMIN / SUPER_ADMIN | Unfreeze vendor |
| PUT | `/api/v1/vendors/:id/field-agent` | LOAN_OFFICER / ADMIN / SUPER_ADMIN | Assign field agent |
| GET | `/api/v1/vendors/market-associations` | Yes | List markets |

### Loans

| Method | Endpoint | Required Roles | Notes |
|---|---|---|---|
| POST | `/api/v1/loans` | VENDOR | Apply for loan |
| GET | `/api/v1/loans?state=UNDER_REVIEW` | ADMIN / SUPER_ADMIN / LOAN_OFFICER | List by state |
| GET | `/api/v1/loans/:id` | Yes | Get loan |
| GET | `/api/v1/loans/:id/schedule` | Yes | Repayment schedule |
| PUT | `/api/v1/loans/:id/approve` | LOAN_OFFICER / ADMIN / SUPER_ADMIN | Approve |
| PUT | `/api/v1/loans/:id/reject` | LOAN_OFFICER / ADMIN / SUPER_ADMIN | Reject |
| PUT | `/api/v1/loans/:id/disburse` | ADMIN / SUPER_ADMIN | Disburse (sends mobile money) |
| PUT | `/api/v1/loans/:id/revert-disbursement` | LOAN_OFFICER / ADMIN / SUPER_ADMIN | Revert failed disbursement |
| GET | `/api/v1/loans/vendor/:vendor_id` | Yes | Vendor's loans |
| GET | `/api/v1/loans/payment-plans` | LOAN_OFFICER / ADMIN / SUPER_ADMIN | List payment plans |

### Repayments

| Method | Endpoint | Required Roles | Notes |
|---|---|---|---|
| POST | `/api/v1/repayments` | VENDOR / ADMIN | `{loan_id, amount, monime_reference}` |
| PUT | `/api/v1/repayments/loans/:id/default` | ADMIN / SUPER_ADMIN | Mark loan as defaulted |

### Groups

| Method | Endpoint | Required Roles | Notes |
|---|---|---|---|
| POST | `/api/v1/groups` | VENDOR / ADMIN / FIELD_AGENT | Create group |
| GET | `/api/v1/groups` | ADMIN / SUPER_ADMIN / LOAN_OFFICER / FIELD_AGENT | List groups |
| GET | `/api/v1/groups/:id` | Yes | Get group |
| POST | `/api/v1/groups/:id/members` | VENDOR / ADMIN / FIELD_AGENT | Add member |
| PUT | `/api/v1/groups/:id/freeze` | LOAN_OFFICER / ADMIN / SUPER_ADMIN | Freeze group |
| PUT | `/api/v1/groups/:id/unfreeze` | LOAN_OFFICER / ADMIN / SUPER_ADMIN | Unfreeze group |

### Payments

| Method | Endpoint | Required Roles | Notes |
|---|---|---|---|
| POST | `/api/v1/payments` | CUSTOMER / VENDOR | Create payment |
| PUT | `/api/v1/payments/:id/complete` | ADMIN | Complete with ref |
| GET | `/api/v1/payments/vendor/:vendor_id` | Yes | Vendor receipts |

### Reports

| Method | Endpoint | Required Roles |
|---|---|---|
| GET | `/api/v1/reports/portfolio` | ADMIN / SUPER_ADMIN / MFI_PARTNER |
| GET | `/api/v1/reports/repayment-rate` | ADMIN / SUPER_ADMIN / MFI_PARTNER |
| GET | `/api/v1/reports/default-rate` | ADMIN / SUPER_ADMIN |
| GET | `/api/v1/reports/disbursement-volume` | ADMIN / SUPER_ADMIN / MFI_PARTNER |
| GET | `/api/v1/reports/vendor-distribution` | ADMIN / SUPER_ADMIN |
| GET | `/api/v1/reports/partner-summary` | ADMIN / SUPER_ADMIN / MFI_PARTNER |
| GET | `/api/v1/reports/officer-queue` | LOAN_OFFICER / ADMIN / SUPER_ADMIN |
| GET | `/api/v1/reports/dashboard-summary` | ADMIN / SUPER_ADMIN / MFI_PARTNER / LOAN_OFFICER |

### Notifications & Audit

| Method | Endpoint | Auth |
|---|---|---|
| GET | `/api/v1/notifications` | Yes (scoped by demo/live mode) |
| GET | `/api/v1/notifications/stream` | Yes (SSE stream) |
| GET | `/api/v1/audit-logs` | SUPER_ADMIN only |
| GET | `/health` | No (health check) |

### USSD (called by Monime gateway, not frontend)

| Method | Endpoint | Body |
|---|---|---|
| POST | `/api/v1/ussd` | form: `sessionId, phoneNumber, serviceCode, text` |

---

## 3. Testing Scenarios

### Scenario 1: Super Admin — Full System Access

1. Login at `https://market-pay.vercel.app` with `superadmin@marketpay.sl` / `password`
2. View dashboard overview
3. Navigate to Audit Logs — see all system actions
4. Navigate to Reports — see all analytics
5. Manage vendors, loans, groups

### Scenario 2: Loan Officer — Loan Lifecycle

1. Login as `joshua@marketpay.sl` / `password123`
2. Navigate to **Vendors** → see Araia, Nelson, Moses, Bernard
3. Navigate to **Loans** → review pending applications
4. Create a loan for a vendor:
   - Click "Apply" or create via API
   - Set amount and payment plan
5. **Approve** the loan
6. **Disburse** the loan (triggers Monime mobile money payout)
   - Araia (`+23276902971`): sends via Orange SL (m17)
   - Moses (`+23233346989`): sends via Africell (m18)

### Scenario 3: Vendor — USSD Loan Application

1. Dial USSD code (vendor flow)
2. Enter PIN to authenticate
3. Select "Apply for loan"
4. Enter amount
5. Confirm
6. Loan officer sees the application in the dashboard

### Scenario 4: Field Agent — Vendor Onboarding

1. Login as `agent@marketpay.sl` / `password`
2. View assigned vendors
3. Approve KYC for pending vendors
4. Navigate to create new vendor

### Scenario 5: MFI Partner — Portfolio View

1. Login as `mfi.demo@marketpay.sl` / `password`
2. View portfolio summary
3. Check disbursement volume and repayment rates
4. View partner summary report

---

## 4. Mobile Money Provider Prefixes

| Network | Monime Provider ID | Phone Prefixes |
|---|---|---|
| **Orange SL** | `m17` | `+23272`, `+23276`, `+23277`, `+23278`, `+23279` |
| **Africell** | `m18` | `+23230`, `+23233`, `+23234` (also the default fallback) |

Configured in `configs/config.yaml` under `monime.payout.provider_mappings`.

---

## 5. Test with curl (PowerShell)

```powershell
# Login
$login = Invoke-RestMethod -Uri "https://market-pay-fpsi.onrender.com/api/v1/auth/login" `
  -Method POST `
  -ContentType "application/json" `
  -Body '{"email":"superadmin@marketpay.sl","password":"password"}'

$token = $login.access_token

# Get market associations
Invoke-RestMethod -Uri "https://market-pay-fpsi.onrender.com/api/v1/vendors/market-associations" `
  -Headers @{ Authorization = "Bearer $token" }

# List vendors
Invoke-RestMethod -Uri "https://market-pay-fpsi.onrender.com/api/v1/vendors" `
  -Headers @{ Authorization = "Bearer $token" }

# List loans
Invoke-RestMethod -Uri "https://market-pay-fpsi.onrender.com/api/v1/loans?state=UNDER_REVIEW" `
  -Headers @{ Authorization = "Bearer $token" }

# Health check (no auth)
Invoke-RestMethod -Uri "https://market-pay-fpsi.onrender.com/health"
```

---

## 6. System Architecture

```
┌────────────────┐     ┌──────────────────┐     ┌───────────────┐
│   Monime       │────▶│  MarketPay API   │────▶│  PostgreSQL   │
│  USSD Gateway  │     │  (Render)        │     │  (Render)     │
│  + Payout API  │◀────│  :8080           │◀────│  savwise_ai   │
└────────────────┘     └──────────────────┘     └───────────────┘
                               │
                               ▼
                        ┌───────────────┐
                        │   Frontend    │
                        │  (Vercel)     │
                        └───────────────┘
```

### Roles & Permissions

| Role | Description |
|---|---|
| **SUPER_ADMIN** | Full system oversight, audit logs, all actions |
| **ADMIN** | Operations management, vendor/loan/group admin |
| **MFI_PARTNER** | Microfinance institution — portfolio reports only |
| **LOAN_OFFICER** | Reviews loans, manages vendors & groups |
| **FIELD_AGENT** | Onboards vendors in the field |
| **VENDOR** | Market vendor — applies for loans, makes payments |
| **CUSTOMER** | End customer — pays vendors via USSD |

---

## 7. Common Issues & Fixes

### Issue: Login returns 401
- Verify you're using `password` (not `Admin@1234`) for accounts from seed migrations 000002 and 000004
- Verify you're using `password123` for accounts from migration 000007
- Check that `is_active = true` in the database
- For vendors: use `/auth/vendor-login` with phone + PIN (not email/password)

### Issue: Loan disbursement fails
- Check the vendor's phone prefix matches a configured provider mapping
- Africell numbers use `m18`, Orange SL numbers use `m17`
- Unknown prefixes fall back to the default provider

### Issue: Demo/Live mode
- All seeded accounts are **Demo** mode
- API responses are scoped by the `X-MarketPay-Mode: demo|live` header
- Quick Login (impersonation) is only available in Demo mode

---

## 8. Quick Start (Local Development)

```powershell
# Clone and configure
Copy-Item .env.example .env
docker compose up --build -d

# Run migrations (after DB is ready)
./migrate up

# Start API
go run ./cmd/api -config configs/config.local.yaml
```

**Local URLs:**
- API: `http://localhost:8080`
- Swagger: `http://localhost:8080/swagger/index.html`
- Frontend: `http://localhost:3000`

---

*Generated from migrations/000002_seed_data.up.sql, 000004_demo_accounts.up.sql, 000007_seed_loan_officer_vendors.up.sql*
