package usecase

import (
	"context"
	"strings"

	"paperid-wallet/internal/domain"
)

// DisburseInput represents the input data required to process a disbursement.
type DisburseInput struct {
	UserID             string  `json:"user_id"`
	Amount             float64 `json:"amount"`
	DestinationBank    string  `json:"destination_bank"`
	DestinationAccount string  `json:"destination_account"`
}

// DisburseOutput represents the result of a processed disbursement.
type DisburseOutput struct {
	Message          string  `json:"message"`
	RemainingBalance float64 `json:"remaining_balance"`
}

// DisbursementUseCase defines the contract for disbursement business operations.
type DisbursementUseCase interface {
	Execute(ctx context.Context, input DisburseInput) (*DisburseOutput, error)
}

// disbursementUseCase implements DisbursementUseCase.
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
	// Validate input
	if strings.TrimSpace(input.UserID) == "" {
		return nil, domain.ErrInvalidUserID
	}
	if input.Amount <= 0 {
		return nil, domain.ErrInvalidAmount
	}

	// Atomically deduct user balance in storage
	updatedUser, err := uc.userRepo.DeductBalance(ctx, input.UserID, input.Amount)
	if err != nil {
		return nil, err
	}

	return &DisburseOutput{
		Message:          "Disbursement successful",
		RemainingBalance: updatedUser.Balance,
	}, nil
}
