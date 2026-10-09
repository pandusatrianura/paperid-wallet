package domain_test

import (
	"testing"

	"paperid-wallet/internal/domain"
)

func TestUser_Deduct(t *testing.T) {
	t.Run("successfully deducts amount when balance is sufficient", func(t *testing.T) {
		user := &domain.User{
			ID:      "U001",
			Name:    "Alice",
			Balance: 5000.00,
		}

		err := user.Deduct(1000.00)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		expectedBalance := 4000.00
		if user.Balance != expectedBalance {
			t.Errorf("expected balance %.2f, got %.2f", expectedBalance, user.Balance)
		}
	})

	t.Run("returns ErrInvalidAmount when amount is zero", func(t *testing.T) {
		user := &domain.User{
			ID:      "U001",
			Name:    "Alice",
			Balance: 5000.00,
		}

		err := user.Deduct(0)
		if err != domain.ErrInvalidAmount {
			t.Errorf("expected ErrInvalidAmount, got: %v", err)
		}
	})

	t.Run("returns ErrInvalidAmount when amount is negative", func(t *testing.T) {
		user := &domain.User{
			ID:      "U001",
			Name:    "Alice",
			Balance: 5000.00,
		}

		err := user.Deduct(-50.00)
		if err != domain.ErrInvalidAmount {
			t.Errorf("expected ErrInvalidAmount, got: %v", err)
		}
	})

	t.Run("returns ErrInsufficientBalance when amount exceeds balance", func(t *testing.T) {
		user := &domain.User{
			ID:      "U002",
			Name:    "Bob",
			Balance: 150.00,
		}

		err := user.Deduct(200.00)
		if err != domain.ErrInsufficientBalance {
			t.Errorf("expected ErrInsufficientBalance, got: %v", err)
		}

		// Balance should remain unchanged
		if user.Balance != 150.00 {
			t.Errorf("expected balance to remain 150.00, got %.2f", user.Balance)
		}
	})
}
