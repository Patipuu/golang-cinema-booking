package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"booking_cinema_golang/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	// ErrNotFound              = errors.New("user not found")
	ErrEmailAlreadyExists    = errors.New("email already exists")
	ErrUsernameAlreadyExists = errors.New("username already exists")
)

type postgresUserRepo struct {
	pool *pgxpool.Pool
}

// NewUserRepository creates a postgres-backed UserRepository.
func NewUserRepository(pool *pgxpool.Pool) UserRepository {
	return &postgresUserRepo{pool: pool}
}

const userColumns = `id, username, email, password_hash, full_name, phone,
					 is_verified, COALESCE(otp_code, ''), otp_expiry, created_at, updated_at`

func (r *postgresUserRepo) scanUser(row pgx.Row) (*domain.User, error) {
	u := &domain.User{}
	err := row.Scan(
		&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.FullName, &u.Phone,
		&u.IsVerified, &u.OTPCode, &u.OTPExpiry, &u.CreatedAt, &u.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return u, err
}

func (r *postgresUserRepo) FindByID(ctx context.Context, id string) (*domain.User, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT `+userColumns+` FROM users WHERE id = $1`, id)
	return r.scanUser(row)
}

func (r *postgresUserRepo) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT `+userColumns+` FROM users WHERE email = $1`, email)
	return r.scanUser(row)
}

func (r *postgresUserRepo) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT `+userColumns+` FROM users WHERE username = $1`, username)
	return r.scanUser(row)
}

func (r *postgresUserRepo) Create(ctx context.Context, user *domain.User) error {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO users (username, email, password_hash, full_name, phone)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, created_at, updated_at`,
		user.Username, user.Email, user.PasswordHash, user.FullName, user.Phone,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			if strings.Contains(pgErr.ConstraintName, "email") {
				return ErrEmailAlreadyExists
			}
			return ErrUsernameAlreadyExists
		}
		return err
	}
	return nil
}

func (r *postgresUserRepo) UpdateOTP(ctx context.Context, userID, otpCode string, expiry time.Time) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE users SET otp_code = $2, otp_expiry = $3, updated_at = NOW() WHERE id = $1`,
		userID, otpCode, expiry,
	)
	return err
}

func (r *postgresUserRepo) SetVerified(ctx context.Context, userID string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE users
		 SET is_verified = true, otp_code = NULL, otp_expiry = NULL, updated_at = NOW()
		 WHERE id = $1`,
		userID,
	)
	return err
}
