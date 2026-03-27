package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"booking_cinema_golang/internal/config"
	"booking_cinema_golang/internal/domain"
	"booking_cinema_golang/internal/repository"
	"booking_cinema_golang/internal/utils"
	"golang.org/x/crypto/bcrypt"
)

type userService struct {
	userRepo repository.UserRepository
	cfg      *config.Config
}

var (
	// ErrUserNotFound is returned when user does not exist.
	ErrUserNotFound = errors.New("user not found")
	// ErrEmailAlreadyInUse is returned when email is already registered.
	ErrEmailAlreadyInUse = errors.New("email already in use")
)

// NewUserService creates a new UserService implementation.
func NewUserService(userRepo repository.UserRepository, cfg *config.Config) UserService {
	return &userService{
		userRepo: userRepo,
		cfg:      cfg,
	}
}

func (s *userService) GetUserByID(id int64) (*domain.User, error) {
	ctx := context.Background()

	if id <= 0 {
		return nil, fmt.Errorf("invalid user id")
	}

	u, err := s.userRepo.FindByID(ctx, strconv.FormatInt(id, 10))
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, ErrUserNotFound
	}
	return u, nil
}

func (s *userService) GetUserByEmail(email string) (*domain.User, error) {
	ctx := context.Background()

	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" {
		return nil, fmt.Errorf("email is required")
	}

	u, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, ErrUserNotFound
	}
	return u, nil
}

func (s *userService) CreateUser(user *domain.User) error {
	ctx := context.Background()

	if user == nil {
		return fmt.Errorf("user is required")
	}

	user.Email = strings.TrimSpace(strings.ToLower(user.Email))
	user.Username = strings.TrimSpace(user.Username)
	user.FullName = strings.TrimSpace(user.FullName)
	user.Phone = strings.TrimSpace(user.Phone)

	if user.Email == "" {
		return fmt.Errorf("email is required")
	}
	if user.Username == "" {
		user.Username = user.Email
	}

	// Ensure email uniqueness.
	if existing, _ := s.userRepo.FindByEmail(ctx, user.Email); existing != nil {
		return ErrEmailAlreadyInUse
	}

	// Hash password if caller provided plain text in PasswordHash field.
	if strings.TrimSpace(user.PasswordHash) == "" {
		return fmt.Errorf("password is required")
	}
	if !strings.HasPrefix(user.PasswordHash, "$2") {
		hash, err := bcrypt.GenerateFromPassword([]byte(user.PasswordHash), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		user.PasswordHash = string(hash)
	}

	// Default verification status.
	if user.ID == "" {
		user.IsVerified = false
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return err
	}

	// Optionally create OTP (same behavior as auth registration).
	if s.cfg != nil {
		if otp, err := utils.GenerateOTP(); err == nil {
			exp := utils.OTPExpiry(s.cfg.OTP.ExpiryMinutes)
			_ = s.userRepo.UpdateOTP(ctx, user.ID, otp, exp)
		}
	}

	return nil
}

func (s *userService) GetProfile(ctx context.Context, userID string) (*domain.User, error) {
	if userID == "" {
		return nil, fmt.Errorf("user id is required")
	}

	u, err := s.userRepo.FindByID(ctx, userID)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find profile: %w", err)
	}

	return u, nil
}

func (s *userService) UpdateProfile(ctx context.Context, userID, fullName, phone string) (*domain.User, error) {
	u, err := s.GetProfile(ctx, userID)
	if err != nil {
		return nil, err
	}

	u.FullName = strings.TrimSpace(fullName)
	u.Phone = strings.TrimSpace(phone)

	if err := s.userRepo.Update(ctx, u); err != nil {
		return nil, fmt.Errorf("update profile: %w", err)
	}

	return u, nil
}

func (s *userService) ChangePassword(ctx context.Context, userID, oldPassword, newPassword string) error {
	u, err := s.GetProfile(ctx, userID)
	if err != nil {
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(oldPassword)); err != nil {
		return errors.New("incorrect old password")
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), 12)
	if err != nil {
		return fmt.Errorf("hash new password: %w", err)
	}

	if err := s.userRepo.UpdatePassword(ctx, userID, string(newHash)); err != nil {
		return fmt.Errorf("update password: %w", err)
	}

	return nil
}
