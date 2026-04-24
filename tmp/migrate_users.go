package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
)

func main() {
	dsn := "postgres://postgres:123456@localhost:5432/booking_cinema?sslmode=disable"
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		fmt.Printf("Connect failed: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(ctx)

	_, err = conn.Exec(ctx, "ALTER TABLE users ADD COLUMN IF NOT EXISTS avatar_url TEXT;")
	if err != nil {
		fmt.Printf("Alter failed: %v\n", err)
	} else {
		fmt.Println("Added avatar_url")
	}

	_, err = conn.Exec(ctx, "ALTER TABLE users ADD COLUMN IF NOT EXISTS is_active BOOLEAN DEFAULT TRUE;")
	if err != nil {
		fmt.Printf("Alter failed: %v\n", err)
	} else {
		fmt.Println("Added is_active")
	}
}
