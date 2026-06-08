# MarketPay USSD API Documentation

## Overview

The MarketPay USSD Flow Service provides HTTP REST endpoints to manage USSD (Unstructured Supplementary Service Data) flows for MarketPay operations. The service uses stateful flow management to track user interactions across multiple USSD sessions.

## Base URL

```
http://localhost:8080
```

## Endpoints

### 1. Advance USSD Flow

**Endpoint:** `POST /api/ussd/advance`

**Description:** Advances the USSD flow to the next page based on the current page and user input.

**Request Headers:**
```
Content-Type: application/json
```

**Request Body:**
```json
{
  "session_id": "string (required)",
  "current_page": "string (required)",
  "values": {
    "key1": "value1",
    "key2": "value2"
  }
}
```

**Response:**
```json
{
  "action": "string (navigate | stop)",
  "next_page": "string",
  "message": "string",
  "data": {
    "key1": "value1",
    "key2": "value2"
  },
  "error": "string (optional)"
}
```

**Status Codes:**
- `200 OK` - Flow advanced successfully
- `400 Bad Request` - Invalid request payload
- `500 Internal Server Error` - Server error during flow processing

**Example Request:**
```bash
curl -X POST http://localhost:8080/api/ussd/advance \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "user-12345",
    "current_page": "mp_select_service",
    "values": {
      "selected_service": "pay_vendor"
    }
  }'
```

**Example Response:**
```json
{
  "action": "navigate",
  "next_page": "mp_collect_payment_vendor_code",
  "message": "Enter vendor code\nUse a code like MP12345",
  "data": {
    "selected_service": "pay_vendor"
  }
}
```

### 2. Health Check

**Endpoint:** `GET /health`

**Description:** Health check endpoint to verify service availability.

**Response:**
```json
{
  "status": "ok",
  "service": "marketpay-ussd"
}
```

**Status Codes:**
- `200 OK` - Service is healthy

**Example Request:**
```bash
curl http://localhost:8080/health
```

**Example Response:**
```json
{
  "status": "ok",
  "service": "marketpay-ussd"
}
```

## Page IDs and Flow Structure

### Main Service Selection
- **Page ID:** `mp_select_service`
- **Description:** Main menu for selecting services
- **Values Expected:** `selected_service` with value from options

### Vendor Registration Flow
1. `mp_select_service` (select "register_vendor")
2. `mp_collect_vendor_name` (enter vendor name)
3. `mp_collect_market_name` (enter market name)
4. `mp_submit_vendor_registration` (submit registration)
5. `mp_show_vendor_registration_result` (display result)

### Vendor Payment Flow
1. `mp_select_service` (select "pay_vendor")
2. `mp_collect_payment_vendor_code` (enter vendor code)
3. `mp_collect_payment_amount` (enter amount)
4. `mp_confirm_payment_choice` (confirm payment)
5. `mp_submit_vendor_payment` (submit payment)
6. `mp_show_payment_result` or `mp_show_payment_cancelled` (display result)

### Transaction History Flow
1. `mp_select_service` (select "transaction_history")
2. `mp_fetch_transaction_history` (fetch data)
3. `mp_show_transaction_history` (display result)

### Balance Check Flow
1. `mp_select_service` (select "balance_check")
2. `mp_fetch_balance` (fetch data)
3. `mp_show_balance` (display result)

### Loan Eligibility Flow
1. `mp_select_service` (select "loan_eligibility")
2. `mp_fetch_loan_eligibility` (fetch data)
3. `mp_show_loan_eligibility` (display result)

### Loan Application Flow
1. `mp_select_service` (select "loan_application")
2. `mp_collect_loan_amount` (enter amount)
3. `mp_confirm_loan_application` (confirm application)
4. `mp_submit_loan_application` (submit application)
5. `mp_show_loan_application_result` or `mp_show_loan_application_cancelled` (display result)

### Exit Flow
1. `mp_select_service` (select "exit")
2. `mp_exit_service` (exit message)

## Input Validation Rules

### Vendor/Market Name
- **Min Length:** 2 characters
- **Max Length:** 50 characters
- **Pattern:** `^[A-Za-z0-9][A-Za-z0-9 ,.&-]{1,49}$`
- **Examples:** "ABC Store", "Tech & Solutions", "Market-01"

