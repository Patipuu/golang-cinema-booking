package service

import "booking_cinema_golang/internal/domain"

// PaymentService defines payment operations (async with goroutine as per requirements).
type PaymentService interface {
	ProcessPayment(bookingID, paymentMethod string, amount float64) (*domain.Payment, error)
	GetPayment(id string) (*domain.Payment, error)
}
