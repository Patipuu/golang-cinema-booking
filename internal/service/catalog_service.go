package service

import (
	"context"
	"time"

	"booking_cinema_golang/internal/domain"
	"booking_cinema_golang/internal/repository"
)

type CatalogService interface {
	// Cinema
	CreateCinema(ctx context.Context, c *domain.Cinema) error
	UpdateCinema(ctx context.Context, c *domain.Cinema) error
	DeleteCinema(ctx context.Context, id string) error
	GetCinema(ctx context.Context, id string) (*domain.Cinema, error)
	ListCinemas(ctx context.Context, page domain.Page) ([]domain.Cinema, domain.PageResult, error)

	// Room
	CreateRoom(ctx context.Context, r *domain.ScreeningRoom) error
	UpdateRoom(ctx context.Context, r *domain.ScreeningRoom) error
	DeleteRoom(ctx context.Context, id string) error
	GetRoomByID(ctx context.Context, id string) (*domain.ScreeningRoom, error)
	ListRooms(ctx context.Context, cinemaID string) ([]domain.ScreeningRoom, error)
	GetSeats(ctx context.Context, roomID string) ([]domain.Seat, error)

	// Movie
	CreateMovie(ctx context.Context, m *domain.Movie) error
	UpdateMovie(ctx context.Context, m *domain.Movie) error
	DeleteMovie(ctx context.Context, id string) error
	GetMovie(ctx context.Context, id string) (*domain.Movie, error)
	ListMovies(ctx context.Context, status, search string, page domain.Page) ([]domain.Movie, domain.PageResult, error)

	// Showtime
	CreateShowtime(ctx context.Context, s *domain.Showtime) error
	UpdateShowtime(ctx context.Context, s *domain.Showtime) error
	DeleteShowtime(ctx context.Context, id string) error
	ListShowtimes(ctx context.Context, cinemaID, movieID string, date time.Time) ([]domain.Showtime, error)
	GetShowtime(ctx context.Context, id string) (*domain.Showtime, error)
}

type catalogService struct {
	repo repository.CatalogRepository
}

func NewCatalogService(repo repository.CatalogRepository) CatalogService {
	return &catalogService{repo: repo}
}

func (s *catalogService) CreateCinema(ctx context.Context, c *domain.Cinema) error {
	return s.repo.CreateCinema(ctx, c)
}

func (s *catalogService) UpdateCinema(ctx context.Context, c *domain.Cinema) error {
	return s.repo.UpdateCinema(ctx, c)
}

func (s *catalogService) DeleteCinema(ctx context.Context, id string) error {
	return s.repo.DeleteCinema(ctx, id)
}

func (s *catalogService) GetCinema(ctx context.Context, id string) (*domain.Cinema, error) {
	return s.repo.GetCinemaByID(ctx, id)
}

func (s *catalogService) ListCinemas(ctx context.Context, page domain.Page) ([]domain.Cinema, domain.PageResult, error) {
	return s.repo.ListCinemas(ctx, page)
}

func (s *catalogService) CreateRoom(ctx context.Context, r *domain.ScreeningRoom) error {
	return s.repo.CreateRoom(ctx, r)
}

func (s *catalogService) UpdateRoom(ctx context.Context, r *domain.ScreeningRoom) error {
	return s.repo.UpdateRoom(ctx, r)
}

func (s *catalogService) DeleteRoom(ctx context.Context, id string) error {
	return s.repo.DeleteRoom(ctx, id)
}

func (s *catalogService) GetRoomByID(ctx context.Context, id string) (*domain.ScreeningRoom, error) {
	return s.repo.GetRoomByID(ctx, id)
}

func (s *catalogService) ListRooms(ctx context.Context, cinemaID string) ([]domain.ScreeningRoom, error) {
	return s.repo.ListRoomsByCinema(ctx, cinemaID)
}

func (s *catalogService) GetSeats(ctx context.Context, roomID string) ([]domain.Seat, error) {
	return s.repo.GetSeatsByRoom(ctx, roomID)
}


func (s *catalogService) CreateMovie(ctx context.Context, m *domain.Movie) error {
	return s.repo.CreateMovie(ctx, m)
}

func (s *catalogService) UpdateMovie(ctx context.Context, m *domain.Movie) error {
	return s.repo.UpdateMovie(ctx, m)
}

func (s *catalogService) DeleteMovie(ctx context.Context, id string) error {
	return s.repo.DeleteMovie(ctx, id)
}

func (s *catalogService) GetMovie(ctx context.Context, id string) (*domain.Movie, error) {
	return s.repo.GetMovieByID(ctx, id)
}

func (s *catalogService) ListMovies(ctx context.Context, status, search string, page domain.Page) ([]domain.Movie, domain.PageResult, error) {
	return s.repo.ListMovies(ctx, status, search, page)
}

func (s *catalogService) CreateShowtime(ctx context.Context, st *domain.Showtime) error {
	// Check for room conflict (simplified)
	// TODO: Add complex conflict check (overlaps)
	return s.repo.CreateShowtime(ctx, st)
}

func (s *catalogService) UpdateShowtime(ctx context.Context, st *domain.Showtime) error {
	return s.repo.UpdateShowtime(ctx, st)
}

func (s *catalogService) DeleteShowtime(ctx context.Context, id string) error {
	return s.repo.DeleteShowtime(ctx, id)
}

func (s *catalogService) ListShowtimes(ctx context.Context, cinemaID, movieID string, date time.Time) ([]domain.Showtime, error) {
	return s.repo.ListShowtimes(ctx, cinemaID, movieID, date)
}

func (s *catalogService) GetShowtime(ctx context.Context, id string) (*domain.Showtime, error) {
	return s.repo.GetShowtimeByID(ctx, id)
}
