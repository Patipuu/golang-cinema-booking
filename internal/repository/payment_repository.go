package repository

import (
	"context"

	"booking_cinema_golang/internal/domain"
)

// PaymentRepository defines data access methods for payments.
type PaymentRepository interface {
	FindByID(ctx context.Context, id string) (*domain.Payment, error)
	FindByBookingID(ctx context.Context, bookingID string) (*domain.Payment, error)
	Create(ctx context.Context, payment *domain.Payment) error
}
