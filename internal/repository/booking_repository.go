package repository

import (
	"context"
	"booking_cinema_golang/internal/domain"
)

type BookingRepository interface {
	Create(ctx context.Context, booking *domain.Booking, seatIDs []string) error
	UpdateStatus(ctx context.Context, bookingID, status string) error
	FindByID(ctx context.Context, id string) (*domain.Booking, error)
	ListByUserID(ctx context.Context, userID string, page domain.Page) ([]domain.Booking, domain.PageResult, error)
	GetTakenSeatIDsForShowtime(ctx context.Context, showtimeID string) (map[string]string, error)
	Delete(ctx context.Context, id string) error
	HasOverlappingBooking(ctx context.Context, userID, showtimeID string) (bool, error)
	GetStats(ctx context.Context) (map[string]any, error)
}
