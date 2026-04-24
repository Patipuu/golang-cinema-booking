package main

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"log"
)

func main() {
	conn, err := pgx.Connect(context.Background(), "postgres://postgres:123456@localhost:5432/booking_cinema?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close(context.Background())
	
	fmt.Println("--- Movies Table ---")
	rows, _ := conn.Query(context.Background(), "SELECT column_name, data_type FROM information_schema.columns WHERE table_name = 'movies'")
	for rows.Next() {
		var name, dtype string
		rows.Scan(&name, &dtype)
		fmt.Printf("%s: %s\n", name, dtype)
	}

	fmt.Println("\n--- Showtimes Table ---")
	rows, _ = conn.Query(context.Background(), "SELECT column_name, data_type FROM information_schema.columns WHERE table_name = 'showtimes'")
	for rows.Next() {
		var name, dtype string
		rows.Scan(&name, &dtype)
		fmt.Printf("%s: %s\n", name, dtype)
	}
}
