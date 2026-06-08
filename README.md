<<<<<<< HEAD
# MarketPay Backend

Production-grade Go backend for a microfinance and USSD payment platform serving informal market vendors in Sierra Leone.

---

## Overview

MarketPay acts as middleware between vendors, customers, MFI partners, and NGOs. It handles vendor onboarding, loan origination, repayment collection, credit scoring, USSD flows, and double-entry ledger accounting.

### Tech Stack

| Concern | Technology |
|---|---|
| Language | Go 1.24+ |
| REST API | Gin |
| RPC | gRPC |
| Database | PostgreSQL 16 + GORM |
| Cache / Rate Limiting | Redis 7 |
| Auth | JWT (access 15min / refresh 7d) |
| Migrations | golang-migrate |
| Logging | Zap (structured JSON) |
| Config | Viper |
| Testing | Testify + Mockery |
| Containers | Docker + Docker Compose |
| API Docs | Swagger / OpenAPI |

---

## Quick Start

### Prerequisites
- [Docker Desktop](https://www.docker.com/products/docker-desktop) — required for all platforms
- [Go 1.24+](https://go.dev/dl/) — required for local builds / `go mod tidy`

### Windows (PowerShell)

```powershell
# Inside the extracted marketpay folder:
Copy-Item .env.example .env
go mod tidy
.\scripts\make.ps1 quickstart
```

### macOS / Linux

```bash
cp .env.example .env
go mod tidy
make quickstart
```

After startup:
| URL | Purpose |
|---|---|
| `http://localhost:8080` | REST API |
| `http://localhost:8080/swagger/index.html` | Swagger UI |
| `http://localhost:8080/health` | Health check |

Default seed credentials:

| Role | Email | Password |
|---|---|---|
| Super Admin | superadmin@marketpay.sl | password |
| Loan Officer | officer@marketpay.sl | password |
| Field Agent | agent@marketpay.sl | password |

---

## Architecture

```
cmd/
  api/        → HTTP + gRPC server entry point (DI wiring)
  worker/     → Outbox event worker
  migrate/    → Database migration runner

internal/
  auth/           → JWT auth, refresh tokens
  vendor/         → Vendor registration, KYC, PIN
  customer/       → Customer records
  loan/           → Loan origination, state machine, schedules
  repayment/      → Repayments, penalties, defaults
  group/          → Group lending (5–10 members, guarantee)
  creditscore/    → Weighted credit scoring engine
  payment/        → Customer→Vendor payments (1% fee)
  ledger/         → Double-entry bookkeeping
  partner/        → MFI/NGO/Bank partner management
  notification/   → SMS / WhatsApp / Email dispatcher
  reporting/      → Admin, partner, officer reports
  audit/          → Immutable audit log
  ussd/           → USSD session state machine
  monime/         → Payment gateway ports/adapters
  shared/         → Base models, value objects

pkg/
  config/         → Viper config loader
  logger/         → Zap wrapper
  errors/         → Typed AppError with HTTP status
  middleware/     → JWT auth, RBAC, rate limiting, security headers
  outbox/         → Transactional outbox publisher + worker
  pagination/     → Page/limit helpers
```

Each bounded context follows Clean Architecture:

```
domain/model/       → Aggregates, entities, value objects, domain logic
domain/repository/  → Repository interfaces
application/        → Use-case services (orchestrate domain)
infrastructure/     → GORM repos, Redis adapters
interfaces/http/    → Gin handlers
interfaces/grpc/    → gRPC server stubs
```

---

## Loan Products

| Product | Amount (SLE) | Term | Interest | Approval |
|---|---|---|---|---|
| Emergency Advance | 50 – 500 | 2 weeks | 5% flat | Auto if score ≥ 75 |
| Starter Loan | 500 – 2,000 | 4–8 weeks | 8% flat | Loan Officer |
| Growth Loan | 2,000 – 5,000 | 12 weeks | 10% declining | Loan Officer |

### Loan States

```
DRAFT → PENDING_REVIEW → AUTO_APPROVED → DISBURSED → ACTIVE → CLOSED
                       → UNDER_REVIEW  → APPROVED  → DISBURSED
                                       → REJECTED
                                                      ACTIVE → DEFAULTED → CLOSED
```

Illegal transitions are rejected by the state machine.

---

## Credit Scoring

| Factor | Weight |
|---|---|
| Transaction Volume | 30 |
| Transaction Consistency | 20 |
| Repayment History | 30 |
| Market Association | 10 |
| KYC Completeness | 10 |
| Group Bonus | +5 |

- Minimum score to apply: **50**
- Auto-approval threshold: **75**

---

## API Endpoints

### Auth
| Method | Path | Description |
|---|---|---|
| POST | `/api/v1/auth/register` | Register user |
| POST | `/api/v1/auth/login` | Login → token pair |
| POST | `/api/v1/auth/refresh` | Refresh access token |
| POST | `/api/v1/auth/logout` | Revoke all tokens |
| GET | `/api/v1/auth/me` | Current user |

### Vendors
| Method | Path | Description |
|---|---|---|
| POST | `/api/v1/vendors` | Register vendor |
| GET | `/api/v1/vendors` | List vendors |
| GET | `/api/v1/vendors/:id` | Get vendor |
| GET | `/api/v1/vendors/:id/eligibility` | Check loan eligibility |
| PUT | `/api/v1/vendors/:id/kyc/approve` | Approve KYC |
| GET | `/api/v1/vendors/market-associations` | List markets |

### Loans
| Method | Path | Description |
|---|---|---|
| POST | `/api/v1/loans` | Apply for loan |
| GET | `/api/v1/loans` | List by state |
| GET | `/api/v1/loans/:id` | Get loan |
| GET | `/api/v1/loans/:id/schedule` | Repayment schedule |
| PUT | `/api/v1/loans/:id/approve` | Approve |
| PUT | `/api/v1/loans/:id/reject` | Reject |
| PUT | `/api/v1/loans/:id/disburse` | Disburse |
| GET | `/api/v1/loans/vendor/:vendor_id` | Vendor loans |

### Repayments
| Method | Path | Description |
|---|---|---|
| POST | `/api/v1/repayments` | Submit repayment |
| PUT | `/api/v1/repayments/loans/:id/default` | Mark defaulted |

### Groups
| Method | Path | Description |
|---|---|---|
| POST | `/api/v1/groups` | Create group |
| GET | `/api/v1/groups` | List groups |
| GET | `/api/v1/groups/:id` | Get group |
| POST | `/api/v1/groups/:id/members` | Add member |
| PUT | `/api/v1/groups/:id/freeze` | Freeze group |

### Payments
| Method | Path | Description |
|---|---|---|
| POST | `/api/v1/payments` | Initiate payment |
| PUT | `/api/v1/payments/:id/complete` | Complete payment |
| GET | `/api/v1/payments/vendor/:vendor_id` | Vendor payments |

### Reports
| Method | Path | Description |
|---|---|---|
| GET | `/api/v1/reports/portfolio` | Portfolio outstanding |
| GET | `/api/v1/reports/repayment-rate` | Repayment rate |
| GET | `/api/v1/reports/default-rate` | Default rate |
| GET | `/api/v1/reports/disbursement-volume` | Disbursement by month |
| GET | `/api/v1/reports/vendor-distribution` | Vendors by market |
| GET | `/api/v1/reports/partner-summary` | Partner commissions |
| GET | `/api/v1/reports/officer-queue` | Officer review queue |

### USSD
| Method | Path | Description |
|---|---|---|
| POST | `/api/v1/ussd` | USSD gateway callback |

---

## USSD Flows (dial `*737#`)

```
1. Register
2. Pay Vendor
3. Sales History
4. Loan Eligibility
5. Apply for Loan
6. Loan Balance
7. Repayment Schedule
8. Repay Loan
9. Group Info
```

All flows are PIN-protected (bcrypt). Rate limited to **10 sessions/hour** per phone number.

---

## Build Commands

### Windows (PowerShell)
```powershell
.\scripts\make.ps1 quickstart     # Start everything via Docker
.\scripts\make.ps1 build          # Build all binaries -> bin/
.\scripts\make.ps1 test           # Run unit tests
.\scripts\make.ps1 test-cover     # Coverage HTML report
.\scripts\make.ps1 docker-up      # docker compose up --build
.\scripts\make.ps1 docker-down    # docker compose down
.\scripts\make.ps1 docker-logs    # Follow API + worker logs
.\scripts\make.ps1 migrate-up     # Run migrations
.\scripts\make.ps1 migrate-down   # Rollback migrations
.\scripts\make.ps1 tidy           # go mod tidy
.\scripts\make.ps1 clean          # Remove build artifacts
```

### macOS / Linux
```bash
make quickstart     # Start everything via Docker
make build          # Build all binaries → bin/
make test           # Run unit tests
make test-cover     # Coverage HTML report
make docker-up      # docker compose up --build
make docker-down    # docker compose down
make migrate-up     # Run migrations
make migrate-down   # Rollback migrations
make swagger        # Generate Swagger docs
make clean          # Remove build artifacts
```

---

## Roles & Permissions

| Role | Capabilities |
|---|---|
| `SUPER_ADMIN` | Full access |
| `ADMIN` | All ops except super-admin config |
| `LOAN_OFFICER` | Review, approve, reject loans |
| `FIELD_AGENT` | Register vendors, assist onboarding |
| `VENDOR` | Apply for loans, make repayments, receive payments |
| `CUSTOMER` | Make payments to vendors |
| `MFI_PARTNER` | View partner reports and commission data |

---

## Environment Variables

See `.env.example` for the full list. Key overrides:

```bash
DATABASE_HOST=postgres
DATABASE_PASSWORD=your-db-password
JWT_ACCESS_SECRET=your-secret
MONIME_API_KEY=your-key
```

---

## Running Tests

```bash
# Unit tests only (no DB required)
make test

# With race detector
make test-race

# Coverage HTML report
make test-cover
open coverage.html
```

---

## Monime Integration

Uses the ports/adapters pattern — swap out the payment gateway by implementing:

```go
type PaymentGateway interface {
    Disburse(ctx, DisbursementRequest) (*DisbursementResponse, error)
    Collect(ctx, CollectionRequest)   (*CollectionResponse, error)
    ValidateWebhook(payload, sig)      bool
    GetTransaction(ctx, reference)    (*MonimeTransaction, error)
}
```

Webhook validation uses HMAC-SHA256. Retry strategy: immediate → 1 hour → 24 hours → manual review.

---

## Ledger Accounts

| Account | Type |
|---|---|
| Loan Receivable | Asset |
| Partner Liability | Liability |
| Interest Income | Income |
| Penalty Income | Income |
| Commission Income | Income |
| Transaction Fee Income | Income |
| Monime Float | Asset |

Every disbursement, repayment, and payment fee creates balanced journal entries. Unbalanced entries are rejected.

---

## License

MIT © MarketPay Sierra Leone
=======
# MarketPay USSD Flow Service

A robust, production-ready USSD (Unstructured Supplementary Service Data) flow service for MarketPay built in Go. This service handles vendor registration, payments, transaction history, balance checks, and loan operations through USSD-based interactions.

## Project Structure

```
marketpay/
├── internal/
│   ├── ussd/
│   │   ├── types.go              # Flow types, constants, and interfaces
│   │   ├── service.go            # Main MarketPayFlowService
│   │   ├── validators.go         # Input validation logic
│   │   └── utils.go              # Utility functions
│   ├── store/
│   │   └── memory.go             # In-memory state store implementation
│   ├── transport/
│   │   └── http/
│   │       └── handler.go        # HTTP endpoint handlers
│   └── domain/
│       └── entities/             # Domain entities (for future extensions)
├── go.mod                        # Go module definition
└── README.md                     # This file
```

## Features

### Services Supported

1. **Register Vendor** - Collect vendor and market information
2. **Pay Vendor** - Send payment to registered vendors with SMS receipt option
3. **Transaction History** - View transaction history
4. **Check Balance** - Check account balance
5. **Loan Eligibility** - Check loan eligibility status
6. **Loan Application** - Apply for a loan
7. **Exit** - Exit the service

### Flow Characteristics

- **State Management**: Maintains session state across multiple USSD interactions
- **Input Validation**: Comprehensive validation for all user inputs
- **Error Handling**: Graceful error handling with user-friendly messages
- **Logging**: Structured logging using zerolog
- **Security**: Sensitive data masking in logs

## USSD Flow Pages

### Main Menu
- `mp_select_service` - Service selection menu

### Vendor Registration
- `mp_collect_vendor_name` - Vendor name input
- `mp_collect_market_name` - Market name input
- `mp_submit_vendor_registration` - Registration API call
- `mp_show_vendor_registration_result` - Result display

### Vendor Payment
- `mp_collect_payment_vendor_code` - Vendor code input
- `mp_collect_payment_amount` - Payment amount input
- `mp_confirm_payment_choice` - Payment confirmation
- `mp_submit_vendor_payment` - Payment API call
- `mp_show_payment_result` - Result display
- `mp_show_payment_cancelled` - Cancellation display

### Transaction History
- `mp_fetch_transaction_history` - Fetch API call
- `mp_show_transaction_history` - History display

### Balance Check
- `mp_fetch_balance` - Fetch API call
- `mp_show_balance` - Balance display

### Loan Eligibility
- `mp_fetch_loan_eligibility` - Eligibility check API call
- `mp_show_loan_eligibility` - Eligibility display

### Loan Application
- `mp_collect_loan_amount` - Loan amount input
- `mp_confirm_loan_application` - Application confirmation
- `mp_submit_loan_application` - Application API call
- `mp_show_loan_application_result` - Result display
- `mp_show_loan_application_cancelled` - Cancellation display

### Exit
- `mp_exit_service` - Exit message

## Usage

### Creating a Flow Service

```go
package main

import (
	"context"
	"marketpay/internal/store"
	"marketpay/internal/ussd"
)

func main() {
	// Create state store
	stateStore := store.NewInMemoryStateStore()
	
	// Create flow service
	flowService := ussd.NewMarketPayFlowService(stateStore)
	
	// Advance the flow
	input := ussd.AdvanceFlowInput{
		SessionID:   "session-123",
		CurrentPage: ussd.PageSelectService,
		Values:      map[string]string{},
	}
	
	ctx := context.Background()
	result, err := flowService.Advance(ctx, input)
	if err != nil {
		// Handle error
	}
	
	// Use result
	println(result.Message)
}
```

### HTTP Integration

```go
package main

import (
	"net/http"
	"marketpay/internal/store"
	"marketpay/internal/ussd"
	"marketpay/internal/transport/http"
)

func main() {
	stateStore := store.NewInMemoryStateStore()
	flowService := ussd.NewMarketPayFlowService(stateStore)
	handler := http.NewMarketPayUSSDHandler(flowService)
	
	http.HandleFunc("/api/ussd/advance", handler.Advance)
	http.HandleFunc("/health", handler.HealthCheck)
	
	http.ListenAndServe(":8080", nil)
}
```

### HTTP Request Example

```bash
curl -X POST http://localhost:8080/api/ussd/advance \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "session-123",
    "current_page": "mp_select_service",
    "values": {
      "selected_service": "pay_vendor"
    }
  }'
```

## Input Validation Rules

### Vendor Name
- Length: 2-50 characters
- Pattern: `^[A-Za-z0-9][A-Za-z0-9 ,.&-]{1,49}$`

### Market Name
- Length: 2-50 characters
- Pattern: `^[A-Za-z0-9][A-Za-z0-9 ,.&-]{1,49}$`

### Vendor Code
- Length: 7-12 characters
- Pattern: `^MP[0-9]{5,10}$` (e.g., MP12345)

### Payment Amount
- Min: 1 SLE
- Max: 10,000,000 SLE

### Loan Amount
- Min: 1 SLE
- Max: 5,000,000 SLE

## State Management

Session state is stored as key-value pairs. Common state keys include:

- `selected_service` - Selected service from main menu
- `registration_vendor_name` - Vendor name for registration
- `registration_market_name` - Market name for registration
- `payment_vendor_code` - Vendor code for payment
- `payment_amount` - Payment amount
- `payment_confirmed` - Payment confirmation status
- `loan_amount` - Loan application amount
- `loan_confirmed` - Loan application confirmation status

## Error Handling

The service provides graceful error handling:

- **Validation Errors**: User-friendly messages for invalid inputs
- **Missing Required Fields**: Prompts user to enter missing data
- **Session Errors**: Logs errors but continues with empty state
- **API Failures**: Gracefully handles external API failures (with TODO implementations)

## Logging

Uses structured logging with zerolog:

```go
log.Debug().Str("session_id", sessionID).Msg("processing page")
log.Error().Err(err).Msg("operation failed")
log.Info().Str("service", "payment").Msg("operation successful")
```

Sensitive data is automatically masked in logs:
- Names: `J***n`
- Emails: `j***n@example.com`
- Vendor Codes: `****2345`
- Phone Numbers: `****1234`

## Dependencies

- `github.com/rs/zerolog/log` - Structured logging

## Future Enhancements

- [ ] PostgreSQL state store implementation
- [ ] Redis state store implementation
- [ ] Integration with MarketPay APIs
- [ ] SMS gateway integration
- [ ] Authentication/Authorization layer
- [ ] Rate limiting
- [ ] Metrics and monitoring
- [ ] Unit tests
- [ ] Integration tests
- [ ] gRPC transport layer

## Configuration

State stores can be swapped by implementing the `StateStore` interface:

```go
type StateStore interface {
	Load(ctx context.Context, sessionID string) (map[string]string, error)
	Save(ctx context.Context, sessionID string, data map[string]string) error
}
```

## License

[Add your license here]
>>>>>>> 36381e2c5538acddd881027bebc31d8897a8e7cb
