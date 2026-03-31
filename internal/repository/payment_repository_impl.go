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
    query := `SELECT id, booking_id, payment_method, amount, status, transaction_id, paid_at, created_at, updated_at 
              FROM payments WHERE id = $1`

    var p domain.Payment
    var paidAt *time.Time

    err := r.db.Pool.QueryRow(ctx, query, id).Scan(
        &p.ID, &p.BookingID, &p.PaymentMethod, &p.Amount,
        &p.Status, &p.TransactionID, &paidAt, &p.CreatedAt, &p.UpdatedAt,
    )

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
    query := `SELECT id, booking_id, payment_method, amount, status, transaction_id, paid_at, created_at, updated_at 
              FROM payments WHERE booking_id = $1 ORDER BY created_at DESC LIMIT 1`

    var p domain.Payment
    var paidAt *time.Time

    err := r.db.Pool.QueryRow(ctx, query, bookingID).Scan(
        &p.ID, &p.BookingID, &p.PaymentMethod, &p.Amount,
        &p.Status, &p.TransactionID, &paidAt, &p.CreatedAt, &p.UpdatedAt,
    )

    if err != nil {
        if err.Error() == "no rows in result set" {
            return nil, nil
        }
        return nil, err
    }

    p.PaidAt = paidAt
    return &p, nil
}

func (r *PaymentRepositoryImpl) FindByTransactionID(ctx context.Context, transactionID string) (*domain.Payment, error) {
    query := `SELECT id, booking_id, payment_method, amount, status, transaction_id, paid_at, created_at, updated_at 
              FROM payments WHERE transaction_id = $1`

    var p domain.Payment
    var paidAt *time.Time

    err := r.db.Pool.QueryRow(ctx, query, transactionID).Scan(
        &p.ID, &p.BookingID, &p.PaymentMethod, &p.Amount,
        &p.Status, &p.TransactionID, &paidAt, &p.CreatedAt, &p.UpdatedAt,
    )

    if err != nil {
        if err.Error() == "no rows in result set" {
            return nil, nil
        }
        return nil, err
    }

    p.PaidAt = paidAt
    return &p, nil
}

func (r *PaymentRepositoryImpl) FindByStatus(ctx context.Context, status string, limit, offset int) ([]*domain.Payment, error) {
    query := `SELECT id, booking_id, payment_method, amount, status, transaction_id, paid_at, created_at, updated_at 
              FROM payments WHERE status = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`

    rows, err := r.db.Pool.Query(ctx, query, status, limit, offset)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var payments []*domain.Payment
    for rows.Next() {
        var p domain.Payment
        var paidAt *time.Time

        err := rows.Scan(
            &p.ID, &p.BookingID, &p.PaymentMethod, &p.Amount,
            &p.Status, &p.TransactionID, &paidAt, &p.CreatedAt, &p.UpdatedAt,
        )
        if err != nil {
            return nil, err
        }

        p.PaidAt = paidAt
        payments = append(payments, &p)
    }

    return payments, nil
}

func (r *PaymentRepositoryImpl) Create(ctx context.Context, payment *domain.Payment) error {
    now := time.Now()
    payment.CreatedAt = now
    payment.UpdatedAt = now

    query := `INSERT INTO payments (id, booking_id, payment_method, amount, status, transaction_id, paid_at, created_at, updated_at) 
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

    _, err := r.db.Pool.Exec(ctx, query,
        payment.ID, payment.BookingID, payment.PaymentMethod, payment.Amount,
        payment.Status, payment.TransactionID, payment.PaidAt, payment.CreatedAt, payment.UpdatedAt,
    )

    return err
}

func (r *PaymentRepositoryImpl) BatchCreate(ctx context.Context, payments []*domain.Payment) error {
    if len(payments) == 0 {
        return nil
    }

    tx, err := r.db.Pool.Begin(ctx)
    if err != nil {
        return err
    }
    defer tx.Rollback(ctx)

    now := time.Now()
    query := `INSERT INTO payments (id, booking_id, payment_method, amount, status, transaction_id, paid_at, created_at, updated_at) 
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

    for _, p := range payments {
        p.CreatedAt = now
        p.UpdatedAt = now

        _, err := tx.Exec(ctx, query,
            p.ID, p.BookingID, p.PaymentMethod, p.Amount,
            p.Status, p.TransactionID, p.PaidAt, p.CreatedAt, p.UpdatedAt,
        )
        if err != nil {
            return err
        }
    }

    return tx.Commit(ctx)
}

func (r *PaymentRepositoryImpl) Update(ctx context.Context, payment *domain.Payment) error {
    payment.UpdatedAt = time.Now()

    query := `UPDATE payments 
              SET booking_id = $1, payment_method = $2, amount = $3, status = $4, 
                  transaction_id = $5, paid_at = $6, updated_at = $7 
              WHERE id = $8`

    _, err := r.db.Pool.Exec(ctx, query,
        payment.BookingID, payment.PaymentMethod, payment.Amount, payment.Status,
        payment.TransactionID, payment.PaidAt, payment.UpdatedAt, payment.ID,
    )

    return err
}

func (r *PaymentRepositoryImpl) UpdateStatus(ctx context.Context, id, status string) error {
    now := time.Now()
    var paidAt *time.Time

    if status == "paid" || status == "completed" {
        paidAt = &now
    }

    query := `UPDATE payments SET status = $1, paid_at = $2, updated_at = $3 WHERE id = $4`
    _, err := r.db.Pool.Exec(ctx, query, status, paidAt, now, id)

    return err
}

func (r *PaymentRepositoryImpl) UpdateStatusByTransactionID(ctx context.Context, transactionID, status string) error {
    now := time.Now()
    var paidAt *time.Time

    if status == "paid" || status == "completed" {
        paidAt = &now
    }

    query := `UPDATE payments SET status = $1, paid_at = $2, updated_at = $3 WHERE transaction_id = $4`
    _, err := r.db.Pool.Exec(ctx, query, status, paidAt, now, transactionID)

    return err
}

func (r *PaymentRepositoryImpl) GetPaymentStats(ctx context.Context, from, to time.Time) (map[string]interface{}, error) {
    query := `
        SELECT 
            COUNT(*) as total_count,
            COALESCE(SUM(amount), 0) as total_amount,
            COUNT(CASE WHEN status = 'paid' THEN 1 END) as paid_count,
            COALESCE(SUM(CASE WHEN status = 'paid' THEN amount ELSE 0 END), 0) as paid_amount,
            COUNT(CASE WHEN status = 'pending' THEN 1 END) as pending_count,
            COUNT(CASE WHEN status = 'failed' THEN 1 END) as failed_count
        FROM payments
        WHERE created_at BETWEEN $1 AND $2
    `

    var stats struct {
        TotalCount   int64
        TotalAmount  float64
        PaidCount    int64
        PaidAmount   float64
        PendingCount int64
        FailedCount  int64
    }

    err := r.db.Pool.QueryRow(ctx, query, from, to).Scan(
        &stats.TotalCount, &stats.TotalAmount,
        &stats.PaidCount, &stats.PaidAmount,
        &stats.PendingCount, &stats.FailedCount,
    )

    if err != nil {
        return nil, err
    }

    return map[string]interface{}{
        "total_count":   stats.TotalCount,
        "total_amount":  stats.TotalAmount,
        "paid_count":    stats.PaidCount,
        "paid_amount":   stats.PaidAmount,
        "pending_count": stats.PendingCount,
        "failed_count":  stats.FailedCount,
        "from":          from,
        "to":            to,
    }, nil
}