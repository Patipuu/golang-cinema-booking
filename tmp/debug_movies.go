package main
import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
)
func main() {
	pool, err := pgxpool.New(context.Background(), "postgres://postgres:123456@localhost:5432/booking_cinema?sslmode=disable")
	if err != nil { panic(err) }
	
	rows, err := pool.Query(context.Background(), "SELECT id, title_vi, trailer_url FROM movies")
	if err != nil { panic(err) }
	defer rows.Close()
	
	for rows.Next() {
		var id, title string
		var trailer interface{}
		rows.Scan(&id, &title, &trailer)
		fmt.Printf("ID: %s, Title: %s, Trailer: %T\n", id, title, trailer)
	}
}
