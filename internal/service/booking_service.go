package service

import (
	"context"

	"booking_cinema_golang/internal/domain"
)

// BookingService defines booking-related operations.
type BookingService interface {
	CreateBooking(ctx context.Context, userID, showtimeID string, seatIDs []string) (*domain.Booking, error)
	GetBooking(ctx context.Context, id string) (*domain.Booking, error)
	ListByUserID(ctx context.Context, userID string, page domain.Page) ([]domain.Booking, domain.PageResult, error)
	GetTakenSeatIDsForShowtime(ctx context.Context, showtimeID string) ([]string, error)
}
