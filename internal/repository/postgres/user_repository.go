package postgres

import (
	"booking_cinema_golang/internal/database"
	"booking_cinema_golang/internal/domain"
	"booking_cinema_golang/internal/repository"
	"context"
	"time"
)

type userRepository struct {
	db *database.DB
}

// NewUserRepository creates a new Postgres-backed UserRepository.
func NewUserRepository(db *database.DB) repository.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	const q = `
SELECT id, username, email, password_hash, full_name, phone,
       is_verified, otp_code, otp_expiry, created_at, updated_at
FROM users
WHERE id = $1
`
	u := &domain.User{}
	err := r.db.Pool.QueryRow(ctx, q, id).Scan(
		&u.ID,
		&u.Username,
		&u.Email,
		&u.PasswordHash,
		&u.FullName,
		&u.Phone,
		&u.IsVerified,
		&u.OTPCode,
		&u.OTPExpiry,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	const q = `
SELECT id, username, email, password_hash, full_name, phone,
       is_verified, otp_code, otp_expiry, created_at, updated_at
FROM users
WHERE email = $1
`
	u := &domain.User{}
	err := r.db.Pool.QueryRow(ctx, q, email).Scan(
		&u.ID,
		&u.Username,
		&u.Email,
		&u.PasswordHash,
		&u.FullName,
		&u.Phone,
		&u.IsVerified,
		&u.OTPCode,
		&u.OTPExpiry,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *userRepository) Create(ctx context.Context, u *domain.User) error {
	const q = `
INSERT INTO users (username, email, password_hash, full_name, phone, is_verified)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, created_at, updated_at
`
	return r.db.Pool.QueryRow(
		ctx, q,
		u.Username,
		u.Email,
		u.PasswordHash,
		u.FullName,
		u.Phone,
		u.IsVerified,
	).Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)
}

func (r *userRepository) UpdateOTP(ctx context.Context, userID, otpCode string, expiry time.Time) error {
	const q = `
UPDATE users
SET otp_code = $1,
    otp_expiry = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $3
`
	_, err := r.db.Pool.Exec(ctx, q, otpCode, expiry, userID)
	return err
}

func (r *userRepository) SetVerified(ctx context.Context, userID string, verified bool) error {
	const q = `
UPDATE users
SET is_verified = $1,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $2
`
	_, err := r.db.Pool.Exec(ctx, q, verified, userID)
	return err
}

