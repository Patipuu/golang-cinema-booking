package repository

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"booking_cinema_golang/internal/config"
	"booking_cinema_golang/internal/database"
	"booking_cinema_golang/internal/domain"
)

func TestBookingCreateRaceCondition(t *testing.T) {
	// Integration test: requires a reachable PostgreSQL configured by internal/config.
	loadDotEnvForTest(t)
	cfg, err := config.Load()
	if err != nil {
		t.Skipf("config load failed: %v", err)
	}
	db, err := database.Connect(context.Background(), cfg.Database.DSN())
	if err != nil {
		t.Skipf("db connect failed (skipping integration test): %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	_, _ = db.Pool.Exec(ctx, `CREATE EXTENSION IF NOT EXISTS pgcrypto;`)

	repo := NewBookingRepository(db)

	runID := fmt.Sprintf("%d", time.Now().UTC().UnixNano())

	userID1 := mustInsertUser(t, ctx, db, "user_race_1_"+runID, "user1_race_"+runID+"@example.com")
	userID2 := mustInsertUser(t, ctx, db, "user_race_2_"+runID, "user2_race_"+runID+"@example.com")
	cinemaID := mustInsertCinema(t, ctx, db, "cinema_race_"+runID, "Somewhere", "CityRace", 10)
	movieID := mustInsertMovie(t, ctx, db, "movie_race_"+runID, "genre", 90, "G")

	// Seats.
	mustInsertSeat(t, ctx, db, cinemaID, "A1", "A", "regular")

	// Showtime.
	showtimeID := mustInsertShowtime(t, ctx, db, cinemaID, movieID, time.Now().UTC().Format("2006-01-02"), "19:00:00", 75000)

	seatCodes := []string{"A1"}

	start := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(2)

	var mu sync.Mutex
	errs := make([]error, 0, 2)
	successCount := 0

	run := func(userID string) {
		defer wg.Done()
		<-start
		b := &domain.Booking{
			UserID:     userID,
			ShowtimeID: showtimeID,
			Status:     "pending",
		}
		err := repo.Create(ctx, b, seatCodes)
		mu.Lock()
		defer mu.Unlock()
		errs = append(errs, err)
		if err == nil {
			successCount++
		}
	}

	go run(userID1)
	go run(userID2)

	close(start)
	wg.Wait()

	// Exactly one should succeed.
	if successCount != 1 {
		t.Fatalf("expected exactly 1 successful booking, got %d (errors: %v)", successCount, summarizeErrors(errs))
	}

	// Ensure seat A1 is taken only once for this showtime.
	taken, err := repo.GetTakenSeatIDsForShowtime(ctx, showtimeID)
	if err != nil {
		t.Fatalf("GetTakenSeatIDsForShowtime failed: %v", err)
	}
	if !contains(taken, "A1") {
		t.Fatalf("expected taken seats to contain A1, got: %v", taken)
	}

	// Best-effort cleanup.
	_, _ = db.Pool.Exec(ctx, `DELETE FROM showtimes WHERE id = $1`, showtimeID)

	// Optional hint if integration data has stale rows.
	_ = os.Setenv("INTEGRATION_TEST_DONE", strings.TrimSpace(time.Now().String()))
}

func loadDotEnvForTest(t *testing.T) {
	t.Helper()

	// config.Load() loads ".env" relative to the current working directory.
	// When `go test ./...` runs, the working directory is per-package, so we
	// proactively load repo-root .env to make the integration test stable.
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		return
	}
	repoRoot := filepath.Clean(filepath.Join(filepath.Dir(thisFile), "..", ".."))
	dotEnvPath := filepath.Join(repoRoot, ".env")

	f, err := os.Open(dotEnvPath)
	if err != nil {
		// If user already sets DB_* via environment, it's fine to skip silently.
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])

		// Don't override already-set env vars.
		if _, alreadySet := os.LookupEnv(key); alreadySet {
			continue
		}
		_ = os.Setenv(key, val)
	}
}

func mustInsertUser(t *testing.T, ctx context.Context, db *database.DB, username, email string) string {
	t.Helper()
	var id string
	err := db.Pool.QueryRow(ctx, `
		INSERT INTO users (username, email, password_hash, full_name, phone, is_verified)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id::text
	`, username, email, "x", "Race User", "000", false).Scan(&id)
	if err != nil {
		t.Fatalf("insert user failed: %v", err)
	}
	return id
}

func mustInsertCinema(t *testing.T, ctx context.Context, db *database.DB, name, location, city string, totalSeats int) string {
	t.Helper()
	var id string
	err := db.Pool.QueryRow(ctx, `
		INSERT INTO cinemas (name, location, city, total_seats)
		VALUES ($1, $2, $3, $4)
		RETURNING id::text
	`, name, location, city, totalSeats).Scan(&id)
	if err != nil {
		t.Fatalf("insert cinema failed: %v", err)
	}
	return id
}

func mustInsertSeat(t *testing.T, ctx context.Context, db *database.DB, cinemaID, seatNumber, rowName, seatType string) {
	t.Helper()
	_, err := db.Pool.Exec(ctx, `
		INSERT INTO seats (cinema_id, seat_number, row_name, seat_type)
		VALUES ($1, $2, $3, $4)
	`, cinemaID, seatNumber, rowName, seatType)
	if err != nil {
		t.Fatalf("insert seat failed: %v", err)
	}
}

func mustInsertMovie(t *testing.T, ctx context.Context, db *database.DB, title, genre string, duration int, rating string) string {
	t.Helper()
	var id string
	err := db.Pool.QueryRow(ctx, `
		INSERT INTO movies (title, genre, duration, rating)
		VALUES ($1, $2, $3, $4)
		RETURNING id::text
	`, title, genre, duration, rating).Scan(&id)
	if err != nil {
		t.Fatalf("insert movie failed: %v", err)
	}
	return id
}

func mustInsertShowtime(t *testing.T, ctx context.Context, db *database.DB, cinemaID, movieID, showDate, showTime string, price float64) string {
	t.Helper()
	var id string
	err := db.Pool.QueryRow(ctx, `
		INSERT INTO showtimes (cinema_id, movie_id, show_date, show_time, price)
		VALUES ($1, $2, $3::date, $4::time, $5)
		RETURNING id::text
	`, cinemaID, movieID, showDate, showTime, price).Scan(&id)
	if err != nil {
		t.Fatalf("insert showtime failed: %v", err)
	}
	return id
}

func contains(arr []string, target string) bool {
	for _, v := range arr {
		if v == target {
			return true
		}
	}
	return false
}

func summarizeErrors(errs []error) []string {
	out := make([]string, 0, len(errs))
	for _, e := range errs {
		if e == nil {
			out = append(out, "nil")
			continue
		}
		// Keep it compact for CI logs.
		if errors.Is(e, ErrSeatLockConflict) {
			out = append(out, "ErrSeatLockConflict")
			continue
		}
		if errors.Is(e, ErrSeatAlreadyTaken) {
			out = append(out, "ErrSeatAlreadyTaken")
			continue
		}
		out = append(out, e.Error())
	}
	return out
}

