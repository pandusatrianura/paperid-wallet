package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	deliveryHttp "paperid-wallet/internal/delivery/http"
	"paperid-wallet/internal/domain"
	"paperid-wallet/internal/usecase"
)

// mockDisbursementUseCase is a test double for usecase.DisbursementUseCase
type mockDisbursementUseCase struct {
	executeFunc func(ctx context.Context, input usecase.DisburseInput) (*usecase.DisburseOutput, error)
}

func (m *mockDisbursementUseCase) Execute(ctx context.Context, input usecase.DisburseInput) (*usecase.DisburseOutput, error) {
	if m.executeFunc != nil {
		return m.executeFunc(ctx, input)
	}
	return nil, errors.New("unimplemented")
}

func TestDisburseHandler(t *testing.T) {
	t.Run("returns 200 and success response when valid request is provided", func(t *testing.T) {
		mockUC := &mockDisbursementUseCase{
			executeFunc: func(ctx context.Context, input usecase.DisburseInput) (*usecase.DisburseOutput, error) {
				return &usecase.DisburseOutput{
					Message:          "Disbursement successful",
					RemainingBalance: 4500.00,
				}, nil
			},
		}

		handler := deliveryHttp.NewDisburseHandler(mockUC)

		body := []byte(`{"user_id":"U001","amount":500.00,"destination_bank":"BCA","destination_account":"123"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/disburse", bytes.NewBuffer(body))
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}

		var resp map[string]interface{}
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp["message"] != "Disbursement successful" {
			t.Errorf("expected message 'Disbursement successful', got '%v'", resp["message"])
		}
		if resp["remaining_balance"] != 4500.00 {
			t.Errorf("expected remaining_balance 4500.00, got %v", resp["remaining_balance"])
		}
	})

	t.Run("returns 405 Method Not Allowed for non-POST requests", func(t *testing.T) {
		mockUC := &mockDisbursementUseCase{}
		handler := deliveryHttp.NewDisburseHandler(mockUC)

		req := httptest.NewRequest(http.MethodGet, "/api/disburse", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusMethodNotAllowed {
			t.Fatalf("expected status 405, got %d", rec.Code)
		}

		var resp map[string]interface{}
		json.Unmarshal(rec.Body.Bytes(), &resp)
		if resp["error"] != "Method not allowed" {
			t.Errorf("expected error 'Method not allowed', got '%v'", resp["error"])
		}
	})

	t.Run("returns 400 Bad Request for malformed JSON", func(t *testing.T) {
		mockUC := &mockDisbursementUseCase{}
		handler := deliveryHttp.NewDisburseHandler(mockUC)

		req := httptest.NewRequest(http.MethodPost, "/api/disburse", bytes.NewBufferString("{invalid json"))
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400, got %d", rec.Code)
		}

		var resp map[string]interface{}
		json.Unmarshal(rec.Body.Bytes(), &resp)
		if resp["error"] != "Invalid JSON payload" {
			t.Errorf("expected error 'Invalid JSON payload', got '%v'", resp["error"])
		}
	})

	t.Run("returns 400 Bad Request when amount is zero or negative", func(t *testing.T) {
		mockUC := &mockDisbursementUseCase{
			executeFunc: func(ctx context.Context, input usecase.DisburseInput) (*usecase.DisburseOutput, error) {
				return nil, domain.ErrInvalidAmount
			},
		}
		handler := deliveryHttp.NewDisburseHandler(mockUC)

		body := []byte(`{"user_id":"U001","amount":-50.00,"destination_bank":"BCA","destination_account":"123"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/disburse", bytes.NewBuffer(body))
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400, got %d", rec.Code)
		}

		var resp map[string]interface{}
		json.Unmarshal(rec.Body.Bytes(), &resp)
		if resp["error"] != "Amount must be greater than zero" {
			t.Errorf("expected error 'Amount must be greater than zero', got '%v'", resp["error"])
		}
	})

	t.Run("returns 404 Not Found when user is not found", func(t *testing.T) {
		mockUC := &mockDisbursementUseCase{
			executeFunc: func(ctx context.Context, input usecase.DisburseInput) (*usecase.DisburseOutput, error) {
				return nil, domain.ErrUserNotFound
			},
		}
		handler := deliveryHttp.NewDisburseHandler(mockUC)

		body := []byte(`{"user_id":"UNKNOWN","amount":100.00,"destination_bank":"BCA","destination_account":"123"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/disburse", bytes.NewBuffer(body))
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected status 404, got %d", rec.Code)
		}

		var resp map[string]interface{}
		json.Unmarshal(rec.Body.Bytes(), &resp)
		if resp["error"] != "User not found" {
			t.Errorf("expected error 'User not found', got '%v'", resp["error"])
		}
	})

	t.Run("returns 422 Unprocessable Entity when balance is insufficient", func(t *testing.T) {
		mockUC := &mockDisbursementUseCase{
			executeFunc: func(ctx context.Context, input usecase.DisburseInput) (*usecase.DisburseOutput, error) {
				return nil, domain.ErrInsufficientBalance
			},
		}
		handler := deliveryHttp.NewDisburseHandler(mockUC)

		body := []byte(`{"user_id":"U002","amount":500.00,"destination_bank":"BCA","destination_account":"123"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/disburse", bytes.NewBuffer(body))
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnprocessableEntity {
			t.Fatalf("expected status 422, got %d", rec.Code)
		}

		var resp map[string]interface{}
		json.Unmarshal(rec.Body.Bytes(), &resp)
		if resp["error"] != "Insufficient balance" {
			t.Errorf("expected error 'Insufficient balance', got '%v'", resp["error"])
		}
	})
}
