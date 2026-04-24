package main

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	dsn := "postgres://postgres:123456@localhost:5432/booking_cinema?sslmode=disable"
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer conn.Close(ctx)

	// Updated to include @gmail.com for both username and email to satisfy user request
	username := "user@gmail.com"
	email := "user@gmail.com"
	password := "password123"
	role := "customer"

	hash, _ := bcrypt.GenerateFromPassword([]byte(password), 12)

	id := uuid.New().String()

	_, err = conn.Exec(ctx, `
		INSERT INTO users (id, username, email, full_name, phone, password_hash, role, is_verified, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, true, NOW(), NOW())
		ON CONFLICT (username) DO UPDATE 
		SET email = EXCLUDED.email, password_hash = EXCLUDED.password_hash, role = EXCLUDED.role, is_verified = true, phone = EXCLUDED.phone
	`, id, username, email, "Regular User", "0987654321", string(hash), role)

	if err != nil {
		log.Fatalf("Failed to create user: %v\n", err)
	}

	fmt.Printf("User created successfully!\nUsername/Email: %s\nPassword: %s\nRole: %s\n", username, password, role)
}
