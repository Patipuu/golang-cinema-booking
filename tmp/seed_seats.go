package main

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"
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

	// Fetch all rooms
	rows, err := conn.Query(ctx, "SELECT id FROM screening_rooms")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var roomIDs []string
	for rows.Next() {
		var id string
		rows.Scan(&id)
		roomIDs = append(roomIDs, id)
	}
	rows.Close()

	if len(roomIDs) == 0 {
		fmt.Println("No rooms found")
		return
	}

	rowsArr := []string{"A", "B", "C", "D", "E", "F", "G", "H"}
	cols := 10

	inserted := 0
	for _, roomID := range roomIDs {
		fmt.Printf("Inserting seats for room %s...\n", roomID)
		for _, rowName := range rowsArr {
			for c := 1; c <= cols; c++ {
				seatNumber := fmt.Sprintf("%s%d", rowName, c)
				seatType := "standard"
				if rowName == "G" || rowName == "H" {
					seatType = "vip"
				}

				id := uuid.New().String()
				_, err := conn.Exec(ctx, `
					INSERT INTO seats (id, room_id, seat_number, row_name, seat_type, status, created_at)
					VALUES ($1, $2, $3, $4, $5, 'available', NOW())
					ON CONFLICT DO NOTHING
				`, id, roomID, seatNumber, rowName, seatType)

				if err != nil {
					log.Printf("Error inserting seat %s: %v\n", seatNumber, err)
				} else {
					inserted++
				}
			}
		}
	}
	fmt.Printf("Done. Inserted %d seats.\n", inserted)
}
