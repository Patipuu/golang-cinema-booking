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

	_, err = conn.Exec(ctx, "INSERT INTO payment_methods (id, name, code, is_active) VALUES (gen_random_uuid(), 'Thanh toán tiền mặt', 'CASH', true) ON CONFLICT (code) DO NOTHING;")
	if err != nil {
		fmt.Printf("Insert failed: %v\n", err)
	} else {
		fmt.Println("Added CASH payment method")
	}
}
