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

	var email, role string
	err = conn.QueryRow(ctx, "SELECT email, role FROM users WHERE email='admin@gmail.com'").Scan(&email, &role)
	if err != nil {
		fmt.Println("Admin user not found")
	} else {
		fmt.Printf("Email: %s | Role: %s\n", email, role)
	}
}
