package main

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
)

func main() {
	conn, _ := pgx.Connect(context.Background(), "postgres://postgres:123456@localhost:5432/booking_cinema?sslmode=disable")
	defer conn.Close(context.Background())
	rows, _ := conn.Query(context.Background(), "SELECT id, status, updated_at FROM bookings ORDER BY created_at DESC LIMIT 5")
	for rows.Next() {
		var id, status string
        var updated_at interface{}

		err := rows.Scan(&id, &status, &updated_at)
        if err != nil { fmt.Println("err:", err) }
		fmt.Printf("%s - status:%s, updated:%v\n", id, status, updated_at)
	}
}
