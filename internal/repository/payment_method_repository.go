package repository

import (
	"booking_cinema_golang/internal/database"
	"booking_cinema_golang/internal/domain"
	"context"
)

type PaymentMethodRepository interface {
	// Nếu id rỗng, trả về tất cả; nếu có id, lọc theo id
	FindAll(ctx context.Context, id string) ([]domain.PaymentMethod, error)
}

type PaymentMethodRepositoryImpl struct {
	db *database.DB
}

func NewPaymentMethodRepository(db *database.DB) PaymentMethodRepository {
	return &PaymentMethodRepositoryImpl{db: db}
}

func (r *PaymentMethodRepositoryImpl) FindAll(ctx context.Context, id string) ([]domain.PaymentMethod, error) {
	query := "SELECT id, name, code, is_active, created_at FROM payment_methods WHERE is_active = TRUE"
	var args []interface{}
	if id != "" {
		query += " AND id = $1"
		args = append(args, id)
	}
	query += " ORDER BY created_at DESC"

	rows, err := r.db.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var methods []domain.PaymentMethod
	for rows.Next() {
		var m domain.PaymentMethod
		if err := rows.Scan(&m.ID, &m.Name, &m.Code, &m.IsActive, &m.CreatedAt); err != nil {
			return nil, err
		}
		methods = append(methods, m)
	}
	return methods, nil
}
