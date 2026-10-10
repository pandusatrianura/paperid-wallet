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
	getByIDFunc                 func(ctx context.Context, id string) (*domain.User, error)
	deductBalanceFunc           func(ctx context.Context, id string, amount float64, tx *domain.Transaction) (*domain.User, error)
	getTransactionsByUserIDFunc func(ctx context.Context, userID string) ([]*domain.Transaction, error)
}

func (m *mockUserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return nil, domain.ErrUserNotFound
}

func (m *mockUserRepository) DeductBalance(ctx context.Context, id string, amount float64, tx *domain.Transaction) (*domain.User, error) {
	if m.deductBalanceFunc != nil {
		return m.deductBalanceFunc(ctx, id, amount, tx)
	}
	return nil, domain.ErrUserNotFound
}

func (m *mockUserRepository) GetTransactionsByUserID(ctx context.Context, userID string) ([]*domain.Transaction, error) {
	if m.getTransactionsByUserIDFunc != nil {
		return m.getTransactionsByUserIDFunc(ctx, userID)
	}
	return nil, nil
}

func TestDisbursementUseCase_Execute_Success(t *testing.T) {
	ctx := context.Background()
	var recordedTx *domain.Transaction

	mockRepo := &mockUserRepository{
		deductBalanceFunc: func(ctx context.Context, id string, amount float64, tx *domain.Transaction) (*domain.User, error) {
			if id == "U001" && amount == 500.00 {
				recordedTx = tx
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
	if out.TransactionID == "" {
		t.Errorf("expected non-empty transaction ID")
	}
	if out.Message != "Disbursement successful" {
		t.Errorf("expected message 'Disbursement successful', got '%s'", out.Message)
	}
	if out.RemainingBalance != 4500.00 {
		t.Errorf("expected remaining balance 4500.00, got %.2f", out.RemainingBalance)
	}

	if recordedTx == nil || recordedTx.DestinationBank != "BCA" || recordedTx.DestinationAccount != "1234567890" {
		t.Errorf("expected recorded transaction with bank and account details, got: %+v", recordedTx)
	}
}

func TestDisbursementUseCase_Execute_ValidationAndErrors(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		input       usecase.DisburseInput
		mockRepo    domain.UserRepository
		expectedErr error
	}{
		{
			name: "empty user id",
			input: usecase.DisburseInput{
				UserID: "",
				Amount: 100.00,
			},
			mockRepo:    &mockUserRepository{},
			expectedErr: domain.ErrInvalidUserID,
		},
		{
			name: "zero amount",
			input: usecase.DisburseInput{
				UserID: "U001",
				Amount: 0,
			},
			mockRepo:    &mockUserRepository{},
			expectedErr: domain.ErrInvalidAmount,
		},
		{
			name: "negative amount",
			input: usecase.DisburseInput{
				UserID: "U001",
				Amount: -25.50,
			},
			mockRepo:    &mockUserRepository{},
			expectedErr: domain.ErrInvalidAmount,
		},
		{
			name: "user not found",
			input: usecase.DisburseInput{
				UserID: "UNKNOWN",
				Amount: 100.00,
			},
			mockRepo: &mockUserRepository{
				deductBalanceFunc: func(ctx context.Context, id string, amount float64, tx *domain.Transaction) (*domain.User, error) {
					return nil, domain.ErrUserNotFound
				},
			},
			expectedErr: domain.ErrUserNotFound,
		},
		{
			name: "insufficient balance",
			input: usecase.DisburseInput{
				UserID: "U002",
				Amount: 500.00,
			},
			mockRepo: &mockUserRepository{
				deductBalanceFunc: func(ctx context.Context, id string, amount float64, tx *domain.Transaction) (*domain.User, error) {
					return nil, domain.ErrInsufficientBalance
				},
			},
			expectedErr: domain.ErrInsufficientBalance,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			uc := usecase.NewDisbursementUseCase(tc.mockRepo)
			_, err := uc.Execute(ctx, tc.input)
			if !errors.Is(err, tc.expectedErr) {
				t.Fatalf("expected error %v, got %v", tc.expectedErr, err)
			}
		})
	}
}
