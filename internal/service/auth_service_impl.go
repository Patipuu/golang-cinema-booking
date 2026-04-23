package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"booking_cinema_golang/internal/domain"
	"booking_cinema_golang/internal/repository"
	"booking_cinema_golang/internal/utils"

	"golang.org/x/crypto/bcrypt"
)

// Sentinel errors — exported so handlers can match them with errors.Is.
var (
	ErrEmailExists        = errors.New("email already registered")
	ErrUsernameExists     = errors.New("username already taken")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrAccountNotVerified = errors.New("account not verified, please check your email for the OTP code")
	ErrInvalidOTP         = errors.New("invalid OTP code")
	ErrExpiredOTP         = errors.New("OTP code has expired")
	ErrAlreadyVerified    = errors.New("account is already verified")
)

type authServiceImpl struct {
	userRepo    repository.UserRepository
	emailSvc    EmailService
	jwtSecret   string
	expiryHours int
	otpMinutes  int
}

// NewAuthService creates a ready-to-use AuthService implementation.
func NewAuthService(
	userRepo repository.UserRepository,
	emailSvc EmailService,
	jwtSecret string,
	expiryHours, otpMinutes int,
) AuthService {
	return &authServiceImpl{
		userRepo:    userRepo,
		emailSvc:    emailSvc,
		jwtSecret:   jwtSecret,
		expiryHours: expiryHours,
		otpMinutes:  otpMinutes,
	}
}

// Register creates an unverified user and sends a verification OTP by email.
func (s *authServiceImpl) Register(ctx context.Context, email, password, username, fullName, phone string) (*domain.User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	username = strings.TrimSpace(username)

	// Duplicate email check
	if _, err := s.userRepo.FindByEmail(ctx, email); err == nil {
		return nil, ErrEmailExists
	} else if !errors.Is(err, repository.ErrNotFound) {
		return nil, fmt.Errorf("check email: %w", err)
	}

	// Duplicate username check
	if _, err := s.userRepo.FindByUsername(ctx, username); err == nil {
		return nil, ErrUsernameExists
	} else if !errors.Is(err, repository.ErrNotFound) {
		return nil, fmt.Errorf("check username: %w", err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user := &domain.User{
		Username:     username,
		Email:        email,
		PasswordHash: string(hash),
		FullName:     fullName,
		Phone:        phone,
	}
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	if err := s.generateAndSendOTP(ctx, user); err != nil {
		// User is created but email failed — they can use resend-verification
		return nil, fmt.Errorf("send verification email: %w", err)
	}

	return user, nil
}

// Login validates credentials and issues a JWT for verified accounts.
func (s *authServiceImpl) Login(ctx context.Context, email, password string) (*domain.User, string, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	user, err := s.userRepo.FindByEmail(ctx, email)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, "", ErrInvalidCredentials
	}
	if err != nil {
		return nil, "", fmt.Errorf("find user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, "", ErrInvalidCredentials
	}

	if !user.IsVerified {
		return nil, "", ErrAccountNotVerified
	}

	token, err := utils.GenerateToken(s.jwtSecret, user.ID, user.Email, user.Username, user.Role, s.expiryHours)
	if err != nil {
		return nil, "", fmt.Errorf("generate token: %w", err)
	}

	return user, token, nil
}

// VerifyOTP validates the OTP code and marks the account as verified.
func (s *authServiceImpl) VerifyOTP(ctx context.Context, userID, otpCode string) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if errors.Is(err, repository.ErrNotFound) {
		return ErrUserNotFound
	}
	if err != nil {
		return fmt.Errorf("find user: %w", err)
	}

	if user.IsVerified {
		return ErrAlreadyVerified
	}
	if user.OTPCode != otpCode {
		return ErrInvalidOTP
	}
	if utils.IsOTPExpired(user.OTPExpiry) {
		return ErrExpiredOTP
	}

	return s.userRepo.SetVerified(ctx, userID)
}

// ResendVerification generates a new OTP and resends the verification email.
func (s *authServiceImpl) ResendVerification(ctx context.Context, email string) error {
	email = strings.ToLower(strings.TrimSpace(email))

	user, err := s.userRepo.FindByEmail(ctx, email)
	if errors.Is(err, repository.ErrNotFound) {
		return ErrUserNotFound
	}
	if err != nil {
		return fmt.Errorf("find user: %w", err)
	}

	if user.IsVerified {
		return ErrAlreadyVerified
	}

	return s.generateAndSendOTP(ctx, user)
}

// generateAndSendOTP creates an OTP, persists it, and sends the verification email.
func (s *authServiceImpl) generateAndSendOTP(ctx context.Context, user *domain.User) error {
	otpCode, err := utils.GenerateOTP()
	if err != nil {
		return fmt.Errorf("generate OTP: %w", err)
	}
	expiry := utils.OTPExpiry(s.otpMinutes)

	if err := s.userRepo.UpdateOTP(ctx, user.ID, otpCode, expiry); err != nil {
		return fmt.Errorf("save OTP: %w", err)
	}

	return s.emailSvc.SendVerificationEmail(user.Email, user.FullName, otpCode, s.otpMinutes)
}
