package service

import "booking_cinema_golang/internal/domain"

// AuthService defines authentication-related operations (register, login, OTP).
type AuthService interface {
	Register(email, password, fullName, username string) (*domain.User, error)
	Login(email, password string) (*domain.User, string, error) // user, token, error
	VerifyOTP(userID, otpCode string) error
}
