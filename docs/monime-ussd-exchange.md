# Monime USSD Exchange

The MarketPay backend supports Monime's encrypted USSD Exchange protocol for vendor registration, payments, balance checks, and loan applications.

## RSA Key Setup

Monime exchange uses RSA-OAEP SHA-256 for key exchange and AES-128-GCM for payload encryption.

### Generate keys

```bash
openssl genrsa -out monime_private_key.pem 2048
openssl rsa -in monime_private_key.pem -pubout -out monime_public_key.pem
```

### Deploy to Render

1. **Private key** – Set as a Render secret environment variable or secret file:

   - `MONIME_RSA_PRIVATE_KEY` – paste the full PEM contents as a single environment variable.
   - OR `MONIME_RSA_KEY_FILE` – path to a mounted secret file containing the PEM.

2. **Public key** – Paste the contents of `monime_public_key.pem` into the Monime USSD flow **Security** section.

3. **Never commit the private key** to GitHub. The `.gitignore` file already excludes the `keys/` directory.

## Verification

After deployment, confirm the logs contain:

```
monime exchange endpoint enabled
```

Test the health endpoint:

```bash
curl https://market-pay-fpsi.onrender.com/api/v1/monime/exchange/health
```

Expected response:

```json
{
  "status": "ok",
  "exchange_enabled": true,
  "key_loaded": true
}
```

## Exchange Endpoint

`POST https://market-pay-fpsi.onrender.com/api/v1/monime/exchange`

The endpoint accepts encrypted requests with:

- `encryptedAesKey` – RSA-OAEP SHA-256 encrypted AES-128 key (base64)
- `encryptedExchangeData` – AES-128-GCM encrypted exchange payload (base64)

Returns `text/plain` containing the base64-encrypted response.

## Access Gate Exchange

The MarketPay USSD flow must start with an access-control exchange page before showing the MarketPay menu.

The first page in the USSD flow must be:

```text
mp_access_gate_exchange
```

This page silently calls:

```text
https://market-pay-fpsi.onrender.com/api/v1/monime/exchange
```

The backend decrypts the Monime exchange payload and reads:

```text
global.subscriberId
```

Do not use `global.subscriberMsisdn` for access control because it may be masked.

The backend normalizes the subscriber ID, hashes it with SHA-256 if it is not already a 64-character lowercase SHA-256 hash, and compares it against the `ussd_allowed_subscribers` table.

If the subscriber is allowed, the backend returns:

```json
{
  "action": "navigate",
  "pageId": "mp_select_service",
  "pageData": {
    "access_granted": true
  }
}
```

If the subscriber is not allowed, missing, inactive, invalid, or if an internal error happens, the backend returns:

```json
{
  "action": "stop",
  "message": "Flow doesn't exist."
}
```

Important:
The user must not see the MarketPay menu unless `mp_access_gate_exchange` succeeds.

### Database table

```sql
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS ussd_allowed_subscribers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    subscriber_id_hash TEXT NOT NULL UNIQUE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    label TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ussd_allowed_subscribers_hash_active
ON ussd_allowed_subscribers (subscriber_id_hash, is_active);
```

### Seed data

```sql
INSERT INTO ussd_allowed_subscribers (subscriber_id_hash, label, is_active)
VALUES
('a81aa760bee43a675750076012f74d4e43248898b077b934547d3f6dceefdc68', 'Nelson', TRUE),
('82a05af957af142cd35af548d76596ba0cf21730023e032adaef2c7c279d7270', 'Moses Moore', TRUE),
('b7e8504f4419ded90df8d15d2846814894b1fdfb0ad84f4974709ee0d94c3e87', 'Mr Cyril', TRUE),
('bdb7b76c58a77bd2dc791a01e19651276e820b4e0ad8f9568482dde3b0e0579d', 'Sanah', TRUE)
ON CONFLICT (subscriber_id_hash)
DO UPDATE SET
    label = EXCLUDED.label,
    is_active = EXCLUDED.is_active,
    updated_at = NOW();
```

### Migration

The access gate table is created and seeded by migration `migrations/000006_ussd_allowed_subscribers.up.sql`.

### Implementation

- `normalizeSubscriberHash(subscriberID)` — SHA-256 hex digest of the raw subscriber ID (never re-hash).
- `handleAccessGateExchange` — checks `ussd_allowed_subscribers` for the hash with `is_active = true`; returns `stop` with `"Flow doesn't exist."` if denied, missing, or empty subscriber ID.
- Route case `"mp_access_gate_exchange"` added to the `route()` switch in `service.go:119`.
- `mp_access_gate_exchange` is the first page in `configs/monime-ussd-flow.json`.
- Idempotency key follows the same `sessionId + "-" + currentPage` pattern as other routes.

### Security note

Do not log raw subscriber IDs, phone numbers, private keys, database passwords, or decrypted exchange payloads. If debugging subscriber authorization, log only a short safe fingerprint such as the first 8 characters of the computed hash.

## Supported Pages

| Page ID | Purpose |
|---|---|
| `mp_access_gate_exchange` | First-page subscriber access check |
| `mp_collect_market_name` | Vendor registration |
| `mp_confirm_payment_receipt` | Payment confirmation |
| `mp_collect_payment_pin` | Payment processing |
| `mp_balance_exchange` | Balance check |
| `mp_loan_eligibility_exchange` | Loan eligibility check |
| `mp_confirm_loan_application` | Loan application |

## Response Format

Navigate to another page:

```json
{
  "action": "navigate",
  "pageId": "page_id",
  "pageData": { ... }
}
```

End the session:

```json
{
  "action": "stop",
  "message": "User-friendly message"
}
```
