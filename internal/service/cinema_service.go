package service

import (
	"context"
	"time"

	"booking_cinema_golang/internal/domain"
	"booking_cinema_golang/internal/repository"
)

// CinemaService defines cinema and showtime operations.
type CinemaService interface {
	GetCinema(id string) (*domain.Cinema, error)
	ListCinemas(page domain.Page) ([]domain.Cinema, domain.PageResult, error)
	ListShowtimes(cinemaID string, date time.Time) ([]domain.Showtime, error)
	ListSeatsByCinema(cinemaID string) ([]domain.Seat, error)
	CreateCinema(ctx context.Context, cinema *domain.Cinema) error
	UpdateCinema(ctx context.Context, cinema *domain.Cinema) error
	DeleteCinema(ctx context.Context, id string) error
}

type cinemaServiceImpl struct {
	cinemaRepo repository.CinemaRepository
}

func NewCinemaService(cinemaRepo repository.CinemaRepository) CinemaService {
	return &cinemaServiceImpl{
		cinemaRepo: cinemaRepo,
	}
}

func (s *cinemaServiceImpl) GetCinema(id string) (*domain.Cinema, error) {
	return s.cinemaRepo.FindByID(context.Background(), id)
}

func (s *cinemaServiceImpl) ListCinemas(page domain.Page) ([]domain.Cinema, domain.PageResult, error) {
	return s.cinemaRepo.ListAll(context.Background(), page)
}

func (s *cinemaServiceImpl) ListShowtimes(cinemaID string, date time.Time) ([]domain.Showtime, error) {
	return s.cinemaRepo.ListShowtimes(context.Background(), cinemaID, date)
}

func (s *cinemaServiceImpl) ListSeatsByCinema(cinemaID string) ([]domain.Seat, error) {
	return s.cinemaRepo.ListSeatsByCinema(context.Background(), cinemaID)
}

func (s *cinemaServiceImpl) CreateCinema(ctx context.Context, cinema *domain.Cinema) error {
	return s.cinemaRepo.Create(ctx, cinema)
}

func (s *cinemaServiceImpl) UpdateCinema(ctx context.Context, cinema *domain.Cinema) error {
	return s.cinemaRepo.Update(ctx, cinema)
}

func (s *cinemaServiceImpl) DeleteCinema(ctx context.Context, id string) error {
	return s.cinemaRepo.Delete(ctx, id)
}
