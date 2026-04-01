package service

import (
	"context"

	"booking_cinema_golang/internal/domain"
)

// UserService defines user operations (async with goroutine as per requirements).
type UserService interface {
	GetUserByID(ctx context.Context, id string) (*domain.User, error)
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	CreateUser(ctx context.Context, user *domain.User) error

	// Profile management
	GetProfile(ctx context.Context, userID string) (*domain.User, error)
	UpdateProfile(ctx context.Context, userID, fullName, phone, avatarURL string) (*domain.User, error)
	ChangePassword(ctx context.Context, userID, oldPassword, newPassword string) error

	// Admin operations
	FindAllUsers(ctx context.Context, page, limit int, search string) ([]*domain.User, *domain.PageResult, error)
	UpdateUserStatus(ctx context.Context, userID string, isActive bool) error
	UpdateUserRole(ctx context.Context, userID, role string) error
}