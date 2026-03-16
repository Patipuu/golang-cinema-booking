package repository

import (
	"context"
	"fmt"

	"booking_cinema_golang/internal/database"
	"booking_cinema_golang/internal/domain"
)

type bookingRepository struct {
	db *database.DB
}

// Create implements [BookingRepository].
func (r *bookingRepository) Create(ctx context.Context, booking *domain.Booking, seatIDs []string) error {
	panic("unimplemented")
}

// GetTakenSeatIDsForShowtime implements [BookingRepository].
func (r *bookingRepository) GetTakenSeatIDsForShowtime(ctx context.Context, showtimeID string) ([]string, error) {
	panic("unimplemented")
}

// ListByUserID implements [BookingRepository].
func (r *bookingRepository) ListByUserID(ctx context.Context, userID string, page domain.Page) ([]domain.Booking, domain.PageResult, error) {
	panic("unimplemented")
}

func NewBookingRepository(db *database.DB) BookingRepository {
	return &bookingRepository{db: db}
}

func (r *bookingRepository) FindByID(ctx context.Context, id string) (*domain.Booking, error) {

	if r.db == nil || r.db.Pool == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	query := `
		SELECT id, user_id, showtime_id, status, total_price, created_at, updated_at
		FROM bookings
		WHERE id = $1`

	var b domain.Booking

	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&b.ID,
		&b.UserID,
		&b.ShowtimeID,
		&b.Status,
		&b.TotalPrice,
		&b.CreatedAt,
		&b.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("tìm booking thất bại: %w", err)
	}

	return &b, nil
}