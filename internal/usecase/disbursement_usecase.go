package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"paperid-wallet/internal/domain"
)

// DisburseInput represents the input data required to process a disbursement.
type DisburseInput struct {
	UserID             string
	Amount             float64
	DestinationBank    string
	DestinationAccount string
}

// DisburseOutput represents the result of a processed disbursement.
type DisburseOutput struct {
	TransactionID    string
	Message          string
	RemainingBalance float64
}

// DisbursementUseCase defines the contract for disbursement business operations.
type DisbursementUseCase interface {
	Execute(ctx context.Context, input DisburseInput) (*DisburseOutput, error)
}

type disbursementUseCase struct {
	userRepo domain.UserRepository
}

// NewDisbursementUseCase instantiates a new DisbursementUseCase with injected dependencies.
func NewDisbursementUseCase(userRepo domain.UserRepository) DisbursementUseCase {
	return &disbursementUseCase{
		userRepo: userRepo,
	}
}

// Execute orchestrates the disbursement transaction.
func (uc *disbursementUseCase) Execute(ctx context.Context, input DisburseInput) (*DisburseOutput, error) {
	if strings.TrimSpace(input.UserID) == "" {
		return nil, domain.ErrInvalidUserID
	}
	if input.Amount <= 0 {
		return nil, domain.ErrInvalidAmount
	}

	tx := &domain.Transaction{
		ID:                 newTransactionID(),
		UserID:             input.UserID,
		Amount:             input.Amount,
		DestinationBank:    input.DestinationBank,
		DestinationAccount: input.DestinationAccount,
		Status:             domain.TransactionStatusSuccess,
		CreatedAt:          time.Now().UTC(),
	}

	updatedUser, err := uc.userRepo.DeductBalance(ctx, input.UserID, input.Amount, tx)
	if err != nil {
		return nil, err
	}

	return &DisburseOutput{
		TransactionID:    tx.ID,
		Message:          "Disbursement successful",
		RemainingBalance: updatedUser.Balance,
	}, nil
}

func newTransactionID() string {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return fmt.Sprintf("TRX-%d-%s", time.Now().Unix(), hex.EncodeToString(b))
}
