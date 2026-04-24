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

	rows, _ := conn.Query(ctx, "SELECT id, room_id FROM showtimes")
	fmt.Println("Showtimes rooms:")
	for rows.Next() {
		var id string
		var roomID string
		rows.Scan(&id, &roomID)
		fmt.Printf("Showtime ID: %s, RoomID: %s\n", id, roomID)
	}
}
