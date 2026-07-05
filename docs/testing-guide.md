# MarketPay — Testing Guide & System Overview

> **Last updated:** 22 June 2026  
> **Base URL (prod):** `https://market-pay-fpsi.onrender.com`  
> **Frontend:** `https://market-pay.vercel.app` (manual redeploy required after backend push)

---

## Table of Contents

1. [System Architecture](#1-system-architecture)
2. [USSD Flows](#2-ussd-flows)
3. [Vendor Registration (New)](#3-vendor-registration)
4. [Loan Disbursement](#4-loan-disbursement)
5. [Repayment Tracking (New)](#5-repayment-tracking)
6. [Test Accounts](#6-test-accounts)
7. [Testing Scenarios](#7-testing-scenarios)
8. [Key Decisions](#8-key-decisions)

---

## 1. System Architecture

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

### Key Components

| Component | Technology | Location |
|---|---|---|
| API Server | Go + Gin + GORM | `cmd/api/main.go` |
| Database | PostgreSQL | Render (savwise_ai) |
| Migrations | SQL files | `migrations/000001`–`000011` |
| USSD Exchange | Monime platform + RSA crypto | `internal/monime/exchange/` |
| Payouts | Monime Payout API | `pkg/monimepayout/` |
| Webhooks | Monime v2 payload | `internal/monime/interfaces/http/webhook.go` |
| Frontend | Next.js (standalone) | `market_pay-frontend/` |

---

## 2. USSD Flows

### 2a. Public Flow (`monime-ussd-public-flow.json`)

No access gate — any mobile subscriber can use it.

```
Welcome to MarketPay
1. Pay a vendor
2. Register as vendor
3. Exit
```

#### Pay a Vendor Path
```
→ Enter vendor code (e.g. MP12345)
→ Enter amount in SLE (min 1)
→ Confirm & pay / Cancel
→ PIN entry (Monime platform handles collection)
→ Result displayed
```

#### Register as Vendor Path
```
→ Enter full name
→ Enter market name
→ Enter phone number (e.g. 23276123456)
→ Confirm / Cancel
→ Result: "Registration submitted. A field agent will contact you."
```

### 2b. Vendor Flow (`monime-ussd-vendor-flow.json`)

Access gate required — subscriber must be in `ussd_allowed_subscribers`.

```
→ Access Gate (SHA-256 subscriberId check)
  → Menu:
    1. My credit score
    2. Apply for loan
    3. Repay loan
    4. Check balance
    5. Transaction history
    6. Exit
```

#### Repay Loan Path
```
→ Shows outstanding balance
→ Enter repayment amount
→ Confirm & pay / Cancel
→ PIN entry (Monime platform handles collection)
→ Result displayed
→ LoanRepayment record created (PENDING)
→ Webhook financial_account.credited → confirmed (COMPLETED)
```

---

## 3. Vendor Registration

### Flow Steps

1. Subscriber selects "Register as vendor" on public USSD menu
2. Collects: name → market → phone number
3. Phone normalization: `normalizePhone()` handles all formats:
   - `23276123456` → `+23276123456`
   - `076123456` → `+232076123456`
   - `+23276123456` → `+23276123456`
4. Creates `users` record with `is_active = false`
5. Creates `vendors` record with `Status = PENDING`
6. Links subscriber to vendor in `ussd_subscribers`
7. NOT added to `ussd_allowed_subscribers` yet

### Approval (Waiting Room)

After USSD registration, the vendor is in a "waiting room":

| State | Can log in? | Can use vendor USSD menu? |
|---|---|---|
| Just registered (PENDING) | No (`is_active = false`) | No (not in allowed_subscribers) |
| Field agent approves KYC | Yes (`is_active = true`) | Yes (if subscriberId added) |

**Approval endpoint:** `PUT /api/v1/vendors/:id/kyc/approve`  
- Sets vendor status → `ACTIVE`
- Sets KYC status → `VERIFIED`
- Sets `users.is_active = true`
- To also grant USSD access, add subscriberId to `ussd_allowed_subscribers` (via DB or future endpoint)

### Registration Message

> *"Registration submitted for [Name] at [Market]. Phone: [Phone]. A field agent will contact you to verify and activate your account."*

---

## 4. Loan Disbursement

### States

```
APPROVED → DISBURSEMENT_PENDING → ACTIVE (on payout.completed)
                                 → APPROVED (on payout.failed)
```

### Disburse Action

`PUT /api/v1/loans/:id/disburse`

1. Validates loan is `APPROVED` and user has permission
2. Determines mobile money provider from phone prefix:
   - `+23276/77/78` → `m17` (Orange SL)
   - `+23230/33/34` → `m18` (Africell)
   - configurable via `provider_mappings` in `config.yaml`
3. Calls Monime Payout API (`POST /v1/payouts`)
4. Sets loan state to `DISBURSEMENT_PENDING`
5. Stores `monime_reference` = payout ID

### Payout Webhooks

| Event | Handler | Effect |
|---|---|---|
| `payout.completed` | `handlePayoutCompleted` | State → `ACTIVE`, generates schedules, sets `provider_ref` |
| `payout.failed` | `handlePayoutFailed` | State → `APPROVED`, clears ref, stores `failure_reason` |

### Failed Payout — Retry

If a payout fails, the loan returns to `APPROVED`. The three-dot menu on the loans page allows re-disbursing.

---

## 5. Repayment Tracking

### Table: `loan_repayments`

| Column | Type | Purpose |
|---|---|---|
| `id` | UUID (PK) | Primary key |
| `loan_id` | UUID (FK→loans) | The loan being repaid |
| `vendor_id` | UUID (FK→vendors) | The vendor repaying |
| `amount` | DECIMAL(15,2) | Repayment amount |
| `monime_ref` | VARCHAR(255) | Monime collection reference (webhook match key) |
| `payment_ref` | VARCHAR(255) UNIQUE | Our internal reference (`REPAY-{session}-{ts}`) |
| `metadata` | JSONB | Session data, masked_phone, subscriber_id |
| `status` | VARCHAR(50) | `PENDING` → `COMPLETED` / `FAILED` |
| `paid_at` | TIMESTAMPTZ | When confirmed |
| `created_at` / `updated_at` | TIMESTAMPTZ | Audit timestamps |

### How Repayments Work

```
USSD Repayment Flow:

1. Vendor selects "Repay loan" → sees balance
2. Enters amount → confirms
3. Monime collect_payment template processes PIN & collection
4. Callback hits handleRepaymentResult:
   - Extracts ExternalRef from ExportedData (or falls back to payment_ref)
   - Creates LoanRepayment record (status=PENDING)
   - Shows result to user
5. Monime sends webhook financial_account.credited:
   - handleAccountCredited matches by monime_ref (object.id)
   - Fallback: matches by metadata.payment_ref
   - Calls RepaymentService.ConfirmRepayment()
     → marks LoanRepayment as COMPLETED
     → calls existing Repay() to apply to loan schedules
```

### Webhook Matching Strategy

1. **Primary:** Match `financial_account.credited.object.id` against `loan_repayments.monime_ref`
2. **Fallback:** Check `data.metadata.payment_ref` against `loan_repayments.payment_ref`

If no match is found, the repayment stays `PENDING` and must be reconciled manually.

---

## 6. Test Accounts

### Vendors (Login via phone + PIN on website)

| Name | Phone | PIN | Mobile Money |
|---|---|---|---|
| Araia | `+23276902971` | `2006` | Orange SL (m17) |
| Moses Moore | `+23233346989` | `2004` | Africell (m18) |
| Nelson | `+23279826564` | `2005` | Orange SL (m17) |
| Bernard (Gamanga) | `+23272571067` | `2005` | Orange SL (m17) |

### Loan Officer (Login via email + password)

| Name | Email | Password | Role |
|---|---|---|---|
| Joshua Yoki | `joshua@marketpay.sl` | `password123` | LOAN_OFFICER |

### Vendor Phones by Network

| Network | Prefix | Monime Provider ID |
|---|---|---|
| Orange SL | `+23276`, `+23277`, `+23278` | `m17` |
| Africell | `+23230`, `+23233`, `+23234` | `m18` |

---

## 7. Testing Scenarios

### Scenario 1: Public Pay Vendor

1. Dial USSD code (public flow)
2. Select "Pay a vendor"
3. Enter vendor code `MP12345` (or any valid code)
4. Enter amount (e.g. `10`)
5. Confirm → PIN entry → success
6. **Verify:** Notification sent to LOAN_OFFICER role

### Scenario 2: Public Vendor Registration

1. Dial USSD code (public flow)
2. Select "Register as vendor"
3. Enter name → market → phone
4. Confirm
5. **Verify:** Message says "Registration submitted. A field agent will contact you."
6. **Verify:** New vendor created with `Status = PENDING`, `users.is_active = false`

### Scenario 3: Field Agent Approval

1. Log into website as LOAN_OFFICER
2. Navigate to vendor management
3. Find the new vendor
4. Click "Approve KYC"
5. **Verify:** Vendor status → `ACTIVE`, user → `is_active = true`
6. **Optional:** Add subscriberId to `ussd_allowed_subscribers` for USSD access

### Scenario 4: Full Loan Cycle (Moses Moore — Africell)

**Should succeed (Africell + m18):**

1. Log in as LOAN_OFFICER
2. Create loan for Moses Moore (`+23233346989`)
3. Approve loan
4. Disburse loan (should use `m18` provider)
5. **Verify:** Loan state → `DISBURSEMENT_PENDING`
6. Wait for webhook `payout.completed`
7. **Verify:** Loan state → `ACTIVE`, schedules generated
8. Moses checks balance via USSD → shows outstanding

### Scenario 5: Full Loan Cycle (Araia — Orange SL)

**Should succeed (Orange + m17):**

1. Log in as LOAN_OFFICER
2. Create loan for Araia (`+23276902971`)
3. Approve loan
4. Disburse loan (should use `m17` provider)
5. **Verify:** Loan state → `DISBURSEMENT_PENDING`
6. Wait for webhook `payout.completed`
7. **Verify:** Loan state → `ACTIVE`, schedules generated

### Scenario 6: Repay via USSD (Vendor Flow)

Prerequisite: Vendor has an ACTIVE loan.

1. Dial USSD code (vendor flow)
2. Select "Repay loan"
3. Enter amount
4. Confirm → PIN entry
5. **Verify:** `LoanRepayment` record created (`status = PENDING`)
6. Wait for webhook `financial_account.credited`
7. **Verify:** `LoanRepayment` status → `COMPLETED`
8. **Verify:** Loan schedule updated (AmountPaid increased)

### Scenario 7: Payout Failure → Retry

1. Disburse loan (any vendor)
2. Webhook returns `payout.failed`
3. **Verify:** Loan state → `APPROVED`, `failure_reason` set
4. On loans page, click three-dot menu → "Retry Disbursement"
5. **Verify:** New payout sent, state → `DISBURSEMENT_PENDING` again

---

## 8. Key Decisions

### Webhook Signature
Monime sends empty `X-Monime-Signature`. We log a warning but accept the webhook. No 401 rejection.

### Phone Normalization
`normalizePhone()` strips non-digits, handles `232...`, `076...`, `+...` formats. Outputs `+232...` format.

### Subscriber Identity
- `global.subscriberId` = SHA-256 hash of MSISDN — used for identification
- `global.subscriberMsisdn` = masked (e.g. `233XX XXX 4567`) — display only, never used as lookup

### Provider Selection
Provider ID (`m17` / `m18`) resolved from `provider_mappings` in `config.yaml` by phone prefix, not hardcoded.

### Migration Auto-Apply
Render runs `./migrate up && ./app` on startup. Migrations 000001–000011 apply automatically.

### Loan Source
- `WEB` (default) — created via website
- `USSD` — created via USSD application flow
