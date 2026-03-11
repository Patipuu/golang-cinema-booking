package repository

import (
	"context"

	"booking_cinema_golang/internal/domain"
)

// BookingRepository defines data access methods for bookings.
type BookingRepository interface {
	FindByID(ctx context.Context, id string) (*domain.Booking, error)
	Create(ctx context.Context, booking *domain.Booking, seatIDs []string) error
	ListByUserID(ctx context.Context, userID string, page domain.Page) ([]domain.Booking, domain.PageResult, error)
	GetTakenSeatIDsForShowtime(ctx context.Context, showtimeID string) ([]string, error)
}
