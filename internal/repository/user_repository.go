package repository

import (
	"context"
	"time"

	"booking_cinema_golang/internal/domain"
)

// UserRepository defines data access methods for users.
type UserRepository interface {
	FindByID(ctx context.Context, id string) (*domain.User, error)
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	FindByUsername(ctx context.Context, username string) (*domain.User, error)
	Create(ctx context.Context, user *domain.User) error
	UpdateOTP(ctx context.Context, userID, otpCode string, expiry time.Time) error
	SetVerified(ctx context.Context, userID string) error
}
