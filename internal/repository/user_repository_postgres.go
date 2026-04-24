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

const userColumns = "id, username, email, password_hash, full_name, phone, avatar_url, role, is_active, is_verified, otp_code, otp_expiry, created_at, updated_at"

type postgresUserRepo struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) UserRepository {
	return &postgresUserRepo{pool: pool}
}

func (r *postgresUserRepo) scanUser(row pgx.Row) (*domain.User, error) {
	u := &domain.User{}
    var avatarURL, phone *string
	err := row.Scan(
		&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.FullName, &phone,
        &avatarURL, &u.Role, &u.IsActive, &u.IsVerified, &u.OTPCode, &u.OTPExpiry, &u.CreatedAt, &u.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
    if phone != nil {
        u.Phone = phone
    }
    if avatarURL != nil {
        u.AvatarURL = *avatarURL
    }
	return u, err
}

func (r *postgresUserRepo) FindByID(ctx context.Context, id string) (*domain.User, error) {
	row := r.pool.QueryRow(ctx, `SELECT `+userColumns+` FROM users WHERE id = $1`, id)
	return r.scanUser(row)
}

func (r *postgresUserRepo) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	row := r.pool.QueryRow(ctx, `SELECT `+userColumns+` FROM users WHERE email = $1`, email)
	return r.scanUser(row)
}

func (r *postgresUserRepo) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	row := r.pool.QueryRow(ctx, `SELECT `+userColumns+` FROM users WHERE username = $1`, username)
	return r.scanUser(row)
}

func (r *postgresUserRepo) Create(ctx context.Context, user *domain.User) error {
	if user.Role == "" {
		user.Role = "customer"
	}
	err := r.pool.QueryRow(ctx,
		`INSERT INTO users (username, email, password_hash, full_name, phone, avatar_url, role, is_active, is_verified)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		 RETURNING id, created_at, updated_at`,
		user.Username, user.Email, user.PasswordHash, user.FullName, user.Phone, user.AvatarURL, user.Role, user.IsActive, user.IsVerified,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			if strings.Contains(pgErr.ConstraintName, "email") {
				return ErrEmailExists
			}
			return ErrUsernameExists
		}
		return err
	}
	return nil
}

func (r *postgresUserRepo) Update(ctx context.Context, user *domain.User) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE users SET full_name = $1, phone = $2, avatar_url = $3, updated_at = NOW() WHERE id = $4`,
		user.FullName, user.Phone, user.AvatarURL, user.ID)
	return err
}

func (r *postgresUserRepo) UpdatePassword(ctx context.Context, userID, newHash string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE users SET password_hash = $1, updated_at = NOW() WHERE id = $2`,
		newHash, userID)
	return err
}

func (r *postgresUserRepo) UpdateStatus(ctx context.Context, userID string, isActive bool) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE users SET is_active = $1, updated_at = NOW() WHERE id = $2`,
		isActive, userID)
	return err
}

func (r *postgresUserRepo) FindAll(ctx context.Context, page, limit int, search string) ([]*domain.User, int, error) {
	offset := (page - 1) * limit
	query := `SELECT ` + userColumns + ` FROM users WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM users WHERE 1=1`
	var args []interface{}
	idx := 1

	if search != "" {
		query += ` AND (LOWER(email) LIKE $` + string(rune(idx+'0')) + ` OR LOWER(full_name) LIKE $` + string(rune(idx+'0')) + `)`
		countQuery += ` AND (LOWER(email) LIKE $` + string(rune(idx+'0')) + ` OR LOWER(full_name) LIKE $` + string(rune(idx+'0')) + `)`
		args = append(args, "%"+strings.ToLower(search)+"%")
		idx++
	}

	query += ` ORDER BY created_at DESC LIMIT $` + string(rune(idx+'0')) + ` OFFSET $` + string(rune(idx+1+'0'))
	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		u, err := r.scanUser(rows)
		if err != nil {
			return nil, 0, err
		}
		users = append(users, u)
	}

	var total int
	_ = r.pool.QueryRow(ctx, countQuery, args[:len(args)-2]...).Scan(&total)

	return users, total, nil
}

func (r *postgresUserRepo) UpdateOTP(ctx context.Context, id, otp string, expiry time.Time) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE users SET otp_code = $1, otp_expiry = $2, updated_at = NOW() WHERE id = $3`,
		otp, expiry, id)
	return err
}

func (r *postgresUserRepo) SetVerified(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE users SET is_verified = TRUE, otp_code = NULL, otp_expiry = NULL, updated_at = NOW() WHERE id = $1`,
		id)
	return err
}

func (r *postgresUserRepo) ListAll(ctx context.Context, page domain.Page) ([]domain.User, domain.PageResult, error) {
	offset := (page.Page - 1) * page.Limit
	rows, err := r.pool.Query(ctx,
		`SELECT `+userColumns+` FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2`,
		page.Limit, offset)
	if err != nil {
		return nil, domain.PageResult{}, err
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		u, err := r.scanUser(rows)
		if err != nil {
			return nil, domain.PageResult{}, err
		}
		users = append(users, *u)
	}

	var total int
	_ = r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&total)

	return users, domain.PageResult{
		TotalCount: total,
		Page:       page.Page,
		Limit:      page.Limit,
	}, nil
}

func (r *postgresUserRepo) UpdateRole(ctx context.Context, id, role string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE users SET role = $1, updated_at = NOW() WHERE id = $2`,
		role, id)
	return err
}