### Vendor Code
- **Min Length:** 7 characters
- **Max Length:** 12 characters
- **Pattern:** `^MP[0-9]{5,10}$`
- **Examples:** "MP12345", "MP0123456789"

### Payment Amount
- **Min Value:** 1 SLE
- **Max Value:** 10,000,000 SLE
- **Type:** Integer

### Loan Amount
- **Min Value:** 1 SLE
- **Max Value:** 5,000,000 SLE
- **Type:** Integer

## Error Handling

### Validation Errors
When user input fails validation, the service returns an error message:

```json
{
  "action": "stop",
  "next_page": "mp_collect_vendor_name",
  "message": "Enter a valid name",
  "data": {
    "registration_vendor_name": "A"
  }
}
```

### Missing Required Fields
If a required field is missing, the service prompts for re-entry:

```json
{
  "action": "navigate",
  "next_page": "mp_collect_payment_vendor_code",
  "message": "Enter vendor code\nUse a code like MP12345",
  "data": {}
}
```

### Server Errors
If an unexpected error occurs:

```json
{
  "action": "stop",
  "message": "Failed to process request",
  "error": "Internal server error"
}
```

## Session Management

### Session ID
- Must be unique per user session
- Should be generated by the client (e.g., phone number + timestamp)
- Session state is persisted across requests
- State is maintained in memory by default (or custom store)

### Session Data
Session data is stored and retrieved automatically:
- Persists across multiple requests
- Contains all user inputs and flow context
- Returned in the `data` field of responses

## Action Types

### Navigate
Advances to the next page in the flow:
```json
{
  "action": "navigate",
  "next_page": "mp_collect_payment_amount",
  "message": "Enter amount in SLE"
}
```

### Stop
Terminates the current flow and ends the session:
```json
{
  "action": "stop",
  "message": "Thanks for using MarketPay.",
  "data": {}
}
```

## Best Practices

1. **Session IDs:** Use phone number or unique identifier
2. **Error Handling:** Always check the `error` field in responses
3. **Timeouts:** Implement client-side timeouts (e.g., 30 seconds)
4. **Retry Logic:** Retry failed requests with exponential backoff
5. **Logging:** Log all USSD interactions for audit trails
6. **Security:** Do not log sensitive data (validate server-side)

## Rate Limiting

Currently not implemented. Future versions may include rate limiting per session or IP.

## Examples

### Example 1: Complete Payment Flow

```bash
# Step 1: Select service
curl -X POST http://localhost:8080/api/ussd/advance \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "user-001",
    "current_page": "mp_select_service",
    "values": {"selected_service": "pay_vendor"}
  }'

# Step 2: Enter vendor code
curl -X POST http://localhost:8080/api/ussd/advance \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "user-001",
    "current_page": "mp_collect_payment_vendor_code",
    "values": {"payment_vendor_code": "MP123456"}
  }'

# Step 3: Enter amount
curl -X POST http://localhost:8080/api/ussd/advance \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "user-001",
    "current_page": "mp_collect_payment_amount",
    "values": {"payment_amount": "100000"}
  }'

# Step 4: Confirm payment
curl -X POST http://localhost:8080/api/ussd/advance \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "user-001",
    "current_page": "mp_confirm_payment_choice",
    "values": {"payment_confirmed": "pay_sms"}
  }'

# Step 5: Submit payment
curl -X POST http://localhost:8080/api/ussd/advance \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "user-001",
    "current_page": "mp_submit_vendor_payment",
    "values": {}
  }'
```

### Example 2: Loan Application Flow

```bash
# Step 1: Select loan application service
curl -X POST http://localhost:8080/api/ussd/advance \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "user-002",
    "current_page": "mp_select_service",
    "values": {"selected_service": "loan_application"}
  }'

# Step 2: Enter loan amount
curl -X POST http://localhost:8080/api/ussd/advance \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "user-002",
    "current_page": "mp_collect_loan_amount",
    "values": {"loan_amount": "500000"}
  }'

# Step 3: Confirm application
curl -X POST http://localhost:8080/api/ussd/advance \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "user-002",
    "current_page": "mp_confirm_loan_application",
    "values": {"loan_confirmed": "submit"}
  }'
```

## Support

For issues or feature requests, please refer to the project documentation or contact the development team.
