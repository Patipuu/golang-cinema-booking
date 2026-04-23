package repository

import (
	"context"
	"time"

	"booking_cinema_golang/internal/database"
	"booking_cinema_golang/internal/domain"

	"github.com/jackc/pgx/v5/pgconn"
)

type PostgresCinemaRepository struct {
	db *database.DB
}

func NewCinemaRepository(db *database.DB) CinemaRepository {
	return &PostgresCinemaRepository{db: db}
}

func (r *PostgresCinemaRepository) FindByID(ctx context.Context, id string) (*domain.Cinema, error) {
	var c domain.Cinema
	row := r.db.Pool.QueryRow(ctx, `
		SELECT id::text, name, location, city, total_seats, created_at, updated_at
		FROM cinemas
		WHERE id = $1
	`, id)
	if err := row.Scan(&c.ID, &c.Name, &c.Location, &c.City, &c.TotalSeats, &c.CreatedAt, &c.UpdatedAt); err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *PostgresCinemaRepository) ListAll(ctx context.Context, page domain.Page) ([]domain.Cinema, domain.PageResult, error) {
	// Total count
	var totalCount int
	if err := r.db.Pool.QueryRow(ctx, `SELECT COUNT(*)::int FROM cinemas`).Scan(&totalCount); err != nil {
		return nil, domain.PageResult{}, err
	}

	limit := page.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	offset := page.Offset()

	rows, err := r.db.Pool.Query(ctx, `
		SELECT id::text, name, location, city, total_seats, created_at, updated_at
		FROM cinemas
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, domain.PageResult{}, err
	}
	defer rows.Close()

	var cinemas []domain.Cinema
	for rows.Next() {
		var c domain.Cinema
		if err := rows.Scan(&c.ID, &c.Name, &c.Location, &c.City, &c.TotalSeats, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, domain.PageResult{}, err
		}
		cinemas = append(cinemas, c)
	}
	if err := rows.Err(); err != nil {
		return nil, domain.PageResult{}, err
	}

	pageResult := domain.PageResult{
		TotalCount: totalCount,
		Page:       page.Page,
		Limit:      limit,
	}

	return cinemas, pageResult, nil
}

func (r *PostgresCinemaRepository) ListShowtimes(ctx context.Context, cinemaID string, date time.Time) ([]domain.Showtime, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT id::text, cinema_id::text, movie_id::text, show_date, show_time, price::float8, created_at
		FROM showtimes
		WHERE cinema_id = $1 AND DATE(show_date) = DATE($2)
		ORDER BY show_time
	`, cinemaID, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var showtimes []domain.Showtime
	for rows.Next() {
		var s domain.Showtime
		if err := rows.Scan(&s.ID, &s.CinemaID, &s.MovieID, &s.ShowDate, &s.ShowTime, &s.Price, &s.CreatedAt); err != nil {
			return nil, err
		}
		showtimes = append(showtimes, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return showtimes, nil
}

func (r *PostgresCinemaRepository) ListSeatsByCinema(ctx context.Context, cinemaID string) ([]domain.Seat, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT id::text, cinema_id::text, seat_number, row_name, seat_type, created_at
		FROM seats
		WHERE cinema_id = $1
		ORDER BY row_name, seat_number
	`, cinemaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var seats []domain.Seat
	for rows.Next() {
		var s domain.Seat
		if err := rows.Scan(&s.ID, &s.CinemaID, &s.SeatNumber, &s.RowName, &s.SeatType, &s.CreatedAt); err != nil {
			return nil, err
		}
		seats = append(seats, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return seats, nil
}

func (r *PostgresCinemaRepository) Create(ctx context.Context, cinema *domain.Cinema) error {
	if cinema == nil {
		return ErrInvalidInput
	}

	err := r.db.Pool.QueryRow(ctx, `
		INSERT INTO cinemas (name, location, city, total_seats)
		VALUES ($1, $2, $3, $4)
		RETURNING id::text, created_at, updated_at
	`, cinema.Name, cinema.Location, cinema.City, cinema.TotalSeats).Scan(&cinema.ID, &cinema.CreatedAt, &cinema.UpdatedAt)

	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			return ErrDuplicate
		}
		return err
	}
	return nil
}

func (r *PostgresCinemaRepository) Update(ctx context.Context, cinema *domain.Cinema) error {
	if cinema == nil || cinema.ID == "" {
		return ErrInvalidInput
	}

	result, err := r.db.Pool.Exec(ctx, `
		UPDATE cinemas
		SET name = $1, location = $2, city = $3, total_seats = $4, updated_at = NOW()
		WHERE id = $5
	`, cinema.Name, cinema.Location, cinema.City, cinema.TotalSeats, cinema.ID)

	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *PostgresCinemaRepository) Delete(ctx context.Context, id string) error {
	if id == "" {
		return ErrInvalidInput
	}

	result, err := r.db.Pool.Exec(ctx, `DELETE FROM cinemas WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
