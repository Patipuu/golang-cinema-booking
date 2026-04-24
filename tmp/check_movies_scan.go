package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
)

func main() {
	conn, _ := pgx.Connect(context.Background(), "postgres://postgres:123456@localhost:5432/booking_cinema?sslmode=disable")
	defer conn.Close(context.Background())

	query := "SELECT id, title_vi, title_en, genre, director, cast_members, duration_mins, language, subtitle, rating_label, description, poster_url, trailer_url, status, avg_rating, created_at FROM movies ORDER BY created_at DESC LIMIT $1 OFFSET $2"
	args := []any{100, 0}

	rows, err := conn.Query(context.Background(), query, args...)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var ID, TitleVI, TitleEN, GenreStr, Director, CastMembers, Language, Subtitle, RatingLabel, Description, PosterURL, TrailerURL, Status string
		var DurationMins int
		var AvgRating float32
		var CreatedAt interface{}

		err := rows.Scan(&ID, &TitleVI, &TitleEN, &GenreStr, &Director, &CastMembers, &DurationMins, &Language, &Subtitle, &RatingLabel, &Description, &PosterURL, &TrailerURL, &Status, &AvgRating, &CreatedAt)
		if err != nil {
			fmt.Printf("Scan error for one row: %v\n", err)
			
			// Trying with pointers to check which fields are null
			var pTitleEN, pDirector, pCastMembers, pLanguage *string
			var pDesc, pPoster, pTrailer *string
			var pSub *string
			var pDur *int
			err2 := conn.QueryRow(context.Background(), "SELECT title_en, director, cast_members, language, subtitle, description, poster_url, trailer_url, duration_mins FROM movies WHERE id=$1", ID).Scan(&pTitleEN, &pDirector, &pCastMembers, &pLanguage, &pSub, &pDesc, &pPoster, &pTrailer, &pDur)
			if err2 == nil {
				fmt.Printf("NULL checks for id %s - En:%v, Dir:%v, Dur:%v\n", ID, pTitleEN==nil, pDirector==nil, pDur==nil)
			}
		} else {
			fmt.Printf("Success: %s - duration: %d\n", TitleVI, DurationMins)
		}
	}
}
