package usecase_test

import (
	"context"
	"errors"
	"testing"

	"paperid-wallet/internal/domain"
	"paperid-wallet/internal/usecase"
)

// mockUserRepository is a test double implementing domain.UserRepository
type mockUserRepository struct {
	getByIDFunc       func(ctx context.Context, id string) (*domain.User, error)
	deductBalanceFunc func(ctx context.Context, id string, amount float64) (*domain.User, error)
}

func (m *mockUserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return nil, domain.ErrUserNotFound
}

func (m *mockUserRepository) DeductBalance(ctx context.Context, id string, amount float64) (*domain.User, error) {
	if m.deductBalanceFunc != nil {
		return m.deductBalanceFunc(ctx, id, amount)
	}
	return nil, domain.ErrUserNotFound
}

func TestDisbursementUseCase_Execute(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully executes disbursement", func(t *testing.T) {
		mockRepo := &mockUserRepository{
			deductBalanceFunc: func(ctx context.Context, id string, amount float64) (*domain.User, error) {
				if id == "U001" && amount == 500.00 {
					return &domain.User{ID: "U001", Name: "Alice", Balance: 4500.00}, nil
				}
				return nil, errors.New("unexpected params")
			},
		}

		uc := usecase.NewDisbursementUseCase(mockRepo)
		input := usecase.DisburseInput{
			UserID:             "U001",
			Amount:             500.00,
			DestinationBank:    "BCA",
			DestinationAccount: "1234567890",
		}

		out, err := uc.Execute(ctx, input)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if out.Message != "Disbursement successful" {
			t.Errorf("expected message 'Disbursement successful', got '%s'", out.Message)
		}
		if out.RemainingBalance != 4500.00 {
			t.Errorf("expected remaining balance 4500.00, got %.2f", out.RemainingBalance)
		}
	})

	t.Run("returns ErrInvalidUserID when user ID is empty", func(t *testing.T) {
		mockRepo := &mockUserRepository{}
		uc := usecase.NewDisbursementUseCase(mockRepo)

		input := usecase.DisburseInput{
			UserID:             "",
			Amount:             100.00,
			DestinationBank:    "BCA",
			DestinationAccount: "123456",
		}

		_, err := uc.Execute(ctx, input)
		if !errors.Is(err, domain.ErrInvalidUserID) {
			t.Errorf("expected ErrInvalidUserID, got: %v", err)
		}
	})

	t.Run("returns ErrInvalidAmount when amount is zero or negative", func(t *testing.T) {
		mockRepo := &mockUserRepository{}
		uc := usecase.NewDisbursementUseCase(mockRepo)

		testCases := []float64{0, -10.50}
		for _, amount := range testCases {
			input := usecase.DisburseInput{
				UserID:             "U001",
				Amount:             amount,
				DestinationBank:    "BCA",
				DestinationAccount: "123456",
			}
			_, err := uc.Execute(ctx, input)
			if !errors.Is(err, domain.ErrInvalidAmount) {
				t.Errorf("for amount %.2f, expected ErrInvalidAmount, got: %v", amount, err)
			}
		}
	})

	t.Run("returns ErrUserNotFound when repository cannot find user", func(t *testing.T) {
		mockRepo := &mockUserRepository{
			deductBalanceFunc: func(ctx context.Context, id string, amount float64) (*domain.User, error) {
				return nil, domain.ErrUserNotFound
			},
		}
		uc := usecase.NewDisbursementUseCase(mockRepo)

		input := usecase.DisburseInput{
			UserID:             "UNKNOWN",
			Amount:             100.00,
			DestinationBank:    "BCA",
			DestinationAccount: "123456",
		}

		_, err := uc.Execute(ctx, input)
		if !errors.Is(err, domain.ErrUserNotFound) {
			t.Errorf("expected ErrUserNotFound, got: %v", err)
		}
	})

	t.Run("returns ErrInsufficientBalance when user has insufficient balance", func(t *testing.T) {
		mockRepo := &mockUserRepository{
			deductBalanceFunc: func(ctx context.Context, id string, amount float64) (*domain.User, error) {
				return nil, domain.ErrInsufficientBalance
			},
		}
		uc := usecase.NewDisbursementUseCase(mockRepo)

		input := usecase.DisburseInput{
			UserID:             "U002",
			Amount:             500.00,
			DestinationBank:    "BCA",
			DestinationAccount: "123456",
		}

		_, err := uc.Execute(ctx, input)
		if !errors.Is(err, domain.ErrInsufficientBalance) {
			t.Errorf("expected ErrInsufficientBalance, got: %v", err)
		}
	})
}
