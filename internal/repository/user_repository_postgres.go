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
					 COALESCE(avatar_url, ''), COALESCE(role, 'customer'), COALESCE(is_active, true),
					 is_verified, COALESCE(otp_code, ''), otp_expiry, created_at, updated_at`

func (r *postgresUserRepo) scanUser(row pgx.Row) (*domain.User, error) {
	u := &domain.User{}
	err := row.Scan(
		&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.FullName, &u.Phone,
		&u.AvatarURL, &u.Role, &u.IsActive,
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

func (r *postgresUserRepo) Update(ctx context.Context, user *domain.User) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE users
		 SET full_name = $2, phone = $3, avatar_url = $4, updated_at = NOW()
		 WHERE id = $1`,
		user.ID, user.FullName, user.Phone, user.AvatarURL,
	)
	return err
}

func (r *postgresUserRepo) UpdatePassword(ctx context.Context, userID, newHash string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE users
		 SET password_hash = $2, updated_at = NOW()
		 WHERE id = $1`,
		userID, newHash,
	)
	return err
}

func (r *postgresUserRepo) FindAll(ctx context.Context, page, limit int, search string) ([]*domain.User, int, error) {
	offset := (page - 1) * limit
	if offset < 0 {
		offset = 0
	}

	searchQuery := "%" + search + "%"
	
	var total int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE email ILIKE $1 OR full_name ILIKE $1 OR username ILIKE $1`, searchQuery).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.pool.Query(ctx,
		`SELECT `+userColumns+` FROM users 
		 WHERE email ILIKE $1 OR full_name ILIKE $1 OR username ILIKE $1
		 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		searchQuery, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		u := &domain.User{}
		err := rows.Scan(
			&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.FullName, &u.Phone,
			&u.AvatarURL, &u.Role, &u.IsActive,
			&u.IsVerified, &u.OTPCode, &u.OTPExpiry, &u.CreatedAt, &u.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		users = append(users, u)
	}
	return users, total, rows.Err()
}

func (r *postgresUserRepo) UpdateStatus(ctx context.Context, userID string, isActive bool) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE users SET is_active = $2, updated_at = NOW() WHERE id = $1`,
		userID, isActive,
	)
	return err
}

func (r *postgresUserRepo) UpdateRole(ctx context.Context, userID string, role string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE users SET role = $2, updated_at = NOW() WHERE id = $1`,
		userID, role,
	)
	return err
}
