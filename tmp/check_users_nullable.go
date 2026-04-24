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

	rows, _ := conn.Query(ctx, "SELECT column_name, is_nullable FROM information_schema.columns WHERE table_name = 'users'")
	fmt.Println("Columns in 'users':")
	for rows.Next() {
		var name, nullable string
		rows.Scan(&name, &nullable)
		fmt.Printf("%s: %s\n", name, nullable)
	}
}
