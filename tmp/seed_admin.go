package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
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

	email := "admin@gmail.com"
	password := "password123"
	role := "admin"

	hash, _ := bcrypt.GenerateFromPassword([]byte(password), 12)

	_, err = conn.Exec(ctx, `
		INSERT INTO users (username, email, password_hash, full_name, role, is_verified)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (email) DO UPDATE 
		SET role = EXCLUDED.role, password_hash = EXCLUDED.password_hash, is_verified = TRUE
	`, "admin", email, string(hash), "System Admin", role, true)

	if err != nil {
		fmt.Printf("Seed failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Admin user seeded successfully!")
}
