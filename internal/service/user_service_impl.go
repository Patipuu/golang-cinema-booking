package service

import (
	"context"
	"errors"
	"fmt"
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

func (s *userService) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	if id == "" {
		return nil, fmt.Errorf("invalid user id")
	}

	u, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, ErrUserNotFound
	}
	return u, nil
}

func (s *userService) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
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

func (s *userService) CreateUser(ctx context.Context, user *domain.User) error {
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

func (s *userService) UpdateProfile(ctx context.Context, userID, fullName, phone, avatarURL string) (*domain.User, error) {
	u, err := s.GetProfile(ctx, userID)
	if err != nil {
		return nil, err
	}

	u.FullName = strings.TrimSpace(fullName)
	u.Phone = strings.TrimSpace(phone)
	u.AvatarURL = strings.TrimSpace(avatarURL)

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

func (s *userService) FindAllUsers(ctx context.Context, page, limit int, search string) ([]*domain.User, *domain.PageResult, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	users, total, err := s.userRepo.FindAll(ctx, page, limit, search)
	if err != nil {
		return nil, nil, fmt.Errorf("find all users: %w", err)
	}

	pageResult := &domain.PageResult{
		Page:       page,
		Limit:      limit,
		TotalCount: total,
	}

	return users, pageResult, nil
}

func (s *userService) UpdateUserStatus(ctx context.Context, userID string, isActive bool) error {
	if userID == "" {
		return fmt.Errorf("user id is required")
	}
	if err := s.userRepo.UpdateStatus(ctx, userID, isActive); err != nil {
		return fmt.Errorf("update user status: %w", err)
	}
	return nil
}

func (s *userService) UpdateUserRole(ctx context.Context, userID, role string) error {
	if userID == "" {
		return fmt.Errorf("user id is required")
	}
	role = strings.ToLower(strings.TrimSpace(role))
	if role != "admin" && role != "customer" && role != "staff" {
		return fmt.Errorf("invalid role")
	}
	if err := s.userRepo.UpdateRole(ctx, userID, role); err != nil {
		return fmt.Errorf("update user role: %w", err)
	}
	return nil
}
