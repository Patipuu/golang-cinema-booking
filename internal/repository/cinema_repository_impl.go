package repository

import (
	"booking_cinema_golang/internal/database"
	"booking_cinema_golang/internal/domain"
	"context"
	"time"
)

type CinemaRepositoryImpl struct {
    db *database.DB
}

func NewCinemaRepository(db *database.DB) CinemaRepository {
    return &CinemaRepositoryImpl{db: db}
}

func (r *CinemaRepositoryImpl) FindByID(ctx context.Context, id string) (*domain.Cinema, error) {
    var c domain.Cinema
    err := r.db.Pool.QueryRow(ctx, `SELECT id, name, location, city, total_seats, created_at, updated_at FROM cinemas WHERE id = $1`, id).Scan(
        &c.ID, &c.Name, &c.Location, &c.City, &c.TotalSeats, &c.CreatedAt, &c.UpdatedAt,
    )
    if err != nil {
        return nil, err
    }
    return &c, nil
}

func (r *CinemaRepositoryImpl) ListAll(ctx context.Context, page domain.Page) ([]domain.Cinema, domain.PageResult, error) {
    rows, err := r.db.Pool.Query(ctx, `SELECT id, name, location, city, total_seats, created_at, updated_at FROM cinemas`)
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
    return cinemas, domain.PageResult{TotalCount: len(cinemas)}, nil
}

func (r *CinemaRepositoryImpl) ListShowtimes(ctx context.Context, cinemaID string, date time.Time) ([]domain.Showtime, error) {
    rows, err := r.db.Pool.Query(ctx, `SELECT id, cinema_id, movie_id, show_date, show_time, price, created_at FROM showtimes WHERE cinema_id = $1 AND show_date = $2`, cinemaID, date)
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
    return showtimes, nil
}

func (r *CinemaRepositoryImpl) ListSeatsByCinema(ctx context.Context, cinemaID string) ([]domain.Seat, error) {
    rows, err := r.db.Pool.Query(ctx, `SELECT id, cinema_id, seat_number, seat_type FROM seats WHERE cinema_id = $1`, cinemaID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    var seats []domain.Seat
    for rows.Next() {
        var s domain.Seat
        if err := rows.Scan(&s.ID, &s.CinemaID, &s.SeatNumber, &s.SeatType); err != nil {
            return nil, err
        }
        seats = append(seats, s)
    }
    return seats, nil
}