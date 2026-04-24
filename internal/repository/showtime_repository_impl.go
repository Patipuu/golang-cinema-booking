package repository

import (
	"booking_cinema_golang/internal/database"
	"booking_cinema_golang/internal/domain"
	"context"
	"fmt"
	"time"
)

type ShowtimeRepositoryImpl struct {
    db *database.DB
}

func NewShowtimeRepository(db *database.DB) ShowtimeRepository {
    return &ShowtimeRepositoryImpl{db: db}
}

func (r *ShowtimeRepositoryImpl) FindByID(ctx context.Context, id string) (*domain.Showtime, error) {
    query := `SELECT id, cinema_id, movie_id, show_date, show_time, price, created_at FROM showtimes WHERE id = $1`
    var s domain.Showtime
    err := r.db.Pool.QueryRow(ctx, query, id).Scan(
        &s.ID, &s.CinemaID, &s.MovieID, &s.ShowDate, &s.ShowTime, &s.Price, &s.CreatedAt,
    )
    if err != nil {
        return nil, err
    }
    return &s, nil
}

func (r *ShowtimeRepositoryImpl) FindByCinema(ctx context.Context, cinemaID string, date time.Time) ([]*domain.Showtime, error) {
    query := `SELECT id, cinema_id, movie_id, show_date, show_time, price, created_at FROM showtimes WHERE cinema_id = $1 AND show_date = $2 ORDER BY show_time`
    rows, err := r.db.Pool.Query(ctx, query, cinemaID, date)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var showtimes []*domain.Showtime
    for rows.Next() {
        var s domain.Showtime
        if err := rows.Scan(&s.ID, &s.CinemaID, &s.MovieID, &s.ShowDate, &s.ShowTime, &s.Price, &s.CreatedAt); err != nil {
            return nil, err
        }
        showtimes = append(showtimes, &s)
    }
    return showtimes, nil
}

func (r *ShowtimeRepositoryImpl) Create(ctx context.Context, showtime *domain.Showtime) error {
    query := `INSERT INTO showtimes (id, cinema_id, movie_id, show_date, show_time, price, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7)`
    _, err := r.db.Pool.Exec(ctx, query,
        showtime.ID, showtime.CinemaID, showtime.MovieID, showtime.ShowDate, showtime.ShowTime, showtime.Price, showtime.CreatedAt,
    )
    return err
}

func (r *ShowtimeRepositoryImpl) Update(ctx context.Context, showtime *domain.Showtime) error {
    query := `UPDATE showtimes SET cinema_id = $1, movie_id = $2, show_date = $3, show_time = $4, price = $5 WHERE id = $6`
    _, err := r.db.Pool.Exec(ctx, query,
        showtime.CinemaID, showtime.MovieID, showtime.ShowDate, showtime.ShowTime, showtime.Price, showtime.ID,
    )
    return err
}

func (r *ShowtimeRepositoryImpl) Delete(ctx context.Context, id string) error {
    query := `DELETE FROM showtimes WHERE id = $1`
    _, err := r.db.Pool.Exec(ctx, query, id)
    return err
}

// SearchShowtimes tìm kiếm nâng cao theo tên phim, cinema, date
func (r *ShowtimeRepositoryImpl) SearchShowtimes(ctx context.Context, movieName, cinemaID string, date *time.Time) ([]*domain.ShowtimeWithMovie, error) {
    // Xây dựng truy vấn động
    query := `SELECT s.id, s.cinema_id, s.movie_id, s.show_date, s.show_time, s.price, s.created_at,
                     m.id, m.title, m.genre, m.duration, m.rating, m.created_at
              FROM showtimes s
              JOIN movies m ON s.movie_id = m.id
              WHERE 1=1`
    var args []interface{}
    idx := 1
    if movieName != "" {
        query += " AND LOWER(m.title) LIKE $" + fmt.Sprint(idx)
        args = append(args, "%"+movieName+"%")
        idx++
    }
    if cinemaID != "" {
        query += " AND s.cinema_id = $" + fmt.Sprint(idx)
        args = append(args, cinemaID)
        idx++
    }
    if date != nil {
        query += " AND s.show_date = $" + fmt.Sprint(idx)
        args = append(args, *date)
        idx++
    }
    query += " ORDER BY s.show_date, s.show_time"
    rows, err := r.db.Pool.Query(ctx, query, args...)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    var showtimes []*domain.ShowtimeWithMovie
    for rows.Next() {
        var s domain.Showtime
        var m domain.Movie
        if err := rows.Scan(
            &s.ID, &s.CinemaID, &s.MovieID, &s.ShowDate, &s.ShowTime, &s.Price, &s.CreatedAt,
            &m.ID, &m.Title, &m.Genre, &m.Duration, &m.Rating, &m.CreatedAt,
        ); err != nil {
            return nil, err
        }
        showtimes = append(showtimes, &domain.ShowtimeWithMovie{Showtime: s, Movie: m})
    }
    return showtimes, nil
}