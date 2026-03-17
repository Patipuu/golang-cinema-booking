package service

import "booking_cinema_golang/internal/domain"

// UserService defines user operations (async with goroutine as per requirements).
type UserService interface {
	GetUserByID(id int64) (*domain.User, error)
	GetUserByEmail(email string) (*domain.User, error)
	CreateUser(user *domain.User) error
}