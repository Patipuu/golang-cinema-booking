package repository

import (
	"context"
	"time"

	"booking_cinema_golang/internal/domain"
)

type CatalogRepository interface {
	// Cinema
	CreateCinema(ctx context.Context, c *domain.Cinema) error
	UpdateCinema(ctx context.Context, c *domain.Cinema) error
	DeleteCinema(ctx context.Context, id string) error
	GetCinemaByID(ctx context.Context, id string) (*domain.Cinema, error)
	ListCinemas(ctx context.Context, page domain.Page) ([]domain.Cinema, domain.PageResult, error)

	// Room
	CreateRoom(ctx context.Context, r *domain.ScreeningRoom) error
	UpdateRoom(ctx context.Context, r *domain.ScreeningRoom) error
	DeleteRoom(ctx context.Context, id string) error
	GetRoomByID(ctx context.Context, id string) (*domain.ScreeningRoom, error)
	ListRoomsByCinema(ctx context.Context, cinemaID string) ([]domain.ScreeningRoom, error)

	// Seat mapping
	CreateSeats(ctx context.Context, seats []domain.Seat) error
	GetSeatsByRoom(ctx context.Context, roomID string) ([]domain.Seat, error)

	// Movie
	CreateMovie(ctx context.Context, m *domain.Movie) error
	UpdateMovie(ctx context.Context, m *domain.Movie) error
	DeleteMovie(ctx context.Context, id string) error
	GetMovieByID(ctx context.Context, id string) (*domain.Movie, error)
	ListMovies(ctx context.Context, status, search string, page domain.Page) ([]domain.Movie, domain.PageResult, error)

	// Showtime
	CreateShowtime(ctx context.Context, s *domain.Showtime) error
	UpdateShowtime(ctx context.Context, s *domain.Showtime) error
	DeleteShowtime(ctx context.Context, id string) error
	GetShowtimeByID(ctx context.Context, id string) (*domain.Showtime, error)
	ListShowtimes(ctx context.Context, cinemaID, movieID string, date time.Time) ([]domain.Showtime, error)
}
