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

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		_ = json.NewEncoder(w).Encode(DisburseResponse{Error: "Method not allowed"})
		return
	}

	var req DisburseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(DisburseResponse{Error: "Invalid JSON payload"})
		return
	}

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

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(DisburseResponse{
		TransactionID: output.TransactionID,
		Message:       output.Message,
		Balance:       output.RemainingBalance,
	})
}

// handleError maps domain errors to appropriate HTTP status codes and responses.
func (h *DisburseHandler) handleError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	msg := "Internal server error"

	switch {
	case errors.Is(err, domain.ErrInvalidAmount), errors.Is(err, domain.ErrInvalidUserID):
		status = http.StatusBadRequest
		msg = err.Error()
	case errors.Is(err, domain.ErrUserNotFound):
		status = http.StatusNotFound
		msg = err.Error()
	case errors.Is(err, domain.ErrInsufficientBalance):
		status = http.StatusUnprocessableEntity
		msg = err.Error()
	}

	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(DisburseResponse{Error: msg})
}
