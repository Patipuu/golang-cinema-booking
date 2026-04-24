package service

import (
	"context"
	"fmt"
	"time"

	"booking_cinema_golang/internal/domain"
	"booking_cinema_golang/internal/infrastructure/redis"
	"booking_cinema_golang/internal/repository"
)

type bookingService struct {
	repo      repository.BookingRepository
	redis     *redis.Client
	pricing   PricingService
	wsHub     interface{} // Using interface{} to avoid circular dependency, will type assert to Hub if needed or just use broadcast method
	lockTimer time.Duration
}

// Hub interface matches the broadcast method we need.
type HubInterface interface {
	BroadcastSeatUpdate(showtimeID, seatID, status string)
}

func NewBookingService(repo repository.BookingRepository, rdb *redis.Client, pricing PricingService, hub HubInterface) BookingService {
	return &bookingService{
		repo:      repo,
		redis:     rdb,
		pricing:   pricing,
		wsHub:     hub,
		lockTimer: 10 * time.Minute,
	}
}

func (s *bookingService) LockSeat(ctx context.Context, showtimeID, seatID string) (bool, error) {
	ok, err := s.redis.LockSeat(ctx, showtimeID, seatID, s.lockTimer)
	if err != nil {
		return false, err
	}
	if ok {
		// Broadcast to all clients
		if hub, ok := s.wsHub.(HubInterface); ok {
			hub.BroadcastSeatUpdate(showtimeID, seatID, "locked")
		}
	}
	return ok, nil
}

func (s *bookingService) UnlockSeat(ctx context.Context, showtimeID, seatID string) error {
	err := s.redis.UnlockSeat(ctx, showtimeID, seatID)
	if err == nil {
		if hub, ok := s.wsHub.(HubInterface); ok {
			hub.BroadcastSeatUpdate(showtimeID, seatID, "available")
		}
	}
	return err
}

func (s *bookingService) CreateBooking(ctx context.Context, userID, showtimeID string, seatIDs []string) (*domain.Booking, error) {
	// 0. Kiểm tra trùng giờ chiếu
	overlap, err := s.repo.HasOverlappingBooking(ctx, userID, showtimeID)
	if err != nil {
		return nil, err
	}
	if overlap {
		return nil, fmt.Errorf("bạn đã có một vé với suất chiếu khác trong khoảng thời gian này")
	}

	// 1. Calculate price
	// For simplicity, we assume we know seat types or retrieve them
	// In a real app, we'd fetch seats first
	total := 0.0
	// ... pricing logic here ...

	b := &domain.Booking{
		UserID:     userID,
		ShowtimeID: showtimeID,
		Status:     "pending",
		TotalPrice: total,
		BookedAt:   time.Now(),
	}

	if err := s.repo.Create(ctx, b, seatIDs); err != nil {
		return nil, err
	}

	// 2. Unlock in Redis as they are now "sold" or "pending" in DB
	for _, sid := range seatIDs {
		s.redis.UnlockSeat(ctx, showtimeID, sid)
		if hub, ok := s.wsHub.(HubInterface); ok {
			hub.BroadcastSeatUpdate(showtimeID, sid, "sold")
		}
	}

	return b, nil
}

func (s *bookingService) GetBooking(ctx context.Context, id string) (*domain.Booking, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *bookingService) ListByUserID(ctx context.Context, userID string, page domain.Page) ([]domain.Booking, domain.PageResult, error) {
	return s.repo.ListByUserID(ctx, userID, page)
}

func (s *bookingService) GetTakenSeatIDsForShowtime(ctx context.Context, showtimeID string) (map[string]string, error) {
	// Combine DB taken seats + Redis locked seats
	dbTaken, err := s.repo.GetTakenSeatIDsForShowtime(ctx, showtimeID)
	if err != nil {
		return nil, err
	}

	redisLocked, err := s.redis.GetLockedSeats(ctx, showtimeID)
	if err != nil {
		// Just log and continue with DB results if Redis fails
		return dbTaken, nil
	}

	// Merge: Redis seats are always 'holding' (pending)
	// DB confirmed seats override Redis locks
	seatMap := make(map[string]string)
	for code, status := range dbTaken {
		seatMap[code] = status
	}
	for _, code := range redisLocked {
		if _, exists := seatMap[code]; !exists {
			seatMap[code] = "pending"
		}
	}

	return seatMap, nil
}

func (s *bookingService) CancelBooking(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *bookingService) UpdateBookingStatus(ctx context.Context, id, status string) error {
	if err := s.repo.UpdateStatus(ctx, id, status); err != nil {
		return err
	}

	if status == "confirmed" || status == "paid" {
		booking, err := s.repo.FindByID(ctx, id)
		if err == nil && booking != nil {
			for _, sid := range booking.Seats {
				if hub, ok := s.wsHub.(HubInterface); ok {
					hub.BroadcastSeatUpdate(booking.ShowtimeID, sid, "sold")
				}
			}
		}
	}
	return nil
}

func (s *bookingService) GetStats(ctx context.Context) (map[string]any, error) {
	return s.repo.GetStats(ctx)
}

func (s *bookingService) ListBookings(ctx context.Context, page, limit int) ([]domain.Booking, error) {
	// TODO: Implement pagination in repository
	return s.repo.ListAll(ctx, page, limit)
}

func (s *bookingService) CancelBooking(ctx context.Context, id string) error {
	return s.repo.Cancel(ctx, id)
}
