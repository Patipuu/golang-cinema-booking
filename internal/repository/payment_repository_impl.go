package repository

import (
    "context"
    "time"

    "booking_cinema_golang/internal/database"
    "booking_cinema_golang/internal/domain"
)

type PaymentRepositoryImpl struct {
    db *database.DB
}

func NewPaymentRepository(db *database.DB) PaymentRepository {
    return &PaymentRepositoryImpl{db: db}
}

func (r *PaymentRepositoryImpl) FindByID(ctx context.Context, id string) (*domain.Payment, error) {
    var p domain.Payment
    var paidAt *time.Time
    err := r.db.Pool.QueryRow(ctx, "SELECT id, booking_id, payment_method, amount, status, transaction_id, paid_at, created_at, updated_at FROM payments WHERE id = $1", id).Scan(&p.ID, &p.BookingID, &p.PaymentMethod, &p.Amount, &p.Status, &p.TransactionID, &paidAt, &p.CreatedAt, &p.UpdatedAt)
    if err != nil {
        if err.Error() == "no rows in result set" {
            return nil, nil
        }
        return nil, err
    }
    p.PaidAt = paidAt
    return &p, nil
}

func (r *PaymentRepositoryImpl) FindByBookingID(ctx context.Context, bookingID string) (*domain.Payment, error) {
    var p domain.Payment
    var paidAt *time.Time
    err := r.db.Pool.QueryRow(ctx, "SELECT id, booking_id, payment_method, amount, status, transaction_id, paid_at, created_at, updated_at FROM payments WHERE booking_id = $1", bookingID).Scan(&p.ID, &p.BookingID, &p.PaymentMethod, &p.Amount, &p.Status, &p.TransactionID, &paidAt, &p.CreatedAt, &p.UpdatedAt)
    if err != nil {
        if err.Error() == "no rows in result set" {
            return nil, nil
        }
        return nil, err
    }
    p.PaidAt = paidAt
    return &p, nil
}

func (r *PaymentRepositoryImpl) Create(ctx context.Context, payment *domain.Payment) error {
    now := time.Now()
    payment.CreatedAt = now
    payment.UpdatedAt = now
    _, err := r.db.Pool.Exec(ctx, "INSERT INTO payments (id, booking_id, payment_method, amount, status, transaction_id, paid_at, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)", payment.ID, payment.BookingID, payment.PaymentMethod, payment.Amount, payment.Status, payment.TransactionID, payment.PaidAt, payment.CreatedAt, payment.UpdatedAt)
    return err
}

func (r *PaymentRepositoryImpl) Update(ctx context.Context, payment *domain.Payment) error {
    payment.UpdatedAt = time.Now()
    _, err := r.db.Pool.Exec(ctx, "UPDATE payments SET booking_id = $1, payment_method = $2, amount = $3, status = $4, transaction_id = $5, paid_at = $6, updated_at = $7 WHERE id = $8", payment.BookingID, payment.PaymentMethod, payment.Amount, payment.Status, payment.TransactionID, payment.PaidAt, payment.UpdatedAt, payment.ID)
    return err
}

func (r *PaymentRepositoryImpl) UpdateStatus(ctx context.Context, id, status string) error {
    now := time.Now()
    var paidAt *time.Time
    if status == "paid" {
        paidAt = &now
    }
    _, err := r.db.Pool.Exec(ctx, "UPDATE payments SET status = $1, paid_at = $2, updated_at = $3 WHERE id = $4", status, paidAt, now, id)
    return err
}