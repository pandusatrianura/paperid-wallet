package integration_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	httphandler "paperid-wallet/internal/delivery/http"
	"paperid-wallet/internal/domain"
	"paperid-wallet/internal/repository/inmemory"
	"paperid-wallet/internal/usecase"
)

func setupTestRouter() http.Handler {
	seedUsers := map[string]*domain.User{
		"U001": {ID: "U001", Name: "Alice", Balance: 5000.00},
		"U002": {ID: "U002", Name: "Bob", Balance: 150.00},
	}
	repo := inmemory.NewUserRepository(seedUsers)
	uc := usecase.NewDisbursementUseCase(repo)
	handler := httphandler.NewDisburseHandler(uc)
	return httphandler.NewRouter(handler)
}

func TestDisbursement_EndToEnd_Scenarios(t *testing.T) {
	router := setupTestRouter()

	t.Run("Scenario 1: Successful disbursement deduction for Alice", func(t *testing.T) {
		payload := `{"user_id":"U001","amount":1000.00,"destination_bank":"BCA","destination_account":"1234567890"}`
		req := httptest.NewRequest(http.MethodPost, "/api/disburse", bytes.NewBufferString(payload))
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d. Body: %s", rec.Code, rec.Body.String())
		}

		var resp map[string]interface{}
		json.Unmarshal(rec.Body.Bytes(), &resp)

		if resp["transaction_id"] == nil || resp["transaction_id"] == "" {
			t.Errorf("expected valid transaction_id, got %v", resp["transaction_id"])
		}
		if resp["message"] != "Disbursement successful" {
			t.Errorf("expected message 'Disbursement successful', got '%v'", resp["message"])
		}
		if resp["remaining_balance"] != 4000.00 {
			t.Errorf("expected remaining_balance 4000.00, got %v", resp["remaining_balance"])
		}
	})

	t.Run("Scenario 2: Insufficient balance error for Bob", func(t *testing.T) {
		payload := `{"user_id":"U002","amount":200.00,"destination_bank":"Mandiri","destination_account":"9876543210"}`
		req := httptest.NewRequest(http.MethodPost, "/api/disburse", bytes.NewBufferString(payload))
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422 Unprocessable Entity, got %d", rec.Code)
		}

		var resp map[string]interface{}
		json.Unmarshal(rec.Body.Bytes(), &resp)

		if resp["error"] != "Insufficient balance" {
			t.Errorf("expected error 'Insufficient balance', got '%v'", resp["error"])
		}
	})

	t.Run("Scenario 3: User not found", func(t *testing.T) {
		payload := `{"user_id":"U999","amount":50.00,"destination_bank":"BRI","destination_account":"111222333"}`
		req := httptest.NewRequest(http.MethodPost, "/api/disburse", bytes.NewBufferString(payload))
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected 404 Not Found, got %d", rec.Code)
		}

		var resp map[string]interface{}
		json.Unmarshal(rec.Body.Bytes(), &resp)

		if resp["error"] != "User not found" {
			t.Errorf("expected error 'User not found', got '%v'", resp["error"])
		}
	})

	t.Run("Scenario 4: Negative or zero amount rejection", func(t *testing.T) {
		payload := `{"user_id":"U001","amount":0,"destination_bank":"BCA","destination_account":"123"}`
		req := httptest.NewRequest(http.MethodPost, "/api/disburse", bytes.NewBufferString(payload))
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 Bad Request, got %d", rec.Code)
		}

		var resp map[string]interface{}
		json.Unmarshal(rec.Body.Bytes(), &resp)

		if resp["error"] != "Amount must be greater than zero" {
			t.Errorf("expected error 'Amount must be greater than zero', got '%v'", resp["error"])
		}
	})

	t.Run("Scenario 5: Concurrent disbursements handle balance accurately", func(t *testing.T) {
		// New isolated router for concurrency test
		concurrencyRouter := setupTestRouter()
		// Alice starts with 5000.00
		// Let's run 50 requests of 100.00 each -> total 5000.00
		const requestsCount = 50
		var wg sync.WaitGroup
		wg.Add(requestsCount)

		for i := 0; i < requestsCount; i++ {
			go func() {
				defer wg.Done()
				payload := `{"user_id":"U001","amount":100.00,"destination_bank":"BCA","destination_account":"123"}`
				r := httptest.NewRequest(http.MethodPost, "/api/disburse", bytes.NewBufferString(payload))
				w := httptest.NewRecorder()
				concurrencyRouter.ServeHTTP(w, r)
				if w.Code != http.StatusOK {
					t.Errorf("unexpected status in concurrent request: %d", w.Code)
				}
			}()
		}

		wg.Wait()

		// Next request of 10.00 should now fail with Insufficient balance
		payload := `{"user_id":"U001","amount":10.00,"destination_bank":"BCA","destination_account":"123"}`
		req := httptest.NewRequest(http.MethodPost, "/api/disburse", bytes.NewBufferString(payload))
		rec := httptest.NewRecorder()
		concurrencyRouter.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422 after exhausting balance, got %d", rec.Code)
		}
	})
}
