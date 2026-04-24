package main

import (
	"context"
	"fmt"
	"log"

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
	rows, _ := conn.Query(ctx, "SELECT id, room_id FROM showtimes LIMIT 5")
	for rows.Next() {
		var id, roomID string
		rows.Scan(&id, &roomID)
		fmt.Printf("Showtime ID: %s | Room ID in table: %s\n", id, roomID)
	}
	rows.Close()

	fmt.Println("\n--- Screening Rooms table ---")
	rows, _ = conn.Query(ctx, "SELECT id, name FROM screening_rooms LIMIT 5")
	for rows.Next() {
		var id, name string
		rows.Scan(&id, &name)
		fmt.Printf("Room ID: %s | Name: %s\n", id, name)
	}
	rows.Close()
}
