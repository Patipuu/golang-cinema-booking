package main

import (
	"context"
	"fmt"

	"booking_cinema_golang/internal/config"
	"booking_cinema_golang/internal/database"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Config error: %v\n", err)
		return
	}
	db, err := database.Connect(context.Background(), cfg.Database.DSN())
	if err != nil {
		fmt.Printf("DB error: %v\n", err)
		return
	}
	defer db.Close()

	hash, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	
	query := `
		INSERT INTO users (username, email, password_hash, full_name, role, is_active, is_verified)
		VALUES ('admin', 'admin@cinema.local', $1, 'Admin', 'admin', true, true)
		ON CONFLICT (email) DO UPDATE SET 
			password_hash = $1, 
			role = 'admin',
			is_active = true,
			is_verified = true
	`
	_, err = db.Pool.Exec(context.Background(), query, string(hash))
	if err != nil {
		fmt.Printf("Update error: %v\n", err)
		return
	}
	fmt.Println("Admin user updated successfully.")
}
