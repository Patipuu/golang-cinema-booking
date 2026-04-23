package repository

import "errors"

var (
	// Generic not found semantics.
	ErrNotFound = errors.New("not found")

	// Generic input validation.
	ErrInvalidInput = errors.New("invalid input")
	ErrDuplicate    = errors.New("duplicate entry")

	// Auth-related.
	ErrEmailExists        = errors.New("email already exists")
	ErrUsernameExists     = errors.New("username already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrAccountNotVerified = errors.New("account not verified")

	// Booking-related.
	ErrSeatAlreadyTaken = errors.New("seat already taken")
	ErrSeatNotFound     = errors.New("seat not found")
	ErrSeatLockConflict = errors.New("seat is being processed")
	ErrBookingNotFound  = errors.New("booking not found")
	ErrBookingBadStatus = errors.New("booking has invalid status")

	// Payment-related.
	ErrPaymentBadStatus = errors.New("payment has invalid status")
)
