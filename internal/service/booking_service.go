package service

import "booking_cinema_golang/internal/domain"

// BookingService defines booking-related operations.
type BookingService interface {
	CreateBooking(userID, showtimeID string, seatIDs []string) (*domain.Booking, error)
	GetBooking(id string) (*domain.Booking, error)
	ListByUserID(userID string, page domain.Page) ([]domain.Booking, domain.PageResult, error)
	GetTakenSeatIDsForShowtime(showtimeID string) ([]string, error)
}
