package inmemory_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"paperid-wallet/internal/domain"
	"paperid-wallet/internal/repository/inmemory"
)

func TestUserRepository_GetByID(t *testing.T) {
	ctx := context.Background()
	repo := inmemory.NewUserRepository(map[string]*domain.User{
		"U001": {ID: "U001", Name: "Alice", Balance: 5000.00},
	})

	tests := []struct {
		name        string
		userID      string
		expectedErr error
		checkUser   func(t *testing.T, u *domain.User)
	}{
		{
			name:        "existing user",
			userID:      "U001",
			expectedErr: nil,
			checkUser: func(t *testing.T, u *domain.User) {
				if u == nil || u.ID != "U001" || u.Name != "Alice" || u.Balance != 5000.00 {
					t.Errorf("unexpected user returned: %+v", u)
				}
			},
		},
		{
			name:        "non-existent user",
			userID:      "UNKNOWN",
			expectedErr: domain.ErrUserNotFound,
			checkUser: func(t *testing.T, u *domain.User) {
				if u != nil {
					t.Errorf("expected nil user, got: %+v", u)
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			user, err := repo.GetByID(ctx, tc.userID)
			if !errors.Is(err, tc.expectedErr) {
				t.Fatalf("expected error %v, got %v", tc.expectedErr, err)
			}
			if tc.checkUser != nil {
				tc.checkUser(t, user)
			}
		})
	}
}

func TestUserRepository_DeductBalance_SuccessAndAudit(t *testing.T) {
	ctx := context.Background()
	repo := inmemory.NewUserRepository(map[string]*domain.User{
		"U001": {ID: "U001", Name: "Alice", Balance: 1000.00},
	})

	tx := &domain.Transaction{
		ID:                 "TRX-001",
		UserID:             "U001",
		Amount:             300.00,
		DestinationBank:    "BCA",
		DestinationAccount: "123456",
		Status:             domain.TransactionStatusSuccess,
	}

	user, err := repo.DeductBalance(ctx, "U001", 300.00, tx)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if user.Balance != 700.00 {
		t.Errorf("expected remaining balance 700.00, got: %.2f", user.Balance)
	}

	persisted, _ := repo.GetByID(ctx, "U001")
	if persisted.Balance != 700.00 {
		t.Errorf("expected persisted balance 700.00, got: %.2f", persisted.Balance)
	}

	history, err := repo.GetTransactionsByUserID(ctx, "U001")
	if err != nil {
		t.Fatalf("expected no error fetching transactions: %v", err)
	}
	if len(history) != 1 || history[0].ID != "TRX-001" {
		t.Errorf("expected 1 recorded transaction TRX-001, got: %+v", history)
	}
}

func TestUserRepository_DeductBalance_Errors(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name         string
		initialData  map[string]*domain.User
		userID       string
		amount       float64
		expectedErr  error
		verifyNoLeak func(t *testing.T, repo *inmemory.UserRepository)
	}{
		{
			name:        "user not found",
			initialData: nil,
			userID:      "NON_EXISTENT",
			amount:      100.00,
			expectedErr: domain.ErrUserNotFound,
		},
		{
			name: "zero amount",
			initialData: map[string]*domain.User{
				"U001": {ID: "U001", Name: "Alice", Balance: 500.00},
			},
			userID:      "U001",
			amount:      0,
			expectedErr: domain.ErrInvalidAmount,
		},
		{
			name: "negative amount",
			initialData: map[string]*domain.User{
				"U001": {ID: "U001", Name: "Alice", Balance: 500.00},
			},
			userID:      "U001",
			amount:      -50.00,
			expectedErr: domain.ErrInvalidAmount,
		},
		{
			name: "insufficient balance preserves original state",
			initialData: map[string]*domain.User{
				"U002": {ID: "U002", Name: "Bob", Balance: 150.00},
			},
			userID:      "U002",
			amount:      200.00,
			expectedErr: domain.ErrInsufficientBalance,
			verifyNoLeak: func(t *testing.T, repo *inmemory.UserRepository) {
				persisted, _ := repo.GetByID(ctx, "U002")
				if persisted.Balance != 150.00 {
					t.Errorf("expected balance to remain 150.00, got %.2f", persisted.Balance)
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := inmemory.NewUserRepository(tc.initialData)
			_, err := repo.DeductBalance(ctx, tc.userID, tc.amount, nil)
			if !errors.Is(err, tc.expectedErr) {
				t.Fatalf("expected error %v, got %v", tc.expectedErr, err)
			}
			if tc.verifyNoLeak != nil {
				tc.verifyNoLeak(t, repo)
			}
		})
	}
}

func TestUserRepository_DeductBalance_Concurrency(t *testing.T) {
	ctx := context.Background()
	const initialBalance = 1000.00
	const deductionAmount = 10.00
	const goroutinesCount = 100 // 100 * 10 = 1000 total deduction

	repo := inmemory.NewUserRepository(map[string]*domain.User{
		"U_CONCURRENT": {ID: "U_CONCURRENT", Name: "Concurrent User", Balance: initialBalance},
	})

	var wg sync.WaitGroup
	wg.Add(goroutinesCount)

	for i := 0; i < goroutinesCount; i++ {
		go func() {
			defer wg.Done()
			_, err := repo.DeductBalance(ctx, "U_CONCURRENT", deductionAmount, nil)
			if err != nil {
				t.Errorf("unexpected error in concurrent deduction: %v", err)
			}
		}()
	}

	wg.Wait()

	finalUser, err := repo.GetByID(ctx, "U_CONCURRENT")
	if err != nil {
		t.Fatalf("failed to fetch user after concurrency test: %v", err)
	}
	if finalUser.Balance != 0.00 {
		t.Errorf("expected final balance 0.00, got: %.2f", finalUser.Balance)
	}
}
