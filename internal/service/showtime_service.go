package service

import (
	"booking_cinema_golang/internal/domain"
	"context"
	"time"
)

type ShowtimeService interface {
    GetShowtime(ctx context.Context, id string) (*domain.Showtime, error)
    ListShowtimesByCinema(ctx context.Context, cinemaID string, date time.Time) ([]*domain.Showtime, error)
    CreateShowtime(ctx context.Context, showtime *domain.Showtime) error
    UpdateShowtime(ctx context.Context, showtime *domain.Showtime) error
    DeleteShowtime(ctx context.Context, id string) error
    // Search nâng cao theo tên phim, cinema, date
    SearchShowtimes(ctx context.Context, movieName, cinemaID string, date *time.Time) ([]*domain.ShowtimeWithMovie, error)
}