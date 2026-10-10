package domain_test

import (
	"errors"
	"testing"

	"paperid-wallet/internal/domain"
)

func TestUser_Deduct(t *testing.T) {
	tests := []struct {
		name            string
		initialBalance float64
		amount          float64
		expectedBalance float64
		expectedErr     error
	}{
		{
			name:            "sufficient balance",
			initialBalance: 5000.00,
			amount:          1000.00,
			expectedBalance: 4000.00,
			expectedErr:     nil,
		},
		{
			name:            "zero amount",
			initialBalance: 5000.00,
			amount:          0,
			expectedBalance: 5000.00,
			expectedErr:     domain.ErrInvalidAmount,
		},
		{
			name:            "negative amount",
			initialBalance: 5000.00,
			amount:          -50.00,
			expectedBalance: 5000.00,
			expectedErr:     domain.ErrInvalidAmount,
		},
		{
			name:            "insufficient balance",
			initialBalance: 150.00,
			amount:          200.00,
			expectedBalance: 150.00,
			expectedErr:     domain.ErrInsufficientBalance,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			user := &domain.User{
				ID:      "U001",
				Name:    "Alice",
				Balance: tc.initialBalance,
			}

			err := user.Deduct(tc.amount)
			if !errors.Is(err, tc.expectedErr) {
				t.Fatalf("expected error %v, got %v", tc.expectedErr, err)
			}
			if user.Balance != tc.expectedBalance {
				t.Errorf("expected balance %.2f, got %.2f", tc.expectedBalance, user.Balance)
			}
		})
	}
}
