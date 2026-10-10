package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	httphandler "paperid-wallet/internal/delivery/http"
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
					TransactionID:    "TRX-TEST-123",
					Message:          "Disbursement successful",
					RemainingBalance: 4500.00,
				}, nil
			},
		}

		handler := httphandler.NewDisburseHandler(mockUC)

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

		if resp["transaction_id"] != "TRX-TEST-123" {
			t.Errorf("expected transaction_id 'TRX-TEST-123', got '%v'", resp["transaction_id"])
		}
		if resp["message"] != "Disbursement successful" {
			t.Errorf("expected message 'Disbursement successful', got '%v'", resp["message"])
		}
		if resp["remaining_balance"] != 4500.00 {
			t.Errorf("expected remaining_balance 4500.00, got %v", resp["remaining_balance"])
		}
	})
}

func TestDisburseHandler_ErrorScenarios(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		body           string
		mockUC         usecase.DisbursementUseCase
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "method not allowed",
			method:         http.MethodGet,
			body:           "",
			mockUC:         &mockDisbursementUseCase{},
			expectedStatus: http.StatusMethodNotAllowed,
			expectedError:  "Method not allowed",
		},
		{
			name:           "malformed json payload",
			method:         http.MethodPost,
			body:           "{invalid json",
			mockUC:         &mockDisbursementUseCase{},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid JSON payload",
		},
		{
			name:   "invalid amount error from usecase",
			method: http.MethodPost,
			body:   `{"user_id":"U001","amount":-50.00,"destination_bank":"BCA","destination_account":"123"}`,
			mockUC: &mockDisbursementUseCase{
				executeFunc: func(ctx context.Context, input usecase.DisburseInput) (*usecase.DisburseOutput, error) {
					return nil, domain.ErrInvalidAmount
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Amount must be greater than zero",
		},
		{
			name:   "user not found",
			method: http.MethodPost,
			body:   `{"user_id":"UNKNOWN","amount":100.00,"destination_bank":"BCA","destination_account":"123"}`,
			mockUC: &mockDisbursementUseCase{
				executeFunc: func(ctx context.Context, input usecase.DisburseInput) (*usecase.DisburseOutput, error) {
					return nil, domain.ErrUserNotFound
				},
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "User not found",
		},
		{
			name:   "insufficient balance",
			method: http.MethodPost,
			body:   `{"user_id":"U002","amount":500.00,"destination_bank":"BCA","destination_account":"123"}`,
			mockUC: &mockDisbursementUseCase{
				executeFunc: func(ctx context.Context, input usecase.DisburseInput) (*usecase.DisburseOutput, error) {
					return nil, domain.ErrInsufficientBalance
				},
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedError:  "Insufficient balance",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			handler := httphandler.NewDisburseHandler(tc.mockUC)
			req := httptest.NewRequest(tc.method, "/api/disburse", bytes.NewBufferString(tc.body))
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != tc.expectedStatus {
				t.Fatalf("expected status %d, got %d", tc.expectedStatus, rec.Code)
			}

			var resp map[string]interface{}
			if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}
			if resp["error"] != tc.expectedError {
				t.Errorf("expected error '%s', got '%v'", tc.expectedError, resp["error"])
			}
		})
	}
}
