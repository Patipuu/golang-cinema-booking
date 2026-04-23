package repository

import (
	"booking_cinema_golang/internal/domain"
	"context"
	"time"
)

type ShowtimeRepository interface {
    FindByID(ctx context.Context, id string) (*domain.Showtime, error)
    FindByCinema(ctx context.Context, cinemaID string, date time.Time) ([]*domain.Showtime, error)
    Create(ctx context.Context, showtime *domain.Showtime) error
    Update(ctx context.Context, showtime *domain.Showtime) error
    Delete(ctx context.Context, id string) error
    // Search nâng cao theo tên phim, cinema, date
    SearchShowtimes(ctx context.Context, movieName, cinemaID string, date *time.Time) ([]*domain.ShowtimeWithMovie, error)
}