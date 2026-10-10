package domain

// User represents the wallet account owner and balance entity.
type User struct {
	ID      string
	Name    string
	Balance float64
}

// Deduct verifies and performs deduction on the user's balance.
// Returns ErrInvalidAmount if amount <= 0, or ErrInsufficientBalance if balance < amount.
func (u *User) Deduct(amount float64) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}
	if u.Balance < amount {
		return ErrInsufficientBalance
	}
	u.Balance -= amount
	return nil
}
