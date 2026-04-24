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

	fmt.Println("--- Recent Bookings ---")
	rows, err := conn.Query(ctx, `
		SELECT b.id, b.status, b.total_price, string_agg(s.seat_number, ', ') 
		FROM bookings b
		JOIN booking_seats bs ON bs.booking_id = b.id
		JOIN seats s ON s.id = bs.seat_id
		GROUP BY b.id, b.status, b.total_price, b.created_at
		ORDER BY b.created_at DESC LIMIT 10
	`)
	if err != nil {
		log.Fatalf("Query failed: %v\n", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id, status, seats string
		var price float64
		rows.Scan(&id, &status, &price, &seats)
		fmt.Printf("ID: %s | Status: %s | Price: %v | Seats: %s\n", id, status, price, seats)
	}
}
