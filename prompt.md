# MarketPay — Project Status

## Phase 1 (completed — foundation)

### 1. Demo vs Live mode
- **Database:** `is_demo` on users, vendors, loans, groups, payments, etc.
- **Seed data (migration 000004):** Demo accounts for all 7 roles:
  `superadmin@marketpay.sl`, `admin.demo@marketpay.sl`, `mfi.demo@marketpay.sl`,
  `officer@marketpay.sl`, `agent@marketpay.sl`, `vendor.demo@marketpay.sl`,
  `customer.demo@marketpay.sl` — all use password `password`
- **Backend:** `X-MarketPay-Mode: demo|live` header via middleware scopes all queries
- **Frontend:** Demo/Live toggle on login + dashboard header; Quick Login only in Demo mode

### 2. Monime USSD Exchange (encrypted)
- `pkg/monimeexchange/crypto.go` — RSA-OAEP SHA256 + AES-128-GCM per Monime spec
- `POST /api/v1/monime/exchange` — decrypts request, runs business logic, returns encrypted text/plain
- Wired actions: vendor registration, payment validation & processing, balance check, loan eligibility, loan application + notifications
- See `configs/monime-ussd-flow.json` — replace `YOUR_API_HOST` with public API URL in Monime dashboard
- Set RSA key: `$env:MONIME_RSA_PRIVATE_KEY = Get-Content path\to\private.pem -Raw`

### 3. Real-time notifications (SSE)
- `in_app_notifications` table
- `GET /api/v1/notifications` — list (scoped by demo/live)
- `GET /api/v1/notifications/stream` — SSE stream
- Events: VendorCreated, LoanRequested, RepaymentReceived, etc.

### 4. Schema extensions
- `field_agent_id`, `vendor_code`, freeze fields on vendors
- `freeze_history` table
- `monime_exchange_sessions` for idempotency

---

## Phase 2 — Current Status

| Area | Status | Details |
|------|--------|---------|
| **Role dashboards (Super Admin, Admin, MFI, Loan Officer)** | ✅ **Done** | All use real API data from `/reports/*` endpoints via `use-reporting.ts` hooks |
| **Analytics page** | ✅ **Done** | Uses `usePortfolioStats`, `useDisbursementVolume`, `useRepaymentRate`, `useDefaultRate` — real data |
| **Field Agent → vendor assignment** | ✅ **Done** | `field_agent_id` set at creation; `PUT /vendors/:id/field-agent` endpoint to reassign (LoanOfficer/Admin/SuperAdmin); frontend has "Assign Field Agent" button + modal with field agent dropdown from `GET /auth/users?role=FIELD_AGENT` |
| **Freeze/unfreeze (vendors + groups)** | ✅ **Done** | Full backend: HTTP handlers (`PUT /:id/freeze`, `PUT /:id/unfreeze`), application service, domain model, `freeze_history` logging. Frontend: buttons + reason modals on vendors & group-lending pages |
| **Group lending workflow** | ✅ **Done** | Real API data. Create group, add member (vendor selector), freeze/unfreeze, detail page at `/group-lending/[id]` |
| **Vendor phone-only auth** | ✅ **Done** | USSD path + dashboard login tab; `POST /auth/vendor-login` with phone+PIN |
| **Monime payment collection API** | ⚠️ **Partial** | MonimeCollector adapter wired in payment service; `Initiate` calls Monime API; webhook auto-completes payments; `GET /payments` endpoint exists. Needs end-to-end production testing |
| **Audit log API** | ✅ **Done** | `GET /api/v1/audit-logs` with filtering; Super Admin frontend page at `/audit-logs` |
| **Reporting filtered by `is_demo`** | ✅ **Done** | All reporting handler queries add `AND is_demo = true/false` |
| **Loan approve/reject/disburse** | ✅ **Done** | Approve/Reject (LoanOfficer+); Disburse with Monime ref modal (Admin+) on `/loans/[id]` |
| **Monime payment collection wiring** | ✅ **Done** | MonimeCollector interface + adapter in `main.go`; `Initiate` calls Monime API; webhook completes payments; `GET /payments`; status fix `COMPLETED`→`SUCCESS` |

