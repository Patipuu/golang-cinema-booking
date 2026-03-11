package service

import (
	"context"
	"errors"
	"strings"

	"booking_cinema_golang/internal/config"
	"booking_cinema_golang/internal/domain"
	"booking_cinema_golang/internal/repository"
	"booking_cinema_golang/internal/utils"
	"golang.org/x/crypto/bcrypt"
)

var (
	// ErrInvalidCredentials is returned when email or password is wrong.
	ErrInvalidCredentials = errors.New("invalid email or password")
	// ErrInvalidOTP is returned when OTP code is invalid or expired.
	ErrInvalidOTP = errors.New("invalid or expired OTP")
)

type authService struct {
	userRepo repository.UserRepository
	cfg      *config.Config
}

// NewAuthService creates a new AuthService implementation.
func NewAuthService(userRepo repository.UserRepository, cfg *config.Config) AuthService {
	return &authService{
		userRepo: userRepo,
		cfg:      cfg,
	}
}

func (s *authService) Register(email, password, fullName, username string) (*domain.User, error) {
	ctx := context.Background()

	email = strings.TrimSpace(strings.ToLower(email))
	if username == "" {
		username = email
	}

	if existing, _ := s.userRepo.FindByEmail(ctx, email); existing != nil {
		return nil, errors.New("email already in use")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	u := &domain.User{
		Email:        email,
		Username:     username,
		FullName:     fullName,
		PasswordHash: string(hash),
		IsVerified:   false,
	}

	if err := s.userRepo.Create(ctx, u); err != nil {
		return nil, err
	}

	otp, err := utils.GenerateOTP()
	if err == nil {
		exp := utils.OTPExpiry(s.cfg.OTP.ExpiryMinutes)
		_ = s.userRepo.UpdateOTP(ctx, u.ID, otp, exp)
	}

	return u, nil
}

func (s *authService) Login(email, password string) (*domain.User, string, error) {
	ctx := context.Background()

	email = strings.TrimSpace(strings.ToLower(email))

	u, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil || u == nil {
		return nil, "", ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return nil, "", ErrInvalidCredentials
	}

	token, err := utils.GenerateToken(
		s.cfg.JWT.Secret,
		u.ID,
		u.Email,
		u.Username,
		s.cfg.JWT.ExpiryHours,
	)
	if err != nil {
		return nil, "", err
	}

	return u, token, nil
}

func (s *authService) VerifyOTP(userID, otpCode string) error {
	ctx := context.Background()

	u, err := s.userRepo.FindByID(ctx, userID)
	if err != nil || u == nil {
		return ErrInvalidOTP
	}

	if utils.IsOTPExpired(u.OTPExpiry) || u.OTPCode != otpCode {
		return ErrInvalidOTP
	}

	if err := s.userRepo.SetVerified(ctx, u.ID, true); err != nil {
		return err
	}

	return nil
}

