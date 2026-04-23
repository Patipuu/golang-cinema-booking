package repository

import (
	"context"
	"time"

	"booking_cinema_golang/internal/domain"
)

// CinemaRepository defines data access methods for cinemas and showtimes.
type CinemaRepository interface {
	FindByID(ctx context.Context, id string) (*domain.Cinema, error)
	ListAll(ctx context.Context, page domain.Page) ([]domain.Cinema, domain.PageResult, error)
	ListShowtimes(ctx context.Context, cinemaID string, date time.Time) ([]domain.Showtime, error)
	ListSeatsByCinema(ctx context.Context, cinemaID string) ([]domain.Seat, error)
	Create(ctx context.Context, cinema *domain.Cinema) error
	Update(ctx context.Context, cinema *domain.Cinema) error
	Delete(ctx context.Context, id string) error
}
