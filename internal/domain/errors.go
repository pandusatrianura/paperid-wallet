package domain

import "errors"

var (
	// ErrUserNotFound is returned when a requested user ID does not exist.
	ErrUserNotFound = errors.New("User not found")

	// ErrInsufficientBalance is returned when a user has less balance than the disbursement amount.
	ErrInsufficientBalance = errors.New("Insufficient balance")

	// ErrInvalidAmount is returned when disbursement amount is less than or equal to zero.
	ErrInvalidAmount = errors.New("Amount must be greater than zero")

	// ErrInvalidUserID is returned when the user ID is empty.
	ErrInvalidUserID = errors.New("User ID cannot be empty")
)
