package http

// DisburseRequest defines the incoming JSON payload for disbursement.
type DisburseRequest struct {
	UserID             string  `json:"user_id"`
	Amount             float64 `json:"amount"`
	DestinationBank    string  `json:"destination_bank"`
	DestinationAccount string  `json:"destination_account"`
}

// DisburseResponse defines the outgoing JSON payload.
type DisburseResponse struct {
	TransactionID string  `json:"transaction_id,omitempty"`
	Message       string  `json:"message,omitempty"`
	Balance       float64 `json:"remaining_balance,omitempty"`
	Error         string  `json:"error,omitempty"`
}
