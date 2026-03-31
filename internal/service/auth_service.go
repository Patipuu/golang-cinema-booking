package service

import (
	"context"

	"booking_cinema_golang/internal/domain"
)

// AuthService defines authentication-related operations (register, login, OTP).
type AuthService interface {
	Register(ctx context.Context, email, password, username, fullName, phone string) (*domain.User, error)
	Login(ctx context.Context, email, password string) (*domain.User, string, error)
	VerifyOTP(ctx context.Context, userID, otpCode string) error
	ResendVerification(ctx context.Context, email string) error
}
