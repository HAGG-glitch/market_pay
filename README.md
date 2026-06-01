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
