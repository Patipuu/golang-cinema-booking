package service

import (
	"context"

	"booking_cinema_golang/internal/domain"
	"booking_cinema_golang/internal/repository"
)

type bookingService struct {
	repo repository.BookingRepository
}

func NewBookingService(repo repository.BookingRepository) BookingService {
	return &bookingService{repo: repo}
}

func (s *bookingService) CreateBooking(ctx context.Context, userID, showtimeID string, seatIDs []string) (*domain.Booking, error) {
	b := &domain.Booking{
		UserID:     userID,
		ShowtimeID: showtimeID,
		Status:     "pending",
	}
	if err := s.repo.Create(ctx, b, seatIDs); err != nil {
		return nil, err
	}
	return b, nil
}

func (s *bookingService) GetBooking(ctx context.Context, id string) (*domain.Booking, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *bookingService) ListByUserID(ctx context.Context, userID string, page domain.Page) ([]domain.Booking, domain.PageResult, error) {
	return s.repo.ListByUserID(ctx, userID, page)
}

func (s *bookingService) GetTakenSeatIDsForShowtime(ctx context.Context, showtimeID string) ([]string, error) {
	return s.repo.GetTakenSeatIDsForShowtime(ctx, showtimeID)
}

