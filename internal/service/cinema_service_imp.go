package service

import (
	"context"
	"booking_cinema_golang/internal/domain"
	"booking_cinema_golang/internal/repository"
	"time"
)

type cinemaService struct {
    repo repository.CinemaRepository
}

func NewCinemaService(repo repository.CinemaRepository) CinemaService {
    return &cinemaService{repo: repo}
}

func (s *cinemaService) GetCinema(id string) (*domain.Cinema, error) {
    return s.repo.FindByID(context.Background(), id)
}

func (s *cinemaService) ListCinemas(page domain.Page) ([]domain.Cinema, domain.PageResult, error) {
    return s.repo.ListAll(context.Background(), page)
}

func (s *cinemaService) ListShowtimes(cinemaID string, date time.Time) ([]domain.Showtime, error) {
    return s.repo.ListShowtimes(context.Background(), cinemaID, date)
}

func (s *cinemaService) ListSeatsByCinema(cinemaID string) ([]domain.Seat, error) {
    return s.repo.ListSeatsByCinema(context.Background(), cinemaID)
}

// Thêm cho method lọc trong bài tập
func (s *cinemaService) FilterRooms(ctx context.Context, cinemaID string, minSeats int, roomType string) ([]domain.ScreeningRoom, error) {
    return s.repo.FilterRooms(ctx, cinemaID, minSeats, roomType)
}