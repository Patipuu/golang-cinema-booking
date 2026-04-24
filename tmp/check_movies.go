package main

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
)

func main() {
	conn, _ := pgx.Connect(context.Background(), "postgres://postgres:123456@localhost:5432/booking_cinema?sslmode=disable")
	defer conn.Close(context.Background())
	rows, _ := conn.Query(context.Background(), "SELECT id, title_vi FROM movies")
	for rows.Next() {
		var id, title string
		rows.Scan(&id, &title)
		fmt.Printf("%s - %s\n", id, title)
	}
}
