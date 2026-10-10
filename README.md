# PaperID Wallet - Disbursement Service

A lightweight, thread-safe wallet disbursement service in Go. Built with standard library Go, layered architecture, and unit/integration test coverage.

---

## Overview & Architecture

The service is organized into decoupled layers to separate transport concerns, business rules, and storage access:

```
[ HTTP Transport ] -> [ Use Case ] -> [ Repository (Storage) ]
   (net/http)       (Orchestration)      (Thread-safe Memory)
                           |
                     [ Domain Model ]
```

- **Domain (`internal/domain`)**: Core business entities (`User`, `Transaction`), domain invariants, and sentinel errors. Contains pure business logic without external dependencies.
- **Use Case (`internal/usecase`)**: Coordinates disbursement operations, input validation, generates transaction records, and executes storage mutations.
- **Repository (`internal/repository/inmemory`)**: Thread-safe in-memory storage managing user balances and transaction audit logs.
- **Delivery (`internal/delivery/http`)**: HTTP handlers, router setup, request decoding, and status code mapping.
- **App Lifecycle (`internal/app`, `cmd/api`)**: Composition root for dependency injection, configuration, and graceful server shutdown.

---

## Engineering Decisions & Design Notes

1. **Zero External Dependencies**  
   The service uses only Go's standard library (`net/http`, `sync`, `context`, `encoding/json`, `crypto/rand`). This keeps compile times near-instant, the binary minimal, and eliminates supply-chain security risks.

2. **Atomic Balance Deduction & Transaction Ledger**  
   Disbursements generate an immutable `Transaction` record containing destination bank, destination account, timestamp, and unique transaction ID. To guarantee consistency without database transactions, `UserRepository` executes both balance deduction and transaction appending inside a single `sync.RWMutex` write-lock.

3. **Concurrency Safety & Defensive Copies**  
   To prevent race conditions during concurrent disbursements, `UserRepository` serializes write operations via `sync.RWMutex`. Methods return value copies of domain pointers (`*userCopy`, `*txCopy`) so internal state cannot be mutated outside repository critical sections.

4. **Financial Precision Considerations**  
   For the scope of this assignment, balances and amounts use `float64` to match the given JSON specification. In a production fintech wallet, currency amounts should be represented as integers in the smallest currency unit (e.g., cents or satoshis) or via an arbitrary-precision decimal type (such as `shopspring/decimal`) to avoid binary floating-point rounding inaccuracies.

5. **Graceful Shutdown**  
   The HTTP server listens for `SIGINT` (Ctrl+C) and `SIGTERM` signals, completing in-flight requests within a 5-second timeout window before exiting.

---

## Seed Data

The service pre-populates two accounts on startup:

| User ID | Name | Initial Balance |
| :--- | :--- | :--- |
| `U001` | Alice | `5000.00` |
| `U002` | Bob | `150.00` |

---

## Getting Started

### Run the Server

```bash
go run ./cmd/api
```

By default, the server runs on port `8080`. You can override this using the `PORT` environment variable:
```bash
PORT=9000 go run ./cmd/api
```

---

## API Testing & Scenarios

### Scenario 1: Successful Disbursement

**Linux / macOS (Bash):**
```bash
curl -X POST http://localhost:8080/api/disburse \
  -H "Content-Type: application/json" \
  -d '{"user_id": "U001", "amount": 1000.50, "destination_bank": "BCA", "destination_account": "1234567890"}'
```

**Windows PowerShell:**
```powershell
curl.exe -X POST http://localhost:8080/api/disburse `
  -H "Content-Type: application/json" `
  --data '{\"user_id\":\"U001\",\"amount\":1000.50,\"destination_bank\":\"BCA\",\"destination_account\":\"1234567890\"}'
```
*Or using PowerShell's native cmdlet:*
```powershell
Invoke-RestMethod -Uri "http://localhost:8080/api/disburse" -Method Post -ContentType "application/json" -Body '{"user_id":"U001","amount":1000.50,"destination_bank":"BCA","destination_account":"1234567890"}'
```

**Response (HTTP 200):**
```json
{
  "transaction_id": "TRX-1728551234-a1b2c3d4",
  "message": "Disbursement successful",
  "remaining_balance": 3999.5
}
```

---

### Scenario 2: Insufficient Balance

**Linux / macOS (Bash):**
```bash
curl -X POST http://localhost:8080/api/disburse \
  -H "Content-Type: application/json" \
  -d '{"user_id": "U002", "amount": 500.00, "destination_bank": "BNI", "destination_account": "0987654321"}'
```

**Windows PowerShell:**
```powershell
curl.exe -X POST http://localhost:8080/api/disburse `
  -H "Content-Type: application/json" `
  --data '{\"user_id\":\"U002\",\"amount\":500.00,\"destination_bank\":\"BNI\",\"destination_account\":\"0987654321\"}'
```

**Response (HTTP 422):**
```json
{
  "error": "Insufficient balance"
}
```

---

### Scenario 3: User Not Found

**Linux / macOS (Bash):**
```bash
curl -X POST http://localhost:8080/api/disburse \
  -H "Content-Type: application/json" \
  -d '{"user_id": "U999", "amount": 50.00}'
```

**Windows PowerShell:**
```powershell
curl.exe -X POST http://localhost:8080/api/disburse `
  -H "Content-Type: application/json" `
  --data '{\"user_id\":\"U999\",\"amount\":50.00}'
```

**Response (HTTP 404):**
```json
{
  "error": "User not found"
}
```

---

### Scenario 4: Invalid Amount (Zero or Negative)

**Linux / macOS (Bash):**
```bash
curl -X POST http://localhost:8080/api/disburse \
  -H "Content-Type: application/json" \
  -d '{"user_id": "U001", "amount": 0}'
```

**Windows PowerShell:**
```powershell
curl.exe -X POST http://localhost:8080/api/disburse `
  -H "Content-Type: application/json" `
  --data '{\"user_id\":\"U001\",\"amount\":0}'
```

**Response (HTTP 400):**
```json
{
  "error": "Amount must be greater than zero"
}
```

---

## Running Tests

All unit, repository, and integration tests can be run using the standard Go test tool:

```bash
# Run all tests
go test -v ./...

# Run layer-specific tests
go test -v ./internal/domain/...
go test -v ./internal/repository/inmemory/...
go test -v ./internal/usecase/...
go test -v ./internal/delivery/http/...
go test -v ./test/integration/...
```

The test suite includes scenario validations and a concurrent execution test (`50 concurrent requests`) to verify balance consistency under load without data races.
