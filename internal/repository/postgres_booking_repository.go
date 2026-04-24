package repository

import (
	"context"
	"fmt"
	"time"

	"booking_cinema_golang/internal/database"
	"booking_cinema_golang/internal/domain"
)

// PostgresBookingRepository is a PostgreSQL implementation of BookingRepository.
type PostgresBookingRepository struct {
	db *database.DB
}

func NewBookingRepository(db *database.DB) BookingRepository {
	return &PostgresBookingRepository{db: db}
}

func (r *PostgresBookingRepository) FindByID(ctx context.Context, id string) (*domain.Booking, error) {
	var b domain.Booking
	row := r.db.Pool.QueryRow(ctx, `
		SELECT id::text, user_id::text, showtime_id::text, status, total_price::float8, created_at, updated_at
		FROM bookings
		WHERE id = $1
	`, id)
	if err := row.Scan(&b.ID, &b.UserID, &b.ShowtimeID, &b.Status, &b.TotalPrice, &b.CreatedAt, &b.UpdatedAt); err != nil {
		return nil, fmt.Errorf("booking: find by id: %w", err)
	}
	b.Seats, _ = r.getSeatsByBookingID(ctx, id)
	return &b, nil
}

func (r *PostgresBookingRepository) ListByUserID(ctx context.Context, userID string, page domain.Page) ([]domain.Booking, domain.PageResult, error) {
	var totalCount int
	if err := r.db.Pool.QueryRow(ctx, `
		SELECT COUNT(*)::int
		FROM bookings
		WHERE user_id = $1
		  AND (status != 'pending' OR created_at > NOW() - INTERVAL '5 minutes')
	`, userID).Scan(&totalCount); err != nil {
		return nil, domain.PageResult{}, fmt.Errorf("booking: count by user: %w", err)
	}

	limit := page.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	offset := page.Offset()

	rows, err := r.db.Pool.Query(ctx, `
		SELECT id::text, user_id::text, showtime_id::text, status, total_price::float8, created_at, updated_at
		FROM bookings
		WHERE user_id = $1
		  AND (status != 'pending' OR created_at > NOW() - INTERVAL '5 minutes')
		ORDER BY created_at DESC, id
		LIMIT $2 OFFSET $3
	`, userID, limit, offset)
	if err != nil {
		return nil, domain.PageResult{}, fmt.Errorf("booking: list by user: %w", err)
	}
	defer rows.Close()

	out := make([]domain.Booking, 0)
	for rows.Next() {
		var b domain.Booking
		if err := rows.Scan(&b.ID, &b.UserID, &b.ShowtimeID, &b.Status, &b.TotalPrice, &b.CreatedAt, &b.UpdatedAt); err != nil {
			return nil, domain.PageResult{}, fmt.Errorf("booking: scan row: %w", err)
		}
		b.Seats, _ = r.getSeatsByBookingID(ctx, b.ID)
		out = append(out, b)
	}
	return out, domain.PageResult{
		Page:       page.Page,
		Limit:      limit,
		TotalCount: totalCount,
	}, nil
}

