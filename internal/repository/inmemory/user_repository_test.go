package inmemory_test

import (
	"context"
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

	t.Run("returns user when ID exists", func(t *testing.T) {
		user, err := repo.GetByID(ctx, "U001")
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if user == nil || user.ID != "U001" || user.Name != "Alice" || user.Balance != 5000.00 {
			t.Errorf("unexpected user returned: %+v", user)
		}
	})

	t.Run("returns ErrUserNotFound when ID does not exist", func(t *testing.T) {
		user, err := repo.GetByID(ctx, "UNKNOWN")
		if err != domain.ErrUserNotFound {
			t.Errorf("expected ErrUserNotFound, got: %v", err)
		}
		if user != nil {
			t.Errorf("expected nil user, got: %+v", user)
		}
	})
}

func TestUserRepository_DeductBalance(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully deducts balance and records transaction", func(t *testing.T) {
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
	})

	t.Run("returns ErrUserNotFound when user does not exist", func(t *testing.T) {
		repo := inmemory.NewUserRepository(nil)

		_, err := repo.DeductBalance(ctx, "NON_EXISTENT", 100.00, nil)
		if err != domain.ErrUserNotFound {
			t.Errorf("expected ErrUserNotFound, got: %v", err)
		}
	})

	t.Run("returns ErrInvalidAmount when amount is negative or zero", func(t *testing.T) {
		repo := inmemory.NewUserRepository(map[string]*domain.User{
			"U001": {ID: "U001", Name: "Alice", Balance: 500.00},
		})

		_, err := repo.DeductBalance(ctx, "U001", 0, nil)
		if err != domain.ErrInvalidAmount {
			t.Errorf("expected ErrInvalidAmount, got: %v", err)
		}

		_, err = repo.DeductBalance(ctx, "U001", -50, nil)
		if err != domain.ErrInvalidAmount {
			t.Errorf("expected ErrInvalidAmount, got: %v", err)
		}
	})

	t.Run("returns ErrInsufficientBalance when amount exceeds balance", func(t *testing.T) {
		repo := inmemory.NewUserRepository(map[string]*domain.User{
			"U002": {ID: "U002", Name: "Bob", Balance: 150.00},
		})

		_, err := repo.DeductBalance(ctx, "U002", 200.00, nil)
		if err != domain.ErrInsufficientBalance {
			t.Errorf("expected ErrInsufficientBalance, got: %v", err)
		}

		persisted, _ := repo.GetByID(ctx, "U002")
		if persisted.Balance != 150.00 {
			t.Errorf("balance should remain 150.00, got: %.2f", persisted.Balance)
		}
	})

	t.Run("concurrent deductions are thread-safe and consistent", func(t *testing.T) {
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
	})
}
