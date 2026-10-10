package inmemory

import (
	"context"
	"sync"

	"paperid-wallet/internal/domain"
)

// UserRepository provides a thread-safe, in-memory implementation of domain.UserRepository.
type UserRepository struct {
	mu           sync.RWMutex
	users        map[string]*domain.User
	transactions map[string][]*domain.Transaction
}

// NewUserRepository initializes and returns a new in-memory UserRepository.
func NewUserRepository(initialData map[string]*domain.User) *UserRepository {
	repo := &UserRepository{
		users:        make(map[string]*domain.User),
		transactions: make(map[string][]*domain.Transaction),
	}

	for k, v := range initialData {
		if v != nil {
			userCopy := *v
			repo.users[k] = &userCopy
		}
	}

	return repo
}

// GetByID retrieves a user by ID. Returns a copy of the user to prevent external mutation.
func (r *UserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, exists := r.users[id]
	if !exists {
		return nil, domain.ErrUserNotFound
	}

	userCopy := *user
	return &userCopy, nil
}

// DeductBalance atomically verifies balance, deducts it, and records the transaction log.
func (r *UserRepository) DeductBalance(ctx context.Context, id string, amount float64, tx *domain.Transaction) (*domain.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	user, exists := r.users[id]
	if !exists {
		return nil, domain.ErrUserNotFound
	}

	if err := user.Deduct(amount); err != nil {
		return nil, err
	}

	if tx != nil {
		txCopy := *tx
		r.transactions[id] = append(r.transactions[id], &txCopy)
	}

	userCopy := *user
	return &userCopy, nil
}

// GetTransactionsByUserID returns all recorded transactions for a user.
func (r *UserRepository) GetTransactionsByUserID(ctx context.Context, userID string) ([]*domain.Transaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	txs := r.transactions[userID]
	result := make([]*domain.Transaction, len(txs))
	for i, tx := range txs {
		txCopy := *tx
		result[i] = &txCopy
	}
	return result, nil
}
