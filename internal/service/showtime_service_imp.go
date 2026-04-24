package service

import (
	"booking_cinema_golang/internal/domain"
	"booking_cinema_golang/internal/repository"
	"context"
	"time"

	"github.com/google/uuid"
)

type showtimeService struct {
    repo repository.ShowtimeRepository
}

func NewShowtimeService(repo repository.ShowtimeRepository) ShowtimeService {
    return &showtimeService{repo: repo}
}

func (s *showtimeService) GetShowtime(ctx context.Context, id string) (*domain.Showtime, error) {
    return s.repo.FindByID(ctx, id)
}

func (s *showtimeService) ListShowtimesByCinema(ctx context.Context, cinemaID string, date time.Time) ([]*domain.Showtime, error) {
    return s.repo.FindByCinema(ctx, cinemaID, date)
}

func (s *showtimeService) CreateShowtime(ctx context.Context, showtime *domain.Showtime) error {
    if showtime.ID == "" {
        showtime.ID = uuid.New().String()
    }
    showtime.CreatedAt = time.Now()
    return s.repo.Create(ctx, showtime)
}

func (s *showtimeService) UpdateShowtime(ctx context.Context, showtime *domain.Showtime) error {
    return s.repo.Update(ctx, showtime)
}

func (s *showtimeService) DeleteShowtime(ctx context.Context, id string) error {
    return s.repo.Delete(ctx, id)
}

func (s *showtimeService) SearchShowtimes(ctx context.Context, movieName, cinemaID string, date *time.Time) ([]*domain.ShowtimeWithMovie, error) {
    return s.repo.SearchShowtimes(ctx, movieName, cinemaID, date)
}