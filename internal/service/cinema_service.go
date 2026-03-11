package service

import (
	"time"

	"booking_cinema_golang/internal/domain"
)

// CinemaService defines cinema and showtime operations.
type CinemaService interface {
	GetCinema(id string) (*domain.Cinema, error)
	ListCinemas(page domain.Page) ([]domain.Cinema, domain.PageResult, error)
	ListShowtimes(cinemaID string, date time.Time) ([]domain.Showtime, error)
	ListSeatsByCinema(cinemaID string) ([]domain.Seat, error)
}
