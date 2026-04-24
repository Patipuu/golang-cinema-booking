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

	rows, _ := conn.Query(ctx, "SELECT id, cinema_id, base_price, price FROM showtimes LIMIT 5")
	fmt.Println("Showtimes data:")
	for rows.Next() {
		var id string
		var cinemaID *string
		var basePrice *float64
		var price *float64
		rows.Scan(&id, &cinemaID, &basePrice, &price)
		
		cinID := "NULL"
		if cinemaID != nil { cinID = *cinemaID }
		
		bp := "NULL"
		if basePrice != nil { bp = fmt.Sprintf("%v", *basePrice) }
		
		p := "NULL"
		if price != nil { p = fmt.Sprintf("%v", *price) }
		
		fmt.Printf("ID: %s, CinemaID: %s, BasePrice: %s, Price: %s\n", id, cinID, bp, p)
	}
}
