package domain

import "context"

// UserRepository defines storage operations for users and transaction records.
type UserRepository interface {
	GetByID(ctx context.Context, id string) (*User, error)
	DeductBalance(ctx context.Context, id string, amount float64, tx *Transaction) (*User, error)
	GetTransactionsByUserID(ctx context.Context, userID string) ([]*Transaction, error)
}
