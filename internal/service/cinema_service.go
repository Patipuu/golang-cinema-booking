package service

import (
	"context"
	"time"

	"booking_cinema_golang/internal/domain"
)

// CinemaService defines cinema and showtime operations.
type CinemaService interface {
	GetCinema(id string) (*domain.Cinema, error)
	ListCinemas(page domain.Page) ([]domain.Cinema, domain.PageResult, error)
	ListShowtimes(cinemaID string, date time.Time) ([]domain.Showtime, error)
	ListSeatsByCinema(cinemaID string) ([]domain.Seat, error)
	// Thêm cho method lọc trong bài tập
	FilterRooms(ctx context.Context, cinemaID string, minSeats int, roomType string) ([]domain.ScreeningRoom, error)
}
