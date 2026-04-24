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
	GetTakenSeatIDsForShowtime(ctx context.Context, showtimeID string) (map[string]string, error)
	LockSeat(ctx context.Context, showtimeID, seatID string) (bool, error)
	UnlockSeat(ctx context.Context, showtimeID, seatID string) error
	CancelBooking(ctx context.Context, id string) error
	GetStats(ctx context.Context) (map[string]any, error)
	UpdateBookingStatus(ctx context.Context, id, status string) error
}
