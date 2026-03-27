package service

import (
	"context"

	"booking_cinema_golang/internal/domain"
)

// UserService defines user operations (async with goroutine as per requirements).
type UserService interface {
	GetUserByID(id int64) (*domain.User, error)
	GetUserByEmail(email string) (*domain.User, error)
	CreateUser(user *domain.User) error

	// Profile management
	GetProfile(ctx context.Context, userID string) (*domain.User, error)
	UpdateProfile(ctx context.Context, userID, fullName, phone string) (*domain.User, error)
	ChangePassword(ctx context.Context, userID, oldPassword, newPassword string) error
}