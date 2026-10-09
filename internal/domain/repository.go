package domain

import "context"

// UserRepository defines the persistence interface for User entities.
// Adheres to Dependency Inversion Principle (DIP) and Interface Segregation (ISP).
type UserRepository interface {
	// GetByID retrieves a user by ID. Returns ErrUserNotFound if not found.
	GetByID(ctx context.Context, id string) (*User, error)

	// DeductBalance atomically validates balance and deducts the amount for the specified user.
	// Returns updated User and nil, or error (ErrUserNotFound, ErrInsufficientBalance, ErrInvalidAmount).
	DeductBalance(ctx context.Context, id string, amount float64) (*User, error)
}
