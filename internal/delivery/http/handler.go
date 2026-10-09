package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"paperid-wallet/internal/domain"
	"paperid-wallet/internal/usecase"
)

// DisburseHandler handles HTTP requests for disbursement.
type DisburseHandler struct {
	disburseUseCase usecase.DisbursementUseCase
}

// NewDisburseHandler creates a new instance of DisburseHandler with injected use case.
func NewDisburseHandler(disburseUseCase usecase.DisbursementUseCase) *DisburseHandler {
	return &DisburseHandler{
		disburseUseCase: disburseUseCase,
	}
}

// ServeHTTP implements http.Handler interface for the disbursement endpoint.
func (h *DisburseHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// 1. Enforce POST method
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		_ = json.NewEncoder(w).Encode(DisburseResponse{Error: "Method not allowed"})
		return
	}

	// 2. Parse request body
	var req DisburseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(DisburseResponse{Error: "Invalid JSON payload"})
		return
	}

	// 3. Delegate to business use case
	input := usecase.DisburseInput{
		UserID:             req.UserID,
		Amount:             req.Amount,
		DestinationBank:    req.DestinationBank,
		DestinationAccount: req.DestinationAccount,
	}

	output, err := h.disburseUseCase.Execute(r.Context(), input)
	if err != nil {
		h.handleError(w, err)
		return
	}

	// 4. Return success response
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(DisburseResponse{
		Message: output.Message,
		Balance: output.RemainingBalance,
	})
}

// handleError maps domain errors to appropriate HTTP status codes and responses.
func (h *DisburseHandler) handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidAmount):
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(DisburseResponse{Error: "Amount must be greater than zero"})
	case errors.Is(err, domain.ErrInvalidUserID):
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(DisburseResponse{Error: "User ID cannot be empty"})
	case errors.Is(err, domain.ErrUserNotFound):
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(DisburseResponse{Error: "User not found"})
	case errors.Is(err, domain.ErrInsufficientBalance):
		w.WriteHeader(http.StatusUnprocessableEntity)
		_ = json.NewEncoder(w).Encode(DisburseResponse{Error: "Insufficient balance"})
	default:
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(DisburseResponse{Error: "Internal server error"})
	}
}
