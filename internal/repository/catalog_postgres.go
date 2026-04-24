package repository

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"booking_cinema_golang/internal/database"
	"booking_cinema_golang/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type CatalogPostgresRepository struct {
	db *database.DB
}

func NewCatalogRepository(db *database.DB) CatalogRepository {
	return &CatalogPostgresRepository{db: db}
}

// Cinema
func (r *CatalogPostgresRepository) CreateCinema(ctx context.Context, c *domain.Cinema) error {
	if c.ID == "" {
		c.ID = uuid.NewString()
	}
	_, err := r.db.Pool.Exec(ctx, `
		INSERT INTO cinemas (id, name, location, city, hotline, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, c.ID, c.Name, c.Location, c.City, c.Hotline, time.Now(), time.Now())
	return err
}

func (r *CatalogPostgresRepository) UpdateCinema(ctx context.Context, c *domain.Cinema) error {
	_, err := r.db.Pool.Exec(ctx, `
		UPDATE cinemas SET name=$1, location=$2, city=$3, hotline=$4, updated_at=$5
		WHERE id=$6
	`, c.Name, c.Location, c.City, c.Hotline, time.Now(), c.ID)
	return err
}

func (r *CatalogPostgresRepository) DeleteCinema(ctx context.Context, id string) error {
	_, err := r.db.Pool.Exec(ctx, "DELETE FROM cinemas WHERE id=$1", id)
	return err
}

func (r *CatalogPostgresRepository) GetCinemaByID(ctx context.Context, id string) (*domain.Cinema, error) {
	var c domain.Cinema
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id, name, location, city, hotline, created_at, updated_at FROM cinemas WHERE id=$1
	`, id).Scan(&c.ID, &c.Name, &c.Location, &c.City, &c.Hotline, &c.CreatedAt, &c.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	}
	return &c, err
}

func (r *CatalogPostgresRepository) ListCinemas(ctx context.Context, page domain.Page) ([]domain.Cinema, domain.PageResult, error) {
	var total int
	r.db.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM cinemas").Scan(&total)

	rows, err := r.db.Pool.Query(ctx, `
		SELECT id, name, location, city, hotline, created_at, updated_at FROM cinemas
		ORDER BY created_at DESC LIMIT $1 OFFSET $2
	`, page.Limit, page.Offset())
	if err != nil {
		return nil, domain.PageResult{}, err
	}
	defer rows.Close()

	var list []domain.Cinema
	for rows.Next() {
		var c domain.Cinema
		rows.Scan(&c.ID, &c.Name, &c.Location, &c.City, &c.Hotline, &c.CreatedAt, &c.UpdatedAt)
		list = append(list, c)
	}
	return list, domain.PageResult{Page: page.Page, Limit: page.Limit, TotalCount: total}, nil
}

// Room
func (r *CatalogPostgresRepository) CreateRoom(ctx context.Context, room *domain.ScreeningRoom) error {
	if room.ID == "" {
		room.ID = uuid.NewString()
	}
	_, err := r.db.Pool.Exec(ctx, `
		INSERT INTO screening_rooms (id, cinema_id, name, room_type, total_seats, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, room.ID, room.CinemaID, room.Name, room.RoomType, room.TotalSeats, time.Now(), time.Now())
	return err
}

func (r *CatalogPostgresRepository) UpdateRoom(ctx context.Context, room *domain.ScreeningRoom) error {
	_, err := r.db.Pool.Exec(ctx, `
		UPDATE screening_rooms SET name=$1, room_type=$2, total_seats=$3, updated_at=$4
		WHERE id=$5
	`, room.Name, room.RoomType, room.TotalSeats, time.Now(), room.ID)
	return err
}

func (r *CatalogPostgresRepository) DeleteRoom(ctx context.Context, id string) error {
	_, err := r.db.Pool.Exec(ctx, "DELETE FROM screening_rooms WHERE id=$1", id)
	return err
}

func (r *CatalogPostgresRepository) GetRoomByID(ctx context.Context, id string) (*domain.ScreeningRoom, error) {
	var room domain.ScreeningRoom
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id, cinema_id, name, room_type, total_seats, created_at, updated_at FROM screening_rooms WHERE id=$1
	`, id).Scan(&room.ID, &room.CinemaID, &room.Name, &room.RoomType, &room.TotalSeats, &room.CreatedAt, &room.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	}
	return &room, err
}

