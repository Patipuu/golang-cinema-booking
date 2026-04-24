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

	rows, _ := conn.Query(ctx, "SELECT username, email, role, is_verified FROM users")
	fmt.Println("--- USERS ---")
	count := 0
	for rows.Next() {
		var u, e, r string
		var v bool
		rows.Scan(&u, &e, &r, &v)
		fmt.Printf("User: %s | Email: %s | Role: %s | Verified: %v\n", u, e, r, v)
		count++
	}
	if count == 0 {
		fmt.Println("No users found in database.")
	}
}
