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

	var count int
	if err := conn.QueryRow(ctx, "SELECT count(*) FROM seats").Scan(&count); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Total seats: %d\n", count)
}