func (r *PostgresBookingRepository) GetTakenSeatIDsForShowtime(ctx context.Context, showtimeID string) (map[string]string, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT s.seat_number, b.status
		FROM bookings b
		JOIN booking_seats bs ON bs.booking_id = b.id
		JOIN seats s ON s.id = bs.seat_id
		WHERE b.showtime_id = $1
			AND (b.status IN ('confirmed', 'paid') OR (b.status = 'pending' AND b.created_at > NOW() - INTERVAL '5 minutes'))
		ORDER BY s.seat_number
	`, showtimeID)
	if err != nil {
		return nil, fmt.Errorf("booking: get taken seats: %w", err)
	}
	defer rows.Close()

	taken := make(map[string]string)
	for rows.Next() {
		var code, status string
		if err := rows.Scan(&code, &status); err != nil {
			return nil, fmt.Errorf("booking: scan taken seat: %w", err)
		}
		taken[code] = status
	}
	return taken, nil
}

func (r *PostgresBookingRepository) Create(ctx context.Context, booking *domain.Booking, seatCodes []string) error {
	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var roomID string
	var pricePerSeat float64
	var startTime time.Time
	if err := tx.QueryRow(ctx, `
		SELECT room_id::text, price::float8, start_time
		FROM showtimes
		WHERE id = $1
	`, booking.ShowtimeID).Scan(&roomID, &pricePerSeat, &startTime); err != nil {
		return err
	}

	if startTime.Before(time.Now()) {
		return ErrShowtimeExpired
	}

	rows, err := tx.Query(ctx, `
		SELECT id::text, seat_number
		FROM seats
		WHERE room_id = $1 AND seat_number = ANY($2::text[])
		ORDER BY id
		FOR UPDATE NOWAIT
	`, roomID, seatCodes)
	if err != nil {
		return err
	}
	defer rows.Close()

	seatIDByCode := make(map[string]string)
	for rows.Next() {
		var sid, sn string
		rows.Scan(&sid, &sn)
		seatIDByCode[sn] = sid
	}

	// Double Check: Ensure these seats are not already in a pending (< 5m) or confirmed booking for THIS showtime
	var existingSeatCount int
	err = tx.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM booking_seats bs
		JOIN bookings b ON b.id = bs.booking_id
		JOIN seats s ON s.id = bs.seat_id
		WHERE b.showtime_id = $1
		  AND s.seat_number = ANY($2::text[])
		  AND (b.status = 'confirmed' OR (b.status = 'pending' AND b.created_at > NOW() - INTERVAL '5 minutes'))
	`, booking.ShowtimeID, seatCodes).Scan(&existingSeatCount)
	if err != nil {
		return err
	}
	if existingSeatCount > 0 {
		return ErrSeatAlreadyTaken
	}

	lockedSeatIDs := make([]string, 0, len(seatCodes))
	for _, code := range seatCodes {
		if id, ok := seatIDByCode[code]; ok {
			lockedSeatIDs = append(lockedSeatIDs, id)
		}
	}

	totalPrice := pricePerSeat * float64(len(lockedSeatIDs))
	err = tx.QueryRow(ctx, `
		INSERT INTO bookings (user_id, showtime_id, status, total_price)
		VALUES ($1, $2, $3, $4)
		RETURNING id::text, created_at
	`, booking.UserID, booking.ShowtimeID, "pending", totalPrice).Scan(&booking.ID, &booking.CreatedAt)
	if err != nil {
		return err
	}
	booking.TotalPrice = totalPrice

	for _, sid := range lockedSeatIDs {
		tx.Exec(ctx, "INSERT INTO booking_seats (booking_id, seat_id) VALUES ($1, $2)", booking.ID, sid)
	}

	return tx.Commit(ctx)
}

func (r *PostgresBookingRepository) UpdateStatus(ctx context.Context, bookingID, status string) error {
	_, err := r.db.Pool.Exec(ctx, "UPDATE bookings SET status = $1 WHERE id = $2", status, bookingID)
	return err
}

func (r *PostgresBookingRepository) HasOverlappingBooking(ctx context.Context, userID, showtimeID string) (bool, error) {
	query := `
		SELECT COUNT(*)
		FROM bookings b
		JOIN showtimes s_old ON b.showtime_id = s_old.id
		JOIN showtimes s_new ON s_new.id = $2
		WHERE b.user_id = $1 
		  AND b.status IN ('pending', 'confirmed')
		  AND s_old.id != s_new.id
		  AND s_old.start_time < s_new.end_time 
		  AND s_old.end_time > s_new.start_time
	`
	var count int
	r.db.Pool.QueryRow(ctx, query, userID, showtimeID).Scan(&count)
	return count > 0, nil
}

func (r *PostgresBookingRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.Pool.Exec(ctx, "DELETE FROM bookings WHERE id=$1", id)
	return err
}

func (r *PostgresBookingRepository) getSeatsByBookingID(ctx context.Context, bookingID string) ([]string, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT s.seat_number
		FROM booking_seats bs
		JOIN seats s ON s.id = bs.seat_id
		WHERE bs.booking_id = $1
		ORDER BY s.seat_number
	`, bookingID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var seats []string
	for rows.Next() {
		var s string
		rows.Scan(&s)
		seats = append(seats, s)
	}
	return seats, nil
}

func (r *PostgresBookingRepository) GetStats(ctx context.Context) (map[string]any, error) {
	stats := make(map[string]any)

	// Revenue today
	var revenue float64
	r.db.Pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(total_price), 0) FROM bookings 
		WHERE created_at::date = CURRENT_DATE AND status IN ('paid', 'confirmed')
	`).Scan(&revenue)
	stats["today_revenue"] = revenue

	// Tickets today
	var tickets int
	r.db.Pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM bookings 
		WHERE created_at::date = CURRENT_DATE AND status IN ('paid', 'confirmed')
	`).Scan(&tickets)
	stats["today_tickets"] = tickets

	// Top movies
	rows, _ := r.db.Pool.Query(ctx, `
		SELECT m.title_vi, COUNT(b.id) as count 
		FROM bookings b 
		JOIN showtimes s ON b.showtime_id = s.id 
		JOIN movies m ON s.movie_id = m.id 
		WHERE b.status IN ('paid', 'confirmed')
		GROUP BY m.title_vi 
		ORDER BY count DESC 
		LIMIT 3
	`)
	defer rows.Close()
	var topMovies []string
	for rows.Next() {
		var title string
		var count int
		rows.Scan(&title, &count)
		topMovies = append(topMovies, title)
	}
	stats["top_movies"] = topMovies

	// Occupancy rate (mock for now or complex)
	stats["occupancy_rate"] = "72%"

	return stats, nil
}
