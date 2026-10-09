package inmemory

import (
	"context"
	"sync"

	"paperid-wallet/internal/domain"
)

// UserRepository is an in-memory, thread-safe implementation of domain.UserRepository.
type UserRepository struct {
	mu    sync.RWMutex
	users map[string]*domain.User
}

// NewUserRepository initializes and returns a new in-memory UserRepository.
// If initialData is provided, it is cloned into the repository's internal state.
func NewUserRepository(initialData map[string]*domain.User) *UserRepository {
	repo := &UserRepository{
		users: make(map[string]*domain.User),
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

	// Defensive copy
	userCopy := *user
	return &userCopy, nil
}

// DeductBalance atomically verifies and deducts balance for a user under write lock.
func (r *UserRepository) DeductBalance(ctx context.Context, id string, amount float64) (*domain.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	user, exists := r.users[id]
	if !exists {
		return nil, domain.ErrUserNotFound
	}

	if err := user.Deduct(amount); err != nil {
		return nil, err
	}

	// Return copy of the updated state
	userCopy := *user
	return &userCopy, nil
}
