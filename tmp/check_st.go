package main

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
)

func main() {
	conn, _ := pgx.Connect(context.Background(), "postgres://postgres:123456@localhost:5432/booking_cinema?sslmode=disable")
	defer conn.Close(context.Background())
	rows, _ := conn.Query(context.Background(), "SELECT id, movie_id, start_time, end_time, base_price, status FROM showtimes")
	for rows.Next() {
		var id, movieId, status string
        var start, end interface{}
        var price interface{}

		err := rows.Scan(&id, &movieId, &start, &end, &price, &status)
        if err != nil { fmt.Println("err:", err) }
		fmt.Printf("%s - start:%v end:%v base_price:%v\n", id, start, end, price)
	}
}
