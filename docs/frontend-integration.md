# MarketPay Frontend Integration Guide

## Base URL
```
http://localhost:8080/api/v1
```

## Swagger UI (interactive docs — test all endpoints here)
```
http://localhost:8080/swagger/index.html
```

---

## 1. Authentication

### Register a user
```http
POST /api/v1/auth/register
Content-Type: application/json

{
  "email": "vendor1@example.com",
  "phone": "+23276123456",
  "password": "SecurePass123",
  "role": "VENDOR"
}
```
Roles: `SUPER_ADMIN` `ADMIN` `LOAN_OFFICER` `FIELD_AGENT` `VENDOR` `CUSTOMER` `MFI_PARTNER`

### Login
```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "superadmin@marketpay.sl",
  "password": "password"
}
```
**Response:**
```json
{
  "access_token": "eyJhbGci...",
  "refresh_token": "eyJhbGci...",
  "expires_in": 900
}
```

### Using the token
Add to every protected request:
```http
Authorization: Bearer eyJhbGci...
```

### Refresh token (before access_token expires)
```http
POST /api/v1/auth/refresh
Content-Type: application/json

{
  "refresh_token": "eyJhbGci..."
}
```

---

## 2. JavaScript / TypeScript Example

```typescript
const API = "http://localhost:8080/api/v1";

// Login and store token
async function login(email: string, password: string) {
  const res = await fetch(`${API}/auth/login`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ email, password }),
  });
  const data = await res.json();
  localStorage.setItem("access_token", data.access_token);
  localStorage.setItem("refresh_token", data.refresh_token);
  return data;
}

// Authenticated request helper
async function apiRequest(path: string, options: RequestInit = {}) {
  const token = localStorage.getItem("access_token");
  const res = await fetch(`${API}${path}`, {
    ...options,
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${token}`,
      ...options.headers,
    },
  });

  // Auto-refresh if token expired
  if (res.status === 401) {
    const refreshed = await refreshToken();
    if (refreshed) return apiRequest(path, options); // retry
  }

  if (!res.ok) {
    const err = await res.json();
    throw new Error(err.error || "Request failed");
  }
  return res.json();
}

async function refreshToken() {
  const rt = localStorage.getItem("refresh_token");
  if (!rt) return false;
  const res = await fetch(`${API}/auth/refresh`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ refresh_token: rt }),
  });
  if (!res.ok) return false;
  const data = await res.json();
  localStorage.setItem("access_token", data.access_token);
  localStorage.setItem("refresh_token", data.refresh_token);
  return true;
}
```

---

## 3. React Example (with hooks)

```typescript
// hooks/useAuth.ts
import { useState } from "react";

const API = "http://localhost:8080/api/v1";

