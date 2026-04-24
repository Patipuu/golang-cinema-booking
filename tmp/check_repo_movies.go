package main

import (
	"context"
	"fmt"
	"booking_cinema_golang/internal/database"
	"booking_cinema_golang/internal/domain"
	"booking_cinema_golang/internal/repository"
)

func main() {
	db, err := database.NewDB("postgres://postgres:123456@localhost:5432/booking_cinema?sslmode=disable")
	if err != nil {
		fmt.Println(err)
		return
	}
	repo := repository.NewCatalogRepository(db)
	
	// Test ""
	list, _, err := repo.ListMovies(context.Background(), "", domain.Page{Limit: 100, Page: 1})
	fmt.Printf("Status '': Got %d movies, err=%v\n", len(list), err)
	
	// Test "now_showing"
	list2, _, err2 := repo.ListMovies(context.Background(), "now_showing", domain.Page{Limit: 100, Page: 1})
	fmt.Printf("Status 'now_showing': Got %d movies, err=%v\n", len(list2), err2)
}
