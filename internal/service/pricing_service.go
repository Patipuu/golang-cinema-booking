package service

import (
	"context"
	"time"

	"booking_cinema_golang/internal/domain"
)

type PricingService interface {
	CalculatePrice(ctx context.Context, showtime *domain.Showtime, seatType string) (float64, error)
}

type pricingService struct{}

func NewPricingService() PricingService {
	return &pricingService{}
}

func (s *pricingService) CalculatePrice(ctx context.Context, st *domain.Showtime, seatType string) (float64, error) {
	price := st.BasePrice

	// 1. Seat type modifier
	switch seatType {
	case "vip":
		price += st.BasePrice * 0.2 // +20%
	case "couple":
		price += st.BasePrice * 0.5 // +50%
	}

	// 2. Showtime modifier (Time-based)
	// Example: +10% on weekends, +5% for evening shows
	now := st.StartTime
	if now.Weekday() == time.Saturday || now.Weekday() == time.Sunday {
		price += st.BasePrice * 0.1
	}

	hour := now.Hour()
	if hour >= 18 {
		price += st.BasePrice * 0.05
	}

	return price, nil
}