export function useAuth() {
  const [user, setUser] = useState(null);

  const login = async (email: string, password: string) => {
    const res = await fetch(`${API}/auth/login`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ email, password }),
    });
    if (!res.ok) throw new Error("Login failed");
    const data = await res.json();
    localStorage.setItem("token", data.access_token);
    return data;
  };

  const logout = async () => {
    const token = localStorage.getItem("token");
    await fetch(`${API}/auth/logout`, {
      method: "POST",
      headers: { Authorization: `Bearer ${token}` },
    });
    localStorage.removeItem("token");
    setUser(null);
  };

  return { user, login, logout };
}
```

```typescript
// hooks/useVendors.ts
export function useVendors() {
  const getVendors = async (page = 1, limit = 20) => {
    const token = localStorage.getItem("token");
    const res = await fetch(
      `${API}/vendors?page=${page}&limit=${limit}`,
      { headers: { Authorization: `Bearer ${token}` } }
    );
    return res.json();
    // Returns: { data: [...], total: 100, page: 1, limit: 20, total_pages: 5 }
  };

  const checkEligibility = async (vendorId: string) => {
    const token = localStorage.getItem("token");
    const res = await fetch(`${API}/vendors/${vendorId}/eligibility`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    return res.json();
    // Returns: { eligible: true } or { eligible: false, reason: "..." }
  };

  return { getVendors, checkEligibility };
}
```

---

## 4. All Endpoints Quick Reference

### Auth
| Method | URL | Auth | Body |
|--------|-----|------|------|
| POST | `/auth/register` | No | `{email, phone, password, role}` |
| POST | `/auth/login` | No | `{email, password}` |
| POST | `/auth/refresh` | No | `{refresh_token}` |
| POST | `/auth/logout` | Yes | — |
| GET | `/auth/me` | Yes | — |

### Vendors
| Method | URL | Auth | Notes |
|--------|-----|------|-------|
| POST | `/vendors` | ADMIN/FIELD_AGENT | Register vendor |
| GET | `/vendors` | ADMIN/OFFICER | List all |
| GET | `/vendors/:id` | Yes | Get one |
| GET | `/vendors/:id/eligibility` | Yes | Loan eligibility check |
| PUT | `/vendors/:id/kyc/approve` | ADMIN | Approve KYC |
| GET | `/vendors/market-associations` | Yes | List markets |

### Loans
| Method | URL | Auth | Notes |
|--------|-----|------|-------|
| POST | `/loans` | VENDOR | Apply for loan |
| GET | `/loans?state=UNDER_REVIEW` | ADMIN/OFFICER | List by state |
| GET | `/loans/:id` | Yes | Get loan |
| GET | `/loans/:id/schedule` | Yes | Repayment schedule |
| PUT | `/loans/:id/approve` | LOAN_OFFICER | Approve |
| PUT | `/loans/:id/reject` | LOAN_OFFICER | Reject |
| PUT | `/loans/:id/disburse` | ADMIN | Disburse |
| GET | `/loans/vendor/:vendor_id` | Yes | Vendor's loans |

### Repayments
| Method | URL | Auth | Body |
|--------|-----|------|------|
| POST | `/repayments` | VENDOR | `{loan_id, amount, monime_reference}` |
| PUT | `/repayments/loans/:id/default` | ADMIN | Mark defaulted |

### Groups
| Method | URL | Auth | Body |
|--------|-----|------|------|
| POST | `/groups` | VENDOR | `{name, description}` |
| GET | `/groups` | ADMIN/OFFICER | List |
| GET | `/groups/:id` | Yes | Get group |
| POST | `/groups/:id/members` | VENDOR | `{vendor_id}` |
| PUT | `/groups/:id/freeze` | ADMIN | `{reason}` |

### Payments
| Method | URL | Auth | Body |
|--------|-----|------|------|
| POST | `/payments` | CUSTOMER/VENDOR | `{customer_id, vendor_id, amount}` |
| PUT | `/payments/:id/complete` | ADMIN | `{monime_reference}` |
| GET | `/payments/vendor/:vendor_id` | Yes | Vendor receipts |

### Reports
| Method | URL | Auth |
|--------|-----|------|
| GET | `/reports/portfolio` | ADMIN/PARTNER |
| GET | `/reports/repayment-rate` | ADMIN/PARTNER |
| GET | `/reports/default-rate` | ADMIN |
| GET | `/reports/disbursement-volume` | ADMIN |
| GET | `/reports/vendor-distribution` | ADMIN |
| GET | `/reports/partner-summary` | ADMIN/PARTNER |
| GET | `/reports/officer-queue` | OFFICER/ADMIN |

### USSD (called by gateway, not frontend)
| Method | URL | Body |
|--------|-----|------|
| POST | `/ussd` | form: `sessionId, phoneNumber, serviceCode, text` |

---

## 5. Pagination

All list endpoints support:
```
GET /vendors?page=1&limit=20
```
Response always includes:
```json
{
  "data": [...],
  "total": 150,
  "page": 1,
  "limit": 20,
  "total_pages": 8
}
```

---

## 6. Error Format

All errors return:
```json
{
  "error": "descriptive message here"
}
```

| HTTP Code | Meaning |
|-----------|---------|
| 400 | Bad request / validation error |
| 401 | Missing or invalid token |
| 403 | Forbidden (wrong role) |
| 404 | Resource not found |
| 409 | Conflict (e.g. duplicate email) |
| 422 | Business rule violation |
| 429 | Rate limit exceeded |
| 500 | Internal server error |

---

## 7. Test with curl (PowerShell)

```powershell
# Login
$login = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/auth/login" `
  -Method POST `
  -ContentType "application/json" `
  -Body '{"email":"superadmin@marketpay.sl","password":"password"}'

$token = $login.access_token

# Get market associations
Invoke-RestMethod -Uri "http://localhost:8080/api/v1/vendors/market-associations" `
  -Headers @{ Authorization = "Bearer $token" }

# Health check (no auth needed)
Invoke-RestMethod -Uri "http://localhost:8080/health"
```

---

## 8. Test with Postman

1. Import the base URL: `http://localhost:8080/api/v1`
2. Create an **Environment** variable: `token`
3. On the Login request → **Tests** tab, add:
   ```javascript
   pm.environment.set("token", pm.response.json().access_token);
   ```
4. On all other requests, set **Authorization** → **Bearer Token** → `{{token}}`

---

## 9. Frontend Framework Quick Setup

### React (Vite)
```bash
npm create vite@latest marketpay-frontend -- --template react-ts
cd marketpay-frontend
npm install axios
npm run dev   # runs on http://localhost:5173  ← already in CORS allowed list
```

### Next.js
```bash
npx create-next-app@latest marketpay-frontend
cd marketpay-frontend
npm run dev   # runs on http://localhost:3000  ← already in CORS allowed list
```

### Vue
```bash
npm create vue@latest marketpay-frontend
cd marketpay-frontend
npm run dev   # runs on http://localhost:5173  ← already in CORS allowed list
```

### Angular
```bash
npm install -g @angular/cli
ng new marketpay-frontend
cd marketpay-frontend
ng serve      # runs on http://localhost:4200  ← already in CORS allowed list
```

All of these ports are pre-configured in CORS — no changes needed.
