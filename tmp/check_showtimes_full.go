package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
)

func main() {
	dsn := "postgres://postgres:123456@localhost:5432/booking_cinema?sslmode=disable"
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer conn.Close(ctx)

	fmt.Println("--- Showtimes table ---")
	rows, err := conn.Query(ctx, "SELECT id, movie_id, room_id, cinema_id, start_time, status FROM showtimes")
	if err != nil {
		log.Fatalf("Query failed: %v\n", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id, movieID, roomID, cinemaID, status string
		var startTime time.Time
		rows.Scan(&id, &movieID, &roomID, &cinemaID, &startTime, &status)
		fmt.Printf("ID: %s | Movie: %s | Room: %s | Cinema: %s | Start: %v | Status: %s\n", 
			id, movieID, roomID, cinemaID, startTime, status)
	}

	fmt.Println("\n--- Screening Rooms table ---")
	rowsR, _ := conn.Query(ctx, "SELECT id, cinema_id, name FROM screening_rooms")
	for rowsR.Next() {
		var id, cinemaID, name string
		rowsR.Scan(&id, &cinemaID, &name)
		fmt.Printf("ID: %s | Cinema: %s | Name: %s\n", id, cinemaID, name)
	}
}
