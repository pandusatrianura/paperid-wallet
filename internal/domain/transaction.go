package domain

import "time"

type TransactionStatus string

const (
	TransactionStatusSuccess TransactionStatus = "SUCCESS"
	TransactionStatusFailed  TransactionStatus = "FAILED"
)

// Transaction records a financial ledger event for an account disbursement.
type Transaction struct {
	ID                 string
	UserID             string
	Amount             float64
	DestinationBank    string
	DestinationAccount string
	Status             TransactionStatus
	CreatedAt          time.Time
}
