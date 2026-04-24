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

	rows, _ := conn.Query(ctx, "SELECT id, room_id, seat_number FROM seats LIMIT 5")
	fmt.Println("Seats data (with room):")
	for rows.Next() {
		var id string
		var roomID *string
		var seatNumber string
		rows.Scan(&id, &roomID, &seatNumber)
		fmt.Printf("ID: %s, RoomID: %v, SeatNumber: %s\n", id, roomID, seatNumber)
	}

	// Also check columns of 'seats' table
	rows2, _ := conn.Query(ctx, "SELECT column_name, data_type FROM information_schema.columns WHERE table_name = 'seats'")
	fmt.Println("\nColumns in 'seats':")
	for rows2.Next() {
		var name, dt string
		rows2.Scan(&name, &dt)
		fmt.Printf("%s (%s)\n", name, dt)
	}
}
