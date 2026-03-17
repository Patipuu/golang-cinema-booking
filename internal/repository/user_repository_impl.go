package repository

import (
	"context"
	"fmt"
	"time"

	"booking_cinema_golang/internal/database"
	"booking_cinema_golang/internal/domain"
	"github.com/jackc/pgx/v5"
)

type userRepository struct {
	db *database.DB
}

func (r *userRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	if r.db == nil || r.db.Pool == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	query := `
		SELECT
			id, username, email, password_hash, full_name, phone,
			is_verified, otp_code, otp_expiry, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	var u domain.User
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
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
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("query user by id: %w", err)
	}

	return &u, nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	if r.db == nil || r.db.Pool == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	query := `
		SELECT
			id, username, email, password_hash, full_name, phone,
			is_verified, otp_code, otp_expiry, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	var u domain.User
	err := r.db.Pool.QueryRow(ctx, query, email).Scan(
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
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("query user by email: %w", err)
	}

	return &u, nil
}

func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	if r.db == nil || r.db.Pool == nil {
		return fmt.Errorf("database connection is nil")
	}
	if user == nil {
		return fmt.Errorf("user is nil")
	}

	query := `
		INSERT INTO users (username, email, password_hash, full_name, phone, is_verified, otp_code, otp_expiry)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at
	`

	err := r.db.Pool.QueryRow(
		ctx,
		query,
		user.Username,
		user.Email,
		user.PasswordHash,
		user.FullName,
		user.Phone,
		user.IsVerified,
		nullIfEmpty(user.OTPCode),
		user.OTPExpiry,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return fmt.Errorf("insert user: %w", err)
	}

	return nil
}

func (r *userRepository) UpdateOTP(ctx context.Context, userID, otpCode string, expiry time.Time) error {
	if r.db == nil || r.db.Pool == nil {
		return fmt.Errorf("database connection is nil")
	}

	query := `
		UPDATE users
		SET otp_code = $2, otp_expiry = $3, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`
	_, err := r.db.Pool.Exec(ctx, query, userID, otpCode, expiry)
	if err != nil {
		return fmt.Errorf("update otp: %w", err)
	}
	return nil
}

func (r *userRepository) SetVerified(ctx context.Context, userID string, verified bool) error {
	if r.db == nil || r.db.Pool == nil {
		return fmt.Errorf("database connection is nil")
	}

	query := `
		UPDATE users
		SET is_verified = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`
	_, err := r.db.Pool.Exec(ctx, query, userID, verified)
	if err != nil {
		return fmt.Errorf("set verified: %w", err)
	}
	return nil
}

// NewUserRepository constructs a repository that satisfies UserRepository.
func NewUserRepository(db *database.DB) UserRepository {
	return &userRepository{db: db}
}

func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}
