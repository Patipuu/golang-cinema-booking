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
	ListAll(ctx context.Context, page domain.Page) ([]domain.User, domain.PageResult, error)
	UpdateRole(ctx context.Context, userID, role string) error
	Update(ctx context.Context, user *domain.User) error
	UpdatePassword(ctx context.Context, userID, newHash string) error
	FindAll(ctx context.Context, page, limit int, search string) ([]*domain.User, int, error)
	UpdateStatus(ctx context.Context, userID string, isActive bool) error
}
