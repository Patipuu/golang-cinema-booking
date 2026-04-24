package repository

import (
	"context"
	"time"

	"booking_cinema_golang/internal/domain"
)

// PaymentRepository defines data access methods for payments.
type PaymentRepository interface {
	// Basic CRUD
    FindByID(ctx context.Context, id string) (*domain.Payment, error)
    FindByBookingID(ctx context.Context, bookingID string) (*domain.Payment, error)
    FindByTransactionID(ctx context.Context, transactionID string) (*domain.Payment, error)
    FindByStatus(ctx context.Context, status string, limit, offset int) ([]*domain.Payment, error)
    
    Create(ctx context.Context, payment *domain.Payment) error
    Update(ctx context.Context, payment *domain.Payment) error
    UpdateStatus(ctx context.Context, id, status string) error
    UpdateStatusByTransactionID(ctx context.Context, transactionID, status string) error
    
    // Batch operations
    BatchCreate(ctx context.Context, payments []*domain.Payment) error
    
    // Stats
    GetPaymentStats(ctx context.Context, from, to time.Time) (map[string]interface{}, error)

}
