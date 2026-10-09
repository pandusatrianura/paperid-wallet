# PaperID Wallet - Disbursement Service

A robust, thread-safe disbursement API built in Go, designed with **Clean Architecture**, **SOLID Principles**, and **Test-Driven Development (TDD)**.

---

## 📌 Architecture & Design Principles

The service strictly adheres to **Clean Architecture** by separating concerns into independent layers, ensuring high testability, maintainability, and loose coupling.

```
+-------------------------------------------------------------+
|                     Frameworks & Drivers                    |
|             (HTTP Delivery, Server, CLI Entrypoints)        |
|  +-------------------------------------------------------+  |
|  |                   Interface Adapters                  |  |
|  |           (In-Memory Repositories, HTTP Handlers)     |  |
|  |  +-------------------------------------------------+  |  |
|  |  |             Application Business Rules          |  |  |
|  |  |               (Disbursement Use Case)           |  |  |
|  |  |  +-------------------------------------------+  |  |  |
|  |  |  |           Enterprise Business Rules       |  |  |  |
|  |  |  |              (Domain Entities, Errors)    |  |  |  |
|  |  |  +-------------------------------------------+  |  |  |
|  |  +-------------------------------------------------+  |  |
|  +-------------------------------------------------------+  |
+-------------------------------------------------------------+
```

### SOLID Principles in Action

- **S - Single Responsibility Principle (SRP):**
  - `domain.User`: Encapsulates entity data and deduction invariant logic.
  - `inmemory.UserRepository`: Handles thread-safe state persistence and concurrency control.
  - `usecase.DisbursementUseCase`: Orchestrates the business workflow.
  - `http.DisburseHandler`: Focuses exclusively on HTTP protocol translation and status mapping.
- **O - Open/Closed Principle (OCP):**
  - Dependent components interact via interfaces (`UserRepository`, `DisbursementUseCase`). Swapping the storage layer (e.g., PostgreSQL, Redis) or adding a 3rd-party banking gateway does not require changing core use case logic.
- **L - Liskov Substitution Principle (LSP):**
  - Any implementation of `domain.UserRepository` (whether in-memory storage or test mocks) can substitute one another without affecting correctness.
- **I - Interface Segregation Principle (ISP):**
  - Interfaces are concise and client-driven. `UserRepository` specifies only methods needed by application interactors (`GetByID`, `DeductBalance`).
- **D - Dependency Inversion Principle (DIP):**
  - High-level business logic does not depend on low-level database or HTTP details; both depend on domain-level abstractions.

---

## 📂 Project Directory Structure

```text
paperid-wallet/
├── cmd/
│   └── api/
│       └── main.go                             # Standard application entry point
├── internal/
│   ├── app/
│   │   └── app.go                              # Dependency injection & server lifecycle
│   ├── domain/                                 # Domain entities & interfaces
│   │   ├── errors.go                           # Domain sentinel errors
│   │   ├── repository.go                       # Repository interface contracts
│   │   ├── user.go                             # User entity & business invariants
│   │   └── user_test.go                        # Domain unit tests
│   ├── repository/                             # Data access adapters
│   │   └── inmemory/
│   │       ├── user_repository.go              # Thread-safe in-memory repository
│   │       └── user_repository_test.go         # Concurrency & repository tests
│   ├── usecase/                                # Application business logic
│   │   ├── disbursement_usecase.go             # Disbursement use case implementation
│   │   └── disbursement_usecase_test.go        # Use case unit tests with mocks
│   └── delivery/                               # Transport layer (HTTP)
│       └── http/
│           ├── handler.go                      # HTTP Disburse handler
│           ├── handler_test.go                 # HTTP tests with httptest
│           ├── response.go                     # Request & Response DTOs
│           └── router.go                       # HTTP route registration
├── test/
│   └── integration/
│       └── disbursement_test.go                # End-to-end integration tests
├── go.mod                                      # Go module definition
└── README.md                                   # Project documentation
```

---

## 🧪 Initial In-Memory Seed Data

The server seeds two default users for testing:

| User ID | Name | Initial Balance |
| :--- | :--- | :--- |
| `U001` | Alice | `5000.00` |
| `U002` | Bob | `150.00` |

---

## 🚀 How to Run and Test

Run the server from your terminal:

```bash
go run ./cmd/api
```

Open a new terminal and use the following curl commands to test the different scenarios:

### Scenario 1: Successful Disbursement

```bash
curl -X POST http://localhost:8080/api/disburse \
-H "Content-Type: application/json" \
-d '{"user_id": "U001", "amount": 1000.50, "destination_bank": "BCA", "destination_account": "1234567890"}'
```

**Response (HTTP 200):**
```json
{"message":"Disbursement successful","remaining_balance":3999.5}
```

---

### Scenario 2: Insufficient Balance

```bash
curl -X POST http://localhost:8080/api/disburse \
-H "Content-Type: application/json" \
-d '{"user_id": "U002", "amount": 500.00, "destination_bank": "BNI", "destination_account": "0987654321"}'
```

**Response (HTTP 422):**
```json
{"error":"Insufficient balance"}
```

---

### Scenario 3: Invalid User

```bash
curl -X POST http://localhost:8080/api/disburse \
-H "Content-Type: application/json" \
-d '{"user_id": "U999", "amount": 50.00}'
```

**Response (HTTP 404):**
```json
{"error":"User not found"}
```

---

### Scenario 4: Invalid Amount (Zero or Negative)

```bash
curl -X POST http://localhost:8080/api/disburse \
-H "Content-Type: application/json" \
-d '{"user_id": "U001", "amount": 0}'
```

**Response (HTTP 400):**
```json
{"error":"Amount must be greater than zero"}
```

---

## 🧪 Running Automated Tests (TDD)

The project includes complete unit and end-to-end integration tests across all layers.

### Run All Tests
```bash
go test -v ./...
```

### Run Layer-Specific Tests
```bash
# Domain tests
go test -v ./internal/domain/...

# Repository tests (includes concurrency tests)
go test -v ./internal/repository/inmemory/...

# Use Case tests
go test -v ./internal/usecase/...

# HTTP Delivery tests
go test -v ./internal/delivery/http/...

# End-to-End Integration tests
go test -v ./test/integration/...
```

---

## ⚙️ Configuration & Features

- **Port Configuration:** Set `PORT` environment variable (e.g. `PORT=9000 go run ./cmd/api`). Defaults to `8080`.
- **Graceful Shutdown:** Handles `SIGINT` (Ctrl+C) and `SIGTERM` signals cleanly with a 5-second timeout window.
- **Zero External Dependencies:** Built entirely with the Go standard library (`net/http`, `sync`, `context`, `encoding/json`).