func (r *CatalogPostgresRepository) ListRoomsByCinema(ctx context.Context, cinemaID string) ([]domain.ScreeningRoom, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT id, cinema_id, name, room_type, total_seats, created_at, updated_at FROM screening_rooms
		WHERE cinema_id=$1 ORDER BY created_at
	`, cinemaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []domain.ScreeningRoom
	for rows.Next() {
		var room domain.ScreeningRoom
		rows.Scan(&room.ID, &room.CinemaID, &room.Name, &room.RoomType, &room.TotalSeats, &room.CreatedAt, &room.UpdatedAt)
		list = append(list, room)
	}
	return list, nil
}

// Seat mapping
func (r *CatalogPostgresRepository) CreateSeats(ctx context.Context, seats []domain.Seat) error {
	batch := &pgx.Batch{}
	for _, s := range seats {
		if s.ID == "" {
			s.ID = uuid.NewString()
		}
		batch.Queue(`
			INSERT INTO seats (id, room_id, seat_number, row_name, seat_type, status, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`, s.ID, s.RoomID, s.SeatNumber, s.RowName, s.SeatType, s.Status, time.Now())
	}
	br := r.db.Pool.SendBatch(ctx, batch)
	return br.Close()
}

func (r *CatalogPostgresRepository) GetSeatsByRoom(ctx context.Context, roomID string) ([]domain.Seat, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT id, room_id, seat_number, row_name, seat_type, status, created_at FROM seats
		WHERE room_id=$1 ORDER BY row_name, seat_number
	`, roomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []domain.Seat
	for rows.Next() {
		var s domain.Seat
		rows.Scan(&s.ID, &s.RoomID, &s.SeatNumber, &s.RowName, &s.SeatType, &s.Status, &s.CreatedAt)
		list = append(list, s)
	}
	return list, nil
}

// Movie
func (r *CatalogPostgresRepository) CreateMovie(ctx context.Context, m *domain.Movie) error {
	if m.ID == "" {
		m.ID = uuid.NewString()
	}
	_, err := r.db.Pool.Exec(ctx, `
		INSERT INTO movies (id, title, title_vi, title_en, genre, director, cast_members, duration_mins, language, subtitle, rating_label, description, poster_url, trailer_url, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	`, m.ID, m.TitleVI, m.TitleVI, m.TitleEN, strings.Join(m.Genre, ", "), m.Director, m.CastMembers, m.DurationMins, m.Language, m.Subtitle, m.RatingLabel, m.Description, m.PosterURL, m.TrailerURL, m.Status, time.Now())
	if err != nil {
		fmt.Printf("[DB ERROR] CreateMovie failed: %v\n", err)
	}
	return err
}


func (r *CatalogPostgresRepository) UpdateMovie(ctx context.Context, m *domain.Movie) error {
	_, err := r.db.Pool.Exec(ctx, `
		UPDATE movies 
		SET title=$1, title_vi=$2, title_en=$3, genre=$4, director=$5, cast_members=$6, duration_mins=$7, language=$8, subtitle=$9, rating_label=$10, description=$11, poster_url=$12, trailer_url=$13, status=$14
		WHERE id=$15
	`, m.TitleVI, m.TitleVI, m.TitleEN, strings.Join(m.Genre, ", "), m.Director, m.CastMembers, m.DurationMins, m.Language, m.Subtitle, m.RatingLabel, m.Description, m.PosterURL, m.TrailerURL, m.Status, m.ID)
	return err
}

func (r *CatalogPostgresRepository) DeleteMovie(ctx context.Context, id string) error {
	_, err := r.db.Pool.Exec(ctx, "DELETE FROM movies WHERE id=$1", id)
	return err
}

func (r *CatalogPostgresRepository) GetMovieByID(ctx context.Context, id string) (*domain.Movie, error) {
	var m domain.Movie
	var genreStr string
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id, title_vi, title_en, genre, director, cast_members, duration_mins, language, subtitle, rating_label, description, poster_url, trailer_url, status, avg_rating, created_at FROM movies WHERE id=$1
	`, id).Scan(&m.ID, &m.TitleVI, &m.TitleEN, &genreStr, &m.Director, &m.CastMembers, &m.DurationMins, &m.Language, &m.Subtitle, &m.RatingLabel, &m.Description, &m.PosterURL, &m.TrailerURL, &m.Status, &m.AvgRating, &m.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	}
	if err == nil {
		m.Genre = strings.Split(genreStr, ", ")
	}
	return &m, err
}


func (r *CatalogPostgresRepository) ListMovies(ctx context.Context, status, search string, page domain.Page) ([]domain.Movie, domain.PageResult, error) {
	var total int
	query := "SELECT COUNT(*) FROM movies WHERE 1=1"
	var args []any
	argCount := 0

	if status != "" {
		argCount++
		query += fmt.Sprintf(" AND status=$%d", argCount)
		args = append(args, status)
	}
	if search != "" {
		argCount++
		query += fmt.Sprintf(" AND (title_vi ILIKE $%d OR title_en ILIKE $%d)", argCount, argCount)
		args = append(args, "%"+search+"%")
	}

	r.db.Pool.QueryRow(ctx, query, args...).Scan(&total)

	query = "SELECT id, COALESCE(title_vi, ''), COALESCE(title_en, ''), COALESCE(genre, ''), COALESCE(director, ''), COALESCE(cast_members, ''), COALESCE(duration_mins, 0), COALESCE(language, ''), COALESCE(subtitle, ''), COALESCE(rating_label, ''), COALESCE(description, ''), COALESCE(poster_url, ''), COALESCE(trailer_url, ''), COALESCE(status, ''), COALESCE(avg_rating, 0), created_at FROM movies WHERE 1=1"
	
	// Reset args because we need to build it again with limit/offset
	args = []any{}
	argCount = 0
	if status != "" {
		argCount++
		query += fmt.Sprintf(" AND status=$%d", argCount)
		args = append(args, status)
	}
	if search != "" {
		argCount++
		query += fmt.Sprintf(" AND (title_vi ILIKE $%d OR title_en ILIKE $%d)", argCount, argCount)
		args = append(args, "%"+search+"%")
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argCount+1, argCount+2)
	args = append(args, page.Limit, page.Offset())

	rows, err := r.db.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, domain.PageResult{}, err
	}
	defer rows.Close()

	var list []domain.Movie
	for rows.Next() {
		var m domain.Movie
		var genreStr string
		err := rows.Scan(&m.ID, &m.TitleVI, &m.TitleEN, &genreStr, &m.Director, &m.CastMembers, &m.DurationMins, &m.Language, &m.Subtitle, &m.RatingLabel, &m.Description, &m.PosterURL, &m.TrailerURL, &m.Status, &m.AvgRating, &m.CreatedAt)
		if err != nil {
			log.Printf("Scan movie failed: %v", err)
			continue
		}
		m.Genre = strings.Split(strings.Trim(genreStr, "{}"), ",")
		list = append(list, m)
	}
	return list, domain.PageResult{Page: page.Page, Limit: page.Limit, TotalCount: total}, nil
}


// Showtime
func (r *CatalogPostgresRepository) CreateShowtime(ctx context.Context, s *domain.Showtime) error {
	if s.ID == "" {
		s.ID = uuid.NewString()
	}
	if s.ShowDate == "" {
		s.ShowDate = s.StartTime.Format("2006-01-02")
	}
	if s.ShowTime == "" {
		s.ShowTime = s.StartTime.Format("15:04:05")
	}
	if s.Price == 0 && s.BasePrice > 0 {
		s.Price = s.BasePrice
	}

	_, err := r.db.Pool.Exec(ctx, `
		INSERT INTO showtimes (id, movie_id, room_id, cinema_id, start_time, end_time, show_date, show_time, price, base_price, price_modifier, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`, s.ID, s.MovieID, s.RoomID, s.CinemaID, s.StartTime, s.EndTime, s.ShowDate, s.ShowTime, s.Price, s.BasePrice, s.PriceModifier, s.Status, time.Now())
	return err
}

func (r *CatalogPostgresRepository) UpdateShowtime(ctx context.Context, s *domain.Showtime) error {
	if s.ShowDate == "" {
		s.ShowDate = s.StartTime.Format("2006-01-02")
	}
	if s.ShowTime == "" {
		s.ShowTime = s.StartTime.Format("15:04:05")
	}
	if s.Price == 0 && s.BasePrice > 0 {
		s.Price = s.BasePrice
	}

	_, err := r.db.Pool.Exec(ctx, `
		UPDATE showtimes SET movie_id=$1, room_id=$2, cinema_id=$3, start_time=$4, end_time=$5, show_date=$6, show_time=$7, price=$8, base_price=$9, price_modifier=$10, status=$11
		WHERE id=$12
	`, s.MovieID, s.RoomID, s.CinemaID, s.StartTime, s.EndTime, s.ShowDate, s.ShowTime, s.Price, s.BasePrice, s.PriceModifier, s.Status, s.ID)
	return err
}

func (r *CatalogPostgresRepository) DeleteShowtime(ctx context.Context, id string) error {
	_, err := r.db.Pool.Exec(ctx, "DELETE FROM showtimes WHERE id=$1", id)
	return err
}

func (r *CatalogPostgresRepository) GetShowtimeByID(ctx context.Context, id string) (*domain.Showtime, error) {
	var s domain.Showtime
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id, movie_id, room_id, cinema_id, start_time, end_time, base_price, price_modifier, status, created_at FROM showtimes WHERE id=$1
	`, id).Scan(&s.ID, &s.MovieID, &s.RoomID, &s.CinemaID, &s.StartTime, &s.EndTime, &s.BasePrice, &s.PriceModifier, &s.Status, &s.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	}
	return &s, err
}

func (r *CatalogPostgresRepository) ListShowtimes(ctx context.Context, cinemaID, movieID string, date time.Time) ([]domain.Showtime, error) {
	query := `
		SELECT s.id, s.movie_id, s.room_id, s.cinema_id, s.start_time, s.end_time, s.base_price, s.price_modifier, s.status, s.created_at
		FROM showtimes s
		JOIN screening_rooms sr ON s.room_id = sr.id
		WHERE 1=1
	`
	var args []any
	if cinemaID != "" {
		args = append(args, cinemaID)
		query += fmt.Sprintf(" AND sr.cinema_id=$%d", len(args))
	}
	if movieID != "" {
		args = append(args, movieID)
		query += fmt.Sprintf(" AND s.movie_id=$%d", len(args))
	}
	if !date.IsZero() {
		args = append(args, date.Format("2006-01-02"))
		query += fmt.Sprintf(" AND s.start_time::date=$%d", len(args))
	}
	query += " ORDER BY s.start_time"

	rows, err := r.db.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []domain.Showtime
	for rows.Next() {
		var s domain.Showtime
		if err := rows.Scan(&s.ID, &s.MovieID, &s.RoomID, &s.CinemaID, &s.StartTime, &s.EndTime, &s.BasePrice, &s.PriceModifier, &s.Status, &s.CreatedAt); err != nil {
			log.Printf("Scan showtime failed: %v", err)
			return nil, fmt.Errorf("scan showtime: %w", err)
		}
		list = append(list, s)
	}
	return list, nil
}
