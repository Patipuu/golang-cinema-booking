package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"

	"booking_cinema_golang/internal/config"
	"booking_cinema_golang/internal/database"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config load: %v", err)
	}

	ctx := context.Background()
	db, err := database.Connect(ctx, cfg.Database.DSN())
	if err != nil {
		log.Fatalf("database connect: %v", err)
	}
	defer db.Close()

	// Read schema.sql
	schemaSQL, err := ioutil.ReadFile("schema.sql")
	if err != nil {
		log.Fatalf("read schema.sql: %v", err)
	}

	// Execute schema
	_, err = db.Pool.Exec(ctx, string(schemaSQL))
	if err != nil {
		log.Fatalf("execute schema: %v", err)
	}

	fmt.Println("Schema created successfully")

	// Read create_admin.sql
	adminSQL, err := ioutil.ReadFile("create_admin.sql")
	if err != nil {
		log.Fatalf("read create_admin.sql: %v", err)
	}

	// Execute admin creation
	_, err = db.Pool.Exec(ctx, string(adminSQL))
	if err != nil {
		log.Fatalf("execute admin creation: %v", err)
	}

	fmt.Println("Admin user created successfully")
}