### Recently Added
- `PUT /vendors/:id/field-agent` — reassign field agent (LoanOfficer, Admin, SuperAdmin)
- `GET /auth/users?role=FIELD_AGENT` — list users by role for dropdown population
- `PUT /groups/:id/freeze` and `PUT /groups/:id/unfreeze` — full freeze/unfreeze for groups
- Frontend: "Assign Field Agent" button + modal on vendor list page

---

## How to Run

### Docker (easiest)
```powershell
cd C:\Users\joshu\marketpay-backend\marketpay
docker compose up --build -d
```
- API: `http://localhost:8080`
- Frontend: `http://localhost:3000`
- Swagger: `http://localhost:8080/swagger/index.html`
- Quick login: `superadmin@marketpay.sl` / `password` → set Demo mode → Quick Login

### Local dev (requires Postgres + Redis)
```powershell
# Terminal 1 — API
go run ./cmd/api -config configs/config.yaml

# Terminal 2 — Worker
go run ./cmd/worker -config configs/config.yaml

# Terminal 3 — Frontend
cd market_pay-frontend
npm run dev
```

---

## Architecture Notes
- **Go backend:** Gin framework, DDD-style packages (`application/`, `domain/model/`, `infrastructure/postgres/`, `interfaces/http/`)
- **Frontend:** Next.js 16 App Router, React Query, Recharts, custom UI components
- **Auth:** JWT (access + refresh), demo/live header middleware
- **Reporting:** Direct SQL queries in HTTP handler (no separate service layer) — all filter by `is_demo`
- **Exceptions:** Reporting module is the only one without a service/repository layer

---

## Key Backend Routes

| Method | Path | Roles | Description |
|--------|------|-------|-------------|
| POST | `/auth/login` | Public | Email/password login |
| POST | `/auth/vendor-login` | Public | Phone/PIN login |
| GET | `/auth/me` | Auth | Current user |
| GET | `/auth/users?role=FIELD_AGENT` | Admin, SuperAdmin, LoanOfficer | List users by role |
| GET/POST | `/vendors` | Various | List/create vendors |
| PUT | `/vendors/:id/freeze` | LoanOfficer+ | Freeze vendor |
| PUT | `/vendors/:id/unfreeze` | LoanOfficer+ | Unfreeze vendor |
| PUT | `/vendors/:id/field-agent` | LoanOfficer+ | Reassign field agent |
| PUT | `/vendors/:id/kyc/approve` | LoanOfficer+ | Approve KYC |
| POST | `/groups` | LoanOfficer+ | Create group |
| PUT | `/groups/:id/freeze` | LoanOfficer+ | Freeze group |
| PUT | `/groups/:id/unfreeze` | LoanOfficer+ | Unfreeze group |
| POST | `/groups/:id/members` | LoanOfficer+ | Add member |
| POST | `/loans` | Vendor+ | Apply for loan |
| PUT | `/loans/:id/approve` | LoanOfficer+ | Approve loan |
| PUT | `/loans/:id/disburse` | Admin+ | Disburse loan |
| GET | `/reports/*` | Various | Dashboard data |
| GET | `/audit-logs` | SuperAdmin | Audit trail |
| GET | `/payments` | Admin+ | Payment history |

---

## Remaining Work / Suggestions
- **Monime payment collection:** Production testing of webhook callback flow
- **Monime USSD exchange:** End-to-end testing with actual Monime dashboard
- **Reporting:** Consider adding a dedicated service/repository layer to match other modules
- **Notifications:** Wire more event types if needed
- **Performance:** Add indexes if reporting queries become slow
- **Tests:** Unit/integration tests across the stack
