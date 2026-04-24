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
	
	fmt.Println("--- Columns in Showtimes ---")
	rows, _ := conn.Query(context.Background(), "SELECT column_name, is_nullable FROM information_schema.columns WHERE table_name = 'showtimes'")
	for rows.Next() {
		var name, nullable string
		rows.Scan(&name, &nullable)
		fmt.Printf("%s: nullable=%s\n", name, nullable)
	}

	fmt.Println("\n--- Columns in Movies ---")
	rows, _ = conn.Query(context.Background(), "SELECT column_name, is_nullable FROM information_schema.columns WHERE table_name = 'movies'")
	for rows.Next() {
		var name, nullable string
		rows.Scan(&name, &nullable)
		fmt.Printf("%s: nullable=%s\n", name, nullable)
	}
}
