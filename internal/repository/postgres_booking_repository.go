package repository

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"booking_cinema_golang/internal/database"
	"booking_cinema_golang/internal/domain"

	"github.com/jackc/pgx/v5/pgconn"
)

// PostgresBookingRepository is a PostgreSQL implementation of BookingRepository.
// It uses pessimistic row locking (SELECT ... FOR UPDATE NOWAIT) to prevent double booking.
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
	return &b, nil
}

func (r *PostgresBookingRepository) ListByUserID(ctx context.Context, userID string, page domain.Page) ([]domain.Booking, domain.PageResult, error) {
	// Total count.
	var totalCount int
	if err := r.db.Pool.QueryRow(ctx, `
		SELECT COUNT(*)::int
		FROM bookings
		WHERE user_id = $1
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
		out = append(out, b)
	}
	if err := rows.Err(); err != nil {
		return nil, domain.PageResult{}, fmt.Errorf("booking: rows: %w", err)
	}

	return out, domain.PageResult{
		Page:       page.Page,
		Limit:      limit,
		TotalCount: totalCount,
	}, nil
}

func (r *PostgresBookingRepository) GetTakenSeatIDsForShowtime(ctx context.Context, showtimeID string) ([]string, error) {
	// Return seat_number codes (frontend uses codes like "A1").
	rows, err := r.db.Pool.Query(ctx, `
		SELECT DISTINCT s.seat_number
		FROM bookings b
		JOIN booking_seats bs ON bs.booking_id = b.id
		JOIN seats s ON s.id = bs.seat_id
		WHERE b.showtime_id = $1
			AND b.status IN ('pending','confirmed')
		ORDER BY s.seat_number
	`, showtimeID)
	if err != nil {
		return nil, fmt.Errorf("booking: get taken seats: %w", err)
	}
	defer rows.Close()

	var taken []string
	for rows.Next() {
		var code string
		if err := rows.Scan(&code); err != nil {
			return nil, fmt.Errorf("booking: scan taken seat: %w", err)
		}
		taken = append(taken, code)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("booking: rows: %w", err)
	}
	return taken, nil
}

func (r *PostgresBookingRepository) Create(ctx context.Context, booking *domain.Booking, seatCodes []string) error {
	if booking == nil {
		return fmt.Errorf("booking: nil booking")
	}
	if booking.Status == "" {
		booking.Status = "pending"
	}
	if booking.UserID == "" || booking.ShowtimeID == "" {
		return fmt.Errorf("booking: missing user_id/showtime_id")
	}

	// Dedupe + keep deterministic order for consistent locking.
	seen := make(map[string]struct{}, len(seatCodes))
	uniq := make([]string, 0, len(seatCodes))
	for _, c := range seatCodes {
		if c == "" {
			continue
		}
		if _, ok := seen[c]; ok {
			continue
		}
		seen[c] = struct{}{}
		uniq = append(uniq, c)
	}
	seatCodes = uniq
	if len(seatCodes) == 0 {
		return fmt.Errorf("booking: no seats")
	}

	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("booking: begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// Load showtime info (cinema_id + price) for lock scope and total_price.
	var cinemaID string
	var pricePerSeat float64
	if err := tx.QueryRow(ctx, `
		SELECT cinema_id::text, price::float8
		FROM showtimes
		WHERE id = $1
	`, booking.ShowtimeID).Scan(&cinemaID, &pricePerSeat); err != nil {
		return fmt.Errorf("booking: load showtime: %w", err)
	}

	// 1) Lock seat rows pessimistically.
	//    ORDER BY ensures a deterministic lock acquisition order.
	//    NOWAIT makes request fail fast if another tx holds the lock.
	type seatRow struct {
		seatID    string
		seatNumber string
	}
	seatIDByCode := make(map[string]seatRow, len(seatCodes))
	rows, err := tx.Query(ctx, `
		SELECT id::text, seat_number
		FROM seats
		WHERE cinema_id = $1 AND seat_number = ANY($2::text[])
		ORDER BY id
		FOR UPDATE NOWAIT
	`, cinemaID, seatCodes)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "55P03" {
			return ErrSeatLockConflict
		}
		return fmt.Errorf("booking: lock seats: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var sr seatRow
		if err := rows.Scan(&sr.seatID, &sr.seatNumber); err != nil {
			return fmt.Errorf("booking: scan seat row: %w", err)
		}
		seatIDByCode[sr.seatNumber] = sr
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("booking: read locked seats: %w", err)
	}

	if len(seatIDByCode) != len(seatCodes) {
		return ErrSeatNotFound
	}

	// Keep deterministic seatID order (helps reproducibility, not required for correctness).
	lockedSeatIDs := make([]string, 0, len(seatCodes))
	for _, code := range seatCodes {
		lockedSeatIDs = append(lockedSeatIDs, seatIDByCode[code].seatID)
	}
	sort.Strings(lockedSeatIDs)

	// 2) Re-check availability while holding locks.
	var taken []string
	takenRows, err := tx.Query(ctx, `
		SELECT DISTINCT s.seat_number
		FROM bookings b
		JOIN booking_seats bs ON bs.booking_id = b.id
		JOIN seats s ON s.id = bs.seat_id
		WHERE b.showtime_id = $1
			AND b.status IN ('pending','confirmed')
			AND bs.seat_id = ANY($2::uuid[])
		ORDER BY s.seat_number
	`, booking.ShowtimeID, lockedSeatIDs)
	if err != nil {
		return fmt.Errorf("booking: check taken seats: %w", err)
	}
	defer takenRows.Close()
	for takenRows.Next() {
		var code string
		if err := takenRows.Scan(&code); err != nil {
			return fmt.Errorf("booking: scan taken seat: %w", err)
		}
		taken = append(taken, code)
	}
	if err := takenRows.Err(); err != nil {
		return fmt.Errorf("booking: read taken seats: %w", err)
	}
	if len(taken) > 0 {
		return ErrSeatAlreadyTaken
	}

	// 3) Insert booking + booking_seats atomically.
	totalPrice := pricePerSeat * float64(len(seatCodes))
	var bookingID string
	if err := tx.QueryRow(ctx, `
		INSERT INTO bookings (user_id, showtime_id, status, total_price)
		VALUES ($1, $2, $3, $4)
		RETURNING id::text
	`, booking.UserID, booking.ShowtimeID, booking.Status, totalPrice).Scan(&bookingID); err != nil {
		return fmt.Errorf("booking: insert booking: %w", err)
	}
	booking.ID = bookingID
	booking.TotalPrice = totalPrice

	if _, err := tx.Exec(ctx, `
		INSERT INTO booking_seats (booking_id, seat_id)
		SELECT $1, unnest($2::uuid[])
	`, bookingID, lockedSeatIDs); err != nil {
		return fmt.Errorf("booking: insert booking_seats: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("booking: commit: %w", err)
	}
	return nil
}

